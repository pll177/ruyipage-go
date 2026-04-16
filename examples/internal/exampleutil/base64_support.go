package exampleutil

import "encoding/base64"

func ruyiBase64Decode(value string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(value)
}
