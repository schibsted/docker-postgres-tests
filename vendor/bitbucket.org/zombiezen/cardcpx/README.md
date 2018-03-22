# cardcpx

[![Build Status](https://travis-ci.org/zombiezen/cardcpx.svg?branch=master)](https://travis-ci.org/zombiezen/cardcpx)

cardcpx is a specialized UI for copying files from a camera card to 1+ replicas.  The replica
copies happen concurrently, so if you are copying N bytes to R replicas, the time is O(N) instead of
O(N * R).

The interface also has includes simple scene/take ingestion, which is stored in a CSV transaction
log.  Selects will be copied first, so you can do a proofing check on a fast disk while your import
finishes.

cardcpx supports a flat directory structure as well as the RED camera directory structure.  It
assumes that your clip names do not overlap.  Attempting to copy the same file name will not
overwrite data.

## Building

You must have GNU Make, Go 1.2, and Ant installed.  All other dependencies are included in the
`third_party` directory.  To build:

    make

This produces `cardcpx` and `ui/js.js`.

## Running

From the project directory, run:

    ./cardcpx -storageDirs=/path/to/replica1:/path/to/replica2 -takeFile=path/to/takes.csv

then open http://localhost:8080/ in your browser of choice.  The UI is restricted to localhost-only
by default.  For more details on options, run `./cardcpx -help`.

The replicas given by `-storageDirs` are the directories that your clips will be copied to.
`-takeFile` specifies the CSV file that will hold all the take information (i.e. take, scene, clip
name, select).

## License

Apache License 2.0.  See LICENSE for details.

This is not an official Google product (experimental or otherwise), it is just code that happens
to be owned by Google.
