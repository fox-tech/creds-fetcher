BIN_NAME="foxta"

test:
	@echo "===> Testing"
	go test -race -count=1 -coverprofile=coverage.txt -covermode=atomic ./...

install:
	@echo "===> Installing"
	go build -o $(BIN_NAME)
	mv ./$(BIN_NAME) /usr/local/bin 

uninstall:
	rm /usr/local/bin/$(BIN_NAME)
