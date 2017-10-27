# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

OUTPUT_DIR?=_output
REGISTRY?=gcr.io/k8s-testimages/fedtidy
VERSION?=0.1

.PHONY: all
all: build build-image push

.PHONY: build
build:
	go build -o $(OUTPUT_DIR)/fedtidy .

build-image:
	cp Dockerfile $(OUTPUT_DIR)
	docker build -t $(REGISTRY):$(VERSION) $(OUTPUT_DIR)

push:
	gcloud docker -- push $(REGISTRY):$(VERSION)

.PHONY: clean
clean:
	rm -r _output
