package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var configStruct interface{}
var configPaths = []string{}

func AddConfigPath(path string) {
	configPaths = append(configPaths, path)
}

func SetConfigStruct(s interface{}) {
	configStruct = s
}

func Parse() error {
	for _, path := range configPaths {
		_, err := os.Stat(path)

		if err != nil {
			continue
		}

		b, err := ioutil.ReadFile(path)

		if err != nil {
			return fmt.Errorf("Error reading config file '%v': %v", path, err.Error())
		}

		if err := json.Unmarshal(b, &configStruct); err != nil {
			return fmt.Errorf("Error unmarshalling config file '%v': %v", path, err.Error())
		}

		return nil
	}

	return errors.New(fmt.Sprintf(
		"No configurations found in paths: \r\n %v",
		strings.Join(configPaths, "\n"),
	))
}
