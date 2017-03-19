# !/bin/bash

project=ddns.d

if [ -a "$GOPATH/bin/ddns" ]; then
	rm "$GOPATH/bin/ddns"
fi

go install github.com/yangsongfwd/ddns/cmd/ddns

cfgdir=$GOPATH/bin/$project
if [ -a $cfgdir ]; then
	rm -fr $cfgdir
fi
mkdir -p $cfgdir

ln -s $GOPATH/src/github.com/yangsongfwd/ddns/conf/ddns.toml $cfgdir/ddns.toml
ln -s $GOPATH/src/github.com/yangsongfwd/ddns/conf/routes $cfgdir/routes
ln -s $GOPATH/src/github.com/yangsongfwd/ddns/conf/resolv.conf $cfgdir/resolv.conf
ln -s $GOPATH/src/github.com/yangsongfwd/ddns/conf/hosts $cfgdir/hosts

static=$GOPATH/bin/ddns_static
if [ -a $static ]; then
	rm $static
fi
ln -s $GOPATH/src/github.com/yangsongfwd/ddns/ddns_static $static
