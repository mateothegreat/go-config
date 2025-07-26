test:
	richgo test -v ./...

test/watch:
	find . -name '*.go' | entr -c $(MAKE) test