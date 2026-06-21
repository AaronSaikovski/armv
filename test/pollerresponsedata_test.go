package test

import (
	"bytes"
	"testing"

	"github.com/AaronSaikovski/armv/cmd/armv/poller"
)

func TestNewPollerResponseData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		respBody       []byte
		respStatusCode int
		respStatus     string
	}{
		{name: "successful 204", respBody: []byte(`{"status":"ok"}`), respStatusCode: 204, respStatus: "No Content"},
		{name: "conflict 409", respBody: []byte(`{"error":"validation failed"}`), respStatusCode: 409, respStatus: "Conflict"},
		{name: "empty body", respBody: []byte{}, respStatusCode: 204, respStatus: "No Content"},
		{name: "nil body", respBody: nil, respStatusCode: 200, respStatus: "OK"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := poller.NewPollerResponseData(tt.respBody, tt.respStatusCode, tt.respStatus)

			if !bytes.Equal(got.RespBody, tt.respBody) {
				t.Errorf("RespBody = %q, want %q", got.RespBody, tt.respBody)
			}
			if got.RespStatusCode != tt.respStatusCode {
				t.Errorf("RespStatusCode = %d, want %d", got.RespStatusCode, tt.respStatusCode)
			}
			if got.RespStatus != tt.respStatus {
				t.Errorf("RespStatus = %q, want %q", got.RespStatus, tt.respStatus)
			}
		})
	}
}
