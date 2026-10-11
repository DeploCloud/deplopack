# Dependency catalog

Both standalone binaries support `--dependency-spec`. The command prints JSON without reading the project or starting Mise.

Schema version 1 declares a Linux platform (`linux/amd64` or `linux/arm64`) and an ordered list of dependencies. `kind` distinguishes verified archives from Docker images. Images with `target: host` belong in Docker storage; `target: buildkit-cache` images must be fully downloaded and unpacked in BuildKit storage.

The Mise archive includes its version, asset and checksum URLs, archive format, expected internal executable and cache filename. The catalog derives Mise and base image versions from the planner constants. Language runtimes and project dependencies are downloaded during the build.

The Agent fingerprints the catalog and platform, prepares dependencies in background, and exports the common images as OCI image layouts. Each job mounts these layouts read-only and uses explicit BuildKit OCI image sources. BuildKit creates the App’s snapshots itself; no internal databases or OverlayFS directories are copied. Existing App caches are preserved. Preparation failures remain visible and can be retried on first use.

`DEPLOPACK_OCI_IMAGES` is an Agent-managed operational variable containing a JSON object mapping original image references to local OCI layout directories. It is excluded from the application environment and build secrets. App overrides cannot set it.
