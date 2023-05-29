# env-sample-sync

Automatically keep `.env` files in sync with `env.sample`

---

`env-sample-sync` checks whether the local repository contains an `.env` file (configurable), scrubs it of secrets/values, and makes the scrubbed version available as `env.sample` (configurable).

This process can be run manually, or automatically as a git-hook, ensuring that all application environment variables are safely and automatically documented without leaking secrets.

Crucially, `env-sample-sync` allows comments in `env` files, which are carried over to `env.sample`. This lets developers add thorough environment variable documentation to source control.

# Installation & Usage

`env-sample-sync` can be run in three ways:

1. Manually
2. As a native git-hook
3. As a [pre-commit plugin](https://pre-commit.com/#install) (a utility for organizing pre-commit hooks)

## Installation

Installation is required to run `env-sample-sync` manually, or as a native git hook. See [pre-commit configuration](#runing-as-a-pre-commit-plugin) for pre-commit usage.


**Install from the releases page**

Download [the latest release](https://github.com/acaloiaro/env-sample-sync/releases/latest) and add it to your `$PATH`.

**Install with `go install`**

```bash
go install github.com/acaloiaro/env-sample-sync@latest
```

## Running manually

`env-sample-sync` can be run with no arguments to use defaults.

**Sync `.env` with `env.sample`**

```bash
env-sample-sync
```


**With a non-default `.env` and `env.sample` location**

```bash
env-sample-sync --env-file=.env-file-name --sample-file=env-sample-file-name
```

**Provide example values for variables**

By default, `env-sample-sync` uses the name of environment variables in `<brackets>` as example values in `env.sample`, e.g. `FOO=secret value` is replaced with `FOO=<FOO>`. This behavior is customizable wit the `--example` flag.

Add custom examples for variables `FOO` and `BAR`.

```bash
env-sample-sync --example=FOO="must be a valid UUID" --example=BAR="bars must be a positive integer"
```

The above invocation yields the following `env.sample`


```bash
FOO=must be a valid UUID

BAR=bars must be a positive integer
```

## Running as a native git-hook

To add `env-sample-sync` as a pre-commit git hook in a git repository, run:

```bash
env-sample-sync install
```

This installs `env-sample-sync` as a pre-commit git hook with default arguments.

The `install` command supports all [command flags](#command-flags).

If you need to change `env-sample-sync` flags, simply run `env-sample-sync install` again with the desired flags.

## Running as a pre-commit plugin

This utility can be used as a [pre-commit plugin](https://pre-commit.com/#install)

## Add configuration
```bash
cat <<EOF > .pre-commit-config.yaml
repos:
-   repo: https://github.com/acaloiaro/env-sample-sync.git
    rev: v2.1.0
    hooks:
      - id: env-sample-sync
EOF
pre-commit install
git add .pre-commit-config.yaml
```

See [pre-commit configuration examples](#pre-commit-configuration-examples) for additional pre-commit documentation.

# Background

It's important to document the environment variables required to run applications, both in production and development. A great way to do so is with `env.sample` files, but sample files tend to get out of date very quickly.

For example, let's say you're adding a new feature to your application and it requires the variable `FOO` to be set. While you're developing locally, you likely have a `.env` file that looks something like:

```bash
APPLICATION_SECRET=supersekrit

# I got this FOO from the detailed process documented at: http:://wiki.example.com/how_to_get_a_foo
FOO="My super secret value for foo"
```

Working on large teams, it's common to share these `.env` files somewhere secure where all developers have access to them. Retrieval is often integrated into application startup and/or bootstrap processes.

But working on open source projects or teams with less trust and less shared infrastructure, it's more common to share an `env.sample`. `env-sample-sync` automatically keeps the sample file in sync with `.env`, so you don't have to. Your `.env` file stays the same, and is automatically converted to the following `env.sample`:

```bash
APPLICATION_SECRET=<APPLICATION_SECRET>

# I got this FOO from the detailed process documented at: http:://wiki.example.com/how_to_get_a_foo
FOO=<FOO>
```

It's even possible to provide default/example values for every environment variables with the `--example` [flag](#command-flags).

## Command Flags

| Name                  | Description                                         | Example                                                   | Default                       |
| --------------------  | --------------------------------------------------- | --------------------------------------------------------- | ----------------------------- |
| `-e`/`--env-file`     | The name of the environment file                    | `--env-file=.secrets`                                     | `--env-file=.env`             |
| `-s`/`--sample-file`  | The name of the sample environment file             | `--sample-file=secrets.example`                           | `--sample-file=env.sample`    |
| `-x`/`--example`      | Provide examples for specific environment variables | `--example=FOO="Example FOO" --example=BAR="Example BAR"` | `--example=VAR=<VAR>`    |

## Pre-commit Configuration Examples

### Default configuration

```yml
repos:
-   repo: https://github.com/acaloiaro/env-sample-sync.git
    rev: v2.1.0
    hooks:
      - id: env-sample-sync
```

### Customize `.env` and `env.sample` paths

```yml
repos:
-   repo: https://github.com/acaloiaro/env-sample-sync.git
    rev: v2.1.0
    hooks:
      - id: env-sample-sync
        args: ['--env-file=.env_file', '--sample-file=env_file.sample']
```

### Customize variable example values

Sometimes environment variables need to conform to specific formats and it's necessary to provide better documentation. For this reason, environment variable examples may be provided in lieu of the default behavior, which is to use the environment variable name surrounded by `<brackets like this>` in sample files.

```yml
repos:
-   repo: https://github.com/acaloiaro/env-sample-sync.git
    rev: v2.1.0
    hooks:
      - id: env-sample-sync
        args: [--example=FOO="Provide your foo here", --example=BAR="You can fetch bars from https://example.com/bars"]
```

Example env file
`.env`
```
FOO=the_value_of_my_secret_foo
BAR=the_value_of_my_secret_bar
```

Example sample file output
`env.sample`
```bash
FOO=Provide your foo here
BAR=You can fetch bars from https://example.com/bars
```

# Limitations

While this project does not directly support [Direnv](https://direnv.net/)-style `.envrc` files, direnv users are free to use its [`dotenv` std lib function](https://direnv.net/man/direnv-stdlib.1.html#codedotenv-ltdotenvpathgtcode) with this utility.

