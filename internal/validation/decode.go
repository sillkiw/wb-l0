package validation

import (
	"bytes"
	"encoding/json"
)

// DecodeStrict декодирует JSON в v и падает на неизвестных полях.
func DecodeStrict(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
