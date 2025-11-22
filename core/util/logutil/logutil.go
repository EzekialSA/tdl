package logutil

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type OutputMode string

const (
	OutputFile   OutputMode = "file"
	OutputStderr OutputMode = "stderr"
	OutputBoth   OutputMode = "both"
)

// noSyncWriter wraps a WriteSyncer and makes Sync a no-op
type noSyncWriter struct {
	zapcore.WriteSyncer
}

func (w noSyncWriter) Sync() error {
	return nil
}

func New(level zapcore.LevelEnabler, path string) *zap.Logger {
	return NewWithOutput(level, path, OutputFile)
}

func NewWithOutput(level zapcore.LevelEnabler, path string, output OutputMode) *zap.Logger {
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	config.EncodeLevel = zapcore.CapitalLevelEncoder

	var cores []zapcore.Core

	// Add file output if needed
	if output == OutputFile || output == OutputBoth {
		rotate := &lumberjack.Logger{
			Filename:   path,
			MaxSize:    10,
			MaxAge:     7,
			MaxBackups: 3,
			LocalTime:  true,
			Compress:   true,
		}
		fileWriter := zapcore.AddSync(rotate)
		cores = append(cores, zapcore.NewCore(zapcore.NewConsoleEncoder(config), fileWriter, level))
	}

	// Add stderr output if needed
	if output == OutputStderr || output == OutputBoth {
		// Wrap stderr with noSyncWriter to prevent sync errors on Windows
		stderrWriter := noSyncWriter{zapcore.Lock(os.Stderr)}
		cores = append(cores, zapcore.NewCore(zapcore.NewConsoleEncoder(config), stderrWriter, level))
	}

	// Combine cores
	core := zapcore.NewTee(cores...)
	return zap.New(core, zap.AddCaller())
}
