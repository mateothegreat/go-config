test:
	go test -v ./...

test/watch:
	find . -name '*.go' | entr -cr $(MAKE) test