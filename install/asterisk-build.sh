#!/bin/bash
LANG=C
SELINUX_STATE=$(getenforce)
OS_ID=$(grep -oP '(?<=^ID=).+' /etc/os-release | tr -d '"')
CENTOS=`rpm --eval '%{centos_ver}'`

if [ "$SELINUX_STATE" == "Enforcing" ]; then
  echo "SELinux is enabled, disabling"
  sed -c -i "s/\SELINUX=.*/SELINUX=disabled/" /etc/sysconfig/selinux
elif [ "$SELINUX_STATE" == "Permissive" ]; then
  echo "SELinux is permissive, disabling"
  sed -c -i "s/\SELINUX=.*/SELINUX=disabled/" /etc/sysconfig/selinux
else
  echo "SELinux is disabled (or missing)"
fi

if [ "$OS_ID" == "centos" ]; then
  PACKAGE_MANAGER="dnf"
$PACKAGE_MANAGER -y group install "Development Tools"
$PACKAGE_MANAGER -y install dnf-plugins-core
if [ $CENTOS == "8"  ]; then
$PACKAGE_MANAGER -y install https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm
$PACKAGE_MANAGER config-manager --set-enabled powertools
elif [ $CENTOS == "9" ]; then
$PACKAGE_MANAGER -y install https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm
$PACKAGE_MANAGER config-manager --set-enabled crb
fi

$PACKAGE_MANAGER install -y git subversion speex wget libedit-devel libuuid-devel jansson-devel libxml2-devel sqlite-devel libsrtp-devel

$PACKAGE_MANAGER install -y mariadb-server mariadb-connector-odbc mariadb-devel
$PACKAGE_MANAGER install -y unixODBC unixODBC-devel libtool-ltdl libtool-ltdl-devel

$PACKAGE_MANAGER install -y epel-release
$PACKAGE_MANAGER install -y pwgen

$PACKAGE_MANAGER install -y libpcap-devel
$PACKAGE_MANAGER install -y opus-devel
elif [ "$OS_ID" == "amzn" ]; then
  PACKAGE_MANAGER="yum"
  COPT="--with-jansson-bundled"
$PACKAGE_MANAGER -y update
$PACKAGE_MANAGER -y group install "Development Tools"
$PACKAGE_MANAGER install -y git subversion speex wget libedit-devel libuuid-devel jansson-devel libxml2-devel sqlite-devel libsrtp-devel
$PACKAGE_MANAGER install -y mariadb-server mysql-connector-odbc mariadb-devel
$PACKAGE_MANAGER install -y unixODBC unixODBC-devel libtool-ltdl libtool-ltdl-devel
$PACKAGE_MANAGER install -y pwgen
$PACKAGE_MANAGER install -y libpcap-devel
$PACKAGE_MANAGER install -y opus-devel
fi


mkdir -p /usr/src/
cd /usr/src/
wget -cq https://downloads.asterisk.org/pub/telephony/certified-asterisk/asterisk-certified-18.9-cert5.tar.gz
wget -cq https://downloads.asterisk.org/pub/telephony/certified-asterisk/asterisk-certified-18.9-cert5.md5
if [ `md5sum -c asterisk-certified-18.9-cert5.md5 | awk '{print $2}'` = "OK" ]; then
echo "asterisk source downlad compete, extracting tgz"
tar -zxf asterisk-certified-18.9-cert5.tar.gz
rm -f asterisk-certified-18.9-cert5/menuselect.makeopts

git clone https://github.com/cisco/libsrtp
cd libsrtp
./configure; make -j8; make install
cd ..


git clone https://github.com/asterisk/pjproject pjproject

if [ ! -f /usr/local/lib/libpjmedia-audiodev.so.2 ]; then
cd pjproject && ./configure --prefix=/usr/local --enable-shared --disable-sound --disable-resample --disable-video --disable-opencore-amr
make dep && make -j8 && make install
cd ..
fi


git clone https://github.com/irontec/sngrep sngrep
cd sngrep
./bootstrap.sh && ./configure --prefix=/usr/local && make -j8 && make install
cd ..


cd asterisk-certified-18.9-cert5
contrib/scripts/get_mp3_source.sh
#workaround to ast install bug
mkdir -p /var/lib/asterisk/phoneprov/
./configure $COPT && make -j8 && make samples && make config && make install
echo asterisk installed!
cd ..

wget -c http://downloads.digium.com/pub/telephony/codec_opus/asterisk-18.0/x86-64/codec_opus-18.0_current-x86_64.tar.gz
tar -zxf codec_opus-18.0_current-x86_64.tar.gz
cd codec_opus-18.0_1.3.0-x86_64
mkdir -p /usr/lib/asterisk/modules
cp codec_opus.so /usr/lib/asterisk/modules
cp format_ogg_opus.so /usr/lib/asterisk/modules
mkdir -p /var/lib/asterisk/documentation/thirdparty
cp codec_opus_config-en_US.xml /var/lib/asterisk/documentation/thirdparty
cd ..

wget -c http://downloads.digium.com/pub/telephony/codec_g729/asterisk-18.0/x86-64/codec_g729a-18.0_3.1.10-x86_64.tar.gz
#todo...
else
 echo "asterisk source file MD5 mismatch build failed!!!"
fi
