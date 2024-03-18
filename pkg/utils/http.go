package utils

import (
	"fmt"
	"net/http"
	"sync"
)

// CallAPI makes an HTTP request to the given URL and sends the response or error message through the provided channel.
//
// url string, wg *sync.WaitGroup, ch chan string
// No return type
func CallAPI(url string, wg *sync.WaitGroup, ch chan string) {
	defer wg.Done()

	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		ch <- fmt.Sprintf("Error: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	// Read response body
	// For simplicity, just printing the response body here
	// You can parse and process the response as needed
	ch <- fmt.Sprintf("Response from %s: %s", url, resp.Status)
}
