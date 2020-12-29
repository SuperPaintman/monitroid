NIX_FILES := $(shell find -type f -name '*.nix')

all:

.PHONY: build-nix
build-nix-monitroid:
	@nix-build \
		--out-link result-monitroid \
		-E 'with import <nixpkgs> {}; callPackage ./. {}' \
		-A monitroid

.PHONY: format
format-nix:
	@nixpkgs-fmt $(NIX_FILES)

.PHONY: clean
clean:
	@rm -f result
	@rm -f result-*
