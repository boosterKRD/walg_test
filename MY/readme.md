git add .  
git commit -m "v2.1.55"
git tag -a v2.1.55 -m "Release v2.1.55"

git push origin main  
git push origin v2.1.55




sudo rm /etc/apt/sources.list.d/walg.list
sudo apt-get update
    echo "deb [trusted=yes] https://boosterKRD.github.io/walg_test stable main" | sudo tee /etc/apt/sources.list.d/walg.list
curl -fsSL https://boosterKRD.github.io/walg_test/public-key.asc | sudo tee /etc/apt/trusted.gpg.d/walg.asc
echo "deb https://boosterKRD.github.io/walg_test $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/walg.list
sudo apt-get update
apt-cache policy wal-g-pg
sudo apt-get install wal-g-pg=1.0.14
sudo apt-get install wal-g-pg
sudo apt remove wal-g-pg
sudo apt purge wal-g-pg
sudo apt-get install --only-upgrade wal-g-pg
wal-g -v





====
https://salsa.debian.org/postgresql/pgbackrest/-/blob/master/debian/control?ref_type=heads
https://www.postgresql.org/message-id/flat/YOMyXyErQ50je0zh%40msg.df7cb.de#c997486f4403d814ff6d1967772e58eb




rm -rf /home/vagrant/wal-g-build/wal-g_1.0.0*
rm -rf submodules/brotli
rm -rf obj-aarch64-linux-gnu/
cd ..
tar --exclude=.git --exclude=debian -czf wal-g_1.0.0.orig.tar.gz wal-g
cd wal-g
dpkg-buildpackage -us -uc
ls -la ../

sudo dpkg -i ../wal-g_1.0.0-1_arm64.deb




---
https://salsa.debian.org/postgresql/pgbackrest/-/blob/master/debian/control?ref_type=heads
https://www.postgresql.org/message-id/flat/YOMyXyErQ50je0zh%40msg.df7cb.de#c997486f4403d814ff6d1967772e58eb




sudo apt update && sudo apt install -y pbuilder debootstrap devscripts debhelper dh-golang dpkg-dev fakeroot golang-any

sudo pbuilder create --distribution jammy --debootstrapopts --variant=buildd

cd /walg_repo/BoosterKRD/wal-g-build/wal-g
go mod tidy
go mod vendor

cd /walg_repo/BoosterKRD/wal-g-build/
tar --exclude='wal-g/debian' -czf wal-g_1.0.0.orig.tar.gz wal-g

cd wal-g
debuild -S -sa

sudo pbuilder build ../wal-g_1.0.0-1.dsc