// logger
package utils

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	LOGGER_DATEFORMAT = "2006-01-02"
	LOGGER_LOG_SCAN   = 10    //定时检查文件是否需要切割间隔时间
	LOGGER_LOG_SEQ    = 10000 //队列长度
)

type Logger struct {
	mu         *sync.RWMutex
	filePrefix string
	fileSuffix string
	logSrv     string
	date       *time.Time
	logFile    *os.File
	logChan    chan string
}

//日志初始化
func NewLogger(logSrv, prefix, suffix string) *Logger {
	lg := &Logger{
		mu:         new(sync.RWMutex),
		filePrefix: prefix,
		fileSuffix: suffix,
		logSrv:     logSrv,
		logChan:    make(chan string, LOGGER_LOG_SEQ),
	}

	t, _ := time.Parse(LOGGER_DATEFORMAT, time.Now().Format(LOGGER_DATEFORMAT))

	lg.date = &t
	lg.mu.Lock()
	defer lg.mu.Unlock()

	if !lg.isMustSplit() {
		filePath := lg.filePath(lg.date.Format(LOGGER_DATEFORMAT))
		lg.logFile, _ = os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	} else {
		lg.split()
	}

	go lg.logWriter()
	go lg.fileMonitor()

	return lg
}

func (lg *Logger) Write(service, level, msg interface{}) {
	nowFormat := time.Now().Format("2006-01-02 15:04:05")
	str := fmt.Sprintf("[%s]\t%s\t%s\t%v", nowFormat, service, level, msg)
	lg.logChan <- str
}

func (lg *Logger) Debug(service, msg interface{}) {
	lg.Write(service, "debug", msg)
}

func (lg *Logger) Info(service, msg interface{}) {
	lg.Write(service, "info", msg)
}

func (lg *Logger) Warning(service, msg interface{}) {
	lg.Write(service, "warning", msg)
}

func (lg *Logger) Error(service, msg interface{}) {
	lg.Write(service, "error", msg)
}

func (lg *Logger) Critical(service, msg interface{}) {
	lg.Write(service, "critical", msg)
}

//异步写日志
func (lg *Logger) logWriter() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Logger logWritter() catch panic: %v", err)
		}
	}()

	for {
		select {
		case str := <-lg.logChan:
			lg.write(str)
		}
	}
}

//
func (lg *Logger) write(str string) {
	defer func() {
		if err := recover(); err != nil {
			lg.localWrite(str)
		}
	}()
	lg.remoteWrite(str)
}

func (lg *Logger) remoteWrite(str string) {
	params := strings.Split(str, "\t")
	if len(params) < 4 {
		panic("Logger params error")
	}
	if params[2] == "debug" {
		panic("debug")
	}
	rand.Seed(int64(time.Now().Nanosecond()))
	randInt := rand.Intn(10)
	if params[2] == "info" && randInt > 0 { //量太大，采样采集到后台
		panic("info")
	}
	uri := fmt.Sprintf(
		"/log?app=goapi&type=file&entry=&service=%s&level=%s",
		url.QueryEscape(params[1]),
		url.QueryEscape(params[2]))
	url := lg.logSrv + uri
	req := Post(url)
	req.SetTimeout(1*time.Second, 10*time.Millisecond)
	req.Param("message", params[3])
	resp, err := req.DoRequest()
	if err != nil {
		return
	}
	if resp.Body == nil {
		return
	}
	resp.Body.Close()
}

//将字符串写入到文件
func (lg *Logger) localWrite(str string) {
	lg.mu.RLock()
	defer lg.mu.RUnlock()
	buf := []byte(str)
	if len(str) == 0 || str[len(str)-1] != '\n' {
		buf = append(buf, '\n')
	}
	lg.logFile.Write(buf)
}

//判断是否需要切割
func (lg *Logger) isMustSplit() bool {
	t, _ := time.Parse(LOGGER_DATEFORMAT, time.Now().Format(LOGGER_DATEFORMAT))
	if t.After(*lg.date) {
		return true
	}
	return false
}

//切割日志
func (lg *Logger) split() {
	logFile := lg.filePath(time.Now().Format(LOGGER_DATEFORMAT))
	if !lg.isExist(logFile) {
		if lg.logFile != nil {
			lg.logFile.Close()
		}

		t, _ := time.Parse(LOGGER_DATEFORMAT, time.Now().Format(LOGGER_DATEFORMAT))
		lg.date = &t
		lg.logFile, _ = os.Create(logFile)
	}
}

func (lg *Logger) filePath(date string) string {
	filePath := lg.filePrefix + date + lg.fileSuffix
	return filePath
}

//文件监控
func (lg *Logger) fileMonitor() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Logger fileMonitor() catch panic: %v", err)
		}
	}()

	timer := time.NewTicker(time.Duration(LOGGER_LOG_SCAN) * time.Second)
	for {
		select {
		case <-timer.C:
			lg.fileCheck()
		}
	}
}

func (lg *Logger) fileCheck() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Logger fileCheck() catch panic: %v", err)
		}
	}()

	if lg.isMustSplit() {
		lg.mu.Lock()
		defer lg.mu.Unlock()
		lg.split()
	}
}

// Determine a file or a path exists in the os
func (lg *Logger) isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
