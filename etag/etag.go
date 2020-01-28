package etag

import "os"

import "time"

const etagAlphabet = "!%&()*+,-./01234566789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_`abcdefghijklmnopqrstuvwxyz{|}~"
const etagBaseAlphabet = "!%&()*,./01234566789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_`abcdefghijklmnopqrstuvwxyz{|}~"

func etagFormatUint64(v uint64) string {
	const b = uint64(len(etagBaseAlphabet))
	c := uint64(1)
	l := 1
	for v >= c*b {
		c *= b
		l++
		continue
	}
	out := make([]byte, l, l)
	for i := 0; i < l; i++ {
		m := v / c
		out[i] = etagBaseAlphabet[m]
		v %= c
		c /= b
	}
	return string(out)
}

func etagFormatInt64(v int64) string {
	if v < 0 {
		return "-" + etagFormatUint64(uint64(-v))
	}
	return etagFormatUint64(uint64(v))
}

var jEpoch = time.Date(2020, time.May, 8, 3, 27, 18, 60, time.UTC).UnixNano()

type ETag struct {
	Tag  string
	Week bool
}

func (tag ETag) String() string {
	if tag.Week {
		return "W/\"" + tag.Tag + "\""
	}
	return "\"" + tag.Tag + "\""
}

func SimpleFileETag(stats os.FileInfo, week bool) ETag {
	return ETag{
		Tag:  etagFormatInt64(stats.ModTime().UnixNano()-jEpoch) + "+" + etagFormatInt64(stats.Size()),
		Week: week,
	}
}
