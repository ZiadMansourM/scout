GOOS_LIST = linux darwin windows
GOARCH_LIST = amd64 arm64
PROJECT_NAME = scout
VERSION = v0.0.1
BUILD_DIR = output
BINARIES_DIR = $(BUILD_DIR)/binaries
RELEASE_DIR = $(BUILD_DIR)/release

all: clean build release

build:
	@mkdir -p $(BINARIES_DIR)
	@for GOOS in $(GOOS_LIST); do \
		for GOARCH in $(GOARCH_LIST); do \
			EXT=""; \
			if [ "$$GOOS" = "windows" ]; then EXT=".exe"; fi; \
			FILENAME=$(PROJECT_NAME)_$${GOOS}_$${GOARCH}$$EXT; \
			echo "Building $$FILENAME"; \
			GOOS=$$GOOS GOARCH=$$GOARCH go build -o $(BINARIES_DIR)/$$FILENAME -ldflags="-X main.version=$(VERSION)" main.go; \
		done; \
	done

release:
	@mkdir -p $(RELEASE_DIR)
	@cd $(BINARIES_DIR) && for FILE in *; do \
		FILENAME=$${FILE}.tgz; \
		echo "Creating release $$FILENAME"; \
		tar -czvf ../release/$$FILENAME $$FILE; \
	done

clean:
	@echo "Cleaning the entire $(BUILD_DIR) directory..."
	@rm -rf $(BUILD_DIR)

