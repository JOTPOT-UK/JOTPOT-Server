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

package jpsd

import (
	"fmt"
	"io"
	"io/ioutil"
	"jpsutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

type procReader struct {
	value string
}

func (r *procReader) Write(data []byte) (n int, err error) {
	defer func() {
		rerr := recover()
		if rerr != nil {
			n = 0
			err = rerr.(error)
		}
	}()
	r.value += string(data)
	return len(data), nil
}

type proc struct {
	sDir        string
	p           *exec.Cmd
	stdout      *procReader
	stderr      *procReader
	index       int
	controlAddr string
	state       byte
}

func (p *proc) wait() {
	err := p.p.Wait()
	fmt.Println(p.p.ProcessState.Success(), err == nil)
	if p.p.ProcessState.Success() && err == nil {
		p.state = 0
	} else {
		p.state = 1
	}
}

var procs []*proc

func newProc(wd string) bool {
	for _, p := range procs {
		if p.sDir == wd && p.state == 2 {
			return false
		}
	}
	awd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	var sock string
	if runtime.GOOS == "windows" {
		socks := [5]string{"127.5.5.5:5", "127.55.55.55:55", "127.7.7.7:7", "127.77.77.77:77", "127.3.5.7:9"}
		done := false
		for _, v := range socks {
			if !jpsutil.CheckAddr(v) {
				sock = v
				done = true
				break
			}
		}
		if !done {
			panic("No available addresses for data server")
		}
	} else {
		sock = filepath.Join(dir, "data"+string(len(procs)+48)+".sock")
	}
	stdout := &procReader{""}
	stderr := &procReader{""}
	c := exec.Command(jpsutil.GetNodePath(), filepath.Join(awd, filepath.Dir(os.Args[0]), "jps", "run"), "-data", sock)
	c.Stdout = stdout
	c.Stderr = stderr
	c.Dir = wd
	tp := proc{wd, c, stdout, stderr, len(procs), sock, 2}
	procs = append(procs, &tp)
	c.Start()
	go tp.wait()
	return true
}

func listProcs(_ net.Conn) ([]byte, bool) {
	out := []byte{9, byte(len(procs))}
	for _, p := range procs {
		out = append(out, byte(p.index), p.state, byte(len(p.sDir)))
		out = append(out, []byte(p.sDir)...)
	}
	return out, true
}

var dir string

//GotMessage determines how each request is handled.
//The request method is looked up in this map, and it's value called with the connection as the only argument.
//THe function should return a slice of bytes and a bool, if the bool is true, the returned slice will be writen to the connection.
var GotMessage = map[byte]func(net.Conn) ([]byte, bool){
	9: listProcs,
	10: func(con net.Conn) ([]byte, bool) {
		buff := make([]byte, 1)
		var n int = 0
		var err error
		for n < 1 {
			n, err = con.Read(buff)
			if err != nil {
				panic(err)
			}
		}
		wdl := uint8(buff[0])
		var wd []byte
		buff = make([]byte, wdl)
		var got uint8
		for got < wdl {
			n, err := con.Read(buff)
			if err != nil {
				panic(err)
			}
			wd = append(wd, buff[:n]...)
			got += uint8(n)
		}
		if newProc(string(wd)) {
			return []byte{123}, true
		}
		return []byte{124}, true
	},
	11: func(con net.Conn) ([]byte, bool) {
		buff := make([]byte, 1)
		var n int
		var err error
		for n < 1 {
			n, err = con.Read(buff)
			if err != nil {
				panic(err)
			}
		}
		tp := procs[int(buff[0])]
		err = tp.p.Process.Signal(syscall.SIGKILL)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return []byte{124}, true
		}
		for tp.state == 2 {
		}
		tp.state = 0
		return []byte{123}, true
	},
	21: func(_ net.Conn) ([]byte, bool) {
		return []byte{123}, true
	},
	22: func(con net.Conn) ([]byte, bool) {
		buff := make([]byte, 1)
		var n int
		var err error
		for n < 1 {
			n, err = con.Read(buff)
			if err != nil {
				panic(err)
			}
		}
		var scon net.Conn
		if runtime.GOOS == "windows" {
			scon, err = net.Dial("tcp", procs[int(buff[0])].controlAddr)
		} else {
			scon, err = net.Dial("unix", procs[int(buff[0])].controlAddr)
		}
		if err != nil {
			panic(err)
		}
		con.Write([]byte{123})
		scon.Write([]byte("getlogs"))
		_, err = io.Copy(con, scon)
		if err != nil {
			panic(err)
		}
		con.Close()
		return []byte{}, false
	},
	23: func(con net.Conn) ([]byte, bool) {
		buff := make([]byte, 1)
		var n int
		var err error
		for n < 1 {
			n, err = con.Read(buff)
			if err != nil {
				panic(err)
			}
		}
		getting := buff[0]
		buff = append([]byte{20}, jpsutil.Uint32ToBytes(uint32(len(procs[getting].stdout.value)))...)
		buff = append(buff, []byte(procs[getting].stdout.value)...)
		return buff, true
	},
	24: func(con net.Conn) ([]byte, bool) {
		buff := make([]byte, 1)
		var n int
		var err error
		for n < 1 {
			n, err = con.Read(buff)
			if err != nil {
				panic(err)
			}
		}
		getting := buff[0]
		buff = append([]byte{20}, jpsutil.Uint32ToBytes(uint32(len(procs[getting].stderr.value)))...)
		buff = append(buff, []byte(procs[getting].stderr.value)...)
		return buff, true
	},
}

func handler(conn net.Conn) {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Fprintln(os.Stderr, "\nHandler panicing!")
			fmt.Fprintln(os.Stderr, err)
			conn.Close()
		}
	}()
	buff := make([]byte, 1)
	var read int
	var err error
	var ok bool
	var toCall func(net.Conn) ([]byte, bool)
	var toWrite []byte
	var doWrite bool
	for {
		read, err = conn.Read(buff)
		if err != nil {
			if err == io.EOF {
				return
			}
			panic(err)
		}
		if read > 0 {
			toCall, ok = GotMessage[buff[0]]
			if ok {
				toWrite, doWrite = toCall(conn)
			} else {
				toWrite, doWrite = []byte{124}, true
			}
			if doWrite {
				_, err := conn.Write(toWrite)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

//Start starts the JOTPOT Server daemon
func Start() {
	var err error
	dir, err = ioutil.TempDir("", "jps-")
	if err != nil {
		panic(err)
	}
	server, err := net.Listen("tcp", ":50551")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := server.Accept()
		if err != nil {
			panic(err)
		}
		go handler(conn)
	}
}
