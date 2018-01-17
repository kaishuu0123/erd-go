all: build

build: depend bindata
	go build

run: depend
	peg erd.peg
	cat examples/nfldb.er | go run *.go -o nfldb.dot

depend:
	glide install

bindata:
	go-bindata -o=templates_bindata.go ./templates/...
