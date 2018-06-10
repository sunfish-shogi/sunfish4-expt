INST_PATH:=$(GOPATH)/bin/hyper-param-tuning
BIN:=bin/hyper-param-tuning
PKG:=github.com/sunfish-shogi/sunfish4-expt/tools/hyper-param-tuning

.PHONY: help
help:
	@echo "USAGE:"
	@echo "  make build"
	@echo "  make install"
	@echo "  make vet"
	@echo "  make test"

.PHONY: install
install: $(INST_PATH)

.PHONY: build
build: $(BIN)

.PHONY: vet
vet:
	go vet $(PKG)/...

.PHONY: test
test:
	go test $(PKG)/...

$(INST_PATH): $(BIN)
	cp $(BIN) $(INST_PATH)

.PHONY: $(BIN)
$(BIN):
	go build -o $(BIN) $(PKG)

.PHONY: $(CLEAN)
$(CLEAN):
	$(RM) $(BIN)
