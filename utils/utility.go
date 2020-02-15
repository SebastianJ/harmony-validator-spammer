package utils

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// CreateDirectory - creates a directory (mkdir -p) if it already doesn't exist
func CreateDirectory(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

// ReadFileToString - check if a file exists, proceed to read it to memory if it does
func ReadFileToString(filePath string) (string, error) {
	if FileExists(filePath) {
		data, err := ioutil.ReadFile(filePath)

		if err != nil {
			return "", err
		}

		return string(data), nil
	} else {
		return "", fmt.Errorf("file %s doesn't exist - make sure it exists or that you've specified the correct path for it", filePath)
	}
}

// FileExists - checks if a given file exists
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// ParseYaml - parses yaml into a specific type
func ParseYaml(path string, entity interface{}) error {
	yamlData, err := ReadFileToString(path)

	if err != nil {
		return err
	}

	err = yaml.Unmarshal([]byte(yamlData), entity)

	if err != nil {
		return err
	}

	return nil
}
