FROM golang:1.18.2-alpine as builder

ENV APP_HOME /go/src/hw2

WORKDIR $APP_HOME
COPY . .

RUN apk add --no-cache git make protobuf-dev && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest && \
    cd portfolio && \
    go mod tidy

RUN rm -rf ./portfolio/internal/pb/* && \
    protoc \
      -I ./api/portfolio \
      -I ${GOPATH}/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis/ \
      --grpc-gateway_out=./portfolio/internal/pb \
      --grpc-gateway_opt logtostderr=true \
      --grpc-gateway_opt paths=source_relative \
      --grpc-gateway_opt generate_unbound_methods=true \
      --go_out=./portfolio/internal/pb \
      --go_opt=paths=source_relative \
      --go-grpc_out=./portfolio/internal/pb \
      --go-grpc_opt=paths=source_relative \
    ./api/portfolio/*.proto && \
    go build $APP_HOME/portfolio/cmd/main.go


FROM alpine

RUN apk --no-cache add ca-certificates

WORKDIR /root/
RUN mkdir ./db
COPY --from=builder /go/src/hw2/main .
COPY --from=builder /go/src/hw2/portfolio/db/ ./db/

CMD ["./main"]
