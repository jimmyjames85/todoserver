package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

func logf(format string, a ...interface{}) {
	fmt.Printf("%s: %s\n", time.Now().String(), fmt.Sprintf(format, a...))
}

func ToJSON(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		fmt.Printf("failed to marshal data: %s", err)
		return fmt.Sprintf("%v", obj)
	}
	return string(b)
}

func ToBase64(data string) string {
	var out bytes.Buffer
	buf := bytes.NewBuffer([]byte(base64.StdEncoding.EncodeToString([]byte(data))))
	row := make([]byte, 72)

	for n, err := buf.Read(row); err == nil; n, err = buf.Read(row) {
		out.Write(row[:n])
		out.WriteByte('\n')
	}
	return out.String()
}
