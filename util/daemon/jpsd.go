package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

type proc struct {
	sDir        string
	p           *exec.Cmd
	stdout      string
	stderr      string
	index       int
	running     byte
	writeContol *os.File
	readControl *os.File
}

func (c *proc) startRead() {
	stdout, err := c.p.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderr, err := c.p.StderrPipe()
	if err != nil {
		panic(err)
	}
	err = c.p.Start()
	if err != nil {
		panic(err)
	}
	buff := make([]byte, 1024)
	var n1 int
	var n2 int
	for {
		n1, err = stdout.Read(buff)
		if err != nil && err != io.EOF {
			panic(err)
		}
		c.stdout += string(buff[:n1])
		n2, err = stderr.Read(buff)
		if err != nil && err != io.EOF {
			panic(err)
		}
		c.stderr += string(buff[:n2])
		err = c.p.Process.Signal(syscall.Signal(0))
		if n1 == 0 && n2 == 0 && err != nil {
			break
		}
	}
	err = c.p.Wait()
	if err != nil {
		c.running = 0
	} else {
		c.running = 1
	}
}

var procs []*proc

func newProc(wd string) bool {
	for _, p := range procs {
		if p.sDir == wd && p.running == 2 {
			return false
		}
	}
	awd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	c := exec.Command(getNodePath(), filepath.Join(awd, filepath.Dir(os.Args[0]), "jps", "run"))
	c.Dir = wd
	controlIn, writeControl, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	readControl, controlOut, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	c.ExtraFiles = append(c.ExtraFiles, controlIn, controlOut)
	tp := proc{wd, c, "", "", len(procs), 2, writeControl, readControl}
	procs = append(procs, &tp)
	go tp.startRead()
	return true
}

func listProcs(_ net.Conn) ([]byte, bool) {
	out := []byte{9, byte(len(procs))}
	for _, p := range procs {
		out = append(out, byte(p.index), p.running, byte(len(p.sDir)))
		out = append(out, []byte(p.sDir)...)
	}
	return out, true
}

func getNodePath() string {
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

var gotMessage = map[byte]func(net.Conn) ([]byte, bool){
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
	21: func(_ net.Conn) ([]byte, bool) {
		return []byte{123}, true
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
		buff = append([]byte{20}, uint32tobytes(uint32(len(procs[getting].stdout)))...)
		buff = append(buff, []byte(procs[getting].stdout)...)
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
		buff = append([]byte{20}, uint32tobytes(uint32(len(procs[getting].stderr)))...)
		buff = append(buff, []byte(procs[getting].stderr)...)
		return buff, true
	},
	40: func(_ net.Conn) ([]byte, bool) {
		buff := make([]byte, 5)
		for {
			//n, err := procs[0].lpR.Read(buff)
			fmt.Println(n, string(buff))
			if err != nil {
				fmt.Println(err)
				break
			}
		}
		return []byte{123}, true
	},
}

func uint32tobytes(in uint32) []byte {
	return []byte{byte(in >> 24), byte((in >> 16) & 255), byte((in >> 8) & 255), byte(in & 255)}
}

func handler(conn net.Conn) {
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
			toCall, ok = gotMessage[buff[0]]
			if ok {
				toWrite, doWrite = toCall(conn)
			} else {
				toWrite, doWrite = []byte{124}, true
			}
			if doWrite {
				n, err := conn.Write(toWrite)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func main() {
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
