const fs = require("fs") ;
const path = require("path") ;
const parseFlags = require("./flag-parser") ;

const flags = parseFlags() ;
if (!flags["-out"]) {
	throw new Error("-out option required") ;
} else if (typeof flags["-out"][0] !== "string") {
	throw new Error("-out option must be a string") ;
}
const out = flags["-out"][0] ;
if (flags["-in"] && typeof flags["-in"][0] !== "string") {
	throw new Error("-in option must be a string") ;
}
const src = (flags["-in"] || [process.cwd()])[0] ;

let config ;
function loadConfig(path) {
	if (fs.existsSync(path)) {
		config = JSON.parse(fs.readFileSync(path).toString()) ;
		return true ;
	}
	return false ;
}
if (!loadConfig(path.join(src, "build", "buildconfig.json"))) {
	if (!loadConfig(path.join(src, "build", "new", "buildconfig.json"))) {
		throw new Error("Cannot find buildconfig.json") ;
	}
}

function createDir(...paths) {
	const p = path.join(...paths) ;
	if (!fs.existsSync(p)) {
		createDir(path.dirname(p)) ;
		fs.mkdirSync(p) ;
	} else {
		if (fs.statSync(p).isFile()) {
			throw new Error(`${p} is a file.`) ;
		}
	}
}

createDir(out, "daemon", "jps") ;

fs.createReadStream(path.join(src, "util", "jps.go")).pipe(fs.createWriteStream(path.join(out, "jps.go"))) ;
fs.createReadStream(path.join(src, "util", "daemon", "jpsd.go")).pipe(fs.createWriteStream(path.join(out, "daemon", "jpsd.go"))) ;

for (let copyer of config["jps-files"]) {
	fs.createReadStream(path.join(src, copyer)).pipe(fs.createWriteStream(path.join(out, "daemon", "jps", copyer))) ;
}
