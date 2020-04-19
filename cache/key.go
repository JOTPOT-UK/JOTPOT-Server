package cache

type Key struct {
	Host     string
	Path     string
	RawQuery string

	Extra []string
}

func slicesOfStringsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type CachedResponse []byte

type ExtraMap struct {
	m []struct {
		Extra []string
		Value CachedResponse
	}
}

func (em ExtraMap) Get(k []string) *CachedResponse {
	for _, v := range em.m {
		if slicesOfStringsEqual(v.Extra, k) {
			return &v.Value
		}
	}
	return nil
}

type CacheTable struct {
	m map[string]map[string]ExtraMap
}

func (ct CacheTable) Get(k Key) *CachedResponse {
	if m, ok := ct.m[k.Host]; ok {
		if m, ok := m[k.Path+"?"+k.RawQuery]; ok {
			return m.Get(k.Extra)
		}
	}
	return nil
}
