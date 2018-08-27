# v0.4.0
* refactor: split out functions readFile readKubeconfig and mergeKubeconfig for better testing possibilities
* tests: write unit tests readfile, readKubeconfig and mergeKubeconfig 
* feature: add parameter `-stdout` to print the output to stdout instead of writing to file
* bug: print help message on -h no matter of position of the flag

# v0.3.1
* bug: fix index out of range when no flag is given

# v0.3.0
* feature: add parameter `-set-current-context=false` to be able to toggle the context overwrite ([#1](https://github.com/chrischdi/k8s-ctx-import/pull/1))
* bug: Fix `-help` and `-h` to exit and print usage

# v0.2

* bug: Fix `for`-loops to use break. Otherwise the last item in loop would have been imported.
* bug: Fix value assignment
* bug: Fix correct use of flagset
