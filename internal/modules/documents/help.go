package documents

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "documents",
			Shows: kernel.NewText("Your receipts and warranty papers, with the warranty status of each product."),
			SetUp: kernel.NewText("Add a document with the product, the purchase day and the end of the warranty. Attach a photo or a PDF of up to 10 MB."),
			Docs:  "modules/documents/",
		},
	}
}
