## 要求
⼀个简易的内存缓存系统
⽀持设定过期时间，精度为秒级
⽀持设定最⼤内存，当内存超出时候做出合理的处理
⽀持并发安全。
为简化编程细节，⽆需实现数据落地。

```
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
```

## 测试
```
go test -v -count=1 ./
go test -bench=.
```

## 使用
```
// 设置最大内存 超出panic
Client.SetMaxMemory("1MB")

//设置⼀个缓存项
key := "key1"
val := "123456789"
Client.Set(key, val, time.Second*2)

// 获取⼀个值
act, ok = Client.Get(key)

//删除⼀个值
ok := Client.Del(key)

// 检测⼀个值
isExist := Client.Exists(key)

// 清空所有值	
Client.Flush()	

// 返回所有的key
Client.keys
```
