PROTO_FILES=$(shell find proto -name *.proto)

.PHONY: init
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: proto
proto:
	rm -rf ./proto/gen
	protoc --proto_path=./proto \
		--go_out=./ \
		$(PROTO_FILES)
	mv ./github.com/laixhe/gonet/proto/gen ./proto
	rm -rf ./github.com
	protoc-go-inject-tag -input="./proto/gen/*/*.pb.go"
	protoc-go-inject-tag -input="./proto/gen/*/*/*.pb.go"
