---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "b2_application_key Resource - terraform-provider-b2"
subcategory: ""
description: |-
  B2 application key resource.
---

# Resource `b2_application_key`

B2 application key resource.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **capabilities** (Set of String) A list of capabilities.
- **key_name** (String) The name of the key.

### Optional

- **bucket_id** (String) The ID of the bucket.
- **id** (String) The ID of this resource.
- **name_prefix** (String) A prefix to restrict access to files

### Read-only

- **application_key** (String, Sensitive) The key.
- **application_key_id** (String) The ID of the key.
- **options** (Set of String) List of application key options.

