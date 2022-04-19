package cache

import (
	"sync"
	"testing"
	"time"
)

var rw sync.RWMutex

func BenchmarkCacheStorage_Set(b *testing.B) {
	key := "key1"
	Client.SetMaxMemory("1GB")
	for i := 0; i < b.N; i++ {
		i = i
		go func(i int) {
			rw.Lock()
			Client.Set(key, i, time.Second)
			val, _ := Client.Get(key)
			rw.Unlock()
			if val != i {
				b.Log("get key false")
			}
		}(i)
	}
}
