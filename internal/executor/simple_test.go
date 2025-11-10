package executor

import (
	"testing"
	"panoptic/internal/config"
)

func TestSimpleAction(t *testing.T) {
	action := config.Action{
		Name: "test",
		Type: "wait",
		WaitTime: 1,
	}
	
	if action.Name != "test" {
		t.Errorf("Expected test, got %s", action.Name)
	}
}