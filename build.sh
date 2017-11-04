#! /usr/bin/env bash

PRIMARY_GOPATH=`echo $GOPATH | sed -e 's/:.*//'`
if [ -z $PRIMARY_GOPATH ]; then
	PRIMARY_GOPATH=`echo $GOPATH | sed -e 's/.*://'`
fi
PATH=$PRIMARY_GOPATH/bin:$PATH

if which git 2>&1 > /dev/null; then
    if [ -z "`git status --porcelain`" ]; then
        STATE=clean
    else
        STATE=dirty
    fi
    GIT_VERSION=`git rev-parse HEAD`-$STATE
else
    GIT_VERSION=Unknown
fi

touch main.go
go install -v -ldflags "-X github.com/richardwilkes/toolbox/cmdline.GitVersion=$GIT_VERSION"
