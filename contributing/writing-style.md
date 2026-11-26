# Writing style

Use these rules for schema descriptions and Markdown documentation. Keep the meaning of existing text when you edit it.

## Rules

- Put one fact in each sentence.
- Use active voice. Name the actor.
- Use plain words and short sentences.
- Use one term for each thing. For example, use “bucket” for a B2 bucket.
- State units for numbers, such as seconds or bytes. For timestamps, state the unit and epoch, such as milliseconds since January 1, 1970 UTC.
- Check facts against the code or the B2 documentation. Do not guess.

The `SchemaDescriptionBuilder` in `b2/provider.go` appends schema details to descriptions. Do not repeat its defaults, required-with conditions, conflicts, exactly-one and at-least-one constraints, or force-new notice in a description.

## Protect meaning

Keep attribute and block names, HCL, commands and numbers unchanged. Preserve the strength of “must” for requirements, “should” for recommendations and “can” for abilities. Do not change a requirement into a recommendation.

## Examples from this provider

These descriptions are from merged changes in `b2/templates.go`:

- File size ([#263](https://github.com/Backblaze/terraform-provider-b2/pull/263)): “The file size.” → “The file size in bytes.”
- Upload timestamp ([#263](https://github.com/Backblaze/terraform-provider-b2/pull/263)): “This is a UTC time when this file was uploaded.” → “Upload time in milliseconds since January 1, 1970 UTC.”
- Lifecycle prefix ([#261](https://github.com/Backblaze/terraform-provider-b2/pull/261)): “It specifies which files in the bucket it applies to.” → “The file name prefix matched by this lifecycle rule.”
- Retention duration ([#261](https://github.com/Backblaze/terraform-provider-b2/pull/261)): “Duration” → “Retention duration in the specified unit of days or years.”

## How to check a change

1. Edit the schema description in `b2/` or a template in `templates/`.
2. Run `make docs`. This command needs the Python bindings build.
3. Review the generated pages in `docs/`. Never edit those pages by hand.
