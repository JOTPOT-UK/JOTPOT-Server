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

#Set to yes to delete directorys
deletedirs=false

if [ "$1" == "" ] || [ "$2" == "" ] || [ "$3" == "" ]
then
	echo "Usage: [commandname] source dest version"
	exit 1
fi

in=$1
out=$2/$3
version=$3
htmlfile="$out/downloads.html"
wd=$(pwd)

if [ -e "$out" ]
then
	echo "$out already exists."
	exit 1
fi

echo "Creating js source package..."
node build -in "$in" -out "$out/js-source" -jsonly
echo "Packaging..."
cd "$out/js-source"
tar -cf "$out/js-source.tar" *
tar -czf "$out/js-source.tar.gz" *
zip -q "$out/js-source.zip" *
cd "$wd"
if [ $deletedirs == true ]
then
	rm -r "$out/js-source"
fi

echo "Packaging source..."
cd "$in"
tar -cf "$out/source.tar" *
tar -czf "$out/source.tar.gz" *
zip -q "$out/source.zip" *
cd "$wd"

echo "<html><head><title>JOTPOT Server downloads for version $3</title></head><body><!--Auto generated, do not edit.--><h1>JOTPOT Server downloads for version $3</h1><ul>">"$htmlfile"
echo "$html<li><a href=\"source.tar\">source.tar</a></li><li><a href=\"source.tar.gz\">source.tar.gz</a></li><li><a href=\"source.zip\">source.zip</a></li>">>"$htmlfile"
echo "$html<li><a href=\"js-source.tar\">js-source.tar</a></li><li><a href=\"js-source.tar.gz\">js-source.tar.gz</a></li><li><a href=\"js-source.zip\">js-source.zip</a></li>">>"$htmlfile"

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
	if [ $deletedirs == true ]
	then
		rm -r "$out/$1-$2"
	fi
	echo "$html<li><a href=\"$1-$2.tar\">$1-$2.tar</a></li><li><a href=\"$1-$2.tar.gz\">$1-$2.tar.gz</a></li><li><a href=\"$1-$2.zip\">$1-$2.zip</a></li>">>"$htmlfile"
}

build "linux" "arm"
exit
build "windows" "amd64"
build "windows" "386"
build "linux" "amd64"
build "linux" "386"
build "linux" "arm"
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

echo "$html</ul>Website: <a href=\"https://www.jotpot.co.uk/server\">jotpot.uk/server</a><br>GitHub: <a href=\"https://github.com/jotpot-uk/jotpot-server\">github.com/jotpot-uk/jotpot-server</a></body></html>">>"$htmlfile"

echo "Done!"
