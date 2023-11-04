package unit

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"
)

const (
	Ldate = 1 << iota
	Ltime
	Lmicroseconds
	DAY     = time.Hour * 24
	MONTH   = DAY * 30
	_TRACE_ = "[TRACE]"
	_DEBUG_ = "[DEBUG]"
	_INFO_  = "[INFO]"
	_WARN_  = "[WARN]"
	_ERROR_ = "[ERROR]"
	_COST_  = "[COST]"
)

type Logger struct {
	out *os.File
	*WorkerPool
	flag       int
	maxAge     time.Duration
	rotateTime time.Duration
	buf        []byte
	fileName   string
	fileTime   time.Time
}

var g_loggge *Logger

func NewLog(file string, maxAge, rotateTime time.Duration) (err error) {
	g_loggge, err = NewLogger(file, maxAge, rotateTime)
	return
}
func NewLogger(file string, maxAge, rotateTime time.Duration) (*Logger, error) {
	r := &Logger{
		flag:       Ltime | Lmicroseconds,
		fileName:   file,
		fileTime:   time.Now(),
		maxAge:     maxAge,
		rotateTime: rotateTime,
		WorkerPool: NewWorkerPool(-1, 1),
	}

	if 0 >= r.maxAge {
		r.maxAge = MONTH
	}
	if 0 >= r.rotateTime {
		r.rotateTime = time.Hour
	}
	if r.rotateTime > r.maxAge {
		r.rotateTime, r.maxAge = r.maxAge, r.rotateTime
	}

	err := r.rotateFile(r.fileTime)
	return r, err
}

func (r *Logger) SetDefLog() {
	g_loggge = r
}

func (r *Logger) Close() {
	r.WorkerPool.Close()
	if nil != r.out {
		r.out.Close()
	}
}
func (r *Logger) close() {
	r.Close()
	os.Exit(-1)
}

type logNode struct {
	logger   *Logger
	data     []any
	prefix   string
	pushTime time.Time
}

func (r *logNode) Callback() {
	now := time.Now()
	if now.After(r.logger.fileTime) {
		if err := r.logger.rotateFile(now); nil != err {
			output(r.logger, &logNode{
				logger:   r.logger,
				data:     []any{err},
				prefix:   "[ERROR]",
				pushTime: now,
			})
		}
		if err := r.logger.delFile(now); nil != err {
			output(r.logger, &logNode{
				logger:   r.logger,
				data:     []any{err},
				prefix:   "[ERROR]",
				pushTime: now,
			})
		}
	}
	output(r.logger, r)
}

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

func newFileExt(t time.Time, duration time.Duration) string {
	str := t.Format(time.RFC3339)
	if time.Minute <= duration && duration < time.Hour {
		return str[:len("2006-01-02T15:04")]
	} else if time.Hour <= duration && duration < time.Hour*24 {
		return str[:len("2006-01-02T15")]
	} else if time.Hour*24 <= duration {
		return str[:len("2006-01-02")]
	}
	return str
}

func (r *Logger) rotateFile(now time.Time) error {
	r.fileTime = now
	oldFileName := r.fileName + "." + newFileExt(r.fileTime, r.rotateTime)
	fp, err := absoluteFile(oldFileName)
	if nil != err {
		return err
	}
	os.Remove(r.fileName)
	os.Link(oldFileName, r.fileName)
	if nil != r.out {
		r.out.Close()
	}
	r.out = fp
	return nil
}
func (r *Logger) delFile(now time.Time) error {
	files, err := filepath.Glob(r.fileName + ".*")
	if nil != err {
		return err
	}
	for _, file := range files {
		info, err := os.Stat(file)
		if nil != err {
			return err
		}
		if info.ModTime().Add(r.maxAge).Before(now) {
			if err = os.Remove(file); nil != err {
				return err
			}
		}
	}
	return nil
}
func absoluteFile(name string) (*os.File, error) {
	dir := path.Dir(name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); nil != err {
			return nil, err
		}
	}
	return os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
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
	if l.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if l.flag&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '-')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '-')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
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
