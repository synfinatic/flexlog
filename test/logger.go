package test

/*
 // Example usage:
 func TestMyFunction(t *testing.T) {
	// setup the test logger
 	tLogger := testlogger.NewTestLogger("DEBUG")
 	defer tLogger.Close()

	oldLogger := log.Copy()
	log = tLogger
	defer func() { log = oldLogger }()

	// call the function(s) you want to test

 	// check the log messages
 	msg := testlogger.LogMessage{}
	assert.NoError(t, tLogger.GetNext(&msg))
	assert.Equal(t, "my message", msg.Message)
	assert.Equal(t, logger.LevelDebug, msg.Level)

	// Clear any remaining log messages for the next test
	tLogger.Reset()
 }

*/

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/synfinatic/flexlog"
)

// TestLogger implements flexlog.FlexLogger
type TestLogger struct {
	*flexlog.Logger
	r        *io.PipeReader
	w        *io.PipeWriter
	errors   chan error
	messages chan LogMessage
	close    bool
}

// LogMessage is a struct to hold the decoded log message
type LogMessage struct {
	LevelStr string `json:"level"`
	Level    slog.Level
	Message  string     `json:"msg"`
	Time     string     `json:"time"`
	Source   FileSource `json:"source"`
	Error    string     `json:"error"` // i standardized on this field name
	// XXX: Problem is that all the extra fields are not being decoded
}

// FileSource is a struct to hold the source file information
type FileSource struct {
	File     string `json:"file"`
	Function string `json:"function"`
	Line     int    `json:"line"`
}

// NewTestLogger creates a new TestLogger
func NewTestLogger(level string) *TestLogger {
	reader, writer := io.Pipe()

	l := flexlog.LevelStrings[strings.ToUpper(level)].Level()

	tl := TestLogger{
		Logger:   flexlog.NewLogger(flexlog.NewJSON, writer, false, l, false),
		w:        writer,
		r:        reader,
		errors:   make(chan error, 10),
		messages: make(chan LogMessage, 10),
		close:    false,
	}

	// start a goroutine to read from the pipe and decode the log messages
	go func() {
		i := 0
		r := bufio.NewReader(tl.r)

		for !tl.close {
			line, err := r.ReadString('\n')
			i++
			if err != nil {
				if err == io.EOF {
					break
				} else {
					tl.errors <- err
					break
				}
			}
			msg := LogMessage{}
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				tl.errors <- fmt.Errorf("unable to decode log message: %s", err.Error())
				break
			}
			msg.Level = flexlog.LevelStrings[msg.LevelStr]

			tl.messages <- msg
		}
	}()
	return &tl
}

// Close the TestLogger.  Should be called via defer after creating the logger.
func (tl *TestLogger) Close() {
	tl.close = true
	tl.w.Close()
	tl.r.Close()
}

// Reset clears all the previously generated log messages from the buffer.
func (tl *TestLogger) Reset() {
	if tl.close {
		return
	}
	r := bufio.NewReader(tl.r)
	for r.Buffered() > 0 {
		_, _ = r.ReadString('\n')
	}
}

// GetNext returns the next available logged message or retruns an error if no messages are available.
func (tl *TestLogger) GetNext(msg *LogMessage) error {
	if tl.close {
		return errors.New("logger closed")
	}

	t := time.NewTicker(100 * time.Millisecond)
	select {
	case err := <-tl.errors:
		return err
	case m := <-tl.messages:
		*msg = m
		return nil
	case <-t.C:
		return errors.New("no log messages found")
	}
}

// GetNextLevel scans the generated log messages and returns the first message of the given level.
// Note that this will consume any messages that are not of the given level.
func (tl *TestLogger) GetNextLevel(level slog.Level, msg *LogMessage) error {
	for {
		err := tl.GetNext(msg)
		if err != nil {
			return err
		}

		if msg.Level == level {
			return nil
		}
	}
}

/*
// CheckNextLevel checks the next logged message of the given level and returns true if the message matches the given message.
func (tl *TestLogger) CheckNextLevelEqual(level slog.Level, field, value string) (bool, error) {
	msg := LogMessage{}
	err := tl.GetNextLevel(level, &msg)
	if err != nil {
		return false, err
	}

	if reflect.ValueOf(msg).FieldByName(field).String() == value {
		return true, nil
	}
	return false, nil
}

func (tl *TestLogger) CheckNextLevelMatch(level slog.Level, field string, match *regexp.Regexp) (bool, error) {
	msg := LogMessage{}
	err := tl.GetNextLevel(level, &msg)
	if err != nil {
		return false, err
	}

	value := reflect.ValueOf(msg).FieldByName(field).String()
	if match.MatchString(value) {
		return true, nil
	}
	return false, nil
}

func (tl *TestLogger) CheckNextEqual(field, value string) (bool, error) {
	msg := LogMessage{}
	err := tl.GetNext(&msg)
	if err != nil {
		return false, err
	}

	if reflect.ValueOf(msg).FieldByName(field).String() == value {
		return true, nil
	}
	return false, nil
}

func (tl *TestLogger) CheckNextMatch(field string, match *regexp.Regexp) (bool, error) {
	msg := LogMessage{}
	err := tl.GetNext(&msg)
	if err != nil {
		return false, err
	}

	value := reflect.ValueOf(msg).FieldByName(field).String()
	if match.MatchString(value) {
		return true, nil
	}
	return false, nil
}
*/

// Log a message at the Fatal level without exiting
func (tl *TestLogger) Fatal(msg string, args ...interface{}) {
	ctx := context.Background()
	tl.Logger.LogWithSource(ctx, flexlog.LevelFatal, flexlog.StackFrames, msg, args...)
}

// Log a message at the Trace level with context without exting
func (tl *TestLogger) FatalContext(ctx context.Context, msg string, args ...interface{}) {
	tl.Logger.LogWithSource(ctx, flexlog.LevelFatal, flexlog.StackFrames, msg, args...)
}
