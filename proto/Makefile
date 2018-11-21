include Makefile.include

.PHONY: all
all: protogen

.PHONY: protogen
protogen:
	@echo "building protofile "
	protoc -I=${GIT_TOP}/grpcUpload/proto/ --go_out=plugins=grpc:${GIT_TOP}/grpcUpload/proto/ ${GIT_TOP}/grpcUpload/proto/upload.proto


