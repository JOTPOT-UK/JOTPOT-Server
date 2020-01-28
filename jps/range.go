package jps

type Range struct {
	Start int64
	End   int64
}

func NewRange(start, end int64) Range {
	return Range{start, end}
}

func (r Range) StartEnd(size int64) (int64, int64) {
	r = r.Abs(size)
	return r.Start, r.End
}

func (r Range) Abs(size int64) Range {
	if r.Start < 0 {
		r.Start += size
	}
	if r.End < 0 {
		r.End += size
	}
	return r
}

func (r Range) IsAbsolute() bool {
	return r.Start >= -1 && r.End >= 0
}

//Length returns the length of the range.
//If the length is dependent on the size, then the sizeMul != 0 and the length is `l+sizeMul*size`
func (r Range) Length() (l int64, sizeMul int64) {
	l = r.End - r.Start + 1
	if r.Start < 0 {
		if r.End > 0 {
			sizeMul = -1
		}
	} else if r.End < 0 {
		sizeMul = 1
	}
	return
}

func (r Range) AbsLength(size int64) int64 {
	l, sizeMul := r.Length()
	return l + sizeMul*size
}
