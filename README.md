[![Build Status](https://travis-ci.org/caeost/gosnap.svg?branch=master)](https://travis-ci.org/caeost/gosnap)

# Go Snap

This is a very basic pluggable site generator I wrote in go to learn a little bit about the language. It's based very heavily off of the functionality of [metalsmith](http://metalsmith.io) since I think it hit a nice minimal level of necessary functionality.

I simplified the structure a bit based off of what I found using metalsmith on a previous project: `metadata` is gone since I think it's not that useful, also the asynchronicity helping hand of metalsmith is gone since go and javascript are different languages...

## Example

An example exists in the creatively named `/example` folder which can be run and will do very minimal work to the set of files defined in `/example/source` to move them into `/example/destination`. 

## Go

Go seems to have good potential for this kind of build system (although I suspect it has less library support) since it has a strong concurrency model and a focus on speed. It also doesn't lose too many of javascript's strengths since it treats functions as first class citizens and isn't too verbose. 

## To do

* Actually useful plugins
* Benchmarking
