#!/usr/bin/env python3
"""Bounded offline verification. No provisioning, installation, or DB operations.

Exit 0: selected checks pass (or documentation-only SKIP); 1: FAIL; 2: BLOCKED.
JSON evidence deliberately excludes command output and container environment.
"""
import argparse
import hashlib
import json
import os
import pathlib
import re
import subprocess
import sys
import uuid

ROOT = pathlib.Path(__file__).resolve().parents[1]


def run(command, root=ROOT):
    owned_name = 'ae-verify-' + uuid.uuid4().hex if command[:2] == ['docker', 'run'] else None
    actual = command[:2] + ['--name', owned_name] + command[2:] if owned_name else command
    try:
        return subprocess.run(actual, cwd=root, capture_output=True, text=True, timeout=300, check=False)
    except subprocess.TimeoutExpired:
        if owned_name:
            # Stop only the uniquely named ephemeral container created by this call.
            subprocess.run(['docker', 'rm', '-f', owned_name], cwd=root, capture_output=True,
                           text=True, timeout=30, check=False)
        raise


def git(*arguments):
    result = run(['git', *arguments])
    if result.returncode:
        raise ValueError('Git scope resolution failed')
    return result.stdout


def validate_path(path):
    if any(ord(character) < 32 for character in path):
        raise ValueError('Control characters are not valid scope paths')
    parsed = pathlib.PurePosixPath(path)
    if parsed.is_absolute() or '..' in parsed.parts or path.startswith('-') or not parsed.parts:
        raise ValueError('Scope must contain repository-relative paths')
    if not (ROOT / parsed).resolve().is_relative_to(ROOT.resolve()):
        raise ValueError('Scope symlink escapes repository')
    return parsed.as_posix()


ACCOUNT_LAYOUT_PATHS = {
    '.env.example',
    'Makefile',
    'backend/.dockerignore',
    'docker-compose.yml',
    'scripts/check-old-db-handoff.sh',
    'scripts/stage-old-db-handoff.sh',
    'scripts/verify_seed.py',
    'scripts/test_account_csv_layout.py',
    'backend/migrations/seeds/002_master/accounts/permission_group_rules.csv',
    'backend/migrations/seeds/002_master/accounts/permission_groups.csv',
}

# Extra go test flags for packages whose default suite needs files outside the
# backend-only offline mount (repo-root Makefile, frontend, docs, scripts).
OFFLINE_GO_PACKAGE_FLAGS = {
    'backend/cmd/migrate': ['-skip=^TestMakefile'],
    'backend/internal/lintscan': [
        '-skip=^TestDemoAccountLoginFormMatches|^TestDocsMigrationInventoryDrift$|^TestERDTableCount_MatchesSchema$|^TestHandoffDeleteBlock_',
    ],
}

# Fixed source inputs for Go contracts that read outside the backend module.
# Only these individual worktree documents are exposed; never mount the repo root.
BACKEND_CONTRACT_DOCUMENTS = {
    './internal/auth': ('docs/ops/deploy/FIRST_SYSTEM_ADMIN.md',),
    './internal/medicalrecord': (
        'docs/ops/deploy/LAB_DEVICE_CONNECTIVITY.md',
        'docs/architecture/adr/007-lab-device-receive-and-commit.md',
    ),
}


def backend_contract_documents(command):
    return sorted({document for package, documents in BACKEND_CONTRACT_DOCUMENTS.items()
                   if package in command for document in documents})


def scoped_contract_documents(paths):
    return sorted({document for job in plan(paths)[0] if job['service'] == 'backend'
                   for document in backend_contract_documents(job['command'])})


E2E_TSCONFIG_BOOTSTRAP = (
    'const fs=require("fs");const path=require("path");'
    'const files=process.argv.slice(1).map((f)=>path.resolve("/app",f));'
    'fs.writeFileSync("/tmp/ae-e2e-tsconfig.json",JSON.stringify({'
    'compilerOptions:{strict:true,module:"CommonJS",moduleResolution:"Node",'
    'target:"ES2022",esModuleInterop:true,skipLibCheck:true,ignoreDeprecations:"6.0",'
    'types:["node"],typeRoots:["/app/node_modules/@types","/app/node_modules"]},'
    'files}));'
)


E2E_TRUSTED_DISCOVERY_ROOT = '/app/e2e'


def normalize_e2e_repo_path(path):
    """Normalize frontend/e2e or e2e paths to frontend/e2e/... identity."""
    validate_path(path)
    relative = path[len('frontend/'):] if path.startswith('frontend/') else path
    if not relative.startswith('e2e/'):
        raise ValueError(f'path is outside trusted e2e root: {path}')
    normalized = pathlib.PurePosixPath('frontend') / relative
    parts = normalized.parts
    if '..' in parts:
        raise ValueError(f'path traversal is not allowed: {path}')
    return pathlib.PurePosixPath(*parts).as_posix()


def list_e2e_spec_paths():
    """Return current repo-relative frontend/e2e/**/*.spec.ts paths (re-counted, not hardcoded)."""
    root = ROOT / 'frontend' / 'e2e'
    if not root.is_dir():
        return []
    specs = []
    for spec in sorted(root.rglob('*.spec.ts')):
        if not spec.is_file():
            continue
        relative = spec.relative_to(ROOT).as_posix()
        validate_path(relative)
        specs.append(relative)
    return specs


def validate_e2e_page_consumers(stdout, expected_pages=None, stderr=''):
    """Fail-closed parse of verify-e2e-page-consumers.mjs JSON evidence."""
    text = (stdout or '').strip() or (stderr or '').strip()
    if not text:
        raise ValueError('E2E page consumer AST produced empty output')
    start = text.find('{')
    if start < 0:
        raise ValueError('E2E page consumer AST output is not JSON')
    try:
        payload = json.loads(text[start:])
    except ValueError as error:
        raise ValueError('E2E page consumer AST output is malformed JSON') from error
    if payload.get('ok') is not True:
        raise ValueError('E2E page consumer AST reported ok!=true')
    pages = payload.get('pages')
    if not isinstance(pages, list) or not pages:
        raise ValueError('E2E page consumer AST pages payload is missing')
    evidence = {}
    for entry in pages:
        if not isinstance(entry, dict):
            raise ValueError('E2E page consumer AST page entry is malformed')
        page = entry.get('page')
        consumers = entry.get('consumers')
        blocking = entry.get('blocking')
        if not isinstance(page, str) or not page:
            raise ValueError('E2E page consumer AST page identity is missing')
        if not isinstance(consumers, list) or not isinstance(blocking, list):
            raise ValueError(f'E2E page consumer AST consumers/blocking malformed for {page}')
        if blocking:
            raise ValueError(f'E2E page consumer AST blocking references for {page}')
        if not consumers:
            raise ValueError(f'E2E page object has no spec consumer: {page}')
        for consumer in consumers:
            if not isinstance(consumer, dict):
                raise ValueError(f'E2E page consumer AST consumer entry malformed for {page}')
            if not consumer.get('file') or not consumer.get('form') or 'specifier' not in consumer:
                raise ValueError(f'E2E page consumer AST consumer identity incomplete for {page}')
        evidence[page] = consumers
    if expected_pages:
        found = {entry.get('page') for entry in pages}
        missing = [page for page in expected_pages if page not in found]
        if missing:
            raise ValueError('E2E page consumer AST missing pages: ' + ', '.join(missing))
    return evidence


def playwright_discovery_counts(stdout, trusted_root=E2E_TRUSTED_DISCOVERY_ROOT):
    """Parse Playwright --list JSON into trusted repo-relative identity -> test count."""
    text = stdout.strip()
    if not text:
        raise ValueError('E2E discovery produced empty output')
    start = text.find('{')
    if start < 0:
        raise ValueError('E2E discovery output is not JSON')
    try:
        payload = json.loads(text[start:])
    except ValueError as error:
        raise ValueError('E2E discovery output is malformed JSON') from error
    root_dir = str((payload.get('config') or {}).get('rootDir') or '')
    if os.path.normpath(root_dir) != os.path.normpath(trusted_root):
        raise ValueError(f'E2E discovery rootDir is not trusted: {root_dir!r}')
    counts = {}

    def walk(node):
        for spec in node.get('specs') or []:
            file_field = str(spec.get('file') or '')
            if not file_field or file_field.startswith('/') or '..' in pathlib.PurePosixPath(file_field).parts:
                raise ValueError(f'E2E discovery file identity is ambiguous or out of root: {file_field!r}')
            if file_field.startswith('./'):
                relative = file_field[2:]
            elif file_field.startswith('.'):
                raise ValueError(f'E2E discovery file identity is ambiguous or out of root: {file_field!r}')
            else:
                relative = file_field
            identity = normalize_e2e_repo_path('frontend/e2e/' + relative)
            tests = spec.get('tests')
            if not isinstance(tests, list):
                raise ValueError(f'E2E discovery tests payload is malformed for {identity}')
            counts[identity] = counts.get(identity, 0) + len(tests)
        for child in node.get('suites') or []:
            walk(child)

    walk(payload)
    return counts


def validate_playwright_discovery(stdout, selected_specs, trusted_root=E2E_TRUSTED_DISCOVERY_ROOT):
    """Require every selected spec identity to appear with at least one registered test."""
    if not selected_specs:
        raise ValueError('E2E discovery selection is empty')
    selected = [normalize_e2e_repo_path(path if path.startswith('frontend/') else 'frontend/' + path)
                for path in selected_specs]
    counts = playwright_discovery_counts(stdout, trusted_root=trusted_root)
    missing = [path for path in selected if counts.get(path, 0) < 1]
    if missing:
        raise ValueError('E2E discovery found no registered tests for: ' + ', '.join(missing))
    return counts


def check_e2e_scope(paths):
    """Fail-closed offline checks for selected E2E/runner paths (no browser runtime)."""
    selected = []
    for path in paths:
        validate_path(path)
        selected.append(path)
    if not selected:
        raise ValueError('E2E scope selection is empty')
    runner = 'frontend/scripts/run-e2e.sh'
    if runner in selected:
        target = ROOT / runner
        if not target.is_file():
            raise ValueError('frontend/scripts/run-e2e.sh is missing')
        text = target.read_text(encoding='utf-8')
        for variable in ('UAT_SYNTHETIC_CLOSING_PASSWORD', 'UAT_SYNTHETIC_CLOSING_API_BASE'):
            if re.search(r'-e\s+' + re.escape(variable) + r'\s*=', text):
                raise ValueError(f'{variable} must use name-only -e (no secret value on argv)')
            if not re.search(r'-e\s+' + re.escape(variable) + r'\b(?!\s*=)', text):
                raise ValueError(f'{variable} must be forwarded with name-only docker -e')
        if '"$@"' not in text:
            raise ValueError('run-e2e.sh must preserve playwright argv via "$@"')
    for path in selected:
        if path == runner:
            continue
        target = ROOT / path
        if not target.is_file():
            raise ValueError(f'E2E path is missing: {path}')


def plan(paths):
    jobs, blocked, frontend = [], [], []
    e2e_ts, e2e_pages, e2e_runner = [], [], False
    for path in paths:
        validate_path(path)
        if path.startswith('frontend/src/') and path.endswith(('.ts', '.tsx', '.js', '.jsx')):
            frontend.append(path.removeprefix('frontend/'))
        elif path.startswith('frontend/e2e/') and path.endswith('.spec.ts'):
            if not (ROOT / path).is_file():
                blocked.append(path)
                continue
            e2e_ts.append(path)
        elif path.startswith('frontend/e2e/pages/') and path.endswith('.ts'):
            if not (ROOT / path).is_file():
                blocked.append(path)
                continue
            e2e_pages.append(path)
            if path not in e2e_ts:
                e2e_ts.append(path)
        elif path in (
            'frontend/scripts/verify-e2e-page-consumers.mjs',
            'frontend/scripts/verify-e2e-page-consumers.test.mjs',
        ):
            jobs.append({
                'service': 'frontend',
                'command': ['node', '--test', 'scripts/verify-e2e-page-consumers.test.mjs'],
            })
        elif path in ('frontend/vite.config.ts', 'frontend/scripts/vite-native-config.test.mjs'):
            jobs.append({'service': 'frontend', 'command': ['node', '--test', 'scripts/vite-native-config.test.mjs']})
        elif path == 'frontend/scripts/run-e2e.sh':
            if not (ROOT / path).is_file():
                blocked.append(path)
                continue
            e2e_runner = True
        elif path.startswith('frontend/e2e/') or path.startswith('frontend/scripts/'):
            blocked.append(path)
        elif (path.startswith('backend/internal/') or path.startswith('backend/cmd/')) and path.endswith('.go'):
            package = pathlib.PurePosixPath(path).parent
            if not list((ROOT / package).glob('*_test.go')):
                blocked.append(path)
                continue
            job = {'service': 'backend', 'command': ['go', 'test', '-json', '-p=2', '-count=1', '-short', './' + str(package).removeprefix('backend/')], 'require_completed_test': True}
            extra = OFFLINE_GO_PACKAGE_FLAGS.get(str(package))
            if extra:
                job['command'] = job['command'] + extra
            if job not in jobs:
                jobs.append(job)
            if (ROOT / path).is_file():
                jobs.append({'service': 'backend', 'command': ['gofmt', '-l', path.removeprefix('backend/')], 'require_empty_stdout': True})
        elif path in ('scripts/verify-agent-task.py', 'scripts/test_verify_agent_task.py', '.githooks/pre-commit', '.githooks/pre-push', '.githooks/lib/check-secrets.sh', 'scripts/run-local-ci.sh'):
            if not any(job['service'] == 'host' for job in jobs):
                jobs.append({'service': 'host', 'command': ['python3', '-B', 'scripts/test_verify_agent_task.py']})
        elif path in ('scripts/test_agent_scope_contracts.py', '.gitignore', '.mcp.json', '.claude/settings.json', '.claude/codex-agent-manifest.json', 'backend/wrangler.jsonc'):
            jobs.append({'service': 'host', 'command': ['python3', '-B', 'scripts/test_agent_scope_contracts.py']})
        elif path in ('.claude/scripts/sync-codex-mirror.py', '.claude/scripts/test_sync_codex_mirror.py', '.claude/scripts/sync-codex-mirror.sh'):
            jobs.append({'service': 'host', 'command': ['python3', '-B', '.claude/scripts/test_sync_codex_mirror.py']})
        elif path in ('.claude/scripts/sync-agents-skills.py', '.claude/scripts/test_sync_agents_skills.py', '.claude/scripts/sync-agents-skills.sh'):
            jobs.append({'service': 'host', 'command': ['python3', '-B', '.claude/scripts/test_sync_agents_skills.py']})
        elif path == '.claude/scripts/mirror_safety.py':
            for test in ('test_sync_codex_mirror.py', 'test_sync_agents_skills.py'):
                jobs.append({'service': 'host', 'command': ['python3', '-B', '.claude/scripts/' + test]})
        elif path in ACCOUNT_LAYOUT_PATHS:
            jobs.append({'service': 'host', 'command': ['python3', '-B', 'scripts/test_account_csv_layout.py']})
        elif path in ('scripts/check-workflow-contracts.test.mjs', '.github/workflows/security-scan.yml', '.github/workflows/README-security-scan.md'):
            jobs.append({'service': 'host', 'command': ['node', '--test', 'scripts/check-workflow-contracts.test.mjs']})
        elif path.startswith('backend/worker/') or path == 'backend/wrangler.jsonc':
            jobs.append({'service': 'host', 'command': ['bash', 'scripts/check-test-worker-makefile.test.sh']})
        elif path.startswith('.claude/skills/') and path.endswith('.md'):
            jobs.append({'service': 'host', 'command': ['python3', '-B', '.claude/scripts/test_instruction_safety_contracts.py'], 'requires_mirrors': True})
        elif path == 'backend/docs/api.yaml':
            jobs.append({'service': 'backend', 'command': ['go', 'test', '-json', '-p=2', '-count=1', '-short', './internal/apicontract'], 'require_completed_test': True})
        elif path == 'backend/cmd/api/testdata/get_head_permissions.json':
            jobs.append({'service': 'backend', 'command': ['go', 'test', '-json', '-p=2', '-count=1', '-short', './cmd/api', '-run=^TestGETHEAD'], 'require_completed_test': True})
        elif path in ('backend/internal/auth/testdata/first_system_admin.sql', 'docs/ops/deploy/FIRST_SYSTEM_ADMIN.md'):
            jobs.append({'service': 'backend', 'command': ['go', 'test', '-json', '-p=2', '-count=1', '-short', './internal/auth', '-run=^TestFirstSystemAdminProcedureMatchesInitSchema$'], 'require_completed_test': True})
        elif path.endswith('.md') and (path.startswith(('docs/', '.claude/', '.codex/', '.agents/', 'frontend/src/features/manual/'))
                                      or '/' not in path or pathlib.PurePosixPath(path).name in ('CLAUDE.md', 'AGENTS.md', 'README.md')):
            continue
        else:
            blocked.append(path)
    for path in paths:
        if path in ('.githooks/pre-commit', '.githooks/pre-push', '.githooks/lib/check-secrets.sh', 'scripts/run-local-ci.sh', '.claude/scripts/sync-codex-mirror.sh', '.claude/scripts/sync-agents-skills.sh'):
            jobs.append({'service': 'host', 'command': ['bash' if path.endswith('.sh') else 'sh', '-n', path]})
    if frontend:
        jobs.append({'service': 'frontend', 'command': ['node', 'node_modules/vitest/vitest.mjs', 'related', '--run', '--configLoader', 'native', '--reporter=json', *frontend], 'require_frontend_tests': True})
        existing = [path for path in frontend if (ROOT / 'frontend' / path).is_file() and '/types/generated/' not in '/' + path]
        if existing:
            jobs.append({'service': 'frontend', 'command': ['node', 'node_modules/eslint/bin/eslint.js', '--max-warnings', '0', *existing]})
            jobs.append({'service': 'frontend', 'command': ['node', 'node_modules/prettier/bin/prettier.cjs', '--check', *existing]})
    if e2e_pages:
        for spec in list_e2e_spec_paths():
            if spec not in e2e_ts:
                e2e_ts.append(spec)
    e2e_selected = []
    for path in e2e_ts:
        if path not in e2e_selected:
            e2e_selected.append(path)
    if e2e_runner:
        e2e_selected.append('frontend/scripts/run-e2e.sh')
    if e2e_selected:
        relative_ts = [path.removeprefix('frontend/') for path in e2e_selected if path.endswith('.ts')]
        relative_specs = [path.removeprefix('frontend/') for path in e2e_selected if path.endswith('.spec.ts')]
        if relative_ts:
            jobs.append({'service': 'frontend', 'command': ['node', 'node_modules/eslint/bin/eslint.js', '--max-warnings', '0', *relative_ts]})
            jobs.append({'service': 'frontend', 'command': ['node', 'node_modules/prettier/bin/prettier.cjs', '--check', *relative_ts]})
            if e2e_pages:
                # Page-only conservative mode typechecks the full e2e project so
                # fixture @/ imports resolve via e2e/tsconfig.json (extends app tsconfig).
                jobs.append({
                    'service': 'frontend',
                    'command': [
                        'node', 'node_modules/typescript/bin/tsc',
                        '-p', 'e2e/tsconfig.json', '--noEmit', '--pretty', 'false',
                    ],
                })
            else:
                jobs.append({
                    'service': 'frontend',
                    'command': [
                        'sh', '-c',
                        'node -e \'' + E2E_TSCONFIG_BOOTSTRAP + '\' "$@" '
                        '&& node node_modules/typescript/bin/tsc -p /tmp/ae-e2e-tsconfig.json --noEmit --pretty false',
                        'ae-e2e-tsc',
                        *relative_ts,
                    ],
                })
        if relative_specs:
            jobs.append({
                'service': 'frontend',
                'command': [
                    'node', 'node_modules/@playwright/test/cli.js', 'test',
                    '--list', '--reporter=json', *relative_specs,
                ],
                'require_e2e_discovery': True,
                'e2e_discovery_specs': relative_specs,
            })
        if e2e_runner:
            jobs.append({'service': 'host', 'command': ['bash', '-n', 'frontend/scripts/run-e2e.sh']})
        jobs.append({'service': 'host', 'command': ['python3', '-B', 'scripts/verify-agent-task.py', '--check-e2e-scope', *e2e_selected]})
        if e2e_pages:
            page_args = []
            page_idents = []
            for page in e2e_pages:
                ident = page.removeprefix('frontend/')
                page_args.extend(['--page', ident])
                page_idents.append(ident)
            jobs.append({
                'service': 'frontend',
                'command': [
                    'node', 'scripts/verify-e2e-page-consumers.mjs',
                    *page_args,
                    '--e2e-root', 'e2e',
                ],
                'require_e2e_page_consumers': True,
                'e2e_pages': page_idents,
            })
    unique = []
    for job in jobs:
        if job not in unique:
            unique.append(job)
    return unique, blocked


def container_error(info, root, service):
    if not info.get('Image'):
        return 'container image identity is missing'
    host = info.get('HostConfig', {})
    if not info.get('State', {}).get('Running'):
        return 'container is not running'
    if host.get('NetworkMode') != 'none':
        return 'container network must be none; shared DB/network access is not allowed'
    if host.get('Privileged') or host.get('PidMode') == 'host' or host.get('IpcMode') == 'host':
        return 'privileged/host PID container is not allowed'
    if 'ALL' not in host.get('CapDrop', []) or host.get('CapAdd') or host.get('Devices') or host.get('DeviceRequests'):
        return 'container must drop all capabilities and expose no devices'
    if not any(option in ('no-new-privileges', 'no-new-privileges=true', 'no-new-privileges:true') for option in host.get('SecurityOpt', [])):
        return 'container requires no-new-privileges'
    if not host.get('ReadonlyRootfs'):
        return 'container root filesystem must be read-only'
    mounts = info.get('Mounts', [])
    working = info.get('Config', {}).get('WorkingDir', '')
    expected = (root / service).resolve()
    matching = [mount for mount in mounts if mount.get('Destination') == working
                and mount.get('Type') == 'bind'
                and pathlib.Path(mount.get('Source', '')).resolve() == expected]
    if any(mount.get('RW', True) for mount in mounts
           if not (mount.get('Type') == 'tmpfs' and mount.get('Destination') in ('/tmp', '/verify'))):
        return 'container mounts must be read-only'
    if not matching:
        return 'container worktree mount does not match selected repository/service'
    if any(mount.get('Type') == 'bind' and pathlib.Path(mount.get('Source', '')).resolve() != expected for mount in mounts):
        return 'container has additional host bind mounts'
    if any(mount.get('Destination', '').startswith(working.rstrip('/') + '/') for mount in mounts
           if not mount.get('Destination', '').endswith('/node_modules')):
        return 'container mount masks selected source files'
    return ''


def inspect_container(service, explicit):
    container = explicit
    if not container:
        result = run(['docker', 'compose', 'ps', '-q', service])
        if result.returncode or len(result.stdout.split()) != 1:
            raise ValueError('No unique running service; supply an isolated --' + service + '-container')
        container = result.stdout.strip()
    if not container or not all(character.isalnum() or character in '_.-' for character in container) or container.startswith('-'):
        raise ValueError('Invalid container identifier')
    result = run(['docker', 'inspect', container])
    if result.returncode:
        raise ValueError('Container inspection failed')
    info = json.loads(result.stdout)[0]
    error = container_error(info, ROOT, service)
    if error:
        raise ValueError(error)
    return container, info.get('Image', '')




def prepare_dependency_mountpoint(service, volume):
    if service != 'frontend' or not volume:
        return
    target = ROOT / 'frontend/node_modules'
    validate_path('frontend/node_modules')
    if (ROOT / 'frontend').is_symlink() or target.is_symlink():
        raise ValueError('Dependency mountpoint must not be a symlink')
    if target.exists():
        if not target.is_dir():
            raise ValueError('Dependency mountpoint is not a directory')
        return
    ignored = run(['git', 'check-ignore', '-q', 'frontend/node_modules'])
    if ignored.returncode:
        raise ValueError('Dependency mountpoint must be ignored before creating empty directory')
    target.mkdir()  # Empty bind mount scaffold only; no install and no replacement.


def inspect_image(reference):
    if reference.startswith('-') or any(character.isspace() for character in reference):
        raise ValueError('Invalid local image reference')
    result = run(['docker', 'image', 'inspect', reference])
    if result.returncode:
        raise ValueError('Selected image is not cached locally; no pull was attempted')
    info = json.loads(result.stdout)[0]
    if info.get('Config', {}).get('Volumes'):
        raise ValueError('Image declares implicit volumes; select an image without automatic volumes')
    identity = info.get('Id', '')
    if not identity.startswith('sha256:'):
        raise ValueError('Local image identity missing')
    return identity


def inspect_dependency_volume(volume):
    if not volume or not volume[0].isalnum() or not all(character.isalnum() or character in '_.-' for character in volume):
        raise ValueError('Invalid named dependency volume')
    result = run(['docker', 'volume', 'inspect', volume])
    if result.returncode:
        raise ValueError('Dependency volume does not exist; no volume was created')
    info = json.loads(result.stdout)[0]
    if info.get('Driver') != 'local' or info.get('Options'):
        raise ValueError('Only ordinary existing local dependency volumes are allowed')


def image_command(identity, service, command, volume=None):
    source = str(ROOT / service)
    if ',' in source:
        raise ValueError('Docker mount source contains an unsupported comma')
    invocation = ['docker', 'run', '--rm', '--pull=never', '--network=none', '--read-only',
                  '--cap-drop=ALL', '--security-opt=no-new-privileges', '--pids-limit=256',
                  '--mount', 'type=bind,src=' + source + ',dst=/app,readonly',
                  '--tmpfs', '/tmp:rw,nosuid,nodev,size=1073741824', '-w', '/app',
                  '--entrypoint', '/usr/bin/env']
    if service == 'backend':
        invocation += ['--tmpfs', '/verify:rw,exec,nosuid,nodev,size=1073741824']
        for document in backend_contract_documents(command):
            validate_path(document)
            source_document = ROOT / document
            if source_document.is_symlink():
                raise ValueError('Backend contract document must not be a symlink')
            if not source_document.is_file():
                raise ValueError('Required backend contract document is missing')
            invocation += ['--mount', 'type=bind,src=' + str(source_document)
                           + ',dst=/' + document + ',readonly']
    if volume:
        destination = '/app/node_modules' if service == 'frontend' else '/go/pkg/mod'
        invocation += ['--mount', 'type=volume,src=' + volume + ',dst=' + destination + ',readonly,volume-nocopy']
    return invocation + [identity, '-i', 'PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin',
                         'HOME=/tmp', 'TMPDIR=/tmp', 'GOCACHE=/tmp/go-cache', 'GOMODCACHE=/go/pkg/mod',
                         'GOPROXY=off', 'GOSUMDB=off', 'GOTOOLCHAIN=local', 'GOTMPDIR=/verify', 'GOMAXPROCS=2', 'CI=true', *command]


def selected_services(paths):
    services = {path.split('/')[0] for path in paths if path.startswith(('backend/', 'frontend/'))}
    services.update(job['service'] for job in plan(paths)[0] if job['service'] in ('backend', 'frontend'))
    return sorted(services)


def reject_unstaged_service_sources(paths):
    for service in selected_services(paths):
        if git('diff', '--name-only', '-z', '--', service) or git('ls-files', '--others', '--exclude-standard', '-z', '--', service):
            raise ValueError('Selected service has unstaged or untracked files; staged verification cannot use them')
    documents = scoped_contract_documents(paths)
    if documents and (git('diff', '--name-only', '-z', '--', *documents)
                      or git('ls-files', '--others', '--exclude-standard', '-z', '--', *documents)):
        raise ValueError('Backend contract documents are unstaged or untracked; cannot prove staged content')


def scope_fingerprint(paths):
    expanded = set(paths)
    expanded.update(scoped_contract_documents(paths))
    for service in selected_services(paths):
        expanded.update(filter(None, git('ls-files', '--cached', '--others', '--exclude-standard', '-z', '--', service).split('\0')))
    # Do not read environment files or generated dependency/cache content.
    paths = [path for path in expanded if not pathlib.PurePosixPath(path).name.startswith('.env')
             and not any(part in ('node_modules', '__pycache__', '.git') for part in pathlib.PurePosixPath(path).parts)]
    digest = hashlib.sha256()
    for path in sorted(paths):
        digest.update(path.encode())
        validate_path(path)
        target = ROOT / path
        if target.is_symlink():
            digest.update(os.readlink(target).encode())
        elif target.is_file():
            digest.update(target.read_bytes())
        else:
            digest.update(b'<deleted-or-directory>')
    return digest.hexdigest()



def local_defaults():
    allowed = ('backend_image', 'frontend_image', 'backend_dependency_volume', 'frontend_dependency_volume')
    defaults = {}
    path = ROOT / '.claude/verification.local.json'
    if path.exists():
        if path.is_symlink() or not path.resolve().is_relative_to(ROOT.resolve()):
            raise ValueError('Local verification configuration must be a regular project-local file')
        try:
            defaults = json.loads(path.read_text())
        except (OSError, ValueError):
            raise ValueError('Local verification configuration is unreadable or invalid JSON') from None
        if not isinstance(defaults, dict) or set(defaults) - set(allowed):
            raise ValueError('Local verification configuration contains unsupported keys')
        if any(not isinstance(value, str) or not value or any(ord(c) < 32 for c in value) for value in defaults.values()):
            raise ValueError('Local verification configuration requires nonempty string selectors')
    return {key: os.environ.get('AGENT_VERIFY_' + key.upper(), defaults.get(key)) for key in allowed}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    try:
        defaults = local_defaults()
    except ValueError as error:
        parser.error(str(error))
    scope = parser.add_mutually_exclusive_group(required=True)
    scope.add_argument('--base', help='Local base commit; includes current tracked and untracked work')
    scope.add_argument('--staged', action='store_true')
    scope.add_argument('--paths', nargs='+')
    scope.add_argument('--check-e2e-scope', nargs='+', help='Offline E2E/runner contract check for exact paths')
    parser.add_argument('--plan', action='store_true', help='Resolve scope only; no check execution')
    parser.add_argument('--frontend-container')
    parser.add_argument('--backend-container')
    parser.add_argument('--frontend-image', default=defaults.get('frontend_image'), help='Explicit existing local image; run an ephemeral offline check')
    parser.add_argument('--backend-image', default=defaults.get('backend_image'), help='Explicit existing local image; run an ephemeral offline check')
    parser.add_argument('--frontend-dependency-volume', default=defaults.get('frontend_dependency_volume'), help='Explicit existing read-only node_modules volume')
    parser.add_argument('--backend-dependency-volume', default=defaults.get('backend_dependency_volume'), help='Explicit existing read-only Go module cache volume')
    parser.add_argument('--evidence', type=pathlib.Path)
    args = parser.parse_args()
    if args.check_e2e_scope is not None:
        try:
            check_e2e_scope(args.check_e2e_scope)
        except ValueError as error:
            print(str(error), file=sys.stderr)
            return 1
        return 0
    evidence = {'status': 'BLOCKED', 'root': str(ROOT), 'checks': []}
    code = 2
    try:
        for service in ('frontend', 'backend'):
            if getattr(args, service + '_image') and getattr(args, service + '_container'):
                raise ValueError('Choose image or existing container per service, not both')
            if getattr(args, service + '_dependency_volume') and not getattr(args, service + '_image'):
                raise ValueError('Dependency volumes require explicit image mode')
        if args.paths:
            paths = [validate_path(path) for path in args.paths]
        elif args.staged:
            paths = git('diff', '--cached', '--name-only', '-z').split('\0')
            if git('diff', '--name-only', '-z', '--', *[path for path in paths if path]):
                raise ValueError('Staged paths also have unstaged changes; cannot prove staged content')
        else:
            base = git('rev-parse', '--verify', '--end-of-options', args.base + '^{commit}').strip()
            paths = git('diff', base, '--name-only', '-z').split('\0')
            paths += git('ls-files', '--others', '--exclude-standard', '-z').split('\0')
        paths = sorted(set(validate_path(path) for path in filter(None, paths)))
        if args.staged:
            reject_unstaged_service_sources(paths)
        evidence.update(paths=paths, head=git('rev-parse', 'HEAD').strip(), fingerprint=scope_fingerprint(paths))
        jobs, blocked = plan(paths)
        evidence['unmapped'] = blocked
        if blocked:
            raise ValueError('Unmapped changes require a scoped verification contract before execution')
        if args.plan:
            evidence.update(status='PLAN', checks=jobs)
            code = 0
        else:
            for job in jobs:
                if job.get('requires_mirrors') and not (ROOT / '.agents/skills').is_dir():
                    raise ValueError('Generated skill mirrors are missing; run the safe mirror generators first')
                command = job['command']
                check = dict(job)
                if job['service'] != 'host':
                    service = job['service']
                    reference = getattr(args, service + '_image')
                    if reference:
                        image = inspect_image(reference)
                        volume = getattr(args, service + '_dependency_volume')
                        if volume:
                            inspect_dependency_volume(volume)
                            prepare_dependency_mountpoint(service, volume)
                        check.update(image=image, dependency_volume=volume, mode='ephemeral-offline')
                        command = image_command(image, service, command, volume)
                    else:
                        if service == 'backend' and backend_contract_documents(command):
                            raise ValueError('Backend document contracts require offline image mode with exact read-only inputs')
                        container, image = inspect_container(service, getattr(args, service + '_container'))
                        check.update(container=container, image=image)
                        command = ['docker', 'exec', container, '/usr/bin/env', '-i', 'PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin', 'HOME=/tmp', 'TMPDIR=/tmp', 'GOCACHE=/tmp/go-cache', 'GOMODCACHE=/go/pkg/mod', 'GOPROXY=off', 'GOSUMDB=off', 'GOTOOLCHAIN=local', 'GOTMPDIR=/verify', 'GOMAXPROCS=2', 'CI=true', *command]
                result = run(command)
                failed = result.returncode != 0 or (job.get('require_empty_stdout', False) and bool(result.stdout.strip()))
                if job.get('require_completed_test') and not failed:
                    events = [json.loads(line) for line in result.stdout.splitlines() if line.strip()]
                    completed = sum(event.get('Action') == 'pass' and bool(event.get('Test')) for event in events)
                    check['completed_tests'] = completed
                    if not completed:
                        check.update(exit_code=result.returncode, status='BLOCKED')
                        evidence['checks'].append(check)
                        raise ValueError('Go package completed no passing test cases; no test proof')
                if job.get('require_frontend_tests') and not failed:
                    report = json.loads(result.stdout)
                    completed = report.get('numPassedTests', 0)
                    check['completed_tests'] = completed
                    if not isinstance(completed, int) or completed < 1:
                        check.update(exit_code=result.returncode, status='BLOCKED')
                        evidence['checks'].append(check)
                        raise ValueError('Frontend completed no passing test cases; no test proof')
                if job.get('require_e2e_discovery'):
                    try:
                        counts = validate_playwright_discovery(result.stdout, job.get('e2e_discovery_specs') or [])
                        check['discovery_counts'] = counts
                    except ValueError as error:
                        failed = True
                        check['discovery_error'] = str(error)
                if job.get('require_e2e_page_consumers'):
                    try:
                        if result.returncode != 0:
                            raise ValueError(
                                'E2E page consumer AST exited nonzero: ' + str(result.returncode)
                            )
                        consumers = validate_e2e_page_consumers(
                            result.stdout,
                            job.get('e2e_pages') or None,
                            stderr=result.stderr,
                        )
                        check['e2e_page_consumers'] = consumers
                    except ValueError as error:
                        failed = True
                        check['e2e_page_consumers_error'] = str(error)
                check.update(exit_code=result.returncode, status='FAIL' if failed else 'PASS')
                evidence['checks'].append(check)
                if failed:
                    evidence['status'] = 'FAIL'
                    code = 1
                    break
            else:
                if evidence['fingerprint'] != scope_fingerprint(paths):
                    raise ValueError('Selected files changed during verification')
                evidence['status'] = 'PASS' if jobs else 'SKIP'
                evidence['reason'] = 'Selected checks passed' if jobs else ('Documentation-only scope; runtime verification not needed' if paths else 'Empty scope; no verification performed')
                code = 0
    except (ValueError, OSError, subprocess.TimeoutExpired, KeyError, IndexError) as error:
        evidence.update(status='BLOCKED', reason=str(error))
    evidence['executed_count'] = sum('exit_code' in check for check in evidence['checks'])
    output = json.dumps(evidence, ensure_ascii=False, indent=2) + '\n'
    if args.evidence:
        # Never overwrite an existing artifact or follow a symlink.
        with args.evidence.open('x') as stream:
            stream.write(output)
    print(output, end='')
    return code


if __name__ == '__main__':
    sys.exit(main())
