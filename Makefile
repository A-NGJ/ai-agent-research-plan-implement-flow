.PHONY: build test clean install uninstall

build:
	go build -o bin/rpi ./cmd/rpi

test:
	go test ./...

clean:
	rm -f bin/rpi

INSTALL_DIR ?= $(HOME)/.local/bin

install: build
	mkdir -p $(INSTALL_DIR)
	rm -f $(INSTALL_DIR)/rpi
	cp $(CURDIR)/bin/rpi $(INSTALL_DIR)/rpi
	@echo "Installed rpi to $(INSTALL_DIR)/"
	@echo "Make sure $(INSTALL_DIR) is in your PATH"

uninstall:
	rm -f $(INSTALL_DIR)/rpi
	@echo "Removed rpi from $(INSTALL_DIR)/"
