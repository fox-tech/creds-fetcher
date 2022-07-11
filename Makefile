test:
	@echo "===> Testing"
	go test -race -count=1 -coverprofile=coverage.txt -covermode=atomic ./...

install:
	@echo "===> Installing"
	go install

uninstall:
	@echo "===> Uninstalling"
	rm "$(GOPATH)/bin/creds-fetcher"
