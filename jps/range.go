package jps

import "strconv"

//RangeAll returns a Range which covers all the data.
//It ranges from the first byte (0), to the last byte (-1).
func RangeAll() Range { return Range{0, -1} }

//Range represents a byte range from Start to End (inclusive).
//If byte values are posative, then they are absolute offsets with 0 being the first byte of the file.
//Negative byte index's are from the end of the file, with -1 being the last byte, and -2 being the second to last byte etc.
type Range struct {
	Start int64
	End   int64
}

//NewRange returns a Range struct with the given start and end indexs.
//Note that Range values are inclusive.
func NewRange(start, end int64) Range {
	return Range{start, end}
}

//StartEnd returns the start and end byte indexs (inclusive) for a file with the given size.
//Aka, if either r.Start or r.End is negative, the size is added to them.
//Also see r.Abs, which returns a Range type given a size.
func (r Range) StartEnd(size int64) (int64, int64) {
	r = r.Abs(size)
	return r.Start, r.End
}

//Abs returns a new Range struct with absolute (non-negative) byte indexs.
//These are calculated using the given size - aka, if either r.Start or r.End is negative, the size is added to them.
//This function is used for r.StartEnd(size).
func (r Range) Abs(size int64) Range {
	if r.Start < 0 {
		r.Start += size
	}
	if r.End < 0 {
		r.End += size
	}
	return r
}

//IsAbsolute returns true if the Range is absolute.
//A Range is absolute when it does not depend on the size of the data it is a range of - aka when neither r.Start nor r.End are nagative.
//r.Abs(size) can be used to generate an absolute byte range.
func (r Range) IsAbsolute() bool {
	return r.Start >= 0 && r.End >= 0
}

//Length returns the length of the range.
//If the length is dependent on the size of the data to range over, then the sizeMul != 0 and the length is `l+sizeMul*size`
//See r.AbsLength(size) if you already know the size of the data.
func (r Range) Length() (l int64, sizeMul int64) {
	/* The cases are as follows:
	 * r.Start > 0 && r.End > 0 (the range is absolute):
	 *     l = r.End - r.Start + 1; sizeMul = 0
	 * r.Start < 0 && e.End < 0 (the range is 'absolute from the end' - both offsets are fixed from the end of the data):
	 *     l = r.End - r.Start + 1; sizeMul = 0
	 * r.Start > 0 && e.End < 0 (a fixed start byte, and a end byte dependent on the length):
	 *     l = (r.End + size) - r.Start + 1
	 *     => l = r.End - r.Start + 1; sizeMul = 1
	 * r.Start < 0 && r.End > 0 (I don't know why you'd want this type of range...):
	 *     l = r.End - (r.Start + size) + 1
	 *     => l = r.End - r.Start + 1; sizeMul = -1
	 */
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

func (r Range) isNoLength() bool {
	if r.Start < 0 && r.End > 0 {
		return false
	}
	if r.End < 0 {
		return false
	}
	return r.End-r.Start < 0
}

//AbsLength : if r is used on data with size `size`, then r.AbsLength(size) returns the abount of bytes within the range r.
//l, sizeMul := r.Length(); return l + sizeMul*size  (see r.Length())
func (r Range) AbsLength(size int64) int64 {
	//Warning: this function works on the basis that Length is defined to return the value which makes this expression correct;
	// you should probably not change this unless you are changing r.Length().
	l, sizeMul := r.Length()
	return l + sizeMul*size
}

func (r Range) String() string {
	return "Range{ " + strconv.FormatInt(r.Start, 10) + " - " + strconv.FormatInt(r.End, 10) + " }"
}

//SortRanges sorts the given ranges (in place) in assending order of their Start value.
func SortRanges(rs []Range) {
	for i := 1; i < len(rs); i++ {
		r := rs[i]
		for j := i - 1; 0 <= j && r.Start < rs[j].Start; j-- {
			rs[j+1] = rs[j]
			rs[j] = r
		}
	}
}

//MergeRanges, for all i, merges range i with i+1 if they overlap.
//This doesn't merge overlaping ranges if there is a non-overlaping range between them. This is
// to allow for data requested first to be delivered first, as noted in the HTTP spec.
//If you don't care about the order of ranges, which is permittable, then sorting before calling
// this function will merge all overlaping ranges.
//This does merging in place, then returns a slice of the origional slice. So the slice passed
// should not be used afterwards.
func MergeRanges(rs []Range) []Range {
	if len(rs) == 0 {
		return rs
	}
	//Remove the first element if it's a 0 length range.
	r1 := rs[0]
	if r1.isNoLength() {
		return MergeRanges(rs[1:])
	}
	for i := 1; i < len(rs); {
		//We will uphold r1 = rs[i-1]; r2 = rs[i]
		r2 := rs[i]
		if r2.Start >= 0 {
			if r2.End >= 0 {
				if r2.End < r2.Start {
					// Remove <=0 length ranges
					rs = rs[:i+copy(rs[i:], rs[i+1:])]
					continue
				}
				if r1.Start >= 0 {
					// Both have absolute starting positions
					if r1.Start <= r2.Start {
						// r1 comes first
						if r2.Start <= r1.End {
							// They overlap!
							// Merge r2 in to r1
							if r1.End < r2.End {
								r1.End = r2.End
								rs[i-1].End = r2.End
							}
							// Remove r2
							rs = rs[:i+copy(rs[i:], rs[i+1:])]
							continue
						} // else they don't overlap
					} else {
						// r2 comes first
						if r1.Start <= r2.End {
							// They overlap...
							// So merge r2 to r1
							r1.Start = r2.Start
							if r1.End < r2.End {
								r1.End = r2.End
							}
							rs[i-1] = r1
							// Remove r2
							rs = rs[:i+copy(rs[i:], rs[i+1:])]
							continue
						} // else they don't overlap
					}
				}
			}
		} // for now, we only merge absolute ranges TODO: check End?
		r1 = r2
		i++
	}
	return rs
}

//TidyRanges uses SortRanges then MergeRanges on rs, producing a list of sorted and non-overlaping
// ranges.
//This algorithm modifies the origional slice so the slice passed should not be used afterwards.
func TidyRanges(rs []Range) []Range {
	SortRanges(rs)
	return MergeRanges(rs)
}
