package testfunc

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type Line struct {
	Level   string
	Message string
}

type MemoryLogger struct {
	lines []Line
	mu    sync.Mutex
}

func (ml *MemoryLogger) log(level string, msg string, args ...any) {
	line := strings.Builder{}
	line.WriteString(msg)
	for i, arg := range args {
		if i%2 == 0 {
			line.WriteString(fmt.Sprintf("\t%v=", arg))
		} else {
			line.WriteString(fmt.Sprintf("%v", arg))
		}
	}
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.lines = append(ml.lines, Line{
		Level:   level,
		Message: line.String(),
	})
}

func (ml *MemoryLogger) Debug(ctx context.Context, msg string, args ...any) {
	ml.log("DEBUG", msg, args...)
}

func (ml *MemoryLogger) Info(ctx context.Context, msg string, args ...any) {
	ml.log("INFO", msg, args...)
}

func (ml *MemoryLogger) Warn(ctx context.Context, msg string, args ...any) {
	ml.log("WARN", msg, args...)
}

func (ml *MemoryLogger) Error(ctx context.Context, msg string, args ...any) {
	ml.log("ERROR", msg, args...)
}

func NewMemoryLogger() *MemoryLogger {
	return &MemoryLogger{
		lines: make([]Line, 0),
	}
}
