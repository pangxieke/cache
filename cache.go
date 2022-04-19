package cache

import (
	"github.com/astaxie/beego/logs"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

/**
⼀个简易的内存缓存系统
⽀持设定过期时间，精度为秒级
⽀持设定最⼤内存，当内存超出时候做出合理的处理
⽀持并发安全。
为简化编程细节，⽆需实现数据落地。
*/

/**
1. 过期数据清理 可以借鉴 redis数据清理模型
2. 数据存储类型 map
3. 内存处理，内存超出时，拒绝服务
4. 并发安全 redis单线程
*/

/**
⽀持过期时间和最⼤内存⼤⼩的的内存缓存库。 按照要求实现这⼀个接⼝。
*/
type Cache interface {
	//size 是⼀个字符串。⽀持以下参数: 1KB，100KB，1MB，2MB，1GB 等
	SetMaxMemory(size string) bool
	// 设置⼀个缓存项，并且在expire时间之后过期
	Set(key string, val interface{}, expire time.Duration)
	// 获取⼀个值
	Get(key string) (interface{}, bool)
	// 删除⼀个值
	Del(key string) bool
	// 检测⼀个值 是否存在
	Exists(key string) bool
	// 清空所有值
	Flush() bool
	// 返回所有的key 多少
	Keys() int64
}

type Node struct {
	Key    string
	Data   interface{}
	Expire time.Time
	Memory int64
}

type CacheStorage struct {
	sync.RWMutex
	MaxMemory int64
	data      map[string]*Node
	Memory    int64 // 使用了多少memory
	keys      int64 //多少key
}

var Client *CacheStorage
var one sync.Once

func init() {
	Client = NewCacheStorage()
}

const (
	_  = iota
	KB = 1 << (10 * iota)
	MB = 1 << (10 * iota)
	GB = 1 << (10 * iota)
	TB = 1 << (10 * iota)
	PB = 1 << (10 * iota)
)
const DefMaxMemory = KB

// 单例
func NewCacheStorage() *CacheStorage {
	if Client == nil {
		one.Do(func() {
			c := &CacheStorage{
				data:      make(map[string]*Node),
				MaxMemory: DefMaxMemory,
			}
			c.Flush()
			Client = c
		})
	}
	return Client
}

// 字符串拆分
func splitMemory(str string, sep string) (n int, err error) {
	s := strings.Split(str, sep)
	n, err = strconv.Atoi(s[0])
	return
}

// 设置最大内存
func (this *CacheStorage) SetMaxMemory(size string) bool {
	var m int
	if n, err := splitMemory(size, "KB"); err == nil {
		m = n * KB
	}
	if n, err := splitMemory(size, "MB"); err == nil {
		m = n * MB
	}
	if n, err := splitMemory(size, "GB"); err == nil {
		m = n * GB
	}
	if n, err := splitMemory(size, "PB"); err == nil {
		m = n * PB
	}
	if m != 0 {
		this.MaxMemory = int64(m)
		return true
	} else {
		return false
	}
}

func (this *CacheStorage) Set(key string, val interface{}, expire time.Duration) {
	node := Node{
		Key:    key,
		Data:   val,
		Expire: time.Now().Add(expire),
		Memory: int64(unsafe.Sizeof(val)),
	}
	m := unsafe.Sizeof(node)

	// 内存是否超出
	if this.Memory+int64(m) > this.MaxMemory {
		logs.Info("memory over", this.MaxMemory)
		panic("memory over")
		return
	}

	this.Lock()
	defer this.Unlock()

	this.data[key] = &node
	this.keys++
	this.Memory += int64(m)
}

func (this *CacheStorage) Get(key string) (interface{}, bool) {
	if this.Exists(key) {
		this.Lock()
		defer this.Unlock()
		if val, ok := this.data[key]; ok {
			return val.Data, true
		}
	}
	return nil, false
}
func (this *CacheStorage) Del(key string) bool {
	this.Lock()
	defer this.Unlock()

	node, ok := Client.data[key]
	if !ok {
		return false
	}
	delete(this.data, key)
	this.keys--
	this.Memory -= node.Memory

	return true
}

func (this *CacheStorage) Exists(key string) bool {
	this.Lock()
	var isExpire bool
	if node, ok := this.data[key]; ok {
		if node.Expire.After(time.Now()) {
			this.Unlock()
			return true
		} else {
			isExpire = true
		}
	}
	this.Unlock()
	if isExpire {
		this.Del(key)
	}
	return false
}

func (this *CacheStorage) Flush() bool {
	this.Lock()
	defer this.Unlock()
	this.data = make(map[string]*Node)
	this.keys = 0
	this.Memory = 0
	return false
}

func (this *CacheStorage) Keys() int64 {
	this.Lock()
	defer this.Unlock()
	return this.keys
}

/**
 * 自动清理过期数据
 * 策略：每2秒取 1/10数据，判断过期，则清理
 */
func (this *CacheStorage) AutoClear() {
	sig := make(chan bool)
	go DelTimer(sig)
}

// 定期清理定时器
func DelTimer(sig chan bool) {
	timer := time.NewTicker(time.Second * 2)
	defer timer.Stop()

Loop:
	for {
		select {
		case <-timer.C:
			logs.Info("clear expired key")
			go SyncDel()
		case <-sig:
			break Loop
		}
	}
}

// 随机取1/10数据，判断是否删除
func SyncDel() {
	n := int(Client.keys / 10)

	var count int
	for key, _ := range Client.data {
		if count > n {
			break
		}
		if !Client.Exists(key) {
			Client.Del(key)
		}
		count++
	}
}
