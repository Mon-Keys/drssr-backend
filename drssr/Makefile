run-api:
	go run cmd/api/main.go

run-bot:
	go run cmd/tg_bot/main.go

background_run-api:
	nohup go run cmd/api/main.go > api_logs.log &

background_run-bot:
	nohup go run cmd/tg_bot/main.go > bot_logs.log &

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
