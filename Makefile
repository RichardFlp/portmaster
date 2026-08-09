BINARY  := portmaster
VERSION ?= 0.1.0
MODULE  := github.com/RichardFlp/portmaster
LDFLAGS := -s -w -X $(MODULE)/internal/version.Version=$(VERSION)

.PHONY: build test vet lint fmt clean release

build:
	mkdir -p bin
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/portmaster

test:
	go test ./...

vet:
	go vet ./...

lint: vet
	@files=$$(gofmt -l .); \
	if [ -n "$$files" ]; then \
		echo "gofmt required on:"; \
		echo "$$files"; \
		exit 1; \
	fi

fmt:
	gofmt -w .

clean:
	rm -rf bin dist coverage

release:
	mkdir -p dist
	@for target in "windows amd64 exe" "windows arm64 exe" "linux amd64" "linux arm64" "darwin amd64" "darwin arm64"; do \
		set -- $$target; \
		os=$$1; arch=$$2; ext=$$3; \
		name=$(BINARY)-$$os-$$arch; \
		if [ -n "$$ext" ]; then \
			name=$$name.$$ext; \
		fi; \
		echo "building $$name"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/$$name ./cmd/portmaster; \
	done
	cd dist && sha256sum $(BINARY)-* > SHA256SUMS.txt
