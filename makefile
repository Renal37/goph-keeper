ALL_TARGETS = client server
GO = go

define build
	${GO} build -o ./cmd/$(1)/ -v ./cmd/$(1)/ 
endef

all: build

.PHONY: build
build: $(patsubst %, build-%, $(ALL_TARGETS))


.PHONY: build-%
build-%:
	@echo === Building $*
	$(call build,$*)

.PHONY: test
test:
	@echo === Tests
	${GO} test -count 1 -v -cover ./...

define clean
	rm ./cmd/$(1)/$(1)
endef

.PHONY: clean
clean: $(patsubst %, clean-%, $(ALL_TARGETS))

clean-%:
	@echo === Cleaning $*
	$(call clean,$*)

# Linter constants
LINTER := golangci-lint 

.PHONY: lint
lint:
	@echo === Lint
	$(LINTER) --version
	$(LINTER) cache clean && $(LINTER) run

generate:
	${GO} generate ./...

.PHONY: test_with_coverage
test_with_coverage:
	${GO} test -count=1 -coverprofile=coverage.out ./...

.PHONY: coverage_total
coverage_total: test_with_coverage
	PERCENTAGE=$$(${GO} tool cover -func=coverage.out | grep "total:" | tr -s '\t' | cut -f3); \
	echo Total coverage: $${PERCENTAGE}

.PHONY: coverage_html
coverage_html: test_with_coverage
	${GO} tool cover -html=coverage.out -o coverage.html && firefox ./coverage.html

run_server:
	docker stop gophkeeper-postgres > /dev/null
	docker rm gophkeeper-postgres > /dev/null
	docker run -d --name gophkeeper-postgres -p 5431:5432 -e POSTGRES_DB=gophkeeper-db -e POSTGRES_USER=gophkeeper-user -e POSTGRES_PASSWORD=12345 postgres:12-alpine && \
	sleep 5 && \
	cmd/server/server -c config.yaml