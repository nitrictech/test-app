
build:
	go build ./functions/store
	go build ./functions/worker

test:
	go test -v ./tests/...

clean:
	rm -f controller store worker
