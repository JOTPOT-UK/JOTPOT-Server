let readable = require("stream").Readable ;

class Data extends readable {
	
	constructor (name,headers) {
		
		super() ;
		this.name = name ;
		this.headers = headers ;
		
	}
	
	_read(s) {
		
		//console.log("Wait, am I supposed to do somthing?") ;
		
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
	let output = new Object() ;
	let stage = 0 ;
	let fields = new Object() ;
	let newLineMade = false ;
	let currentHeaders = new Object() ;
	let currentDataPipe = null ;
	const parseTick =_=> {
		
		if (dataString.indexOf(boundary) === 0 && stage === 0) {
			
			dataString = dataString.substring(boundary.length,dataString.length) ;
			data = Buffer.from(dataString) ;
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
				data = Buffer.from(dataString) ;
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
						
						name = currentHeaders["content-disposition"].match(/; name="(\\"|[^"])*"/g)[0].substring(8,currentHeaders["content-disposition"].match(/; name="(\\"|[^"])*"/g)[0].length-1) ;
						console.log(name) ;
						
					}
					currentDataPipe = new Data(name,currentHeaders) ;
					cb(currentDataPipe) ;
					return parseTick() ;
					
				}
				else {
					
					//console.log("Header:",thisBit) ;
					thisBit = thisBit.split(": ") ;
					currentHeaders[thisBit[0].toLowerCase()] = thisBit[1] ;
					
				}
				
			} while (1)
			
		}
		
		else if (stage === 2) {
			
			let thisData ;
			if (data.indexOf(boundaryB) !== -1) {
				
				thisData = data.slice(0,data.indexOf(boundaryB)) ;
				data = data.slice(data.indexOf(boundaryB),data.length) ;
				dataString = data.toString() ;
				stage = 0 ;
				
			}
			else {
				
				thisData = data.slice(0,data.length) ;
				data = Buffer.alloc(0) ;
				dataString = "" ;
				
			}
			console.log("Data:",thisData.toString()) ;
			currentDataPipe.push(thisData) ;
			if (!stage) {
				
				return parseTick() ;
				
			}
			return false ;
			
		}
		
	} ;
	req.on("data",d=>{
		
		//console.log(d.toString()) ;
		data = Buffer.concat([data,d]) ;
		dataString += d.toString() ;
		endReady = parseTick() ;
		if (endReady) {console.log("End ready")} ;
		
	}) ;
	req.on("end",_=>{
		
		resolve() ;
		
	}) ;
	
}) ;