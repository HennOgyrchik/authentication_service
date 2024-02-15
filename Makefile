TAG = "0.0.1"

docker-build:
	sudo docker build -t medods:$(TAG) .

run: docker-build
	docker compose up