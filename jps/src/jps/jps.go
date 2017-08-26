/*

	JOTPOT Server
	Version 26A-1

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
	"io"
	"jpsutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
)

func generateServiceFile(dir string) string {
	return `[Unit]
Description=JOTPOT Server Daemon
Wants=network-online.target
After=network.target network-online.target
ConditionPathExists=` + dir + `

[Service]
ExecStart=` + filepath.Join(dir, "jps") + ` start-daemon

[Install]
WantedBy=multi-user.target
Alias=jpsd.service`
}

//Commands determine commands that the user can use
var Commands = map[string]func(){
	"run": func() {
		//                                                       Module path........................................... , User args...........................
		c := exec.Command(jpsutil.GetNodePath(), append([]string{filepath.Join(filepath.Dir(os.Args[0]), "jps-main", "run")}, jpsutil.Args(os.Args[2:]).ToServer()...)...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Run()
	},
	"start": func() {
		fmt.Println("Starting")
		//Dial daemon
		con, err := net.Dial("tcp", "127.5.5.5:50551")
		if err != nil {
			panic(err)
		}

		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		//Create buffer to send
		//                           10, Length of wd.,  [wd..........], How many args.......
		buff := append(append([]byte{10, byte(len(wd))}, []byte(wd)...), byte(len(os.Args)-2))
		for _, arg := range os.Args[2:] {
			//                         Length of arg.., this arg......
			buff = append(append(buff, byte(len(arg))), []byte(arg)...)
		}

		//Write it
		_, err = con.Write(buff)
		if err != nil {
			panic(err)
		}

		//Read return
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
		fmt.Println("Stopping server...")
		con, err := net.Dial("tcp", "127.5.5.5:50551")
		if err != nil {
			panic(err)
		}
		buff := []byte{11, byte(toGet)}
		_, err = con.Write(buff)
		if err != nil {
			panic(err)
		}
		buff = []byte{0}
		n := 0
		for n < 1 {
			n, err = con.Read(buff)
			if err != nil {
				panic(err)
			}
		}
		if buff[0] == 126 {
			fmt.Println("Process", toGet, "doesn't exist!")
		} else if buff[0] != 123 {
			fmt.Println("Somehing went wrong while stopping the server.")
		} else {
			fmt.Println("Server stopped.")
		}
		con.Close()
	},
	"restart": func() {
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
		fmt.Println("Restarting server...")
		con, err := net.Dial("tcp", "127.5.5.5:50551")
		if err != nil {
			panic(err)
		}
		buff := []byte{12, byte(toGet)}
		_, err = con.Write(buff)
		if err != nil {
			panic(err)
		}
		buff = []byte{0}
		n := 0
		for n < 1 {
			n, err = con.Read(buff)
			if err != nil {
				panic(err)
			}
		}
		if buff[0] == 126 {
			fmt.Println("Process", toGet, "doesn't exist!")
		} else if buff[0] != 123 {
			fmt.Println("Somehing went wrong while stopping the server.")
		} else {
			fmt.Println("Server restarted.")
		}
		con.Close()
	},
	"list": func() {
		fmt.Print("Loading list...")
		con, err := net.Dial("tcp", "127.5.5.5:50551")
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
				fmt.Print(strconv.Itoa(int(buff[0])) + ": ")
				if buff[1] > 1 {
					fmt.Print("running")
				} else if buff[1] == 0 {
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
		con, err := net.Dial("tcp", "127.5.5.5:50551")
		if err != nil {
			panic(err)
		}
		buff := []byte{23, byte(toGet)}
		_, err = con.Write(buff)
		if err != nil {
			panic(err)
		}
		buff = jpsutil.GetData(con, 5)
		if buff[0] == 126 {
			fmt.Println("Process", toGet, "doesn't exist!")
		} else if buff[0] != 20 {
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
		con, err := net.Dial("tcp", "127.5.5.5:50551")
		if err != nil {
			panic(err)
		}
		buff := []byte{24, byte(toGet)}
		_, err = con.Write(buff)
		if err != nil {
			panic(err)
		}
		buff = jpsutil.GetData(con, 5)
		if buff[0] == 126 {
			fmt.Println("Process", toGet, "doesn't exist!")
		} else if buff[0] != 20 {
			fmt.Println("\rSomehing went wrong while getting the logs.")
		} else {
			buff = jpsutil.GetData(con, int(jpsutil.Uint32FromBytes(buff[1:])))
			fmt.Println("\rError logs from server:")
			fmt.Println(string(buff))
		}
		con.Close()
	},
	"logs": func() {
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
		con, err := net.Dial("tcp", "127.5.5.5:50551")
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
		if buff[0] == 126 {
			fmt.Println("Process", toGet, "doesn't exist!")
			con.Close()
			return
		} else if buff[0] != 123 {
			panic("123 not returned by read request")
		}
		buff = make([]byte, 1024)
		fmt.Print("\n")
		for {
			n, err = con.Read(buff)
			if err != nil && err != io.EOF {
				panic(err)
			}
			fmt.Print(string(buff[:n]))
			if err == io.EOF {
				break
			}
		}
		fmt.Print("\n\n")
	},
	"setup": func() {
		fmt.Println("Setting this directory up as a simple server...")
		srcDir := path.Dir(os.Args[0])
		dstDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		os.MkdirAll(filepath.Join(dstDir, "sites", "default"), 0664)
		err = jpsutil.CopyFile(filepath.Join(srcDir, "jps-main", "defaultConfig.json"), filepath.Join(dstDir, "config.json"))
		if err != nil {
			panic(err)
		}
		err = jpsutil.CopyFile(filepath.Join(srcDir, "jps-main", "defaultErrorTemp.jpt"), filepath.Join(dstDir, "errorTemp.jpt"))
		if err != nil {
			panic(err)
		}
		err = jpsutil.CopyFile(filepath.Join(srcDir, "jps-main", "defaultIndex.html"), filepath.Join(dstDir, "sites", "default", "index.html"))
		if err != nil {
			panic(err)
		}
		fmt.Println("Successfully set up :)")
	},
	"make-unit-file": func() {
		fmt.Print(generateServiceFile(path.Dir(os.Args[0])))
	},
}

//StartSync runs the startsync command
func StartSync() {
	Commands["run"]()
}

//Go runs the utility
func Go() {
	if len(os.Args) < 2 {
		Commands["run"]()
	} else {
		toCall, ok := Commands[os.Args[1]]
		if ok {
			toCall()
		} else {
			fmt.Fprintln(os.Stderr, "Unknown command:", os.Args[1])
		}
	}
}
