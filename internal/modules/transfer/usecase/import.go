package usecase

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"trisse-home/internal/modules/transfer/domain"
	"trisse-home/internal/shared/kernel"
)

// ImportSource reads the files of another dashboard (output port, one
// adapter per format).
type ImportSource interface {
	Parse(files map[string][]byte) (domain.ImportPlan, error)
}

// ImportOutcome is the merged trisse-home.yaml with what it adds; Result is
// set when the import was applied.
type ImportOutcome struct {
	YAML    []byte
	Summary domain.ImportSummary
	Result  *ApplyResult
}

// Importer turns another dashboard into additions to trisse-home.yaml
// (#77); the input port.
type Importer interface {
	// Import previews the merged file, or applies it when apply is true.
	Import(ctx context.Context, format string, files map[string][]byte, apply bool) (ImportOutcome, error)
}

type importer struct {
	sources map[string]ImportSource
	config  Config
	codec   ConfigCodec
	kinds   func(string) bool
}

// NewImporter returns the Importer. sources maps format names to
// adapters; kinds reports whether a widget kind is available.
func NewImporter(sources map[string]ImportSource, config Config, codec ConfigCodec, kinds func(string) bool) Importer {
	return &importer{sources: sources, config: config, codec: codec, kinds: kinds}
}

var errFormat = kernel.NewValidationError("format", "Unknown import format.")

// importSections are the parts of the file an import can add to.
var importSections = []string{"services", "bookmarks", "dashboard"}

func (uc *importer) Import(ctx context.Context, format string, files map[string][]byte, apply bool) (ImportOutcome, error) {
	src, ok := uc.sources[strings.ToLower(format)]
	if !ok {
		return ImportOutcome{}, errFormat
	}
	plan, err := src.Parse(files)
	if err != nil {
		return ImportOutcome{}, err
	}
	current, err := uc.config.Export(ctx)
	if err != nil {
		return ImportOutcome{}, err
	}
	// The export is read back as plain maps and lists, so the transfer
	// module can add to sections it does not know the types of.
	identity := func(s string) (string, error) { return s, nil }
	generic := func(string) (any, bool) { return new(any), true }
	version, sections, err := uc.codec.Decode(current, identity, generic)
	if err != nil {
		return ImportOutcome{}, fmt.Errorf("read current settings: %w", err)
	}
	values := map[string]any{}
	for _, s := range sections {
		values[s.Name] = *(s.Value.(*any))
	}
	sum := domain.ImportSummary{Warnings: plan.Warnings}
	var touched []string
	if svc, ok := values["services"]; ok && len(plan.Tiles) > 0 {
		list, _ := svc.([]any)
		var added, skipped int
		values["services"], added, skipped = domain.MergeTiles(list, plan.Tiles)
		sum.Tiles, sum.Skipped = added, sum.Skipped+skipped
		touched = append(touched, "services")
	}
	if bm, ok := values["bookmarks"]; ok && len(plan.Folders) > 0 {
		list, _ := bm.([]any)
		var added, skipped int
		values["bookmarks"], added, skipped = domain.MergeFolders(list, plan.Folders)
		sum.Bookmarks, sum.Skipped = added, sum.Skipped+skipped
		touched = append(touched, "bookmarks")
	}
	if d, ok := values["dashboard"].(map[string]any); ok {
		name, n, warnings := domain.AddImportTab(d, plan, sum.Tiles > 0, sum.Bookmarks > 0, uc.kinds)
		sum.Tab, sum.Widgets = name, n
		sum.Warnings = append(sum.Warnings, warnings...)
		if n > 0 {
			touched = append(touched, "dashboard")
		}
	}
	// Only the sections the import adds to are part of the result, so
	// applying it cannot change anything else.
	var out []NamedSection
	for _, name := range importSections {
		if slices.Contains(touched, name) {
			out = append(out, NamedSection{Name: name, Value: values[name]})
		}
	}
	data, err := uc.codec.Encode(version, out)
	if err != nil {
		return ImportOutcome{}, err
	}
	res := ImportOutcome{YAML: data, Summary: sum}
	if apply && len(out) > 0 {
		r, err := uc.config.Apply(ctx, data)
		if err != nil {
			return ImportOutcome{}, err
		}
		res.Result = &r
	}
	return res, nil
}
