package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path"
)

func InitLogger() *zap.SugaredLogger {
	config := zap.NewProductionEncoderConfig()
	config.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncodeTime = zapcore.TimeEncoderOfLayout("2006.01.02 | 15:04:05")

	fileEncoder := zapcore.NewJSONEncoder(config)
	consoleEncoder := zapcore.NewConsoleEncoder(config)

	err := os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		panic(err)
	}
	logFilePath := path.Join("logs", "debug.log")

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    100, // megabytes
		MaxBackups: 15,
		MaxAge:     3, // days
	})

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, fileWriter, zapcore.DebugLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
	)
	logger := zap.New(core, zap.AddCaller())

	defer logger.Sync()

	return logger.Sugar()
}
