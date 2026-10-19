#!/usr/bin/env python3
"""Export/restore the STG data that must survive a schema rebuild.

This utility is USER-RUN. It connects with psql using an env file, exports
selected tables to CSV, and restores them after cmd/migrate. It never prints
the password and never writes secrets to the manifest.

Export:
  python3 scripts/old-db/stg-preserve.py export

Restore (dry-run by default):
  python3 scripts/old-db/stg-preserve.py restore --dir <export-dir>
  python3 scripts/old-db/stg-preserve.py restore --dir <export-dir> --execute

Verify:
  python3 scripts/old-db/stg-preserve.py verify --dir <export-dir>
"""
from __future__ import annotations

import argparse
import csv
import hashlib
import json
import os
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
DEFAULT_ENV_FILE = ROOT / "scripts/old-db/stg-uat-old-db-handoff.local.env"
DEFAULT_OUT_ROOT = ROOT / "sensitive-local/stg-preserve"

# Restore order matters because of foreign keys.
# The second item is the ON CONFLICT target used by restore.
TABLES: list[tuple[str, str]] = [
    ("clinics", "(id)"),
    ("clinic_settings", "(clinic_id)"),
    ("clinic_holidays", "(id)"),
    ("closing_special_periods", "(id)"),
    ("occupations", "(id)"),
    ("accounts", "(id)"),
    ("permission_groups", "(id)"),
    ("permission_group_rules", "(id)"),
    ("staffs", "(id)"),
    # pkey is the serial id; the (staff_id, clinic_id) pair collision is
    # resolved by TARGET_PRE_DELETES before the INSERT.
    ("staff_clinic_assignments", "(id)"),
    ("staff_permission_groups", "(staff_id, group_id)"),
    ("manual_articles", "(id)"),
    ("manual_article_versions", "(id)"),
    ("support_bug_reports", "(id)"),
    ("support_chat_messages", "(id)"),
]

STRICT_TABLES = {
    "manual_articles",
    "manual_article_versions",
    "support_bug_reports",
    "support_chat_messages",
}

# SQL statements applied to the stage table before the INSERT, per table.
# They resolve collisions between the seed-managed demo identities (recreated
# by the login seed with @example.test emails) and the exported copies, so the
# restore converges without violating accounts_email_key / FK constraints.
STAGE_FILTERS: dict[str, list[str]] = {
    # Demo accounts keep their old serial ids in the export; the login seed
    # already recreated them under new ids. Skip the exported copies.
    "accounts": [
        "DELETE FROM stage WHERE email LIKE '%@example.test'",
    ],
    # Demo staffs keep their fixed band ids across rebuilds, but their
    # account_id pointed at the skipped demo accounts. Remap to the seed
    # accounts by the deterministic staff-id email pattern. Staffs whose
    # account is gone are dropped — except those referenced by preserved
    # support rows, which are kept with account_id = NULL (column is
    # nullable) so bug reports keep their reporter identity.
    "staffs": [
        (
            "UPDATE stage SET account_id = na.id FROM public.accounts na "
            "WHERE na.email = 'stg-staff-' || stage.id::text || '@example.test'"
        ),
        (
            "DELETE FROM stage WHERE account_id IS NOT NULL AND NOT EXISTS ("
            "SELECT 1 FROM public.accounts a WHERE a.id = stage.account_id)"
        ),
    ],
    "staff_clinic_assignments": [
        "DELETE FROM stage WHERE staff_id NOT IN (SELECT id FROM public.staffs)",
    ],
    "staff_permission_groups": [
        "DELETE FROM stage WHERE staff_id NOT IN (SELECT id FROM public.staffs)",
    ],
    # Drop only rows referencing staffs that could not be restored; the rest
    # of the preserved content is kept (verify reports the final counts).
    "support_bug_reports": [
        (
            "DELETE FROM stage WHERE reporter_staff_id IS NOT NULL AND NOT EXISTS ("
            "SELECT 1 FROM public.staffs s WHERE s.id = stage.reporter_staff_id)"
        ),
    ],
    "support_chat_messages": [
        (
            "DELETE FROM stage WHERE staff_id IS NOT NULL AND NOT EXISTS ("
            "SELECT 1 FROM public.staffs s WHERE s.id = stage.staff_id)"
        ),
    ],
}

# SQL statements executed against the target table (stage is available as a
# temp table) right before the INSERT. staff_clinic_assignments has a serial
# pkey plus a (staff_id, clinic_id) unique constraint; an exported row whose
# id differs from an existing row with the same pair would violate the unique
# constraint. Remove pair-matching rows first; the login seed re-creates its
# own assignments on the next migrate.
TARGET_PRE_DELETES: dict[str, list[str]] = {
    "staff_clinic_assignments": [
        (
            "DELETE FROM public.staff_clinic_assignments t USING stage s "
            "WHERE t.staff_id = s.staff_id AND t.clinic_id = s.clinic_id"
        ),
    ],
}

FK_CHECKS: list[tuple[str, str]] = [
    ("support_bug_reports", "reporter_staff_id"),
    ("support_bug_reports", "clinic_id"),
    ("support_chat_messages", "staff_id"),
    ("support_chat_messages", "clinic_id"),
    ("manual_articles", "updated_by_staff_id"),
    ("manual_article_versions", "edited_by_staff_id"),
    ("clinics", "company_id"),
    ("occupations", "clinic_id"),
    ("staffs", "account_id"),
    ("staffs", "occupation_id"),
    ("staffs", "clinic_id"),
    ("staff_clinic_assignments", "staff_id"),
    ("staff_clinic_assignments", "clinic_id"),
    ("staff_permission_groups", "staff_id"),
    ("staff_permission_groups", "group_id"),
    ("permission_groups", "clinic_id"),
    ("permission_group_rules", "group_id"),
    ("clinic_settings", "clinic_id"),
    ("clinic_holidays", "clinic_id"),
    ("closing_special_periods", "clinic_id"),
]


def die(message: str) -> None:
    print(f"BLOCKED: {message}", file=sys.stderr)
    raise SystemExit(1)


def read_env(path: Path) -> dict[str, str]:
    if not path.exists():
        die(f"env file not found: {path}")
    mode = path.stat().st_mode & 0o777
    if mode != 0o600:
        die(f"env file must be 0600: {path} (current {oct(mode)})")
    values: dict[str, str] = {}
    for raw in path.read_text().splitlines():
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        if line.startswith("export "):
            line = line[len("export "):]
        key, sep, value = line.partition("=")
        if not sep:
            continue
        values[key.strip()] = value.strip().strip("'\"")
    for key in ("DB_HOST", "DB_USER", "DB_PASSWORD"):
        if not values.get(key):
            die(f"env file is missing {key}")
    return values


def psql_env(cfg: dict[str, str]) -> dict[str, str]:
    env = os.environ.copy()
    env.update(
        {
            "PGHOST": cfg["DB_HOST"],
            "PGPORT": cfg.get("DB_PORT", "5432"),
            "PGUSER": cfg["DB_USER"],
            "PGPASSWORD": cfg["DB_PASSWORD"],
            "PGDATABASE": cfg.get("DB_NAME", "postgres"),
            "PGSSLMODE": cfg.get("DB_SSL_MODE", "verify-full"),
            # The app role is not BYPASSRLS; use the app's ops bypass contract.
            "PGOPTIONS": "-c app.bypass_rls=on",
        }
    )
    ssl_root = cfg.get("DB_SSL_ROOT_CERT", "")
    if ssl_root == "system":
        # libpq <17 does not understand sslrootcert=system.
        ssl_root = "/etc/ssl/cert.pem"
    if ssl_root:
        env["PGSSLROOTCERT"] = ssl_root
    return env


def run_psql(env: dict[str, str], sql: str | None = None, script: str | None = None) -> str:
    cmd = ["psql", "-X", "-q", "-v", "ON_ERROR_STOP=1", "-At"]
    if script is None:
        cmd += ["-c", sql or ""]
        result = subprocess.run(cmd, env=env, text=True, capture_output=True)
    else:
        cmd += ["-f", "-"]
        result = subprocess.run(cmd, env=env, input=script, text=True, capture_output=True)
    if result.returncode != 0:
        die((result.stderr or result.stdout).strip())
    return result.stdout


def quote_ident(value: str) -> str:
    return '"' + value.replace('"', '""') + '"'


def table_columns(env: dict[str, str], table: str) -> list[str]:
    sql = f"""
        SELECT column_name
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = '{table}'
          AND is_generated = 'NEVER'
        ORDER BY ordinal_position
    """
    return [line for line in run_psql(env, sql).splitlines() if line]


def table_count(env: dict[str, str], table: str) -> int:
    return int(run_psql(env, f"SELECT count(*) FROM public.{quote_ident(table)}").strip())


def content_hash(env: dict[str, str], table: str) -> str:
    sql = (
        "SELECT md5(COALESCE(string_agg(t::text, E'\\n' ORDER BY t::text), '')) "
        f"FROM public.{quote_ident(table)} t"
    )
    return run_psql(env, sql).strip()


def csv_content_diffs(
    env: dict[str, str],
    export_dir: Path,
    table: str,
    columns: list[str],
    compare_columns: list[str] | None = None,
) -> list[str]:
    """Compare a restored strict table against its export CSV, row by row.

    CSV export cannot distinguish NULL from the empty string in text
    columns, so every column is compared as COALESCE(col::text, ''). Any
    remaining difference is reported per row with the differing column
    names (post-restore app writes show up here too). compare_columns may
    restrict which columns are compared while the stage still loads the
    full CSV column set.
    """
    csv_path = export_dir / f"{table}.csv"
    csv_literal = str(csv_path).replace("'", "''")
    column_list = ", ".join(quote_ident(column) for column in columns)
    table_name = quote_ident(table)
    compared = compare_columns if compare_columns is not None else columns

    selects: list[str] = []
    for column in compared:
        quoted = quote_ident(column)
        selects.append(
            "SELECT t.id, '%s' AS col FROM public.%s t JOIN stage s ON s.id = t.id "
            "WHERE COALESCE(t.%s::text, '') IS DISTINCT FROM COALESCE(s.%s::text, '')"
            % (column, table_name, quoted, quoted)
        )
    script = (
        f"CREATE TEMP TABLE stage (LIKE public.{table_name} INCLUDING DEFAULTS);\n"
        f"\\copy stage ({column_list}) FROM '{csv_literal}' WITH (FORMAT csv, HEADER true)\n"
        "SELECT t.id::text || ' missing-in-target' "
        f"FROM stage s LEFT JOIN public.{table_name} t ON t.id = s.id WHERE t.id IS NULL\n"
        "UNION ALL "
        "SELECT s.id::text || ' missing-in-stage' "
        f"FROM public.{table_name} t LEFT JOIN stage s ON s.id = t.id WHERE s.id IS NULL\n"
        "UNION ALL "
        "SELECT id::text || ':' || string_agg(col, ',') FROM (\n  "
        + "\n  UNION ALL ".join(selects)
        + "\n) d GROUP BY id;\n"
        "DROP TABLE stage;\n"
    )
    out = run_psql(env, script=script)
    return [line.strip() for line in out.splitlines() if line.strip()]


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def guard_target(cfg: dict[str, str]) -> None:
    app_env = os.environ.get("APP_ENV", "").lower()
    host = cfg.get("DB_HOST", "").lower()
    if app_env in {"production", "prod"} or "prod" in host:
        die("refusing to touch a production-looking target")


def export(args: argparse.Namespace) -> None:
    cfg = read_env(Path(args.env_file))
    guard_target(cfg)
    out_dir = Path(args.out_dir) if args.out_dir else DEFAULT_OUT_ROOT / datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    out_dir.mkdir(parents=True, exist_ok=False)
    out_dir.chmod(0o700)
    env = psql_env(cfg)

    tables = []
    for table, _conflict in TABLES:
        columns = table_columns(env, table)
        if not columns:
            die(f"table not found or has no columns: {table}")
        csv_path = out_dir / f"{table}.csv"
        select_list = ", ".join(quote_ident(column) for column in columns)
        csv_literal = str(csv_path).replace("'", "''")
        run_psql(
            env,
            script=(
                f"\\copy (SELECT {select_list} FROM public.{quote_ident(table)}) "
                f"TO '{csv_literal}' WITH (FORMAT csv, HEADER true)\n"
            ),
        )
        rows = table_count(env, table)
        csv_path.chmod(0o600)
        tables.append(
            {
                "name": table,
                "columns": columns,
                "rows": rows,
                "bytes": csv_path.stat().st_size,
                "sha256": sha256_file(csv_path),
                "content_hash": content_hash(env, table),
            }
        )
        print(f"exported {table}: rows={rows}")

    manifest = {
        "format_version": 1,
        "created_at": datetime.now(timezone.utc).isoformat(),
        "source": {
            "host": cfg["DB_HOST"],
            "database": cfg.get("DB_NAME", "postgres"),
            "app_env": os.environ.get("APP_ENV", ""),
        },
        "tables": tables,
    }
    manifest_path = out_dir / "manifest.json"
    manifest_path.write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n")
    manifest_path.chmod(0o600)
    print(f"manifest: {manifest_path}")


def load_manifest(export_dir: Path) -> dict:
    path = export_dir / "manifest.json"
    if not path.exists():
        die(f"manifest not found: {path}")
    manifest = json.loads(path.read_text())
    if manifest.get("format_version") != 1:
        die(f"unsupported manifest format: {manifest.get('format_version')}")
    return manifest


def validate_export(export_dir: Path, manifest: dict) -> None:
    names = [entry["name"] for entry in manifest["tables"]]
    expected = [table for table, _ in TABLES]
    if names != expected:
        die(f"manifest tables are not in restore order: {names}")
    for entry in manifest["tables"]:
        csv_path = export_dir / f"{entry['name']}.csv"
        if not csv_path.exists():
            die(f"missing CSV: {csv_path}")
        if sha256_file(csv_path) != entry["sha256"]:
            die(f"sha256 mismatch: {csv_path}")
        header = csv_path.read_text(errors="replace").splitlines()[0].split(",")
        if header != entry["columns"]:
            die(f"CSV header mismatch: {csv_path}")


def conflict_columns(conflict: str) -> set[str]:
    return {part.strip() for part in conflict.strip("()").split(",")}


SUPPORT_STAFF_REFERENCES: list[tuple[str, str]] = [
    ("support_bug_reports", "reporter_staff_id"),
    ("support_chat_messages", "staff_id"),
]


def support_referenced_staff_ids(export_dir: Path) -> set[int]:
    """Staff ids referenced by preserved support rows in the export CSVs.

    Those staffs are restored with account_id = NULL so the support rows keep
    their reporter/author identity even though the referenced accounts were
    seed-managed synthetic accounts skipped by the restore.
    """
    referenced: set[int] = set()
    for table, column in SUPPORT_STAFF_REFERENCES:
        path = export_dir / f"{table}.csv"
        if not path.exists():
            continue
        with path.open(newline="") as handle:
            for row in csv.DictReader(handle):
                value = (row.get(column) or "").strip()
                if value:
                    referenced.add(int(value))
    return referenced


def restore(args: argparse.Namespace) -> None:
    export_dir = Path(args.dir).resolve()
    manifest = load_manifest(export_dir)
    validate_export(export_dir, manifest)

    if not args.execute:
        print("DRY RUN: no database connection made.")
        for entry in manifest["tables"]:
            print(f"  {entry['name']}: {entry['rows']} rows")
        print("Re-run with --execute after reviewing the target.")
        return

    cfg = read_env(Path(args.env_file))
    guard_target(cfg)
    print(f"restore target: {cfg['DB_HOST']}/{cfg.get('DB_NAME', 'postgres')}")
    env = psql_env(cfg)

    report: dict[str, dict[str, int]] = {}
    for entry in manifest["tables"]:
        table = entry["name"]
        conflict = dict(TABLES)[table]
        columns = table_columns(env, table)
        if columns != entry["columns"]:
            die(f"column mismatch for {table}: export={entry['columns']} target={columns}")
        csv_path = export_dir / f"{table}.csv"
        csv_literal = str(csv_path).replace("'", "''")
        column_list = ", ".join(quote_ident(column) for column in columns)
        update_columns = [column for column in columns if column not in conflict_columns(conflict)]
        if update_columns:
            conflict_clause = "DO UPDATE SET " + ", ".join(
                f"{quote_ident(column)} = EXCLUDED.{quote_ident(column)}" for column in update_columns
            )
        else:
            conflict_clause = "DO NOTHING"
        script = (
            f"CREATE TEMP TABLE stage (LIKE public.{quote_ident(table)} INCLUDING DEFAULTS);\n"
            f"\\copy stage ({column_list}) FROM '{csv_literal}' WITH (FORMAT csv, HEADER true)\n"
            "SELECT count(*) FROM stage;\n"
        )
        filters = list(STAGE_FILTERS.get(table, []))
        if table == "staffs":
            referenced = sorted(support_referenced_staff_ids(export_dir))
            if referenced:
                id_list = ", ".join(str(value) for value in referenced)
                # Retain support-referenced staffs (account_id = NULL) before
                # the orphan-account delete drops the rest.
                keep_filter = (
                    "UPDATE stage SET account_id = NULL "
                    "WHERE account_id IS NOT NULL AND NOT EXISTS ("
                    "SELECT 1 FROM public.accounts a WHERE a.id = stage.account_id) "
                    f"AND stage.id IN ({id_list})"
                )
                filters.insert(max(len(filters) - 1, 0), keep_filter)
        for statement in filters:
            script += statement + ";\n"
        if filters:
            script += "SELECT count(*) FROM stage;\n"
        for statement in TARGET_PRE_DELETES.get(table, []):
            script += statement + ";\n"
        script += (
            f"INSERT INTO public.{quote_ident(table)} ({column_list})\n"
            f"SELECT {column_list} FROM stage\n"
            f"ON CONFLICT {conflict} {conflict_clause};\n"
            "DROP TABLE stage;\n"
        )
        out = run_psql(env, script=script)
        lines = [line for line in out.splitlines() if line.strip()]
        staged = int(lines[0]) if lines else 0
        kept = staged
        if filters:
            kept = int(lines[1]) if len(lines) > 1 else staged
            print(
                f"restored {table}: {entry['rows']} export rows, "
                f"{staged} staged, {staged - kept} skipped by restore filters"
            )
        else:
            print(f"restored {table}: {entry['rows']} rows")
        report[table] = {"staged": staged, "kept": kept, "skipped": staged - kept}

    report_path = export_dir / "restore-report.json"
    report_path.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n")
    os.chmod(report_path, 0o600)
    print(f"restore report written: {report_path}")

    sequence_tables = ", ".join(f"'{table}'" for table, _ in TABLES if table not in {"clinic_settings", "staff_permission_groups"})
    run_psql(
        env,
        script=f"""
DO $$
DECLARE
    table_name text;
    max_id bigint;
BEGIN
    FOREACH table_name IN ARRAY ARRAY[{sequence_tables}]
    LOOP
        IF pg_get_serial_sequence('public.' || table_name, 'id') IS NOT NULL THEN
            EXECUTE format('SELECT COALESCE(MAX(id), 1) FROM public.%I', table_name) INTO max_id;
            PERFORM setval(pg_get_serial_sequence('public.' || table_name, 'id'), max_id, true);
        END IF;
    END LOOP;
END $$;
""",
    )
    print("sequences advanced")
    print("NOTE: run cmd/migrate again to refresh demo login passwords, then run verify.")


def verify(args: argparse.Namespace) -> None:
    export_dir = Path(args.dir).resolve()
    manifest = load_manifest(export_dir)
    validate_export(export_dir, manifest)
    cfg = read_env(Path(args.env_file))
    guard_target(cfg)
    env = psql_env(cfg)

    # Restore filters may intentionally skip rows (seed-managed demo
    # identities, staffs whose accounts are gone). The report records what
    # the restore actually staged and kept per table.
    report: dict[str, dict[str, int]] = {}
    report_path = export_dir / "restore-report.json"
    if report_path.exists():
        report = json.loads(report_path.read_text())

    failures: list[str] = []
    for entry in manifest["tables"]:
        table = entry["name"]
        rows = table_count(env, table)
        entry_report = report.get(table)
        kept = entry_report["kept"] if entry_report else entry["rows"]
        if table in STRICT_TABLES:
            if kept == entry["rows"]:
                if rows != entry["rows"]:
                    failures.append(f"{table}: rows {rows} != export {entry['rows']}")
                diff_columns = list(entry["columns"])
                if getattr(args, "allow_sync_drift", False):
                    # The STG plane_sync scheduled job rewrites status /
                    # updated_at on preserved bug reports shortly after a
                    # restore; exclude those columns from the strict compare.
                    diff_columns = [
                        column for column in diff_columns if column not in {"status", "updated_at"}
                    ]
                    if diff_columns != list(entry["columns"]):
                        print(f"NOTE: {table}: status/updated_at excluded from the compare (--allow-sync-drift)")
                diffs = csv_content_diffs(
                    env, export_dir, table, entry["columns"], compare_columns=diff_columns
                )
                if diffs:
                    sample = "; ".join(diffs[:5])
                    failures.append(f"{table}: {len(diffs)} content diff(s) vs export ({sample})")
            else:
                if rows != kept:
                    failures.append(f"{table}: rows {rows} != restored kept {kept}")
                print(
                    f"NOTE: {table}: {entry['rows'] - kept} row(s) skipped by restore filters; "
                    "content comparison skipped for this table"
                )
        elif rows < kept:
            failures.append(f"{table}: rows {rows} < restored expectation {kept}")
        if entry_report:
            print(
                f"verify {table}: rows={rows} export={entry['rows']} kept={kept} skipped={entry_report['skipped']}"
            )
        else:
            print(f"verify {table}: rows={rows} export={entry['rows']}")

    for child, column in FK_CHECKS:
        sql = (
            f"SELECT count(*) FROM public.{quote_ident(child)} c "
            f"LEFT JOIN public.{quote_ident(column_target(column))} p ON p.id = c.{quote_ident(column)} "
            f"WHERE c.{quote_ident(column)} IS NOT NULL AND p.id IS NULL"
        )
        orphans = int(run_psql(env, sql).strip())
        if orphans:
            failures.append(f"{child}.{column}: {orphans} orphan row(s)")
        print(f"fk {child}.{column}: orphans={orphans}")

    if failures:
        for failure in failures:
            print(f"FAIL: {failure}", file=sys.stderr)
        raise SystemExit(1)
    print("PASS: preserved data matches the export")


def column_target(column: str) -> str:
    mapping = {
        "reporter_staff_id": "staffs",
        "staff_id": "staffs",
        "updated_by_staff_id": "staffs",
        "edited_by_staff_id": "staffs",
        "account_id": "accounts",
        "occupation_id": "occupations",
        "clinic_id": "clinics",
        "group_id": "permission_groups",
        "company_id": "companies",
    }
    if column not in mapping:
        die(f"no FK target mapping for {column}")
    return mapping[column]


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)

    export_parser = sub.add_parser("export", help="export preservation tables to CSV")
    export_parser.add_argument("--env-file", default=str(DEFAULT_ENV_FILE))
    export_parser.add_argument("--out-dir", default="")
    export_parser.set_defaults(func=export)

    restore_parser = sub.add_parser("restore", help="restore an export; dry-run unless --execute")
    restore_parser.add_argument("--dir", required=True)
    restore_parser.add_argument("--env-file", default=str(DEFAULT_ENV_FILE))
    restore_parser.add_argument("--execute", action="store_true")
    restore_parser.set_defaults(func=restore)

    verify_parser = sub.add_parser("verify", help="verify restored counts, hashes, and FKs")
    verify_parser.add_argument("--dir", required=True)
    verify_parser.add_argument("--env-file", default=str(DEFAULT_ENV_FILE))
    verify_parser.add_argument(
        "--allow-sync-drift",
        action="store_true",
        help="exclude status/updated_at from the strict compare (the STG plane_sync job legitimately rewrites them after a restore)",
    )
    verify_parser.set_defaults(func=verify)

    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
