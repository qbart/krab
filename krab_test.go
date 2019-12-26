package krab

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestMigration(t *testing.T) {
	t.Errorf("Not implemented")
}
