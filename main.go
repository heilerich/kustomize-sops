package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/getsops/sops/v3"
	"github.com/getsops/sops/v3/decrypt"
	yaml "go.yaml.in/yaml/v3"
	"sigs.k8s.io/kustomize/kustomize/v5/commands"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

type errorReader struct{ error }

func (e errorReader) Read([]byte) (int, error) { return 0, e }

func newDecoder(reader io.Reader) *yaml.Decoder {
	data, err := io.ReadAll(reader)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to read data: %w", err)
		return yaml.NewDecoder(errorReader{wrappedErr})
	}

	cleartext, err := decrypt.Data(data, "yaml")
	if err != nil {
		if errors.Is(err, sops.MetadataNotFound) {
			return yaml.NewDecoder(bytes.NewBuffer(data))
		}
		wrappedErr := fmt.Errorf("failed to decrypt sops file: %w", err)
		return yaml.NewDecoder(errorReader{wrappedErr})
	}

	buffer := bytes.NewBuffer(cleartext)

	return yaml.NewDecoder(buffer)
}

func init() {
	kyaml.NewDecoder = newDecoder
}

func main() {
	if err := commands.NewDefaultCommand().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
