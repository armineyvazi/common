package log

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/armineyvazi/common/pkg/port"
)

type log struct {
	debug bool
}

func New(debug bool) port.Logger {
	customFormatter := logrus.TextFormatter{
		TimestampFormat: time.DateTime,
		FullTimestamp:   true,
	}
	logrus.SetFormatter(&customFormatter)

	return &log{
		debug: debug,
	}
}

func (l *log) Info(msg string, params ...interface{}) {
	if l.debug {
		frame := getFrame(1)
		file := strings.Split(frame.File, "/")
		logrus.Infof(file[len(file)-1]+":"+fmt.Sprintf("%d ", frame.Line)+msg, params...)
	}
}

func (l *log) Warn(msg string, params ...interface{}) {
	if l.debug {
		frame := getFrame(1)
		file := strings.Split(frame.File, "/")
		logrus.Warnf(file[len(file)-1]+":"+fmt.Sprintf("%d ", frame.Line)+msg, params...)
	}
}

func (l *log) Error(msg string, params ...interface{}) {
	if l.debug {
		frame := getFrame(1)
		file := strings.Split(frame.File, "/")
		logrus.Errorf(file[len(file)-1]+":"+fmt.Sprintf("%d ", frame.Line)+msg, params...)
	}
}

func getFrame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}
