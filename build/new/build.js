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

const parseFlags = require("./flag-parser") ;

const flags = parseFlags() ;
if (!flags["-out"]) {
	throw new Error("-out option required") ;
} else if (typeof flags["-out"][0] !== "string") {
	throw new Error("-out option must be a string") ;
}
const out = flags["-out"][0] ;
if (flags["-in"] && typeof flags["-in"][0] !== "string") {
	throw new Error("-in option must be a string") ;
}
const src = (flags["-in"] || [process.cwd()])[0] ;
const goPath = (flags["-go"] || ["go"])[0] ;

const fs = require("fs") ;
const path = require("path") ;

let config ;
function loadConfig(path) {
	if (fs.existsSync(path)) {
		config = JSON.parse(fs.readFileSync(path).toString()) ;
		return true ;
	}
	return false ;
}
if (!loadConfig(path.join(src, "build", "buildconfig.json"))) {
	if (!loadConfig(path.join(src, "build", "new", "buildconfig.json"))) {
		throw new Error("Cannot find buildconfig.json") ;
	}
}

function createDir(...paths) {
	const p = path.join(...paths) ;
	if (!fs.existsSync(p)) {
		createDir(path.dirname(p)) ;
		fs.mkdirSync(p) ;
	} else {
		if (fs.statSync(p).isFile()) {
			throw new Error(`${p} is a file.`) ;
		}
	}
}

function copy(p1, p2) {
	fs.writeFileSync(p2, fs.readFileSync(p1)) ;
}

if (flags["-jsonly"]) {
	createDir(out) ;
	for (let copyer of config["jps-files"]) {
		copy(path.join(src, copyer), path.join(out, copyer)) ;
	}
	copy(path.join(src, "LICENSE"), path.join(out, "LICENSE")) ;
	process.exit(0) ;
}

createDir(out, "jps-main") ;

for (let copyer of config["jps-files"]) {
	copy(path.join(src, copyer), path.join(out, "jps-main", copyer)) ;
}
copy(path.join(src, "config.json"), path.join(out, "jps-main", "defaultConfig.json")) ;
copy(path.join(src, "errorTemp.jpt"), path.join(out, "jps-main", "defaultErrorTemp.jpt")) ;
copy(path.join(src, "sites", "default", "index.html"), path.join(out, "jps-main", "defaultIndex.html")) ;
copy(path.join(src, "LICENSE"), path.join(out, "LICENSE")) ;

const isWindows = (process.env.GOOS || process.platform).replace("win32", "windows") === "windows" ;
const cp = require("child_process") ;

let built = true ;
try {
	cp.execSync(`${goPath} build -o ${path.join(out, "jps" + (isWindows?".exe":""))} ${path.join(src, "jps", "jps-main.go")}`, {
		env: {
			"GOPATH": path.join(src, "jps"),
			"GOOS": process.env.GOOS||"",
			"GOARCH": process.env.GOARCH||"",
			"GOARM": process.env.GOARM||"",
			"GO386": process.env.GO386||"",
			"GOROOT": process.env.GOROOT||"",
			"PATH": process.env.PATH,
			"CC": process.env.GOARCH.indexOf("arm")===0?"arm-linux-gnueabi-gcc":""
		},
		stdio: flags["--hide-errors"]?"ignore":"inherit"
	}) ;
} catch (err) {
	built = false ;
	console.warn("!!! Failed to build main executable!") ;
}
if (built) {
	let origDir = process.cwd() ;
	process.chdir(out) ;
	fs.linkSync("./jps" + (isWindows?".exe":""), "jpsd" + (isWindows?".exe":"")) ;
	process.chdir(origDir) ;
}
try {
	cp.execSync(`${goPath} build -o ${path.join(out, "jpslb" + (isWindows?".exe":""))} ${path.join(src, "lb.go")}`, {
		env: {
			"GOPATH": path.join(src, "jps"),
			"GOOS": process.env.GOOS||"",
			"GOARCH": process.env.GOARCH||"",
			"GOARM": process.env.GOARM||"",
			"GO386": process.env.GO386||"",
			"GOROOT": process.env.GOROOT||"",
			"PATH":process.env.PATH,
			"CC": process.env.GOARCH.indexOf("arm")===0?"arm-linux-gnueabi-gcc":""
		},
		stdio: flags["--hide-errors"]?"ignore":"inherit"
	}) ;
} catch (err) {
	console.warn("!!! Failed to build load balencer executable!") ;
	built = false ;
}
if (!built) {
	process.exit(1) ;
}
