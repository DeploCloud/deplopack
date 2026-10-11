package deplopack

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/mise"
	"github.com/railwayapp/railpack/core/plan"
)

type dependency struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	Target       string `json:"target"`
	Image        string `json:"image,omitempty"`
	Version      string `json:"version,omitempty"`
	URL          string `json:"url,omitempty"`
	ChecksumsURL string `json:"checksumsUrl,omitempty"`
	Asset        string `json:"asset,omitempty"`
	Format       string `json:"format,omitempty"`
	Entry        string `json:"entry,omitempty"`
	Executable   string `json:"executable,omitempty"`
}

// Both binaries describe the Linux environment used by the Agent.
func WriteDependencySpec(output io.Writer) error {
	if runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64" {
		return fmt.Errorf("dependencies support Linux amd64 and arm64")
	}
	asset, base, err := mise.LinuxAsset(runtime.GOARCH)
	if err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(struct {
		SchemaVersion int          `json:"schemaVersion"`
		Platform      string       `json:"platform"`
		Dependencies  []dependency `json:"dependencies"`
	}{1, "linux/" + runtime.GOARCH, []dependency{
		{ID: "mise", Kind: "archive", Target: "host", Version: mise.Version, URL: base + asset, ChecksumsURL: base + "SHASUMS256.txt", Asset: asset, Format: "tar.gz", Entry: "mise/bin/mise", Executable: "mise-" + mise.Version},
		{ID: "buildkit", Kind: "docker-image", Target: "host", Image: "moby/buildkit:latest"},
		{ID: "docker-cli", Kind: "docker-image", Target: "host", Image: "docker:29-cli"},
		{ID: "builder-base", Kind: "docker-image", Target: "buildkit-cache", Image: generate.RailpackBuilderImage},
		{ID: "runtime-base", Kind: "docker-image", Target: "buildkit-cache", Image: plan.RailpackRuntimeImage},
		{ID: "alpine", Kind: "docker-image", Target: "buildkit-cache", Image: "alpine:latest"},
	}})
}
