package utils

import (
	"encoding/json"
)

func JsonEncode(s interface{}) string {
	jBytes, _ := json.Marshal(s)
	return string(jBytes)
}
