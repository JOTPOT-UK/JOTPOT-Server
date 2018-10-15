package varify

func IsDigit(b byte) bool {
	return '0' <= b && b <= '9'
}

func IsAlpha(b byte) bool {
	return ('A' <= b && b <= 'Z') || ('a' <= b && b <= 'z')
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
