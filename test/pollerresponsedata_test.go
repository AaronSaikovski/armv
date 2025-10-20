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
	"testing"

	"github.com/AaronSaikovski/armv/cmd/armv/poller"
)

func TestNewPollerResponseData(t *testing.T) {
	tests := []struct {
		name           string
		respBody       []byte
		respStatusCode int
		respStatus     string
	}{
		{
			name:           "successful response",
			respBody:       []byte(`{"status":"ok"}`),
			respStatusCode: 204,
			respStatus:     "No Content",
		},
		{
			name:           "error response",
			respBody:       []byte(`{"error":"validation failed"}`),
			respStatusCode: 409,
			respStatus:     "Conflict",
		},
		{
			name:           "empty body",
			respBody:       []byte{},
			respStatusCode: 204,
			respStatus:     "No Content",
		},
		{
			name:           "nil body",
			respBody:       nil,
			respStatusCode: 200,
			respStatus:     "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := poller.NewPollerResponseData(tt.respBody, tt.respStatusCode, tt.respStatus)

			if string(got.RespBody) != string(tt.respBody) {
				t.Errorf("NewPollerResponseData() RespBody = %v, want %v", got.RespBody, tt.respBody)
			}
			if got.RespStatusCode != tt.respStatusCode {
				t.Errorf("NewPollerResponseData() RespStatusCode = %v, want %v", got.RespStatusCode, tt.respStatusCode)
			}
			if got.RespStatus != tt.respStatus {
				t.Errorf("NewPollerResponseData() RespStatus = %v, want %v", got.RespStatus, tt.respStatus)
			}
		})
	}
}
