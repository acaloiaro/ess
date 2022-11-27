# env-sample-sync 

This [pre-commit plugin](https://pre-commit.com/#install) safely keeps `.env` files in sync with `env.sample`.

It checks whether the local repository has an `.env` file and if one exists, it is scrubbed of secrets/values and made available as `env.sample`. This ensures that all application environment variables are safely and automatically documented without leaking secrets.

For details on configuring and installing `pre-commit`, refer to https://pre-commit.com/#install

# Background

It's important to document the environment variables required to run applications, both in production and development. A great way to do so is with `env.sample` files.

For example, let's say you're adding a new feature to your application and it requires the variable `FOO` to be set. While you're developing locally, you likely have a `.env` file that looks something like:

```bash
APPLICATION_SECRET=supersekrit

# I got this FOO from the detailed process documented at: http:://wiki.example.com/how_to_get_a_foo
FOO="My super secret value for foo"
```

Working on large teams, it's common to share these `.env` files somewhere secure where all developers have access to them. Retrieval is often integrated into application startup and/or bootstrap processes.

But working on open source projects or teams with less trust and less shared infrastructure, it's more common to share an `env.sample`. This pre-commit plugin automatically keeps the sample file in sync with `.env`, so you don't have to. Your `.env` file stays the same, and is automatically converted to the following `env.sample`:

```bash
APPLICATION_SECRET=<APPLICATION_SECRET>

# I got this FOO from the detailed process documented at: http:://wiki.example.com/how_to_get_a_foo
FOO=<FOO>
```

It's even possible to provide default/example values for for every environment variables with the `--example` [argument](#arguments).

# Usage

Add this hook to your repository as follows:

## Add configuration
```bash
cat <<EOF > .pre-commit-config.yaml
repos:
-   repo: https://github.com/acaloiaro/pre-commit-env-sample-sync.git
    rev: 8c51d1e303eeb619ad42dd2259c3692faf456569
    hooks:
      - id: env-sample-sync
EOF
pre-commit install
```

## Hook arguments

This plugin accepts the following arguments

| Name                  | Description                                               | Example                                                   | Default                       |
| --------------------  | --------------------------------------------------------- | --------------------------------------------------------- | ----------------------------- |
| `-e`/`--env-file`     | The name of the environment file in the repository        | `--env-file=.secrets`                                     | `--env-file=.env`             |
| `-s`/`--sample-file`  | The name of the sample environment file in the repository | `--sample-file=secrets.example`                           | `--sample-file=env.sample`    |
| `-x`/`--example`      | Provide examples for specific environment variables       | `--example=FOO="Example FOO" --example=BAR="Example BAR"` | `--example=<ENV_VAR_NAME>`    |

## Configuration Examples

### Default configuration

```yml
repos:
-   repo: https://github.com/acaloiaro/pre-commit-env-sample-sync.git
    rev: 8c51d1e303eeb619ad42dd2259c3692faf456569
    hooks:
      - id: env-sample-sync
```

### Customize `.env` and `env.sample` paths

```yml
repos:
-   repo: https://github.com/acaloiaro/pre-commit-env-sample-sync.git
    rev: 8c51d1e303eeb619ad42dd2259c3692faf456569
    hooks:
      - id: env-sample-sync
        args: ['--env-file=.env_file', '--sample-file=env_file.sample']
```

### Customize variable example values

Sometimes environment variables need to conform to specific formats and it's necessary to provide better documentation. For this reason, environment variable examples may be provided in lieu of the default behavior, which is to use the environment variable name surrounded by `<brackets like this>` in sample files.

`.pre-commit-config.yml`
```yml
repos:
-   repo: https://github.com/acaloiaro/pre-commit-env-sample-sync.git
    rev: 8c51d1e303eeb619ad42dd2259c3692faf456569
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

Currently, this hook only supports `.env` files with environment variables of the form `VAR_NAME=VAR_VALUE`. [Direnv](https://direnv.net/)-style `.envrc` is not currently supported. Pull requests welcome!

