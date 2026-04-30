.PHONY: lint format format-check check build build-cli build-extension package-extension deploy-dev

lint:
	cd cli && go run ./cmd/specs lint --style

format:
	cd cli && go run ./cmd/specs format

format-check:
	cd cli && go run ./cmd/specs format --check

check: format-check lint

build: build-cli build-extension

build-cli:
	cd cli/cmd/specs && go build -o ../../../specs

build-extension:
	cd extension && pnpm run compile

package-extension:
	cd extension && pnpm run package:bundled

deploy-dev: build-cli build-extension
	cd extension && pnpm run symlink
