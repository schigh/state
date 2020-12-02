UNIT_PKGS = $(shell go list ./... | grep -v sandbox | paste -sd "," -)
COVER_DIR = $(CURDIR)/.coverage

.PHONY: test
test: $(COVER_DIR)
	go test -v -race -cover -coverpkg="$(UNIT_PKGS)" -coverprofile="$(COVER_DIR)/unit.out" ./...

$(COVER_DIR):
	@mkdir -p $(COVER_DIR)
