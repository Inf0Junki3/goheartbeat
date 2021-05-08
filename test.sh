# This file performs the base tests for heartbeat. If something does not pass, first check that the 
# endpoints are correct (e.g. that they're live. You can do this with curl, netcat or ping), then
# check the code.
go test *.go -v -coverprofile=coverage.out
go tool cover -html coverage.out -o coverage.html