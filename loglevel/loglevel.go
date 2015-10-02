// Copyright 2015 Kevin Bowrin All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// This package provides a way to filter log messages
// depending on their log level.
package loglevel

import (
	"log"
	"strings"
	"sync"
)

type LogLevel int

const (
	ErrorMessage LogLevel = iota
	WarnMessage
	InfoMessage
	DebugMessage
	TraceMessage
)

var logLevelToString = map[LogLevel]string{
	ErrorMessage: "ERROR",
	WarnMessage:  "WARN",
	InfoMessage:  "INFO",
	DebugMessage: "DEBUG",
	TraceMessage: "TRACE",
}

var logMessageLevel = ErrorMessage
var logMessageLevelMutex = new(sync.RWMutex)

func Set(level LogLevel) {
	logMessageLevelMutex.Lock()
	defer logMessageLevelMutex.Unlock()

	logMessageLevel = level
}

//Log a message if the level is below or equal to the set LogMessageLevel
func Log(message interface{}, messagelevel LogLevel) {
	logMessageLevelMutex.RLock()
	defer logMessageLevelMutex.RUnlock()

	if messagelevel <= logMessageLevel {
		log.Printf("%v: %v\n", messagelevel, message)
	}
}

func (level LogLevel) String() string {
	return logLevelToString[level]
}

func ParseLogLevel(parseThis string) LogLevel {
	for logLevel, logLevelString := range logLevelToString {
		if logLevelString == strings.ToUpper(parseThis) {
			return logLevel
		}

	}
	return TraceMessage
}
