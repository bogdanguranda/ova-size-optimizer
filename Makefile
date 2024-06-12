# Makefile for running syft-analyzer checks

# List of image IDs (can be overridden from the command line)
IMAGE_IDS ?= e8a9ee02cba6 2e1c800b7bd7

# Default target for running checks followed by cleanup
.PHONY: all
all:
	echo "Starting all steps... "
	-$(MAKE) generate
	-$(MAKE) analyze
	$(MAKE) clean

.PHONY: generate
generate:
	echo "Generating syft github-json files for images... "
	for image_id in $(IMAGE_IDS); do \
		echo "Generating Syft GitHub JSON for image $$image_id..."; \
		syft $$image_id -o github-json > syft_$$image_id.json; \
	done

.PHONY: analyze
analyze:
	echo "Running syft-analyzer..."
	go run syft-analyzer.go syft_*.json

.PHONY: clean
clean:
	echo "Cleaning up json files..."
	rm -f syft_*.json
