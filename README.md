Terraform Provider B2
=====================

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.12.x
-	[Go](https://golang.org/doc/install) >= 1.12
-	[Python](https://github.com/pyenv/pyenv) >= 3.9

Building
--------

1. Build python bindings:
    ```sh
    $ cd python-bindings
    $ pip install --user nox
    $ nox -s bundle
    ```
1. Build the provider: 
    ```sh
    $ cd ..
    $ make install
    ```

Testing
-------

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

Adding Dependencies
-------------------

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.


