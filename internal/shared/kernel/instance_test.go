package kernel

import (
	"context"
	"testing"
)

func TestInstanceNameFallsBackToTheProduct(t *testing.T) {
	var none InstanceName
	if none.Of(context.Background()) != ProductName {
		t.Fatal("nil")
	}
	empty := InstanceName(func(context.Context) string { return "" })
	org := InstanceName(func(context.Context) string { return "Example GmbH" })
	if empty.Of(context.Background()) != ProductName || org.Of(context.Background()) != "Example GmbH" {
		t.Fatal("values")
	}
}
