

build:
	go build -o bin/history ./functions/history
	go build -o bin/store ./functions/store
	go build -o bin/worker ./functions/worker