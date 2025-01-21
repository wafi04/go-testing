module github.com/wafi04/go-testing/auth

go 1.22.0

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.2
	github.com/wafi04/common v0.0.0-20250118112750-604b0b0abc8f
	github.com/wafi04/shared v0.0.0-20250116124558-f6dedf29cbd0
	golang.org/x/crypto v0.32.0
	google.golang.org/grpc v1.69.4
	google.golang.org/protobuf v1.35.1
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241015192408-796eee8c2d53 // indirect
)

replace github.com/wafi04/go-testing/common => ../common

