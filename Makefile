include .env
export

LOCAL_BIN:=$(CURDIR)/bin
PATH:=$(LOCAL_BIN):$(PATH)

swag-v1: ### swag init
	swag init -g internal/controller/http/v1/router.go
.PHONY: swag-v1

run: swag-v1 ### swag run
	go mod tidy && go mod download && \
	DISABLE_SWAGGER_HTTP_HANDLER='' GIN_MODE=debug CGO_ENABLED=0 go run -tags migrate ./cmd/app
.PHONY: run

build:
	go build -o ./bin/app -tags migrate ./cmd/app
.PHONY: build

mock: ### run mockgen
	mockgen -source ./internal/usecase/interfaces.go -package usecase_test > ./internal/usecase/mocks_test.go
.PHONY: mock

migrate-create:  ### create new migration
	migrate create -ext sql -dir migrations 'migrate_name'
.PHONY: migrate-create

migrate-down:
	migrate -path migrations -database '$(PG_URL)?sslmode=disable' down 

migrate-up: ### migration up
	migrate -path migrations -database '$(PG_URL)?sslmode=disable' up
.PHONY: migrate-up

bin-deps:
	GOBIN=$(LOCAL_BIN) go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	GOBIN=$(LOCAL_BIN) go install github.com/golang/mock/mockgen@latest

psql:
	docker exec -it postgres psql -U user -d postgres

push:
    git add . \
    git commit -m "feat: $(msg)"\
    git push

send:
	scp -r ./* nodirbek@192.168.100.17:/home/nodirbek/news

rr:
	sudo docker rm app-news -f
	sudo docker rmi news-app-news -f
	git pull
	sudo docker compose up