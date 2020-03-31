package mylogger

import (
	"fmt"
	"os"
	"path"
	"time"
)

// 往文件里面写日志相关代码

var (
	// MaxSize 日志通道缓冲区大小
	MaxSize = 5000
)

// FileLogger 结构体
type FileLogger struct {
	Level       LogLevel
	filePath    string // 日志文件保存路径
	fileName    string // 日志文件保存的文件名
	fileObj     *os.File
	errFileObj  *os.File
	maxFileSize int64
	logChan     chan *logMsg
}

type logMsg struct {
	level     LogLevel
	msg       string
	funcName  string
	fileName  string
	timestamp string
	line      int
}

// NewFileLogger 构造函数
func NewFileLogger(levelStr, fp, fn string, maxSize int64) *FileLogger {
	logLevel, err := parseLogLevel(levelStr)
	if err != nil {
		panic(err)
	}
	f1 := &FileLogger{
		Level:       logLevel,
		filePath:    fp,
		fileName:    fn,
		maxFileSize: maxSize,
		logChan:     make(chan *logMsg, MaxSize),
	}
	err = f1.initFile() // 按照文件路径和文件名将文件打开
	if err != nil {
		panic(err)
	}
	return f1
}

func (f *FileLogger) initFile() error {
	// 初始化文件
	logPath := path.Join(f.filePath, f.fileName)
	fileObj, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("open log file failed, err:%v\n", err)
		return err
	}
	errFileObj, err := os.OpenFile(logPath+".err", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("open log file failed, err:%v\n", err)
		return err
	}
	// 此时 日志文件打开完毕
	f.fileObj = fileObj
	f.errFileObj = errFileObj
	// 开启后台goroutine去异步写日志
	go f.writeLog()
	return nil
}

// 检查日志文件大小（是否需要切割）
func (f *FileLogger) checkSize(file *os.File) bool {
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("get file info failed, err:%v\n", err)
		return false
	}
	// 如果当前文件大小 >= 日志文件的最大值，就应该返回true
	return fileInfo.Size() >= f.maxFileSize
}

// 切割文件
func (f *FileLogger) splitFile(file *os.File) (*os.File, error) {
	// 需要切割日志文件
	nowStr := time.Now().Format("20060102150405000")
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("get file info failed,err:%v\n", err)
		return nil, err
	}

	logPath := path.Join(f.filePath, fileInfo.Name())      // 拿到当前的日志文件完整路径
	newlogPath := fmt.Sprintf("%s.bak%s", logPath, nowStr) // 拼接一个日志文件备份的名字

	// 1. 关闭当前日志文件
	file.Close()
	// 2. 备份一下 rename  xx.log -> xx.log.bak202003091952
	os.Rename(logPath, newlogPath)
	// 3. 打开一个新的日志文件
	fileObj, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("open new log file failed, err:%v\n", err)
		return nil, err
	}
	// 4. 将打开的新日志文件对象赋值给 f.fileObj
	return fileObj, nil
}

// 后台异步写日志方法
func (f *FileLogger) writeLog() {
	for {
		// 判断是否需要切割日志文件
		if f.checkSize(f.fileObj) {
			newFile, err := f.splitFile(f.fileObj)  // 得到日志文件对象
			if err != nil {
				return
			}
			f.fileObj = newFile
		}
		if f.checkSize(f.errFileObj) {
			newFile, err := f.splitFile(f.errFileObj)
			if err != nil {
				return
			}
			f.errFileObj = newFile
		}
		select {
		case logTmp := <-f.logChan:
			// 拼接日志信息：
			logInfo := fmt.Sprintf("[%s] [%s] [%s:%s:%d] %s\n", logTmp.timestamp, unparseLogLevle(logTmp.level), logTmp.fileName, logTmp.funcName, logTmp.line, logTmp.msg)
	
			// 写入日志
			fmt.Fprintf(f.fileObj, logInfo)
			if logTmp.level >= ERROR {
				// 如果记录的日志级别大于ERROR，还要在err日志文件中再记录一次
				fmt.Fprintf(f.errFileObj, logInfo)
			}
		default:
			// 取不到日志先休息500ms
			time.Sleep(time.Millisecond * 500)
		}
		
	}
}

// 记录日志的方法
func (f *FileLogger) log(lv LogLevel, format string, a ...interface{}) {

	if lv >= f.Level {
		msg := fmt.Sprintf(format, a...)
		now := time.Now()
		funcName, fileName, lineNo := getInfo(3)
		// 先把日志发送到通道中
		logTmp := &logMsg{
			level:     lv,
			msg:       msg,
			funcName:  funcName,
			fileName:  fileName,
			timestamp: now.Format("2006-01-02 15:04:05"),
			line:      lineNo,
		}
		select {
		case f.logChan <- logTmp:
		default:
			// 把日志丢弃保证不出现阻塞
		}
	}
}

// Debug 级别
func (f *FileLogger) Debug(format string, a ...interface{}) {
	f.log(DEBUG, format, a...)
}

// Info 级别
func (f *FileLogger) Info(format string, a ...interface{}) {
	f.log(INFO, format, a...)
}

// Warning 级别
func (f *FileLogger) Warning(format string, a ...interface{}) {
	f.log(WARNING, format, a...)
}

// Error 级别
func (f *FileLogger) Error(format string, a ...interface{}) {
	f.log(ERROR, format, a...)
}

// Fatal 级别
func (f *FileLogger) Fatal(format string, a ...interface{}) {
	f.log(FATAL, format, a...)
}

// Close 关闭资源
func (f *FileLogger) Close() {
	f.fileObj.Close()
	f.errFileObj.Close()
}
