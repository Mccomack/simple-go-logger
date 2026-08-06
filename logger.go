package logger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	defaultDirectory   = "./logs"
	defaultProjectName = "unknown"
)

type Config struct {
	Directory   string
	ProjectName string
}

type Manager struct {
	mu sync.Mutex

	file     *os.File
	filePath string
	closed   bool
}

type Logger struct {
	manager *Manager
	name    string
}

func Open(configs ...Config) (*Manager, error) {
	if len(configs) > 1 {
		return nil, errors.New("only one Config value can be provided")
	}

	config := Config{}

	if len(configs) == 1 {
		config = configs[0]
	}

	directory := strings.TrimSpace(config.Directory)
	if directory == "" {
		directory = defaultDirectory
	}

	projectName := strings.TrimSpace(config.ProjectName)
	if projectName == "" {
		projectName = defaultProjectName
	}

	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, fmt.Errorf(
			"failed to create the log directory: %w",
			err,
		)
	}

	fileName := createFileName(projectName)
	filePath := filepath.Join(directory, fileName)

	file, err := os.OpenFile(
		filePath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0o644,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to open the log file: %w",
			err,
		)
	}

	return &Manager{
		file:     file,
		filePath: filePath,
	}, nil
}

func createFileName(projectName string) string {
	timestamp := time.Now().
		UTC().
		Format("2006-01-02T15-04-05Z")

	buildType := "Release"
	if debugEnabled {
		buildType = "Debug"
	}

	return fmt.Sprintf(
		"%s-%s-%s.log",
		projectName,
		timestamp,
		buildType,
	)
}

func (m *Manager) New(name string) *Logger {
	name = strings.TrimSpace(name)

	if name == "" {
		name = defaultProjectName
	}

	return &Logger{
		manager: m,
		name:    name,
	}
}

func (m *Manager) FilePath() string {
	if m == nil {
		return ""
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	return m.filePath
}

func (m *Manager) Close() error {
	if m == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}

	m.closed = true

	if m.file == nil {
		return nil
	}

	err := m.file.Close()
	m.file = nil

	if err != nil {
		return fmt.Errorf(
			"failed to close the log file: %w",
			err,
		)
	}

	return nil
}

func currentTime() string {
	return time.Now().
		UTC().
		Format("2006-01-02T15:04:05.000000Z")
}

func (l *Logger) format(
	message string,
	logLevel string,
) string {
	message = strings.TrimRight(message, "\r\n")

	if logLevel == "" {
		return fmt.Sprintf(
			"%s [%s]: %s\n",
			currentTime(),
			l.name,
			message,
		)
	}

	return fmt.Sprintf(
		"%s %s [%s]: %s\n",
		currentTime(),
		logLevel,
		l.name,
		message,
	)
}

func (m *Manager) write(message string) error {
	if m == nil {
		return errors.New("logger Manager is not initialized")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("logger Manager is already closed")
	}

	if m.file == nil {
		return errors.New("log file is not open")
	}

	var writeErrors []error

	if err := writeString(m.file, message); err != nil {
		writeErrors = append(
			writeErrors,
			fmt.Errorf("failed to write to the log file: %w", err),
		)
	}

	if err := writeString(os.Stdout, message); err != nil {
		writeErrors = append(
			writeErrors,
			fmt.Errorf("failed to write to the console: %w", err),
		)
	}

	return errors.Join(writeErrors...)
}

func writeString(writer io.Writer, value string) error {
	written, err := io.WriteString(writer, value)
	if err != nil {
		return err
	}

	if written != len(value) {
		return io.ErrShortWrite
	}

	return nil
}

func (l *Logger) validate() error {
	if l == nil {
		return errors.New("Logger is nil")
	}

	if l.manager == nil {
		return errors.New("Logger is not initialized")
	}

	return nil
}

// logLevel is optional
func (l *Logger) Log(
	message string,
	logLevel *string,
) error {
	if err := l.validate(); err != nil {
		return err
	}

	level := ""

	if logLevel != nil {
		level = *logLevel
	}

	return l.manager.write(l.format(message, level))
}

func (l *Logger) Logf(
	formatString string,
	args ...any,
) error {
	if err := l.validate(); err != nil {
		return err
	}

	return l.Log(fmt.Sprintf(formatString, args...), nil)
}

func (l *Logger) Debug(message string) error {
	if err := l.validate(); err != nil {
		return err
	}

	if !debugEnabled {
		return nil
	}

	dbgstr := "DEBUG"

	return l.Log(message, &dbgstr)
}

func (l *Logger) Debugf(
	formatString string,
	args ...any,
) error {
	if err := l.validate(); err != nil {
		return err
	}

	if !debugEnabled {
		return nil
	}

	return l.Debug(
		fmt.Sprintf(formatString, args...),
	)
}
