include Makefile.include

.PHONY: all
all: protogen

.PHONY: protogen
protogen:
	@echo "building protofile "
	protoc -I=${GIT_TOP}/proto/ --go_out=plugins=grpc:${GIT_TOP}/proto/ ${GIT_TOP}/proto/upload.proto


