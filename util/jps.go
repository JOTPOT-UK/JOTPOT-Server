package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
)

func getData(con net.Conn, toGet int) (out []byte) {
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
		}
		got += n
		out = append(out, buff[:n]...)
	}
	return
}

var commands = map[string]func(){
	"startsync": func() {
		c := exec.Command(getNodePath(), filepath.Join(path.Dir(os.Args[0]), "jps", "run"))
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Run()
	},
	"start": func() {
		fmt.Println("Starting")
		con, err := net.Dial("tcp", ":50551")
		if err != nil {
			panic(err)
		}
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		buff := append([]byte{10, byte(len(wd))}, []byte(wd)...)
		_, err = con.Write(buff)
		if err != nil {
			panic(err)
		}
		buff = make([]byte, 1)
		var n int
		for n < 1 {
			n, err = con.Read(buff)
			if err != nil {
				panic(err)
			}
		}
		if buff[0] == 123 {
			fmt.Println("Server started succesfully.")
		} else {
			fmt.Println("Server failed to start.")
		}
		con.Close()
	},
	"stop": func() {
		fmt.Println("Stopping")
	},
	"node": func() {
		fmt.Println("nodee stuff")
	},
	"list": func() {
		fmt.Print("Loading list...")
		con, err := net.Dial("tcp", ":50551")
		if err != nil {
			panic(err)
		}
		buff := []byte{9}
		_, err = con.Write(buff)
		if err != nil {
			panic(err)
		}
		buff = getData(con, 2)
		if buff[0] != 9 {
			fmt.Println("\rSomehing went wrong while loading the server list.")
		} else {
			got := int(buff[1])
			fmt.Println("\rServers under this daemon:")
			for got > 0 {
				buff = getData(con, 3)
				fmt.Print(string(buff[0]+48) + ": ")
				if buff[1] > 1 {
					fmt.Print("running")
				} else if buff[1] == 1 {
					fmt.Print("stopped")
				} else {
					fmt.Print("errored")
				}
				buff = getData(con, int(buff[2]))
				fmt.Println(" (" + string(buff) + ")")
				got--
			}
		}
		con.Close()
	},
	"info": func() {
		if len(os.Args) < 3 {
			fmt.Println("Second argument must be an int")
			os.Exit(1)
			return
		}
		toGet, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Second argument must be an int")
			os.Exit(1)
			return
		}
		fmt.Print("Getting logs...")
		con, err := net.Dial("tcp", ":50551")
		if err != nil {
			panic(err)
		}
		buff := []byte{23, byte(toGet)}
		_, err = con.Write(buff)
		if err != nil {
			panic(err)
		}
		buff = getData(con, 5)
		if buff[0] != 20 {
			fmt.Println("\rSomehing went wrong while getting the logs.")
		} else {
			buff = getData(con, int(uint32frombytes(buff[1:])))
			fmt.Println("\rInfo logs from server:")
			fmt.Println(string(buff))
		}
		con.Close()
	},
	"error": func() {
		if len(os.Args) < 3 {
			fmt.Println("Second argument must be an int")
			os.Exit(1)
			return
		}
		toGet, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Second argument must be an int")
			os.Exit(1)
			return
		}
		fmt.Print("Getting logs...")
		con, err := net.Dial("tcp", ":50551")
		if err != nil {
			panic(err)
		}
		buff := []byte{24, byte(toGet)}
		_, err = con.Write(buff)
		if err != nil {
			panic(err)
		}
		buff = getData(con, 5)
		if buff[0] != 20 {
			fmt.Println("\rSomehing went wrong while getting the logs.")
		} else {
			buff = getData(con, int(uint32frombytes(buff[1:])))
			fmt.Println("\rError logs from server:")
			fmt.Println(string(buff))
		}
		con.Close()
	},
	"read": func() {
		if len(os.Args) < 3 {
			fmt.Println("Second argument must be an int")
			os.Exit(1)
			return
		}
		toGet, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Second argument must be an int")
			os.Exit(1)
			return
		}
		fmt.Print("Sending read request...")
		con, err := net.Dial("tcp", ":50551")
		if err != nil {
			panic(err)
		}
		buff := []byte{22, byte(toGet)}
		_, err = con.Write(buff)
		if err != nil {
			panic(err)
		}
		var n int
		for n < 1 {
			n, err = con.Read(buff)
			if err != nil {
				panic(err)
			}
		}
		if buff[0] != 123 {
			panic("123 not returned by read request")
		}
		buff = make([]byte, 1024)
		for {
			n, err = con.Read(buff)
			if err != nil {
				panic(err)
			}
			fmt.Print(string(buff[:n]))
		}
	},
}

func uint32frombytes(in []byte) uint32 {
	return (uint32(in[0]) << 24) + (uint32(in[1]) << 16) + (uint32(in[2]) << 8) + uint32(in[3])
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

func main() {
	if len(os.Args) < 2 {
		commands["startsync"]()
	} else {
		toCall, ok := commands[os.Args[1]]
		if ok {
			toCall()
		} else {
			fmt.Fprintln(os.Stderr, "Unknown command:", os.Args[1])
		}
	}
}
