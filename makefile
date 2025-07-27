build:
	go build -o bin/go-validate cmd/main.go

link:
	rm -f /usr/local/bin/go-validate
	ln -s $(PWD)/bin/go-validate /usr/local/bin/go-validate

test:
	richgo test -v ./...

test/watch:
	find . -name '*.go' | entr -c $(MAKE) test