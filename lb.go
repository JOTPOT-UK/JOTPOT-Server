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

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
)

var ignorePort bool

//Panic if err is not nil
func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

type rule struct {
	regexp     *regexp.Regexp
	serveraddr []string
}
type uncompiledrule struct {
	regexp     string
	serveraddr []string
}
type director struct {
	DefaultAddr string
	DefaultHost string
	Directions  map[string][]rule
	Servers     map[string][2]uint64
}
type uncompileddirector struct {
	DefaultAddr string
	DefaultHost string
	Directions  map[string][]uncompiledrule
}

type handler struct {
	IsHTTPS  bool
	Director *director
}

type lowestLoad struct {
	val  uint64
	addr string
}

type ruleset map[string][]uncompiledrule

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

			//If there is only 1 server, it must be the one with the lowest load
			if len(exp.serveraddr) == 1 {
				return exp.serveraddr[0]
			}

			//Now we need to return tu server with tu lowest load
			ll := lowestLoad{1<<64 - 1, t.DefaultAddr}
			for _, tuServer := range exp.serveraddr {
				load := t.Servers[tuServer][0]
				if load < ll.val {
					ll = lowestLoad{load, tuServer}
				}
			}
			return ll.addr

		}
	}

	//Fallback to default address
	return t.DefaultAddr
}

//Compile uncompiled director
func (t uncompileddirector) Compile(serverConfig map[string]interface{}) (out director, err error) {
	var compiled *regexp.Regexp

	//Copy basic properties
	out.DefaultAddr = t.DefaultAddr
	out.DefaultHost = t.DefaultHost

	//Go through the hosts and rulesets in the uncompiled one
	for host, tf := range t.Directions {
		//If it is a nil map, we cannot just add the key
		if out.Directions != nil {
			out.Directions[host] = []rule{}
		} else {
			out.Directions = map[string][]rule{host: []rule{}}
		}

		//Compile all the rules and set up the server loads
		for _, tuRule := range tf {

			for _, tuServer := range tuRule.serveraddr {
				_, ok := out.Servers[tuServer]
				if !ok {
					capI, ok := serverConfig[tuServer]
					var cap uint64 = 50
					if ok {
						cap = uint64(capI.(float64))
					}
					if out.Servers != nil {
						out.Servers[tuServer] = [2]uint64{0, cap}
					} else {
						out.Servers = map[string][2]uint64{tuServer: {0, cap}}
					}
				}
			}

			//Compile the regexp
			compiled, err = regexp.Compile(tuRule.regexp)
			if err != nil {
				//Return with err if there was an err
				return
			}

			//Append it
			out.Directions[host] = append(out.Directions[host], rule{compiled, tuRule.serveraddr})
		}
	}
	return
}

func (m *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

	//Add the host to the URL object for WhereIs
	req.URL.Host = req.Host

	//What server should we dial
	connectTo := m.Director.WhereIs(req.URL)
	m.Director.Servers[connectTo] = [2]uint64{m.Director.Servers[connectTo][0] + m.Director.Servers[connectTo][1], m.Director.Servers[connectTo][1]}

	//Dial server
	forward, err := net.Dial("tcp", connectTo)
	panicIfErr(err)

	//Add load balancer headers
	req.Header["jp-source-ip"] = []string{req.RemoteAddr}
	if m.IsHTTPS {
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
		n, err = req.Body.Read(buf)
		if n > 0 {
			forward.Write(buf[:n])
		}
		if err != nil {
			if err == io.EOF {
				bClosed = true
			} else {
				panic(err)
			}
		}

		//Server -> Client
		n, err = forwardResp.Body.Read(buf)
		if n > 0 {
			resp.Write(buf[:n])
		}
		if err != nil {
			if err == io.EOF {
				fClosed = true
			} else {
				panic(err)
			}
		}

		if bClosed && fClosed {
			forward.Close()
			m.Director.Servers[connectTo] = [2]uint64{m.Director.Servers[connectTo][0] - m.Director.Servers[connectTo][1], m.Director.Servers[connectTo][1]}
			return
		}
	}

}

func getConfig() (map[string]interface{}, error) {

	//Load file
	file, err := ioutil.ReadFile("lbconfig.json")
	if err != nil {
		return nil, err
	}

	//Parse and return file
	var data interface{}
	err = json.Unmarshal(file, &data)
	if err != nil {
		return nil, err
	}
	return data.(map[string]interface{}), nil
}

func main() {

	fmt.Println("JOTPTO Server Load Balencer is now loading...")

	//Get config file
	conf, err := getConfig()
	panicIfErr(err)

	//Load defaults and set up empty ruleset
	toCompile := uncompileddirector{conf["defaultAddr"].(string), conf["defaultHost"].(string), ruleset{}}

	//Fill the directions
	for key, val := range conf["directions"].(map[string]interface{}) {
		if toCompile.Directions != nil {
			toCompile.Directions[key] = []uncompiledrule{}
		} else {
			toCompile.Directions = ruleset{key: []uncompiledrule{}}
		}
		vals := val.(map[string]interface{})
		for k2, v2 := range vals {
			toCompile.Directions[key] = append(toCompile.Directions[key], uncompiledrule{k2, []string{}})
			for _, v3 := range v2.([]interface{}) {
				toCompile.Directions[key][len(toCompile.Directions[key])-1].serveraddr = append(toCompile.Directions[key][len(toCompile.Directions[key])-1].serveraddr, v3.(string))
			}
		}
	}

	//Other settings
	ignorePort = conf["ignorePort"].(bool)

	//Compile it
	mainDirector, err := toCompile.Compile(conf["serverProperties"].(map[string]interface{}))
	panicIfErr(err)

	//Set up server
	fmt.Println("Ready to go!")
	err = http.ListenAndServe(conf["listenOn"].(string), &handler{false, &mainDirector})
	panicIfErr(err)

}
