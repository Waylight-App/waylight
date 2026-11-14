package gateway

import (
	"bytes"
	"errors"
	"io"
	"strconv"

	"go.yaml.in/yaml/v3"

	"trisse-home/internal/modules/transfer/usecase"
	"trisse-home/internal/shared/kernel"
)

// SchemaURL is the published JSON Schema of trisse-home.yaml; editors with
// the YAML language server validate the file against it.
const SchemaURL = "https://raw.githubusercontent.com/itadori26/trisse-home/main/docs/schema/trisse-home.v1.schema.json"

// YAMLCodec implements usecase.ConfigCodec (#88).
type YAMLCodec struct{}

// Encode implements usecase.ConfigCodec. Section values are encoded in
// struct field order, so the same settings always give the same bytes.
func (YAMLCodec) Encode(version int, sections []usecase.NamedSection) ([]byte, error) {
	root := &yaml.Node{Kind: yaml.MappingNode, HeadComment: "yaml-language-server: $schema=" + SchemaURL}
	root.Content = append(root.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "version"},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.Itoa(version)})
	for _, s := range sections {
		var v yaml.Node
		if err := v.Encode(s.Value); err != nil {
			return nil, err
		}
		root.Content = append(root.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: s.Name}, &v)
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{root}}); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func invalid(err error) error {
	return kernel.NewValidationErrorf("file", "The file is not valid: {reason}", "reason", err.Error())
}

// Decode implements usecase.ConfigCodec. Unknown fields are errors, so a
// misspelt key is reported instead of silently ignored.
func (YAMLCodec) Decode(data []byte, expand func(string) (string, error), target func(string) (any, bool)) (int, []usecase.NamedSection, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return 0, nil, invalid(err)
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return 0, nil, kernel.NewValidationError("file", "The file must be a YAML mapping with version and sections.")
	}
	root := doc.Content[0]
	version := 0
	var out []usecase.NamedSection
	seen := map[string]bool{}
	for i := 0; i+1 < len(root.Content); i += 2 {
		key, val := root.Content[i].Value, root.Content[i+1]
		if seen[key] {
			return 0, nil, kernel.NewValidationErrorf("file", "The section {name} appears twice.", "name", key)
		}
		seen[key] = true
		if key == "version" {
			if err := val.Decode(&version); err != nil {
				return 0, nil, kernel.NewValidationError("version", "The version must be a number.")
			}
			continue
		}
		dst, ok := target(key)
		if !ok {
			return 0, nil, kernel.NewValidationErrorf("file", "Unknown section {name} (or its module is disabled).", "name", key)
		}
		if err := expandStrings(val, expand); err != nil {
			return 0, nil, err
		}
		if err := decodeStrict(val, dst); err != nil {
			return 0, nil, kernel.NewValidationErrorf("file", "Section {name}: {reason}", "name", key, "reason", err.Error())
		}
		out = append(out, usecase.NamedSection{Name: key, Value: dst})
	}
	return version, out, nil
}

// expandStrings replaces ${NAME} in every string scalar below n. Keys are
// left alone; only values can come from the environment.
func expandStrings(n *yaml.Node, expand func(string) (string, error)) error {
	switch n.Kind {
	case yaml.ScalarNode:
		if n.ShortTag() != "!!str" {
			return nil
		}
		v, err := expand(n.Value)
		if err != nil {
			return err
		}
		n.Value = v
	case yaml.MappingNode:
		for i := 1; i < len(n.Content); i += 2 {
			if err := expandStrings(n.Content[i], expand); err != nil {
				return err
			}
		}
	case yaml.SequenceNode, yaml.DocumentNode:
		for _, c := range n.Content {
			if err := expandStrings(c, expand); err != nil {
				return err
			}
		}
	case yaml.AliasNode:
		return kernel.NewValidationError("file", "YAML aliases (*name) are not supported; write the value out.")
	}
	return nil
}

// decodeStrict decodes n into dst and rejects fields dst does not have;
// yaml.Node.Decode has no such option, so n takes a trip through bytes.
func decodeStrict(n *yaml.Node, dst any) error {
	b, err := yaml.Marshal(n)
	if err != nil {
		return err
	}
	dec := yaml.NewDecoder(bytes.NewReader(b))
	dec.KnownFields(true)
	if err := dec.Decode(dst); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
