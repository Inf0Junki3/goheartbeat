package main

import (
    "flag"
    "fmt"
    "encoding/json"
    "errors"
    "io/ioutil"
    "log"
    "log/syslog"
    "net"
    "net/http"
    "strings"
    "sync"
    "time"
)

type Config struct {
    HeartbeatIntervalSeconds int
    Urls []string
    TcpEndpoints []string
    TimeoutSeconds int
}

func main(){
    
    configFilePath := flag.String("c", "/etc/heartbeat.config", "The heartbeat config file. Defaults to /etc/heartbeat.config")
    flag.Parse()
    
    config := readConfig(*configFilePath)

    totalIssues := []string{}
    
    logWriter, err := syslog.New(syslog.LOG_ALERT, "heartbeat")
    if err == nil {
        log.SetOutput(logWriter)
    }
    
    var waitGroup sync.WaitGroup

    for {
        go heartbeatUrls(&waitGroup, config.Urls, &totalIssues, time.Duration(config.TimeoutSeconds) * time.Second)
        waitGroup.Add(1)
        go heartbeatTcp(&waitGroup, config.TcpEndpoints, &totalIssues, time.Duration(config.TimeoutSeconds) * time.Second)
        waitGroup.Add(1)

        waitGroup.Wait()

        if len(totalIssues) > 0{
            log.Print(strings.Join(totalIssues, "\n"))
        }
        
        time.Sleep(time.Duration(config.HeartbeatIntervalSeconds) * time.Second)
    }
}

func readConfig(configPath string) Config {
    configJson, err := ioutil.ReadFile(configPath)
    if err != nil {
        panic(err)
    }

    var config Config
    if err := json.Unmarshal(configJson, &config); err != nil {
        panic(err)
    }

    if config.Urls == nil {
        panic(errors.New("Urls parameter is missing."))
    }

    if config.TcpEndpoints == nil {
        panic(errors.New("TcpEndpoints parameter is missing."))
    }

    return config
}

func heartbeatUrls(waitGroup *sync.WaitGroup, urls []string, issues *[]string, timeout time.Duration) {
    defer waitGroup.Done()

    client := http.Client{
        Timeout: timeout,
    }
    
    for _, curSite := range urls {
        response, err := client.Get(curSite)

        if err != nil {
            print(fmt.Sprintf("Ooops. %s\n", err))
            *issues = append(*issues, err.Error())
        } else {
            print(fmt.Sprintf("%s: %s\n", curSite, response.Status))
        }
    }
}

func heartbeatTcp(waitGroup *sync.WaitGroup, endpoints []string, issues *[]string, timeout time.Duration) {
    defer waitGroup.Done()
    for _, curEndpoint := range(endpoints) {
        conn, err := net.DialTimeout("tcp", curEndpoint, timeout)
        if err != nil {
            print(fmt.Sprintf("Oops. %s\n", err))
            *issues = append(*issues, err.Error())
        } else {
            print(fmt.Sprintf("Connection to %s successful.\n", curEndpoint))
            conn.Close()
        }
    }
}
