include help.mk

# get content of .env as environment variables
-include .env
export

# get root dir
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
IMAGE_NAME := whatsapp-reminder
IMAGE_VERSION := latest

.DEFAULT_GOAL := start

.PHONY: update
update: ## pulls git repo
	@git -C ${ROOT_DIR} pull
	go mod tidy

.PHONY: test
test: ## run golang test (including integration tests)
	go test -timeout 0  ${ROOT_DIR}...

.PHONY: start
start: ## start via docker
	docker build . -t ${IMAGE_NAME}
	docker run --rm -p 8080:8080 ${IMAGE_NAME}

.PHONY: start-cluster
start-cluster: ## starts k3d cluster and registry
	@k3d cluster create --config ${ROOT_DIR}k3d/clusterconfig.yaml

.PHONY: start-k3d
start-k3d: start-cluster push-k3d ## start k3d cluster and deploy local code
	@helm install ${IMAGE_NAME} ${ROOT_DIR}charts/${IMAGE_NAME} \
		--set image.repository=registry.localhost:5000/${IMAGE_NAME} \
		--set image.tag=${IMAGE_VERSION} \
		--set config.googleSheets.spreadsheetId=$(SPREADSHEET_ID) \
		--set-string config.googleSheets.sheetName="$(SHEET_NAME)" \
		--set secrets.serviceAccountJsonBase64=$(SERVICE_ACCOUNT_JSON_BASE64) \
		$(if $(EMAIL_ORIGIN_ADDRESS),--set secrets.emailOriginAddress=$(EMAIL_ORIGIN_ADDRESS),) \
		$(if $(EMAIL_ORIGIN_NAME),--set-string secrets.emailOriginName="$(EMAIL_ORIGIN_NAME)",) \
		$(if $(SCHEDULE),--set-string schedule="$(SCHEDULE)",) \
		$(if $(TIME_LOCATION),--set-string config.app.timeLocation="$(TIME_LOCATION)",) \
		$(if $(RETENTION_TIME),--set config.app.retentionTime=$(RETENTION_TIME),) \
		$(if $(EMAIL_SERVICE_URL),--set config.email.serviceUrl=$(EMAIL_SERVICE_URL),)

.PHONY: stop-k3d
stop-k3d: ## stop K3d cluster
	@k3d cluster delete --config ${ROOT_DIR}k3d/clusterconfig.yaml

.PHONY: restart-k3d
restart-k3d: stop-k3d start-k3d ## restarts K3d cluster
	
.PHONY: push-k3d
push-k3d: ## build and push docker image to local registry
	@docker build -f ${ROOT_DIR}Dockerfile . -t ${IMAGE_NAME}
	@docker tag ${IMAGE_NAME} localhost:5000/${IMAGE_NAME}:${IMAGE_VERSION}
	@docker push localhost:5000/${IMAGE_NAME}:${IMAGE_VERSION}

.PHONY: lint
lint: ## run golangci-lint
	golangci-lint run ${ROOT_DIR}...

.PHONY: generate-helm-docs
generate-helm-docs: ## re-generates helm docs using docker
	@docker run --rm --volume "$(ROOT_DIR)charts:/helm-docs" jnorwood/helm-docs:latest
