package mongosearch

var mapFuncImmediate = `function() { emit(this._id, {}) }`

var mapFunc = `
function() {
	// Checks to see if the phrase (array of single words) exists in the
	// all-words array of the object
	var hasPhrase = function(phrase, all) {
		var phraseLen = phrase.length
		for (var i = 0; i + phraseLen < all.length; i++) {
			// See if phrase exists at all
			idx = all.indexOf(phrase[0], i)
			if (idx == -1) {
				return false
			}
			i += idx

			// Check to see if the phrase matches
			var slice = all.slice(i, i + phraseLen)
			for (var a = 0; a < phraseLen; a++) {
				if (slice[a] != phrase[a]) {
					return false
				}
			}
			return true
		}
		return false
	}

	// Walks over every phrase and tests its presence in the all-words array
	var boolPhrases = function(o, text) {
		for (var k in o) {
			for (var i = 0; i < o[k].length; i++) {
				var v = o[k][i]
				switch (typeof v) {
				case "string":
					o[k][i] = hasPhrase(v.split(/ /), text)
					break
				case "object":
					boolPhrases(o[k][i], text)
				}
			}
		}
	}

	// Parses all the boolean operations from boolPhrases and turns it into
	// a single boolean value
	var boolResult = function(o) {
		var ret = true
		var bools = {}
		for (var k in o) {
			bools[k] = []
			for (var i = 0; i < o[k].length; i++) {
				if (typeof o[k][i] == "object") {
					bools[k][i] = boolResult(o[k][i])
				} else {
					bools[k][i] = o[k][i]
				}
			}
			switch (k) {
			case "or":
				if (bools[k].length == 0) {
					break
				}
				var subRet = bools[k][0]
				for (var i = 1; i < bools[k].length; i++) {
					subRet = subRet || bools[k][i]
				}
				ret = ret && subRet
				break
			case "and":
				if (bools[k].length == 0) {
					break
				}
				var subRet = bools[k][0]
				for (var i = 1; i < bools[k].length; i++) {
					subRet = subRet && bools[k][i]
				}
				ret = ret && subRet
				break
			case "nor":
				if (bools[k].length == 0) {
					break
				}
				var subRet = bools[k][0]
				for (var i = 1; i < bools[k].length; i++) {
					subRet = subRet || bools[k][i]
				}
				ret = ret && !subRet
				break
			}
		}
		return ret
	}

	// Put the funcs to good use
	var all = this.%s

	var o = {
		query: query
	}

	if (caseSensitive) {
		o.result = JSON.parse(JSON.stringify(query))
	} else {
		// lowercase everything
		for (var i = 0; i < all.length; i++) {
			all[i] = all[i].toLowerCase()
		}
		o.result = JSON.parse(JSON.stringify(query).toLowerCase())
	}

	boolPhrases(o.result, all)
	if (boolResult(o.result)) {
		emit(this._id, o)
	}
}
`
