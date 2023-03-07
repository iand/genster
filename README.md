# genster

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/iand/genster)
[![Check Status](https://github.com/iand/genster/actions/workflows/check.yml/badge.svg)](https://github.com/iand/genster/actions/workflows/check.yml)
[![Test Status](https://github.com/iand/genster/actions/workflows/test.yml/badge.svg)](https://github.com/iand/genster/actions/workflows/test.yml)

## Usage

`genster` is an application for building browseable family history sites from GEDCOM files.
It generates markdown documents that cam be processed with a site generator such as Jekyll or Hugo.

I created it for myself so it makes many assumptions about my particular style of using GEDCOM.
It may or may not work for other people.

Notably:

 - I use Ancestry extensively and use this as the primary source of my GEDCOMs. Because of this genster includes various idioms that are peculiar to Ancestry.
 - I use Ancestry custom events to convey extra information. These appear in the exported GEDCOM as a general `EVEN` with the "fact label" of the event becoming its type. 


## Getting Started

As of Go 1.19, install the latest genster executable using:

	go install github.com/iand/genster@latest

This will download and build a binary in $GOBIN.


## Conventions

Some custom event "fact labels" with specific handling:

 - `Nickname` - holds the preferred nickname for a person (unfortunately alternate names and AKAs are very messy in Ancestry and I wanted a single value to refer to the person)
 - `OLB` - one line bio, a short sentence that summarises the person's life

Other examples of custom event "fact labels" that are used verbatim in the generated pages:

 - `Admitted to Shipmeadow Workhouse`
 - `Discharged to Shipmeadow Workhouse`
 - `Enlisted in Royal Fusiliers`
 - `Discharged from Royal Fusiliers`
 - `Posted to Malta` - use when assigned or sent to a new post
 - `Posted Home`
 - `Stationed at Woolwich`
 - `Promoted to Corporal Wheeler`
 - `Awarded Queen's South Africa Medal` - include clasps in details
 - `Missing in action`
 - `Injured in action`

Events with values that start with one of these phrases will be included verbatim in the generated narrative:

 - He was recorded as
 - She was recorded as
 - It was recorded that
 - George is recorded 
 - George was recorded 


## License

This is free and unencumbered software released into the public domain. For more
information, see <http://unlicense.org/> or the accompanying [`UNLICENSE`](UNLICENSE) file.
