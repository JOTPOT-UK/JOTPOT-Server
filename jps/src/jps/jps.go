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

package jps

import (
	"fmt"
	"jpsutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
)

//Commands determine commands that the user can use
var Commands = map[string]func(){
	"startsync": func() {
		c := exec.Command(jpsutil.GetNodePath(), filepath.Join(path.Dir(os.Args[0]), "jps", "run"))
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
		buff = jpsutil.GetData(con, 2)
		if buff[0] != 9 {
			fmt.Println("\rSomehing went wrong while loading the server list.")
		} else {
			got := int(buff[1])
			fmt.Println("\rServers under this daemon:")
			for got > 0 {
				buff = jpsutil.GetData(con, 3)
				fmt.Print(string(buff[0]+48) + ": ")
				if buff[1] > 1 {
					fmt.Print("running")
				} else if buff[1] == 1 {
					fmt.Print("stopped")
				} else {
					fmt.Print("errored")
				}
				buff = jpsutil.GetData(con, int(buff[2]))
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
		buff = jpsutil.GetData(con, 5)
		if buff[0] != 20 {
			fmt.Println("\rSomehing went wrong while getting the logs.")
		} else {
			buff = jpsutil.GetData(con, int(jpsutil.Uint32FromBytes(buff[1:])))
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
		buff = jpsutil.GetData(con, 5)
		if buff[0] != 20 {
			fmt.Println("\rSomehing went wrong while getting the logs.")
		} else {
			buff = jpsutil.GetData(con, int(jpsutil.Uint32FromBytes(buff[1:])))
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
	"setup": func() {
		fmt.Println("Setting this directory up as a simple server...")
		srcDir := path.Dir(os.Args[0])
		dstDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		os.MkdirAll(filepath.Join(dstDir, "sites", "default"), 0664)
		err = jpsutil.CopyFile(filepath.Join(srcDir, "jps", "defaultConfig.json"), filepath.Join(dstDir, "config.json"))
		if err != nil {
			panic(err)
		}
		err = jpsutil.CopyFile(filepath.Join(srcDir, "jps", "defaultErrorTemp.jpt"), filepath.Join(dstDir, "errorTemp.jpt"))
		if err != nil {
			panic(err)
		}
		err = jpsutil.CopyFile(filepath.Join(srcDir, "jps", "defaultIndex.html"), filepath.Join(dstDir, "sites", "default", "index.html"))
		if err != nil {
			panic(err)
		}
		fmt.Println("Successfully set up :)")
	},
}

//StartSync runs the startsync command
func StartSync() {
	Commands["startsync"]()
}

//Go runs the utility
func Go() {
	if len(os.Args) < 2 {
		Commands["startsync"]()
	} else {
		toCall, ok := Commands[os.Args[1]]
		if ok {
			toCall()
		} else {
			fmt.Fprintln(os.Stderr, "Unknown command:", os.Args[1])
		}
	}
}