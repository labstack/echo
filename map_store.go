package echo

// MapStore is a string -> arbitrary value store with Get/Set methods.
type MapStore map[string]interface{}

// Set sets a key value into the store.
func (m MapStore) Set(key string, val interface{}) {
	m[key] = val
}

// Get retrieves a key from the store (may be nil)
func (m MapStore) Get(key string) interface{} {
	return m[key]
}

// Del removes a key from the store
func (m MapStore) Del(key string) {
	delete(m, key)
}
