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
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

type handler struct{}

func (m *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	fmt.Println("Got request")

	//Dial server
	fmt.Println("Dialing server...")
	forward, err := net.Dial("tcp", ":80")
	panicIfErr(err)

	//Write headers etc. to server
	fmt.Println("Writing header to server...")
	err = req.Write(forward)
	panicIfErr(err)

	//Parse as response
	fmt.Println("Parsing server response...")
	forwardResp, err := http.ReadResponse(bufio.NewReader(forward), nil)
	panicIfErr(err)

	//Write headers etc. to client
	fmt.Println("Copying headers...")
	for key, val := range forwardResp.Header {
		//resp.Header().Set(key, val)
		resp.Header().Set(key, val[0])
		if len(val) > 1 {
			for _, val := range val {
				resp.Header().Add(key, val)
			}
		}
	}

	fmt.Println("Writing header to client...")
	resp.WriteHeader(forwardResp.StatusCode)
	panicIfErr(err)

	//Buf to store data in and n already set up
	fmt.Println("Setting stuff up...")
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
	http.ListenAndServe(":8081", &handler{})
}
