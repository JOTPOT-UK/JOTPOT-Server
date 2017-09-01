/*

	JOTPOT Server
	Version 26B-0

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
	"fmt"
	"jps"
	"jpsd"
	"os"
	"path/filepath"
	"strings"
)

func endHandler() {
	err := recover()
	if err != nil {
		fmt.Fprintln(os.Stderr, "\nProcess panicing!")
		fmt.Fprintln(os.Stderr, err)
	}
}

func main() {
	defer endHandler()
	exec, err := os.Executable()
	if err != nil {
		panic(err)
	}
	os.Args[0] = exec
	if strings.Index(filepath.Base(os.Args[0]), "jpsd") == 0 {
		jpsd.Start()
	} else if len(os.Args) < 2 {
		os.Args = append(os.Args, "run")
		jps.Go()
	} else {
		if os.Args[1] == "start-daemon" {
			jpsd.Start()
		} else {
			jps.Go()
		}
	}
}
