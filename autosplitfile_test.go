package autosplitfile

import (
	"testing"
	"time"
	"fmt"
)

func Test_AutoSplitFile(t *testing.T) {
	t.Run("new File error options", func(t *testing.T) {
		_, err := New(&FileOptions{
			"/tmp/abc",
			4096,
			100,
			"3ss",
		})
		if err == nil {
			t.Errorf("New() error = %v", err)
			return
		}
	})

	var file *File
	t.Run("new File", func(t *testing.T) {
		l, err := New(&FileOptions{
			"/tmp/abc",
			4096,
			100,
			"3s",
		})
		if err != nil {
			t.Errorf("New() error = %v", err)
			return
		}

		file = l
	})

	t.Run("write", func(t *testing.T) {
		file.Write([]byte("121\n"))
		time.Sleep(3 * time.Second)
	})

	t.Run("refresh file path", func(t *testing.T) {
		path1 := file.currentActualFilePath()
		fmt.Printf("path1 %s\n", path1)

		file.Write([]byte("llllllllllllllllllllllllllllllllllllllllllllllllll\n"))
		file.Write([]byte("llllllllllllllllllllllllllllllllllllllllllllllllll\n"))
		time.Sleep(3 * time.Second)

		path2 := file.currentActualFilePath()
		fmt.Printf("path2 %s\n", path2)

		file.Write([]byte("row1\n"))
		time.Sleep(3 * time.Second)

		path3 := file.currentActualFilePath()
		fmt.Printf("path3 %s\n", path3)
		if path2 == path3 {
			t.Errorf("path2 %s, path3 %s\n", path2, path3)
		}

		time.Sleep(61 * time.Second)

		file.Write([]byte("row2\n"))
		time.Sleep(3 * time.Second)

		path4 := file.currentActualFilePath()
		fmt.Printf("path4 %s\n", path4)
		if path3 == path4 {
			t.Errorf("path3 %s, path4 %s\n", path3, path4)
		}

		file.Close()
	})
}
