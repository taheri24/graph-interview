package utils

import "encoding/json"

func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

func JsonEncode[T any](val T) []byte {
	return Must(json.Marshal(val))
}

func JsonDecode[T any](raw []byte) T {
	var result T
	err := json.Unmarshal(raw, &result)
	return Must(result, err)
}
