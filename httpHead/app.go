/*
# HttpHead
# 2018/11/6
*/

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultTestURL1 = "https://www.baidu.com"
	defaultTestURL2 = "https://www.qq.com"
	defaultFrequency = 10
	defaultTimeout = 1
	defaultStdout = false
)

var (
	cnt            int
	fromCfg        string
	targetWebsites []string
)

func init() {
	flag.IntVar(&cnt, "c", defaultFrequency, "[] repeat N times to request the URL.")
	flag.StringVar(&fromCfg, "f", "", "[] load URLs from file.")
	runtime.GOMAXPROCS(runtime.NumCPU())
}

type taskState struct {
	v   map[string]int
	mux sync.Mutex
}

func (ts *taskState) Inc(key string) {
	ts.mux.Lock()
	ts.v[key]++
	ts.mux.Unlock()
}

func (ts *taskState) Value(key string) int {
	ts.mux.Lock()
	defer ts.mux.Unlock()
	return ts.v[key]
}

func loadDataFromFile(f string) []string {
	var urls []string
	data, err := ioutil.ReadFile(f)
	if err != nil {
		log.Printf("[E] %s : %v", "ioutil.ReadFile", err)
		return urls
	}

	for _, h := range strings.Split(string(data), "\n") {
		if h == "" {
			continue
		}
		urls = append(urls, h)
	}
	return urls
}

func handleHTTPHeadRequest(cnt int, url string, ch chan string, stat *taskState) {
	for i := 0; i < cnt; i++ {
		header, err := http.Head(url)
		if err != nil {
			log.Printf("[E] %s : %v", "http.Head", err)
			stat.Inc("failure")
			ch <- "[" + strconv.Itoa(i) + "]" + url + " : failed."
		} else {
			stat.Inc("success")
			status := header.Status
			ch <- "[" + strconv.Itoa(i) + "]" + url + " : " + status
		}
	}

}

func main() {
	flag.Parse()
	if len(flag.Args()) > 0 {
		targetWebsites = flag.Args()
	} else if len(fromCfg) > 0 {
		targetWebsites = loadDataFromFile(fromCfg)
	} else {
		targetWebsites = []string{defaultTestURL1, defaultTestURL2}
	}

	dtStart := time.Now()
	stat := taskState{v: make(map[string]int)}
	ch := make(chan string)
	chTheEnd := make(chan bool)

	for _, url := range targetWebsites {
		go handleHTTPHeadRequest(cnt, url, ch, &stat)
	}

	go func() {
		timer := time.NewTimer(time.Second * defaultTimeout)
		for {
			if !timer.Stop() {
				select {
				case <-timer.C: //try to drain from the channel
				default:
				}
			}
			timer.Reset(time.Second * defaultTimeout)
			select {
			case msg, ok := <-ch:
				if ok {
					fmt.Println(msg)
					continue
				}
				chTheEnd <- true
			case <-timer.C:
				log.Printf("timer expired (%ds)", defaultTimeout)
				chTheEnd <- true
			}
		}
	}()

	<-chTheEnd
	log.Printf("success: %d, failure: %d, Time Cost: %v\n", stat.Value("success"), stat.Value("failure"), time.Since(dtStart))

}
