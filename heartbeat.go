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
    "time"
)

type Config struct {
    HeartbeatIntervalSeconds int
    Urls []string
    TcpEndpoints []string
    TimeoutSeconds int
}

func main(){
    
    config_file_path := flag.String("c", "/etc/heartbeat.config", "The heartbeat config file. Defaults to /etc/heartbeat.config")
    flag.Parse()
    
    config := read_config(*config_file_path)

    total_issues := []string{}
    
    logWriter, err := syslog.New(syslog.LOG_ALERT, "heartbeat")
    if err == nil {
        log.SetOutput(logWriter)
    }
    
    for {
        heartbeat_urls(config.Urls, &total_issues, time.Duration(config.TimeoutSeconds) * time.Second)
        heartbeat_tcp(config.TcpEndpoints, &total_issues, time.Duration(config.TimeoutSeconds) * time.Second)

        if len(total_issues) > 0{
            log.Print(strings.Join(total_issues, "\n"))
        }
        
        time.Sleep(time.Duration(config.HeartbeatIntervalSeconds) * time.Second)
    }
}

func read_config(config_path string) Config {
    config_json, err := ioutil.ReadFile(config_path)
    if err != nil {
        panic(err)
    }

    var config Config
    if err := json.Unmarshal(config_json, &config); err != nil {
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

func heartbeat_urls(urls []string, issues *[]string, timeout time.Duration) {
    client := http.Client{
        Timeout: timeout,
    }
    
    for _, cur_site := range urls {
        response, err := client.Get(cur_site)

        if err != nil {
            print(fmt.Sprintf("Ooops. %s\n", err))
            *issues = append(*issues, err.Error())
        } else {
            print(fmt.Sprintf("%s: %s\n", cur_site, response.Status))
        }
    }
}

func heartbeat_tcp(endpoints []string, issues *[]string, timeout time.Duration) {
    for _, cur_endpoint := range(endpoints) {
        conn, err := net.DialTimeout("tcp", cur_endpoint, timeout)
        if err != nil {
            print(fmt.Sprintf("Oops. %s\n", err))
            *issues = append(*issues, err.Error())
        } else {
            print(fmt.Sprintf("Connection to %s successful.\n", cur_endpoint))
            conn.Close()
        }
    }
}
