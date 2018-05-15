# k8s-ctx-import
[![Build Status](https://travis-ci.org/chrischdi/k8s-ctx-import.svg?branch=master)](https://travis-ci.org/chrischdi/k8s-ctx-import)

`k8s-ctx-import` is an utility to merge kubernetes contexts to a single kubeconfig.

```
$ k8s-ctx-import -h
`k8s-ctx-import` is an utility to merge kubernetes contexts to a single kubeconfig.
It imports the context either to `~/.kube/config` or to the file defined by the `KUBECONFIG` environment variable.
Usage of k8s-ctx-import:
  -force
        force import of context
  -help
        display this help and exit
  -name string
        renames the context for the import
Example:
    cat /some/kubeconfig | k8s-ctx-import
```

## Install pre-compiled version

* Download a pre-compiled version from the [releases](https://github.com/chrischdi/k8s-ctx-import/releases) page
* Unpack tar.gz
* Make sure it is executable
* Move the binary into `$PATH`

## Install from source

Install or update from current master:
```
go get -u github.com/chrischdi/k8s-ctx-import
```

## Contribute

Feel free to clone or fork this repo to start contributing.
