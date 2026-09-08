"""Exercise staging/checking with synthetic bundles; no database or real CSVs."""

import json
import os
from pathlib import Path
import shutil
import stat
import subprocess
import tempfile
import unittest


ROOT = Path(__file__).resolve().parent.parent


def isolated_env(extra=None):
    env = dict(os.environ)
    for key in (
        "GIT_DIR",
        "GIT_WORK_TREE",
        "GIT_INDEX_FILE",
        "GIT_OBJECT_DIRECTORY",
        "GIT_COMMON_DIR",
        "GIT_PREFIX",
        "GIT_TEMPLATE_DIR",
        "GIT_CEILING_DIRECTORIES",
    ):
        env.pop(key, None)
    if extra:
        env.update(extra)
    return env


class AccountCSVLayoutTest(unittest.TestCase):
    def test_make_account_source_with_trailing_slash(self):
        with tempfile.TemporaryDirectory(prefix="ae-account-make-") as temp:
            root = Path(temp).resolve()
            source = root / "backend/migrations/seeds/_old_db_handoff/fixture"
            accounts = root / "backend/migrations/seeds/002_master/accounts/_old_db_handoff/fixture"
            accounts.mkdir(parents=True)
            env = isolated_env()
            env.pop("CSV_IMPORT_ACCOUNT_SOURCE_DIR", None)
            makefile = Path(__file__).resolve().parent.parent / "Makefile"
            for suffix in ["", "/", "//"]:
                env["CSV_IMPORT_SOURCE_DIR"] = str(source) + suffix
                result = subprocess.run(
                    ["make", "-s", "-f", str(makefile), "-f", "-", "account-layout-print"],
                    input="account-layout-print:\n\t@printf '%s\\n' '$(CSV_IMPORT_ACCOUNT_SOURCE_DIR)'\n",
                    cwd=root, env=env, capture_output=True, text=True,
                )
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(result.stdout.strip(), str(accounts))

    def test_staging_normalizes_account_directory_and_check_rejects_public_mode(self):
        with tempfile.TemporaryDirectory(prefix="ae-account-layout-") as temp:
            root = Path(temp) / "repo"
            scripts = root / "scripts"
            scripts.mkdir(parents=True)
            git_env = isolated_env()
            subprocess.run(["git", "init", "-q", str(root)], check=True, env=git_env)
            for name in ["stage-old-db-handoff.sh", "check-old-db-handoff.sh"]:
                shutil.copy2(Path(__file__).parent / name, scripts / name)
            source = Path(temp) / "source"
            account_dir = source / "accounts"
            account_dir.mkdir(parents=True)
            account_dir.chmod(0o755)
            (account_dir / "staffs.csv").write_text("id\n", encoding="utf-8")
            for index in range(20):
                (source / f"table_{index}.csv").write_text("id\n", encoding="utf-8")
            (source / "manifest.json").write_text(json.dumps({
                "clinicCode": "fixture", "sourceRunId": "fixture-run",
                "tables": [{} for _ in range(21)],
            }), encoding="utf-8")
            env = isolated_env({"CLINIC_CODE": "fixture", "MIGRATION_RUN_ID": "fixture-run",
                                 "CSV_IMPORT_SOURCE_DIR": str(source)})
            staged = subprocess.run(["bash", str(scripts / "stage-old-db-handoff.sh")],
                                    env=env, capture_output=True, text=True)
            self.assertEqual(staged.returncode, 0, staged.stderr)
            dest = root / "backend/migrations/seeds/002_master/accounts/_old_db_handoff/fixture"
            clinical = root / "backend/migrations/seeds/_old_db_handoff/fixture"
            self.assertFalse((clinical / "accounts").exists())
            self.assertFalse((clinical / "staffs.csv").exists())
            self.assertEqual((dest / "staffs.csv").read_bytes(), (account_dir / "staffs.csv").read_bytes())
            self.assertEqual((clinical / "manifest.json").read_bytes(), (source / "manifest.json").read_bytes())
            self.assertEqual(len(list(clinical.glob("*.csv"))), 20)
            self.assertEqual(stat.S_IMODE((dest / "staffs.csv").stat().st_mode), 0o600)
            self.assertEqual(stat.S_IMODE(dest.stat().st_mode), 0o700)
            command = ["bash", str(scripts / "check-old-db-handoff.sh")]
            checked = subprocess.run(command, env=env, capture_output=True, text=True)
            self.assertEqual(checked.returncode, 0, checked.stderr)
            # Restaging the split repository layout automatically selects central staff.
            restaged = subprocess.run(["bash", str(scripts / "stage-old-db-handoff.sh")],
                                     env=dict(env, CSV_IMPORT_SOURCE_DIR=str(clinical)),
                                     capture_output=True, text=True)
            self.assertEqual(restaged.returncode, 0, restaged.stderr)
            self.assertEqual((dest / "staffs.csv").read_bytes(), b"id\n")
            # A failed second promotion restores both previous destinations.
            (account_dir / "staffs.csv").write_text("id\nnew\n", encoding="utf-8")
            (source / "table_0.csv").write_text("id\nnew\n", encoding="utf-8")
            executable_dir = Path(temp) / "bin"
            executable_dir.mkdir()
            mv = executable_dir / "mv"
            mv.write_text(
                '#!/usr/bin/env bash\n'
                'if [[ "$1" == *"002_master/accounts/_old_db_handoff/.stage-"*"/fixture" ]]; then\n'
                '  exit 42\n'
                'fi\n'
                'exec /bin/mv "$@"\n', encoding="utf-8")
            mv.chmod(0o700)
            failed = subprocess.run(["bash", str(scripts / "stage-old-db-handoff.sh")],
                                    env=dict(env, PATH=str(executable_dir) + os.pathsep + os.environ["PATH"]),
                                    capture_output=True, text=True)
            self.assertNotEqual(failed.returncode, 0)
            self.assertEqual((dest / "staffs.csv").read_bytes(), b"id\n")
            self.assertEqual((clinical / "table_0.csv").read_bytes(), b"id\n")
            self.assertEqual((clinical / "manifest.json").read_bytes(), (source / "manifest.json").read_bytes())
            # Flat producer bundles are accepted too.
            (account_dir / "staffs.csv").rename(source / "staffs.csv")
            account_dir.rmdir()
            flat = subprocess.run(["bash", str(scripts / "stage-old-db-handoff.sh")],
                                  env=env, capture_output=True, text=True)
            self.assertEqual(flat.returncode, 0, flat.stderr)
            self.assertEqual((dest / "staffs.csv").read_bytes(), b"id\nnew\n")
            dest.chmod(0o755)
            rejected = subprocess.run(command, env=env, capture_output=True, text=True)
            self.assertNotEqual(rejected.returncode, 0)
            self.assertIn("owner-only", rejected.stderr)
            dest.chmod(0o700)
            (dest / "staffs.csv").chmod(0o644)
            rejected_file = subprocess.run(command, env=env, capture_output=True, text=True)
            self.assertNotEqual(rejected_file.returncode, 0)
            self.assertIn("owner-only", rejected_file.stderr)
            (dest / "staffs.csv").chmod(0o600)
            (clinical / "accounts").mkdir()
            duplicate = subprocess.run(command, env=env, capture_output=True, text=True)
            self.assertNotEqual(duplicate.returncode, 0)
            self.assertIn("duplicate", duplicate.stderr)

    def test_staging_in_linked_worktree_with_explicit_account_source(self):
        with tempfile.TemporaryDirectory(prefix="ae-account-worktree-") as temp:
            root = Path(temp) / "repo"
            git_env = isolated_env()
            subprocess.run(["git", "init", "-q", str(root)], check=True, env=git_env)
            subprocess.run(["git", "-C", str(root), "-c", "user.name=Fixture", "-c",
                            "user.email=fixture@example.invalid", "commit", "--allow-empty", "-qm", "fixture"],
                           check=True, env=git_env)
            worktree = Path(temp) / "worktree"
            subprocess.run(["git", "-C", str(root), "worktree", "add", "--detach", "-q", str(worktree)],
                           check=True, env=git_env)
            scripts = worktree / "scripts"
            scripts.mkdir()
            for name in ["stage-old-db-handoff.sh", "check-old-db-handoff.sh"]:
                shutil.copy2(Path(__file__).parent / name, scripts / name)
            source = Path(temp) / "source"
            source.mkdir()
            accounts = Path(temp) / "accounts"
            accounts.mkdir()
            (accounts / "staffs.csv").write_bytes(b"id\r\n1\r\n")
            for index in range(20):
                (source / f"table_{index}.csv").write_bytes(b"id\r\n")
            (source / "manifest.json").write_text(json.dumps({
                "clinicCode": "fixture", "sourceRunId": "fixture-run",
                "tables": [{} for _ in range(21)],
            }), encoding="utf-8")
            env = isolated_env({"CLINIC_CODE": "fixture", "MIGRATION_RUN_ID": "fixture-run",
                                 "CSV_IMPORT_SOURCE_DIR": str(source),
                                 "CSV_IMPORT_ACCOUNT_SOURCE_DIR": str(accounts)})
            for name in ["stage-old-db-handoff.sh", "check-old-db-handoff.sh"]:
                result = subprocess.run(["bash", str(scripts / name)], env=env, capture_output=True, text=True)
                self.assertEqual(result.returncode, 0, result.stderr)
            staff = worktree / "backend/migrations/seeds/002_master/accounts/_old_db_handoff/fixture/staffs.csv"
            self.assertEqual(staff.read_bytes(), b"id\r\n1\r\n")

    def test_ignore_compose_and_env_keep_account_handoff_private(self):
        root = Path(__file__).resolve().parent.parent
        dockerignore = (root / "backend/.dockerignore").read_text(encoding="utf-8")
        self.assertIn("002_master/accounts/_old_db_handoff/", dockerignore)
        compose = (root / "docker-compose.yml").read_text(encoding="utf-8")
        self.assertIn("/migration-accounts", compose)
        makefile = (root / "Makefile").read_text(encoding="utf-8")
        self.assertIn("CSV_IMPORT_ACCOUNT_SOURCE_DIR", makefile)
        example = (root / ".env.example").read_text(encoding="utf-8")
        self.assertIn("# SEEDLOGIN_OPERATOR_EMAIL=", example)
        self.assertNotRegex(example, r"^SEEDLOGIN_OPERATOR_PASSWORD=.+$", "operator password must stay commented")
        groups = root / "backend/migrations/seeds/002_master/accounts/permission_groups.csv"
        self.assertTrue(groups.is_file())
        self.assertIn("執行", groups.read_text(encoding="utf-8"))


if __name__ == "__main__":
    unittest.main()
