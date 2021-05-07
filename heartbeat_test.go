package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "strings"
    "sync"
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

var testParameters TestParameters
var waitGroup sync.WaitGroup

func TestMain(m *testing.M) {
    parametersJson, _ := ioutil.ReadFile("tests/test_parameters.config")
    if err := json.Unmarshal(parametersJson, &testParameters); err != nil {
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
    config := readConfig("tests/heartbeat_base.config")
    
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
    readConfig("thisfiledoesnotexist.config")
}

func TestReadConfigMalformedJson (t *testing.T) {
    defer func() {
        if r:= recover(); r != nil {
            print("Test passed.\n")
        } else {
            t.Fatalf("An error is expected here!")
        }
    }()
    readConfig("tests/heartbeat_malformed.config")
}

func TestReadConfigMissingInterval (t *testing.T) {
    config := readConfig("tests/heartbeat_missing_interval.config")
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
    readConfig("tests/heartbeat_missing_urls.config")
}

func TestReadConfigMissingTcpEndpoints (t *testing.T) {
    defer func() {
        if r:= recover(); r == nil {
            t.Fatalf("A missing TCP endpoint list should result in an error.")
        } else {
            fmt.Printf("%v\n", r)
        }
    }()
    readConfig("tests/heartbeat_missing_tcpendpoints.config")
}

func TestUrlBase (t *testing.T) {
    totalIssues := []string{}
    waitGroup.Add(1)
    go heartbeatUrls(&waitGroup, []string{fmt.Sprintf("http://%s", testParameters.UrlAlwaysWorking)}, &totalIssues, time.Duration(3) * time.Second)
    waitGroup.Wait()
    if len(totalIssues) > 0 {
        t.Fatalf("The URL is not live as expected.")
    }
}

func TestUrlEmpty (t *testing.T) {
    totalIssues := []string{}
    waitGroup.Add(1)
    go heartbeatUrls(&waitGroup, []string{}, &totalIssues, time.Duration(3) * time.Second)
    waitGroup.Wait()
}

func TestUrlDoesNotExist (t *testing.T) {    
    totalIssues := []string{}
    waitGroup.Add(1)
    go heartbeatUrls(&waitGroup, []string{fmt.Sprintf("http://%s", testParameters.UrlDomainDoesNotExist),},
                   &totalIssues,
                   time.Duration(1) * time.Second)
    waitGroup.Wait()
    if !(len(totalIssues) == 1 && totalIssues[0] == fmt.Sprintf("Get http://%s: dial tcp: lookup www.thisdoesntexist.local: no such host", testParameters.UrlDomainDoesNotExist)) {
        t.Fatalf("The non-existing URL did not trigger an alert.")
    }
}

func TestUrlTimeout (t *testing.T) {    
    totalIssues := []string{}
    waitGroup.Add(1)
    go heartbeatUrls(&waitGroup, []string{fmt.Sprintf("http://%s", testParameters.UrlServiceDoesNotExist),}, &totalIssues, time.Duration(1) * time.Second)
    waitGroup.Wait()
    if !(len(totalIssues) == 1 && strings.Contains(totalIssues[0], "(Client.Timeout exceeded while awaiting headers)")) {
        print(totalIssues[0])
        t.Fatalf("The time-out did not trigger an alert.")
    }
}

func TestTcpBase (t *testing.T) {
    totalIssues := []string{}
    waitGroup.Add(1)
    go heartbeatTcp(&waitGroup, []string{testParameters.TcpEndpointAlwaysWorking}, &totalIssues, time.Duration(1) * time.Second)
    waitGroup.Wait()
    if len(totalIssues) > 0 {
        t.Fatalf("The endpoint is not live as expected.")
    }
}

func TestTcpEmpty (t *testing.T) {
    totalIssues := []string{}
    waitGroup.Add(1)
    go heartbeatTcp(&waitGroup, []string{}, &totalIssues, time.Duration(1) * time.Second)
    waitGroup.Wait()
}

func TestTcpDoesNotExist (t *testing.T) {
    totalIssues := []string{}
    waitGroup.Add(1)
    go heartbeatTcp(&waitGroup, []string{testParameters.TcpEndpointNeverWorking,}, &totalIssues, time.Duration(1) * time.Second)
    waitGroup.Wait()
    if !(len(totalIssues) == 1 && totalIssues[0] == fmt.Sprintf("dial tcp %s: i/o timeout", testParameters.TcpEndpointNeverWorking)) {
        t.Fatalf("The non-existing service did not trigger an alert.")
    }
}