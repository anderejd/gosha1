# gosha1
Concurrent, recursive file checksum calculator.

* Outputs sha1sum compatible format (sha1sum --check FILE)
* One goroutine worker per logical core.
* Prints checksums to stdout
* Prints stats to stderr


