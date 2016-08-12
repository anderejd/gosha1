# gosha1
Concurrent SHA-1 checksum calculator for file trees.

* Outputs **sha1sum compatible format** (sha1sum --check FILE).
* Walks the input directory and all subdirs.
* One goroutine worker per logical core.
* Prints checksums to stdout.
* Prints stats to stderr.


