package observability

import (
	"bytes"
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLogger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logger Suite")
}

var _ = Describe("Logger", Label("scope:integration", "loop:g7-ops", "layer:server", "b:structured-logging", "r:high"), func() {
	Describe("NewLogger", func() {
		It("creates a logger with JSON output format", func() {
			logger := NewLogger()
			Expect(logger).NotTo(BeNil())
		})

		It("outputs structured JSON logs", func() {
			var buf bytes.Buffer
			config := zap.NewProductionConfig()
			config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
			config.OutputPaths = []string{"stdout"}
			config.ErrorOutputPaths = []string{"stderr"}
			
			// Create a logger that writes to buffer for testing
			core := zapcore.NewCore(
				zapcore.NewJSONEncoder(config.EncoderConfig),
				zapcore.AddSync(&buf),
				zapcore.InfoLevel,
			)
			testLogger := zap.New(core)
			logger := NewLoggerFromZap(testLogger)

			logger.Info("test message", "key", "value", "number", 42)

			// Verify JSON output
			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			Expect(err).NotTo(HaveOccurred())
			Expect(logEntry).To(HaveKey("msg"))
			Expect(logEntry["msg"]).To(Equal("test message"))
		})

		It("supports different log levels", func() {
			logger := NewLogger()
			Expect(logger).NotTo(BeNil())

			// These should not panic
			logger.Info("info message")
			logger.Error(nil, "error message")
			logger.V(1).Info("debug message")
		})

		It("supports context fields", func() {
			var buf bytes.Buffer
			config := zap.NewProductionConfig()
			config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
			core := zapcore.NewCore(
				zapcore.NewJSONEncoder(config.EncoderConfig),
				zapcore.AddSync(&buf),
				zapcore.InfoLevel,
			)
			testLogger := zap.New(core)
			logger := NewLoggerFromZap(testLogger)

			logger.WithValues("connection_id", "conn-123", "session_id", "sess-456").
				Info("connection established", "message_type", "connect")

			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			Expect(err).NotTo(HaveOccurred())
			Expect(logEntry).To(HaveKey("connection_id"))
			Expect(logEntry["connection_id"]).To(Equal("conn-123"))
			Expect(logEntry).To(HaveKey("session_id"))
			Expect(logEntry["session_id"]).To(Equal("sess-456"))
			Expect(logEntry).To(HaveKey("message_type"))
			Expect(logEntry["message_type"]).To(Equal("connect"))
		})
	})

	Describe("Log Levels", func() {
		It("uses ERROR level for errors", func() {
			var buf bytes.Buffer
			config := zap.NewProductionConfig()
			config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
			core := zapcore.NewCore(
				zapcore.NewJSONEncoder(config.EncoderConfig),
				zapcore.AddSync(&buf),
				zapcore.ErrorLevel,
			)
			testLogger := zap.New(core)
			logger := NewLoggerFromZap(testLogger)

			logger.Error(nil, "error occurred", "error", "test error")

			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			Expect(err).NotTo(HaveOccurred())
			Expect(logEntry).To(HaveKey("level"))
			Expect(logEntry["level"]).To(Equal("error"))
		})

		It("uses INFO level for informational messages", func() {
			var buf bytes.Buffer
			config := zap.NewProductionConfig()
			config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
			core := zapcore.NewCore(
				zapcore.NewJSONEncoder(config.EncoderConfig),
				zapcore.AddSync(&buf),
				zapcore.InfoLevel,
			)
			testLogger := zap.New(core)
			logger := NewLoggerFromZap(testLogger)

			logger.Info("info message")

			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			Expect(err).NotTo(HaveOccurred())
			Expect(logEntry).To(HaveKey("level"))
			Expect(logEntry["level"]).To(Equal("info"))
		})
	})
})

