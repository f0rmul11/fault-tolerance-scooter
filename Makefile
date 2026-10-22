SERVICES := profile scooter rental pricing api-gateway
MODULES := $(addprefix services/,$(SERVICES)) scooterctl

.PHONY: all cli build up deploy stop status test vet generate generate-local proto-tools xo-tools generate-xo migrate
all: cli
cli:
	mkdir -p bin
	cd scooterctl && GOWORK=off go build -trimpath -o ../bin/scooterctl ./cmd/scooterctl
up: cli
	./bin/scooterctl up --timeout 15m
build: cli
	./bin/scooterctl build --timeout 15m
deploy: cli
	./bin/scooterctl deploy --timeout 10m
stop: cli
	./bin/scooterctl stop
status: cli
	./bin/scooterctl status
test:
	@set -e; for module in $(MODULES); do (cd "$$module" && GOWORK=off go test -race ./...); done
vet:
	@set -e; for module in $(MODULES); do (cd "$$module" && GOWORK=off go vet ./...); done
generate:
	@set -e; for service in $(SERVICES); do $(MAKE) -C services/$$service generate; done
generate-local:
	@set -e; for service in $(SERVICES); do $(MAKE) -C services/$$service generate-local; done
proto-tools:
	$(MAKE) -C services/api-gateway proto-tools
xo-tools:
	./tools/install-xo.sh
generate-xo:
	./tools/generate-xo.sh "$(SERVICE)"
migrate:
	@test -n "$(SERVICE)" || (echo 'Specify SERVICE and POSTGRES_DSN' >&2; exit 1)
	goose -dir "services/$(SERVICE)/tools/migrations" postgres "$$POSTGRES_DSN" up
