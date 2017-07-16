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

let readable = require("stream").Readable ;

class Data extends readable {
	
	constructor (name,headers) {
		
		super() ;
		this.name = name ;
		this.headers = headers ;
		this.setup = false ;
		
	}
	
	_read(s) {
		
		if (!this.setup && this.isPaused()) {
			
			this.setup = true ;
			this.resume() ;
			
		}
		
	}
	
}

module.exports = (req,cb) => new Promise((resolve,reject)=>{
	
	if (typeof req.headers["content-type"] === "undefined") {
		
		reject("No content type header.") ;
		return ;
		
	}
	const types = req.headers["content-type"].split("; ") ;
	if (types[0] !== "multipart/form-data") {
		
		reject([0,"Not multipart/form-data"]) ;
		return ;
		
	}
	let boundary = "" ;
	for (let doing = 1 ; doing < types.length ; doing++) {
		
		if (types[doing].split("=")[0] === "boundary") {
			
			boundary = `--${types[doing].split("=")[1]}` ;
			break ;
			
		}
		
	}
	if (!boundary) {
		
		reject([1,"No boundary"]) ;
		return ;
		
	}
	let boundaryB = Buffer.from(boundary) ;
	let data = Buffer.alloc(0) ;
	let dataString = new String() ;
	let endReady = false ;
	let reqEnded = false ;
	let output = new Object() ;
	let stage = 0 ;
	let fields = new Object() ;
	let newLineMade = false ;
	let currentHeaders = new Object() ;
	let currentDataPipe = null ;
	const parseTick =_=> {
		
		if (dataString.indexOf(boundary) === 0 && stage === 0) {
			
			dataString = dataString.substring(boundary.length,dataString.length) ;
			data = data.slice(boundaryB.length,data.length) ;
			stage++ ;
			newLineMade = false ;
			return parseTick() ;
			
		}
		
		else if (stage === 1) {
			
			do {
				
				if (!dataString) {
					
					return false ;
					
				}
				let thisBit = dataString.split("\r\n")[0] ;
				dataString = dataString.substring(dataString.indexOf("\r\n")+2,dataString.length) ;
				data = data.slice(data.indexOf(Buffer.from("\r\n"))+2,data.length) ;
				if (thisBit === "--") {
					
					return true ;
					
				}
				else if (thisBit === "") {
					
					if (!newLineMade) {
						
						newLineMade = true ;
						continue ;
						
					}
					stage++ ;
					let name = null ;
					if (typeof currentHeaders["content-disposition"] !== "undefined") {
						
						name = currentHeaders["content-disposition"].match(/; name="(\\"|[^"])*"/g)[0]
																	.substring(8,
																			   currentHeaders["content-disposition"].match(/; name="(\\"|[^"])*"/g)[0].length-1) ;
						
					}
					currentDataPipe = new Data(name,currentHeaders) ;
					cb(currentDataPipe) ;
					return parseTick() ;
					
				}
				else {
					
					thisBit = thisBit.split(": ") ;
					currentHeaders[thisBit[0].toLowerCase()] = thisBit[1] ;
					
				}
				
			} while (1)
			
		}
		
		else if (stage === 2) {
			
			let thisData ;
			if (data.indexOf(boundaryB) !== -1) {
				
				thisData = data.slice(0,data.indexOf(boundaryB)-2) ;
				data = data.slice(data.indexOf(boundaryB),data.length) ;
				dataString = data.toString() ;
				stage = 0 ;
				
			}
			else {
				
				thisData = data.slice(0,data.length) ;
				data = Buffer.alloc(0) ;
				dataString = "" ;
				
			}
			currentDataPipe.push(thisData) ;
			if (!stage) {
				
				currentDataPipe.emit("end") ;
				return parseTick() ;
				
			}
			return false ;
			
		}
		
	} ;
	req.on("data",d=>{
		
		data = Buffer.concat([data,d]) ;
		dataString += d.toString() ;
		/*parseTick().then(hi=>{
			
			endReady = hi ;
			if (endReady && reqEnded) {
				
				resolve() ;
				
			}
			
		}) ;*/
		endReady = parseTick() ;
		if (endReady && reqEnded) {
			
			resolve() ;
			
		}
		
	}) ;
	req.on("end",_=>{
		
		if (endReady) {
			
			resolve() ;
			
		}
		
		else {
			
			reqEnded = true ;
			
		}
		
	}) ;
	
}) ;
