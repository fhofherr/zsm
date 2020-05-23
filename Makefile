.PHONY: release
release:
	goreleaser --rm-dist

.PHONY: release-snapshot
release-snapshot:
	goreleaser --rm-dist --snapshot

.PHONY: test
test:
	./scripts/test-coverage.sh

.PHONY: clean
clean:
	rm -rf dist
