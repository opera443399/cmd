package main

import (
    "flag"
    "runtime"
    "log"
    "io/ioutil"
    "net/http"
    "sync"
    "strconv"
    "strings"
    "time"
)

const (
    defaultRepeatTimes = 3
    defaultTimeout = 1
)

var (
    repeatTimes 	int
    fromCfg        	string
    showDetails     bool
    taskList        []string
)

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

//Task task
type Task struct {
    url string
    ts taskState
}

//NewTask new task
func NewTask(url string) *Task {
    t := &Task{
        url: url,
        ts: taskState{
            v: make(map[string]int),
        },
    }

    return t
}

//Start task start
func (t *Task) Start(wg *sync.WaitGroup) {
    dtStart := time.Now()
    timeout := time.Duration(defaultTimeout * time.Second)
    wg.Add(1)
    defer wg.Done()

    c := http.Client{
        Timeout: timeout,
    }

    for i := 0; i < repeatTimes; i++ {
        if header, err := c.Head(t.url); err != nil {
            t.ts.Inc("failure")
            if !showDetails { continue }
            log.Printf("[%s] %s: failed(%v)", strconv.Itoa(i), t.url, err)
        } else {
            if header.StatusCode > 400 {
                t.ts.Inc("failure")
                if !showDetails { continue }
                log.Printf("[%s] %s: %s", strconv.Itoa(i), t.url, header.Status)
            } else {
                t.ts.Inc("success")
                if !showDetails { continue }
                log.Printf("[%s] %s: %s", strconv.Itoa(i), t.url, header.Status)
            }
        }
    }
    log.Printf("-> %d %d %s <- %s\n", 
                t.ts.Value("success"), t.ts.Value("failure"), time.Since(dtStart), t.url)
}

func init() {
    flag.IntVar(&repeatTimes, "c", defaultRepeatTimes, "Repeat N times")
    flag.StringVar(&fromCfg, "f", "", "Load URL list from config file")
    flag.BoolVar(&showDetails, "s", false, "Show details in stdout")
    runtime.GOMAXPROCS(runtime.NumCPU())
}

func loadDataFromFile(f string) []string {
    var urls []string
    if data, err := ioutil.ReadFile(f); err != nil {
        log.Fatalf("[E] %v", err)
    } else {
        for _, h := range strings.Split(string(data), "\n") {
            if h == "" {
                continue
            }
            urls = append(urls, h)
        }
    }
    
    return urls
}

func main() {
    flag.Parse()
    if len(flag.Args()) > 0 {
        taskList = flag.Args()
    } else if len(fromCfg) > 0 {
        taskList = loadDataFromFile(fromCfg)
    } else {
        log.Fatal("[E] no urls specifiled...")
    }

    wg := &sync.WaitGroup{}
    for _, url := range taskList {
        task := NewTask(url)
        go task.Start(wg)
        time.Sleep(defaultTimeout * time.Second)
    }
    wg.Wait()

}
