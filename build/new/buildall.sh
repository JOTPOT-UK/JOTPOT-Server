#!/bin/bash

#	JOTPOT Server
#	Version 26A-0
#
#	Copyright (c) 2016-2017 Jacob O'Toole
#
#	Permission is hereby granted, free of charge, to any person obtaining a copy
#	of this software and associated documentation files (the "Software"), to deal
#	in the Software without restriction, including without limitation the rights
#	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
#	copies of the Software, and to permit persons to whom the Software is
#	furnished to do so, subject to the following conditions:
#
#	The above copyright notice and this permission notice shall be included in all
#	copies or substantial portions of the Software.
#
#	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
#	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
#	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
#	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
#	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
#	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
#	SOFTWARE.

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
wd=$(pwd)

function build {
	echo "Building for $1 on $2..."
	export GOOS="$1"
	export GOARCH="$2"
	node build -in "$in" -out "$out/$1-$2" -go "/usr/local/go/bin/go"
	echo "Packaging..."
	cd "$out/$1-$2"
	tar -cf "$out/$1-$2.tar" *
	tar -czf "$out/$1-$2.tar.gz" *
	zip -q "$out/$1-$2.zip" *
	cd "$wd"
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

echo "Done!"
