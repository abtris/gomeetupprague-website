# https://suva.sh/posts/well-documented-makefiles/#simple-makefile
.DEFAULT_GOAL := help
SHELL := /bin/bash

TITLE := $(shell echo $(name) | tr [:upper:] [:lower:] | sed -e 's/ /_/g')

.PHONY: help deps clean build watch test test-videos update-videos

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## Build the project
	hugo

watch: ## Watch file changes and build
	hugo server -w

test: ## Run responsive Playwright tests before pushing
	npm test

test-videos: ## Test the Go video fetcher
	cd scripts && go test ./...

update-videos: test-videos ## Fetch the latest YouTube videos (requires API environment variables)
	cd scripts && go run .

post: ## Make new post: make post name="Title"
	hugo new --kind post "post/$$(date +%Y-%m-%d)-$(TITLE).markdown"
