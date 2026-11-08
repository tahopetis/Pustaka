package testutils

import (
	"strings"
	"testing"
)

func TestTestDatabaseConfig(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		allowUnsafe bool
		wantSkip    bool
		wantDB      string
		wantErr     string
	}{
		{
			name:     "unset skips",
			raw:      "",
			wantSkip: true,
		},
		{
			name:     "whitespace only skips",
			raw:      "   \n",
			wantSkip: true,
		},
		{
			name:     "test database is accepted",
			raw:      "postgres://user:pass@localhost:5433/porta_test?sslmode=disable",
			wantDB:   "porta_test",
			wantSkip: false,
		},
		{
			name:     "test substring anywhere in the name is accepted",
			raw:      "postgres://user:pass@localhost:5433/my_testdb",
			wantDB:   "my_testdb",
			wantSkip: false,
		},
		{
			name:    "non-test database is rejected",
			raw:     "postgres://user:pass@localhost:5433/porta",
			wantErr: "must contain \"test\"",
		},
		{
			name:        "non-test database allowed with override",
			raw:         "postgres://user:pass@localhost:5433/porta",
			allowUnsafe: true,
			wantDB:      "porta",
		},
		{
			name:    "invalid DSN is rejected",
			raw:     "not-a-dsn",
			wantErr: "invalid TEST_DATABASE_URL",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg, skip, err := testDatabaseConfig(tc.raw, tc.allowUnsafe)

			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected an error containing %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if skip != tc.wantSkip {
				t.Fatalf("skip = %v, want %v", skip, tc.wantSkip)
			}
			if tc.wantDB != "" {
				if cfg == nil {
					t.Fatal("expected a pool config, got nil")
				}
				if got := cfg.ConnConfig.Database; got != tc.wantDB {
					t.Fatalf("connected database = %q, want %q", got, tc.wantDB)
				}
			}
		})
	}
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
