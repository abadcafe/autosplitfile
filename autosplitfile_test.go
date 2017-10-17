package autosplitfile

import (
	"testing"
	"time"
	"fmt"
)

func TestAutoSplitFile(t *testing.T) {
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

	t.Run("refresh file path", func(t *testing.T) {
		file.Write([]byte("121\n"))
		time.Sleep(3 * time.Second)

		path1 := file.currentActualFilePath()
		fmt.Printf("path1 %s, original\n", path1)

		file.Write([]byte("llllllllllllllllllllllllllllllllllllllllllllllllll\n"))
		file.Write([]byte("llllllllllllllllllllllllllllllllllllllllllllllllll\n"))
		time.Sleep(3 * time.Second)

		path2 := file.currentActualFilePath()
		fmt.Printf("path2 %s, sequence increased\n", path2)
		if path2 == path1 {
			t.Errorf("path2 %s should not equal path1 %s\n", path2, path1)
		}

		file.Write([]byte("row1\n"))
		time.Sleep(3 * time.Second)

		path3 := file.currentActualFilePath()
		fmt.Printf("path3 %s should equals path2, but when the time spans two minutes, it will fail.\n", path3)
		if path2 != path3 {
			t.Errorf("path3 %s should equals path2 %s\n", path3, path2)
		}

		time.Sleep(61 * time.Second)

		file.Write([]byte("row2\n"))
		time.Sleep(3 * time.Second)

		path4 := file.currentActualFilePath()
		fmt.Printf("path4 %s, time increased\n", path4)
		if path3 == path4 {
			fmt.Errorf("path4 %s should not equal path3 %s\n", path4, path3)
		}

		file.Close()
	})
}
