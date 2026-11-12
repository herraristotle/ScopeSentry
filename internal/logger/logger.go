package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

// Add simplified methods
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

func init() {
	// Parse the log level
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(config.GlobalConfig.Log.Level)); err != nil {
		fmt.Println(err)
		return
	}

	// Load the timezone
	location, err := time.LoadLocation(config.GlobalConfig.System.Timezone)
	if err != nil {
		// If loading fails, use the local timezone
		location = time.Local
		fmt.Println("load Timezone error")
	}

	// Encoder configuration
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalColorLevelEncoder, // colored level
		EncodeTime: func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			// Format time using the timezone
			t = t.In(location)
			encoder.AppendString(t.Format("2006-01-02 15:04:05"))
		}, // ISO time format
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	// Build the output sinks
	var cores []zapcore.Core

	// Always output to the console
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
	cores = append(cores, consoleCore)

	// If a log file path is configured, also output to the file
	if config.GlobalConfig.Log.Filename != "" {
		logDir := filepath.Dir(config.GlobalConfig.Log.Filename)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Println(err)
			return
		}

		fileWriter := &lumberjack.Logger{
			Filename: config.GlobalConfig.Log.Filename,
			MaxSize:  300, // in MB
			Compress: true,
		}

		fileCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(fileWriter), level)
		cores = append(cores, fileCore)
	}

	// Merge all outputs
	core := zapcore.NewTee(cores...)

	// Create the Logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return
}
