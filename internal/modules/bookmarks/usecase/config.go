package usecase

import (
	"context"
	"fmt"
	"strings"

	"trisse-home/internal/modules/bookmarks/domain"
	"trisse-home/internal/shared/kernel"
)

// ConfigBookmark is one bookmark in trisse-home.yaml (#88).
type ConfigBookmark struct {
	Title       string
	URL         string
	Description string
	Tags        []string
}

// ConfigFolder is a folder with its bookmarks.
type ConfigFolder struct {
	Name      string
	Bookmarks []ConfigBookmark
}

// ConfigSection holds the bookmarks without a folder and the folders in
// their order. Bookmarks are listed by title, as everywhere else.
type ConfigSection struct {
	Unfiled []ConfigBookmark
	Folders []ConfigFolder
}

// ConfigRepository adds the atomic replacement of all bookmarks.
type ConfigRepository interface {
	Repository
	// ReplaceAll makes folders and bookmarks the complete set; rows whose
	// IDs are missing are deleted together with their tags, icons and
	// link checks.
	ReplaceAll(ctx context.Context, folders []domain.Folder, bookmarks []domain.Bookmark) error
}

// Config exports and applies the bookmarks section (input port).
type Config interface {
	Export(ctx context.Context) (ConfigSection, error)
	Apply(ctx context.Context, s ConfigSection) error
}

type config struct {
	repo  ConfigRepository
	clock kernel.Clock
	ids   kernel.IDGenerator
}

// NewConfig returns the Config use case.
func NewConfig(repo ConfigRepository, clock kernel.Clock, ids kernel.IDGenerator) Config {
	return &config{repo: repo, clock: clock, ids: ids}
}

func toConfig(b domain.Bookmark) ConfigBookmark {
	return ConfigBookmark{Title: b.Title(), URL: b.URL(), Description: b.Description(), Tags: b.Tags()}
}

// Export and Apply work on the admin's bookmarks: config as code describes
// the instance, not members' personal bookmarks (ADR 0018), and they also
// run at start as the system, which would see everyone's.
func (uc *config) Export(ctx context.Context) (ConfigSection, error) {
	ctx = kernel.WithPrincipal(ctx, kernel.AdminPrincipal())
	folders, err := uc.repo.ListFolders(ctx)
	if err != nil {
		return ConfigSection{}, fmt.Errorf("list folders: %w", err)
	}
	list, err := uc.repo.ListBookmarks(ctx)
	if err != nil {
		return ConfigSection{}, fmt.Errorf("list bookmarks: %w", err)
	}
	byFolder := map[string][]ConfigBookmark{}
	for _, b := range list {
		byFolder[b.FolderID()] = append(byFolder[b.FolderID()], toConfig(b))
	}
	out := ConfigSection{Unfiled: byFolder[""]}
	for _, f := range folders {
		out.Folders = append(out.Folders, ConfigFolder{Name: f.Name(), Bookmarks: byFolder[f.ID()]})
	}
	return out, nil
}

var errDuplicateFolder = kernel.NewValidationError("folder", "Each folder name can be used only once in the file.")

// Apply keeps the IDs of folders with the same name and of bookmarks with
// the same URL, so fetched icons and link checks survive a re-import.
func (uc *config) Apply(ctx context.Context, s ConfigSection) error {
	ctx = kernel.WithPrincipal(ctx, kernel.AdminPrincipal())
	oldFolders, err := uc.repo.ListFolders(ctx)
	if err != nil {
		return fmt.Errorf("list folders: %w", err)
	}
	oldBookmarks, err := uc.repo.ListBookmarks(ctx)
	if err != nil {
		return fmt.Errorf("list bookmarks: %w", err)
	}
	folderIDs := map[string]string{}
	for _, f := range oldFolders {
		folderIDs[strings.ToLower(f.Name())] = f.ID()
	}
	byURL := map[string]domain.Bookmark{}
	for _, b := range oldBookmarks {
		byURL[b.URL()] = b
	}
	now := uc.clock.Now()
	var bookmarks []domain.Bookmark
	add := func(folderID string, c ConfigBookmark) error {
		in := domain.BookmarkInput{FolderID: folderID, Title: c.Title, URL: c.URL, Description: c.Description, Tags: c.Tags}
		if old, ok := byURL[strings.TrimSpace(c.URL)]; ok {
			delete(byURL, old.URL()) // a URL listed twice gets a new ID the second time
			if err := old.Update(in); err != nil {
				return err
			}
			bookmarks = append(bookmarks, old)
			return nil
		}
		b, err := domain.NewBookmark(uc.ids.NewID(), in, now)
		if err != nil {
			return err
		}
		bookmarks = append(bookmarks, b)
		return nil
	}
	for _, c := range s.Unfiled {
		if err := add("", c); err != nil {
			return err
		}
	}
	seen := map[string]bool{}
	folders := make([]domain.Folder, 0, len(s.Folders))
	for i, cf := range s.Folders {
		key := strings.ToLower(strings.TrimSpace(cf.Name))
		if seen[key] {
			return errDuplicateFolder
		}
		seen[key] = true
		id, ok := folderIDs[key]
		if !ok {
			id = uc.ids.NewID()
		}
		f, err := domain.NewFolder(id, cf.Name, i)
		if err != nil {
			return err
		}
		folders = append(folders, f)
		for _, c := range cf.Bookmarks {
			if err := add(id, c); err != nil {
				return err
			}
		}
	}
	return uc.repo.ReplaceAll(ctx, folders, bookmarks)
}
