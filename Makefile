gen:
	buf generate
.PHONY: gen

run-unit-tests:
	go test -v -tags=unit ./...

run-integration-tests-ci:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit integration-tests

run-integration-tests: clean-test-db run-integration-tests-ci

clean-test-db:
	docker-compose -f docker-compose.test.yml down -v
