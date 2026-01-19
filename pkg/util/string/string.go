package util_string

import "encoding/json"

func SafeMarshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
