package yaml

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func ReadYaml[T any](filename string) (*T, error) {
	item := new(T)

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, item)
	if err != nil {
		return nil, fmt.Errorf("invalid config, err=%w", err)
	}

	return item, nil
}

func WriteYaml(path string, data any) error {
	bin, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		err = os.Mkdir(filepath.Dir(path), 0700)
		if err != nil && !os.IsExist(err) {
			return err
		}
	}

	return os.WriteFile(path, bin, 0600)
}
