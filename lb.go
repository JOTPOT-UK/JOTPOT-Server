/*

	JOTPOT Server
	Version 25E

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

package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
)

const ignorePort bool = true

//Panic if err is not nil
func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

type form struct {
	regexp     *regexp.Regexp
	serveraddr string
}
type uncompiledform struct {
	regexp     string
	serveraddr string
}
type director struct {
	DefaultAddr string
	DefaultHost string
	Directions  map[string][]form
}
type uncompileddirector struct {
	DefaultAddr string
	DefaultHost string
	Directions  map[string][]uncompiledform
}

type handler struct {
	isHTTPS  bool
	director *director
}

//Returns the server address based on the directions from the director
func (t director) WhereIs(u *url.URL) string {
	//Get relivent directions
	var hostname string
	if ignorePort {
		hostname = u.Hostname()
	} else {
		hostname = u.Host
	}
	checks, ok := t.Directions[hostname]

	//Fallback to default host/default address
	if !ok {
		if t.DefaultHost == "" {
			return t.DefaultAddr
		}
		checks, ok = t.Directions[t.DefaultHost]
		if !ok {
			return t.DefaultAddr
		}
	}

	//Test for reg exp matches
	for _, exp := range checks {
		if exp.regexp.MatchString(u.Path) {
			return exp.serveraddr
		}
	}

	//Fallback to default address
	return t.DefaultAddr
}

//Compile uncompiled director
func (t uncompileddirector) Compile() (out director, err error) {
	var compiled *regexp.Regexp

	//Copy basic properties
	out.DefaultAddr = t.DefaultAddr
	out.DefaultHost = t.DefaultHost

	for host, tf := range t.Directions {
		//If it is a nil map, we cannot just add the key
		if out.Directions != nil {
			out.Directions[host] = []form{}
		} else {
			out.Directions = map[string][]form{host: []form{}}
		}

		//Compile all the rules
		for _, exp := range tf {
			compiled, err = regexp.Compile(exp.regexp)
			if err != nil {
				//Return with err if there was an err
				return
			}
			out.Directions[host] = append(out.Directions[host], form{compiled, exp.serveraddr})
		}
	}
	return
}

func (m *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	fmt.Println("Got request")

	//Add the host to the URL object for WhereIs
	req.URL.Host = req.Host

	//What server should we dial
	connectTo := m.director.WhereIs(req.URL)
	fmt.Println(connectTo)

	//Dial server
	forward, err := net.Dial("tcp", connectTo)
	panicIfErr(err)

	//Add load balancer headers
	req.Header["jp-source-ip"] = []string{req.RemoteAddr}
	if m.isHTTPS {
		req.Header["jp-source-secure"] = []string{"https"}
	} else {
		req.Header["jp-source-secure"] = []string{"http"}
	}

	//Write headers etc. to server
	err = req.Write(forward)
	panicIfErr(err)

	//Parse as response
	forwardResp, err := http.ReadResponse(bufio.NewReader(forward), nil)
	panicIfErr(err)

	//Write headers etc. to client
	for key, val := range forwardResp.Header {
		//resp.Header().Set(key, val)
		resp.Header().Set(key, val[0])
		if len(val) > 1 {
			for _, val := range val {
				resp.Header().Add(key, val)
			}
		}
	}

	resp.WriteHeader(forwardResp.StatusCode)
	panicIfErr(err)

	//Buf to store data in and n already set up
	buf := make([]byte, 1024)
	var n int
	var bClosed bool
	var fClosed bool

	//Body
	for {
		//Client -> Server
		fmt.Println("Client -> Server")
		n, err = req.Body.Read(buf)
		if err == nil {
			forward.Write(buf[:n])
		} else if err == io.EOF {
			bClosed = true
		}

		//Server -> Client
		fmt.Println("Server -> Client")
		n, err = forwardResp.Body.Read(buf)
		if err == nil {
			resp.Write(buf[:n])
		} else if err == io.EOF {
			fClosed = true
		}

		if bClosed && fClosed {
			fmt.Println("We are at the end!")
			forward.Close()
			return
		}
	}
}

func main() {
	//To be used later
	toCompile := uncompileddirector{
		//Default address and host
		"LIMBO",
		"LIMBO",
		//Simple test director
		map[string][]uncompiledform{
			"www.jotpot.co.uk": []uncompiledform{
				uncompiledform{".*", "192.168.1.11:80"},
			},
			"localhost": []uncompiledform{
				uncompiledform{".*", ":8080"},
			},
		},
	}

	//Compile it
	mainDirector, err := toCompile.Compile()
	panicIfErr(err)

	//Set up server
	http.ListenAndServe(":8081", &handler{false, &mainDirector})
}
