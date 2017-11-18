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

const {Transform} = require("stream") ;

const openCB = Buffer.from("{") ;
const dblOpenCB = Buffer.from("{{") ;
//const closeCB = Buffer.from("}") ;
const dblCloseCB = Buffer.from("}}") ;

class addVars extends Transform {
	//It will add both sets of vars (in for {key: value}), however for vars with the
	// same keys, hpVars overwrites lpVars
	constructor(hpVars, lpVars) {
		super() ;
		//Create a vars object, hpVars take the highest priority
		this.vars = new Object() ;
		Object.assign(this.vars, lpVars) ;
		Object.assign(this.vars, hpVars) ;
		//Init variables
		this.data = Buffer.alloc(0) ;
		this.open = false ;
	}
	willDoAnything() {
		for (let v in this.vars) {
			if (this.vars[v]) {
				return true ;
			}
		}
		return false ;
	}
	//For detecting an opening
	stage1() {
		let openIndex = this.data.indexOf(dblOpenCB) ;
		//If we aren't opening a variable, push the data we have.
		if (openIndex === -1) {
			//If the last character is {, then we need to hold on to it
			if (this.data[this.data.length-1] === openCB[0]) {
				this.push(this.data.slice(0, this.data.length - 1)) ;
				this.data = Buffer.from(openCB) ;
				return ;
			}
			this.push(this.data) ;
			this.data = Buffer.alloc(0) ;
			return ;
		}
		//Push and remove everything up until we open the var
		this.push(this.data.slice(0, openIndex)) ;
		this.data = this.data.slice(openIndex + 2, this.data.length) ;
		this.open = true ;
		this.stage2() ;
	}
	//For detecting the close and adding the var
	stage2() {
		let closeIndex = this.data.indexOf(dblCloseCB, 0) ;
		//We need more data if it hasn't ended
		if (closeIndex === -1) {
			return ;
		}
		//This should be the variable name
		let theVar = this.data.slice(0, closeIndex) ;
		let theVarString = theVar.toString() ;
		//If we have something to insert
		if (typeof this.vars[theVarString] !== "undefined") {
			let toPush ;
			try {
				//Convert it to a buffer
				toPush = Buffer.from(this.vars[theVarString]) ;
			} catch (err) {
				console.warn("Unable to insert variable", theVarString) ;
				console.warn(err.stack) ;
			}
			//Push it
			this.push(toPush) ;
		} else {
			//Otherwise, inser what the syntax would have been
			this.push(Buffer.concat([dblOpenCB, theVar, dblCloseCB], dblOpenCB.length + theVar.length + dblCloseCB.length)) ;
		}
		//Remove the variable name and closing syntax from the data
		this.data = this.data.slice(closeIndex + 2, this.data.length) ;
		//Carry on
		this.open = false ;
		this.stage1() ;
	}
	_transform(chunk, encoding, callback) {
		//Add the data to our buffer
		this.data = Buffer.concat([this.data, chunk], this.data.length + chunk.length) ;
		//Call the correct function
		if (this.open) {
			this.stage2() ;
		} else {
			this.stage1() ;
		}
		//OK, we can take more data
		callback() ;
	}
}

module.exports = addVars ;