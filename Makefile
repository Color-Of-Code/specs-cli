.PHONY: lint format format-check check test vet build build-engine build-extension package-extension deploy-dev

lint:
	cd engine && go run ./cmd/specs lint --style

format:
	cd engine && go run ./cmd/specs format

format-check:
	cd engine && go run ./cmd/specs format --check

vet:
	cd engine && go vet ./...

test:
	cd engine && go test ./...

check: format-check lint vet test

build: build-engine build-extension

build-engine:
	cd engine/cmd/specs && go build -o ../../../specs

build-extension:
	cd extension && pnpm run compile

package-extension:
	cd extension && pnpm run package:bundled

deploy-dev: build-engine build-extension
	cd extension && pnpm run symlink
