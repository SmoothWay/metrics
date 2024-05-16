package logger

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	mockLogger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.InfoLevel,
	))
	loggger := &Logger{logger: mockLogger}

	// Call the method being tested
	loggger.Info("test message", zap.String("key", "value"))

	// Check the logged message
	assert.Contains(t, buf.String(), `"level":"info"`)
	assert.Contains(t, buf.String(), `"msg":"test message"`)
	assert.Contains(t, buf.String(), `"key":"value"`)
}
func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	mockLogger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.WarnLevel,
	))
	loggger := &Logger{logger: mockLogger}

	loggger.Warn("test warning", zap.String("key", "value"))

	assert.Contains(t, buf.String(), `"level":"warn"`)
	assert.Contains(t, buf.String(), `"msg":"test warning"`)
	assert.Contains(t, buf.String(), `"key":"value"`)
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	mockLogger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.ErrorLevel,
	))
	loggger := &Logger{logger: mockLogger}

	loggger.Error("test error", zap.String("key", "value"))

	assert.Contains(t, buf.String(), `"level":"error"`)
	assert.Contains(t, buf.String(), `"msg":"test error"`)
	assert.Contains(t, buf.String(), `"key":"value"`)
}
func TestLog(t *testing.T) {
	// Assuming that log is already initialized
	err := Init("info")
	assert.NoError(t, err)
	assert.NotNil(t, log)
}
