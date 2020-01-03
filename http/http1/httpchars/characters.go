package httpchars

func IsWhitespace(b byte) bool {
	return b == ' ' || b == '\t'
}

func IsDigit(b byte) bool {
	return '0' <= b && b <= '9'
}

func IsAlpha(b byte) bool {
	return ('A' <= b && b <= 'Z') || ('a' <= b && b <= 'z')
}

func IsVChar(b byte) bool {
	return 0x21 <= b && b <= 0x7E
}

func IsOBSText(b byte) bool {
	return 0x80 <= b && b <= 0xFF
}

func IsFieldChar(b byte) bool {
	return IsOBSText(b) || IsVChar(b)
}

func IsQuotedPairChar(b byte) bool {
	/*
		RFC 7230, section 3.2.6 says:

		A sender SHOULD NOT generate a quoted-pair in a quoted-string except
		where necessary to quote DQUOTE and backslash octets occurring within
		that string.

		Therefore, for efficiency, we will check those 2 characters first.
	*/
	return b == '"' || b == '\\' || b == ' ' || b == '\t' || IsVChar(b) || IsOBSText(b)
}

func IsQDText(b byte) bool {
	return (0x21 <= b && b <= 0x7E && b != '\\' && b != '"') ||
		b == ' ' || b == '\t' ||
		IsOBSText(b)
}

func IsTokenChar(b byte) bool {
	//35 is #
	return IsDigit(b) || IsAlpha(b) || (35 <= b && b <= '\'') || b == '!' || b == '*' || b == '+' || b == '-' || b == '.' || b == '^' || b == '_' || b == '`' || b == '|' || b == '~'
}

func IsValidToken(bytes []byte) bool {
	for _, b := range bytes {
		if !IsTokenChar(b) {
			return false
		}
	}
	return true
}

func IsValidTokenString(s string) bool {
	for _, r := range s {
		if r > 255 {
			return false
		}
		b := byte(r)
		if !IsTokenChar(b) {
			return false
		}
	}
	return true
}

func RemoveWhitespacePrefix(v string) string {
	var i int
	for IsWhitespace(v[i]) {
		i++
	}
	return v[i:]
}

func RemoveWhitespacePostfix(v string) string {
	i := len(v) - 1
	for IsWhitespace(v[i]) {
		i--
	}
	return v[:i+1]
}

//RemoveWhitespace takes a string, and returns the same string, but with any preceding or trailing tabs or spaces removed.
func RemoveWhitespace(v string) string {
	return RemoveWhitespacePostfix(RemoveWhitespacePrefix(v))
}
