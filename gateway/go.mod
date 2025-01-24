module github.com/wafi04/go-testing/gateway

go 1.22.0

require (
	github.com/gorilla/mux v1.8.1
	github.com/wafi04/go-testing/auth v0.0.0
	github.com/wafi04/go-testing/category v0.0.0
	github.com/wafi04/go-testing/common v0.0.0
	github.com/wafi04/go-testing/product v0.0.0
	golang.org/x/exp v0.0.0-20250106191152-7588d65b2ba8
	google.golang.org/grpc v1.70.0
)

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241202173237-19429a94021a // indirect
	google.golang.org/protobuf v1.36.3 // indirect
)

replace github.com/wafi04/go-testing/common => ../common

replace github.com/wafi04/go-testing/auth => ../auth

replace github.com/wafi04/go-testing/category => ../category

replace github.com/wafi04/go-testing/product => ../product
