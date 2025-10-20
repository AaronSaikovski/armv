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

package poller

import (
	"testing"
	"time"
)

func TestConstants(t *testing.T) {
	t.Run("progress bar max should be positive", func(t *testing.T) {
		if progressBarMax <= 0 {
			t.Errorf("progressBarMax = %d, want positive value", progressBarMax)
		}
	})

	t.Run("sleep duration should be reasonable", func(t *testing.T) {
		if sleepDuration < 0 {
			t.Errorf("sleepDuration = %v, want non-negative value", sleepDuration)
		}
		if sleepDuration > 10*time.Second {
			t.Errorf("sleepDuration = %v, seems too long for polling", sleepDuration)
		}
	})

	t.Run("polling timeout should be reasonable", func(t *testing.T) {
		if pollingTimeout <= 0 {
			t.Errorf("pollingTimeout = %v, want positive value", pollingTimeout)
		}
		if pollingTimeout < 1*time.Minute {
			t.Errorf("pollingTimeout = %v, seems too short for Azure operations", pollingTimeout)
		}
	})

	t.Run("API constants should have expected values", func(t *testing.T) {
		if API_SUCCESS != 202 {
			t.Errorf("API_SUCCESS = %d, want 202", API_SUCCESS)
		}
		if API_RESOURCE_MOVE_OK != 204 {
			t.Errorf("API_RESOURCE_MOVE_OK = %d, want 204", API_RESOURCE_MOVE_OK)
		}
		if API_RESOURCE_MOVE_FAIL != 409 {
			t.Errorf("API_RESOURCE_MOVE_FAIL = %d, want 409", API_RESOURCE_MOVE_FAIL)
		}
	})
}
