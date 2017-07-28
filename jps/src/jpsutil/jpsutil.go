/*

	JOTPOT Server
	Version 26A-0

	Copyright (c) 2016-2017 Jacob O'Toole

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE.

*/

package jpsutil

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
)

//GetNodePath checks PATH env varaible and cwd/node for the node binary and returns the path of it.
func GetNodePath() string {
	var thisPath string
	var err error
	command := "node"
	if runtime.GOOS == "windows" {
		command = "node.exe"
	}
	for _, p := range filepath.SplitList(os.Getenv("PATH")) {
		thisPath = filepath.Join(p, command)
		_, err = os.Stat(thisPath)
		if err != nil {
			continue
		}
		return thisPath
	}
	thisPath, err = os.Getwd()
	thisPath = filepath.Join(thisPath, "node", command)
	_, err = os.Stat(thisPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "No Node.js binary")
		os.Exit(1)
		panic("No Node.js binary")
	}
	return thisPath
}

//Uint32ToBytes creates a slice of 4 bytes as a representation of a uint32
func Uint32ToBytes(in uint32) []byte {
	return []byte{byte(in >> 24), byte((in >> 16) & 255), byte((in >> 8) & 255), byte(in & 255)}
}

//Uint32FromBytes creates a uint32 from a slice of 4 bytes
func Uint32FromBytes(in []byte) uint32 {
	return (uint32(in[0]) << 24) + (uint32(in[1]) << 16) + (uint32(in[2]) << 8) + uint32(in[3])
}

//GetData takes a connection and reads the specified amout of bytes and returns them in a slice
func GetData(con net.Conn, toGet int) (out []byte) {
	var got int
	var buff []byte
	if toGet < 1024 {
		buff = make([]byte, toGet)
	} else {
		buff = make([]byte, 1024)
	}
	var n int
	var err error
	for got < toGet {
		if toGet-got < len(buff) {
			buff = buff[:toGet-got]
		}
		n, err = con.Read(buff)
		if err != nil && err != io.EOF {
			panic(err)
		} else if n == 0 && err == io.EOF {
			panic(err)
		}
		got += n
		out = append(out, buff[:n]...)
	}
	return
}

//CheckAddr checks if an address exists, if so, it returns true. If not, false.
func CheckAddr(addr string) bool {
	con, err := net.Dial("tcp", addr)
	if err != nil {
		return false
	}
	con.Close()
	return true
}

//CopyFile copies file p1 to p2
func CopyFile(p1, p2 string) error {
	src, err := os.Open(p1)
	if err != nil {
		return err
	}
	dest, err := os.OpenFile(p2, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return err
	}
	stats, err := src.Stat()
	if err != nil {
		return err
	}
	err = dest.Truncate(stats.Size())
	if err != nil {
		return err
	}
	buff := make([]byte, 1024)
	var n int
	var err2 error
	for {
		n, err = src.Read(buff)
		if err != nil && err != io.EOF {
			return err
		}
		_, err2 = dest.Write(buff[:n])
		if err2 != nil {
			return err2
		}
		if err == io.EOF {
			break
		}
	}
	if err = src.Close(); err != nil {
		return err
	}
	if err = dest.Close(); err != nil {
		return err
	}
	return nil
}
