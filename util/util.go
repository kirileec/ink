package util

import (
	"github.com/linxlib/logs"
	"sync"

	"io"
	"os"
	"time"
)

const (
	DATE_FORMAT               = "2006-01-02T15:04:05"
	DATE_FORMAT1              = "2006-01-02 15:04:05"
	DATE_FORMAT_WITH_TIMEZONE = "2006-01-02T15:04:05Z08:00"
)

// Parse date by std date string
func ParseDate(dateStr string) time.Time {
	date, err := time.ParseInLocation(DATE_FORMAT_WITH_TIMEZONE, dateStr, time.Now().Location())
	if err != nil {
		date, err = time.ParseInLocation(DATE_FORMAT, dateStr, time.Now().Location())
		if err != nil {
			date, err = time.ParseInLocation(DATE_FORMAT1, dateStr, time.Now().Location())
			if err != nil {
				logs.Error(err)
			}
		}
	}
	return date
}

// Check file if exist
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// Check file if is directory
func IsDir(path string) bool {
	file, err := os.Stat(path)
	if err != nil {
		return false
	}
	return file.IsDir()
}

// Copy folder and file
// Refer to https://www.socketloop.com/tutorials/golang-copy-directory-including-sub-directories-files
func CopyFile(source string, dest string, wg *sync.WaitGroup) {
	sourceFile, err := os.Open(source)
	defer sourceFile.Close()
	if err != nil {
		logs.Fatal(err)
	}
	destfile, err := os.Create(dest)
	if err != nil {
		logs.Fatal(err)
	}
	defer destfile.Close()
	defer wg.Done()
	_, err = io.Copy(destfile, sourceFile)
	if err == nil {
		sourceInfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceInfo.Mode())
			if err != nil {
				logs.Error(err)
			}
		}
	} else {
		logs.Error(err)
	}
}

func CopyDir(source string, dest string, wg *sync.WaitGroup) {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		logs.Error(err)
	}
	err = os.MkdirAll(dest, sourceInfo.Mode())
	if err != nil {
		logs.Fatal(err)
	}
	directory, _ := os.Open(source)
	defer directory.Close()
	defer wg.Done()
	objects, err := directory.Readdir(-1)
	for _, obj := range objects {
		sourceFilePointer := source + "/" + obj.Name()
		destinationFilePointer := dest + "/" + obj.Name()
		if obj.IsDir() {
			wg.Add(1)
			CopyDir(sourceFilePointer, destinationFilePointer, wg)
		} else {
			wg.Add(1)
			go CopyFile(sourceFilePointer, destinationFilePointer, wg)
		}
	}
}
