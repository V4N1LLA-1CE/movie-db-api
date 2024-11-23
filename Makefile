# load env
include .env
export

# create postgres container on your machine
pg-container:
	docker run --name ${POSTGRES_CONTAINER_NAME} \
		-p 5432:5432 \
		-e POSTGRES_USER=${POSTGRES_SUPERUSER} \
		-e POSTGRES_PASSWORD=${POSTGRES_SUPERUSER_PASSWORD} \
		-d postgres:latest

# hop into container
pg-exec:
	docker exec -it ${POSTGRES_CONTAINER_NAME} /bin/bash

# run with live reloading
watch:
	air

.PHONY: watch pg-container pg-exec
