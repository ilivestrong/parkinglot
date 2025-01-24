input:=

build:
	go build -o bin/parking_lot main.go
run:
	go run main.go $(input)
test:
	go test -v -count=1  ./... 