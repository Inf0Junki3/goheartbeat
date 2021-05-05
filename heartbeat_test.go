package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "testing"
    "time"
)

type TestParameters struct {
    UrlAlwaysWorking string
    UrlDomainDoesNotExist string
    UrlServiceDoesNotExist string
    TcpEndpointAlwaysWorking string
    TcpEndpointNeverWorking string
}

var test_parameters TestParameters

func TestMain(m *testing.M) {
    parameters_json, _ := ioutil.ReadFile("tests/test_parameters.config")
    if err := json.Unmarshal(parameters_json, &test_parameters); err != nil {
        panic(err)
    }
    os.Exit(m.Run())
}

func TestReadConfig (t *testing.T) {
    defer func() {
        if r:= recover(); r == nil {
            print("File loaded.\n")
        } else {
            t.Fatalf("Problem parsing default config file. JSON might be malformed.")
        }
    }()
    config := read_config("tests/heartbeat_base.config")
    
    if config.HeartbeatIntervalSeconds != 30 {
        t.Fatalf("HeartbeatIntervalSeconds did not load with the expected value.")
    }
    
    if config.TimeoutSeconds != 5 {
        t.Fatalf("TimeoutSeconds did not load with the expected value.")
    }
    
    if !(len(config.Urls) == 2 && 
         config.Urls[0] == "http://www.google.com" &&
         config.Urls[1] == "https://www.google.com") {
        t.Fatalf("Urls did not load with the expected value.")
    }
    
    if !(len(config.TcpEndpoints) == 1 && config.TcpEndpoints[0] == "192.168.1.1:22") {
        t.Fatalf("TcpEndpoints did not load with the expected value.")
    }
}

func TestReadConfigDoesNotExist (t *testing.T) {
    defer func() {
        if r:= recover(); r != nil {
            print("Test passed.\n")
        } else {
            t.Fatalf("An error is expected here!")
        }
    }()
    read_config("thisfiledoesnotexist.config")
}

func TestReadConfigMalformedJson (t *testing.T) {
    defer func() {
        if r:= recover(); r != nil {
            print("Test passed.\n")
        } else {
            t.Fatalf("An error is expected here!")
        }
    }()
    read_config("tests/heartbeat_malformed.config")
}

func TestReadConfigMissingInterval (t *testing.T) {
    config := read_config("tests/heartbeat_missing_interval.config")
    if config.HeartbeatIntervalSeconds != 0 {
        t.Fatalf("A missing heartbeat interval should result in a value of 0.")
    }
}

func TestReadConfigMissingUrls (t *testing.T) {
    defer func() {
        if r:= recover(); r == nil {
            t.Fatalf("A missing URL list should result in an error.")
        } else {
            fmt.Printf("%v\n", r)
        }
    }()
    read_config("tests/heartbeat_missing_urls.config")
}

func TestReadConfigMissingTcpEndpoints (t *testing.T) {
    defer func() {
        if r:= recover(); r == nil {
            t.Fatalf("A missing TCP endpoint list should result in an error.")
        } else {
            fmt.Printf("%v\n", r)
        }
    }()
    read_config("tests/heartbeat_missing_tcpendpoints.config")
}

func TestUrlBase (t *testing.T) {
    total_issues := []string{}
    heartbeat_urls([]string{fmt.Sprintf("http://%s", test_parameters.UrlAlwaysWorking)}, &total_issues, time.Duration(3) * time.Second)
    if len(total_issues) > 0 {
        t.Fatalf("The URL is not live as expected.")
    }
}

func TestUrlEmpty (t *testing.T) {
    total_issues := []string{}
    heartbeat_urls([]string{}, &total_issues, time.Duration(3) * time.Second)
}

func TestUrlDoesNotExist (t *testing.T) {    
    total_issues := []string{}
    heartbeat_urls([]string{fmt.Sprintf("http://%s", test_parameters.UrlDomainDoesNotExist),}, 
                   &total_issues, 
                   time.Duration(1) * time.Second)
    if !(len(total_issues) == 1 && total_issues[0] == fmt.Sprintf("Get http://%s: dial tcp: lookup www.thisdoesntexist.local: no such host", test_parameters.UrlDomainDoesNotExist)) {
        t.Fatalf("The non-existing URL did not trigger an alert.")
    }
}

func TestUrlTimeout (t *testing.T) {    
    total_issues := []string{}
    heartbeat_urls([]string{fmt.Sprintf("http://%s", test_parameters.UrlServiceDoesNotExist),}, &total_issues, time.Duration(1) * time.Second)
    if !(len(total_issues) == 1 && total_issues[0] == fmt.Sprintf("Get http://%s: net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers)", test_parameters.UrlServiceDoesNotExist)) {
        print(total_issues[0])
        t.Fatalf("The time-out did not trigger an alert.")
    }
}

func TestTcpBase (t *testing.T) {
    total_issues := []string{}
    heartbeat_tcp([]string{test_parameters.TcpEndpointAlwaysWorking}, &total_issues, time.Duration(1) * time.Second)
    if len(total_issues) > 0 {
        t.Fatalf("The endpoint is not live as expected.")
    }
}

func TestTcpEmpty (t *testing.T) {
    total_issues := []string{}
    heartbeat_tcp([]string{}, &total_issues, time.Duration(1) * time.Second)
}

func TestTcpDoesNotExist (t *testing.T) {
    total_issues := []string{}
    heartbeat_tcp([]string{test_parameters.TcpEndpointNeverWorking,}, &total_issues, time.Duration(1) * time.Second)
    if !(len(total_issues) == 1 && total_issues[0] == fmt.Sprintf("dial tcp %s: i/o timeout", test_parameters.TcpEndpointNeverWorking)) {
        t.Fatalf("The non-existing service did not trigger an alert.")
    }
}
