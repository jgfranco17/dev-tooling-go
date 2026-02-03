package load

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPerson struct {
	Name string `json:"name" yaml:"name"`
	Age  int    `json:"age" yaml:"age"`
}

func TestFromJSON_Success(t *testing.T) {
	data := `{"name":"Ada","age":42}`
	got, err := FromJSON[testPerson](strings.NewReader(data))

	require.NoError(t, err)
	assert.Equal(t, testPerson{Name: "Ada", Age: 42}, got)
}

func TestFromJSON_Invalid(t *testing.T) {
	invalidData := `{not valid}`
	got, err := FromJSON[testPerson](strings.NewReader(invalidData))

	require.Error(t, err)
	assert.Equal(t, testPerson{}, got)
}

func TestFromYAML_Success(t *testing.T) {
	data := "name: Ada\nage: 42\n"
	got, err := FromYAML[testPerson](strings.NewReader(data))

	require.NoError(t, err)
	assert.Equal(t, testPerson{Name: "Ada", Age: 42}, got)
}

func TestFromYAML_Invalid(t *testing.T) {
	invalidData := ": invalid"
	got, err := FromYAML[testPerson](strings.NewReader(invalidData))

	require.Error(t, err)
	assert.Equal(t, testPerson{}, got)
}
