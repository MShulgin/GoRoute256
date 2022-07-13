run_infra:
	docker-compose up -d zookeeper kafka-1 etcd redis pg-offer pg-post pg-shipment-1 pg-shipment-2
	./init_cluster.sh

run_offer:
	go run offer/cmd/main.go -conf offer/config/offer.yml

run_post:
	go run post/cmd/main.go -conf post/config/post.yml

run_shipment:
	go run shipment/cmd/main.go -conf shipment/config/shipment.yml

run_test:
	go run test/main.go

docker-all:
	docker build -t offer:1.0 -f deploy/docker/Dockerfile.offer .