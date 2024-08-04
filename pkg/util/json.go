package util

import "encoding/json"

func ToJSON(v any) string {
	return string(ToJSONByte(v))
}

func ToJSONByte(v any) []byte {
	if b, err := json.Marshal(v); err == nil {
		return b
	}
	return nil
}
