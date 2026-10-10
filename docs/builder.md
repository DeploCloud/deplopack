# Standalone builder

The Agent clones the selected commit, provides the builder binary and starts it
with the project root as its working directory. The builder generates the plan
with the selected provider and executes the existing BuildKit pipeline.
It calls the planner and BuildKit directly without importing the Railpack CLI.
The build task strips debug symbols and local source paths from the binary.

```sh
mise run build-builder
cd path/to/project
DEPLOPACK_PROVIDER=node BUILDKIT_HOST=docker-container://buildkit /path/to/deplopack-builder
```

Configuration is supplied entirely through process environment variables:

| Variable | Purpose |
| --- | --- |
| `DEPLOPACK_PROVIDER` | Required Railpack provider selected from detector results |
| `DEPLOPACK_BUILD_CMD` | Optional custom build command |
| `DEPLOPACK_START_CMD` | Optional custom start command |
| `DEPLOPACK_CONFIG_FILE` | Optional config path relative to the project root |
| `BUILDKIT_HOST` | Required endpoint of an existing BuildKit daemon |

The selected provider overrides `railpack.json`; unknown or failed provider
selections fail the build. The planner initializes that provider directly without
repeating the general detection scan. Provider-specific analysis needed to plan
the build, including secondary runtimes, still runs.
Omitted build/start commands use upstream provider defaults and configuration.
Explicit `RAILPACK_*` environment configuration takes precedence over
`railpack.json` in the selected-provider build path and is forwarded directly
to the upstream planner.

Both binaries convert `DEPLOPACK_*` configuration to `RAILPACK_*`, overriding
the original variable even when the Deplopack value is empty. Precedence is
independent of environment ordering. `DEPLOPACK_PROVIDER` is reserved for provider
selection and is not converted. Original `RAILPACK_*` names remain supported.

Other environment variables and the converted `RAILPACK_*` configuration are
available to build steps as upstream BuildKit secrets. `DEPLOPACK_PROVIDER`,
`BUILDKIT_HOST`, `DOCKER_HOST` and `DOCKER_CONFIG` are excluded from project
secrets. The Agent must construct the child
environment explicitly instead of inheriting unrelated host credentials.
`GITHUB_TOKEN` retains upstream behavior for authenticated tool downloads.

The builder prints a readable package, command and startup summary, followed by
BuildKit progress; it does not print the plan JSON. Its banner and internal progress
prefixes identify DeploPack, while upstream image names, URLs and configuration
variables retain their original names. It produces an image loaded
into Docker. The image name defaults to the project directory name in lowercase;
the target platform defaults to Linux with the host architecture.
The pipeline requires Docker CLI to load the resulting image. Docker authentication
uses the existing Docker config. Neither Git cloning nor application startup is performed here.
Dockerfile and Compose remain separate deployment methods.
An injected executable inside the checkout should be excluded by the Agent's
build context ignore rules, so it is not copied into the application image.

Logs and progress retain the upstream format. The exit status is zero for success,
one for build failure, and 75 for transient planning failures. `--version` prints
the binary version; build options and source directory arguments are not accepted.

Image builds require a running BuildKit daemon and Docker.
