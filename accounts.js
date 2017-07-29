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


let fs = require("fs") ;
let Events = require("events") ;

let stamp = 0 ;

//Create the emitter.
class NewEmitter extends Events {}
const procUpdate = new NewEmitter() ;

//Object for storing users.
let loggedIn = new Object() ;

//When we get a message
process.on("message",toDo=>{
	
	if (toDo[0] === "proc-authed") {
		
		procUpdate.emit(`authed-${toDo[1]}-${toDo[2]}-${toDo[4]}`,toDo[3]) ;
		
	}
	
	else if (toDo[0] === "proc-usn") {
		
		procUpdate.emit(`username-${toDo[1]}-${toDo[2]}`,[toDo[3],toDo[4]||""]) ;
		
	}
	
	//If it id for us.
	else if (toDo[0] === "proc-update") {
		
		//Update our object.
		loggedIn = toDo[1] ;
		//And emit an event to say we are up to date.
		procUpdate.emit("update") ;
		
	}
	
	else if (toDo[0] === "proc-added") {
		
		procUpdate.emit("added-"+toDo[1]) ;
		
	}
	
	else if (toDo[0] === "proc-deled") {
		
		procUpdate.emit("deled-"+toDo[1]) ;
		
	}
	
}) ;

//Gets a user ID from the request object.
function getUserID(req) {
	
	//If there are no cookies.
	if (typeof req.headers.cookie === "undefined") {
		
		//Then they have no ID.
		return false ;
		
	}
	
	//Split their cookies up.
	let cookieArray = req.headers.cookie.split("; ") ;
	
	//Object for storing cookies.
	let cookies = new Object() ;
	
	//Go through each cookie.
	for (let doing in cookieArray) {
		
		//Add it to the object.
		cookies[cookieArray[doing].split("=")[0]] = cookieArray[doing].split("=")[1] ;
		
	}
	
	//If the JOTPOTUID cookie doesn't exist.
	if (typeof cookies.JOTPOTUID === "undefined") {
		
		//Then they dont have one.
		return false ;
		
	}
	
	//Otherwise, generate their ID.
	return `${req.headers["x-forwarded-for"] || req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress}(${req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress})---${cookies.JOTPOTUID}` ;
	
}

//Function to create a new UID and add it to the set-cookie header.
function makeNewUID(req,resp) {
	
	//Gen code...
	let newUID = Math.random() + Math.random() + Math.random() + Math.random() + Math.random() + Math.random() + Math.random() + Math.random() + Math.random() + Math.random() ;
	newUID *= 10e10 ;
	newUID *= Math.random() * Math.random() * Math.random() ;
	newUID *= 10e30 ;
	newUID *= Math.random() * Math.random() * Math.random() ;
	newUID *= 10e30 ;
	newUID += Math.random() * 10e5 ;
	newUID += Math.random() * 10e5 ;
	newUID += Math.random() * 10e5 ;
	
	//Set their cookies.
	resp.setHeader("Set-Cookie","JOTPOTUID=" + Math.round(newUID).toString(36)) ;
	
	//Return their ID.
	return `${req.headers["x-forwarded-for"] || req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress}(${req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress})---${newUID}` ;
	
}

module.exports.getUserID = getUserID ;
module.exports.makeNewUID = makeNewUID ;

//Class for an account system.
class proc {
	
	constructor (name,db,pages,pagesEx,login,loginPage,logout,logoutPage,reg,regPage,loginRedirect,https) {
		
		//Grab the database as an object.
		this.accounts = JSON.parse(fs.readFileSync(db).toString()) ;
		
		//Create the pages and starts array
		this.pages = new Array() ;
		this.starts = new Array() ;
		
		//Go through the pages.
		for (let doing in pages) {
			
			//If it contains a wildcard at the end
			if (pages[doing].lastIndexOf("*") === pages[doing].length - 1) {
				
				//It is a start (without the wildcard)
				this.starts.push(pages[doing].substring(0,pages[doing].length - 1)) ;
				
			}
			
			//But if not
			else {
				
				//It is a normal page
				this.pages.push(pages[doing]) ;
				
			}
			
		}
		
		
		//Create the special pages objects.
		this.specialPages = new Object() ;
		this.specialPagesP = new Object() ;
		
		//And fill them up.
		this.specialPages[login] = "login" ;
		this.specialPages[logout] = "logout" ;
		this.specialPages[reg] = "reg" ;
		this.specialPagesP[loginPage] = "loginPage" ;
		this.specialPagesP[logoutPage] = "logoutPage" ;
		this.specialPagesP[regPage] = "regPage" ;
		this.specialPages["login"] = login ;
		this.specialPages["logout"] = logout ;
		this.specialPages["reg"] = reg ;
		this.specialPagesP["loginPage"] = loginPage ;
		this.specialPagesP["logoutPage"] = logoutPage ;
		this.specialPagesP["regPage"] = regPage ;
		
		this.loginRedirect = loginRedirect || "/" ;
		
		//If https is not defined
		if (typeof https === "undefined") {
			
			this.https = false ;
			
		}
		
		else {
			
			this.https = https ;
			
		}
		
		
		//Bind the doAnything function.
		this.doAnything = this.doAnything.bind(this) ;
		
		//Create an object to store the usernames.
		loggedIn[name] = new Object() ;
		
		//Set the ID
		this.ID = name ;
		
		
		//Set up the pages objects.
		this.pagesEx = new Array() ;
		this.pagesExS = new Array() ;
		
		//Go through the pages argument.
		for (let doing in pagesEx) {
			
			//If there is a wildcard at the end.
			if (pagesEx[doing].lastIndexOf("*") === pagesEx[doing].length - 1) {
				
				//It is a start without the wildcard.
				this.pagesExS.push(pagesEx[doing].substring(0,pagesEx[doing].length - 1)) ;
				
			}
			
			else {
				
				//Meh, normal page.
				this.pagesEx.push(pagesEx[doing]) ;
				
			}
			
		}
		
		this.isAuthed = this.isAuthed.bind(this) ;
		this.getUsername = this.getUsername.bind(this) ;
		
		//Get the master process to set up a sync for it.
		process.send(["proc","new",name]) ;
		
		Object.defineProperty(this,"users",{
			
			get: function () {
				
				return loggedIn[name] ;
				
			},
			
			set: function (v) {
				
				loggedIn[name] = v ;
				
			}
			
		}) ;
		
	}
	
	//Resolves true if the user has permission to access it from this system, false if not.
	doAnything(req,resp) {
		
		return new Promise((resolve) => {
			
			let page = req.url.value ;
			
			//If this doesn't protect it.
			if (typeof this.specialPagesP[page] !== "undefined" || this.pagesEx.indexOf(page) !== -1) {
				
				//The user can access it.
				resolve([true]) ;
				return ;
				
			}
			
			//Go through the ends.
			for (let doing in this.pagesExS) {
				
				//If excluded
				if (page.indexOf(this.pagesExS[doing]) === 0) {
					
					//Then the user can access it.
					resolve([true]) ;
					return ;
					
				}
				
			}
			
			//let shouldCheckBlock = true ;
			
			//If it is a special page, then we dont need to check.
			//if (typeof this.specialPages[page] !== "undefined" ) {
				
			//	shouldCheckBlock = false ;
				
			//}
			
			//If we do need to checl.
			if (typeof this.specialPages[page] === "undefined") {
				
				let isAuthed = true ;
				
				//If it is protected, the user is not authed.
				if (this.pages.indexOf(page) !== -1) {
					
					isAuthed = false ;
					
				}
				
				//And the starts, but only if we need to.
				if (isAuthed) {
					
					for (let doing in this.starts) {
						
						if (page.indexOf(this.starts[doing]) === 0) {
							
							isAuthed = false ;
							break ;
							
						}
						
					}
					
				}
				
				//If the user can still access it, then resolve true.
				if (isAuthed) {
					
					resolve([true]) ;
					return ;
					
				}
				
			}
			
			let user = req.user || getUserID(req) ;
			if (user === false) {
				
				user = makeNewUID(req,resp) ;
				
			}
			req.user = user ;
			
			if (typeof this.specialPages[page] !== "undefined") {
				
				let returned = this[this.specialPages[page]](req,resp,user) ;
				
				if (typeof returned[0] !== "undefined") {
					
					resolve(returned) ;
					
				}
				
				else {
					
					returned.then((...args)=>resolve(...args)) ;
					
				}
				
				return ;
				
			}
			
			this.isAuthed(user).then(isAuthed=>{
				
				if (isAuthed) {
					
					resolve([true]) ;
					
					
				}
				
				else {
					
					resolve([false,"redirect",(this.https?"https://":"http://") + removeCommonPorts(this.specialPagesP["loginPage"], this.https)]) ;
					
				}
				
			}) ;
			
		}) ;
		
	}
	
	isAuthed (user) {
		
		let thisStamp = stamp++ ;
		return new Promise((resolve) => {
			
			procUpdate.once(`authed-${this.ID}-${user}-${thisStamp}`,rv=>{
				
				resolve(rv) ;
				
			}) ;
			process.send(["proc","authed",this.ID,user,thisStamp]) ;
			
		}) ;
		
	}
	
	getUsername (user) {
		
		return new Promise((resolve) => {
			
			procUpdate.once(`username-${this.ID}-${user}`,rv=>{
				
				if (rv[0]) {
					
					resolve(rv[1]) ;
					
				}
				
				else {
					
					resolve(false) ;
					
				}
				
			}) ;
			process.send(["proc","usn",this.ID,user]) ;
			
		}) ;
		
	}
	
	login (req,resp,user) {
		
		return new Promise((resolve)=>{req.on("data",(d)=>{
			
			d = decodeURIComponent(decodeURI(d.toString()).replace(/\+/g," ")).split("&") ;
			let args = new Object() ;
			
			for (let doing in d) {
				
				args[d[doing].split("=")[0]] = d[doing].split("=")[1] ;
				
			}
			
			let sendIUOP =()=> {
				
				resolve([false,"redirect",(this.https?"https://":"http://") + removeCommonPorts(this.specialPagesP.loginPage, this.https)]) ;
				
			} ;
			
			let isLoggedIn =()=> {
				
				loggedIn[this.ID][user] = args.username ;
				//resp.writeHead(200,{"Content-Type":"text/plain"}) ;
				//resp.write("You are logged in...") ;
				//resp.end() ;
				procUpdate.once(`added-${user}`, ()=>{
					
					resolve( [false,"redirect",(this.https?"https://":"http://") + removeCommonPorts(this.loginRedirect, this.https)]) ;
					
				}) ;
				
				process.send(["proc","add",this.ID,user,args.username]) ;
				
			} ;
			
			if (typeof args.username === "undefined" || typeof args.password === "undefined") {
				
				sendIUOP() ;
				resolve([false,"redirect",(this.https?"https://":"http://") + removeCommonPorts(this.specialPagesP.loginPage, this.https)]) ;
				
			}
			
			else if (typeof this.accounts[args.username] === "undefined") {
				
				sendIUOP() ;
				resolve([false,"redirect",(this.https?"https://":"http://") + removeCommonPorts(this.specialPagesP.loginPage, this.https)]) ;
				
			}
			
			else if (this.accounts[args.username].password === args.password) {
				
				isLoggedIn() ;
				//resolve([true]) ;
				
			}
			
			else {
				
				sendIUOP() ;
				resolve([false,"redirect",(this.https?"https://":"http://") + removeCommonPorts(this.specialPagesP.loginPage, this.https)]) ;
				
			}
			
		});});
		
	}
	
	logout (req,resp,user) {
		
		return new Promise(resolve=>{
			
			delete loggedIn[this.ID][user] ;
			
			procUpdate.once(`deled-${user}`, ()=>{
				
				resolve([false,"redirect",(this.https?"https://":"http://") + this.specialPagesP.logoutPage]) ;
				
			}) ;
			
			process.send(["proc","del",this.ID,user]) ;
			
		}) ;
		
	}
	
	register () {
		
		//Coming soon.
		//To a cinima near you.
		
	}
	
}

function removeCommonPorts(path, https) {
	path = path.split("/") ;
	path[0] = path[0].split(":") ;
	if (path[path.length-1] === https?"443":"80") {
		path[0].pop() ;
	}
	path[0] = path[0].join(":") ;
	path = path.join("/") ;
	return path ;
}

module.exports.proc = proc ;