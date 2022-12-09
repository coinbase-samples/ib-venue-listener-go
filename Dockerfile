# Copyright 2022 Coinbase Global, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ARG ACCOUNT_ID
ARG REGION
ARG ENV_NAME

FROM $ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com/go-base-$ENV_NAME:latest as builder

ARG CACHEBUST=1

RUN mkdir -p /build
WORKDIR /build
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -o main cmd/main.go

FROM scratch

COPY --from=builder /build/main /main
COPY --from=builder /etc/ssl/certs /etc/ssl/certs


EXPOSE 8443
CMD ["/main"]
