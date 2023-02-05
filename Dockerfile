# Copyright 2021 The Events Exporter authors
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
FROM golang:1.19-alpine3.17 AS builder

RUN apk add --no-cache --update alpine-sdk bash

WORKDIR /usr/local/src/events_exporter

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""

ENV GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT}
ARG GOPROXY

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build


FROM alpine:3.17.1 AS main

RUN apk add --no-cache --update ca-certificates
COPY --from=builder /usr/local/src/events_exporter/bin/events_exporter /usr/local/bin/events_exporter

# nobody
USER 1001:1001

ENTRYPOINT ["events_exporter"]
