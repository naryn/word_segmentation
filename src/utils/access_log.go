// access_log
package utils

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"
)

const (
	LOG_DATEFORMAT   = "2006-01-02"
	DEFAULT_LOG_SCAN = 10    //定时检查文件是否需要切割间隔时间
	DEFAULT_LOG_SEQ  = 10000 //队列长度
)

type AccessLogger struct {
	mu         *sync.RWMutex
	filePrefix string
	fileSuffix string
	date       *time.Time
	logFile    *os.File
	logChan    chan string
}

//访问日志初始化
func NewAccLogger(prefix, suffix string) *AccessLogger {
	lg := &AccessLogger{
		mu:         new(sync.RWMutex),
		filePrefix: prefix,
		fileSuffix: suffix,
		logChan:    make(chan string, DEFAULT_LOG_SEQ),
	}

	t, _ := time.Parse(LOG_DATEFORMAT, time.Now().Format(LOG_DATEFORMAT))

	lg.date = &t
	lg.mu.Lock()
	defer lg.mu.Unlock()

	if !lg.isMustSplit() {
		filePath := lg.filePath(lg.date.Format(LOG_DATEFORMAT))
		lg.logFile, _ = os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	} else {
		lg.split()
	}

	go lg.logWriter()
	go lg.fileMonitor()

	return lg
}

//生成访问日志
func (lg *AccessLogger) LogTimer(path, ext string, timer *Timer) {
	nowFormat := time.Now().Format("2006-01-02 15:04:05")
	str := fmt.Sprintf("[%v]\t\"%v\"", nowFormat, path)
	last := timer.Start
	execAll := time.Now().Sub(*last).Nanoseconds() / 1000000
	str += fmt.Sprintf("\t\"execAll:%v\"", execAll)
	for _, v := range timer.List {
		str += fmt.Sprintf("\t\"exec%v:%v\"", v.Msg, v.Timer.Sub(*last).Nanoseconds()/1000000)
		last = v.Timer
	}
	str += fmt.Sprintf("\t\"%v\"", url.QueryEscape(ext))
	lg.logChan <- str
}

func (lg *AccessLogger) Log(path, ext string) {
	nowFormat := time.Now().Format("2006-01-02 15:04:05")
	str := fmt.Sprintf("[%v]\t\"%v\"", nowFormat, path)
	str += fmt.Sprintf("\t\"%v\"", url.QueryEscape(ext))
	lg.logChan <- str
}

//异步写日志
func (lg *AccessLogger) logWriter() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("logWritter() catch panic: %v", err)
		}
	}()

	for {
		select {
		case str := <-lg.logChan:
			lg.write(str)
		}
	}
}

//将字符串写入到文件
func (lg *AccessLogger) write(str string) {
	lg.mu.RLock()
	defer lg.mu.RUnlock()
	buf := []byte(str)
	if len(str) == 0 || str[len(str)-1] != '\n' {
		buf = append(buf, '\n')
	}
	lg.logFile.Write(buf)
}

//判断是否需要切割
func (lg *AccessLogger) isMustSplit() bool {
	t, _ := time.Parse(LOG_DATEFORMAT, time.Now().Format(LOG_DATEFORMAT))
	if t.After(*lg.date) {
		return true
	}
	return false
}

//切割日志
func (lg *AccessLogger) split() {
	logFile := lg.filePath(time.Now().Format(LOG_DATEFORMAT))
	if !lg.isExist(logFile) {
		if lg.logFile != nil {
			lg.logFile.Close()
		}

		t, _ := time.Parse(LOG_DATEFORMAT, time.Now().Format(LOG_DATEFORMAT))
		lg.date = &t
		lg.logFile, _ = os.Create(logFile)
	}
}

func (lg *AccessLogger) filePath(date string) string {
	filePath := lg.filePrefix + date + lg.fileSuffix
	return filePath
}

//文件监控
func (lg *AccessLogger) fileMonitor() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("fileMonitor() catch panic: %v", err)
		}
	}()

	timer := time.NewTicker(time.Duration(DEFAULT_LOG_SCAN) * time.Second)
	for {
		select {
		case <-timer.C:
			lg.fileCheck()
		}
	}
}

func (lg *AccessLogger) fileCheck() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("fileCheck() catch panic: %v", err)
		}
	}()

	if lg.isMustSplit() {
		lg.mu.Lock()
		defer lg.mu.Unlock()
		lg.split()
	}
}

// Determine a file or a path exists in the os
func (lg *AccessLogger) isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
