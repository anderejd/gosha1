# gosha1
Concurrent SHA-1 checksum calculator for file trees.

* Outputs **_sha1sum compatible format_** (sha1sum --check FILE).
* Walks the input directory and all subdirs.
* Limits the number of worker goroutines to os.NumCPU().
* Prints checksums to stdout.
* Prints stats to stderr.
* Returns 0 on success.


