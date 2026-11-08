package testutils

import (
	"strings"
	"testing"
)

func TestDatabaseURLFromEnv(t *testing.T) {
	t.Run("unset returns empty", func(t *testing.T) {
		got := databaseURLFromEnv(func(string) string { return "" })
		if got != "" {
			t.Fatalf("expected empty URL, got %q", got)
		}
	})

	t.Run("returns the configured DSN", func(t *testing.T) {
		const dsn = "postgres://testuser:postgres@localhost:5432/testdb?sslmode=disable"
		got := databaseURLFromEnv(func(key string) string {
			if key != testDatabaseURLEnv {
				t.Fatalf("unexpected env key %q", key)
			}
			return dsn
		})
		if got != dsn {
			t.Fatalf("expected %q, got %q", dsn, got)
		}
	})

	t.Run("trims surrounding whitespace", func(t *testing.T) {
		got := databaseURLFromEnv(func(string) string {
			return "  postgres://user@host:5432/db  \n"
		})
		want := "postgres://user@host:5432/db"
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}

// TestSetupTestDBSkipsWithoutURL pins the contract that switched this helper
// away from a Docker daemon: no TEST_DATABASE_URL means skip, not fatalf.
func TestSetupTestDBSkipsWithoutURL(t *testing.T) {
	var skipped bool

	t.Run("inner", func(t *testing.T) {
		t.Setenv(testDatabaseURLEnv, "")
		defer func() { skipped = t.Skipped() }()
		SetupTestDB(t)
	})

	if !skipped {
		t.Fatal("expected SetupTestDB to skip the test when TEST_DATABASE_URL is unset")
	}
}

func TestSkipMessageMentionsEnvVar(t *testing.T) {
	msg := skipMessage()
	if !strings.Contains(msg, testDatabaseURLEnv) {
		t.Fatalf("skip message should name %s, got %q", testDatabaseURLEnv, msg)
	}
}
