miia
====

[![Build Status](https://travis-ci.org/DarinM223/miia.svg?branch=master)](https://travis-ci.org/DarinM223/miia)

![thumb](/miia.png "Miia")

Lisp-style programming language that tries to exploit concurrency and parallelism
easier by compiling into a graph of goroutines because goroutines can be run in parallel
but also be blocked with less memory usage than traditional threads. The idea for this
project is that although it may not be as performant as a traditional language at
non-concurrent work, if the majority of work is done by sending lots of asynchronous calls
that may take a long time other nodes might be able to start processing much sooner,
since every node is only waiting for its dependencies and is not limited to sequential execution.

This project is still work in progress and because there are no benchmarks yet so it is unknown
how performant this idea would be. This project is mostly to test the limits for green threads
and to learn more Go programming.

## Compiling
First you have to run `go get` inside the project directory in order to download the dependencies.
(right now the only dependency is goquery for extracting data from css selectors)

After that you can build the project by running `go build`.

## Running
When runnning the executable, you have to specify the path to the file to compile and run.

For example, running:
```
./miia test/simple.scrape
```
in the project directory should run a simple program for the language.

## Testing
In order to run the unit tests in all the packages, run `go test ./...` in the project directory.

## TODO
- ~~Add collect node that collects a from a stream of messages.~~
- ~~Add array literal support.~~
- Add more options for selector nodes (right now it only parses the text of the dom element with the selector).
- Add more options for goto nodes (allow for GET and POST and custom auth types).
- ~~Determine exactly how many for loop subnodes to run in parallel (right now if a for loop contains a nested for loop
it only runs each subnode one at a time).~~
- Add benchmarks.
