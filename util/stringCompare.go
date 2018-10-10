package util

type ComparisonResult byte

const (
	ComparisonEqual ComparisonResult = iota
	ComparisonLessThan
	ComparisonGreaterThan
)

func CompareStrings(a, b string) ComparisonResult {
	l := MinInt(len(a), len(b))
	for i := 0; i < l; i++ {
		if a[i] < b[i] {
			return ComparisonLessThan
		} else if a[i] > b[i] {
			return ComparisonGreaterThan
		}
	}
	if len(a) < len(b) {
		return ComparisonLessThan
	} else if len(a) > len(b) {
		return ComparisonGreaterThan
	}
	return ComparisonEqual
}
