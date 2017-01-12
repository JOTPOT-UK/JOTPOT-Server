let fs = require("fs") ;
let Events = require("events") ;

class NewEmitter extends Events {}
const procUpdate = new NewEmitter() ;
let loggedIn = new Object() ;

process.on("message",toDo=>{
	
	if (toDo[0] === "proc-update") {
		
		loggedIn = toDo[1] ;
		procUpdate.emit("update") ;
		
	}
	
}) ;

function getUserID(req) {
	
	if (typeof req.headers.cookie === "undefined") {
		
		return false ;
		
	}
	
	let cookieArray = req.headers.cookie.split("; ") ;
	let cookies = new Object() ;
	for (let doing in cookieArray) {
		
		cookies[cookieArray[doing].split("=")[0]] = cookieArray[doing].split("=")[1] ;
		
	}
	
	if (typeof cookies.JOTPOTUID === "undefined") {
		
		return false ;
		
	}
	
	return `${req.headers['x-forwarded-for'] || req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress}(${req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress})---${cookies.JOTPOTUID}` ;
	
}

function makeNewUID(req,resp) {
	
	let newUID = Math.random() + Math.random() + Math.random() + Math.random() + Math.random() + Math.random() + Math.random() + Math.random() + Math.random() + Math.random() ;
	newUID *= 10e10 ;
	newUID *= Math.random() * Math.random() * Math.random() ;
	newUID *= 10e30 ;
	newUID *= Math.random() * Math.random() * Math.random() ;
	newUID *= 10e30 ;
	newUID += Math.random() * 10e5 ;
	newUID += Math.random() * 10e5 ;
	newUID += Math.random() * 10e5 ;
	resp.setHeader("Set-Cookie","JOTPOTUID=" + Math.round(newUID).toString(36)) ;
	return `${req.headers['x-forwarded-for'] || req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress}(${req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress})---${newUID}` ; ;
	
}

class proc {
	
	constructor (name,db,pages,login,loginPage,logout,logoutPage,reg,regPage) {
		
		this.accounts = JSON.parse(fs.readFileSync(db).toString()) ;
		this.pages = new Array() ;
		this.starts = new Array() ;
		for (let doing in pages) {
			
			if (pages[doing].lastIndexOf("*") === pages[doing].length - 1) {
				
				this.starts.push(pages[doing].substring(0,pages[doing].length - 1)) ;
				
			}
			
			else {
				
				this.pages.push(pages[doing]) ;
				
			}
			
		}
		this.specialPages = new Object() ;
		this.specialPagesP = new Object() ;
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
		this.doAnything = this.doAnything.bind(this) ;
		loggedIn[name] = new Object() ;
		this.ID = name ;
		process.send(["proc","new",name]) ;
		
	}
	
	doAnything(req,resp) {
		
		return new Promise((resolve,regect) => {
			
			let page = req.url ;
			
			if (typeof this.specialPagesP[page] !== "undefined") {
				
				resolve([true]) ;
				return ;
				
			}
			
			let shouldCheckBlock = true ;
			if (typeof this.specialPages[page] !== "undefined" ) {
				
				shouldCheckBlock = false ;
				
			}
			
			if (shouldCheckBlock) {
				
				let isAuthed = true ;
				
				if (this.pages.indexOf(page) !== -1) {
					
					isAuthed = false ;
					
				}
				
				if (isAuthed) {for (let doing in this.starts) {
					
					if (page.indexOf(this.starts[doing]) === 0) {
						
						isAuthed = false ;
						break ;
						
					}
					
				}}
				
				if (isAuthed) {
					
					resolve([true]) ;
					return ;
					
				}
				
			}
			
			let user = getUserID(req) ;
			if (user === false) {
				
				user = makeNewUID(req,resp) ;
				
			}
			
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
					return ;
					
				}
				
				else {
					
					resolve([false,"redirect","http://" + this.specialPagesP["loginPage"]]) ;
					return ;
					
				}
				
			}) ;
			
		}) ;
		
	}
	
	isAuthed (user) {
		
		return new Promise((resolve,reject) => {
			
			if (typeof loggedIn[this.ID][user] === "undefined") {
				
				process.send(["proc","get"]) ;
				procUpdate.once("update",_=>{
					
					if (typeof loggedIn[this.ID][user] === "undefined") {
						
						resolve(false) ;
						
					}
					
					else {
						
						resolve(true) ;
						
					}
					
				}) ;
				
			}
			
			else {
				
				resolve(true) ;
				
			}
			
		}) ;
		
	}
	
	login (req,resp,user) {
		
		return new Promise((resolve,reject)=>{req.on("data",(d)=>{
			
			d = decodeURIComponent(decodeURI(d.toString()).replace(/\+/g," ")).split("&") ;
			let args = new Object() ;
			
			for (let doing in d) {
				
				args[d[doing].split("=")[0]] = d[doing].split("=")[1] ;
				
			}
			
			let sendIUOP =_=> {
				
				//resp.writeHead(200,{"Content-Type":"text/plain"}) ;
				//resp.write("Incorrect username or password...") ;
				//resp.end() ;
				resolve([false,"redirect",this.specialPagesP.loginPage]) ;
				
			}
			
			let isLoggedIn =_=> {
				
				loggedIn[this.ID][user] = args.username ;
				process.send(["proc","add",this.ID,user,args.username]) ;
				//resp.writeHead(200,{"Content-Type":"text/plain"}) ;
				//resp.write("You are logged in...") ;
				//resp.end() ;
				resolve( [false,"redirect","/"]) ;
				
			}
			
			if (typeof args.username === "undefined" || typeof args.password === "undefined") {
				
				sendIUOP() ;
				resolve([false,"redirect",this.specialPagesP.loginPage]) ;
				
			}
			
			else if (typeof this.accounts[args.username] === "undefined") {
				
				sendIUOP() ;
				resolve([false,"redirect",this.specialPagesP.loginPage]) ;
				
			}
			
			else if (this.accounts[args.username].password === args.password) {
				
				isLoggedIn() ;
				resolve([true]) ;
				
			}
			
			else {
				
				sendIUOP() ;
				resolve([false,"redirect",this.specialPagesP.loginPage]) ;
				
			}
			
		}) ;}) ;
		return [null] ;
		
	}
	
	logout (req,resp,user) {
		
		delete loggedIn[this.ID][user] ;
		process.send(["proc","del",this.ID,user]) ;
		return [false,"redirect",this.specialPagesP.logoutPage] ;
		
	}
	
	register (req,resp,user) {
		
		//Coming soon.
		
	}
	
}

module.exports = proc ;