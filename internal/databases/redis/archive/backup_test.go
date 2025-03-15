package archive

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wal-g/wal-g/internal/printlist"
)

func TestBackup_PrintableFields(t *testing.T) {
	b := Backup{
		BackupName:      "my first backup",
		StartLocalTime:  time.Unix(1692811111, 0).UTC(),
		FinishLocalTime: time.Unix(1692822222, 0).UTC(),
		DataSize:        100000,
		BackupSize:      200000,
		Permanent:       true,
		UserData:        []string{"a", "b", "c"},
		Version:         "4.5.4",
		BackupType:      "rdb",
	}
	got := b.PrintableFields()
	prettyStartTime := "Wednesday, 23-Aug-23 17:18:31 UTC"
	prettyFinishTime := "Wednesday, 23-Aug-23 20:23:42 UTC"
	want := []printlist.TableField{
		{
			Name:        "name",
			PrettyName:  "Name",
			Value:       "my first backup",
			PrettyValue: nil,
		},
		{
			Name:        "start_time",
			PrettyName:  "Start time",
			Value:       "2023-08-23T17:18:31Z",
			PrettyValue: &prettyStartTime,
		},
		{
			Name:        "finish_time",
			PrettyName:  "Finish time",
			Value:       "2023-08-23T20:23:42Z",
			PrettyValue: &prettyFinishTime,
		},
		{
			Name:        "user_data",
			PrettyName:  "UserData",
			Value:       "[\"a\",\"b\",\"c\"]",
			PrettyValue: nil,
		},
		{
			Name:        "data_size",
			PrettyName:  "Data size",
			Value:       "100000",
			PrettyValue: nil,
		},
		{
			Name:        "backup_size",
			PrettyName:  "Backup size",
			Value:       "200000",
			PrettyValue: nil,
		},
		{
			Name:        "permanent",
			PrettyName:  "Permanent",
			Value:       "true",
			PrettyValue: nil,
		},
		{
			Name:        "backup_type",
			PrettyName:  "Backup type",
			Value:       "rdb",
			PrettyValue: nil,
		},
		{
			Name:        "version",
			PrettyName:  "Backup version",
			Value:       "4.5.4",
			PrettyValue: nil,
		},
		{
			Name:        "used_memory",
			PrettyName:  "Used memory (limited by maxmemory)",
			Value:       "0",
			PrettyValue: nil,
		},
		{
			Name:        "used_memory_rss",
			PrettyName:  "Used memory (as seen by OS))",
			Value:       "0",
			PrettyValue: nil,
		},
	}
	assert.Equal(t, want, got)
}
