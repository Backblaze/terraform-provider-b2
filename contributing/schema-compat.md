# Schema compatibility check

The schema is the public contract of this provider. A change to it can break existing configurations or plans. The `schema-compat` check compares the schema of your working tree with a base ref and reports breaking changes.

## What it checks

The check builds the provider at the base ref (default `origin/master`) and at your working tree. It dumps each schema with `terraform providers schema -json` through development overrides, as `scripts/examples-check.sh` does. It needs no credentials and downloads nothing.

It reports these changes as BREAKING:

- A removed resource or data source.
- A removed attribute or nested block.
- A changed attribute type.
- An optional attribute that became required.
- A new required attribute or block.
- A sensitive attribute that became non-sensitive.
- A changed nesting mode, a higher minimum item count or a lower maximum item count on a nested block.

It reports added resources, data sources, optional attributes and optional blocks as NON-BREAKING. It lists deprecation changes separately as DEPRECATION.

## What it cannot see

The schema JSON has no field for `ForceNew` or for defaults. These issue rules are therefore not checked:

- A new `ForceNew` on an attribute.
- A removed or changed default.

Review these by hand. To cover them later, you could add a unit test that reads the provider schema in Go (`schema.Schema.ForceNew` and `schema.Schema.Default`) and compares it with a committed snapshot. This check does not do that.

## How to run it

```sh
make schema-compat
make schema-compat SCHEMA_COMPAT_ARGS="--base origin/some-branch --warn-only"
```

The command exits with status 1 when it finds BREAKING changes. The `--warn-only` option reports them and exits with status 0. The `--base` option sets the base ref.

The command needs Go and Terraform. To test the comparison logic alone, run `scripts/schema-compat_test.sh`. That test needs only Python 3.

A BREAKING change needs a `Changed` or `Removed` changelog entry.
