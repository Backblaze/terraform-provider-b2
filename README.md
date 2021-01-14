Terraform Provider B2
=====================

Terraform provider for Backblaze B2.

The provider is written in go, but it uses official B2 python SDK embedded into the binary.

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.15
-	[Python](https://github.com/pyenv/pyenv) >= 3.9

Dependencies
------------
*Note:* You can run it inside python virtualenv as it installs the dependencies for the python bindings as well.

```
make deps
```

Building
--------

```
make build
```

Installing
----------

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

Release History
-----------------

Please refer to the [changelog](CHANGELOG.md).
