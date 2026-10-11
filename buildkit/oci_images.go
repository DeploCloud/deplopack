package buildkit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/containerd/containerd/v2/core/content"
	"github.com/containerd/containerd/v2/plugins/content/local"
	"github.com/moby/buildkit/client/llb"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

func loadOCIImages(images map[string]string, platform specs.Platform) (map[string]llb.State, map[string]content.Store, error) {
	states := map[string]llb.State{}
	stores := map[string]content.Store{}
	for ref, directory := range images {
		file := filepath.Join(directory, "index.json")
		info, err := os.Lstat(file)
		if err != nil || !info.Mode().IsRegular() || info.Size() > 64<<10 {
			return nil, nil, fmt.Errorf("invalid local OCI image index: %s", ref)
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, nil, err
		}
		var index specs.Index
		if json.Unmarshal(data, &index) != nil || index.SchemaVersion != 2 || len(index.Manifests) != 1 || index.Manifests[0].Digest.Validate() != nil {
			return nil, nil, fmt.Errorf("invalid local OCI image descriptor: %s", ref)
		}
		store, err := local.NewStore(directory)
		if err != nil {
			return nil, nil, err
		}
		id := index.Manifests[0].Digest.Encoded()
		stores[id] = store
		states[ref] = llb.OCILayout("deplopack/image@"+index.Manifests[0].Digest.String(), llb.OCIStore("", id), llb.Platform(platform), llb.WithCustomName("[deplopack] local image "+ref))
	}
	return states, stores, nil
}
