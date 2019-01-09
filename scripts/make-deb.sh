#!/bin/bash

set -eo pipefail

if [[ -z "$VERSION" ]]; then
    echo "VERSION unset!"
    exit 1
fi

case $GOARCH in
    ''|amd64)   ARCH=amd64 ;;
    mips*)      ARCH=mips ;;
    *)          ARCH=$GOARCH ;;
esac

# init system, either sysv or systemd
: ${INIT:=systemd}
case $INIT in
    systemd|sysv) : ;;
    *) echo "Invalid value for INIT ($INIT)" ; exit 1 ;;
esac

wolssh="wolssh-${GOARCH}"
wolssh="${wolssh%-}"

tmpdir="$(mktemp -d)"
if [[ -z $SAVETMP ]]; then
    trap "rm -rf $tmpdir" EXIT
else
    echo "working in $tmpdir"
fi

echo "preparing data"
datadir=$tmpdir/data
install -Dm755 $wolssh $datadir/usr/bin/wolssh
install -Dm644 default.ini $datadir/etc/wolssh.ini
install -Dm644 debian/wolssh.default $datadir/etc/default/wolssh
if [[ $INIT == sysv ]]; then
    install -Dm755 debian/wolssh.init $datadir/etc/init.d/wolssh
else # systemd
    install -Dm644 debian/wolssh.service $datadir/lib/systemd/system/wolssh.service
fi
tar --owner=root:0 --group=root:0 -czf $tmpdir/data.tar.gz -C $datadir .

echo "creating debian control"
debdir=$tmpdir/debian
mkdir $debdir
sed -e "s|@VERSION@|$VERSION|g; s|@ARCH@|$ARCH|g" \
    debian/control >$debdir/control
install -m755 debian/postinst $debdir/
install -m644 debian/conffiles $debdir/
tar --owner=root:0 --group=root:0 -czf $tmpdir/control.tar.gz -C $debdir .
echo '2.0' >$tmpdir/debian-binary

pkgname="wolssh-$VERSION-$ARCH.deb"
echo "creating $pkgname"
rm -f $pkgname
ar -rcs $pkgname $tmpdir/debian-binary $tmpdir/control.tar.gz $tmpdir/data.tar.gz
