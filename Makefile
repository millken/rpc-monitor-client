.PHONY: proto run

proto:
	for f in */*.proto; do \
		protoc --go_out=plugins=grpc:. $$f; \
		echo compiled: $$f; \
	done

docker:
	docker build -t dnsoa/romaer .
run:
	docker-compose build
	docker-compose up --remove-orphans
