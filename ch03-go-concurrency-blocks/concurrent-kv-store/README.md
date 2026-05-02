# Concurrent Key-Value Store
A simple implementation of a KV Store interface
```
type KVStore interface {
    Get(key string) (string, bool)
    Set(key, value string)
}
```

That is concurrent safe and has no race conditions. Uses sync.RWMutex for concurrent read access and exclusive writes.