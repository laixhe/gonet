PROTOCOL_FILES=$(shell find protocol/api -name *.proto)

.PHONY: init
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: protocol
protocol:
	rm -rf ./protocol/gen
	protoc --proto_path=./protocol/api \
		--go_out=./ \
		$(PROTOCOL_FILES)
	mv ./github.com/laixhe/gonet/protocol/gen ./protocol
	rm -rf ./github.com
	#protoc-go-inject-tag -input="./protocol/gen/*/*.pb.go"
	protoc-go-inject-tag -input="./protocol/gen/*/*/*.pb.go"
