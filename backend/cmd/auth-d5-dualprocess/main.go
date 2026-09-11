// Command auth-d5-dualprocess validates disposable DSN allowlisting, then
// delegates to scripts/auth-d5-dualprocess.sh for the dual-process harness.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	host := envOr("D5_DB_HOST", "ae-auth-fix-disposable-pg")
	port := envOr("D5_DB_PORT", "5432")
	name := envOr("D5_DB_NAME", "auth_d1_db")
	if err := refuseUnsafeDSN(host, port, name); err != nil {
		fmt.Fprintf(os.Stderr, "REFUSE: %v\n", err)
		os.Exit(2)
	}

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintln(os.Stderr, "unable to resolve source path")
		os.Exit(1)
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	script := filepath.Join(root, "scripts", "auth-d5-dualprocess.sh")
	if filepath.Base(script) != "auth-d5-dualprocess.sh" {
		fmt.Fprintln(os.Stderr, "unexpected harness script path")
		os.Exit(1)
	}
	//nolint:gosec // G204: fixed repo-local script resolved from source path; args are intentionally forwarded.
	cmd := exec.Command(script, os.Args[1:]...)
	cmd.Dir = root
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			os.Exit(ee.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "harness failed: %v\n", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func refuseUnsafeDSN(host, port, name string) error {
	lower := strings.ToLower(host + "/" + name)
	switch {
	case name == "ekarte_db":
		return fmt.Errorf("DB_NAME ekarte_db is shared app DB")
	case port == "15432":
		return fmt.Errorf("port 15432 is reserved/unsafe")
	case host == "db":
		return fmt.Errorf("host db is compose shared DB")
	case host == "animalekarte-db-1" || host == "old-db-postgres":
		return fmt.Errorf("host %s is not disposable auth DB", host)
	case strings.Contains(lower, "stg") || strings.Contains(lower, "prod"):
		return fmt.Errorf("host/name looks like stg/prod")
	case host != "ae-auth-fix-disposable-pg" && host != "127.0.0.1" && host != "localhost":
		return fmt.Errorf("unexpected DB_HOST=%s", host)
	case name != "auth_d1_db" && name != "auth_fix_db" && name != "auth_fix_db_test":
		return fmt.Errorf("unexpected DB_NAME=%s", name)
	}
	return nil
}
