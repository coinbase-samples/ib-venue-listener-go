REGION ?= us-east-1
PROFILE ?= sa-infra
ENV_NAME ?= dev
SYSTEM_ID ?= ib
BASE_URL ?= http://localhost:4566
PRICE_FEED_NAME ?= priceFeed
ORDER_FEED_NAME ?= orderFeed
ORDER_FILL_QUEUE_NAME ?= orderFillQueue.fifo

ACCOUNT_ID := $(shell aws sts get-caller-identity --profile $(PROFILE) --query 'Account' --output text)

setup-kinesis:
	@aws kinesis create-stream --endpoint=$(BASE_URL) --stream-name $(PRICE_FEED_NAME) --shard-count 1
	@aws kinesis create-stream --endpoint=$(BASE_URL) --stream-name $(ORDER_FEED_NAME) --shard-count 1

setup-sqs:
	@aws sqs create-queue --endpoint=$(BASE_URL) --queue-name $(ORDER_FILL_QUEUE_NAME) --attributes FifoQueue=true,ContentBasedDeduplication=true

docker-build:
	@docker build --platform linux/amd64 --build-arg REGION=$(REGION) --build-arg ENV_NAME=$(ENV_NAME) --build-arg ACCOUNT_ID=$(ACCOUNT_ID) .

docker-build-local:
	@docker build --tag ib-venue-listener-go:local --build-arg REGION=$(REGION) --build-arg ENV_NAME=local --build-arg ACCOUNT_ID=$(ACCOUNT_ID) .


.PHONY: docker-build setup-kinesis setup-sqs docker-build-local