package yaml

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYamlMarshaling(t *testing.T) {
	type testType struct {
		L []int          `yaml:"list"`
		M map[string]any `yaml:"map"`
	}

	obj := &testType{
		L: []int{1, 2, 3},
		M: map[string]any{
			"k": "v",
			"n": 100,
		},
	}

	filename := "test.file"
	path := filepath.Join(t.TempDir(), filename)

	fmt.Println(path)

	err := WriteYaml(path, obj)
	require.NoError(t, err)

	actualObj, err := ReadYaml[testType](path)
	require.NoError(t, err)

	require.Equal(t, obj, actualObj)
}
