---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "livebox Provider"
subcategory: ""
description: |-
  A terraform provider to interact with a Livebox. It currently only supports configuring port forwarding rules.
---

# livebox Provider

A terraform provider to interact with a Livebox. It currently only supports configuring port forwarding rules.

## Example Usage

```terraform
provider "livebox" {
  host = "https://192.168.1.1"
  password = "some-password"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `host` (String) URI exposing the Livebox API. May also be provided via LIVEBOX_HOST environment variable.
- `password` (String, Sensitive) Password for accessing the Livebox API. May also be provided via LIVEBOX_PASSWORD environment variable.
