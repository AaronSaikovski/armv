package test

import (
	"encoding/json"
	"testing"

	"github.com/AaronSaikovski/armv/pkg/utils"
)

type testStruct struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

func TestUnmarshalDataToStruct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   []byte
		want    testStruct
		wantErr bool
	}{
		{
			name:  "valid JSON",
			input: []byte(`{"name":"John Doe","age":30,"email":"john@example.com"}`),
			want:  testStruct{Name: "John Doe", Age: 30, Email: "john@example.com"},
		},
		{
			name:    "invalid JSON",
			input:   []byte(`{"name":"John Doe","age":30,`),
			wantErr: true,
		},
		{
			name:  "empty JSON",
			input: []byte(`{}`),
		},
		{
			name:  "partial JSON",
			input: []byte(`{"name":"Jane"}`),
			want:  testStruct{Name: "Jane"},
		},
		{
			name:    "empty input",
			input:   []byte(``),
			wantErr: true,
		},
		{
			name:  "extra fields silently ignored",
			input: []byte(`{"name":"K","age":1,"email":"a@b","extra":"ignored"}`),
			want:  testStruct{Name: "K", Age: 1, Email: "a@b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got testStruct
			err := utils.UnmarshalDataToStruct(tt.input, &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UnmarshalDataToStruct error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("UnmarshalDataToStruct = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestUnmarshalDataToStructTargetIsActuallyMutated pins the fix for the
// previous double-indirection bug (`json.Unmarshal(b, &targetStruct)` which
// unmarshalled into the interface header, not the caller's value).
func TestUnmarshalDataToStructTargetIsActuallyMutated(t *testing.T) {
	t.Parallel()

	var got testStruct
	if err := utils.UnmarshalDataToStruct([]byte(`{"name":"X","age":7,"email":"x@y"}`), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Name != "X" || got.Age != 7 || got.Email != "x@y" {
		t.Errorf("target struct not populated: %+v", got)
	}
}

func TestMarshalStructToJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name:  "valid struct",
			input: testStruct{Name: "Alice", Age: 25, Email: "alice@example.com"},
			want:  `{"name":"Alice","age":25,"email":"alice@example.com"}`,
		},
		{
			name:  "empty struct",
			input: testStruct{},
			want:  `{"name":"","age":0,"email":""}`,
		},
		{
			name:  "simple map",
			input: map[string]string{"key": "value"},
			want:  `{"key":"value"}`,
		},
		{
			name:  "nil input",
			input: nil,
			want:  `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := utils.MarshalStructToJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("MarshalStructToJSON error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && string(got) != tt.want {
				t.Errorf("MarshalStructToJSON = %s, want %s", string(got), tt.want)
			}
		})
	}
}

func TestPrettyJsonString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "compact JSON", input: `{"name":"Bob","age":35}`},
		{name: "already pretty JSON", input: "{\n    \"name\": \"Charlie\"\n}"},
		{name: "empty object", input: `{}`},
		{name: "nested JSON", input: `{"user":{"name":"David","address":{"city":"NYC"}}}`},
		{name: "invalid JSON", input: `{"name":"Invalid",`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := utils.PrettyJsonString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("PrettyJsonString error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			// Output must be valid JSON.
			var v any
			if err := json.Unmarshal([]byte(got), &v); err != nil {
				t.Errorf("output is not valid JSON: %v (output: %q)", err, got)
			}
		})
	}
}
