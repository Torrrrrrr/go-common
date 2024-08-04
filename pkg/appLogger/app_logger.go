package applogger

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	AppLogger interface {
		WithContext(ctx *gin.Context) AppLogger
		WithField(key FieldKey, value string) AppLogger

		// Return new logger with configuration from current logger,
		// exclude current context and fields
		Clone() AppLogger

		// Log message and results to [Error] level
		Error(message string, refID any, serviceName string, result any)
		// Log message and results to [Warn] level
		Warn(message string, refID any, serviceName string, result any)
		// Log message and results to [Info] level
		Info(message string, refID any, serviceName string, result any)
		// Log message and results to [Debug] level
		Debug(message string, refID any, serviceName string, result any)

		// Log request
		LogRequest()
		// Log response
		LogResponse()
	}

	AppLoggerImpl struct {
		zapLogger *zap.Logger
		ctx       *gin.Context
		fields    map[string]string

		// For log request & response
		t            time.Time
		requestBody  string
		responseBody *bodyLogWriter
	}

	// Response body writer
	bodyLogWriter struct {
		gin.ResponseWriter
		body *bytes.Buffer
	}
)

type FieldKey string

const (
	RefID       FieldKey = "refID"
	AppName     FieldKey = "appName"
	ServiceName FieldKey = "serviceName"
)

func (a *AppLoggerImpl) WithContext(ctx *gin.Context) AppLogger {
	a.ctx = ctx
	return a
}

func (a *AppLoggerImpl) WithField(key FieldKey, value string) AppLogger {
	a.fields[string(key)] = value
	return a
}

func (a *AppLoggerImpl) Clone() AppLogger {
	return NewAppLogger(a.zapLogger.Level().String())
}

func (a *AppLoggerImpl) generateFields() (field []zap.Field) {
	// Generate log id
	const LOG_ID = "logID"
	field = append(field, zap.Int64(LOG_ID, time.Now().UnixNano()))

	if a.ctx != nil {
		// Standard fields from context
		// session_id
		const SID = "sid"
		if sid := a.ctx.GetString(SID); sid != "" {
			field = append(field, zap.String(SID, sid))
		}

		// cust_id
		const CUST_ID = "custID"
		if custID := a.ctx.GetString(CUST_ID); custID != "" {
			field = append(field, zap.String(CUST_ID, custID))
		}
	}

	for k, v := range a.fields {
		field = append(field, zap.String(k, v))
	}

	return field
}

func (a *AppLoggerImpl) Error(message string, refID any, serviceName string, result any) {
	log := a.zapLogger.WithOptions(zap.Fields(a.generateFields()...))
	log.Error(message, zap.Any("refID", refID), zap.Any("serviceName", serviceName), zap.Any("result", result))
}

func (a *AppLoggerImpl) Warn(message string, refID any, serviceName string, result any) {
	log := a.zapLogger.WithOptions(zap.Fields(a.generateFields()...))
	log.Warn(message, zap.Any("refID", refID), zap.Any("serviceName", serviceName), zap.Any("result", result))
}

func (a *AppLoggerImpl) Info(message string, refID any, serviceName string, result any) {
	log := a.zapLogger.WithOptions(zap.Fields(a.generateFields()...))
	log.Info(message, zap.Any("refID", refID), zap.Any("serviceName", serviceName), zap.Any("result", result))
}

func (a *AppLoggerImpl) Debug(message string, refID any, serviceName string, result any) {
	log := a.zapLogger.WithOptions(zap.Fields(a.generateFields()...))
	log.Debug(message, zap.Any("refID", refID), zap.Any("serviceName", serviceName), zap.Any("result", result))
}

func (a *AppLoggerImpl) LogRequest() {
	a.t = time.Now()
	// Check content type
	contentType := a.ctx.Request.Header.Get("Content-Type")
	isFileUpload := strings.HasPrefix(contentType, "multipart/form-data")
	var reader io.ReadCloser

	// Get raw body
	if a.ctx.Request.Body != nil && !isFileUpload {
		buf, _ := io.ReadAll(a.ctx.Request.Body)
		reader = io.NopCloser(bytes.NewBuffer(buf))
		a.ctx.Request.Body = io.NopCloser(bytes.NewBuffer(buf))

		a.requestBody = readBody(reader)
	}

	// Set body writer
	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: a.ctx.Writer}
	a.ctx.Writer = blw
	a.responseBody = blw

	if a.ctx != nil {
		log := a.zapLogger.WithOptions(zap.Fields(a.generateFields()...))
		log.Info(
			"request",
			zap.String("url", a.ctx.Request.URL.Path),
			zap.String("input", a.requestBody),
			zap.String("output", ""),
			zap.Int64("latency", 0),
			zap.Int("status", 0),
		)
	}
}

func (a *AppLoggerImpl) LogResponse() {
	if a.ctx != nil {
		// Check content type
		contentType := a.ctx.Writer.Header().Get("Content-Type")
		var body string
		if strings.HasPrefix(contentType, "application/json") {
			body = a.responseBody.body.String()
		} else {
			body = ""
		}

		log := a.zapLogger.WithOptions(zap.Fields(a.generateFields()...))
		log.Info(
			"response",
			zap.String("url", a.ctx.Request.URL.Path),
			zap.String("input", a.requestBody),
			zap.String("output", body),
			zap.Int64("latency", time.Since(a.t).Milliseconds()),
			zap.Int("status", a.ctx.Writer.Status()),
		)
	}
}

func initZapLogger(level string) *zap.Logger {
	cfg := zap.NewProductionConfig()
	lvl, _ := zap.ParseAtomicLevel(level)
	cfg.Level = zap.NewAtomicLevelAt(lvl.Level())

	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.CallerKey = zapcore.OmitKey
	cfg.EncoderConfig.FunctionKey = zapcore.OmitKey
	cfg.EncoderConfig.MessageKey = "message"

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	return zap.Must(cfg.Build())
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func readBody(reader io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	s := buf.String()
	return s
}

// New app logger
func NewAppLogger(level string) AppLogger {
	return &AppLoggerImpl{
		zapLogger: initZapLogger(level),
		fields:    make(map[string]string),
	}
}
