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
	"strings"
	"testing"

	"github.com/AaronSaikovski/armv/cmd/armv/poller"
)

func TestPollerResponseDataFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       []byte
		statusCode int
		status     string
		wantAll    []string // substrings that must appear in output
	}{
		{
			name:       "success 204 renders success banner",
			body:       nil,
			statusCode: 204,
			status:     "No Content",
			wantAll:    []string{"SUCCESS", "No Azure Resource Validation issues", "No Content"},
		},
		{
			name:       "409 with valid JSON is pretty-printed",
			body:       []byte(`{"error":{"code":"X","message":"bad"}}`),
			statusCode: 409,
			status:     "Conflict",
			wantAll:    []string{`"error":`, `"code": "X"`, `"message": "bad"`},
		},
		{
			name:       "409 with empty body renders no-body sentinel",
			body:       []byte{},
			statusCode: 409,
			status:     "Conflict",
			wantAll:    []string{"no response body", "409", "Conflict"},
		},
		{
			name:       "409 with nil body renders no-body sentinel",
			body:       nil,
			statusCode: 409,
			status:     "Conflict",
			wantAll:    []string{"no response body", "409"},
		},
		{
			name:       "409 with non-JSON body persists verbatim",
			body:       []byte("<!doctype html><html>upstream proxy error</html>"),
			statusCode: 409,
			status:     "Conflict",
			wantAll:    []string{"Raw body:", "upstream proxy error", "409"},
		},
		{
			name:       "500 with JSON body is pretty-printed",
			body:       []byte(`{"error":"internal"}`),
			statusCode: 500,
			status:     "Internal Server Error",
			wantAll:    []string{`"error": "internal"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := poller.NewPollerResponseData(tt.body, tt.statusCode, tt.status)
			got := resp.Format()
			for _, want := range tt.wantAll {
				if !strings.Contains(got, want) {
					t.Errorf("Format() missing substring %q in:\n%s", want, got)
				}
			}
		})
	}
}

func TestPollerResponseDataFormatIsPureFunction(t *testing.T) {
	t.Parallel()

	body := []byte(`{"x":1}`)
	resp := poller.NewPollerResponseData(body, 409, "Conflict")

	a := resp.Format()
	b := resp.Format()
	if a != b {
		t.Errorf("Format() not deterministic:\na=%s\nb=%s", a, b)
	}

	// Caller's body byte slice must not be mutated.
	if string(body) != `{"x":1}` {
		t.Errorf("Format() mutated input body: %s", string(body))
	}
}
