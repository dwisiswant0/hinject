package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/logrusorgru/aurora"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {

	client := newClient()
	var wg sync.WaitGroup
	sc := bufio.NewScanner(os.Stdin)
	var forwarded = "nobody.dw1.io"

	for sc.Scan() {
		rawURL := sc.Text()
		wg.Add(1)

		go func() {
			defer wg.Done()

			req, err := http.NewRequest("GET", rawURL, nil)
			req.Header.Set("X-Forwarded-Host", forwarded)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to create request: %s\n", err)
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "request failed: %s\n", err)
				return
			}
			defer resp.Body.Close()

			data, _ := ioutil.ReadAll(resp.Body)
			if strings.Contains(string(data), forwarded) {
				fmt.Printf("[%s] %s\n", aurora.Green("VULN").String(), rawURL)
			}
		}()
	}

	wg.Wait()
}

func newClient() *http.Client {

	tr := &http.Transport{
		MaxIdleConns:    30,
		IdleConnTimeout: time.Second,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   time.Second * 10,
			KeepAlive: time.Second,
		}).DialContext,
	}

	re := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &http.Client{
		Transport:     tr,
		CheckRedirect: re,
		Timeout:       time.Second * 10,
	}

}
