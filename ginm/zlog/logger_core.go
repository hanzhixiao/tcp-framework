package zlog

import (
	"bytes"
	"fmt"
	"mmo/ginm/pkg/utils"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// (日志头部信息标记位，采用bitmap方式，用户可以选择头部需要哪些标记位被打印)
const (
	BitDate         = 1 << iota                            // Date flag bit 2019/01/23 (日期标记位)
	BitTime                                                // Time flag bit 01:23:12 (时间标记位)
	BitMicroSeconds                                        // Microsecond flag bit 01:23:12.111222 (微秒级标记位)
	BitLongFile                                            // Complete file name /home/go/src/zinx/main.go (完整文件名称)
	BitShortFile                                           // Last file name main.go (最后文件名)
	BitLevel                                               // Current log level: 0(Debug), 1(Info), 2(Warn), 3(Error), 4(Panic), 5(Fatal) (当前日志级别)
	BitStdFlag      = BitDate | BitTime                    // Standard log header format (标准头部日志格式)
	BitDefault      = BitLevel | BitShortFile | BitStdFlag // Default log header format (默认日志头部格式)
)

const (
	LOG_MAX_BUF = 1024 * 1024
)

// Log Level
const (
	LogDebug = iota
	LogInfo
	LogWarn
	LogError
	LogPanic
	LogFatal
)

var levels = []string{
	"[DEBUG]",
	"[INFO]",
	"[WARN]",
	"[ERROR]",
	"[PANIC]",
	"[FATAL]",
}

type LoggerCore struct {
	mu             sync.Mutex
	prefix         string
	flag           int
	buf            bytes.Buffer
	isolationLevel int
	callDepth      int
	fw             *utils.Writer
	onLogHook      func([]byte)
}

func NewZinxLog(prefix string, flag int) *LoggerCore {
	zlog := &LoggerCore{prefix: prefix, flag: flag, isolationLevel: 0, callDepth: 2}
	runtime.SetFinalizer(zlog, CleanZinxLog)
	return zlog
}

func CleanZinxLog(log *LoggerCore) {
	log.closeFile()
}

func (log *LoggerCore) closeFile() {
	if log.fw != nil {
		log.fw.Close()
	}
}

func (log *LoggerCore) Infof(format string, v ...interface{}) {
	if log.verifyLogIsolation(LogInfo) {
		return
	}
	_ = log.OutPut(LogInfo, fmt.Sprintf(format, v...))
}

func (log *LoggerCore) Info(v ...interface{}) {
	if log.verifyLogIsolation(LogInfo) {
		return
	}
	_ = log.OutPut(LogInfo, fmt.Sprintln(v...))
}

func (log *LoggerCore) Debugf(format string, v ...interface{}) {
	if log.verifyLogIsolation(LogDebug) {
		return
	}
	_ = log.OutPut(LogDebug, fmt.Sprintf(format, v...))
}
func (log *LoggerCore) Debug(v ...interface{}) {
	if log.verifyLogIsolation(LogDebug) {
		return
	}
	_ = log.OutPut(LogDebug, fmt.Sprintln(v...))
}

func (log *LoggerCore) Warnf(format string, v ...interface{}) {
	if log.verifyLogIsolation(LogWarn) {
		return
	}
	_ = log.OutPut(LogWarn, fmt.Sprintf(format, v...))
}

func (log *LoggerCore) Warn(v ...interface{}) {
	if log.verifyLogIsolation(LogWarn) {
		return
	}
	_ = log.OutPut(LogWarn, fmt.Sprintln(v...))
}

func (log *LoggerCore) Errorf(format string, v ...interface{}) {
	if log.verifyLogIsolation(LogError) {
		return
	}
	_ = log.OutPut(LogError, fmt.Sprintf(format, v...))
}

func (log *LoggerCore) Error(v ...interface{}) {
	if log.verifyLogIsolation(LogError) {
		return
	}
	_ = log.OutPut(LogError, fmt.Sprintln(v...))
}

func (log *LoggerCore) Fatalf(format string, v ...interface{}) {
	if log.verifyLogIsolation(LogFatal) {
		return
	}
	_ = log.OutPut(LogFatal, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (log *LoggerCore) Fatal(v ...interface{}) {
	if log.verifyLogIsolation(LogFatal) {
		return
	}
	_ = log.OutPut(LogFatal, fmt.Sprintln(v...))
	os.Exit(1)
}

func (log *LoggerCore) Panicf(format string, v ...interface{}) {
	if log.verifyLogIsolation(LogPanic) {
		return
	}
	s := fmt.Sprintf(format, v...)
	_ = log.OutPut(LogPanic, s)
	panic(s)
}

func (log *LoggerCore) Panic(v ...interface{}) {
	if log.verifyLogIsolation(LogPanic) {
		return
	}
	s := fmt.Sprintln(v...)
	_ = log.OutPut(LogPanic, s)
	panic(s)
}

func (log *LoggerCore) verifyLogIsolation(logLevel int) bool {
	return log.isolationLevel > logLevel
}

func (log *LoggerCore) OutPut(level int, s string) error {
	now := time.Now()
	var file string
	var line int
	log.mu.Lock()
	defer log.mu.Unlock()
	if log.flag&(BitShortFile|BitLongFile) != 0 {
		log.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(log.callDepth)
		if !ok {
			file = "unknown-file"
			line = 0
		}
		log.mu.Lock()
	}
	var err error
	log.buf.Reset()
	log.formatHeader(now, file, line, level)
	log.buf.WriteString(s)
	if len(s) > 0 && s[len(s)-1] != '\n' {
		log.buf.WriteByte('\n')
	}
	if log.fw == nil {
		_, _ = os.Stderr.Write(log.buf.Bytes())
	} else {
		_, err = log.fw.Write(log.buf.Bytes())
	}
	if log.onLogHook != nil {
		log.onLogHook(log.buf.Bytes())
	}
	return err
}

func (log *LoggerCore) formatHeader(t time.Time, file string, line int, level int) {
	var buf *bytes.Buffer = &log.buf
	// If the current prefix string is not empty, write the prefix first.
	if log.prefix != "" {
		buf.WriteByte('<')
		buf.WriteString(log.prefix)
		buf.WriteByte('>')
	}

	// If the time-related flags are set, add the time information to the log header.
	if log.flag&(BitDate|BitTime|BitMicroSeconds) != 0 {
		// Date flag is set
		if log.flag&BitDate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			buf.WriteByte('/') // "2019/"
			itoa(buf, int(month), 2)
			buf.WriteByte('/') // "2019/04/"
			itoa(buf, day, 2)
			buf.WriteByte(' ') // "2019/04/11 "
		}

		// Time flag is set
		if log.flag&(BitTime|BitMicroSeconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			buf.WriteByte(':') // "11:"
			itoa(buf, min, 2)
			buf.WriteByte(':') // "11:15:"
			itoa(buf, sec, 2)  // "11:15:33"
			// Microsecond flag is set
			if log.flag&BitMicroSeconds != 0 {
				buf.WriteByte('.')
				itoa(buf, t.Nanosecond()/1e3, 6) // "11:15:33.123123
			}
			buf.WriteByte(' ')
		}

		// Log level flag is set
		if log.flag&BitLevel != 0 {
			buf.WriteString(levels[level])
		}

		// Short file name flag or long file name flag is set
		if log.flag&(BitShortFile|BitLongFile) != 0 {
			// Short file name flag is set
			if log.flag&BitShortFile != 0 {
				short := file
				for i := len(file) - 1; i > 0; i-- {
					if file[i] == '/' {
						// Get the file name after the last '/' character, e.g. "zinx.go" from "/home/go/src/zinx.go"
						short = file[i+1:]
						break
					}
				}
				file = short
			}
			buf.WriteString(file)
			buf.WriteByte(':')
			itoa(buf, line, -1) // line number
			buf.WriteString(": ")
		}
	}
}

func (log *LoggerCore) Flags() int {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.flag
}

// ResetFlags resets the log Flags bitmap flags
// (重新设置日志Flags bitMap 标记位)
func (log *LoggerCore) ResetFlags(flag int) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.flag = flag
}

// AddFlag adds a flag to the bitmap flags
// (添加flag标记)
func (log *LoggerCore) AddFlag(flag int) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.flag |= flag
}

// SetPrefix sets a custom prefix for the log
// (设置日志的 用户自定义前缀字符串)
func (log *LoggerCore) SetPrefix(prefix string) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.prefix = prefix
}

// SetLogFile sets the log file output
// (设置日志文件输出)
func (log *LoggerCore) SetLogFile(fileDir string, fileName string) {
	if log.fw != nil {
		log.fw.Close()
	}
	log.fw = utils.New(filepath.Join(fileDir, fileName))
}

// SetMaxAge 最大保留天数
func (log *LoggerCore) SetMaxAge(ma int) {
	if log.fw == nil {
		return
	}
	log.mu.Lock()
	defer log.mu.Unlock()
	log.fw.SetMaxAge(ma)
}

// SetMaxSize 单个日志最大容量 单位：字节
func (log *LoggerCore) SetMaxSize(ms int64) {
	if log.fw == nil {
		return
	}
	log.mu.Lock()
	defer log.mu.Unlock()
	log.fw.SetMaxSize(ms)
}

// SetCons 同时输出控制台
func (log *LoggerCore) SetCons(b bool) {
	if log.fw == nil {
		return
	}
	log.mu.Lock()
	defer log.mu.Unlock()
	log.fw.SetCons(b)
}

func (log *LoggerCore) SetLogLevel(logLevel int) {
	log.isolationLevel = logLevel
}

func (log *LoggerCore) Stack(v ...interface{}) {
	s := fmt.Sprint(v...)
	s += "\n"
	buf := make([]byte, LOG_MAX_BUF)
	n := runtime.Stack(buf, true) //得到当前堆栈信息
	s += string(buf[:n])
	s += "\n"
	_ = log.OutPut(LogError, s)
}

func itoa(buf *bytes.Buffer, i int, wID int) {
	var u uint = uint(i)
	if u == 0 && wID <= 1 {
		buf.WriteByte('0')
		return
	}

	// Assemble decimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; u > 0 || wID > 0; u /= 10 {
		bp--
		wID--
		b[bp] = byte(u%10) + '0'
	}

	// avoID slicing b to avoID an allocation.
	for bp < len(b) {
		buf.WriteByte(b[bp])
		bp++
	}
}
