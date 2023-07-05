package maps

import "sync"

func GetOrCreateLocked[K comparable, V any](m map[K]V, key K, mx *sync.RWMutex, createFn func(key K) (V, error)) (V, bool, error) {
	mx.RLock()
	v, ok := m[key]
	mx.RUnlock()
	if ok {
		return v, false, nil
	}
	mx.Lock()
	defer mx.Unlock()

	v, ok = m[key]

	if ok {
		return v, false, nil
	}

	v, err := createFn(key)
	if err != nil {
		return v, false, err
	}

	m[key] = v
	return v, true, nil
}

func ExistsLocked[K comparable, V any](m map[K]V, key K, mx *sync.RWMutex) bool {
	mx.RLock()
	_, ok := m[key]
	mx.RUnlock()

	return ok
}
