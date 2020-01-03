package jps

type Range struct {
	Start int64
	End   int64
}

func (r Range) StartEnd(length int64) (int64, int64) {
	start := r.Start
	end := r.End
	if start < 0 {
		start += length
	}
	if end < 0 {
		end += length
	}
	return start, end
}
