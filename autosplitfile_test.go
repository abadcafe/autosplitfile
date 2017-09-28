package main

import (
	"testing"
	"time"
	"fmt"
	"os"
)

func Test_AutoSplitFile(t *testing.T) {
	t.Run("new AutoSplitFile error options", func(t *testing.T) {
		_, err := NewAutoSplitFile(&AutoSplitFileOptions{
			"/tmp/abc",
			4096,
			100,
			"3ss",
		})
		if err == nil {
			t.Errorf("NewAutoSplitFile() error = %v", err)
			return
		}
	})

	var AutoSplitFile *AutoSplitFile
	t.Run("new AutoSplitFile", func(t *testing.T) {
		l, err := NewAutoSplitFile(&AutoSplitFileOptions{
			"/tmp/abc",
			4096,
			100,
			"3s",
		})
		if err != nil {
			t.Errorf("NewAutoSplitFile() error = %v", err)
			return
		}

		AutoSplitFile = l
	})

	t.Run("write", func(t *testing.T) {
		AutoSplitFile.Write([]byte("121\n"))
		time.Sleep(3 * time.Second)
	})

	t.Run("refresh file path", func(t *testing.T) {
		path1 := AutoSplitFile.currentLogFilePath()
		fmt.Printf("path1 %s\n", path1)

		AutoSplitFile.Write([]byte("llllllllllllllllllllllllllllllllllllllllllllllllll\n"))
		AutoSplitFile.Write([]byte("llllllllllllllllllllllllllllllllllllllllllllllllll\n"))
		time.Sleep(3 * time.Second)

		path2 := AutoSplitFile.currentLogFilePath()
		fmt.Printf("path2 %s\n", path2)

		AutoSplitFile.Write([]byte("row1\n"))
		time.Sleep(3 * time.Second)

		path3 := AutoSplitFile.currentLogFilePath()
		fmt.Printf("path3 %s\n", path3)
		if path2 == path3 {
			t.Errorf("path2 %s, path3 %s\n", path2, path3)
		}

		time.Sleep(61 * time.Second)

		AutoSplitFile.Write([]byte("row2\n"))
		time.Sleep(3 * time.Second)

		path4 := AutoSplitFile.currentLogFilePath()
		fmt.Printf("path4 %s\n", path4)
		if path3 == path4 {
			t.Errorf("path3 %s, path4 %s\n", path3, path4)
		}

		AutoSplitFile.Close()
	})
}
