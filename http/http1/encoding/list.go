package encoding

import (
	"strings"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps/pipe"
	"github.com/JOTPOT-UK/JOTPOT-Server/util"
)

//List contains a set of Encodings
type List struct {
	l []*Encoding
}

//NewList returns a new List, with the underlying slice capacity set to capacity.
func NewList(capacity int) List {
	return List{
		l: make([]*Encoding, 0, capacity),
	}
}

func (l *List) index(encoding string) (bool, int) {
	//If the list is empty, then insert at index 0!
	if len(l.l) == 0 {
		return false, 0
	}
	//Start with the full list
	a := 0
	b := len(l.l) - 1
	c := (a + b) / 2
	var r util.ComparisonResult
	for {
		r = util.CompareStrings(l.l[c].Name, encoding)
		if r == util.ComparisonEqual {
			//This is the index if the strings are equal
			return true, c
		} else if r == util.ComparisonGreaterThan {
			a = c + 1
			//If we have finished the search, then the item needs to go after the current one
			if b < a {
				return false, c + 1
			}
		} else {
			b = c - 1
			//If we have finished the search, then the item needs to go before the current one
			if b < a {
				return false, c
			}
		}
		c = (a + b) / 2
	}
}

//Contains returns true if an encoding with the name given is in the List
func (l *List) Contains(encoding string) bool {
	r, _ := l.index(encoding)
	return r
}

//Get returns the encoding within with the name given, or nil if the List doesn't have that encoding
func (l *List) Get(encoding string) *Encoding {
	r, i := l.index(encoding)
	if !r {
		return nil
	}
	return l.l[i]
}

//Remove removes the encoding with the given name from the list. If the encoding does not exist, it returns false.
func (l *List) Remove(encoding string) bool {
	r, i := l.index(encoding)
	if !r {
		return false
	}
	copy(l.l[i:], l.l[i+1:])
	l.l = l.l[:len(l.l)-1]
	return true
}

//Add adds the encoding to the List if an encoding with the same name does not already exist.
// If one already exists, it returns false, otherwise, on success, it returns true!
func (l *List) Add(encoding *Encoding) bool {
	//If the List is empty, then insert it
	if len(l.l) == 0 {
		l.l = append(l.l, encoding)
		return true
	}
	r, i := l.index(encoding.Name)
	if r {
		return false
	}
	if i == len(l.l) {
		//If we want to add it to the end, then just append it.
		l.l = append(l.l, encoding)
	} else {
		//Copy the last item to the place after it, to extend the list
		l.l = append(l.l, l.l[len(l.l)-1])
		//Make space for the new item
		copy(l.l[i+1:len(l.l)-1], l.l[i:len(l.l)-2])
		//Add it
		l.l[i] = encoding
	}
	return true
}

//GetReaderPipeGenerators returns the list of pipe.ReaderPipeGenerators
// that should be used to decode a message  that was encoded with the given encoding names.
// codes should be in the order in which the encodings were applied to the data - as they would be in a HTTP header.
func (l *List) GetReaderPipeGenerators(codes []string) ([]*pipe.ReaderPipeGenerator, bool) {
	out := make([]*pipe.ReaderPipeGenerator, len(codes), len(codes))
	var thisOk bool
	var code int
	var count int
	//Iterate backwards as we are undoing the encodings
	for i := len(codes) - 1; i > -1; i-- {
		//Get the current encoding, if it isn't in the list, then return not ok.
		thisOk, code = l.index(strings.ToLower(codes[i]))
		if !thisOk {
			return nil, false
		}
		//Add it to the output, and inc count
		out[count] = &l.l[code].Reader
		count++
	}
	return out, true
}

func (l *List) GetWriterPipeGenerators(codes []string) ([]*pipe.WriterPipeGenerator, bool) {
	out := make([]*pipe.WriterPipeGenerator, len(codes), len(codes))
	ok := true
	var thisOk bool
	var code int
	for i := range codes {
		thisOk, code = l.index(codes[i])
		if thisOk {
			out[i] = &l.l[code].Writer
		} else {
			ok = false
		}
	}
	return out, ok
}
