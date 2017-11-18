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

const url = require("url") ;

const validProtocols = ["http:", "https:", "ws:"] ;

function parseHost(host, https) {
	let splithost = host.split(":") ;
	//If the host doesn't conatain the port, add 80/443
	if (splithost.length < 2) {
		return [`${host}:${https?443:80}`, host, https?443:80] ;
	} 
	let portshouldbe = splithost.pop() ;
	if (parseInt(portshouldbe, 10).toString() === portshouldbe) {
		return [host, splithost.join(":"), portshouldbe] ;
	} 
	return [`${host}:${https?443:80}`, host, https?443:80] ;
		
	
}

class URL {
	constructor (req, defaultHost, slashes=true) {
		
		let hostLocked = false ;
		
		//Parse the URL
		let purl = url.parse(req.url, false) ;
		
		//Add http(s):// properties
		purl.protocol = `http${req.overHttps?"s":""}:` ;
		purl.slashes = slashes ;
		
		//Split host into hostname and port
		//	This isn't the same method as the Node.js url.parse
		let parsedHost = parseHost((req.headers.host || defaultHost).toLowerCase(), req.overHttps) ;
		purl.host = parsedHost[0] ;
		purl.hostname = parsedHost[1] ;
		purl.port = parsedHost[2] ;
		
		Object.defineProperty(this, "accualURL", {
			configurable: false,
			enumerable: true,
			writable: false,
			value: url.format(purl)
		}) ;
		
		Object.defineProperty(this, "pathname", {get:()=>{
			return purl.pathname ;
		}, set:val=>{
			if (val.indexOf("/") !== 0) {
				throw new Error("pathname must begin with a '/'") ;
			}
			purl.pathname = val ;
			purl.path = val + (purl.search || "") ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "path", {get:()=>{
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
		
		Object.defineProperty(this, "search", {get:()=>{
			return purl.search ;
		}, set:val=>{
			if (val.indexOf("?") !== 0) {
				throw new Error("search must begin with a '?'") ;
			}
			purl.search = val ;
			purl.query = val.substring(1,val.length) ;
			purl.path = purl.pathname + (purl.search || "") ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "query", {get:()=>{
			return purl.query ;
		}, set:val=>{
			purl.query = val ;
			purl.search = "?" + val ;
			purl.path = purl.pathname + (purl.search || "") ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "hash", {get:()=>{
			return purl.hash ;
		}, set:val=>{
			if (val.indexOf("#") !== 0) {
				throw new Error("hash must begin with a '#'") ;
			}
			purl.hash = val ;
		}, enumerable:true, configurable:false}) ;
		
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
		Object.defineProperty(this, "protocol", {get:()=>{
			return purl.protocol ;
		}, set:val=>{
			if (validProtocols.indexOf(val) === -1) {
				console.warn("Protocol has been changed to", val, "this is not a valid web protocol (http: or https:)") ;
			}
			purl.protocol = val ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "slashes", {get:()=>{
			return purl.slashes ;
		}, set:val=>{
			if (typeof slashes !== "boolean") {
				console.warn("slashes must be a boolean") ;
			}
			purl.slashes = val ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "href", {get:()=>{
			return url.format(purl) ;
		}, set:val=>{
			if (!hostLocked) {
				purl = url.parse(val) ;
			} else {
				const origHost = purl.host ;
				purl = url.parse(val) ;
				if (purl.host !== origHost) {
					console.warn("An attempt was made to change the host be setting the href, however the host is locked so was not changed.") ;
					this.host = origHost ;
				}
			}
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "hrefnoport", {get:()=>{
			return url.format(purl).replace(purl.host, purl.hostname) ;
		}, set:val=>{
			const port = purl.port ;
			if (!hostLocked) {
				purl = url.parse(val) ;
				purl.hostname = purl.host ;
				purl.port = port ;
				purl.host = purl.hostname + ":" + port ;
			} else {
				const origHost = purl.host ;
				purl = url.parse(val) ;
				purl.hostname = purl.host ;
				purl.port = port ;
				purl.host = purl.hostname + ":" + port ;
				if (origHost !== purl.host) {
					console.warn("An attempt was made to change the host be setting the hrefnoport, however the host is locked so was not changed.") ;
					this.host = origHost ;
				}
			}
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "location", {get:()=>{
			if ((purl.port === 80 && purl.protocol === "http:") || (purl.port === 443 && purl.protocol === "https:")) {
				return url.format(purl).replace(purl.host, purl.hostname) ;
			} 
			return url.format(purl) ;
			
		}, set:val=>{
			purl = url.parse(val) ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "host", {get:()=>{
			return purl.host ;
		}, set:val=>{
			if (!hostLocked) {
				parsedHost = parseHost(val, purl.protocol==="https:") ;
				purl.host = parsedHost[0] ;
				purl.hostname = parsedHost[1] ;
				purl.port = parsedHost[2] ;
			} else {
				console.warn("URL host change attempted, although host is locked.") ;
			}
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "hostname", {get:()=>{
			return purl.hostname ;
		}, set:val=>{
			if (!hostLocked) {
				purl.hostname = val ;
				purl.host = `${val}:${purl.port}` ;
			} else {
				console.warn("URL hostname change attempted, although host is locked.") ;
			}
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "port", {get:()=>{
			return purl.port ;
		}, set:val=>{
			if (!hostLocked) {
				purl.port = val ;
				purl.host = `${purl.hostname}:${val}` ;
			} else {
				console.warn("URL port change attempted, although host is locked.") ;
			}
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "value", {get:()=>{
			return purl.host + purl.pathname ;
		}, set:val=>{
			val = val.split("/") ;
			this.host = val.shift() ;
			purl.pathname = "/" + val.join("/") ;
			purl.path = purl.pathname + (purl.search || "") ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "fullvalue", {get:()=>{
			return purl.host + purl.path ;
		}, set:val=>{
			val = val.split("/") ;
			this.host = val.shift() ;
			purl.path = "/" + val.join("/") ;
			let npurl = url.parse(val) ;
			purl.pathname = npurl.pathname ;
			purl.search = npurl.search ;
			purl.query = npurl.query ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "valuenoport", {get:()=>{
			return purl.hostname + purl.pathname ;
		}, set:val=>{
			val = val.split("/") ;
			this.hostname = val.shift() ;
			purl.pathname = "/" + val.join("/") ;
			purl.path = purl.pathname + (purl.search || "") ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "fullvaluenoport", {get:()=>{
			return purl.hostname + purl.path ;
		}, set:val=>{
			val = val.split("/") ;
			this.hostname = val.shift() ;
			purl.path = "/" + val.join("/") ;
			let npurl = url.parse(val) ;
			purl.pathname = npurl.pathname ;
			purl.search = npurl.search ;
			purl.query = npurl.query ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "origin", {get:()=>{
			return purl.protocol + (purl.slashes?"//":"") + String(purl.host) ;
		}, set:val=>{
			let splitter = ":" ;
			if (val.indexOf(":") === val.indexOf("://")) {
				splitter = "://" ;
				purl.slashes = true ;
			} else {
				purl.slashes = false ;
			}
			purl.protocol = val.substring(0, val.indexOf(splitter)) + ":" ;
			if (validProtocols.indexOf(purl.protocol) === -1) {
				console.warn("Protocol has been changed to", val, "this is not a valid web protocol (http: or https:)") ;
			}
			this.host = val.substring(val.indexOf(splitter), val.length) ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "username", {get:()=>{
			if (!purl.auth) {
				return null ;
			}
			return purl.auth.split(":")[0] ;
		}, set:val=>{
			if (val.indexOf(":") !== -1) {
				throw new Error("username/password must not contain ':'") ;
			}
			purl.auth = val + ":" + this.password ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "password", {get:()=>{
			if (!purl.auth) {
				return null ;
			}
			return purl.auth.split(":")[1] ;
		}, set:val=>{
			if (val.indexOf(":") !== -1) {
				throw new Error("username/password must not contain ':'") ;
			}
			purl.auth = this.username + ":" + val ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "auth", {get:()=>{
			return purl.auth ;
		}, set:val=>{
			if (val.match(/:/g).length !== 2) {
				throw new Error("auth must have 1 ':'") ;
			}
			purl.auth = val ;
		}, enumerable:true, configurable:false}) ;
		
		Object.defineProperty(this, "lockHost", {value:()=>{
			if (hostLocked) {
				return () => {} ;
			} 
			hostLocked = true ;
			return () => hostLocked = false ;
			
		}, enumerable: false, configurable: false, writable: false}) ;
		
		this.valueOf = this.valueOf.bind(this) ;
		this.toString = this.toString.bind(this) ;
		this.toJSON = this.toJSON.bind(this) ;
		this.normalize = this.normalize.bind(this) ;
		
		Object.seal(this) ;
		
	}
	valueOf() {
		return this.value ;
	}
	toString() {
		return this.value ;
	}
	toJSON() {
		return this.location ;
	}
	normalize() {
		do {
			this.pathname = this.pathname.replace(/\.\./g, "").replace(/\/\//g, "/") ;
		} while (this.pathname.indexOf("//") !== -1) ;
		while (this.pathname.length > 1 && this.pathname.indexOf("/") === this.pathname.length - 1) {
			this.pathname = this.pathname.substring(0, this.pathname.length - 1) ;
		}
	}
	static toDir(v) {
		return v.replace(/:/g, ';') ;
	}
}

function createURL(opts) {
	return new URL({
		url: opts.path || "/",
		overHttps: opts.https || opts.overHttps || opts.overHTTPS || opts.secure || false,
		secureToServer: opts.https || opts.overHttps || opts.overHTTPS || opts.secure || false,
		headers: {
			host: opts.host || module.exports.defaultHost || "default:0"
		}
	}, module.exports.defaultHost, opts.slashes || true) ;
}

function createURLFromString(u) {
	let pu = url.parse(u) ;
	return new URL({
		url: pu.path || "/",
		overHttps: pu.protocol === "https:",
		secureToServer: pu.protocol === "https:",
		headers: {
			host: pu.host || module.exports.defaultHost || "default:0"
		}
	}, module.exports.defaultHost, pu.slashes || false) ;
}

module.exports = {
	URL,
	createURL,
	createURLFromString,
	defaultHost: "default:0"
} ;
