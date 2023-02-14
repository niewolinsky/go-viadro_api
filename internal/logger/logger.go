package logger

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

type Level int8

const (
	LevelInfo Level = iota
	LevelError
	LevelFatal
	LevelOff
)

func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

func Log(level Level, message string, err_message error) {
	log := struct {
		Level   string `json:"level"`
		Time    string `json:"time"`
		Message string `json:"message"`
		Error   string `json:"error,omitempty"`
	}{
		Level:   level.String(),
		Time:    time.Now().UTC().Format(time.DateTime),
		Message: message,
		Error:   err_message.Error(),
	}

	line, _ := json.MarshalIndent(log, "", " ")

	fmt.Println(string(line))
}

func LogInfo(message string) {
	Log(LevelInfo, message, errors.New(""))
}

func LogError(message string, err_message error) {
	Log(LevelError, message, err_message)
}

func LogFatal(message string, err_message error) {
	Log(LevelFatal, message, err_message)
	os.Exit(1)
}
