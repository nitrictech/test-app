
build:
	go build ./functions/controller
	go build ./functions/store
	go build ./functions/worker

test:
	go run github.com/onsi/ginkgo/ginkgo ./tests/...

clean:
	rm -f controller store worker
