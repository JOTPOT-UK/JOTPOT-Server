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

const path = require("path") ;
const fs = require("fs") ;

//Not implemented: link, linkSync, lstat, lstatSync, mkdtemp, mkdtempSync, readlink, readlinkSync, symlink, symlinkSync
const pathFuncs = [
	"access",
	"accessSync",
	"appendFile",
	"appendFileSync",
	"createReadStream",
	"createWriteStream",
	"exists",
	"existsSync",
	"mkdir",
	"mkdirSync",
	"mkdirSync",
	"readFile",
	"readFileSync",
	"rmdir",
	"rmdirSync",
	"stat",
	"statSync",
	"truncate",
	"truncateSync",
	"unlink",
	"unlinkSync",
	"unwatchFile",
	"utimes",
	"utimesSync",
	"watch",
	"watchFile",
	"writeFile",
	"writeFileSync"
] ;
const fdFuncs = [
	"fdatasync",
	"fdatasyncSync",
	"fstat",
	"fstatSync",
	"fsync",
	"fsyncSync",
	"ftruncate",
	"ftruncateSync",
	"futimes",
	"futimesSync",
	"futimesSync",
	"read",
	"readSync",
	"write",
	"writeSync"
] ;
const otherVars = [
	"constants"
] ;

//False fs class
class FileSystem {
	constructor (dirs) {
		this.ownedFDs = new Array() ;
		//Faster creation for common setup
		if (typeof dirs === "string") {
			this.fdDirs = {
				root: `./extensions/${dirs}`,
				mounts: {
					webspace: `./sites/${dirs}`
				}
			} ;
		} else {
			this.fdDirs = dirs || {
				mounts: {}
			} ;
			//Check that mounts is an object
			if (typeof this.fdDirs.mounts !== "object") {
				throw new Error("mounts must be an object") ;
			}
		}
		//Get the real root path
		this.rootPath = this.getPath("/") ;
		this.public = {
			open: (...args) => {
				//Rewrite the path
				args[0] = this.getPath(args[0]) ;
				//Get the callback, and rewrite the argument
				let callback = args.pop() ;
				args.push((err, fd) => {
					//Push it as an owned fd if there wasnt an error
					if (!err) {
						this.ownedFDs.push(fd) ;
					}
					callback(err, fd) ;
				}) ;
				//Call open
				fs.open(...args) ;
			},
			openSync: (...args) => {
				args[0] = this.getPath(args[0]) ;
				let fd = fs.openSync(...args) ;
				this.ownedFDs.push(fd) ;
				return fd ;
			},
			close: (...args) => {
				//Check we own it
				if (this.ownedFDs.indexOf(args[0]) === -1) {
					throw "You can only use FDs that you have opened." ;
				} else {
					//We no longer own the FD we are closing
					this.ownedFDs.splice(this.ownedFDs.indexOf(args[0]), 1) ;
					return fs.close(...args) ;
				}
			},
			closeSync: (...args) => {
				//Check we own it
				if (this.ownedFDs.indexOf(args[0]) === -1) {
					throw "You can only use FDs that you have opened." ;
				} else {
					//We no longer own the FD we are closing
					this.ownedFDs.splice(this.ownedFDs.indexOf(args[0]), 1) ;
					return fs.closeSync(...args) ;
				}
			},
			rename: (...args) => {
				args[0] = this.getPath(args[0]) ;
				args[1] = this.getPath(args[1]) ;
				return fs.rename(...args) ;
			},
			renameSync: (...args) => {
				args[0] = this.getPath(args[0]) ;
				args[1] = this.getPath(args[1]) ;
				return fs.renameSync(...args) ;
			},
			realpath: (p, ...args) => {
				p = path.normalize("/"+p) ;
				process.nextTick(()=>{
					(args[1]||args[0])(null, p) ;
				}) ;
			},
			realpathSync: (p, opts) => {
				//If the 'real' option is true, then use getPath
				if (typeof opts === "object") {
					if (opts.real) {
						return this.getPath(path.normalize("/"+p)) ;
					}
				}
				return path.normalize("/"+p) ;
			},
			readdir: (...args) => {
				args[0] = this.getPath(args[0]) ;
				//If we are in the root path
				if (args[0] === this.rootPath) {
					//Store the origional callback and rewrite it
					let origCB = args[args.length===2?1:2] ;
					args[args.length===2?1:2] = (err, dir) => {
						if (!err) {
							//Concat the mounts before the callback
							dir = dir.concat(Object.keys(this.fdDirs.mounts)) ;
						}
						origCB(err, dir.sort()) ;
					} ;
					fs.readdir(...args) ;
				} else {
					fs.readdir(...args) ;
				}
			},
			readdirSync: (...args) => {
				args[0] = this.getPath(args[0]) ;
				let dir = fs.readdirSync(...args) ;
				//If we are in the root, add the mounts
				if (args[0] === this.rootPath) {
					dir = dir.concat(Object.keys(this.fdDirs.mounts)) ;
				}
				//Return a sorted output
				return dir.sort() ;
			}
		} ;
		//Add the functions and constants
		for (let doing in pathFuncs) {
			this.createPublic(pathFuncs[doing]) ;
		}
		for (let doing in fdFuncs) {
			this.createPublicFD(fdFuncs[doing]) ;
		}
		for (let doing in otherVars) {
			fdFuncs[otherVars[doing]] = new Object() ;
			Object.assign(fdFuncs[otherVars[doing]], fs[otherVars[doing]]) ;
		}
	}
	//Function to get real path
	getPath(p) {
		//Split thr path up
		p = path.normalize(p).split(path.sep) ;
		//Remove the start if we are at the root
		while (p[0] === "") {
			p.shift() ;
		}
		let root = this.fdDirs.root || "/" ;
		//Go through he mounts
		for (let doing in this.fdDirs.mounts) {
			//If we are in a mount, change the root to that of the mount and change to relitive path
			if (p[0] === doing) {
				root = this.fdDirs.mounts[doing] ;
				p.splice(0, 1) ;
				break ;
			}
		}
		//Join the root and the path
		return path.join(root, ...p) ;
	}
	createPublic(func) {
		this.public[func] = (...args) => {
			//Rewrite the path to the real path
			args[0] = this.getPath(args[0]) ;
			return fs[func](...args) ;
		} ;
	}
	createPublicFD(func) {
		this.public[func] = (...args) => {
			//Check we own the FD, throw if not
			if (this.ownedFDs.indexOf(args[0]) === -1) {
				throw new Error("You can only use FDs that you have opened.") ;
			} else {
				fs[func](...args) ;
			}
		} ;
	}
}

module.exports = FileSystem ;
