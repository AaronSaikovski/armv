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

	"github.com/AaronSaikovski/armv/pkg/utils"
)

func TestCheckValidSubscriptionID(t *testing.T) {
	tests := []struct {
		name           string
		subscriptionID string
		want           bool
	}{
		{
			name:           "valid subscription ID",
			subscriptionID: "12345678-1234-1234-1234-123456789012",
			want:           true,
		},
		{
			name:           "valid subscription ID with lowercase",
			subscriptionID: "abcdef12-abcd-abcd-abcd-123456789abc",
			want:           true,
		},
		{
			name:           "valid subscription ID with uppercase",
			subscriptionID: "ABCDEF12-ABCD-ABCD-ABCD-123456789ABC",
			want:           true,
		},
		{
			name:           "invalid subscription ID - too short",
			subscriptionID: "12345678-1234-1234-1234",
			want:           false,
		},
		{
			name:           "invalid subscription ID - missing hyphens",
			subscriptionID: "12345678123412341234123456789012",
			want:           false,
		},
		{
			name:           "invalid subscription ID - empty string",
			subscriptionID: "",
			want:           false,
		},
		{
			name:           "invalid subscription ID - wrong format",
			subscriptionID: "not-a-valid-uuid",
			want:           false,
		},
		{
			name:           "invalid subscription ID - extra characters",
			subscriptionID: "12345678-1234-1234-1234-123456789012-extra",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.CheckValidSubscriptionID(tt.subscriptionID)
			if got != tt.want {
				t.Errorf("CheckValidSubscriptionID(%q) = %v, want %v", tt.subscriptionID, got, tt.want)
			}
		})
	}
}
