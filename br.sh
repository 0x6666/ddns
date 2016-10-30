# !/bin/bash

set -e

if [ -a "$GOPATH/bin/ddns" ]; then
	rm "$GOPATH/bin/ddns"
fi

go install github.com/inimei/ddns/cmd/ddns

cfg=$GOPATH/bin/ddns.toml
if [ -a "$cfg" ]; then
	echo "$cfg already exist...."
else 
	ln -s $GOPATH/src/github.com/inimei/ddns/ddns.toml $cfg
fi

static=$GOPATH/bin/ddns_static
if [ -a "$static" ]; then
	echo "$static already exist...."
else 
	ln -s $GOPATH/src/github.com/inimei/ddns/ddns_static $static
fi

/home/inimei/go/gopath/bin/ddns
