package unit

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	DAY     = time.Hour * 24
	WEEK    = DAY * 7
	MONTH   = DAY * 30
	_TRACE_ = "[TRACE]"
	_DEBUG_ = "[DEBUG]"
	_INFO_  = "[INFO]"
	_WARN_  = "[WARN]"
	_ERROR_ = "[ERROR]"
	_SEP_   = "."
)

func (r *Logger) Trace(v ...any) { r.cache <- &logNode{data: v, prefix: _TRACE_, pushTime: time.Now()} }
func (r *Logger) Debug(v ...any) { r.cache <- &logNode{data: v, prefix: _DEBUG_, pushTime: time.Now()} }
func (r *Logger) Info(v ...any)  { r.cache <- &logNode{data: v, prefix: _INFO_, pushTime: time.Now()} }
func (r *Logger) Warn(v ...any)  { r.cache <- &logNode{data: v, prefix: _WARN_, pushTime: time.Now()} }
func (r *Logger) Error(v ...any) { r.cache <- &logNode{data: v, prefix: _ERROR_, pushTime: time.Now()} }
func (r *Logger) Print(prefix string, v ...any) {
	r.cache <- &logNode{data: v, prefix: prefix, pushTime: time.Now()}
}

func Trace(v ...any)                { g_loggge.Trace(v...) }
func Debug(v ...any)                { g_loggge.Debug(v...) }
func Info(v ...any)                 { g_loggge.Info(v...) }
func Warn(v ...any)                 { g_loggge.Warn(v...) }
func Error(v ...any)                { g_loggge.Error(v...) }
func Print(prefix string, v ...any) { g_loggge.Print(prefix, v...) }

type (
	fileEntry struct {
		createTime time.Time
		path       string
	}
	logNode struct {
		data     []any
		prefix   string
		pushTime time.Time
	}
	Logger struct {
		out            *os.File
		cache          chan *logNode
		stop           chan bool
		maxAge         time.Duration
		rotateInterval time.Duration
		buf            []byte
		filePrefix     string
		createdFile    []*fileEntry
	}
)

func CloseLog() {
	g_loggge.Close()
}
func (r *Logger) Close() {
	close(r.cache)
	<-r.stop
	if nil != r.out {
		r.out.Close()
	}
}

var g_loggge *Logger

func NewLog(dir string, rotateInterval, maxAge time.Duration) (err error) {
	g_loggge, err = NewSyncLog(dir, rotateInterval, maxAge)
	return err
}

func NewSyncLog(dir string, rotateInterval, maxAge time.Duration) (*Logger, error) {
	r := &Logger{
		filePrefix:     path.Join(dir, "log"),
		maxAge:         maxAge,
		rotateInterval: rotateInterval,
		cache:          make(chan *logNode, 8192),
		stop:           make(chan bool),
	}

	if time.Second > r.rotateInterval {
		r.rotateInterval = time.Second
	}
	if r.rotateInterval >= r.maxAge {
		r.maxAge = r.rotateInterval + 3*time.Second
	}

	//将文件夹中的文件按时间排序
	if matches, err := filepath.Glob(r.filePrefix + _SEP_ + "*"); nil == err {
		for _, match := range matches {
			info, err := os.Stat(match)
			if nil == err && nil != info {
				r.createdFile = append(r.createdFile, &fileEntry{createTime: info.ModTime(), path: match})
			}
		}
		sort.Slice(r.createdFile, func(i, j int) bool { return r.createdFile[i].createTime.Before(r.createdFile[j].createTime) })
	}
	if err := r.rotateFile(); nil != err {
		return nil, err
	}
	go r.output()
	return r, nil
}

func (r *Logger) rotateFile() error {
	now := time.Now()
	file := r.filePrefix + _SEP_ + strings.ReplaceAll(now.Format(time.RFC3339)[:len(time.DateTime)], ":", "-")
	if info, err := os.Stat(file); nil == info || os.IsNotExist(err) {
		dir := path.Dir(file)
		if info, err = os.Stat(dir); nil == info || os.IsNotExist(err) {
			if err = os.MkdirAll(dir, 0755); nil != err {
				return err
			}
		}
		r.createdFile = append(r.createdFile, &fileEntry{createTime: now, path: file})
	}

	fp, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if nil != err {
		return err
	}
	os.Remove(r.filePrefix)
	os.Link(file, r.filePrefix)
	if nil != r.out {
		r.out.Close()
	}
	r.out = fp

	for i := 0; i < len(r.createdFile); i++ {
		if r.createdFile[i].createTime.Add(r.maxAge).After(now) {
			break
		}
		if nil == os.Remove(r.createdFile[i].path) {
			r.createdFile = r.createdFile[1:]
		}
	}
	return err
}

func (r *Logger) output() {
	t := time.Tick(r.rotateInterval)
	for {
		select {
		case <-t:
			r.rotateFile()
		case node, ok := <-r.cache:
			if !ok {
				r.stop <- true
				close(r.stop)
				return
			}
			output(r, node)
		}
	}
}

func itoa(buf *[]byte, i int, wid int) {
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func (l *Logger) formatHeader(buf *[]byte, t time.Time, prefix string) {
	h, m, s := t.Clock()
	itoa(buf, h, 2)
	*buf = append(*buf, ':')
	itoa(buf, m, 2)
	*buf = append(*buf, ':')
	itoa(buf, s, 2)
	*buf = append(*buf, '.')
	itoa(buf, t.Nanosecond()/1e3, 6)
	*buf = append(*buf, ' ')
	*buf = append(*buf, prefix...)
	*buf = append(*buf, ' ')
}

func output(l *Logger, node *logNode) {
	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, node.pushTime, node.prefix)
	l.buf = fmt.Appendln(l.buf, node.data...)
	if len(l.buf) > 1 && l.buf[len(l.buf)-2] == '\n' {
		l.buf = l.buf[:len(l.buf)-1]
	}
	l.out.Write(l.buf)
}
