package utils

import (
	"bytes"
	"encoding/json"
)

// UnmarshalDataToStruct unmarshals respBody into the value pointed to by
// targetStruct (which must be a pointer, as required by encoding/json).
func UnmarshalDataToStruct(respBody []byte, targetStruct any) error {
	return json.Unmarshal(respBody, targetStruct)
}

// MarshalStructToJSON returns the JSON encoding of targetStruct.
func MarshalStructToJSON(targetStruct any) ([]byte, error) {
	return json.Marshal(targetStruct)
}

// PrettyJsonString returns an indented version of the provided JSON string.
func PrettyJsonString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}
