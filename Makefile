.PHONY: help clean test test-ci start-services docker-clean

.DEFAULT: help

help:
	@echo "Make Help for $(SERVICE_NAME)"
	@echo ""
	@echo "make clean   	- removes dynamodb data directory"
	@echo "make test    	- run tests locally using docker containers"
	@echo "make test-ci 	- used by Jenkins to run tests without exposing ports"
	@echo "start-dynamodb 	- Start local DynamoDB container for testing"

test:
	docker compose -f docker-compose.test.yml down --remove-orphans
	docker compose -f docker-compose.test.yml up --exit-code-from local_tests local_tests

test-ci:
	mkdir -p test-dynamodb-data
	chmod -R 777 test-dynamodb-data
	docker compose -f docker-compose.test.yml down --remove-orphans
	docker compose -f docker-compose.test.yml up --exit-code-from ci_tests ci_tests

# Start clean external service containers for local testing
start-services: docker-clean
	docker compose -f docker-compose.test.yml up -d dynamodb pennsievedb


# Spin down active docker containers.
docker-clean:
	docker compose -f docker-compose.test.yml down

# Remove dydb database
clean: docker-clean
	rm -rf test-dynamodb-data
