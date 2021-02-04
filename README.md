# erd-go

This is forked from <https://github.com/kaishuu0123/erd-go> and added following:

1. Support `?` as `{0,1}` in relations
1. Added `colors` segment to manage color palette see: [simple example](https://github.com/ray-g/erd-go/blob/master/examples/simple.er#L8-L23)
1. Changed to use <github.com/kevinburke/go-bindata> since the original `bin-data` is not maintained
1. Add support `label` to name
1. Add support special characters in table name, such as `[Users.Profile]`
1. Use horizontal layout for isolated nodes

install this `go-bindata` by `go get -u -v github.com/kevinburke/go-bindata/...`

Get binary from this [releases page](https://github.com/ray-g/erd-go/releases).  [![Build Status](https://travis-ci.org/ray-g/erd-go.svg?branch=master)](https://travis-ci.org/ray-g/erd-go)

## ===

[![Build Status](https://travis-ci.org/kaishuu0123/erd-go.svg?branch=master)](https://travis-ci.org/kaishuu0123/erd-go)
[![Coverage Status](https://coveralls.io/repos/github/kaishuu0123/erd-go/badge.svg)](https://coveralls.io/github/kaishuu0123/erd-go)

Translates a plain text description of a relational database schema to a graphical entity-relationship diagram.(convert to dot file)

![ER diagram for nfldb](https://github.com/kaishuu0123/erd-go/blob/master/examples/outputs/nfldb.png)

## Install

get binary from [releases page](https://github.com/kaishuu0123/erd-go/releases).

or

```
go get github.com/kaishuu0123/erd-go
```

or (for Mac)

```
brew tap kaishuu0123/erd-go
brew install erd-go
```

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

support input from STDIN.

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
1. install peg
    ```
    go get github.com/pointlander/peg
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

