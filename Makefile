# Makefile for running syft-analyzer checks

# List of image IDs (can be overridden from the command line)
IMAGE_IDS ?= 313b71942cfa
# ef0c1c2f2e79  cp-schmea
# bogdan: e8a9ee02cba6 2e1c800b7bd7

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
		syft $$image_id -o github-json > syft_github_json_$$image_id.json; \
		echo "Generating Syft JSON for image $$image_id..."; \
		syft $$image_id -o json > temp_syft; \
		cat temp_syft | jq > syft_json_$$image_id.json; \
		rm -f temp_syft; \
		echo "Generating Dive json for image $$image_id..."; \
		dive $$image_id -j dive_$$image_id.json; \
	done

.PHONY: analyze
analyze:
	echo "Running oci-analyzer..."
	$(eval SYFT_GITHUB_JSON_FILES := $(shell ls syft_github_json_*.json))
	$(eval SYFT_JSON_FILES := $(shell ls syft_json_*.json))
	$(eval DIVE_FILES := $(shell ls dive_*.json))
	go run oci-analyzer.go \
		--syft-github-json-files "$(SYFT_GITHUB_JSON_FILES)" \
		--syft-json-files "$(SYFT_JSON_FILES)" \
		--dive-files "$(DIVE_FILES)"

.PHONY: clean
clean:
	echo "Cleaning up json files..."
	rm -f syft_*.json
	rm -f dive_*.json
