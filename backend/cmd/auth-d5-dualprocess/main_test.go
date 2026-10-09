package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRefuseUnsafeDSN(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		host    string
		port    string
		dbName  string
		wantErr bool
	}{
		{name: "allow disposable", host: "ae-auth-fix-disposable-pg", port: "5432", dbName: "auth_d1_db", wantErr: false},
		{name: "allow localhost auth_fix_db_test", host: "127.0.0.1", port: "5432", dbName: "auth_fix_db_test", wantErr: false},
		{name: "refuse shared ekarte_db", host: "ae-auth-fix-disposable-pg", port: "5432", dbName: "ekarte_db", wantErr: true},
		{name: "refuse host db", host: "db", port: "5432", dbName: "auth_d1_db", wantErr: true},
		{name: "refuse 15432", host: "ae-auth-fix-disposable-pg", port: "15432", dbName: "auth_d1_db", wantErr: true},
		{name: "refuse stg-looking name", host: "ae-auth-fix-disposable-pg", port: "5432", dbName: "stg_auth", wantErr: true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := refuseUnsafeDSN(tc.host, tc.port, tc.dbName)
			if tc.wantErr && err == nil {
				t.Fatalf("expected refuse for host=%s port=%s name=%s", tc.host, tc.port, tc.dbName)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected refuse: %v", err)
			}
		})
	}
}

// Validate the launcher's checked-in target without starting the DB harness.
func TestHarnessScriptPathResolvesCheckedInRunner(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test's repository")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "..", "..", ".."))
	script := harnessScriptPath(root)
	info, err := os.Stat(script)
	if err != nil {
		t.Fatalf("launcher target does not exist: %v", err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
		t.Fatal("launcher target must be a regular executable file")
	}
}
