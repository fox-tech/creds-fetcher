package fsmanager

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type defaultFileSystemManager struct {
}

func NewDefault() defaultFileSystemManager {
	return defaultFileSystemManager{}
}

var (
	ErrCouldNotReadFile  = errors.New("read from file failed")
	ErrCouldNotWriteFile = errors.New("write file failed")
)

// readFile tries to read the given filename, if the file doesn't exist,
// it is created
func (defaultFileSystemManager) ReadFile(dir, filename string) ([]byte, error) {
	fp := filepath.Join(dir, filename)
	data := []byte{}

	hd, err := os.UserHomeDir()
	if err != nil {
		return data, fmt.Errorf("%w: failed to get home dir: %v", ErrCouldNotReadFile, err)
	}

	if err = os.Chdir(hd); err != nil {
		return data, fmt.Errorf("%w: failed to change dir: %v", ErrCouldNotReadFile, err)
	}

	data, err = ioutil.ReadFile(fp)
	if err == nil {
		return data, nil
	}

	if err != nil && !os.IsNotExist(err) {
		return data, fmt.Errorf("%w: failed to read file %s: %v", ErrCouldNotReadFile, fp, err)
	}

	_, err = os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return data, fmt.Errorf("%w: failed to open dir: %v", ErrCouldNotReadFile, err)
		}

		log.Print("credentials directory not found, creating...")
		if err = os.Mkdir(dir, 0766); err != nil {
			return data, fmt.Errorf("%w: failed to create dir: %v", ErrCouldNotWriteFile, err)
		}
		log.Print("credentials directory created")
	}

	log.Print("credentials file not found, creating...")
	_, err = os.Create(fp)
	if err != nil {
		return data, fmt.Errorf("%w: failed to create file: %v", ErrCouldNotWriteFile, err)
	}
	log.Printf("credentials file created: %s", fp)

	return data, nil
}

func (defaultFileSystemManager) WriteFile(name string, data []byte) error {
	return os.WriteFile(name, data, fs.FileMode(os.O_RDWR))
}
