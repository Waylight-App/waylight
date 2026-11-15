package usecase

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"trisse-home/internal/modules/bookmarks/domain"
	"trisse-home/internal/shared/kernel"
)

func toView(ctx context.Context, b domain.Bookmark) BookmarkView {
	vis, mine := accessOf(ctx, b.Access())
	return BookmarkView{ID: b.ID(), FolderID: b.FolderID(), Title: b.Title(), URL: b.URL(), Description: b.Description(), Tags: b.Tags(), CreatedAt: b.CreatedAt(),
		Visibility: vis, Mine: mine}
}

// accessOf maps owner and visibility for a view. A record without an owner
// was just created by the viewer and has the default visibility.
func accessOf(ctx context.Context, a kernel.Access) (string, bool) {
	p := kernel.PrincipalFrom(ctx)
	if a.Owner == "" {
		return string(kernel.DefaultVisibility(p)), true
	}
	return string(a.Visibility), a.Mine(p)
}

func (in Input) toDomain() domain.BookmarkInput {
	return domain.BookmarkInput{FolderID: in.FolderID, Title: in.Title, URL: in.URL, Description: in.Description, Tags: in.Tags}
}

func matches(b domain.Bookmark, f Filter) bool {
	switch {
	case f.FolderID == UnfiledFolderID && b.FolderID() != "":
		return false
	case f.FolderID != "" && f.FolderID != UnfiledFolderID && b.FolderID() != f.FolderID:
		return false
	case f.Tag != "" && !slices.Contains(b.Tags(), strings.ToLower(f.Tag)):
		return false
	}
	q := strings.ToLower(strings.TrimSpace(f.Query))
	return q == "" || strings.Contains(strings.ToLower(b.Title()+" "+b.URL()+" "+b.Description()), q)
}

type listBookmarks struct{ d Deps }

// NewListBookmarks returns the ListBookmarks use case.
func NewListBookmarks(d Deps) ListBookmarks { return &listBookmarks{d: d} }

// Execute implements ListBookmarks.
func (uc *listBookmarks) Execute(ctx context.Context, f Filter) (ListResponse, error) {
	all, err := uc.d.Repo.ListBookmarks(ctx)
	if err != nil {
		return ListResponse{}, fmt.Errorf("list bookmarks: %w", err)
	}
	folders, err := uc.d.Repo.ListFolders(ctx)
	if err != nil {
		return ListResponse{}, fmt.Errorf("list folders: %w", err)
	}
	statuses := map[string]domain.LinkStatus{}
	if uc.d.Links != nil {
		if statuses, err = uc.d.Links.ListLinkStatuses(ctx); err != nil {
			return ListResponse{}, fmt.Errorf("list link statuses: %w", err)
		}
	}
	icons := map[string]bool{}
	if uc.d.Icons != nil {
		if icons, err = uc.d.Icons.IconIDs(ctx); err != nil {
			return ListResponse{}, fmt.Errorf("list icons: %w", err)
		}
	}
	res := ListResponse{Bookmarks: []BookmarkView{}, Folders: []FolderView{}, Tags: []TagView{}}
	perFolder := map[string]int{}
	perTag := map[string]int{}
	for _, b := range all {
		perFolder[b.FolderID()]++
		for _, t := range b.Tags() {
			perTag[t]++
		}
		if matches(b, f) && (!f.Broken || statuses[b.ID()].State == domain.LinkBroken) {
			v := toView(ctx, b)
			v.Link, v.HasIcon = statuses[b.ID()], icons[b.ID()]
			res.Bookmarks = append(res.Bookmarks, v)
		}
	}
	for _, fo := range folders {
		res.Folders = append(res.Folders, withCount(folderView(ctx, fo), perFolder[fo.ID()]))
	}
	for name, n := range perTag {
		res.Tags = append(res.Tags, TagView{Name: name, Count: n})
	}
	slices.SortFunc(res.Tags, func(a, b TagView) int { return strings.Compare(a.Name, b.Name) })
	return res, nil
}

type createBookmark struct{ d Deps }

// NewCreateBookmark returns the CreateBookmark use case.
func NewCreateBookmark(d Deps) CreateBookmark { return &createBookmark{d: d} }

// Execute implements CreateBookmark.
func (uc *createBookmark) Execute(ctx context.Context, in Input) (BookmarkView, error) {
	if err := ensureFolder(ctx, uc.d.Repo, in.FolderID); err != nil {
		return BookmarkView{}, err
	}
	b, err := domain.NewBookmark(uc.d.IDs.NewID(), in.toDomain(), uc.d.Clock.Now())
	if err != nil {
		return BookmarkView{}, err
	}
	if err := uc.d.Repo.SaveBookmark(ctx, b); err != nil {
		return BookmarkView{}, fmt.Errorf("save bookmark: %w", err)
	}
	v := toView(ctx, b)
	// Title and favicon are fetched right away when enabled (#72); a page
	// that cannot be read leaves the bookmark as entered.
	if s, err := uc.d.settings(ctx); err == nil && uc.d.canEnrich(s) {
		if enriched, hasIcon, err := uc.d.enrich(ctx, b); err == nil {
			v = toView(ctx, enriched)
			v.HasIcon = hasIcon
		}
	}
	return v, nil
}

type updateBookmark struct{ d Deps }

// NewUpdateBookmark returns the UpdateBookmark use case.
func NewUpdateBookmark(d Deps) UpdateBookmark { return &updateBookmark{d: d} }

// Execute implements UpdateBookmark.
func (uc *updateBookmark) Execute(ctx context.Context, id string, in Input) (BookmarkView, error) {
	b, err := uc.d.Repo.GetBookmark(ctx, id)
	if err != nil {
		return BookmarkView{}, err
	}
	if err := ensureFolder(ctx, uc.d.Repo, in.FolderID); err != nil {
		return BookmarkView{}, err
	}
	if err := b.Update(in.toDomain()); err != nil {
		return BookmarkView{}, err
	}
	if err := uc.d.Repo.SaveBookmark(ctx, b); err != nil {
		return BookmarkView{}, fmt.Errorf("save bookmark: %w", err)
	}
	return toView(ctx, b), nil
}

type deleteBookmark struct{ d Deps }

// NewDeleteBookmark returns the DeleteBookmark use case.
func NewDeleteBookmark(d Deps) DeleteBookmark { return &deleteBookmark{d: d} }

// Execute implements DeleteBookmark.
func (uc *deleteBookmark) Execute(ctx context.Context, id string) error {
	if _, err := uc.d.Repo.GetBookmark(ctx, id); err != nil {
		return err
	}
	if err := uc.d.Repo.DeleteBookmark(ctx, id); err != nil {
		return fmt.Errorf("delete bookmark: %w", err)
	}
	return nil
}

func ensureFolder(ctx context.Context, repo Repository, id string) error {
	if strings.TrimSpace(id) == "" {
		return nil
	}
	_, err := repo.GetFolder(ctx, id)
	return err
}
