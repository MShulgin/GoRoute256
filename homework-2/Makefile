proto:
	MODULES=("portfolio" "bot"); \
	for module in $${MODULES[@]}; do \
		protoc \
			-I ./api/portfolio \
			-I ${GOPATH}/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis/ \
			--grpc-gateway_out="./$$module/internal/pb" \
			--grpc-gateway_opt logtostderr=true \
			--grpc-gateway_opt paths=source_relative \
			--grpc-gateway_opt generate_unbound_methods=true \
			--go_out="./$$module/internal/pb" \
			--go_opt=paths=source_relative \
			--go-grpc_out="./$$module/internal/pb" \
			--go-grpc_opt=paths=source_relative \
		./api/portfolio/*.proto; \
	done;

test:
	go clean -testcache && go test ./commons/... && go test ./portfolio/... && go test ./bot/...

run_app:
	go run portfolio/cmd/main.go -conf config/portfolio.yml

run_bot:
	go run bot/cmd/main.go -conf config/bot.yml

build:
	docker-compose build

run:
	docker-compose up -d

shutdown:
	docker-compose down

cleanup:
	docker-compose down -v

migrate_new:
	goose -dir ./portfolio/db create $(NAME) sql

migrate_status:
	goose -dir ./portfolio/db postgres "user=app password=pwd dbname=portfolio sslmode=disable" status

migrate_up:
	goose -dir ./portfolio/db postgres "user=app password=pwd dbname=portfolio sslmode=disable" up

migrate_down:
	goose -dir ./portfolio/db postgres "user=app password=pwd dbname=portfolio sslmode=disable" down
