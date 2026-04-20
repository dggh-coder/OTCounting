package main

import (
	"strings"
	"testing"
)

func TestDatabaseURLFromEnvDefaultsToOtUser(t *testing.T) {
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_PASSWORD_FILE", "")

	dsn, err := databaseURLFromEnv()
	if err != nil {
		t.Fatalf("databaseURLFromEnv error: %v", err)
	}
	if !strings.Contains(dsn, "postgres://ot_user:secret@opengauss:5432/postgres") {
		t.Fatalf("expected DSN to use ot_user default user, got: %s", dsn)
	}
}

func TestDatabaseURLFromEnvRespectsDBUser(t *testing.T) {
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_USER", "user_ot_backend")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_PASSWORD_FILE", "")

	dsn, err := databaseURLFromEnv()
	if err != nil {
		t.Fatalf("databaseURLFromEnv error: %v", err)
	}
	if !strings.Contains(dsn, "postgres://user_ot_backend:secret@") {
		t.Fatalf("expected DSN to use DB_USER value, got: %s", dsn)
	}
}
