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

const {createHash} = require("crypto") ;
const emmiter = require("events") ;

const jpsUtil = requireJPS("jps-util") ;

function newParser(req, resp) {
	let currentOpCode = 0 ;
	let currentDataFrame = Buffer.alloc(0) ;
	let dataToPush = Buffer.alloc(0) ;
	let pushOpCode = 0 ;
	let mask = Buffer.alloc(4) ;
	let currentDataPosition = 0 ;
	let isThisTheEnd = false ;
	let dataLeft = 0 ;
	let finalDataLeft = 0 ;
	let parseStage = 0 ;
	function gotData(d, i) {
		if (i >= d.length) {
			return ;
		}
		if (parseStage === 0) {
			if (d[i]&128) {
				isThisTheEnd = true ;
			} else {
				isThisTheEnd = false ;
			}
			currentOpCode = d[i]&15 ;
			parseStage = 1 ;
			gotData(d, i+1) ;
		} else if (parseStage === 1) {
			//Spec says we must be masked
			if (!(d[i]&128)) {
				req.socket.end() ;
				return ;
			}
			finalDataLeft = d[i] - 128 ;
			if (finalDataLeft < 126) {
				parseStage = 4 ;
				dataLeft = 4 ;
				mask = Buffer.alloc(4) ;
			} else {
				parseStage = dataLeft - 124 ;
				dataLeft = parseStage===2?2:8 ;
				currentDataFrame = Buffer.alloc(dataLeft) ;
			}
			currentDataPosition = 0 ;
			gotData(d, i+1) ;
		} else if (parseStage < 4) {
			if (d.length - i > dataLeft) {
				let readTo = i + dataLeft ;
				d.copy(currentDataFrame, currentDataPosition, i, readTo) ;
				if (parseStage === 2) {
					finalDataLeft = currentDataFrame.readUInt16BE(0) ;
				} else {
					finalDataLeft = currentDataFrame.readUInt32BE(0)<<32 ;
					finalDataLeft += currentDataFrame.readUInt32BE(4) ;
				}
				currentDataPosition = 0 ;
				parseStage = 4 ;
				dataLeft = 4 ;
				mask = Buffer.alloc(4) ;
				gotData(d, readTo) ;
			} else {
				d.copy(currentDataFrame, currentDataPosition, i, d.length) ;
			}
		} else if (parseStage === 4) {
			if (d.length - i > dataLeft) {
				let readTo = i + dataLeft ;
				d.copy(mask, currentDataPosition, i, readTo) ;
				currentDataPosition = 0 ;
				dataLeft = finalDataLeft ;
				parseStage = 5 ;
				currentDataFrame = Buffer.alloc(dataLeft) ;
				gotData(d, readTo) ;
			} else {
				d.copy(mask, currentDataPosition, i, d.length) ;
			}
		} else if (parseStage === 5) {
			if (d.length - i >= dataLeft) {
				let readTo = i + dataLeft ;
				d.copy(currentDataFrame, currentDataPosition, i, readTo) ;
				for (let doing = 0; doing < currentDataFrame.length; doing++) {
					currentDataFrame[doing] = currentDataFrame[doing] ^ mask[doing%4] ;
				}
				if (currentOpCode === 0) {
					dataToPush = Buffer.concat([dataToPush, currentDataFrame], dataToPush.length + currentDataFrame.length) ;
					if (isThisTheEnd) {
						if (pushOpCode === 1) {
							req.emit("text", dataToPush.toString("utf8")) ;
						} else if (pushOpCode === 2) {
							req.emit("binary", dataToPush) ;
						} else if (pushOpCode === 9) {
							let pongData = Buffer.allocUnsafe(dataToPush.length) ;
							dataToPush.copy(pongData) ;
							resp.que(()=>{
								resp.end(pongData, 10) ;
							}) ;
						}
					}
				} else if (!isThisTheEnd) {
					dataToPush = currentDataFrame ;
					pushOpCode = currentOpCode ;
				} else if (currentOpCode === 1) {
					req.emit("text", currentDataFrame.toString("utf8")) ;
				} else if (currentOpCode === 2) {
					req.emit("binary", currentDataFrame) ;
				} else if (currentOpCode === 9) {
					let pongData = Buffer.allocUnsafe(currentDataFrame.length) ;
					currentDataFrame.copy(pongData) ;
					resp.que(()=>{
						resp.end(pongData, 10) ;
					}) ;
				}
				currentDataFrame = Buffer.alloc(0) ;
				parseStage = 0 ;
			} else {
				d.copy(currentDataFrame, currentDataPosition, i, d.length) ;
				dataLeft -= d.length - i ;
			}
		}
	}
	req.socket.on("data", d=>gotData(d, 0)) ;
}

function newResponse(req) {
	let ended = true ;
	let que = new Array() ;
	let resp = {
		write: (data, opcode=NaN, isEnd=false)=>{
			try {
				let buff ;
				if (data.length < 126) {
					buff = Buffer.allocUnsafe(2) ;
					buff[1] = data.length ;
				} else if (data.length < (1<<16)-1) {
					buff = Buffer.allocUnsafe(3) ;
					buff.writeUInt16BE(data.length, 1) ;
				} else if (data.length < (1<<63)-1) {
					buff = Buffer.allocUnsafe(9) ;
					buff.writeUInt32BE(data.length>>32, 1) ;
					buff.writeUInt32BE(data.length&((1<<32)-1), 5) ;
				} else {
					throw new Error("Too much data to send in a single frame!") ;
				}
				if (isEnd) {
					buff[0] = 1<<7 ;
				} else {
					buff[0] = 0 ;
				}
				if (!ended) {
					opcode = 0 ;
				} else if (isNaN(opcode)) {
					if (typeof data === "string") {
						opcode = 1 ;
					} else {
						opcode = 2 ;
					}
				}
				ended = isEnd ;
				buff[0] += opcode ;
				req.socket.write(buff) ;
				req.socket.write(data) ;
				if (isEnd && que.length > 0) {
					que.shift()() ;
				}
			} catch (err) {
				console.warn("Error sending frame:") ;
				console.warn(err.stack) ;
			}
		},
		end: (data, opcode)=>resp.write(data, opcode, true),
		isEnded: ()=>ended,
		que: (cb, fasttrack=false)=>{
			if (typeof cb !== "function") {
				return resp.que(()=>resp.end(cb)) ;
			}
			let len ;
			if (fasttrack) {
				len =  que.unshift(cb) ;
			} else {
				len = que.push(cb) ;
			}
			if (len === 1 && ended) {
				que.shift()() ;
			}
			return len ;
		}
	} ;
	return resp ;
}

function handleUpgrade(req, s, secure) {
	try {
		//Close if we are not upgrading to web sockets
		if (req.headers.upgrade && req.headers["sec-websocket-key"] && req.headers.upgrade === "websocket") {
			let aReq = new emmiter() ;
			Object.assign(aReq, {
				//Normal request properties
				socket: s,
				headers: req.headers,
				httpVersion: req.httpVersion,
				method: req.method,
				rawHeaders: req.rawHeaders,
				rawTrailers: req.rawTrailers,
				destroy: req.destroy,
				url: req.url,
				
				end: s.end
			}) ;
			//Add the jps request properties
			module.exports.serverCalls.addReqProps(req, secure) ;
			//Function to accept the upgrade (must be called by the extension)
			aReq.accept = cb => {
				const hash = createHash("sha1") ;
				hash.on("readable", ()=>{
					const data = hash.read() ;
					if (data) {
						s.write("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + data.toString("base64") + "\r\n\r\n") ;
						aReq.url.protocol = "ws:" ;
						cb() ; //eslint-disable-line callback-return
					}
				}) ;
				hash.write(req.headers["sec-websocket-key"] + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11") ;
				hash.end() ;
			} ;
			//Create the response object
			let resp = newResponse(req) ;
			//Set up the parser
			newParser(aReq, resp) ;
			module.exports.serverCalls.doEvent("websocket", aReq.url.host, ()=>{
				if (handlersWS[aReq.url.host]) {
					handlersWS[aReq.url.host](req, resp) ;
				} else if (handlers[aReq.url.hostname]) {
					handlers[aReq.url.hostname](req, resp) ;
				}
			}, aReq, resp) ;
		} else {
			s.end() ;
		}
	} catch (err) {
		s.end() ;
		jpsUtil.coughtError(err, " in upgrade handler", null, req.jpid||"") ;
	}
}

let handlers = new Object() ;
let handlersWS = new Object() ;
function handleWebSocket(url, handler, incSearch=false) {
	if (incSearch) {
		handlersWS[url] = handler ;
	} else {
		handlers[url] = handler ;
	}
}
function isHandled(url, incSearch=false) {
	if (incSearch) {
		return Boolean(handlersWS[url]) ;
	}
	return Boolean(handlers[url]) ;
}

module.exports = {
	handleUpgrade,
	serverCalls: {
		addReqProps: ()=>{},
		doEvent: ()=>{}
	},
	handleWebSocket,
	isHandled
} ;
