package test

import (
	"testing"

	"github.com/AaronSaikovski/armv/pkg/utils"
)

func TestCheckValidSubscriptionID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		subscriptionID string
		want           bool
	}{
		{name: "valid lowercase digits", subscriptionID: "12345678-1234-1234-1234-123456789012", want: true},
		{name: "valid lowercase hex", subscriptionID: "abcdef12-abcd-abcd-abcd-123456789abc", want: true},
		{name: "valid uppercase hex", subscriptionID: "ABCDEF12-ABCD-ABCD-ABCD-123456789ABC", want: true},
		{name: "valid mixed case hex", subscriptionID: "AbCdEf12-1234-ABCD-abcd-123456789AbC", want: true},
		{name: "too short", subscriptionID: "12345678-1234-1234-1234", want: false},
		{name: "missing hyphens", subscriptionID: "12345678123412341234123456789012", want: false},
		{name: "empty", subscriptionID: "", want: false},
		{name: "non-hex characters", subscriptionID: "not-a-valid-uuid", want: false},
		{name: "trailing garbage", subscriptionID: "12345678-1234-1234-1234-123456789012-extra", want: false},
		{name: "leading garbage", subscriptionID: "prefix-12345678-1234-1234-1234-123456789012", want: false},
		// Braces are no longer accepted — Azure subscription IDs are bare UUIDs.
		{name: "braced UUID rejected", subscriptionID: "{12345678-1234-1234-1234-123456789012}", want: false},
		{name: "unbalanced open brace rejected", subscriptionID: "{12345678-1234-1234-1234-123456789012", want: false},
		{name: "unbalanced close brace rejected", subscriptionID: "12345678-1234-1234-1234-123456789012}", want: false},
		// Anchors must reject surrounding/injected whitespace and newlines.
		{name: "leading whitespace rejected", subscriptionID: " 12345678-1234-1234-1234-123456789012", want: false},
		{name: "trailing whitespace rejected", subscriptionID: "12345678-1234-1234-1234-123456789012 ", want: false},
		{name: "trailing newline rejected", subscriptionID: "12345678-1234-1234-1234-123456789012\n", want: false},
		{name: "embedded newline injection rejected", subscriptionID: "12345678-1234-1234-1234-123456789012\nevil", want: false},
		{name: "non-hex g in first group", subscriptionID: "g2345678-1234-1234-1234-123456789012", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := utils.CheckValidSubscriptionID(tt.subscriptionID); got != tt.want {
				t.Errorf("CheckValidSubscriptionID(%q) = %v, want %v", tt.subscriptionID, got, tt.want)
			}
		})
	}
}
