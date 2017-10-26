#!/bin/bash

#	JOTPOT Server
#	Version 26B-0
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

#Set to true to delete directorys
deletedirs=true

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
node build -in "$in" -out "$out/jps-js-source" -jsonly
echo "Packaging..."
cd "$out/jps-js-source"
tar -cf "$out/jps-js-source.tar" *
tar -czf "$out/jps-js-source.tar.gz" *
zip -q "$out/jps-js-source.zip" *
cd "$wd"
if [ $deletedirs == true ]
then
	rm -r "$out/jps-js-source"
fi

echo "Packaging source..."
cd "$in"
tar -cf "$out/jps-source.tar" *
tar -czf "$out/jps-source.tar.gz" *
zip -q "$out/jps-source.zip" *
cd "$wd"

echo "<html><head><title>JOTPOT Server downloads for version $3</title></head><body><!--Auto generated, do not edit.--><h1>JOTPOT Server downloads for version $3</h1><a href=\"https://github.com/JOTPOT-UK/JOTPOT-Server/releases/tag/$3\" target=\"_blank\">View release (with release notes) on GitHub (github.com/JOTPOT-UK/JOTPOT-Server/releases/tag/$3)</a><ul>">"$htmlfile"

echo "<li>jps-source.tar: <a href=\"jps-source.tar\">Download from JOTPOT</a></li>">>"$htmlfile"
echo "<li>jps-source.tar.gz: <a href=\"jps-source.tar.gz\">Download from JOTPOT</a><a href=\"https://github.com/JOTPOT-UK/JOTPOT-Server/archive/$version.tar.gz\">Download from GitHub</a></li>">>"$htmlfile"
echo "<li>jps-source.zip: <a href=\"jps-source.zip\">Download from JOTPOT</a><a href=\"https://github.com/JOTPOT-UK/JOTPOT-Server/archive/$version.zip\">Download from GitHub</a></li>">>"$htmlfile"

echo "<li>jps-js-source.tar: <a href=\"jps-js-source.tar\">Download from JOTPOT</a></li>">>"$htmlfile"
echo "<li>jps-js-source.tar.gz: <a href=\"jps-js-source.tar.gz\">Download from JOTPOT</a><a href=\"https://github.com/JOTPOT-UK/JOTPOT-Server/releases/download/$version/jps-js-source.tar.gz\">Download from GitHub</a></li>">>"$htmlfile"
echo "<li>jps-js-source.zip: <a href=\"jps-js-source.zip\">Download from JOTPOT</a><a href=\"https://github.com/JOTPOT-UK/JOTPOT-Server/releases/download/$version/jps-js-source.zip\">Download from GitHub</a></li>">>"$htmlfile"

function build {
	export GOOS="$1"
	export GOARCH="$2"
	arch="$2"
	archlabel="$arch"
	if [ "$2" == "arm" ]
	then
		arch="armv$3"
		archlabel="ARMv$3"
		export GOARM="$3"
	elif [ "$2" == "arm64" ]
	then
		arch="armv8"
		archlabel="ARMv8"
	elif [ "$2" == "386" ]
	then
		arch="x86-$3"
		export GO386="$3"
		if [ "$3" == "387" ]
		then
			archlabel="x86-x87"
		else
			archlabel="x86-SSE2"
		fi
	elif [ "$2" == "amd64" ]
	then
		arch="x86-64"
		archlabel="x86-64"
	fi
	echo "Building for $1 on $arch..."
	if node build -in "$in" -out "$out/jps-$1-$arch" -go "/usr/local/go/bin/go" --hide-errors
	then
		echo "Packaging..."
		cd "$out/jps-$1-$arch"
		tar -cf "$out/jps-$1-$arch.tar" *
		tar -czf "$out/jps-$1-$arch.tar.gz" *
		zip -q "$out/jps-$1-$arch.zip" *
		cd "$wd"
		echo "<li>jps-$1-$archlabel.tar: <a href=\"jps-$1-$arch.tar\">Download from JOTPOT</a></li>">>"$htmlfile"
		echo "<li>jps-$1-$archlabel.tar.gz: <a href=\"jps-$1-$arch.tar.gz\">Download from JOTPOT</a><a href=\"https://github.com/JOTPOT-UK/JOTPOT-Server/releases/download/$version/jps-$1-$arch.tar.gz\">Download from GitHub</a></li>">>"$htmlfile"
		echo "<li>jps-$1-$archlabel.zip: <a href=\"jps-$1-$arch.zip\">Download from JOTPOT</a><a href=\"https://github.com/JOTPOT-UK/JOTPOT-Server/releases/download/$version/jps-$1-$arch.zip\">Download from GitHub</a></li>">>"$htmlfile"
	else
		echo "!!! $1 on $arch failed to build"
	fi
	if [ $deletedirs == true ]
	then
		rm -r "$out/jps-$1-$arch"
	fi
}

build "android" "arm" "5"
build "android" "arm" "6"
build "android" "arm" "7"
build "darwin" "386" "387"
build "darwin" "386" "sse2"
build "darwin" "amd64"
build "darwin" "arm" "5"
build "darwin" "arm" "6"
build "darwin" "arm" "7"
build "darwin" "arm64"
build "dragonfly" "amd64"
build "freebsd" "386" "387"
build "freebsd" "386" "sse2"
build "freebsd" "amd64"
build "freebsd" "arm" "5"
build "freebsd" "arm" "6"
build "freebsd" "arm" "7"
build "linux" "386" "387"
build "linux" "386" "sse2"
build "linux" "amd64"
build "linux" "arm" "5"
build "linux" "arm" "6"
build "linux" "arm" "7"
build "linux" "arm64"
build "linux" "ppc64"
build "linux" "ppc64le"
build "linux" "mips"
build "linux" "mipsle"
build "linux" "mips64"
build "linux" "mips64le"
build "netbsd" "386" "387"
build "netbsd" "386" "sse2"
build "netbsd" "amd64"
build "netbsd" "arm" "5"
build "netbsd" "arm" "6"
build "netbsd" "arm" "7"
build "openbsd" "386" "387"
build "openbsd" "386" "sse2"
build "openbsd" "amd64"
build "openbsd" "arm" "5"
build "openbsd" "arm" "6"
build "openbsd" "arm" "7"
build "plan9" "386" "387"
build "plan9" "386" "sse2"
build "plan9" "amd64"
build "solaris" "amd64"
build "windows" "386" "387"
build "windows" "386" "sse2"
build "windows" "amd64"

echo "</ul>Website: <a href=\"https://www.jotpot.co.uk/server\">jotpot.uk/server</a><br>GitHub: <a href=\"https://github.com/jotpot-uk/jotpot-server\">github.com/jotpot-uk/jotpot-server</a></body></html>">>"$htmlfile"

echo "Done!"
