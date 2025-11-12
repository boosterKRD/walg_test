# PostgreSQL Switchover Plan

Switchover plan step by step

1. Check the cluster state (if the state is okay, then go further)
2. Stop sending application queries to the primary
3. Wait until the lag is 0
4. Stop the primary
5. Promote the replica
6. [Convert the old primary to a replica](#6-convert-the-old-primary-to-a-replica)

[Convert the old primary to a replica](#6-convert-the-old-primary-to-a-replica)



<a id="check-cluster-state"></a>
## 1. Check clutser state 
Before procces to switchover we must check the cluster state and if it it satisfises all requerments start the process:
- All members in clusters must be reachable (Primary and Replica)
- The replication lag must be small (within 100MB)

On primary execute the fellowing SQL query
```sql
sudo -u postgres psql -d postgres -c "SELECT client_addr, state, sync_state, replay_lag, pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn)) AS total_lag FROM pg_stat_replication;"
```

> Example of th eoutput
>  client_addr   |   state   | sync_state | replay_lag       | total_lag
> ----------------+-----------+------------+------------------+-----------
> 35.176.231.227 | streaming | async      | 00:00:00.002719  | 0 bytes

> - client_addr – the IP address of the replica connected to the primary server.
> - state – current streaming status; it must be streaming, otherwise the replica is not receiving WAL data.
> - sync_state – synchronization mode: async, sync, quorum, or potential. We use the async mode, which means the primary server does not wait for the replica to confirm WAL writes.
> - replay_lag – delay in applying WAL records on the replica, measured in seconds.
> - total_lag – total WAL difference between the primary and the replica, measured in bytes (0 means the replica is fully caught up).

## 1. Shutdown the Old Primary (if reachable)
**Check current role:**
```bash
sudo -u postgres psql -tA -d postgres -c "SELECT CASE WHEN pg_is_in_recovery() THEN 'Replica' ELSE 'Primary' END;"
sudo -u postgres psql -tA -d postgres -c "SELECT COALESCE((now() - last_msg_receipt_time)::text, 'REPLICATION NOT WORK') AS replication_lag_sec FROM pg_stat_wal_receiver where status='streaming' UNION ALL SELECT 'REPLICATION NOT WORK' WHERE NOT EXISTS (SELECT 1 FROM pg_stat_wal_receiver where status='streaming');"

```
**Stop PostgreSQL service:**
```bash
sudo systemctl stop postgresql@17-main.service
sudo systemctl status postgresql@17-main.service
```

## 2. Reroute Applications to the New Primary
Update all application connections to point to the new primary node.

## 3. Promote the New Primary
```bash
sudo -u postgres psql -d postgres -c "SELECT pg_promote();"
```

## 4. Recreate the Former Primary as Replica
Use either 
 
 - **pg_rewind** 
 - **pgbackrest** 
 - **pg_basebackup**

---

### Environment
| Hostname    | IP             | Provider  |
|--------------|----------------|-----------|
| t4b_testdb1  | 23.81.34.232   | LeaseWeb  |
| t4b_testdb2  | 35.176.231.227 | AWS       |

**Credentials:**
```
export SUPERUSER_NAME=postgres
export SUPERUSER_PASSWORD='GMpkZgX1j00L'
export REPLICA_USER=replica
export PASSWORD_REPLICA_USER='b8zzZ6VWtuR2'
export PRIMARY_IP=23.81.34.232
export PG_VER=17
export STANZA_NAME=test1
```

---

### Commands on Replica
```bash
# make sure that it's an replica and get the curremt replication lag 
sudo -u postgres psql -tA -d postgres -c "SELECT CASE WHEN pg_is_in_recovery() THEN 'Replica' ELSE 'Primary' END;"
sudo -u postgres psql -tA -d postgres -c "SELECT pg_size_pretty(pg_wal_lsn_diff(pg_last_wal_receive_lsn(), pg_last_wal_replay_lsn())) AS replay_lag;"

# switch replica to primary mode
sudo -u postgres psql -d postgres -c "SELECT pg_promote();"
sudo -u postgres psql -tA -d postgres -c "SELECT CASE WHEN pg_is_in_recovery() THEN 'Replica' ELSE 'Primary' END;"
```

### Commands on Old Primary and new Primary
```bash
sudo -u postgres psql -d testdb -c "SELECT pg_switch_wal();"
```

---

## 6. Convert the old primary to a replica

Run on the old primary:
```bash
sudo systemctl stop postgresql@${PG_VER}-main.service
sudo systemctl status postgresql@${PG_VER}-main.service
sudo -u ${SUPERUSER_NAME} env \
  SUPERUSER_NAME=${SUPERUSER_NAME} \
  SUPERUSER_PASSWORD=${SUPERUSER_PASSWORD} \
  REPLICA_USER=${REPLICA_USER} \
  PASSWORD_REPLICA_USER=${PASSWORD_REPLICA_USER} \
  PRIMARY_IP=${PRIMARY_IP} \
  PGPASSWORD="${SUPERUSER_PASSWORD}" \
  /usr/lib/postgresql/${PG_VER}/bin/pg_rewind \
    --restore-target-wal \
    --target-pgdata=/var/lib/postgresql/${PG_VER}/main/ \
    --source-server="host=${PRIMARY_IP} port=5432 user=${SUPERUSER_NAME}" -R
```

If you see output like this, it’s normal:
```text
pg_rewind: servers diverged at WAL location 2D/A0 on timeline 1
pg_rewind: no rewind required
OR
pg_rewind: servers diverged at WAL location 2D/B005640 on timeline 2
pg_rewind: rewinding from last common checkpoint at 2D/B005590 on timeline 2
pg_rewind: Done!
```

Then edit the replication config:
```bash

if grep -q "^primary_conninfo" /var/lib/postgresql/${PG_VER}/main/postgresql.auto.conf; then
  sudo sed -i "s|^primary_conninfo.*|primary_conninfo = 'user=${REPLICA_USER} password=${PASSWORD_REPLICA_USER} host=${PRIMARY_IP} port=5432 sslmode=prefer target_session_attrs=any'|" /var/lib/postgresql/${PG_VER}/main/postgresql.auto.conf
else
  echo "primary_conninfo = 'user=${REPLICA_USER} password=${PASSWORD_REPLICA_USER} host=${PRIMARY_IP} port=5432 sslmode=prefer target_session_attrs=any'" | sudo tee -a /var/lib/postgresql/${PG_VER}/main/postgresql.auto.conf
fi
```
Restart the service:
```bash
sudo systemctl restart postgresql@${PG_VER}-main.service
sudo systemctl status postgresql@${PG_VER}-main.service
sudo -u postgres psql -tA -d postgres -c "SELECT CASE WHEN pg_is_in_recovery() THEN 'Replica' ELSE 'Primary' END;"
sudo -u postgres psql -tA -d postgres -c "SELECT pg_size_pretty(pg_wal_lsn_diff(pg_last_wal_receive_lsn(), pg_last_wal_replay_lsn())) AS replay_lag;"
```

---

### Verify on Primary
```bash
sudo -u postgres psql -d testdb -c "SELECT * FROM pg_stat_replication;"
sudo -u postgres psql -d testdb -c "SELECT pg_switch_wal();"
```

---

## 6. Restore from Backup (if needed)
If rewind fails, restore from backup (S3 retrieval may incur costs).
```bash
sudo systemctl stop postgresql@${PG_VER}-main.service
sudo systemctl status postgresql@${PG_VER}-main.service
sudo rm -rf /var/lib/postgresql/${PG_VER}/main/*

# Get info about last backup
pgbackrest info --output=json | jq '.[0].backup | last'
# Restore last backup
sudo -u postgres env STANZA_NAME=${STANZA_NAME} \
  pgbackrest --stanza=${STANZA_NAME} \
  restore --delta --log-level-console=info --type=standby

if grep -q "^primary_conninfo" /var/lib/postgresql/${PG_VER}/main/postgresql.auto.conf; then
  sudo sed -i "s|^primary_conninfo.*|primary_conninfo = 'user=${REPLICA_USER} password=${PASSWORD_REPLICA_USER} host=${PRIMARY_IP} port=5432 sslmode=prefer target_session_attrs=any'|" /var/lib/postgresql/${PG_VER}/main/postgresql.auto.conf
else
  echo "primary_conninfo = 'user=${REPLICA_USER} password=${PASSWORD_REPLICA_USER} host=${PRIMARY_IP} port=5432 sslmode=prefer target_session_attrs=any'" | sudo tee -a /var/lib/postgresql/${PG_VER}/main/postgresql.auto.conf
fi

sudo systemctl restart postgresql@${PG_VER}-main.service
sudo systemctl status postgresql@${PG_VER}-main.service  
```

Then verify on Primary:
```bash
sudo -u postgres psql -d testdb -c "SELECT * FROM pg_stat_replication;"
sudo -u postgres psql -d postgres -c \
sudo -u postgres psql -d postgres -c "SELECT client_addr, state, sync_state, write_lag, flush_lag, replay_lag, pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), sent_lsn)) AS send_lag, pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn)) AS total_lag FROM pg_stat_replication;"
```

sudo -u postgres psql -tA -d postgres -c "SELECT COALESCE((now() - last_msg_receipt_time)::text, 'REPLICATION NOT WORK') AS replication_lag_sec FROM pg_stat_wal_receiver where status='streaming' UNION ALL SELECT 'REPLICATION NOT WORK' WHERE NOT EXISTS (SELECT 1 FROM pg_stat_wal_receiver where status='streaming');"



select now() - last_msg_receipt_time as replication_lag_seconds from pg_stat_wal_receiver;


