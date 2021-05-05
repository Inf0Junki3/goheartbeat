# GoHeartbeat: a simple(r) heartbeat service written in Go

## Overview

This is a simple service that allows you to continuously ping web and TCP services to ensure that
they are live.

The interval with which you ping your services is configurable. This and the list of URL's and TCP
services you want to ping are defined in /etc/heartbeat.config.

Installation is fairly simple - you first compile heartbeat with go; then, you install it to the
directory of your choice. Finally, you set up a service that calls the heartbeat executable.

We've included a setup.sh bash script in this repository that performs these steps, installing the
heartbeat executable to /bin/ and the heartbeat.config file to /etc/, then creating a systemd service.

## Setting up a heartbeat.config file

Once you've installed GoHeartbeat, the first thing you'll want to do is configure your heartbeats.
Here is what a typicaly configuration file looks like:

```json
{
    "HeartbeatIntervalSeconds": 30, // This is the amount of time between heartbeats.
    "Urls": ["http://www.google.com",
             "https://www.google.com"], // These are the URL's you want to check. http/https work.
    "TcpEndpoints": ["192.168.1.1:22"], // These are the services in IP:port format.
    "TimeoutSeconds": 5 // This is the amount of time to wait for an unresponsive connection.
}
```

## Getting the status of your heartbeats

When a heartbeat fails, a log entry of type `ALERT` is created in your syslog. However, GoHeartbeat
can also be consulted for its current status using `systemctl status`:

```bash
# systemctl status heartbeat.service
● heartbeat.service - Heartbeat service
   Loaded: loaded (/etc/systemd/system/heartbeat.service; enabled; vendor preset: enabled)
   Active: active (running) since Sun 2021-05-02 13:12:51 UTC; 2min 9s ago
 Main PID: XXXX (heartbeat)
    Tasks: 17 (limit: XXXX)
   CGroup: /system.slice/heartbeat.service
           └─XXXX /bin/heartbeat

May 02 13:14:52 whatever-host-is-running-heartbeat heartbeat[XXXX]: http://google.com: 200 OK
May 02 13:14:52 whatever-host-is-running-heartbeat heartbeat[XXXX]: https://google.com: 200 OK
May 02 13:14:52 whatever-host-is-running-heartbeat heartbeat[XXXX]: Connection to 192.168.1.1:22 successful.
```

## Tests

The tests in this repository are in `heartbeat_test.go`. A `test.sh` bash file is there to help you run the 
tests. When you run them, you should see something like this:

```bash
~/heartbeat$ bash test.sh
=== RUN   TestReadConfig
File loaded.
--- PASS: TestReadConfig (0.00s)
=== RUN   TestReadConfigDoesNotExist
Test passed.
--- PASS: TestReadConfigDoesNotExist (0.00s)
=== RUN   TestReadConfigMalformedJson
Test passed.
--- PASS: TestReadConfigMalformedJson (0.00s)
=== RUN   TestReadConfigMissingInterval
--- PASS: TestReadConfigMissingInterval (0.00s)
=== RUN   TestReadConfigMissingUrls
Urls parameter is missing.
--- PASS: TestReadConfigMissingUrls (0.00s)
=== RUN   TestReadConfigMissingTcpEndpoints
TcpEndpoints parameter is missing.
--- PASS: TestReadConfigMissingTcpEndpoints (0.00s)
=== RUN   TestUrlBase
http://www.google.com: 200 OK
--- PASS: TestUrlBase (0.23s)
=== RUN   TestUrlEmpty
--- PASS: TestUrlEmpty (0.00s)
=== RUN   TestUrlDoesNotExist
Ooops. Get http://www.thisdoesntexist.local: dial tcp: lookup www.thisdoesntexist.local: no such host
--- PASS: TestUrlDoesNotExist (0.00s)
=== RUN   TestUrlTimeout
Ooops. Get http://www.google.com:1234: net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers)
--- PASS: TestUrlTimeout (1.00s)
=== RUN   TestTcpBase
Connection to 8.8.8.8:53 successful.
--- PASS: TestTcpBase (0.01s)
=== RUN   TestTcpEmpty
--- PASS: TestTcpEmpty (0.00s)
=== RUN   TestTcpDoesNotExist
Oops. dial tcp 8.8.8.8:1234: i/o timeout
--- PASS: TestTcpDoesNotExist (1.00s)
PASS
ok      command-line-arguments  2.237s
```

## Raison d'être

There are other, really great free heartbeat services out there such as Elastic's Heartbeat service.
So why roll out GoHeartbeat?

GoHeartbeat's focus is on simplicity. You have a single executable that runs as a service. When services
go down, it logs the services down to your syslog. You can then use swatch, OSSEC, Splunk, or countless
other tools to alert you to issues. The idea here is to avoid setting up a complex system in an 
environment that does not call for it.

GoHeartbeat is functional, but unpolished. If you have any feature requests, please do open an issue.
Better yet, if you have some code that you wish to integrate, please don't hesitate to send a pull
request.