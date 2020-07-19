# erd-go

[![GitHub release](https://img.shields.io/github/release/kaishuu0123/erd-go.svg)][releases]
[![Build Status](https://travis-ci.org/kaishuu0123/erd-go.svg?branch=master)](https://travis-ci.org/kaishuu0123/erd-go)
[![Coverage Status](https://coveralls.io/repos/github/kaishuu0123/erd-go/badge.svg)](https://coveralls.io/github/kaishuu0123/erd-go)
[![Docker Pulls](https://img.shields.io/docker/pulls/kaishuu0123/erd-go.svg)][docker]

[releases]: https://github.com/kaishuu0123/erd-go/releases
[docker]: https://hub.docker.com/r/kaishuu0123/erd-go/

Translates a plain text description of a relational database schema to a graphical entity-relationship diagram.(convert to dot file)

![ER diagram for nfldb](https://raw.githubusercontent.com/kaishuu0123/erd-go/master/examples/outputs/nfldb.png)

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

## Usage (Used by Docker container)

```
cat examples/nfldb.er | docker run --rm -i kaishuu0123/erd-go | docker run --rm -i risaacson/graphviz dot -T png > nfldb.png
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

