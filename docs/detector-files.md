# Files needed by repository analysis

This inventory follows `internal/detection.Analyze`, provider `Detect` and `Inspect`, their reachable helpers, input defaults, Procfile parsing and Mise declarations. It does not include files used only by `Plan`, installation or compilation.

`cmd/detector/checkout-spec.json` is embedded in the detector and returned by `--checkout-spec`. Each release carries its own inventory. Names and basename patterns match at any depth so workspace manifests and nested projects remain available. Some lockfiles are retained conservatively even when the provider only checks their existence. All TOML files and `*.lock` files remain selected.

## Content and existence are different

The complete tree, directory hierarchy, file kinds and executable bits must remain available. `HasFile`, `HasMatch`, `FindFiles` and `FindDirectories` affect candidates, framework selection and default commands. Missing content must not make an existing source file appear absent.

Files whose contents are not selected are represented by empty placeholders. Correctness depends on the release inventory covering every content read: provider changes must update the specification alongside the code. Configurations with dynamic dependencies request a complete checkout.

## Provider inventory

All paths below are relative to the analyzed directory unless marked recursive.

| Reader | Contents needed | Existence or names needed |
| --- | --- | --- |
| Shared configuration | `.dockerignore`, `railpack.json` or explicit config path, `Procfile` | Config presence and excluded paths |
| Mise | Supported TOML configuration, fragments, environment variants and corresponding lockfiles; `.tool-versions`; enabled idiomatic version files | Configuration hierarchy and excluded configuration paths |
| Node | `package.json` / `package.json5`; workspace manifests selected by workspace globs; `pnpm-workspace.yaml`; `pnpm-lock.yaml`; `.yarnrc.yml`; version declarations; Next, Vite, Astro and React Router configuration; `angular.json`; Expo `app.json` | Package-manager lockfiles; framework config markers including `nx.json`; workspace paths; root `index.js` / `index.ts`; supporting installation files and directories used to infer commands |
| Python | Recursive `requirements.txt`, `pyproject.toml`, `Pipfile`; `.python-version`; `runtime.txt`; **recursive `*.py` for Django application discovery** | Entrypoints `main.py`, `app.py`, `start.py`, `bot.py`, `hello.py`, `server.py`; `manage.py`; package-manager lockfiles |
| Ruby | `Gemfile`, `Gemfile.lock`, `.ruby-version`, `config/application.rb` | `rails`, `bin/rails`, `config/environment.rb`, `script`, `config.ru`, `Rakefile` |
| Go | `go.mod`, plus declarations read by Mise | `go.mod`, `go.work`, `main.go`, root `*.go`; recursive module paths |
| .NET | Root `*.csproj`, `global.json`, declarations read by Mise | Project filenames determine the startup DLL |
| Rust | `Cargo.toml`, declarations read by Mise | Cargo manifest presence |
| Java | Declarations read by Mise | Maven `pom.{xml,atom,clj,groovy,rb,scala,yaml,yml}`, `gradlew`; recursive Spring Boot JAR/class names and `org/springframework/boot` paths. Build-system file contents are not read by the current `Inspect`. |
| Elixir | `.elixir-version`, `.erlang-version`, declarations read by Mise | `mix.exs` |
| Deno | Declarations read by Mise; root `main.{ts,js,mjs,mts}` is opened by `ReadFirstFileOf` to choose a readable entrypoint, although its contents are discarded | `deno.json` / `deno.jsonc`; first recursive `*.{ts,js,mjs,mts}` entrypoint when there is no root main file |
| PHP | Declarations read by Mise | `index.php`, `composer.json`, `artisan`. Current `Inspect` checks Laravel presence without parsing Composer or PHP sources. |
| C++ | Declarations read by Mise | `CMakeLists.txt`, `meson.build` |
| Gleam | Declarations read by Mise | `gleam.toml` |
| Shell | `start.sh` or explicitly configured script, including its shebang | Script path and file metadata |
| Static files | `Staticfile` | `public` directory or root `index.html`; static assets themselves are not read |
| Dockerfile / Compose candidates | No file content needed by current candidate discovery | Root `Dockerfile`; recursive `**/{compose,docker-compose}{,.*}.{yaml,yml}`, including paths excluded by `.dockerignore` |

## Mise declarations

Fixed configuration paths: `mise.toml`, `.mise.toml`, `mise/config.toml`, `.mise/config.toml`, `.config/mise.toml`, `.config/mise/config.toml`, `.config/mise/mise.toml`, `.config/mise/mise.local.toml`, `.tool-versions`.

Environment patterns: `mise.*.toml`, `.mise.*.toml`, `.config/mise.*.toml`, `mise/config.*.toml`, `.mise/config.*.toml`, `.config/mise/config.*.toml`. Fragment patterns: `{mise,.mise,.config/mise}/conf.d/*.toml`. Preserve the lockfile corresponding to each configuration file, including environment variants.

The existing declaration inventory includes `.python-version`, `.python-versions`, `.node-version`, `.nvmrc`, `.ruby-version`, `Gemfile`, `.go-version`, `.java-version`, `.sdkmanrc`, `.exenv-version`, `.deno-version`, `rust-toolchain.toml`, `.bun-version`, `.yvmrc`, `global.json`. Provider-specific `.elixir-version` and `.erlang-version` must also remain available. Mise may read additional idiomatic declarations, such as a legacy `rust-toolchain`.

Mise is an external reader: configuration can reference arbitrary local files through env files, templates or other settings. A literal filename inventory alone cannot guarantee equivalent results. Explicit config/script paths and internal symlink targets override the selection. Resolve local references or fall back to a broader checkout; never silently replace a required file with empty contents.

## Local benchmark

The benchmark uses the same Reloop commit and detector binary for both checkouts. The baseline excludes the Agent's current image/video extensions and restores empty directories. The experimental checkout downloads only inventory-selected contents, preserving all other file names with placeholders. Compare the complete detector JSON, including candidates, metadata, inputs and logs; measure checkout contents and Git object storage separately.

Git object storage is a proxy for downloaded compressed data, not an HTTP byte counter. Worktree size means logical file contents, excluding `.git`; allocation overhead and placeholders are measured separately. No claim about arbitrary repositories follows from this one comparison.

### Reloop results

Repository: `https://github.com/reloop-labs/reloop.git`. Commit: `d523f8de99e4abaa953f103786146510f28d0f20`. Detector source: DeploPack `d8108dbd`, compiled with the existing `build-detector` Mise task on macOS arm64. Three independent shallow partial clones for each selection, followed by two detector executions per checkout. Results and reproducible script are retained under the ignored `tmp/detector-filter/` directory.

| Measurement | Current media filter | Content selection | Reduction |
| --- | ---: | ---: | ---: |
| Materialized file contents | 4,555 files | 49 files | 98.9% |
| Worktree logical contents | 28,360,176 B | 720,792 B | 97.5% |
| Git object storage | 9,282,146–9,282,182 B | 472,420 B | 94.9% |
| Worktree allocated storage, excluding `.git` | 39,012 KiB | 844 KiB | 97.8% |
| Checkout command, median | 7.61 s | 1.10 s | 85.5% |
| Preparation including fetch and placeholders, median | 9.18 s | 3.30 s | 64.1% |
| Detector, median of second execution | 0.277 s | 0.275 s | Essentially unchanged |

The complete detector JSON matched across all six checkouts and repeated executions: `node` followed by the three Compose paths, the same Node metadata, all 19 inputs, and the same logs. The very first detector execution took 1.89 s while initializing its local cache; subsequent runs took 0.25–0.33 s. Timing results are local measurements, not predictions for the Linux test server.

The 4,683 tree entries and their directory hierarchy were preserved in the experimental checkout. Unselected files have empty placeholder contents and their original executable bits. The measured repository has no symlinks or submodules; the benchmark refuses those cases rather than pretending to support them. Arbitrary Mise references and custom configuration/script paths were not exercised. These measurements preceded the Agent implementation below.


## Agent implementation

Before checkout, the Agent executes the verified detector with `--checkout-spec`, using a ten-second timeout and a 64 KiB output limit. It accepts schema version 1 and rejects unknown fields, invalid paths, malformed glob patterns and oversized rule lists. Missing, invalid or unsupported specifications fall back to a complete shallow checkout with an informational result log. Cancellation still aborts the request. The deploy checkout is unchanged.

Schema version 1 fields (all lists are required, may be empty and are limited to 256 entries each):

| Field | Meaning |
| --- | --- |
| `schemaVersion` | Contract version; currently `1` |
| `preserveTree` | Must be `true`; preserve directories, names and executable bits |
| `contentBasenames` | Exact filenames whose content is needed, at any depth |
| `contentPatterns` | Basename globs using Go `path.Match`, e.g. `*.py`; no recursive glob syntax |
| `contentPathSuffixes` | Exact relative path suffixes whose content is needed, e.g. `config/application.rb` |
| `fullCheckoutBasenames` | Exact filenames that require all content |
| `fullCheckoutPatterns` | Basename globs that require all content |
| `fullCheckoutDirectories` | Directory component names that require all content |
| `pathEnvironment` | `DEPLOPACK_*` variables containing explicit project-relative file paths to materialize |

Explicit configuration/script paths and internal symlink targets retain real content. Symlinks are validated before placeholders are created; submodules remain unsupported. All fields describe selection rules, never commands or paths outside the checkout.

The current specification conservatively requests a complete checkout for `.tool-versions`, Mise TOML basenames or files under `mise`/`.mise` directories. This preserves arbitrary template, env-file and local-backend dependencies without another Mise parser. Idiomatic version files remain selected normally. Servers without partial clone support still transfer all blobs; selection only reduces local materialized contents there.

The inventory lives exclusively in `cmd/detector/checkout-spec.json`. Compatible rule updates ship with the detector without an Agent update. Changing the contract itself requires a new schema version; older Agents then use a complete checkout.

### Agent verification

The actual Agent checkout functions were compared locally using temporary self-checks under `tmp/analysis-filter/`. Full detector JSON matched on ten isolated examples (Node/Vite, Python/Django, Go, Ruby, Rust, .NET, Java/Gradle, PHP, Elixir and Deno), as well as a fixture with a media directory and an executable symlink. The Go and Elixir Mise fixtures used the complete-content fallback. Explicit script selection, invalid subdirectory rejection and cancellation were checked. Linux vet and builds for amd64/arm64 passed.

A fresh Reloop comparison using the previous Agent media filter and the implemented content filter measured preparation at 8.289 s versus 2.979 s, and Git object storage at 9,285,194 B versus 472,435 B (94.9% less). The complete detector JSON was identical. This is one local comparison, not a server timing guarantee. No push or server update was performed.


After moving ownership to the detector, the temporary self-check was rerun using `--checkout-spec`. Outputs matched on all ten examples; invalid schemas, unknown fields, null lists, unsafe paths, malformed patterns, oversized output, unsupported flags and cancellation used the expected fallback/error paths. DeploPack `mise run check` completed with zero lint issues, and Linux Agent vet/builds passed. A fresh full Reloop clone failed with a Git error; a subsequent fresh selective clone was compared successfully against the preserved baseline checkout of the same commit, producing identical detector JSON. No release, commit, push or server update was performed.
