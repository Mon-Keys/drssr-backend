run:
	go run cmd/main.go

.PHONY: docker
docker:
	docker stop drssr_api || true
	docker rm drssr_api || true
	docker build . -t drssr_api
	docker run --restart always \
	--add-host=host.docker.internal:host-gateway \
	--name drssr_api \
	-p 3001:3001 -d \
	-v drssr:/posts \
	-v drssr:/looks \
	drssr_api
