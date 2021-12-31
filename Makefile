# Dinkur the task time tracking utility.
# <https://github.com/dinkur/dinkur>
#
# SPDX-FileCopyrightText: 2021 Kalle Fagerberg
# SPDX-License-Identifier: CC0-1.0

.PHONY: install clean tidy deps grpc docs

ifeq ($(OS),Windows_NT)
dinkur.exe:
else
dinkur:
endif
	go build -tags='fts5' -ldflags='-s -w'

install:
	go install -tags='fts5' -ldflags='-s -w'

clean:
	rm -rfv ./dinkur.exe ./dinkur

tidy:
	go mod tidy

deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
	go install github.com/mgechev/revive@latest
	go install golang.org/x/tools/cmd/goimports@latest

docs:
	go run internal/cmd/docgen/docgen.go docs/cmd

grpc:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		api/dinkurapi/v1/dinkurapi.proto
