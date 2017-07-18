/*
	
	JOTPOT Server
	Version 25F
	
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

const url = require("url") ;

function parseHost(host, https) {
	let splithost = host.split(":") ;
	//If the host doesn't conatain the port, add 80/443
	if (splithost.length < 2) {
		return [`${host}:${https?443:80}`, host, https?443:80] ;
	} else {
		let portshouldbe = splithost.pop() ;
		if (parseInt(portshouldbe).toString() === portshouldbe) {
			return [host, splithost.join(":"), portshouldbe] ;
		} else {
			return [`${host}:${https?443:80}`, host, https?443:80] ;
		}
	}
}

class URL {
	constructor (req, defaultHost) {
		
		//Parse the URL
		let purl = url.parse(req.url, false) ;
		
		//Add http(s):// properties
		purl.protocol = `http${req.overHttps?"s":""}:` ;
		purl.slashes = true ;
		
		Object.defineProperty(this, "accualHost", {
			value: req.headers.host || defaultHost,
			enumerable: true,
			writable: false,
			configurable: false
		}) ;
		
		//Split host into hostname and port
		//	This isn't the same method as the Node.js url.parse
		let parsedHost = parseHost((req.headers.host || defaultHost).toLowerCase(), req.overHttps) ;
		purl.host = parsedHost[0] ;
		purl.hostname = parsedHost[1] ;
		purl.port = parsedHost[2] ;
		
		Object.defineProperty(this, "pathname", {get:_=>{
			return purl.pathname ;
		}, set:val=>{
			if (val.indexOf("/") !== 0) {
				throw new Error("pathname must begin with a '/'") ;
			}
			purl.pathname = val ;
			purl.path = val + purl.search ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "path", {get:_=>{
			return purl.path ;
		}, set:val=>{
			if (val.indexOf("/") !== 0) {
				throw new Error("path must begin with a '/'") ;
			}
			purl.path = val ;
			let npurl = url.parse(val) ;
			purl.pathname = npurl.pathname ;
			purl.search = npurl.search ;
			purl.query = npurl.query ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "search", {get:_=>{
			return purl.search ;
		}, set:val=>{
			if (val.indexOf("?") !== 0) {
				throw new Error("search must begin with a '?'") ;
			}
			purl.search = val ;
			purl.query = val.substring(1,val.length) ;
			purl.path = purl.pathname + purl.search ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "query", {get:_=>{
			return purl.query ;
		}, set:val=>{
			purl.query = val ;
			purl.search = "?" + val ;
			purl.path = purl.pathname + purl.search ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "hash", {get:_=>{
			return purl.hash ;
		}, set:val=>{
			if (val.indexOf("#") !== 0) {
				throw new Error("hash must begin with a '#'") ;
			}
			purl.hash = val ;
		}, enumerable:true, configurable:false}) ;
		
		const validProtocols = ["http:", "https:"] ;
		Object.defineProperty(this, "accualProtocol", {
			configurable: false,
			enumerable: true,
			value: purl.protocol,
			writable: false
		}) ;
		Object.defineProperty(this, "accualAccualProtocol", {
			configurable: false,
			enumerable: true,
			value: `http${req.secureToServer?"s":""}:`,
			writable: false
		}) ;
		Object.defineProperty(this, "protocol", {get:_=>{
			return purl.protocol ;
		}, set:val=>{
			if (validProtocols.indexOf(val) === -1) {
				console.warn("Protocol has been changed to", val, "this is not a valid web protocol (http: or https:)") ;
			}
			purl.protocol = val ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "slashes", {
			configurable: false,
			enumerable: true,
			value: true,
			writable: false
		}) ;
		
		Object.defineProperty(this, "auth", {get:_=>{
			return purl.auth ;
		}, set:val=>{
			purl.auth = val ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "href", {get:_=>{
			return url.format(purl) ;
		}, set:val=>{
			purl = url.parse(val) ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "hrefnoport", {get:_=>{
			return url.format(purl).replace(purl.host, purl.hostname) ;
		}, set:val=>{
			purl = url.parse(val) ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "location", {get:_=>{
			if ((purl.port === 80 && purl.protocol === "http:") || purl.port === 443 && purl.protocol === "https:") {
				return url.format(purl).replace(purl.host, purl.hostname) ;
			} else {
				return url.format(purl) ;
			}
		}, set:val=>{
			purl = url.parse(val) ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "host", {get:_=>{
			return purl.host ;
		}, set:val=>{
			parsedHost = parseHost(val, purl.protocol==="https:") ;
			purl.host = parsedHost[0] ;
			purl.hostname = parsedHost[1] ;
			purl.port = parsedHost[2] ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "hostname", {get:_=>{
			return purl.hostname ;
		}, set:val=>{
			purl.hostname = val ;
			purl.host = `${val}:${purl.port}` ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "port", {get:_=>{
			return purl.port ;
		}, set:val=>{
			purl.port = val ;
			purl.host = `${purl.hostname}:${val}` ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "value", {get:_=>{
			return purl.host + purl.pathname ;
		}, set:val=>{
			val = val.split("/") ;
			purl.host = val.shift() ;
			purl.pathname = "/" + val.join("/") ;
			purl.path = purl.pathname + purl.search ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "fullvalue", {get:_=>{
			return purl.host + purl.path ;
		}, set:val=>{
			val = val.split("/") ;
			purl.host = val.shift() ;
			purl.path = "/" + val.join("/") ;
			let npurl = url.parse(val) ;
			purl.pathname = npurl.pathname ;
			purl.search = npurl.search ;
			purl.query = npurl.query ;
		}, enumerable:true, configurable:false}) ;
		
		Object.seal(this) ;
		
	}
	valueOf() {
		return this.value ;
	}
	toString() {
		return this.value ;
	}
	toJSON() {
		return this.href ;
	}
	static toDir(v) {
		return v.replace(/:/g,";") ;
	}
}

module.exports = URL ;

//req.host, req.accualHost, req.port, req.fullurl