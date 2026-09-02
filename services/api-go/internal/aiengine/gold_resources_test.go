package aiengine

import (
	"testing"
)

func TestGoldResourceLookup(t *testing.T) {
	lookup := GetGoldResourceLookup()

	if len(lookup.Roles) != 11 {
		t.Fatalf("expected 11 roles in Gold Tier lookup, got %d", len(lookup.Roles))
	}

	// Verify data_engineer is not present
	if _, ok := lookup.GetRole("data_engineer"); ok {
		t.Fatalf("data_engineer should not be accessible in Gold Tier lookup")
	}

	// Verify full_stack role and module 1
	fsRole, ok := lookup.GetRole("full_stack")
	if !ok {
		t.Fatalf("expected full_stack role to be present")
	}

	if len(fsRole.Modules) != 10 {
		t.Errorf("expected 10 modules for full_stack, got %d", len(fsRole.Modules))
	}

	mod, ok := lookup.GetGoldModuleResources("full_stack", "1")
	if !ok {
		t.Fatalf("expected module 1 for full_stack")
	}

	if len(mod.Resources.Video) == 0 {
		t.Errorf("expected video resources in module 1")
	}
	if len(mod.Resources.Documentation) == 0 {
		t.Errorf("expected documentation resources in module 1")
	}
	if len(mod.Resources.HandsOn) == 0 {
		t.Errorf("expected hands_on resources in module 1")
	}

	// Verify machine_learning role and module 1 lookup by various queries
	mlMod, ok := lookup.GetGoldModuleResources("machine_learning", "1")
	if !ok {
		t.Fatalf("expected module 1 for machine_learning by '1'")
	}
	if mlMod.ModuleName != "Programming Fundamentals for ML" {
		t.Errorf("unexpected module name: %s", mlMod.ModuleName)
	}

	mlMod2, ok := lookup.GetGoldModuleResources("machine_learning", "ml_01_programming_fundamentals")
	if !ok {
		t.Fatalf("expected module 1 for machine_learning by legacy ID 'ml_01_programming_fundamentals'")
	}
	if mlMod2.ModuleNumber != 1 {
		t.Errorf("unexpected module number: %d", mlMod2.ModuleNumber)
	}

	if len(mlMod.Resources.Video) != 2 {
		t.Errorf("expected 2 video resources for ML Module 1, got %d", len(mlMod.Resources.Video))
	}
	if len(mlMod.Resources.Documentation) != 1 {
		t.Errorf("expected 1 doc resource for ML Module 1, got %d", len(mlMod.Resources.Documentation))
	}
	if len(mlMod.Resources.HandsOn) != 2 {
		t.Errorf("expected 2 hands_on resources for ML Module 1, got %d", len(mlMod.Resources.HandsOn))
	}
}
