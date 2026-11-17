package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"waylight/internal/modules"
	"waylight/internal/platform/config"
)

// configFileApplier is implemented by the transfer module (#88); the
// composition root uses it without the module knowing about start-up.
type configFileApplier interface {
	ApplyManagedFile(ctx context.Context, data []byte) ([]string, error)
	Managed() []string
}

func configApplier(registry *modules.Registry) configFileApplier {
	for _, m := range registry.All() {
		if a, ok := m.(configFileApplier); ok {
			return a
		}
	}
	return nil
}

// applyConfigFile applies CONFIG_FILE before the server accepts requests.
// Any problem stops the start: running with settings that silently differ
// from the file would defeat the point of managing them.
func applyConfigFile(ctx context.Context, cfg config.Config, log *slog.Logger, registry *modules.Registry) error {
	if cfg.ConfigFile == "" {
		return nil
	}
	a := configApplier(registry)
	if a == nil {
		return fmt.Errorf("CONFIG_FILE is set, but the transfer module is disabled")
	}
	data, err := os.ReadFile(cfg.ConfigFile)
	if err != nil {
		return fmt.Errorf("read CONFIG_FILE: %w", err)
	}
	sections, err := a.ApplyManagedFile(ctx, data)
	if err != nil {
		return fmt.Errorf("apply %s: %w", cfg.ConfigFile, err)
	}
	log.InfoContext(ctx, "applied the configuration file; these settings are read-only in the UI", "file", cfg.ConfigFile, "sections", sections)
	return nil
}

// managedSections returns the sections CONFIG_FILE owns.
func managedSections(registry *modules.Registry) map[string]bool {
	out := map[string]bool{}
	if a := configApplier(registry); a != nil {
		for _, s := range a.Managed() {
			out[s] = true
		}
	}
	return out
}
