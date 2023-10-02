package zaplog

import (
	"errors"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	folderName = "Logs"
	filename   = "current.log"
)

var (
	ErrFailedToMakeLogger = errors.New("failed to build logger")
)

func NewLogger() (*zap.Logger, error) {
	config := makeEncoderConfig()
	path, err := logFilepath()
	if err != nil {
		return nil, errors.Join(ErrFailedToMakeLogger, err)
	}

	core, err := makeCore(path, config)
	if err != nil {
		return nil, errors.Join(ErrFailedToMakeLogger, err)
	}

	return newLogger(core, path), nil
}

func makeCore(path string, config zapcore.EncoderConfig) (zapcore.Core, error) {
	rollingWriter := &lumberjack.Logger{
		Filename: path,
		MaxSize:  50,
		MaxAge:   14,
		Compress: true,
	}

	if err := rollingWriter.Rotate(); err != nil {
		return nil, errors.Join(errors.New("error rotating log file"), err)
	}

	var (
		consoleEncoder = zapcore.NewConsoleEncoder(config)
		jsonEncoder    = zapcore.NewJSONEncoder(config)

		consoleCore = zapcore.NewCore(consoleEncoder, zapcore.AddSync(zapcore.Lock(os.Stdout)), zap.DebugLevel)
		jsonCore    = zapcore.NewCore(jsonEncoder, zapcore.AddSync(rollingWriter), zapcore.ErrorLevel)

		core = zapcore.NewTee(consoleCore, jsonCore)
	)

	return core, nil
}

func makeEncoderConfig() zapcore.EncoderConfig {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return encoderConfig
}

func logFilepath() (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	executableDir := filepath.Dir(executablePath)
	dir := tryMakeLogDir(executableDir, folderName)
	path := filepath.Join(dir, filename)

	return path, nil
}

func tryMakeLogDir(path, dirName string) string {
	logFolderPath := filepath.Join(path, dirName)
	if err := os.MkdirAll(logFolderPath, os.ModePerm); err != nil {
		return ""
	}

	return logFolderPath
}

func newLogger(core zapcore.Core, path string) *zap.Logger {
	logger := zap.New(core, zap.AddCaller())
	logger.Info("logger initialized", zap.String("directory", filepath.Dir(path)))
	return logger
}
