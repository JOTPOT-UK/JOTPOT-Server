#!/bin/bash

if [ "$1" == "" ]
then
	echo "Source dir required."
fi
if [ "$2" == "" ]
then
	echo "Dest dir required."
fi

in=$1
out=$2

function build {
	export GOOS="$1"
	export GOARCH="$2"
	node build -in "$in" -out "$out/$1-$2" -go "/usr/local/go/bin/go"
}

build "windows" "amd64"
build "windows" "386"
build "linux" "amd64"
build "linux" "386"
build "linux" "ppc64"
build "linux" "ppc64le"
build "linux" "mips"
build "linux" "mipsle"
build "linux" "mips64"
build "linux" "mips64le"
build "android" "arm"
build "darwin" "386"
build "darwin" "amd64"
build "darwin" "arm"
build "darwin" "arm64"
build "dragonfly" "amd64"
build "freebsd" "386"
build "freebsd" "amd64"
build "freebsd" "arm"
build "netbsd" "386"
build "netbsd" "amd64"
build "netbsd" "arm"
build "openbsd" "386"
build "openbsd" "amd64"
build "openbsd" "arm"
build "plan9" "386"
build "plan9" "amd64"
build "solaris" "amd64"
