package gateway

import (
	"cmp"
	"slices"
	"strings"

	"trisse-home/internal/modules/todos/domain"
)

func slicesSortTodos(todos []domain.Todo) {
	slices.SortStableFunc(todos, func(a, b domain.Todo) int {
		ad, bd := a.Due(), b.Due()
		switch {
		case ad.IsZero() != bd.IsZero():
			if ad.IsZero() {
				return 1
			}
			return -1
		case !ad.Equal(bd):
			return ad.Compare(bd)
		}
		return cmp.Or(a.CreatedAt().Compare(b.CreatedAt()), strings.Compare(a.ID(), b.ID()))
	})
}
