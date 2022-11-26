# env-sample-pre-commit-hook

This [pre-commit hook](https://pre-commit.com/#install) checks whether the local repository has an `.env` file and if one exists, it is scrubbed of secrets/values and made available as `env.sample` (configurable). This ensures that all application environment variables are safely and automatically documented without leaking secrets.

# Usage

For details on configuring and installing `pre-commit`, refer to https://pre-commit.com/#install

To use this pre-commit hook, add one of the following configurations to your repository as `.pre-commit-config.yaml` and run `pre-commit install`

## Add configuration
```bash
cat <<EOF > .pre-commit-config.yaml
repos:
-   repo: git+ssh://github.com:/acaloiaro/env-sample-sync
    rev: v1
    hooks:
      - id: env-sample-sync
EOF
pre-commit install
```

### Default configuration
```yml
repos:
-   repo: git+ssh://github.com:/acaloiaro/env-sample-sync
    rev: v1
    hooks:
      - id: env-sample-sync
```

### Custom `.env` and `env.sample` coniguration

```yml
repos:
-   repo: git+ssh://github.com:/acaloiaro/env-sample-sync
    rev: v1
    hooks:
      - id: env-sample-sync
        args: ['--env-file=.env_file', '--sample-file=env_file.sample']
```

# Example Output

`.env`
```
# This key is used as cryptographic salt throughout the application
APPLICATION_KEY=supersekrit
```

`env.sample`
```bash
# This key is used as cryptographic salt throughout the application
APPLICATION_KEY=<APPLICATION_KEY>
```



