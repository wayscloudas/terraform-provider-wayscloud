---
page_title: "wayscloud_app Resource - WAYSCloud"
description: |-
  Manages a container app in WAYSCloud App Platform.
---

# wayscloud_app (Resource)

Manages a container app in WAYSCloud App Platform. Apps support automatic deployments from Docker images or Git repositories.

## Example Usage

### From Docker Image

```terraform
resource "wayscloud_app" "api" {
  name   = "my-api"
  slug   = "my-api"
  region = "no"
  plan   = "app-starter"
  image  = "nginx:latest"

  env_vars = {
    NODE_ENV = "production"
  }

  instances     = 1
  scale_to_zero = true
}
```

### From Git Repository

```terraform
resource "wayscloud_app" "frontend" {
  name       = "my-frontend"
  slug       = "my-frontend"
  region     = "no"
  plan       = "app-starter"
  git_repo   = "https://github.com/myorg/frontend.git"
  git_branch = "main"
}
```

## Schema

### Required

- `name` (String) - App display name.
- `slug` (String) - URL slug (lowercase, hyphens). Changing this forces a new resource.
- `region` (String) - Region code. Changing this forces a new resource.
- `plan` (String) - Plan code.

### Optional

- `image` (String) - Docker image to deploy.
- `git_repo` (String) - Git repository URL.
- `git_branch` (String) - Git branch. Default: "main".
- `env_vars` (Map of String) - Environment variables.
- `instances` (Number) - Number of instances. Default: 1.
- `scale_to_zero` (Boolean) - Scale to zero when idle. Default: false.

### Read-Only

- `id` (String) - The app ID (ULID format).
- `url` (String) - App URL.
- `status` (String) - App status.
- `created_at` (String) - Timestamp when created.

## Import

Apps can be imported using the app ID:

```bash
terraform import wayscloud_app.api app_01ARZ3NDEKTSV4RRFFQ69G5FAV
```
