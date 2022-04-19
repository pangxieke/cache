package cache

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCacheStorage_Get(t *testing.T) {
	a := assert.New(t)
	key := "key1"
	val := "123456789"
	Client.Set(key, val, time.Second*2)

	act, ok := Client.Get(key)
	a.Equal(true, ok)
	a.Equal(val, act)

	time.Sleep(time.Second * 2)
	act, ok = Client.Get(key)
	a.Equal(false, ok)
	a.Equal(nil, act)
}

func TestCacheStorage_Set(t *testing.T) {
	a := assert.New(t)
	key := "key1"
	val := "123456789"
	Client.Set(key, val, time.Second*2)
	a.EqualValues(1, Client.keys)
}

func TestCacheStorage_Exists(t *testing.T) {
	a := assert.New(t)
	key := "key1"
	val := "123456789"
	Client.Set(key, val, time.Second*2)

	isExist := Client.Exists(key)
	a.Equal(true, isExist)

	isExist = Client.Exists(key)
	a.Equal(true, isExist)

	isExist2 := Client.Exists("key2")
	a.Equal(false, isExist2)
}

func TestCacheStorage_Flush(t *testing.T) {
	a := assert.New(t)
	key := "key1"
	val := "123456789"
	Client.Set(key, val, time.Second*2)

	Client.Flush()
	isExist := Client.Exists(key)
	a.Equal(false, isExist)
}

func TestCacheStorage_SetMaxMemory(t *testing.T) {
	a := assert.New(t)
	a.EqualValues(DefMaxMemory, Client.MaxMemory)

	Client.SetMaxMemory("1MB")
	a.EqualValues(MB, Client.MaxMemory)
}

func TestCacheStorage_Keys(t *testing.T) {
	a := assert.New(t)
	Client.Flush()
	Client.Set("key1", 111, time.Second)

	a.EqualValues(1, Client.keys)
	Client.Set("key2", 111, time.Second)
	a.EqualValues(2, Client.keys)
}

func TestCacheStorage_Del(t *testing.T) {
	a := assert.New(t)

	key := "key1"
	val := "123456789"
	Client.Set(key, val, time.Second*2)

	ok := Client.Del(key)
	a.Equal(true, ok)

	isExist := Client.Exists(key)
	a.Equal(false, isExist)
}

func TestSplitMemory(t *testing.T) {
	a := assert.New(t)
	str := "12KB"

	var m int
	if n, err := splitMemory(str, "KB"); err == nil {
		m = n * KB
	}
	if n, err := splitMemory(str, "MB"); err == nil {
		m = n * MB
	}

	a.EqualValues(12*KB, m)
}

func TestDelTimer(t *testing.T) {
	var sig chan bool
	sig = make(chan bool)

	go DelTimer(sig)
	Client.SetMaxMemory("100MB")

	for i := 0; i < 100; i++ {
		Client.Set(fmt.Sprintf("%d", i), i, time.Second)
	}

	time.Sleep(time.Second * 6)
	t.Log(Client.keys)
	sig <- true

}

func TestSyncDel(t *testing.T) {
	for i := 0; i < 10; i++ {
		Client.Set(fmt.Sprintf("%d", i), i, time.Second)
	}

	time.Sleep(time.Second)
	SyncDel()

	t.Log(Client.keys)
}

// 自动清理策略
func TestCacheStorage_Clear(t *testing.T) {
	Client.AutoClear()

	for i := 0; i < 10; i++ {
		Client.Set(fmt.Sprintf("%d", i), i, time.Second)
	}

	t.Log(Client.keys)
	time.Sleep(time.Second * 2)
	t.Log(Client.keys)

	time.Sleep(time.Second * 2)
	t.Log(Client.keys)

	time.Sleep(time.Second * 2)
	t.Log(Client.keys)

	time.Sleep(time.Second * 2)
	t.Log(Client.keys)
}
