package phonelist

import (
	"context"

	"waylight/internal/modules/phonelist/domain"
	"waylight/internal/shared/kernel"
)

// SeedDemo implements modules.DemoSeeder: a small team directory with
// made-up names, numbers and addresses.
func (m *Module) SeedDemo(ctx context.Context) error {
	ctx = kernel.WithPrincipal(ctx, kernel.AdminPrincipal())
	existing, err := m.uc.List(ctx)
	if err != nil || len(existing) > 0 {
		return err
	}
	for _, in := range []domain.EntryInput{
		{Name: "Anna Schröder", Role: "Managing director", Team: "Management", Phone: "+49 441 555 0100", Email: "anna@example.org"},
		{Name: "Jonas Meyer", Role: "IT administrator", Team: "IT", Phone: "+49 441 555 0120", Email: "jonas@example.org"},
		{Name: "Mira Hoffmann", Role: "Developer", Team: "IT", Phone: "+49 441 555 0121", Email: "mira@example.org"},
		{Name: "Tobias Brandt", Role: "Help desk", Team: "IT", Phone: "+49 441 555 0122", Email: "helpdesk@example.org"},
		{Name: "Lea Wagner", Role: "Office manager", Team: "Office", Phone: "+49 441 555 0130", Email: "lea@example.org"},
		{Name: "Reception", Role: "Front desk", Team: "Office", Phone: "+49 441 555 0000", Email: "hello@example.org"},
		{Name: "Felix Krüger", Role: "Sales", Team: "Sales", Phone: "+49 441 555 0140", Email: "felix@example.org"},
		{Name: "Sara Becker", Role: "Customer success", Team: "Sales", Phone: "+49 441 555 0141", Email: "sara@example.org"},
		{Name: "Nils Peters", Role: "Accounting", Team: "Finance", Phone: "+49 441 555 0150", Email: "finance@example.org"},
		{Name: "Facility service", Role: "Building and keys", Team: "Office", Phone: "+49 441 555 0190"},
	} {
		if _, err := m.uc.Add(ctx, in); err != nil {
			return err
		}
	}
	return nil
}
