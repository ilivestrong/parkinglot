file:=

run:
	go run main.go $(file)
test:
	go test -v -count=1  ./... 