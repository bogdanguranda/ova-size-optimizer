# Makefile for analyzing a multi-image OCI archive
# Dependencies: gunzip, tar, jq, syft, skopeo

COMPRESSED_MULTI_ARCHIVE ?= test-multi-archive.tar.gz

TEMP_DIR = temp-oci
MULTI_ARCHIVE = $(TEMP_DIR)/$(shell basename $(COMPRESSED_MULTI_ARCHIVE))
MULTI_ARCHIVE_EXTRACTED_DIR = $(TEMP_DIR)/multi-archive-extracted
INDIVIDUAL_ARCHIVES_DIR = $(TEMP_DIR)/individual-archives

# Default target for running checks followed by cleanup
.PHONY: all
all:
	@echo "Starting all steps... "
	@-$(MAKE) all-no-clean
	@$(MAKE) clean

.PHONY: all-no-clean
all-no-clean: prepare generate analyze

.PHONY: prepare
prepare:
	@mkdir -p $(TEMP_DIR) $(MULTI_ARCHIVE_EXTRACTED_DIR) $(INDIVIDUAL_ARCHIVES_DIR)

	@echo "Decompressing $(COMPRESSED_MULTI_ARCHIVE) into $(MULTI_ARCHIVE)"
	@gunzip -c $(COMPRESSED_MULTI_ARCHIVE) > $(MULTI_ARCHIVE)

	@echo "Extracting archive $(MULTI_ARCHIVE) into $(MULTI_ARCHIVE_EXTRACTED_DIR)"
	@tar -xf $(MULTI_ARCHIVE) -C $(MULTI_ARCHIVE_EXTRACTED_DIR)

	@echo "Searching for images in the multi-image archive..."
	@jq -r '.manifests[] | @base64' $(MULTI_ARCHIVE_EXTRACTED_DIR)/index.json > $(MULTI_ARCHIVE_EXTRACTED_DIR)/manifests.txt

	@cat $(MULTI_ARCHIVE_EXTRACTED_DIR)/manifests.txt | while read manifest; do \
		manifest_json=$$(echo $$manifest | base64 --decode); \
		image_name=$$(echo $$manifest_json | jq -r '.annotations["io.containerd.image.name"]' | sed 's|.*/\([^:]*\).*|\1|'); \
		image_ref=$$(echo $$manifest_json | jq -r '.annotations["org.opencontainers.image.ref.name"]'); \
		image_archive_src=$(MULTI_ARCHIVE):$${image_ref}; \
		image_archive_dest=$(INDIVIDUAL_ARCHIVES_DIR)/$${image_name}-$${image_ref}.tar; \
		echo "Extracting individual image from source $${image_archive_src} into $${image_archive_dest}"; \
		skopeo copy oci-archive:$${image_archive_src} docker-archive:$${image_archive_dest}; \
	done

.PHONY: generate
generate:
	@echo "Generating json files for images... "
	@for tar_file in $(INDIVIDUAL_ARCHIVES_DIR)/*.tar; do \
		echo "Processing $$tar_file"; \
		syft docker-archive:$$tar_file -o github-json > $(INDIVIDUAL_ARCHIVES_DIR)/syft_github_json_$$(basename $$tar_file .tar).json; \
		syft docker-archive:$$tar_file -o json > $(INDIVIDUAL_ARCHIVES_DIR)/syft_json_$$(basename $$tar_file .tar).json; \
	done

.PHONY: analyze
analyze:
	@echo "Running oci-analyzer..."
	@$(eval SYFT_GITHUB_JSON_FILES := $(shell ls $(INDIVIDUAL_ARCHIVES_DIR)/syft_github_json_*.json))
	@$(eval SYFT_JSON_FILES := $(shell ls $(INDIVIDUAL_ARCHIVES_DIR)/syft_json_*.json))
	@go run oci-analyzer.go \
		--syft-github-json-files "$(SYFT_GITHUB_JSON_FILES)" \
		--syft-json-files "$(SYFT_JSON_FILES)"

.PHONY: clean
clean:
	@echo "Cleaning up all temporary files..."
	@rm -rf $(TEMP_DIR)
