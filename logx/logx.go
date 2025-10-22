package logx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	baseDir      string
	logLevel     zapcore.Level
	maxDays      int
	autoClean    bool
	currentDate  string
	flushOnWrite bool
	callerSkip   int
	loggerMap    map[string]*zap.Logger
	mu           sync.Mutex
}

func NewLogger(level string, callerSkip int, dir string, keepDays int, clean bool, flush bool) *Logger {
	l := &Logger{
		baseDir:      dir,
		maxDays:      keepDays,
		autoClean:    clean,
		flushOnWrite: flush,
		callerSkip:   callerSkip,
		loggerMap:    make(map[string]*zap.Logger),
		currentDate:  time.Now().Format("2006-01-02"),
	}

	_ = os.MkdirAll(filepath.Join(l.baseDir, l.currentDate), 0755)

	switch strings.ToLower(level) {
	case "debug":
		l.logLevel = zap.DebugLevel
	case "info":
		l.logLevel = zap.InfoLevel
	case "warn":
		l.logLevel = zap.WarnLevel
	case "error":
		l.logLevel = zap.ErrorLevel
	default:
		l.logLevel = zap.InfoLevel
	}

	if l.autoClean {
		go l.cleanOldLogs()
	}

	return l
}

func (l *Logger) getLogger(fileName string) *zap.Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	if today != l.currentDate {
		l.currentDate = today
		l.loggerMap = make(map[string]*zap.Logger)
		_ = os.MkdirAll(filepath.Join(l.baseDir, l.currentDate), 0755)
	}

	if logger, exists := l.loggerMap[fileName]; exists {
		return logger
	}

	logPath := filepath.Join(l.baseDir, l.currentDate, fmt.Sprintf("%s.log", fileName))
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("无法创建日志文件: %v", err))
	}

	// 创建 MultiWriter 实现同时输出到文件和控制台
	console := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:      "datetime",
			LevelKey:     "level",
			NameKey:      "logger",
			CallerKey:    "caller",
			MessageKey:   "message",
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format(
					"2006-01-02 15:04:05.000"))
			}}),
		zapcore.AddSync(os.Stdout),
		// 输出到终端
		zap.DebugLevel,
	)

	fileEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:      "datetime",
		LevelKey:     "level",
		NameKey:      "logger",
		CallerKey:    "caller",
		MessageKey:   "message",
		EncodeLevel:  zapcore.LowercaseLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		},
	})

	// 写入文件和终端的 logCore
	logCore := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(file), l.logLevel),
		console,
	)

	// 生成 logger
	logger := zap.New(logCore, zap.AddCaller(), zap.AddCallerSkip(l.callerSkip))
	l.loggerMap[fileName] = logger
	return logger
}

func (l *Logger) Info(fileName, module, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String("module", module))
	l.getLogger(fileName).Info(msg, fields...)
	if l.flushOnWrite {
		l.Sync()
	}
}

func (l *Logger) Debug(fileName, module, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String("module", module))
	l.getLogger(fileName).Debug(msg, fields...)
	if l.flushOnWrite {
		l.Sync()
	}
}

func (l *Logger) Warn(fileName, module, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String("module", module))
	l.getLogger(fileName).Warn(msg, fields...)
	if l.flushOnWrite {
		l.Sync()
	}
}

func (l *Logger) Error(fileName, module, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String("module", module))
	l.getLogger(fileName).Error(msg, fields...)
	if l.flushOnWrite {
		l.Sync()
	}
}

func (l *Logger) Sync() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, logger := range l.loggerMap {
		_ = logger.Sync()
	}
}

func (l *Logger) cleanOldLogs() {
	if l.maxDays <= 0 {
		return
	}

	for {
		dirs, err := os.ReadDir(l.baseDir)
		if err != nil {
			time.Sleep(24 * time.Hour)
			continue
		}

		expire := time.Now().AddDate(0, 0, -l.maxDays)
		for _, dir := range dirs {
			if !dir.IsDir() {
				continue
			}
			dirTime, err := time.Parse("2006-01-02", dir.Name())
			if err == nil && dirTime.Before(expire) {
				_ = os.RemoveAll(filepath.Join(l.baseDir, dir.Name()))
			}
		}
		time.Sleep(24 * time.Hour)
	}
}
