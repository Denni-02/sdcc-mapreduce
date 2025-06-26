package utils

import "sync"

// threadSafeMap rappresenta una mappa protetta da mutex
type ThreadSafeMap struct {
	Data map[string]bool
	Mu   *sync.Mutex
}

// Get restituisce il valore associato a una chiave in modo thread-safe
func (t *ThreadSafeMap) Get(key string) bool {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	return t.Data[key]
}

// Set imposta un valore per una chiave in modo thread-safe
func (t *ThreadSafeMap) Set(key string, value bool) {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	t.Data[key] = value
}

// Copy restituisce una copia della mappa
func (t *ThreadSafeMap) Copy() map[string]bool {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	copy := make(map[string]bool)
	for k, v := range t.Data {
		copy[k] = v
	}
	return copy
}
