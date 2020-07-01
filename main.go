package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/logrusorgru/aurora"
)

func main() {

	client := newClient()
	sc := bufio.NewScanner(os.Stdin)
	var wg sync.WaitGroup
	var forwarded = "0a6d8cfc90fb8ef81240cd1f127409098dd846e1.local"
	var verboseMode bool
	flag.BoolVar(&verboseMode, "v", false, "be verbose")

	flag.Parse()

	for sc.Scan() {
		rawURL := sc.Text()
		wg.Add(1)

		go func() {
			defer wg.Done()

			req, err := http.NewRequest("GET", rawURL, nil)
			req.Header.Set("X-Forwarded-Host", forwarded)
			if err != nil && verboseMode {
				fmt.Printf("[  %s  ] %s\n", aurora.Red("FAILED").String(), rawURL)
				return
			}

			resp, err := client.Do(req)
			if err != nil && verboseMode {
				fmt.Printf("[  %s  ] %s\n", aurora.Red("FAILED").String(), rawURL)
				return
			}
			defer resp.Body.Close()

			data, _ := ioutil.ReadAll(resp.Body)
			if strings.Contains(string(data), forwarded) {
				fmt.Printf("[%s] %s\n", aurora.Green("VULNERABLE").String(), rawURL)
			} else {
				if verboseMode {
					fmt.Printf("[ %s ] %s\n", aurora.Yellow("NOT VULN").String(), rawURL)
				}
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
