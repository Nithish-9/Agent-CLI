package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(service string) (*zap.Logger, func()) {

	home, err := os.UserHomeDir()
	if err != nil {
		home = "/tmp"
	}
	logPath := filepath.Join(home, ".agent", "logs", "agent.log")
	os.MkdirAll(filepath.Dir(logPath), 0755)

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return zap.NewNop(), func() {}
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(file),
		zap.InfoLevel,
	)

	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(&unbufferedWriter{os.Stdout}),
		zap.InfoLevel,
	)

	logger := zap.New(
		zapcore.NewTee(fileCore, consoleCore),
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	).With(
		zap.String("service", service),
	)

	cleanup := func() {
		logger.Sync()
		file.Close()
	}

	return logger, cleanup
}

type unbufferedWriter struct {
	w *os.File
}

func (u *unbufferedWriter) Write(p []byte) (n int, err error) {
	n, err = u.w.Write(p)
	u.w.Sync()
	return
}
