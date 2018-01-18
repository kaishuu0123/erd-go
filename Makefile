all: build

build: depend peg bindata
	go build

peg:
	peg erd.peg

run: build
	cat examples/nfldb.er | ./erd-go -o nfldb.dot

depend:
	glide install

bindata:
	go-bindata -o=templates_bindata.go ./templates/...
