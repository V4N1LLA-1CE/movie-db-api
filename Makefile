# load env
include .env
export

# create postgres container on your machine
pg-container:
	docker run --name ${POSTGRES_CONTAINER_NAME} \
		-p 5432:5432 \
		-e POSTGRES_USER=${POSTGRES_USER} \
		-e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} \
		-d postgres:latest

# create db inside container
pg-createdb:
	docker exec -it ${POSTGRES_CONTAINER_NAME} createdb --username=${POSTGRES_USER} --owner=${POSTGRES_USER} ${POSTGRES_DB_NAME}

# drop db inside container
pg-dropdb:
	docker exec -it ${POSTGRES_CONTAINER_NAME} dropdb ${POSTGRES_DB_NAME}

# hop into container
pg-exec:
	docker exec -it ${POSTGRES_CONTAINER_NAME} /bin/bash

# db migrate up
pg-migrateup:
	migrate -path db/migrations -database "${PG_DSN}" -verbose up

# db migrate down
pg-migratedown:
	migrate -path db/migrations -database "${PG_DSN}" -verbose down

pg-build:
	make pg-container && \
	until docker exec ${POSTGRES_CONTAINER_NAME} pg_isready -U ${POSTGRES_USER}; do sleep 1; done && \
	make pg-createdb && \
	make pg-migrateup

# run with live reloading
watch:
	air

.PHONY: watch pg-container pg-exec pg-migrateup pg-migratedown pg-createdb pg-dropdb pg-build
