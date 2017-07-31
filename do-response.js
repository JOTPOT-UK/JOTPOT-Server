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

//incSearch argument: whether or not the search should be included in the link/cache URL/func URL


const path = require("path") ;
const fs = require("fs") ;

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
		delete linksWS[from] ;
	} else {
		delete links[from] ;
	}
}

//Cache, sets URL to respond with the content provided in the cache argument
let pages = new Object() ;
let pagesWS = new Object() ;
function addCache(from, cache, incSearch=false) {
	if (incSearch) {
		pagesWS[from] = cache ;
	} else {
		pages[from] = cache ;
	}
}
function cacheFile(url) {
	return new Promise((resolve, reject) => fs.readFile(path.join(process.cwd(), "sites", url), (err, data) => {
		if (err) {
			reject(err) ;
		} else {
			addCache(url, data, false) ;
		}
	})) ;
}
function cacheFileAs(url, file) {
	return new Promise((resolve, reject) => fs.readFile(path.join(process.cwd(), "sites", file), (err, data) => {
		if (err) {
			reject(err) ;
		} else {
			addCache(url, data, false) ;
		}
	})) ;
}
function cacheFileSync(url) {
	const data = fs.readFileSync(path.join(process.cwd(), "sites", url)) ;
	addCache(url, data, false) ;
}
function cacheFileAsSync(url, file) {
	const data = fs.readFileSync(path.join(process.cwd(), "sites", file)) ;
	addCache(url, data, false) ;
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
		delete pagesWS[url] ;
	} else {
		delete pages[url] ;
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
		delete funcsWS[page] ;
	} else {
		delete funcs[page] ;
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
		if (doing.indexOf(url + (checkLevel===1)?"?":"") === 0) {
			isLearned = true ;
			break ;
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
		delete learning[url] ;
	} else {
		for (let doing in learning) {
			if (doing.indexOf(url + (level===1)?"?":"") === 0) {
				delete learning[doing] ;
			}
		}
	}
}

function doLinks(req) {
	let doTheLoop = false ;
	let origValue ;
	do {
		doTheLoop = false ;
		while (linksWS[req.url.fullvalue]) {
			origValue = req.url.fullvalue ;
			req.url.fullvalue = linksWS[req.url.fullvalue] ;
			if (req.url.fullvalue === origValue) {
				break ;
			}
		}
		if (links[req.url.fullvalue]) {
			origValue = req.url.value ;
			req.url.value = links[req.url.value] ;
			if (req.url.value !== origValue) {
				doTheLoop = true ;
			}
		}
	} while (doTheLoop) ;
}

//Function that sends a response for the given request
function createResponse(req, resp, timeRecieved=[-1,-1]) {
	/* eslint-disable consistent-return */
	return new Promise((resolver)=>{
		const resolve =(...args)=> {
			resolver(...args) ;
		} ;
		//If we have leared how to handle the request
		if (module.exports.enableLearning && learning[req.url.fullvalue]) {
			resp.setHeader("JP-Was-Learned", "1") ;
			if (timeRecieved[0] !== -1 && timeRecieved[1] !== -1) {
				let timeTaken = process.hrtime(timeRecieved) ;
				console.log(`${req.jpid}\tRequest took ${timeTaken[0] * 1000 + timeTaken[1] * 10e-6}ms to process.`) ;
			}
			//Learn types are:
			//                0: Cache with search
			//                1: Cache without search
			//                2: File found
			//                3: 404
			if (learning[req.url.fullvalue][0] === 0) {
				//Unlearn this if it is now invalid
				if (!pagesWS[learning[req.url.fullvalue][2]]) {
					delete learning[req.url.fullvalue] ;
					return createResponse(req, resp) ;
				}
				module.exports.sendCache(learning[req.url.fullvalue][1], pagesWS[learning[req.url.fullvalue][2]], resp, resp.vars, req, 200) ;
				resolve([200, true]) ;
			} else if (learning[req.url.fullvalue][0] === 1) {
				//Unlearn this if it is now invalid
				if (!pages[learning[req.url.fullvalue][2]]) {
					delete learning[req.url.fullvalue] ;
					return createResponse(req, resp) ;
				}
				module.exports.sendCache(learning[req.url.fullvalue][1], pages[learning[req.url.fullvalue][2]], resp, resp.vars, req, 200) ;
				resolve([200, true]) ;
			} else if (learning[req.url.fullvalue][0] === 2) {
				return module.exports.sendFile(learning[req.url.fullvalue][1], resp, resp.vars, req).then(done=>{
					//Unlearn this if it is now invalid
					if (!done[0]) {
						delete learning[req.url.fullvalue] ;
						return createResponse(req, resp) ;
					} 
					resolve([200, true]) ;
					
				}) ;
			} else if (learning[req.url.fullvalue][0] === 3) {
				module.exports.sendError(404, errorMessages[404] || errorMessages[0], resp, req.jpid) ;
				resolve([404, true]) ;
			} else {
				//Try again if this isn't valid
				delete learning[req.url.fullvalue] ;
				return createResponse(req, resp) ;
			}
		}
		
		resp.setHeader("JP-Was-Learned", "0") ;
		
		//Get normilized file path, replaces ':' with ';' in the host and only includes the port if req.usePortInDirectory is true
		let file ;
		
		let origValue = "" ;
		let canLearn = module.exports.enableLearning ;
		const learnValue = req.url.fullvalue ;
		
		const handleeThing =()=> {
			doLinks(req) ;
			file = path.normalize((req.usePortInDirectory?req.url.host:req.url.hostname).replace(/:/g, ";") + req.url.pathname) ;
			let val = req.usePortInDirectory?req.url.fullvalue:req.url.fullvaluenoport ;
			if (funcsWS[val]) {
				//As the function may have changed the URL, or headers etc. we can no longer learn from this request
				canLearn = false ;
				origValue = req.url.href ;
				//Return if we have handled it
				const rv = funcsWS[val](req, resp) ;
				if (!rv || !rv[0]) {
					//So, we haven't handled it hey...
					if (req.url.href !== origValue) {
						//Start again...
						return handleeThing() ;
					}
					//Otherwise, nothing changed so we just carry on
				} else if (typeof rv.then === "function") {
					//Ah, we have a promise, so lets return one!
					return rv.then(rv=>{
						if (!rv || !rv[0]) {
							//Not handled, so start again
							return handleeThing() ;
						}
						//YAY!!! Handled
						return [true, 0] ;
					}) ;
				} else {
					//Yup, we are handled
					return [true, 0] ;
				}
			}
			if (pagesWS[val]) {
				//Learn if we can learn
				if (canLearn) {
					learning[learnValue] = [0, file, val] ;
				}
				module.exports.sendCache(file, pagesWS[val], resp, resp.vars, req, 200) ;
				return [true, 200] ;
			}
			val = req.usePortInDirectory?req.url.value:req.url.valuenoport ;
			if (funcs[val]) {
				//As the function may have changed the URL, or headers etc. we can no longer learn from this request
				canLearn = false ;
				//Return if we have handled it
				const rv = funcs[val](req, resp) ;
				if (!rv || !rv[0]) {
					//So, we haven't handled it hey...
					//We are going to start again incase it has changed metadata and thus a previous handler may need it
					return handleeThing() ;
				} else if (typeof rv.then === "function") {
					//Ah, we have a promise, so lets return one!
					return rv.then(rv=>{
						if (!rv || !rv[0]) {
							//Not handled, so start again
							return handleeThing() ;
						}
						//YAY!!! Handled
						return [true, 0] ;
					}) ;
				}
				//Yup, we are handled
				return [true, 0] ;
			}
			if (pages[val]) {
				//Learn if we can learn
				if (canLearn) {
					learning[learnValue] = [1, file, val] ;
				}
				module.exports.sendCache(file, pages[val], resp, resp.vars, req, 200) ;
				return [true, 200] ;
			}
			return [false] ;
		} ;
		const toDo =()=> {
			if (timeRecieved[0] !== -1 && timeRecieved[1] !== -1) {
				let timeTaken = process.hrtime(timeRecieved) ;
				console.log(`${req.jpid}\tRequest took ${timeTaken[0] * 1000 + timeTaken[1] * 10e-6}ms to process.`) ;
			}
			let code = 500 ;
			//Try sending
			module.exports.sendFile(file, resp, resp.vars, req).then(done=>{
				if (done[0]) {
					if (canLearn) {
						learning[learnValue] = [2, file] ;
					}
					resolve([false, 200]) ;
					return ;
				} else if (done[1] === "DIR") {
					req.url.pathname = req.url.pathname + "/index.html" ;
					//Ensure there aren't any //
					while (req.url.pathname.indexOf("//") !== -1) {
						req.url.pathname = req.url.pathname.replace(/\/\//g, "/") ;
					}
				} else if (done[1].code === "ENOENT") {
					//Not found, so now 404
					code = 404 ;
					req.url.pathname = req.url.pathname + ".page" ;
				} else if (done[1].code === "EACCES") {
					//No perms, so now 403
					code = 403 ;
					req.url.pathname = req.url.pathname + ".page" ;
				} else {
					//NO IDEA!!!
					req.url.pathname = req.url.pathname + ".page" ;
				}
				okThen(()=>module.exports.sendFile(file, resp, resp.vars, req).then(done=>{
					if (done[0]) {
						if (canLearn) {
							learning[learnValue] = [2, file] ;
						}
						resolve([false, 200]) ;
						return ;
					} else if (done[1].code === "EACCES") {
						//Now 403
						code = 403 ;
					} else if (done[1].code !== "ENOENT") {
						//If it does exist or another error has occured, send a 500.
						code = 500 ;
					}
					//Learn that this is a 404 if we can
					if (code === 404 && canLearn) {
						learning[learnValue] = [3] ;
					}
					//Send the error, fallback to error code 0 if there is no message for the current code.
					module.exports.sendError(code, errorMessages[code] || errorMessages[0], resp, req.jpid) ;
					resolve([false, code]) ;
				})) ;
			}) ;
		} ;
		//runs the handlers and checks cache, then if the request is not handles, calls the cb
		const okThen = cb => {
			const wasHandled = handleeThing() ;
			if (wasHandled[0]) {
				resolve([false, wasHandled[1]]) ;
			} else if (typeof wasHandled.then === "function") {
				wasHandled.then(isItHandledYet=>{
					if (isItHandledYet[0]) {
						resolve([false, isItHandledYet[1]]) ;
					} else {
						// cb return would be pointless
						cb() ;  //eslint-disable-line callback-return
					}
				}) ;
			} else {
				// cb return would be pointless
				cb() ;  //eslint-disable-line callback-return
			}
		} ;
		okThen(toDo) ;
	}) ;
	/* eslint-enable consistent-return */
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
	enableLearning: true
} ;
