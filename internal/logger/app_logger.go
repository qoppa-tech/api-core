package logger

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type AppLogger struct {
	isProd bool
	level  Level
}

var defaultLogger *AppLogger

func Init() {
	env := os.Getenv("APP_ENV")
	isProd := env == "production"

	level := LevelDebug
	if isProd {
		level = LevelInfo
	}

	defaultLogger = &AppLogger{
		isProd: isProd,
		level:  level,
	}
}

func GetLogger() *AppLogger {
	if defaultLogger == nil {
		Init()
	}
	return defaultLogger
}

func Debug(msg string, fields ...Field) {
	GetLogger().log(LevelDebug, msg, fields...)
}

func Info(msg string, fields ...Field) {
	GetLogger().log(LevelInfo, msg, fields...)
}

func Warn(msg string, fields ...Field) {
	GetLogger().log(LevelWarn, msg, fields...)
}

func Error(msg string, fields ...Field) {
	GetLogger().log(LevelError, msg, fields...)
}

type Field struct {
	Key   string
	Value interface{}
}

func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

func (l *AppLogger) log(level Level, msg string, fields ...Field) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format(time.RFC3339)
	levelStr := levelToString(level)

	if l.isProd {
		l.logJSON(timestamp, levelStr, msg, fields)
	} else {
		l.logDebug(timestamp, levelStr, msg, fields)
	}
}

func (l *AppLogger) logJSON(timestamp, level, msg string, fields []Field) {
	json := fmt.Sprintf(`{"timestamp":"%s","level":"%s","message":"%s"`, timestamp, level, escapeJSON(msg))

	if _, file, line, ok := runtime.Caller(3); ok {
		json += fmt.Sprintf(`,"caller":"%s:%d"`, file, line)
	}

	for _, f := range fields {
		if f.Value == nil {
			json += fmt.Sprintf(`,"%s":null`, f.Key)
		} else {
			switch v := f.Value.(type) {
			case string:
				json += fmt.Sprintf(`,"%s":"%s"`, f.Key, escapeJSON(v))
			case int, int32, int64, uint, uint32, uint64:
				json += fmt.Sprintf(`,"%s":%d`, f.Key, v)
			case float32, float64:
				json += fmt.Sprintf(`,"%s":%f`, f.Key, v)
			case bool:
				json += fmt.Sprintf(`,"%s":%t`, f.Key, v)
			default:
				json += fmt.Sprintf(`,"%s":"%v"`, f.Key, v)
			}
		}
	}

	json += "}"
	fmt.Fprintln(os.Stdout, json)
}

func (l *AppLogger) logDebug(timestamp, level, msg string, fields []Field) {
	color := getLevelColor(level)
	reset := "\033[0m"

	caller := ""
	if _, file, line, ok := runtime.Caller(3); ok {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		caller = fmt.Sprintf(" \033[90m%s:%d\033[0m", short, line)
	}

	output := fmt.Sprintf("%s | %s%s%s%s | %s", timestamp, color, level, reset, caller, msg)

	if len(fields) > 0 {
		output += " \033[90m|"
		for _, f := range fields {
			output += fmt.Sprintf(" %s=%v", f.Key, f.Value)
		}
		output += reset
	}

	fmt.Fprintln(os.Stdout, output)
}

func levelToString(level Level) string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func getLevelColor(level string) string {
	switch level {
	case "DEBUG":
		return "\033[96m" // Cyan
	case "INFO":
		return "\033[92m" // Green
	case "WARN":
		return "\033[93m" // Yellow
	case "ERROR":
		return "\033[91m" // Red
	default:
		return "\033[97m" // White
	}
}
