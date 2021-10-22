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


PROTO_BASE_PATH = .
TEST_PROTO_BASE_PATH = $(PROTO_BASE_PATH)/tests/pb

INCLUDE_PROTO_PATH = -I$(PROTO_BASE_PATH) \
	-I $(shell go list -m -f {{.Dir}} google.golang.org/protobuf)

PROTO_OPTS = paths=source_relative

DEFAULTS_PROTO = $(PROTO_BASE_PATH)/defaults/defaults.proto

$(shell mkdir -p .bin)

export GOBIN=$(PWD)/.bin

export PATH := $(GOBIN):$(PATH)

bin:
	@go install github.com/golang/protobuf/protoc-gen-go
	@go install github.com/lyft/protoc-gen-star/protoc-gen-debug

clean:
	@rm -rf .bin
	@find $(PROTO_BASE_PATH) -name '*.pb*.go' -type f -exec rm {} \;

.PHONY: proto
proto: gen-proto lint

.PHONY: defaults-proto
defaults-proto:
	@protoc $(INCLUDE_PROTO_PATH) --go_out=$(PROTO_OPTS):. $(DEFAULTS_PROTO)

.PHONY: gen-proto
gen-proto: defaults-proto install
	@find $(PROTO_BASE_PATH) -name '*.proto' -type f -not -path "$(DEFAULTS_PROTO)" -exec \
    	protoc $(INCLUDE_PROTO_PATH) --go_out=$(PROTO_OPTS):. --defaults_out=$(PROTO_OPTS):. {} \;

.PHONY: lint
lint:
	@goimports -w -local $(MODULE) $(PWD)
	@gofmt -w $(PWD)

.PHONY: tests
tests: proto gen-tests
	@go test -v ./module
	@go test -v ./tests


.PHONY: install
install:
	@go install .

.PHONY: gen-debug
gen-debug: defaults-proto
	@protoc -I. --debug_out="debug:." tests/pb/test.proto

.PHONY: gen-tests
gen-tests:
	@@find $(TEST_PROTO_BASE_PATH) -name '*.proto' -type f -exec \
         	protoc $(INCLUDE_PROTO_PATH) --go_out=$(PROTO_OPTS):. --defaults_out=$(PROTO_OPTS):. {} \;
