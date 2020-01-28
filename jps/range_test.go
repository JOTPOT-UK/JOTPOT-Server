package jps

import "testing"

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

}
