.PHONY: docker
docker:
	docker build . -t post-api:latest
