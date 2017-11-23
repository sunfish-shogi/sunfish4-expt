PKG:=github.com/sunfish-shogi/sunfish4-expt

.PHONY: help
help:
	@echo "USAGE:"
	@echo "  make vet"
	@echo "  make test"

.PHONY: vet
vet:
	go vet $(PKG)/...

.PHONY: test
test:
	go test $(PKG)/...
