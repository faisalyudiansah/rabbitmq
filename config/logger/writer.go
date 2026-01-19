package logger

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogWriter struct {
	lock     sync.Mutex
	filename string
	output   string
	writer   *os.File
	curDate  string
	fullpath string
}

func NewLogWriter(output, filename string) (*LogWriter, error) {
	w := &LogWriter{
		output:   output,
		filename: filename,
	}
	err := w.rotate()
	return w, err
}

func (w *LogWriter) Write(s []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	today := time.Now().Format("2006-01-02")
	if today != w.curDate {
		w.rotate()
	}
	return w.writer.Write(s)
}

func (w *LogWriter) rotate() error {
	if w.writer != nil {
		w.writer.Close()
	}
	now := time.Now()
	dir := filepath.Join(w.output, now.Format("2006"), now.Format("01"), now.Format("02"))
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	w.fullpath = filepath.Join(dir, w.filename+".log")
	file, err := os.OpenFile(w.fullpath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	w.writer = file
	w.curDate = now.Format("2006-01-02")
	return nil
}

func (w *LogWriter) GetOutputfilepath() string {
	return w.fullpath
}
