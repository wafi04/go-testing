gen-product:
	@protoc --experimental_allow_proto3_optional \
	--go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	pkg/proto/product/product.proto

gen-auth:
	@protoc --experimental_allow_proto3_optional \
	--go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	pkg/proto/auth/auth.proto