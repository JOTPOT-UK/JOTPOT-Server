const fs = require("fs") ;

function formatObject(file) {
	
	let docs = JSON.parse(fs.readFileSync(file).toString()) ;
	let out = new String() ;
	for (let t of docs) {
		
		let supportList = new String() ;
		for (let doing in t.support) {
			
			supportList += `<li>${["Master","Worker","Limited master","Limited worker"][doing]} extention: <span class="${t.support[doing]?"supported":"notsupported"}">${t.support[doing]?"Yes":"No"}</span></li>` ;
			
		}
		let argsList = new String() ;
		for (let doing in t.args) {
			
			argsList += `<li>${t.args[doing]}</li>` ;
			
		}
		out += `<div ID="${t.id}"><h3 class="func-h3">${t.title}</h3><div class="details">${t.support?`Availability:<ul>${supportList}</ul>`:''}Added: ${t.added}${t.extra?`<br>${t.extra}`:""}</div><ul class="args-list">${argsList}</ul><div class="desc">${t.desc}</div></div>` ;
		
	}
	return eval(`\`${out}\``)
	
}

fs.writeFileSync("out.html",eval(`\`${fs.readFileSync("template.html").toString()}\``)) ;