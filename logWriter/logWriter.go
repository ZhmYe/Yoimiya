package logWriter

import (
	"bufio"
	"fmt"
	//"main/src/Config"
	"os"
	"time"
)

type LogWriter struct {
	path   string
	writer *bufio.Writer
	file   *os.File
}

func NewLogWriter(path string) *LogWriter {
	logger := new(LogWriter)
	format := "2006-01-02-15-04-05"
	currentTime := time.Now()
	filePath := "/root/Yoimiya/logWriter/log/" + path + "_" + currentTime.Format(format) + ".txt"
	//fmt.Println(filePath)
	logger.path = filePath
	var err error
	logger.file, err = os.OpenFile(logger.path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("open file err=%v\n", err)
		panic(err)
	}
	writer := bufio.NewWriter(logger.file)
	logger.writer = writer
	return logger
}
func (l *LogWriter) Write(str string) {
	_, err := l.writer.WriteString(str)
	if err != nil {
		return
	}
}
func (l *LogWriter) Writeln(str string) {
	_, err := l.writer.WriteString(str)
	if err != nil {
		return
	}
	l.Wrap()
}
func (l *LogWriter) Finish() {
	err := l.writer.Flush()
	if err != nil {
		return
	}
	err = l.file.Close()
	if err != nil {
		return
	}
}
func (l *LogWriter) Wrap() {
	l.Write("\n")
}
