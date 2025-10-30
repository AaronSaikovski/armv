/*
MIT License

# Copyright (c) 2024 Aaron Saikovski

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

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
	tests := []struct {
		name    string
		input   []byte
		want    testStruct
		wantErr bool
	}{
		{
			name:  "valid JSON",
			input: []byte(`{"name":"John Doe","age":30,"email":"john@example.com"}`),
			want: testStruct{
				Name:  "John Doe",
				Age:   30,
				Email: "john@example.com",
			},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   []byte(`{"name":"John Doe","age":30,`),
			want:    testStruct{},
			wantErr: true,
		},
		{
			name:    "empty JSON",
			input:   []byte(`{}`),
			want:    testStruct{},
			wantErr: false,
		},
		{
			name:  "partial JSON",
			input: []byte(`{"name":"Jane"}`),
			want: testStruct{
				Name: "Jane",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got testStruct
			err := utils.UnmarshalDataToStruct(tt.input, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalDataToStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("UnmarshalDataToStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshalStructToJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		{
			name: "valid struct",
			input: testStruct{
				Name:  "Alice",
				Age:   25,
				Email: "alice@example.com",
			},
			want:    `{"name":"Alice","age":25,"email":"alice@example.com"}`,
			wantErr: false,
		},
		{
			name:    "empty struct",
			input:   testStruct{},
			want:    `{"name":"","age":0,"email":""}`,
			wantErr: false,
		},
		{
			name:    "simple map",
			input:   map[string]string{"key": "value"},
			want:    `{"key":"value"}`,
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			want:    `null`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.MarshalStructToJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalStructToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(got) != tt.want {
				t.Errorf("MarshalStructToJSON() = %s, want %s", string(got), tt.want)
			}
		})
	}
}

func TestPrettyJsonString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(string) bool
	}{
		{
			name:    "compact JSON",
			input:   `{"name":"Bob","age":35,"email":"bob@example.com"}`,
			wantErr: false,
			check: func(output string) bool {
				// Check if output is properly indented
				var data interface{}
				err := json.Unmarshal([]byte(output), &data)
				return err == nil && len(output) > len(`{"name":"Bob","age":35,"email":"bob@example.com"}`)
			},
		},
		{
			name:    "already pretty JSON",
			input:   "{\n    \"name\": \"Charlie\"\n}",
			wantErr: false,
			check: func(output string) bool {
				var data interface{}
				return json.Unmarshal([]byte(output), &data) == nil
			},
		},
		{
			name:    "invalid JSON",
			input:   `{"name":"Invalid",`,
			wantErr: true,
			check:   func(output string) bool { return true },
		},
		{
			name:    "empty object",
			input:   `{}`,
			wantErr: false,
			check: func(output string) bool {
				return len(output) > 0
			},
		},
		{
			name:    "nested JSON",
			input:   `{"user":{"name":"David","address":{"city":"NYC","zip":"10001"}}}`,
			wantErr: false,
			check: func(output string) bool {
				// Should contain newlines and indentation
				var data interface{}
				err := json.Unmarshal([]byte(output), &data)
				return err == nil && len(output) > 50
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.PrettyJsonString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrettyJsonString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.check(got) {
				t.Errorf("PrettyJsonString() output validation failed for input: %s", tt.input)
			}
		})
	}
}
