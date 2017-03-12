# !/bin/bash

if [ -a "$GOPATH/bin/ddns" ]; then
	rm "$GOPATH/bin/ddns"
fi

go install github.com/yangsongfwd/ddns/cmd/ddns

cfg=$GOPATH/bin/ddns.toml
if [ -a $cfg ]; then
	rm $cfg
fi
ln -s $GOPATH/src/github.com/yangsongfwd/ddns/ddns.toml $cfg

static=$GOPATH/bin/ddns_static
if [ -a $static ]; then
	rm $static
fi
ln -s $GOPATH/src/github.com/yangsongfwd/ddns/ddns_static $static

cfg=$GOPATH/bin/resolv.conf
if [ -a $cfg ]; then
	rm $cfg
fi
ln -s $GOPATH/src/github.com/yangsongfwd/ddns/resolv.conf $cfg
