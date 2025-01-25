file:=

build:
	go build -o bin/parking_lot main.go
run:
	go run main.go $(file)
test:
	go test -v -count=1  ./... 