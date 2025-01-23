build:
	go build -o bin/parking_lot main.go
run:
	go run main.go
test:
	go test -v -count=1  ./... 