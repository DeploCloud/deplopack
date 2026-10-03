# Standalone detector

The detector analyzes an existing local checkout. The Agent owns Git credentials,
checkout creation, commit resolution, timeouts and cleanup.
It does not import the Railpack CLI or use an argument parser.
The build task strips debug symbols and local source paths from the binary.

```sh
mise run build-detector
cd path/to/project
/path/to/deplopack-detector
DEPLOPACK_SHELL_SCRIPT=serve.sh /path/to/deplopack-detector
```

The current working directory is the project root. Configuration and project
variables are supplied through the process environment, matching the builder.
Use `DEPLOPACK_CONFIG_FILE` for an alternate Railpack config relative to that root.
Both binaries convert `DEPLOPACK_*` configuration to `RAILPACK_*`. The Deplopack
value overrides the original variable, including empty values, independently of
environment ordering. Original `RAILPACK_*` names remain supported.
`DEPLOPACK_PROVIDER`, `BUILDKIT_HOST`, `DOCKER_HOST` and `DOCKER_CONFIG` are excluded
from the project environment. The Agent must construct the child environment explicitly
instead of inheriting unrelated host variables.
Arguments are not accepted except `--version`, which prints the binary version.

The detector calls every provider's existing `Detect` method, in upstream order.
It does not call `Initialize`, `Plan` or `GenerateBuildPlan`, install Mise, resolve
tool versions, or execute project commands. Matches are candidates, not proof that
a build will succeed. Framework-specific build/start suggestions are outside this
initial detection contract.

JSON is written to stdout; process diagnostics go to stderr. For example, a Node
project with a Dockerfile returns:

```json
{
  "success": true,
  "detections": [
    {"type": "node"},
    {"type": "dockerfile", "path": "Dockerfile"}
  ]
}
```

Paths are relative to the current working directory. The detector recognizes
`Dockerfile` at that root and searches recursively for Compose files named
`compose.yaml`, `compose.yml`, `docker-compose.yaml`, `docker-compose.yml` or
variants such as `compose.dev.yaml` and `docker-compose.prod.yml`.
Compose results are ordered by standard filename in the order above, followed
by variants. Within each priority, shallower paths come first, then paths are
sorted alphabetically. All matches are returned, including files excluded from
image build contexts, without selecting a default or describing how to build them.
Language detection respects the upstream
file-access behavior, `.dockerignore`, Railpack config exclusions and explicitly
supplied environment variables.

A configured provider does not add a detection or override matches. Configuration
is still read for the inputs used by upstream detection rules and exclusions.
The builder is responsible for interpreting build configuration and selections.

An empty result is a successful analysis with no recognized project type or file. Config or
detection errors produce a nonzero exit status and `success: false` with any
partial result. Unsupported arguments or inaccessible working directories fail before analysis
and may have no JSON output. The Agent must check both exit status and success.

The upstream provider registry is unchanged. The small detection adapter
constructs only the context fields currently needed by `Detect`; review
this assumption when merging upstream changes to provider detection.
