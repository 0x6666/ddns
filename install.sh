# !/bin/bash

if [ -a "$GOPATH/bin/ddns" ]; then
	rm "$GOPATH/bin/ddns"
fi

go install github.com/inimei/ddns/cmd/ddns

cfg=$GOPATH/bin/ddns.toml
if [ ! -f $cfg ]; then
	ln -s $GOPATH/src/github.com/inimei/ddns/ddns.toml $cfg
fi

static=$GOPATH/bin/ddns_static
if [ ! -d $static ]; then
	ln -s $GOPATH/src/github.com/inimei/ddns/ddns_static $static
fi
