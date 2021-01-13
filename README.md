Terraform Provider B2
=====================

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.15
-	[Python](https://github.com/pyenv/pyenv) >= 3.9

Build
-----

```
make build
```

Install
-------

```
make install
```

Testing
-------

*Note:* Acceptance tests create real resources, and often cost money to run.

```
export B2_APPLICATION_KEY=your_app_key
export B2_APPLICATION_KEY_ID=your_app_key_id
make testacc
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


