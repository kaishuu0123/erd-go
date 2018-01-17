# erd-go

Translates a plain text description of a relational database schema to a graphical entity-relationship diagram.(convert to dot file)

![ER diagram for nfldb](https://github.com/kaishuu0123/erd-go/blob/master/examples/outputs/nfldb.png)

## Install

```
go get github.com/kaishuu0123/erd-go
```

or get binary.

## Usage

```
Usage:
  erd-go [OPTIONS] PATTERN [PATH]

Application Options:
  -i, --input=  input will be read from the given file.
  -o, --output= output will be written to the given file.

Help Options:
  -h, --help    Show this help message
```

support input from stdin.

```
cat examples/nfldb.er | erd-go
```

ex.) convert to png from dot (use dot command)

```
cat examples/nfldb.er | erd-go | dot -Tpng -o nfldb.png
```

## Example

see [examples directory](https://github.com/kaishuu0123/erd-go/blob/master/examples)

## Build Instruction

1. install glide
```
go get github.com/Masterminds/glide
```
1. install go-bindata
```
go get github.com/jteeuwen/go-bindata
```
1. make
```
make
```

## LICENSE

MIT

## Credits
This work is based off of several existing projects:
* https://github.com/BurntSushi/erd
* https://github.com/unok/erdm

