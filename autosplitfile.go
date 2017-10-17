package autosplitfile

import (
	"fmt"
	"path"
	"os"
	"strconv"
	"time"
	"strings"
	"bytes"
)

type FileOptions struct {
	PathPrefix    string
	BufferedLines int
	MaxSize       int64
	MaxTime       string
}

type record struct {
	time time.Time
	data []byte
}

type File struct {
	pathPrefix      string
	maxSize         int64
	maxTime         time.Duration
	actualFile      *os.File
	occurredError   error
	waitWriteRecord chan *record
	waitSyncRoutine chan struct{}
}

const timeLayout = "20060102-15-04"
const sequenceStart = 1

func New(options *FileOptions) (fp *File, err error) {
	maxTime, err := time.ParseDuration(options.MaxTime)
	if err != nil {
		return
	} else if maxTime < time.Minute {
		// maxTime can not less than 1 minute.
		maxTime = time.Minute
	}

	fp = &File{
		path.Clean(options.PathPrefix),
		options.MaxSize,
		maxTime,
		nil,
		nil,
		make(chan *record, options.BufferedLines),
		make(chan struct{}),
	}

	// sync routine
	go func() {
		defer func() {
			close(fp.waitSyncRoutine)
		}()

		for rec := range fp.waitWriteRecord {
			if fp.occurredError != nil {
				continue
			}

			err := fp.writeRecordToActualFile(rec)
			if err != nil {
				// todo: atomic write to the pointer
				fp.occurredError = err
			}
		}
	}()

	return
}

func (fp *File) Write(p []byte) (n int, err error) {
	// todo: atomic read the pointer.
	if fp.occurredError != nil {
		return 0, fp.occurredError
	}

	fp.waitWriteRecord <- &record{time.Now(), bytes.Repeat(p, 1)}
	return len(p), nil
}

func (fp *File) Close() error {
	close(fp.waitWriteRecord)
	<-fp.waitSyncRoutine
	return fp.actualFile.Close()
}

func (fp *File) refreshActualFile(currentRecordTime time.Time, increasingSize int) (err error) {
	var expectedFilename string

	if fp.actualFile != nil {
		expectedFilename, err = fp.expectedActualFilePath(currentRecordTime, increasingSize)
		if fp.currentActualFilePath() == expectedFilename {
			return
		} else {
			err = fp.actualFile.Close()
			if err != nil {
				return
			}

			fp.actualFile = nil
		}
	} else {
		expectedFilename = fp.buildActualFilePath(currentRecordTime, sequenceStart)
	}

	fp.actualFile, err = os.OpenFile(expectedFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	return
}

func (fp *File) currentActualFilePath() string {
	return path.Clean(fp.actualFile.Name())
}

func (fp *File) buildActualFilePath(t time.Time, seq int) string {
	// the log file name is in this format: prefix.20060102-15-04.0001 .
	// Truncate() can only deal with UTC, so add the timezone offset before Truncate() and sub it after.
	_, offsetSeconds := t.Local().Zone()
	offsetDuration := time.Second * time.Duration(offsetSeconds)
	t = t.Add(offsetDuration).Truncate(fp.maxTime).Add(-offsetDuration)
	return fmt.Sprintf("%s.%s.%04d", fp.pathPrefix, t.Format(timeLayout), seq)
}

func (fp *File) expectedActualFilePath(currentRecordTime time.Time, increasingSize int) (res string, err error) {
	// extract time and sequence number from current filename.
	filePath := fp.currentActualFilePath()
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

	if currentRecordTime.Sub(fileTime) > fp.maxTime  {
		expectedTime = currentRecordTime
		expectedSeq = sequenceStart
	} else {
		var info os.FileInfo
		info, err = fp.actualFile.Stat()
		if err != nil {
			return
		} else if info.Size() + int64(increasingSize) >= fp.maxSize {
			expectedSeq++
		}
	}

	res = fp.buildActualFilePath(expectedTime, expectedSeq)
	return
}

func (fp *File) writeRecordToActualFile(record *record) (err error) {
	err = fp.refreshActualFile(record.time, len(record.data))
	if err != nil {
		return
	}

	ret, err := fp.actualFile.Write(record.data)
	if err == nil && ret < len(record.data) {
		// unlikely, but should care.
		err = fmt.Errorf("written length less than expected")
	}

	return
}
