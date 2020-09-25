package main

import (
	"flag"
	"net/http"
	"crypto/tls"
	"math/rand"
	"fmt"
	"time"
	"bufio"
	"os"
	"sync"
)

var userAgent = []string {
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.157 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36",
	"Mozilla/5.0 (Windows NT 5.1; rv:7.0.1) Gecko/20100101 Firefox/7.0.1",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:54.0) Gecko/20100101 Firefox/54.0",
}

type addedHeader struct {
	name string
	value string
}

var headers []addedHeader

func getRandomUserAgent() string {
	rand.Seed(time.Now().Unix())
	return userAgent[rand.Intn(len(userAgent))]
}

func getStatusCode(url string, header *addedHeader, timeout int) (int, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	var client = &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if nil != err {
		return -1, err
	}

	req.Header.Add("User-Agent", getRandomUserAgent())
	req.Header.Add(header.name, header.value)

	resp, err := client.Do(req)
	if nil != err {
		return -1, err
	}

	return resp.StatusCode, err
}

func worker(url string, header addedHeader, timeout int, wg *sync.WaitGroup) {
	defer wg.Done()

	statusCode, err := getStatusCode(url, &header, timeout)

	if nil != err {
		fmt.Fprintln(os.Stderr, err)
	}

	if -1 != statusCode {
		fmt.Println(url, header.name, statusCode)
	}
}

func initHeaders() {
	headers = []addedHeader {
		addedHeader {name: "Forwarded-For-Ip",value: "127.0.0.1",},
		addedHeader {name: "X-Forwarded-For",value: "127.0.0.1",},
		addedHeader {name: "Forwarded-For",value: "127.0.0.1",},
		addedHeader {name: "Forwarded",value: "127.0.0.1",},
		addedHeader {name: "X-Forwarded-For-Original",value: "127.0.0.1",},
		addedHeader {name: "X-Forwarded-By",value: "127.0.0.1",},
		addedHeader {name: "X-Forwarded",value: "127.0.0.1",},
		addedHeader {name: "X-Real-IP",value: "127.0.0.1",},
		addedHeader {name: "X-Custom-IP-Authorization",value: "127.0.0.1",},
	}
}

func main() {

	timeout := flag.Int("timeout", 10, "timeout for request")
	threads := flag.Int("threads", 5, "number of threads")
	flag.Parse()

	if 0 == *threads {
		return
	}


	var urls []string
	var wg sync.WaitGroup

	reader := bufio.NewScanner(os.Stdin)
	for reader.Scan() {
		urls = append(urls, reader.Text())
	}

	initHeaders()

	var currentNumOfThreads int = 0
	for _, url := range urls {
		if currentNumOfThreads == *threads {
			wg.Wait()
			currentNumOfThreads = 0
		}
		for _, header := range headers {
			currentNumOfThreads++
			wg.Add(1)
			go worker(url, header, *timeout, &wg)
		}

	}
	wg.Wait()
}
