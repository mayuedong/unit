package unit

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
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
	_COST_  = "[COST]"
)

func (r *Logger) Trace(v ...any) {
	r.Push(&logNode{logger: r, data: v, prefix: _TRACE_, pushTime: time.Now()})
}
func (r *Logger) Debug(v ...any) {
	r.Push(&logNode{logger: r, data: v, prefix: _DEBUG_, pushTime: time.Now()})
}
func (r *Logger) Info(v ...any) {
	r.Push(&logNode{logger: r, data: v, prefix: _INFO_, pushTime: time.Now()})
}
func (r *Logger) Warn(v ...any) {
	r.Push(&logNode{logger: r, data: v, prefix: _WARN_, pushTime: time.Now()})
}
func (r *Logger) Error(v ...any) {
	r.Push(&logNode{logger: r, data: v, prefix: _ERROR_, pushTime: time.Now()})
}

func Cost(v ...any) {
	g_loggge.Push(&logNode{logger: g_loggge, data: v, prefix: _COST_, pushTime: time.Now()})
}
func Trace(v ...any) {
	g_loggge.Push(&logNode{logger: g_loggge, data: v, prefix: _TRACE_, pushTime: time.Now()})
}
func Debug(v ...any) {
	g_loggge.Push(&logNode{logger: g_loggge, data: v, prefix: _DEBUG_, pushTime: time.Now()})
}
func Info(v ...any) {
	g_loggge.Push(&logNode{logger: g_loggge, data: v, prefix: _INFO_, pushTime: time.Now()})
}
func Warn(v ...any) {
	g_loggge.Push(&logNode{logger: g_loggge, data: v, prefix: _WARN_, pushTime: time.Now()})
}
func Error(v ...any) {
	g_loggge.Push(&logNode{logger: g_loggge, data: v, prefix: _ERROR_, pushTime: time.Now()})
}

type fileEntry struct {
	createTime time.Time
	path       string
}
type Logger struct {
	out *os.File
	*WorkerPool
	maxAge         time.Duration
	rotateInterval time.Duration
	buf            []byte
	filePrefix     string
	rotateAt       time.Time
	createdFile    List[*fileEntry]
}

func Close() {
	g_loggge.WorkerPool.Close()
}
func (r *Logger) Close() {
	r.WorkerPool.Close()
	if nil != r.out {
		r.out.Close()
	}
}

var g_loggge = &Logger{
	out:        os.Stdout,
	WorkerPool: NewWorkerPool(-1, 1),
}

func NewLog(file string, rotateInterval, maxAge time.Duration) (err error) {
	g_loggge, err = NewSyncLog(file, rotateInterval, maxAge)
	return err
}

func NewSyncLog(file string, rotateInterval, maxAge time.Duration) (*Logger, error) {
	r := &Logger{
		filePrefix:     file,
		maxAge:         maxAge,
		rotateInterval: rotateInterval,
		WorkerPool:     NewWorkerPool(-1, 1),
	}
	r.createdFile.Init(-1)

	if time.Hour > r.rotateInterval {
		r.rotateInterval = time.Hour
	}
	if r.rotateInterval >= r.maxAge {
		r.maxAge = r.rotateInterval + time.Hour
	}

	//将文件夹中的文件按时间排序
	if matches, err := filepath.Glob(r.filePrefix + ".*"); nil == err {
		for _, match := range matches {
			info, err := os.Stat(match)
			if nil != err || nil == info {
				continue
			}
			entry := &fileEntry{
				createTime: info.ModTime(),
				path:       match,
			}

			if oldNode := r.createdFile.FromHeadFindNode(func(cacheFile *fileEntry) bool {
				return info.ModTime().Before(cacheFile.createTime)
			}); oldNode != nil {
				r.createdFile.InsertPrev(entry, oldNode)
			} else {
				r.createdFile.Push(entry)
			}
		}
	}
	return r, r.rotateFile(time.Now())
}

func (r *Logger) rotateFile(now time.Time) error {
	file := r.filePrefix + "." + now.Format(time.RFC3339)[:len("2006-01-02T15")]
	fileExist := true
	for i := 0; i < 1; i++ {
		if info, err := os.Stat(file); nil == info || os.IsNotExist(err) {
			fileExist = false
		} else {
			break
		}

		dir := path.Dir(file)
		if info, err := os.Stat(dir); nil == info || os.IsNotExist(err) {
			if err = os.MkdirAll(dir, 0755); nil != err {
				return err
			}
		}
	}
	fp, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if nil != err {
		return err
	}
	if !fileExist {
		r.createdFile.Push(&fileEntry{
			createTime: now,
			path:       file,
		})
	}

	os.Remove(r.filePrefix)
	os.Link(file, r.filePrefix)
	if nil != r.out {
		r.out.Close()
	}
	r.out = fp
	r.rotateAt = now

	for entry, ok := r.createdFile.GetHead(); ok && entry.createTime.Add(r.maxAge).Before(now); entry, ok = r.createdFile.GetHead() {
		r.createdFile.PopHead()
		if err = os.Remove(entry.path); nil != err {
			return err
		}
	}
	return err
}

type logNode struct {
	logger   *Logger
	data     []any
	prefix   string
	pushTime time.Time
}

func (r *logNode) Callback() {
	output(r.logger, r)
	if len(r.logger.filePrefix) == 0 {
		return
	}
	if r.pushTime.Sub(r.logger.rotateAt) >= r.logger.rotateInterval {
		now := time.Now()
		if err := r.logger.rotateFile(now); nil != err {
			output(r.logger, &logNode{
				logger:   r.logger,
				data:     []any{err},
				prefix:   "[ERROR]",
				pushTime: now,
			})
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
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func (l *Logger) formatHeader(buf *[]byte, t time.Time, prefix string) {
	hour, min, sec := t.Clock()
	itoa(buf, hour, 2)
	*buf = append(*buf, ':')
	itoa(buf, min, 2)
	*buf = append(*buf, ':')
	itoa(buf, sec, 2)
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
