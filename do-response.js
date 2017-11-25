/*
	
	JOTPOT Server
	Version 27A-0
	
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

"use strict";

//incSearch argument: whether or not the search should be included in the link/cache URL/func URL


const path = require("path") ;
const fs = require("fs") ;

const jpsUtil = requireJPS("jps-util") ;

let errorMessages = ["An unknown error occured."] ;
errorMessages[403] = "Sorry, however you are not permitted to access this file." ;
errorMessages[404] = "The page you are looking for may have been removed or moved to a new location!" ;

//Links, link sends one URL to another (like an alias)
let links = new Object() ;
let linksWS = new Object() ;
function createLink(from, to, incSearch=false) {
	if (incSearch) {
		linksWS[from] = to ;
	} else {
		links[from] = to ;
	}
}
function isLink(from, incSearch=false) {
	if (incSearch) {
		return Boolean(linksWS[from]) ;
	} 
	return Boolean(links[from]) ;
	
}
function getLink(from, incSearch=false) {
	if (incSearch) {
		return linksWS[from] ;
	} 
	return links[from] ;
	
}
function removeLink(from, incSearch=false) {
	if (incSearch) {
		linksWS[from] = undefined ;
	} else {
		links[from] = undefined ;
	}
}

//Cache, sets URL to respond with the content provided in the cache argument
let pages = new Object() ;
let pagesMTime = new Object() ;
let pagesWS = new Object() ;
let pagesWSMTime = new Object() ;
function addCache(from, cache, incSearch=false, modTime=NaN) {
	if (incSearch) {
		pagesWS[from] = cache ;
		pagesWSMTime[from] = modTime ;
	} else {
		pages[from] = cache ;
		pagesMTime[from] = modTime ;
	}
}
function cacheFile(url) {
	return new Promise((resolve, reject) => fs.stat(path.join(process.cwd(), "sites", url), (err, stats) => {
		if (err) {
			reject(err) ;
		} else {
			fs.readFile(path.join(process.cwd(), "sites", url), (err, data) => {
				if (err) {
					reject(err) ;
				} else {
					addCache(url, data, false, stats.mtime.getTime()) ;
				}
			}) ;
		}
	})) ;
}
function cacheFileAs(url, file) {
	return new Promise((resolve, reject) => fs.stat(path.join(process.cwd(), "sites", file), (err, stats) => {
		if (err) {
			reject(err) ;
		} else {
			fs.readFile(path.join(process.cwd(), "sites", file), (err, data) => {
				if (err) {
					reject(err) ;
				} else {
					addCache(url, data, false, stats.mtime.getTime()) ;
				}
			}) ;
		}
	})) ;
}
function cacheFileSync(url) {
	const stats = fs.statSync(path.join(process.cwd(), "sites", url)) ;
	const data = fs.readFileSync(path.join(process.cwd(), "sites", url)) ;
	addCache(url, data, false, stats.mtime.getTime()) ;
}
function cacheFileAsSync(url, file) {
	const stats = fs.statSync(path.join(process.cwd(), "sites", file)) ;
	const data = fs.readFileSync(path.join(process.cwd(), "sites", file)) ;
	addCache(url, data, false, stats.mtime.getTime()) ;
}
function isCache(url, incSearch=false) {
	if (incSearch) {
		return Boolean(pagesWS[url]) ;
	} 
	return Boolean(pages[url]) ;
}
function getCache(url, incSearch=false) {
	if (incSearch) {
		return pagesWS[url] ;
	} 
	return pages[url] ;
}
function removeCache(url, incSearch=false) {
	if (incSearch) {
		pagesWS[url] = undefined ;
		pagesWSMTime[url] = undefined ;
	} else {
		pages[url] = undefined ;
		pagesMTime[url] = undefined ;
	}
}

//Funcs, handles a page with a given function.
//Function is called with the req and resp as the arguments. If it returns/resolves true, the request is stoped.
//Otherwise, it carrys on.
let funcs = new Object() ;
let funcsWS = new Object() ;
function handlePage(page, func, incSearch=false) {
	if (incSearch) {
		funcsWS[page] = func ;
	} else {
		funcs[page] = func ;
	}
}
function isHandled(page, incSearch=false) {
	if (incSearch) {
		return Boolean(funcsWS[page]) ;
	} 
	return Boolean(funcs[page]) ;
	
}
function handle(page, args, incSearch=false) {
	if (incSearch) {
		return funcsWS[page](...args) ;
	} 
	return funcs[page](...args) ;
	
}
function removePageHandler(page, incSearch=false) {
	if (incSearch) {
		funcsWS[page] = undefined ;
	} else {
		funcs[page] = undefined ;
	}
}

let learning = new Object() ;

//Is the URL learned?
//checkLevel:
//			 0: Exact URL
//			 1: URL but with any search
//			 2: Any subpath with an search
function isLearned(url, checkLevel=0) {
	if (checkLevel === 0) {
		return Boolean(learning[url]) ;
	}
	let isLearned = false ;
	for (let doing in learning) {
		if (doing) {
			if (doing.indexOf(url + (checkLevel===1)?"?":"") === 0) {
				isLearned = true ;
				break ;
			}
		}
	}
	return isLearned ;
	
}

//Unlearn, will be releared on next request?
//level:
//		0: Exact URL
//		1: URL but with any search
//		2: Any subpath with an search
function unlearn(url, level=0) {
	if (level === 0) {
		learning[url] = undefined ;
	} else {
		for (let doing in learning) {
			if (doing) {
				if (doing.indexOf(url + (level===1)?"?":"") === 0) {
					learning[doing] = undefined ;
				}
			}
		}
	}
}

function doLinks(req) {
	let doTheLoop = false ;
	let wasChanged = false ;
	let origValue ;
	do {
		doTheLoop = false ;
		while (linksWS[req.url.fullvalue]) {
			origValue = req.url.fullvalue ;
			req.url.fullvalue = linksWS[req.url.fullvalue] ;
			if (req.url.fullvalue === origValue) {
				break ;
			}
			wasChanged = true ;
		}
		if (links[req.url.fullvalue]) {
			origValue = req.url.value ;
			req.url.value = links[req.url.value] ;
			if (req.url.value !== origValue) {
				doTheLoop = true ;
				wasChanged = true ;
			}
		}
	} while (doTheLoop) ;
	return wasChanged ;
}

//Calls any function handlers.
//Calls cb with int arg:
//	0 if there were no handlers
//	1 if it should carry on
//	2 if it should carry on but it needs to catch the next function
//	3 if the handleeThing needs to be called again
//	4 same as 3 but needs to be in try catch
function callHandlers(req, resp, val, ws, cb, rp) {
	let origVal = req.url.fullvalue ;
	let func ;
	if (ws) {
		func = funcsWS[val] ;
	} else {
		func = funcs[val] ;
	}
	if (!func) {
		cb(0) ;
		return ;
	}
	//Call it
	let rv ;
	try {
		rv = func(req, resp) ;
	} catch (err) {
		jpsUtil.coughtError(err, " in a request handler", resp, req.jpid, "An error occured in a request handler!") ;
		return ;
	}
	//If we havn't handled it
	if (!rv) {
		//1 if we need to start again (URL change), or 0 if we can carry on.
		if (req.url.fullvalue !== origVal) {
			cb(3) ; //eslint-disable-line callback-return
		} else {
			cb(1) ; //eslint-disable-line callback-return
		}
	} else if (typeof rv.then === "function") {
		//If it is a promise, do the same behavior when it resolves.
		//Or send an error on a rejection
		rv.then(v=>{
			if (v) {
				rp.callback(false) ;
			} else if (req.url.fullvalue !== origVal) {
				cb(4) ; //eslint-disable-line callback-return
			} else {
				cb(2) ; //eslint-disable-line callback-return
			}
		}, err=>{
			jpsUtil.coughtError(err, " in a request handler", resp, req.jpid, "An error occured in a request handler!") ;
		}) ;
	} else {
		rp.callback(false) ;
	}
}

function afterFirstSendFile(req, resp, file, rp, done) {
	if (done[0]) {
		if (rp.canLearn) {
			learning[rp.learnValue] = [2, file, done[1]] ;
		}
		rp.callback(false) ;
		return ;
	} else if (done[1] === "DIR") {
		req.url.pathname = req.url.pathname + "/index.html" ;
		req.url.normalize() ;
	} else if (done[1].code === "ENOENT") {
		//Not found, so now 404
		rp.code = 404 ;
		req.url.normalize() ;
		req.url.pathname = req.url.pathname + ".page" ;
	} else if (done[1].code === "EACCES") {
		//No perms, so now 403
		rp.code = 403 ;
		req.url.normalize() ;
		req.url.pathname = req.url.pathname + ".page" ;
	} else {
		//NO IDEA!!!
		req.url.normalize() ;
		req.url.pathname = req.url.pathname + ".page" ;
	}
	if (rp.canLearn && learning[req.url.fullvalue]) {
		learning[rp.learnValue] = learning[req.url.fullvalue] ;
		createResponse(req, resp, rp.callback) ;
		return ;
	}
	rp.isFinal = true ;
	handleeThing0(req, resp, rp) ;
}

function afterFinalSendFile(req, resp, file, rp, done) {
	if (done[0]) {
		if (rp.canLearn) {
			learning[rp.learnValue] = [2, file, done[2]] ;
		}
		rp.callback(false) ;
		return ;
	} else if (done[1].code === "EACCES") {
		//Now 403
		rp.code = 403 ;
	} else if (done[1].code !== "ENOENT") {
		//If it does exist or another error has occured, send a 500.
		rp.code = 500 ;
	}
	//Learn that this is a 404 if we can
	if (rp.code === 404 && rp.canLearn) {
		learning[rp.learnValue] = [3] ;
	}
	//Send the error, fallback to error code 0 if there is no message for the current code.
	module.exports.sendError(rp.code, errorMessages[rp.code] || errorMessages[0], resp, req.jpid) ;
	rp.callback(false) ;
}

class internalResponseProps {
	constructor(callback, learnValue, canLearn=true, code=500, isFinal=false) {
		this.learnValue = learnValue ;
		this.canLearn = canLearn ;
		this.code = code ;
		this.callback = callback ;
		this.isFinal = isFinal ;
	}
}

function handleeThing0(req, resp, rp) {
	let origURL = req.url.fullvalue ;
	if (module.exports.isThereAHandler("handle", req.url.host)) {
		rp.canLearn = false ;
		module.exports.doEvent("handle", req.url.host, ()=>{
			if (req.url.fullvalue === origURL) {
				handleeThing1(req, resp, rp) ;
			} else {
				handleeThing0(req, resp, rp) ;
			}
		}) ;
	} else {
		handleeThing1(req, resp, rp) ;
	}
}

function handleeThing1(req, resp, rp) {
	if (doLinks(req)) {
		handleeThing0(req, resp, rp) ;
	}
	let val = req.usePortInDirectory?req.url.fullvalue:req.url.fullvaluenoport ;
	callHandlers(req, resp, val, true, handleLevel=>{
		if (handleLevel) {
			rp.canLearn = false ;
			if (handleLevel === 1) {
				handleeThing2(req, resp, val, rp) ;
			} else if (handleLevel === 3) {
				handleeThing0(req, resp, rp) ;
			} else {
				try {
					if (handleLevel === 2) {
						handleeThing2(req, resp, val, rp) ;
					} else {
						handleeThing0(req, resp, rp) ;
					}
				} catch (err) {
					jpsUtil.coughtError(err, " creating the response", resp, req.jpid) ;
				}
			}
		} else {
			handleeThing2(req, resp, val, rp) ;
		}
	}, rp) ;
}

function handleeThing2(req, resp, val, rp) {
	if (pagesWS[val]) {
		let file = path.normalize((req.usePortInDirectory?req.url.host:req.url.hostname).replace(/:/g, ";") + req.url.pathname) ;
		//Learn if we can learn
		if (rp.canLearn) {
			learning[rp.learnValue] = [0, file, val] ;
		}
		if (!module.exports.doMethodLogic(req, resp, false)) {
			module.exports.sendCache(
				file,
				pagesWS[val],
				resp,
				resp.vars,
				req,
				200,
				pagesWSMTime[val],
				rp.canLearn?'0':'-1'
			) ;
		}
		rp.callback(true) ;
		return ;
	}
	val = req.usePortInDirectory?req.url.value:req.url.valuenoport ;
	callHandlers(req, resp, val, false, handleLevel=>{
		if (handleLevel) {
			rp.canLearn = false ;
			if (handleLevel === 1) {
				handleeThing3(req, resp, val, rp) ;
			} else if (handleLevel === 3) {
				handleeThing0(req, resp, rp) ;
			} else {
				try {
					if (handleLevel === 2) {
						handleeThing3(req, resp, val, rp) ;
					} else {
						handleeThing0(req, resp, rp) ;
					}
				} catch (err) {
					jpsUtil.coughtError(err, " creating the response", resp, req.jpid) ;
				}
			}
		} else {
			handleeThing3(req, resp, val, rp) ;
		}
	}, rp) ;
}

function handleeThing3(req, resp, val, rp) {
	let file = path.normalize((req.usePortInDirectory?req.url.host:req.url.hostname).replace(/:/g, ";") + req.url.pathname) ;
	if (pages[val]) {
		//Learn if we can learn
		if (rp.canLearn) {
			learning[rp.learnValue] = [1, file, val] ;
		}
		if (!module.exports.doMethodLogic(req, resp, false)) {
			module.exports.sendCache(
				file,
				pages[val],
				resp,
				resp.vars,
				req,
				200,
				pagesMTime[val],
				rp.canLearn?'0':'-1'
			) ;
		}
		rp.callback(true) ;
		return ;
	}
	if (rp.isFinal) {
		module.exports.sendFile(
			file,
			resp,
			resp.vars,
			req,
			()=>module.exports.doMethodLogic(req, resp, false),
			rp.canLearn?'0':'-1'
		).then(done=>afterFinalSendFile(req, resp, file, rp, done)) ;
	} else {
		module.exports.sendFile(
			file,
			resp,
			resp.vars,
			req,
			()=>module.exports.doMethodLogic(req, resp, false),
			rp.canLearn?'0':'-1'
		).then(done=>afterFirstSendFile(req, resp, file, rp, done)) ;
	}
}

//Function that sends a response for the given request
function createResponse(req, resp, callback) {
	try {
		/* eslint-disable consistent-return */
		//If we have leared how to handle the request
		if (module.exports.enableLearning && learning[req.url.fullvalue]) {
			resp.setHeader("JP-Was-Learned", "1") ;
			//Learn types are:
			//                0: Cache with search
			//                1: Cache without search
			//                2: File found
			//                3: 404
			if (learning[req.url.fullvalue][0] === 0) {
				//Unlearn this if it is now invalid
				if (!pagesWS[learning[req.url.fullvalue][2]]) {
					learning[req.url.fullvalue] = undefined ;
					createResponse(req, resp, callback) ;
					return ;
				}
				if (!module.exports.doMethodLogic(req, resp, false)) {
					module.exports.sendCache(
						learning[req.url.fullvalue][1],
						pagesWS[learning[req.url.fullvalue][2]],
						resp,
						resp.vars,
						req,
						200,
						pagesWSMTime[learning[req.url.fullvalue][2]],
						'1'
					) ;
				}
				callback(true) ; //eslint-disable-line callback-return
			} else if (learning[req.url.fullvalue][0] === 1) {
				//Unlearn this if it is now invalid
				if (!pages[learning[req.url.fullvalue][2]]) {
					learning[req.url.fullvalue] = undefined ;
					createResponse(req, resp, callback) ;
					return ;
				}
				if (!module.exports.doMethodLogic(req, resp, false)) {
					module.exports.sendCache(
						learning[req.url.fullvalue][1],
						pages[learning[req.url.fullvalue][2]],
						resp,
						resp.vars,
						req,
						200,
						pagesMTime[learning[req.url.fullvalue][2]],
						'1'
					) ;
				}
				callback(true) ; //eslint-disable-line callback-return
			} else if (learning[req.url.fullvalue][0] === 2) {
				module.exports.sendFile(
					learning[req.url.fullvalue][1],
					resp,
					resp.vars,
					req,
					()=>module.exports.doMethodLogic(req, resp, false),
					'1'
				).then(done=>{
					//Unlearn this if it is now invalid
					if (!done[0]) {
						learning[req.url.fullvalue] = undefined ;
						createResponse(req, resp, callback) ;
						return ;
					}
					callback(true) ;
				}) ;
			} else if (learning[req.url.fullvalue][0] === 3) {
				module.exports.sendError(404, errorMessages[404] || errorMessages[0], resp, req.jpid) ;
				callback(true) ; //eslint-disable-line callback-return
			} else {
				//Try again if this isn't valid
				learning[req.url.fullvalue] = undefined ;
				createResponse(req, resp, callback) ;
			}
			return ;
		}
		
		resp.setHeader("JP-Was-Learned", "0") ;
		
		handleeThing0(req, resp, new internalResponseProps(callback, req.url.fullvalue)) ;
		
		/* eslint-enable consistent-return */
	} catch (err) {
		jpsUtil.coughtError(err, " creating the response", resp, req.jpid) ;
	}
}

module.exports = {
	//Main exports:
	createResponse,
	doLinks,
	
	//Export functions
	createLink,
	isLink,
	getLink,
	removeLink,
	addCache,
	cacheFile,
	cacheFileAs,
	cacheFileSync,
	cacheFileAsSync,
	isCache,
	getCache,
	removeCache,
	handlePage,
	isHandled,
	handle,
	removePageHandler,
	isLearned,
	unlearn,
	
	//To get functions and config from loader of the module
	sendFile: ()=>{},
	sendCache: ()=>{},
	sendError: ()=>{},
	doMethodLogic: ()=>{},
	doEvent: ()=>{},
	isThereAHandler: ()=>{},
	enableLearning: true
} ;
