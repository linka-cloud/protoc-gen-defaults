# Copyright 2021 Linka Cloud  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

MODULE = go.linka.cloud/protoc-gen-defaults

DEFAULTS_PROTO = defaults/defaults.proto

$(shell mkdir -p .bin)

export GOBIN=$(PWD)/.bin

export PATH := $(GOBIN):$(PATH)

include .bingo/Variables.mk

bin:
	@go install github.com/bwplotka/bingo@latest
	@bingo get
	@bingo list|tail -n +3|awk '{print $$2}'|xargs -I{} -n1 bash -c 'ln -sf {} $(GOBIN)/$$(bingo list|grep {}|awk "{print \$$1}")'

clean:
	@rm -rf .bin
	@find . -name '*.pb*.go' -type f -exec rm {} \;

.PHONY: proto
proto: gen-proto lint

.PHONY: gen-proto
gen-proto: defaults-proto install
	@buf generate

.PHONY: defaults-proto
defaults-proto: bin
	@buf generate --template buf.go.yaml --path $(DEFAULTS_PROTO)

.PHONY: lint
lint:
	@goimports -w -local $(MODULE) $(PWD)
	@gofmt -w $(PWD)

.PHONY: tests
tests: proto
	@go test -v ./module
	@go test -v ./tests


.PHONY: install
install:
	@go install .

.PHONY: gen-debug
gen-debug: proto
	@protoc -I. --debug_out="debug:." tests/pb/test.proto
