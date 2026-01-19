package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// LoggerConfig represents logger config
type LoggerConfig struct {
	OutputDir string `validate:"required"`
	FileName  string `validate:"required"`
}

// Logger represents Logger
type Logger struct {
	Logger *logrus.Logger `validate:"required"`
	// writer         *CustomLogWriter
	writer *LogWriter
	Config LoggerConfig
}

// NewLogger returns logger instance
func NewLogger(l *logrus.Logger, cfg LoggerConfig) (*Logger, error) {
	logger := &Logger{
		Logger: logrus.New(),
		Config: cfg,
	}

	v := validator.New()
	if err := v.Struct(l); err != nil {
		return nil, err
	}

	// if err := logger.Setup(); err != nil {
	// 	return nil, err
	// }
	return logger, nil
}

// LogFields represents log data structure
type LogFields struct {
	Time      time.Time `json:"time,omitempty"`
	Level     string    `json:"level,omitempty"`
	Message   string    `json:"msg,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
}

func (l *Logger) Format(entry *logrus.Entry) ([]byte, error) {
	var buf bytes.Buffer

	// Create a map to hold the log entry fields
	logData := make(map[string]interface{})
	logData["time"] = entry.Time.Format(time.RFC3339)
	logData["level"] = entry.Level.String()
	logData["msg"] = entry.Message

	// Include additional fields
	for key, value := range entry.Data {
		logData[key] = value

		// Special handling for errors to extract stack trace
		if key == "error" {
			switch err := value.(type) {
			case error:
				// Store the error message
				logData["error"] = err.Error()

				// Check if it's a stacktracer error from pkg/errors
				if stackTracer, ok := err.(interface {
					StackTrace() errors.StackTrace
				}); ok {
					// Convert the stack trace to a string representation
					frames := stackTracer.StackTrace()
					stackFrames := make([]string, len(frames))
					for i, f := range frames {
						stackFrames[i] = fmt.Sprintf("%+v", f)
					}
					logData["stack_trace"] = stackFrames
				}
			default:
				logData[key] = value
			}
		}
	}

	// Marshal the map into pretty JSON
	prettyJSON, err := json.MarshalIndent(logData, "", " ")
	if err != nil {
		return nil, err
	}
	buf.Write(prettyJSON)
	buf.WriteByte('\n') // Add a newline for better readability

	return buf.Bytes(), nil
}

func (l *Logger) validate() error {
	if l.Config.OutputDir == "" {
		return errors.New("output directory is mandatory")
	}
	if l.Config.FileName == "" {
		l.Config.FileName = "logs"
	}
	return nil
}

// Setup setup logger format and log locations.
func (l *Logger) Setup() error {

	if err := l.validate(); err != nil {
		return err
	}
	writer, err := NewLogWriter(l.Config.OutputDir, l.Config.FileName)
	if err != nil {
		return err
	}
	l.writer = writer
	l.Logger.SetOutput(writer)
	l.Logger.SetFormatter(l)
	return nil
}

// GetLogFilePath generates the path for the log file
func (l *Logger) GetLogFilePath(filename string, y int, m time.Month, d int) (*string, error) {
	dir := l.BuildLogDirPath(y, m, d)
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find absolute log directory")
	}
	path := filepath.Join(abs, fmt.Sprintf("%s.log", filename))
	return &path, nil
}

// BuildLogDirPath return log dir path based on the given date
func (l *Logger) BuildLogDirPath(year int, month time.Month, day int) string {
	return filepath.Join(l.Config.OutputDir, strconv.Itoa(year), fmt.Sprintf("%02d", month), fmt.Sprintf("%02d", day))
}
