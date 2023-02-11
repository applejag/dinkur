# Dinkur the task time tracking utility.
# <https://github.com/dinkur/dinkur>
#
# SPDX-FileCopyrightText: 2021 Kalle Fagerberg
# SPDX-License-Identifier: CC0-1.0

ifeq ($(OS),Windows_NT)
OUT_FILE = dinkur.exe
else
OUT_FILE = dinkur
endif

GO_FILES = $(shell git ls-files "*.go" ":!:api")

.PHONY: all
all: grpc build docs

.PHONY: build
build: $(OUT_FILE)

$(OUT_FILE): $(GO_FILES) dinkur.schema.json
	go build -o dinkur -tags='fts5' -ldflags='-s -w'

.PHONY: install
install:
	go install -tags='fts5' -ldflags='-s -w'

.PHONY: clean
clean:
	rm -rfv ./dinkur.exe ./dinkur

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: deps
deps: deps-go deps-pip deps-npm

.PHONY: deps-go
deps-go:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
	go install github.com/mgechev/revive@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/yoheimuta/protolint/cmd/protolint@latest

.PHONY: deps-pip
deps-pip:
	python3 -m pip install --upgrade --user reuse

.PHONY: deps-npm
deps-npm: node_modules

node_modules:
	npm install

.PHONY: docs
docs: docs/cmd/*.md dinkur.schema.json

docs/cmd/dinkur_%.md: cmd/%.go internal/cmd/docgen/docgen.go
	go run internal/cmd/docgen/docgen.go docs/cmd

dinkur.schema.json: $(GO_FILES) cmd/config_schema.go
	go run . config schema --output dinkur.schema.json

.PHONY: grpc
grpc: api/dinkurapi/v1/*.pb.go api/dinkurapi/v1/*_grpc.pb.go

api/dinkurapi/v1/%.pb.go api/dinkurapi/v1/%_grpc.pb.go: api/dinkurapi/v1/%.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		api/dinkurapi/v1/event.proto \
		api/dinkurapi/v1/entries.proto \
		api/dinkurapi/v1/statuses.proto

.PHONY: lint
lint: lint-md lint-go lint-proto lint-license

.PHONY: lint-fix
lint-fix: lint-md-fix lint-go-fix lint-proto-fix

.PHONY: lint-md
lint-md: node_modules
	npx remark . .github

.PHONY: lint-md-fix
lint-md-fix: node_modules
	npx remark . .github -o

.PHONY: lint-go
lint-go:
	@echo goimports -d '**/*.go'
	@goimports -d $(GO_FILES)
	revive -formatter stylish -config revive.toml ./...

.PHONY: lint-go-fix
lint-fix-go:
	@echo goimports -d -w '**/*.go'
	@goimports -d -w $(GO_FILES)

.PHONY: lint-proto
lint-proto:
	protolint lint api/dinkurapi/v1

.PHONY: lint-proto-fix
lint-proto-fix:
	protolint lint -fix api/dinkurapi/v1

.PHONY: lint-license
lint-license:
	reuse lint
