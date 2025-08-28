package ollama

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalJSON(t *testing.T) {
	data, err := json.Marshal(instance)
	t.Run("MarshalJSON does not return an error", func(t *testing.T) {
		assert.Nil(t, err)
	})
	t.Run("MarshalJSON returns expected JSON", func(t *testing.T) {
		assert.JSONEq(t, `{"description":"Ollama local inference provider","local":true,"models":null,"name":"ollama","public":false,"reason":""}`, string(data))
	})
}
