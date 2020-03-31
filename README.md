# go_logger
一个日志库软件，可以将五种（debug, info, warning, error, fatal)级别的日志保存至文件中。
## 快速使用：
```go
package main

import (
	mylogger "github.com/lincoln77/go_logger"
)

// 测试我们自己写 的日志库

var log mylogger.Logger

func main() {
	log = mylogger.NewFileLogger("info", "./", "lincoln.log", 10*2024*1024)
	defer log.Close()

	log.Debug("这是一条Debug日志")
	log.Info("这是一条info日志")
	log.Warning("这是一条warning日志")
	id := 10010
	name := "lyh"
	log.Error("这是一条error日志, id:%d, name:%s", id, name)
	log.Fatal("这是一条fatal日志")
	// time.Sleep(time.Second)
}
```
