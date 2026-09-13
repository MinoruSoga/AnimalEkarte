#!/usr/bin/env python3
"""Map changed paths to backend domains / frontend features for scoped CI and verify.

Partial runs skip aggregate coverage ratchet. Shared/cross-cutting hits force full mode.
"""
from __future__ import annotations

import argparse
import json
import pathlib
import sys

# ADR-006 domain packages (lintscan domainPackages) minus httpapi (treated as shared).
BACKEND_DOMAINS = frozenset({
    'auth',
    'billing',
    'clinic',
    'identitylink',
    'inventory',
    'lstep',
    'manualarticle',
    'medicalrecord',
    'owner',
    'pet',
    'reservation',
    'staff',
    'trimming',
})

BACKEND_SHARED_PACKAGES = frozenset({
    'apicontract',
    'apperrors',
    'audit',
    'authjwt',
    'clinicale2e',
    'config',
    'csvimport',
    'dbconn',
    'httpapi',
    'infra',
    'labdeviceagent',
    'lintscan',
    'logger',
    'middleware',
    'model',
    'persistence',
    'scheduler',
    'seedbundle',
    'seedlogin',
    'sharedkernel',
    'testdb',
    'textsearch',
    'timeutil',
})

BACKEND_FULL_PATH_PREFIXES = (
    'backend/cmd/',
    'backend/migrations/',
    'backend/worker/',
    'backend/docs/',
)

BACKEND_FULL_EXACT = frozenset({
    'backend/go.mod',
    'backend/go.sum',
    'backend/wrangler.jsonc',
    '.github/workflows/ci.yml',
    'scripts/merge_go_coverprofiles.py',
    'scripts/merge_go_coverprofiles_test.py',
    'scripts/ci_scope_plan.py',
    'scripts/ci_scope_plan_test.py',
    'scripts/verify_seed.py',
})

FRONTEND_SHARED_PREFIXES = (
    'frontend/src/components/',
    'frontend/src/hooks/',
    'frontend/src/lib/',
    'frontend/src/app/',
    'frontend/src/types/',
    'frontend/src/shared-liff/',
    'frontend/src/testing/',
)

FRONTEND_FULL_EXACT = frozenset({
    'frontend/package.json',
    'frontend/pnpm-lock.yaml',
    'frontend/vite.config.ts',
    'frontend/tsconfig.json',
    'frontend/tsconfig.app.json',
    'frontend/tsconfig.node.json',
    '.github/workflows/ci.yml',
})


def _posix(path: str) -> str:
    return path.replace('\\', '/')


def classify_backend_path(path: str) -> tuple[str, str | None]:
    """Return ('domain'|'shared'|'ignore', name_or_reason)."""
    path = _posix(path)
    if path in BACKEND_FULL_EXACT:
        return 'shared', path
    if path.startswith(BACKEND_FULL_PATH_PREFIXES):
        return 'shared', path
    if not path.startswith('backend/'):
        return 'ignore', None
    if path.startswith('backend/internal/'):
        parts = pathlib.PurePosixPath(path).parts
        if len(parts) < 3:
            return 'shared', path
        pkg = parts[2]
        if pkg in BACKEND_DOMAINS:
            return 'domain', pkg
        if pkg in BACKEND_SHARED_PACKAGES:
            return 'shared', pkg
        # Unknown internal package: fail-closed full
        return 'shared', pkg
    return 'shared', path


def classify_frontend_path(path: str) -> tuple[str, str | None]:
    path = _posix(path)
    if path in FRONTEND_FULL_EXACT:
        return 'shared', path
    if not path.startswith('frontend/'):
        return 'ignore', None
    if path.startswith('frontend/src/features/'):
        parts = pathlib.PurePosixPath(path).parts
        # frontend / src / features / <name> / ...
        if len(parts) >= 4 and parts[3] not in ('CLAUDE.md',):
            name = parts[3]
            if name.endswith('.md'):
                return 'ignore', None
            return 'feature', name
        return 'shared', path
    if path.startswith(FRONTEND_SHARED_PREFIXES):
        return 'shared', path
    if path.startswith('frontend/e2e/') or path.startswith('frontend/scripts/'):
        return 'ignore', None
    # Other frontend paths (public, index.html, etc.): fail-closed
    return 'shared', path


def plan_scope(paths: list[str]) -> dict:
    backend_domains: set[str] = set()
    frontend_features: set[str] = set()
    shared_hits: list[str] = []
    saw_backend = False
    saw_frontend = False

    for raw in paths:
        path = _posix(raw)
        kind_b, name_b = classify_backend_path(path)
        if kind_b == 'domain':
            saw_backend = True
            backend_domains.add(name_b)  # type: ignore[arg-type]
        elif kind_b == 'shared':
            saw_backend = True
            shared_hits.append(f'backend:{name_b}')

        kind_f, name_f = classify_frontend_path(path)
        if kind_f == 'feature':
            saw_frontend = True
            frontend_features.add(name_f)  # type: ignore[arg-type]
        elif kind_f == 'shared':
            saw_frontend = True
            shared_hits.append(f'frontend:{name_f}')

    if shared_hits:
        mode = 'full'
        reason = 'shared-hit: ' + shared_hits[0]
        coverage = 'run'
        backend_matrix = _full_backend_matrix() if saw_backend else []
        frontend_matrix = _full_frontend_matrix() if saw_frontend else []
        # Full mode uses legacy full shards for the layers that were actually touched.
        return {
            'mode': mode,
            'backend_domains': sorted(backend_domains),
            'frontend_features': sorted(frontend_features),
            'backend': saw_backend,
            'frontend': saw_frontend,
            'run_backend_tests': saw_backend,
            'run_frontend_tests': saw_frontend,
            'coverage_ratchet': coverage,
            'reason': reason,
            'backend_matrix': backend_matrix,
            'frontend_matrix': frontend_matrix,
            'backend_shard_count': len(backend_matrix),
            'frontend_shard_count': len(frontend_matrix),
        }

    mode = 'partial'
    coverage = 'skip'
    reasons = []
    if backend_domains:
        reasons.append('backend-domains=' + ','.join(sorted(backend_domains)))
    if frontend_features:
        reasons.append('frontend-features=' + ','.join(sorted(frontend_features)))
    if not reasons:
        return {
            'mode': 'skip',
            'backend_domains': [],
            'frontend_features': [],
            'backend': False,
            'frontend': False,
            'run_backend_tests': False,
            'run_frontend_tests': False,
            'coverage_ratchet': 'skip',
            'reason': 'no-testable-backend-or-frontend-paths',
            'backend_matrix': [],
            'frontend_matrix': [],
            'backend_shard_count': 0,
            'frontend_shard_count': 0,
        }

    backend_matrix = [
        {
            'shard': domain,
            'select': 'packages',
            # Space-separated module paths for bash arrays in ci.yml
            'packages': f'./internal/{domain}',
        }
        for domain in sorted(backend_domains)
    ]
    frontend_matrix = [
        {
            'shard': feature,
            'select': 'feature',
            'feature_path': f'src/features/{feature}',
            'vitest_shard': '',
        }
        for feature in sorted(frontend_features)
    ]
    return {
        'mode': mode,
        'backend_domains': sorted(backend_domains),
        'frontend_features': sorted(frontend_features),
        'backend': bool(backend_domains),
        'frontend': bool(frontend_features),
        'run_backend_tests': bool(backend_domains),
        'run_frontend_tests': bool(frontend_features),
        'coverage_ratchet': coverage,
        'reason': '; '.join(reasons),
        'backend_matrix': backend_matrix,
        'frontend_matrix': frontend_matrix,
        'backend_shard_count': len(backend_matrix),
        'frontend_shard_count': len(frontend_matrix),
    }


def _full_backend_matrix() -> list[dict]:
    return [
        {'shard': 'medicalrecord', 'select': 'packages', 'packages': './internal/medicalrecord'},
        {'shard': 'auth', 'select': 'packages', 'packages': './internal/auth'},
        {
            'shard': 'staff-billing-reservation',
            'select': 'packages',
            'packages': './internal/staff ./internal/billing ./internal/reservation',
        },
        {'shard': 'remaining', 'select': 'remaining', 'packages': ''},
    ]


def _full_frontend_matrix() -> list[dict]:
    return [
        {'shard': '1', 'select': 'vitest-shard', 'feature_path': '', 'vitest_shard': '1/2'},
        {'shard': '2', 'select': 'vitest-shard', 'feature_path': '', 'vitest_shard': '2/2'},
    ]


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--paths', nargs='+', default=[], help='Changed repo-relative paths')
    parser.add_argument('--paths-file', help='File with one path per line')
    parser.add_argument('--github-output', help='Write key=value pairs for GITHUB_OUTPUT')
    args = parser.parse_args(argv)
    paths = list(args.paths)
    if args.paths_file:
        text = pathlib.Path(args.paths_file).read_text(encoding='utf-8')
        paths.extend(line.strip() for line in text.splitlines() if line.strip())
    plan = plan_scope(paths)
    print(json.dumps(plan, ensure_ascii=False, indent=2))
    if args.github_output:
        out = pathlib.Path(args.github_output)
        # GITHUB_OUTPUT: use heredoc delimiters for values that may contain spaces.
        chunks = [
            f"mode={plan['mode']}\n",
            f"coverage_ratchet={plan['coverage_ratchet']}\n",
            f"backend={str(plan['backend']).lower()}\n",
            f"frontend={str(plan['frontend']).lower()}\n",
            f"run_backend_tests={str(plan['run_backend_tests']).lower()}\n",
            f"run_frontend_tests={str(plan['run_frontend_tests']).lower()}\n",
            f"backend_shard_count={plan['backend_shard_count']}\n",
            f"frontend_shard_count={plan['frontend_shard_count']}\n",
            "backend_matrix<<EOF_BACKEND_MATRIX\n"
            + json.dumps(plan['backend_matrix'], separators=(',', ':'))
            + "\nEOF_BACKEND_MATRIX\n",
            "frontend_matrix<<EOF_FRONTEND_MATRIX\n"
            + json.dumps(plan['frontend_matrix'], separators=(',', ':'))
            + "\nEOF_FRONTEND_MATRIX\n",
            "reason<<EOF_REASON\n" + plan['reason'] + "\nEOF_REASON\n",
        ]
        out.write_text(''.join(chunks), encoding='utf-8')
    return 0


if __name__ == '__main__':
    sys.exit(main())
