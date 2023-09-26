# ess (env sample sync)

[![Gitter chat](https://badges.gitter.im/gitterHQ/gitter.png)](https://app.gitter.im/#/room/#env-sample-sync-dev:gitter.im)

Automatically keep `.env` files in sync with `env.sample`

---

`ess` checks whether the local repository contains an `.env` file (configurable), scrubs it of secrets/
values, and makes the scrubbed version available as `env.sample` (configurable).

This process can be run manually, or automatically as a git-hook, ensuring that all application environment variables
are safely and automatically documented without leaking secrets.

Crucially, `ess` allows comments in `env` files, which are carried over to `env.sample`. This lets
developers add thorough environment variable documentation to source control.

# Installation & Usage

`ess` can be run in three ways:

1. Manually
2. As a native git-hook
3. As a [pre-commit plugin](https://pre-commit.com/#install) (a utility for organizing pre-commit hooks)

## Installation

Installation is required to run `ess` manually, or as a native git hook. See [pre-commit configuration]
(#running-as-a-pre-commit-plugin) for pre-commit usage.


**Install from the releases page**

Download [the latest release](https://github.com/acaloiaro/ess/releases/latest) and add it to your `$PATH`.

**Install with `go install`**

```bash
go install github.com/acaloiaro/ess@latest
```

**Nix Flake**

This application may be added as a flake input

**flake.nix**
```nix
inputs.ess = {
  url = "github:acaloiaro/ess";
  inputs.nixpkgs.follows = "nixpkgs";
};
```

**configuration.nix**
```nix
users.users.<USERNAME>.packages = [
  ...
  inputs.ess.packages.${system}.default
];
```

Or simly run directly
```bash
nix run github:acaloiaro/env-sample-sync
```

## Running manually

`ess` can be run with no arguments to use defaults.

**Sync `.env` with `env.sample`**

```bash
ess
```


**With a non-default `.env` and `env.sample` location**

```bash
ess --env-file=.env-file-name --sample-file=env-sample-file-name
```

**Provide example values for variables**

By default, `ess` uses the name of environment variables in `<brackets>` as example values in `env.sample`,
e.g. `FOO=secret value` is replaced with `FOO=<FOO>`. This behavior is customizable wit the `--example` flag.

Add custom examples for variables `FOO` and `BAR`.

```bash
ess --example=FOO="must be a valid UUID" --example=BAR="bars must be a positive integer"
```

The above invocation yields the following `env.sample`


```bash
FOO=must be a valid UUID

BAR=bars must be a positive integer
```

## Running as a native git-hook

To add `ess` as a pre-commit git hook in a git repository, run:

```bash
ess install
```

This installs `ess` as a pre-commit git hook with default arguments.

The `install` command supports all [command flags](#command-flags).

If you need to change `ess` flags, simply run `ess install` again with the desired flags and
choose the overwrite [o] option when prompted what to do with the existing pre-commit hook.

## Running as a pre-commit plugin

This utility can be used as a [pre-commit plugin](https://pre-commit.com/#install)

## Add configuration
```bash
cat <<EOF > .pre-commit-config.yaml
repos:
-   repo: https://github.com/acaloiaro/ess.git
    rev: v2.5.0
    hooks:
      - id: ess
EOF
pre-commit install
git add .pre-commit-config.yaml
```

See [pre-commit configuration examples](#pre-commit-configuration-examples) for additional pre-commit documentation.

# Background

It's important to document the environment variables required to run applications, both in production and development. A
great way to do so is with `env.sample` files, but sample files tend to get out of date very quickly.

For example, let's say you're adding a new feature to your application and it requires the variable `FOO` to be set.
While you're developing locally, you likely have a `.env` file that looks something like:

```bash
APPLICATION_SECRET=supersekrit

# I got this FOO from the detailed process documented at: http:://wiki.example.com/how_to_get_a_foo
FOO="My super secret value for foo"
```

Working on large teams, it's common to share these `.env` files somewhere secure where all developers have access to
them. Retrieval is often integrated into application startup and/or bootstrap processes.

But working on open source projects or teams with less trust and less shared infrastructure, it's more common to share
an `env.sample`. `ess` automatically keeps the sample file in sync with `.env`, so you don't have to. Your
`.env` file stays the same, and is automatically converted to the following `env.sample`:

```bash
APPLICATION_SECRET=<APPLICATION_SECRET>

# I got this FOO from the detailed process documented at: http:://wiki.example.com/how_to_get_a_foo
FOO=<FOO>
```

It's even possible to provide default/example values for every environment variables with the `--example` [flag](#command-flags).

## Command Flags

| Name             | Description                                         | Example                                                   | Default                       |
| ---------------  | --------------------------------------------------- | --------------------------------------------------------- | ----------------------------- |
| `--env-file`     | The name of the environment file                    | `--env-file=.secrets`                                     | `--env-file=.env`             |
| `--sample-file`  | The name of the sample environment file             | `--sample-file=secrets.example`                           | `--sample-file=env.sample`    |
| `--example`      | Provide examples for specific environment variables | `--example=FOO="Example FOO" --example=BAR="Example BAR"` | `--example=VAR=<VAR>`    |

## Pre-commit Configuration Examples

### Default configuration

```yml
repos:
-   repo: https://github.com/acaloiaro/ess.git
    rev: v2.5.0
    hooks:
      - id: ess
```

### Customize `.env` and `env.sample` paths

```yml
repos:
-   repo: https://github.com/acaloiaro/ess.git
    rev: v2.5.0
    hooks:
      - id: ess
        args: ['--env-file=.env_file', '--sample-file=env_file.sample']
```

### Customize variable example values

Sometimes environment variables need to conform to specific formats and it's necessary to provide better documentation.
For this reason, environment variable examples may be provided in lieu of the default behavior, which is to use the
environment variable name surrounded by `<brackets like this>` in sample files.

```yml
repos:
-   repo: https://github.com/acaloiaro/ess.git
    rev: v2.5.0
    hooks:
      - id: ess
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

