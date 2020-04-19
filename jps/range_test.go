package jps

import (
	"reflect"
	"testing"
)

func TestNewRange(t *testing.T) {
	if NewRange(0, 0) != (Range{
		Start: 0,
		End:   0,
	}) {
		t.Error("NewRange(0,0) failed")
	} else if NewRange(0, 2) != (Range{
		Start: 0,
		End:   2,
	}) {
		t.Error("NewRange(0,2) failed")
	} else if NewRange(1, 2) != (Range{
		Start: 1,
		End:   2,
	}) {
		t.Error("NewRange(1,2) failed")
	} else if NewRange(-1, 2) != (Range{
		Start: -1,
		End:   2,
	}) {
		t.Error("NewRange(-1,2) failed")
	} else if NewRange(-1, -2) != (Range{
		Start: -1,
		End:   -2,
	}) {
		t.Error("NewRange(-1,-2) failed")
	}
}

func TestRangeIsAbsolute(t *testing.T) {
	if !NewRange(0, 0).IsAbsolute() {
		t.Error("Range{0,0} is not absolute?")
	} else if !NewRange(0, 6).IsAbsolute() {
		t.Error("Range{0,6} is not absolute?")
	} else if !NewRange(2, 6).IsAbsolute() {
		t.Error("Range{2,6} is not absolute?")
	} else if !NewRange(2, 0).IsAbsolute() {
		t.Error("Range{2,0} is not absolute?")
	} else if NewRange(-1, 6).IsAbsolute() {
		t.Error("Range{-1,6} is absolute?")
	} else if NewRange(-10, 6).IsAbsolute() {
		t.Error("Range{-10,6} is absolute?")
	} else if NewRange(-10, -6).IsAbsolute() {
		t.Error("Range{-10,-6} is absolute?")
	} else if NewRange(0, -6).IsAbsolute() {
		t.Error("Range{0,-6} is absolute?")
	} else if NewRange(0, -1).IsAbsolute() {
		t.Error("Range{0,-1} is absolute?")
	} else if NewRange(-10, -4).IsAbsolute() {
		t.Error("Range{-10,-4} is absolute?")
	}
}

func TestRangeLength(t *testing.T) {
	if l, m := NewRange(0, 0).Length(); l != 1 || m != 0 {
		t.Error("Range{0, 0}.Length() returned", l, m, "; expecting 1, 0")
	} else if l, m := NewRange(0, 6).Length(); l != 7 || m != 0 {
		t.Error("Range{0, 6}.Length() returned", l, m, "; expecting 7, 0")
	} else if l, m := NewRange(5, 9).Length(); l != 5 || m != 0 {
		t.Error("Range{5, 9}.Length() returned", l, m, "; expecting 5, 0")
	} else if l, m := NewRange(-5, -1).Length(); l != 5 || m != 0 {
		t.Error("Range{-5, -1}.Length() returned", l, m, "; expecting 5, 0")
	} else if l, m := NewRange(-1, -1).Length(); l != 1 || m != 0 {
		t.Error("Range{-1, -1}.Length() returned", l, m, "; expecting 1, 0")
	} else if l, m := NewRange(-6, -3).Length(); l != 4 || m != 0 {
		t.Error("Range{-6, -3}.Length() returned", l, m, "; expecting 4, 0")
	} else if l, m := NewRange(0, -1).Length(); l != 0 || m != 1 {
		t.Error("Range{0, -1}.Length() returned", l, m, "; expecting 0, 1")
	} else if l, m := NewRange(5, -2).Length(); l != -6 || m != 1 {
		t.Error("Range{0, 0}.Length() returned", l, m, "; expecting -6, 1")
	} else if l := NewRange(-8, 20).AbsLength(25); l != 4 {
		t.Error("Range{-8, 20}.AbsLength(25) returned", l, "; expecting 4")
	}
}

func TestRangeAbs(t *testing.T) {
	if NewRange(1, 2) != (Range{1, 2}) {
		t.Error("Equality assumption failed")
	} else if NewRange(1, 2) == (Range{2, 2}) {
		t.Error("Equality assumption failed")
	} else if r := NewRange(0, 5).Abs(19); r != NewRange(0, 5) {
		t.Error("Range{0, 5}.Abs(19) =", r, "expecting Range{0, 5}")
	} else if r := NewRange(0, -1).Abs(19); r != NewRange(0, 18) {
		t.Error("Range{0, -1}.Abs(19) =", r, "expecting Range{0, 18}")
	} else if r := NewRange(-10, 26).Abs(30); r != NewRange(20, 26) {
		t.Error("Range{-10, 26}.Abs(30) =", r, "expecting Range{20, 26}")
	} else if r := NewRange(-11, -3).Abs(219); r != NewRange(208, 216) {
		t.Error("Range{-11, -3}.Abs(219) =", r, "expecting Range{208, 216}")
	}
}

func tSortRanges(rs []Range) []Range {
	SortRanges(rs)
	return rs
}

func TestSortRanges(t *testing.T) {
	if !reflect.DeepEqual(tSortRanges([]Range{}), []Range{}) {
		t.Error("Failed to sort 0 ranges.")
	} else if !reflect.DeepEqual(tSortRanges([]Range{NewRange(1, 2)}), []Range{NewRange(1, 2)}) {
		t.Error("Failed to sort 1 range.")
	} else if !reflect.DeepEqual(tSortRanges([]Range{NewRange(1, 5), NewRange(2, 3)}), []Range{NewRange(1, 5), NewRange(2, 3)}) {
		t.Error("Failed to sort 2 in order ranges.")
	} else if !reflect.DeepEqual(tSortRanges([]Range{NewRange(2, 5), NewRange(1, 3)}), []Range{NewRange(1, 3), NewRange(2, 5)}) {
		t.Error("Failed to sort 2 out of order ranges.")
	} else if !reflect.DeepEqual(
		tSortRanges([]Range{NewRange(0, 5), NewRange(3, 3), NewRange(-6, 10), NewRange(11, -1), NewRange(2, 50)}),
		[]Range{NewRange(-6, 10), NewRange(0, 5), NewRange(2, 50), NewRange(3, 3), NewRange(11, -1)}) {
		t.Error("SortRanges failed (1)")
	}
}

func TestMergeRanges(t *testing.T) {
	if !reflect.DeepEqual(
		MergeRanges([]Range{}),
		[]Range{},
	) {
		t.Error("Failed to merge no ranges.")
	} else if !reflect.DeepEqual(
		MergeRanges([]Range{NewRange(1, 10)}),
		[]Range{NewRange(1, 10)},
	) {
		t.Error("Failed to merge 1 range.")
	} else if !reflect.DeepEqual(
		MergeRanges([]Range{NewRange(1, 5), NewRange(6, 12), NewRange(15, 100)}),
		[]Range{NewRange(1, 5), NewRange(6, 12), NewRange(15, 100)},
	) {
		t.Error("Failed to not merge disjoint ranges in order.")
	} else if !reflect.DeepEqual(
		MergeRanges([]Range{NewRange(100, 105), NewRange(6, 12), NewRange(15, 100)}),
		[]Range{NewRange(100, 105), NewRange(6, 12), NewRange(15, 100)},
	) {
		t.Error("Failed to not merge disjoint ranges out of order.")
	} else if !reflect.DeepEqual(
		MergeRanges([]Range{NewRange(5, 9), NewRange(9, 15)}),
		[]Range{NewRange(5, 15)},
	) {
		t.Error("Failed to merge 2 non-overlaping ranges.")
	} else if !reflect.DeepEqual(
		MergeRanges([]Range{NewRange(5, 12), NewRange(9, 15)}),
		[]Range{NewRange(5, 15)},
	) {
		t.Error("Failed to merge 2 overlaping ranges.")
	} else if !reflect.DeepEqual(
		MergeRanges([]Range{NewRange(5, 12), NewRange(11, 16), NewRange(16, 80)}),
		[]Range{NewRange(5, 80)},
	) {
		t.Error("Failed to merge all ranges (overlap).")
	} else if !reflect.DeepEqual(
		MergeRanges([]Range{NewRange(5, 12), NewRange(11, 16), NewRange(14, 80)}),
		[]Range{NewRange(5, 80)},
	) {
		t.Error("Failed to merge all ranges (overlap twice).")
	} else if !reflect.DeepEqual(
		MergeRanges([]Range{NewRange(5, 12), NewRange(11, 16), NewRange(17, 80)}),
		[]Range{NewRange(5, 16), NewRange(17, 80)},
	) {
		t.Error("Failed to merge first 2 ranges.")
	} else if !reflect.DeepEqual(
		MergeRanges([]Range{NewRange(5, 12), NewRange(15, 16), NewRange(15, 80)}),
		[]Range{NewRange(5, 12), NewRange(15, 80)},
	) {
		t.Error("Failed to merge final 2 ranges.")
	} else if !reflect.DeepEqual(
		MergeRanges([]Range{NewRange(0, 10), NewRange(11, 20), NewRange(20, 23), NewRange(50, 100)}),
		[]Range{NewRange(0, 10), NewRange(11, 23), NewRange(50, 100)},
	) {
		t.Error("Failed to merge middle ranges.")
	} else if !reflect.DeepEqual(
		MergeRanges([]Range{NewRange(200, 500), NewRange(100, 300), NewRange(600, 700), NewRange(30, 40), NewRange(25, 35)}),
		[]Range{NewRange(100, 500), NewRange(600, 700), NewRange(25, 40)},
	) {
		t.Error("Failed to merge reversed order ranges.")
	} else if !reflect.DeepEqual(
		MergeRanges([]Range{NewRange(8, 12), NewRange(8, 12), NewRange(10, 12), NewRange(8, 10), NewRange(10, 10), NewRange(5, 9), NewRange(12, 50), NewRange(6, 3), NewRange(2, 0)}),
		[]Range{NewRange(5, 50)},
	) {
		t.Errorf("Failed to merge complex ranges, got %v.", MergeRanges([]Range{NewRange(8, 12), NewRange(8, 12), NewRange(10, 12), NewRange(8, 10), NewRange(10, 10), NewRange(5, 9), NewRange(12, 50), NewRange(6, 3), NewRange(2, 0)}))
	} else if !reflect.DeepEqual(
		MergeRanges([]Range{NewRange(12, 0), NewRange(0, 0), NewRange(1, 0), NewRange(6, 5), NewRange(6, 4)}),
		[]Range{NewRange(0, 0)},
	) {
		t.Error("Failed to remove 0 or negative length ranges.")
	}
}
