package logger

import (
	"io"
	"net/http"
	"strings"
	"time"
)

type Logger struct {
	writer io.Writer
}

func New(writer io.Writer) *Logger {
	return &Logger{writer: writer}
}

func (l *Logger) LogRequest(req *http.Request) {
	var b strings.Builder

	b.WriteString("\n=== Request ===\n")
	b.WriteString(time.Now().Format(time.RFC3339) + "\n")
	b.WriteString(req.Method + " " + req.URL.String() + "\n")
	b.WriteString("Host: " + req.Host + "\n")

	for k, v := range req.Header {
		b.WriteString(k + ": " + strings.Join(v, ", ") + "\n")
	}

	l.writer.Write([]byte(b.String()))
}

func (l *Logger) LogResponse(res *http.Response, duration time.Duration) {
	var b strings.Builder

	b.WriteString("\n=== Response ===\n")
	b.WriteString(time.Now().Format(time.RFC3339) + "\n")
	b.WriteString("Duration: " + duration.String() + "\n")
	b.WriteString("Status: " + res.Status + "\n")

	for k, v := range res.Header {
		b.WriteString(k + ": " + strings.Join(v, ", ") + "\n")
	}

	l.writer.Write([]byte(b.String()))
}
