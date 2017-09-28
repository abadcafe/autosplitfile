package main

import (
	"fmt"
	"path"
	"os"
	"strconv"
	"time"
	"strings"
)

type AutoSplitFileOptions struct {
	pathPrefix    string
	bufferedLines int
	maxSize       int
	maxTime       string
}

type AutoSplitFile struct {
	pathPrefix      string
	maxSize         int
	maxTime         time.Duration
	actualFile      *os.File
	occurredError   error
	waitFileLine    chan []byte
	waitSyncRoutine chan struct{}
}

const timeLayout = "20060102-15-04"
const sequenceStart = 1

func NewAutoSplitFile(options *AutoSplitFileOptions) (fp *AutoSplitFile, err error) {
	maxTime, err := time.ParseDuration(options.maxTime)
	if err != nil {
		return
	} else if maxTime < time.Minute {
		// maxTime can not less than 1 minute.
		maxTime = time.Minute
	}

	fp = &AutoSplitFile{
		path.Clean(options.pathPrefix),
		options.maxSize,
		maxTime,
		nil,
		nil,
		make(chan []byte, options.bufferedLines),
		make(chan struct{}),
	}

	// sync routine
	go func() {
		defer func() {
			close(fp.waitSyncRoutine)
		}()

		for line := range fp.waitFileLine {
			if fp.occurredError != nil {
				continue
			}

			err := fp.writeFileLine(line)
			if err != nil {
				fp.occurredError = err
			}
		}
	}()

	return
}

func (fp *AutoSplitFile) Write(p []byte) (n int, err error) {
	if fp.occurredError != nil {
		return 0, fp.occurredError
	}

	fp.waitFileLine <- p
	return len(p), nil
}

func (fp *AutoSplitFile) Close() error {
	close(fp.waitFileLine)
	<-fp.waitSyncRoutine
	return fp.actualFile.Close()
}

func (fp *AutoSplitFile) refreshLogFile() (err error) {
	var expectedFilename string

	if fp.actualFile != nil {
		expectedFilename, err = fp.expectedLogFilePath()
		if fp.currentLogFilePath() == expectedFilename {
			return
		} else {
			err = fp.actualFile.Close()
			if err != nil {
				return
			}

			fp.actualFile = nil
		}
	} else {
		expectedFilename = fp.buildLogFilePath(time.Now(), sequenceStart)
	}

	fp.actualFile, err = os.OpenFile(expectedFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	return
}

func (fp *AutoSplitFile) currentLogFilePath() string {
	return path.Clean(fp.actualFile.Name())
}

func (fp *AutoSplitFile) buildLogFilePath(time time.Time, seq int) string {
	// the log file name is in this format: prefix.20060102-15-04.0001 .
	return fmt.Sprintf("%s.%s.%04d", fp.pathPrefix, time.Local().Format(timeLayout), seq)
}

func (fp *AutoSplitFile) expectedLogFilePath() (res string, err error) {
	// extract time and sequence number from current filename.
	filePath := fp.currentLogFilePath()
	suffix := strings.TrimPrefix(filePath, fp.pathPrefix+".")
	parts := strings.Split(suffix, ".")
	timePart := parts[0]
	seqPart := parts[1]

	fileTime, err := time.ParseInLocation(timeLayout, timePart, time.Local)
	if err != nil {
		return
	}

	fileSeq, err := strconv.Atoi(seqPart)
	if err != nil {
		return
	}

	// make expected filename.
	expectedTime := fileTime
	expectedSeq := fileSeq

	now := time.Now()
	if fp.maxTime < now.Sub(fileTime) {
		expectedTime = now
		expectedSeq = sequenceStart
	} else {
		var info os.FileInfo
		info, err = fp.actualFile.Stat()
		if err != nil {
			return
		} else if info.Size() >= int64(fp.maxSize) {
			expectedSeq++
		}
	}

	res = fp.buildLogFilePath(expectedTime, expectedSeq)
	return
}

func (fp *AutoSplitFile) writeFileLine(p []byte) (err error) {
	err = fp.refreshLogFile()
	if err != nil {
		return
	}

	ret, err := fp.actualFile.Write(p)
	if err == nil && ret < len(p) {
		// unlikely, but should care.
		err = fmt.Errorf("written length less than expected")
	}

	return
}
