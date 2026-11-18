package usecase

import (
	"context"

	"waylight/internal/modules/study/domain"
	"waylight/internal/shared/kernel"
)

// SeedDemo adds the modules of a part-time computer science degree to the
// public demo: some passed with grades, some in progress.
func SeedDemo(ctx context.Context, repo Repository, ids kernel.IDGenerator) error {
	existing, err := repo.List(ctx)
	if err != nil || len(existing) > 0 {
		return err
	}
	uc := New(repo, ids)
	grade := func(g float64) *float64 { return &g }
	for _, in := range []domain.ModuleInput{
		{Name: "Programming 1", ECTS: 10, Semester: 1, Status: domain.StatusPassed, Grade: grade(1.7)},
		{Name: "Mathematics for Computer Science", ECTS: 10, Semester: 1, Status: domain.StatusPassed, Grade: grade(2.3)},
		{Name: "Databases", ECTS: 5, Semester: 2, Status: domain.StatusPassed, Grade: grade(1.3)},
		{Name: "Software Engineering", ECTS: 10, Semester: 2, Status: domain.StatusPassed, Grade: grade(2.0)},
		{Name: "Computer Networks", ECTS: 5, Semester: 3, Status: domain.StatusInProgress},
		{Name: "IT Security", ECTS: 5, Semester: 3, Status: domain.StatusInProgress},
		{Name: "Bachelor's thesis", ECTS: 12, Semester: 6, Status: domain.StatusOpen},
	} {
		if _, err := uc.Create(ctx, in); err != nil {
			return err
		}
	}
	_, err = uc.SetTotalECTS(ctx, 180)
	return err
}
