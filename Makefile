# Dinkur the task time tracking utility.
# <https://github.com/dinkur/dinkur>
#
# SPDX-FileCopyrightText: 2021 Kalle Fagerberg
# SPDX-License-Identifier: CC0-1.0

.PHONY: install clean tidy deps grpc docs \
	lint lint-md lint-go lint-proto lint-license \
	lint-fix lint-md-fix lint-proto-fix

ifeq ($(OS),Windows_NT)
dinkur.exe:
else
dinkur:
endif
	go1.18beta1 build -tags='fts5' -ldflags='-s -w'

install:
	go1.18beta1 install -tags='fts5' -ldflags='-s -w'

clean:
	rm -rfv ./dinkur.exe ./dinkur

tidy:
	go1.18beta1 mod tidy

deps:
	go1.18beta1 install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
	go1.18beta1 install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
	go1.18beta1 install github.com/mgechev/revive@latest
	go1.18beta1 install golang.org/x/tools/cmd/goimports@latest
	go1.18beta1 install github.com/yoheimuta/protolint/cmd/protolint@latest
	python3 -m pip install --upgrade --user reuse
	npm install

docs:
	go run internal/cmd/docgen/docgen.go docs/cmd

grpc:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		api/dinkurapi/v1/event.proto \
		api/dinkurapi/v1/entries.proto \
		api/dinkurapi/v1/alerter.proto

lint: lint-md lint-go lint-license lint-proto
lint-fix: lint-md-fix lint-proto-fix

lint-md:
	npx remark . .github

lint-md-fix:
	npx remark . .github -o

lint-go:
	revive -formatter stylish -config revive.toml ./...

lint-proto:
	protolint lint api/dinkurapi/v1

lint-proto-fix:
	protolint lint -fix api/dinkurapi/v1

lint-license:
	reuse lint
