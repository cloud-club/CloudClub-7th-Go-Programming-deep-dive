// Package utils는 애플리케이션 전체에서 사용되는 유틸리티 함수들을 제공합니다.
package utils

import (
	"log"
	"os"
)

// Logger는 애플리케이션의 로깅 기능을 제공합니다
type Logger struct {
	*log.Logger
}

// NewLogger는 새로운 로거 인스턴스를 생성합니다
func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "[TinyPrometheus] ", log.LstdFlags|log.Lshortfile),
	}
}

// Error는 에러 메시지를 로깅합니다
func (l *Logger) Error(format string, v ...interface{}) {
	l.Printf("ERROR: "+format, v...)
}

// Info는 정보 메시지를 로깅합니다
func (l *Logger) Info(format string, v ...interface{}) {
	l.Printf("INFO: "+format, v...)
}

// Debug는 디버그 메시지를 로깅합니다
func (l *Logger) Debug(format string, v ...interface{}) {
	l.Printf("DEBUG: "+format, v...)
}
