package header

import (
	"errors"
	"sync"
)

//LockedSlice has a slice of strings, and a mutex lock!
type LockedSlice struct {
	Slice []string
	Lock  sync.RWMutex
}

//Header is used to store header values for header keys.
// It has locks to aid mutlithreaded usage
type Header struct {
	//Values contains map which maps raw keys to raw values.
	Values map[string]*LockedSlice
	//Processor defines the Processor that should be used when processing and transforming keys and values.
	Processor  *Processor
	KeyModLock sync.RWMutex
}

type UnsafeAdder struct {
	h *Header
}

func (h *UnsafeAdder) Add(key, value string) {
	h.h.AddUnsafe(key, value)
}

var ErrNoHeaderToRelease = errors.New("no header to released")

func (h *Header) UnsafeAdder() UnsafeAdder {
	h.KeyModLock.Lock()
	h.Processor.Lock.RLock()
	return UnsafeAdder{h}
}

func (h *UnsafeAdder) Release() error {
	if h.h == nil {
		return ErrNoHeaderToRelease
	}
	h.h.Processor.Lock.RUnlock()
	h.h.KeyModLock.Unlock()
	h.h = nil
	return nil
}

func New(capacity int, processor *Processor) *Header {
	return &Header{
		Values:    make(map[string]*LockedSlice, capacity),
		Processor: processor,
	}
}

//Processor defines how keys and values should be processed when being added to a header
type Processor struct {
	//KeyTransforms is a slice of functions, which are used to transform header keys. When a key is used, it will be passed through these functions in order.
	KeyTransforms []func(string) string
	//ValueProcessors is a slice of functions, each of which will be called in sequence when values are modified.
	// The first argument is the key for which the values are being modified.
	// The second argument is a slice containing all the values for the key of which the values have been modified for.
	// The third and fourth arguments are the range of values which have been modified.
	// The values returned are the arguments passed to the call of the next function in the slice.
	//Note that returning a different header key does not change the header key under which the values are stored, it just changes the value passed to the next function.
	ValuesProcessors []func(string, []string, int, int) (string, []string, int, int)
	//Lock is a RWMutex associated with the Processor
	Lock sync.RWMutex
}

//GetKey applies the KeyTransforms to the given key, and returns the resulting raw key.
// It is important to note that this function DOES NOT lock the Processor. The programmer should either ensure that the Lock is locked throughout this call, or use the GetKeySafe method.
func (p *Processor) GetKey(key string) string {
	for i := range p.KeyTransforms {
		key = p.KeyTransforms[i](key)
	}
	return key
}

//GetKeySafe calls p.Lock.RLock, then p.GetKey(key), then p.Lock.RUnlock. It returns the value returned by the GetKey call.
func (p *Processor) GetKeySafe(key string) string {
	p.Lock.RLock()
	defer p.Lock.RUnlock()
	return p.GetKey(key)
}

//ProcessValues takes the arguments to pass the the first function in p.ValuesProcessors, and then chains them!
// It returns the final value of the values slice.
// It is important to note that this function DOES NOT lock the Processor. The programmer should either ensure that the Lock is locked throughout this call, or use the ProcessValuesSafe method.
func (p *Processor) ProcessValues(key string, values []string, a, b int) []string {
	for i := range p.ValuesProcessors {
		key, values, a, b = p.ValuesProcessors[i](key, values, a, b)
	}
	return values
}

//ProcessValuesSafe calls p.Lock.RLock, then p.ProcessValues(key, values, a, b), then p.Lock.RUnlock. It returns the value returned by the ProcessValues call.
func (p *Processor) ProcessValuesSafe(key string, values []string, a, b int) []string {
	p.Lock.RLock()
	defer p.Lock.RUnlock()
	return p.ProcessValues(key, values, a, b)
}

//GetValuesRawKey returns a copy of the values assosiated with the given key (as it is the raw key, it is not passed through h.Processor.GetKey).
func (h *Header) GetValuesRawKey(key string) []string {
	vs, ok := h.Values[key]
	if !ok {
		return nil
	}
	vs.Lock.RLock()
	defer vs.Lock.RUnlock()
	if len(vs.Slice) == 0 {
		//Return nil rather than vs to make sure we don't return a reference to vs
		return nil
	}
	//Return a copy of the values
	rv := make([]string, len(vs.Slice), len(vs.Slice))
	copy(rv, vs.Slice)
	return rv
}

func (h *Header) GetValuesRawKeyUnsafe(key string) []string {
	vs, ok := h.Values[key]
	if !ok || len(vs.Slice) == 0 {
		return nil
	}
	rv := make([]string, len(vs.Slice), len(vs.Slice))
	copy(rv, vs.Slice)
	return rv
}

//GetValues returns the values stored under the given key
func (h *Header) GetValues(key string) []string {
	return h.GetValuesRawKey(h.Processor.GetKeySafe(key))
}

//Get returns the first element returned by h.GetValues(key), or "" if there isn't one.
func (h *Header) Get(key string) string {
	vs := h.GetValues(key)
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

//SetValuesRaw sets the values under key to values.
func (h *Header) SetValuesRaw(key string, values []string) {
	h.KeyModLock.RLock()
	vs, ok := h.Values[key]
	if ok {
		//If it exists, then defer RUnlock and carry on
		defer h.KeyModLock.RUnlock()
	} else {
		//OK, so now we need to hold the write lock because we want ot add a key
		h.KeyModLock.RUnlock()
		h.KeyModLock.Lock()
		defer h.KeyModLock.Unlock()
		vs, ok = h.Values[key]
		//Hopfully, we still don't have the key.. In which case, add it!
		if !ok {
			h.Values[key] = &LockedSlice{Slice: values}
			return
		}
		//Oh poo! It exists now... We could just grab a read lock, but for the sake of getting this done, let's just defer the write unlock and use the existing slice.
	}
	//If the key exists, then replace the slice within!
	vs.Lock.Lock()
	vs.Slice = values
	vs.Lock.Unlock()
}

//SetValues sets the values under the given key to the values given.
func (h *Header) SetValues(key string, values []string) {
	h.Processor.Lock.RLock()
	key = h.Processor.GetKey(key)
	values = h.Processor.ProcessValues(key, values, 0, len(values))
	h.Processor.Lock.RUnlock()
	h.SetValuesRaw(key, values)
}

//Set sets the value to the only value under key (an alias for h.SetValues(key, []string{value}))
func (h *Header) Set(key, value string) {
	h.SetValues(key, []string{value})
}

//AddRawKey appends value to the set of values under raw key
func (h *Header) AddRawKey(key, value string) {
	h.KeyModLock.RLock()
	vs, ok := h.Values[key]
	if ok {
		//If it exists, then defer RUnlock and carry on
		defer h.KeyModLock.RUnlock()
	} else {
		//OK, so now we need to hold the write lock because we want ot add a key
		h.KeyModLock.RUnlock()
		h.KeyModLock.Lock()
		defer h.KeyModLock.Unlock()
		vs, ok = h.Values[key]
		//Hopfully, we still don't have the key.. In which case, add it!
		if !ok {
			h.Values[key] = &LockedSlice{
				Slice: h.Processor.ProcessValues(
					key, []string{value}, 0, 1,
				),
			}
			return
		}
		//Oh poo! It exists now... We could just grab a read lock, but for the sake of getting this done, let's just defer the write unlock and use the existing slice.
	}
	//If the key exists, then replace the slice within!
	vs.Lock.Lock()
	vs.Slice = h.Processor.ProcessValues(
		key, append(vs.Slice, value), len(vs.Slice)-1, len(vs.Slice),
	)
	vs.Lock.Unlock()
}

func (h *Header) AddRawKeyUnsafe(key, value string) {
	vs, ok := h.Values[key]
	if !ok {
		h.Values[key] = &LockedSlice{
			Slice: h.Processor.ProcessValues(
				key, []string{value}, 0, 1,
			),
		}
	} else {
		vs.Slice = h.Processor.ProcessValues(
			key, append(vs.Slice, value), len(vs.Slice)-1, len(vs.Slice),
		)
	}
}

//Add appends the value to the set of values under the given key
func (h *Header) Add(key, value string) {
	h.Processor.Lock.RLock()
	defer h.Processor.Lock.RUnlock()
	h.AddRawKey(h.Processor.GetKey(key), value)
}

func (h *Header) AddUnsafe(key, value string) {
	h.AddRawKeyUnsafe(h.Processor.GetKey(key), value)
}

func (h *Header) DelRawKey(key string) bool {
	//We must make sure that 2 threads don't try and delete the same key at the same time,
	// becuase if they did, they could both get ok==true, but then only one of them could take the lock.
	// So the other would be left in deadlock.
	//We could lock on a per key basis, however that would require a map, and these can't shrink again, making them potentially inefficient.
	//In addition, that wouldn't speed things up much, so for now, it will be implemented with a lock on all deletes.

	//Grab the read lock, as we want to check it hasn't been deleted already.
	h.KeyModLock.RLock()
	_, ok := h.Values[key]
	if ok {
		//Now we need to get the write lock
		h.KeyModLock.RUnlock()
		h.KeyModLock.Lock()
		defer h.KeyModLock.Unlock()
		_, ok = h.Values[key]
		//If we've already been deleted, then unlock and return
		if !ok {
			return false
		}
		//TODO: Do we need?: vs.Lock.Lock()
		delete(h.Values, key)
		return true
	}
	h.KeyModLock.RUnlock()
	return false
}

func (h *Header) Del(key string) {
	h.DelRawKey(h.Processor.GetKeySafe(key))
}

func (h *Header) Has(key string, values []string) bool {
	h.Processor.Lock.RLock()
	key = h.Processor.GetKey(key)
	values = h.Processor.ProcessValues(key, values, 0, len(values))
	if len(values) == 0 {
		return true
	}
	h.Processor.Lock.RUnlock()
	have := h.GetValuesRawKey(key)
	left := len(values)
	wants := make([]bool, left, left)
	for i := range have {
		for j := range values {
			if wants[j] == false && have[i] == values[j] {
				left--
				if left == 0 {
					return true
				}
				wants[j] = true
				break
			}
		}
	}
	return false
}

func (h *Header) ModifyRawKey(key string, cb func(key string, values []string) ([]string, error)) error {
	h.KeyModLock.RLock()
	vs, ok := h.Values[key]
	if ok {
		defer h.KeyModLock.RUnlock()
	} else {
		h.KeyModLock.RUnlock()
		h.KeyModLock.Lock()
		defer h.KeyModLock.Unlock()
		vs, ok = h.Values[key]
		if !ok {
			values, err := cb(key, nil)
			if err == nil {
				h.Values[key] = &LockedSlice{
					Slice: values,
				}
			}
			return err
		}
	}
	vs.Lock.Lock()
	values, err := cb(key, vs.Slice)
	if err == nil {
		vs.Slice = values
	}
	vs.Lock.Unlock()
	return err
}

func (h *Header) ForEach(cb func(key string, values []string) error, lock bool) (err error) {
	h.KeyModLock.RLock()
	defer h.KeyModLock.RUnlock()
	for key, vs := range h.Values {
		vs.Lock.RLock()
		if len(vs.Slice) != 0 {
			values := make([]string, len(vs.Slice), len(vs.Slice))
			copy(values, vs.Slice)
			if lock {
				err = cb(key, values)
				vs.Lock.RUnlock()
			} else {
				vs.Lock.RUnlock()
				err = cb(key, values)
			}
		} else {
			vs.Lock.RUnlock()
		}
	}
	return
}
