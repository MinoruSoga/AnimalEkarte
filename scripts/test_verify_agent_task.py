#!/usr/bin/env python3
"""Offline regression tests; never invokes Docker or application tests."""
import importlib.util
import io
import json
import pathlib
import subprocess
import tempfile
import unittest
from unittest import mock

SPEC = importlib.util.spec_from_file_location('verify', pathlib.Path(__file__).with_name('verify-agent-task.py'))
verify = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(verify)


class VerificationTests(unittest.TestCase):
    def test_unknown_backend_is_blocked(self):
        jobs, blocked = verify.plan(['backend/internal/repository/pet.go'])
        self.assertFalse(jobs)
        self.assertTrue(blocked)

    def test_frontend_uses_related_tests_without_empty_success(self):
        jobs, blocked = verify.plan(['frontend/src/features/pets/Pet.tsx'])
        self.assertFalse(blocked)
        self.assertIn('related', jobs[0]['command'])
        self.assertNotIn('--passWithNoTests', jobs[0]['command'])

    def test_docs_skip_and_unknown_script_blocks(self):
        self.assertEqual(verify.plan(['docs/ops/example.md']), ([], []))
        self.assertEqual(verify.plan(['frontend/src/features/manual/content/screens/01-login.md']), ([], []))
        self.assertTrue(verify.plan(['scripts/new-script.sh'])[1])
        self.assertTrue(verify.plan(['frontend/src/content/manual.md'])[1])

    def test_auth_d5_dualprocess_script_maps_to_bash_n(self):
        jobs, blocked = verify.plan(['scripts/auth-d5-dualprocess.sh'])
        self.assertFalse(blocked)
        self.assertEqual(jobs[0]['command'], ['bash', '-n', 'scripts/auth-d5-dualprocess.sh'])

    def test_reject_path_escape(self):
        for path in ('../secret', '/tmp/x', '-option', 'frontend/../../x'):
            with self.assertRaises(ValueError):
                verify.validate_path(path)

    def test_container_requires_exact_mount_and_no_network(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            info = {'State': {'Running': True}, 'HostConfig': {'NetworkMode': 'none', 'Privileged': False, 'CapDrop': ['ALL'], 'ReadonlyRootfs': True, 'SecurityOpt': ['no-new-privileges']},
                    'Mounts': [{'Type': 'bind', 'Source': str(root / 'frontend'), 'Destination': '/app', 'RW': False}],
                    'Config': {'WorkingDir': '/app'}, 'Image': 'sha256:fixture'}
            self.assertEqual(verify.container_error(info, root, 'frontend'), '')
            info['Mounts'][0]['RW'] = True
            self.assertIn('read-only', verify.container_error(info, root, 'frontend'))
            info['Mounts'][0]['RW'] = False
            info['HostConfig']['Privileged'] = True
            self.assertIn('privileged', verify.container_error(info, root, 'frontend'))
            info['HostConfig']['Privileged'] = False
            info['HostConfig']['SecurityOpt'] = ['no-new-privileges=false']
            self.assertIn('no-new-privileges', verify.container_error(info, root, 'frontend'))
            info['HostConfig']['SecurityOpt'] = ['no-new-privileges']
            info['Mounts'].append({'Type': 'tmpfs', 'Destination': '/tmp', 'RW': True})
            self.assertEqual(verify.container_error(info, root, 'frontend'), '')
            image = info.pop('Image')
            self.assertIn('image', verify.container_error(info, root, 'frontend'))
            info['Image'] = image
            info['Mounts'][0]['Source'] = '/other/frontend'
            self.assertIn('mount', verify.container_error(info, root, 'frontend'))
            info['Mounts'][0]['Source'] = str(root / 'frontend')
            info['HostConfig']['NetworkMode'] = 'default'
            self.assertIn('network', verify.container_error(info, root, 'frontend'))

    def test_go_scope_stays_bounded(self):
        jobs, blocked = verify.plan(['backend/internal/apperrors/errors.go'])
        self.assertFalse(blocked)
        self.assertEqual(jobs[0]['command'][-2:], ['-short', './internal/apperrors'])

    def test_openapi_yaml_uses_apicontract_package(self):
        jobs, blocked = verify.plan(['backend/docs/api.yaml'])
        self.assertFalse(blocked)
        self.assertEqual(jobs[0]['command'][-2:], ['-short', './internal/apicontract'])
        self.assertTrue(jobs[0]['require_completed_test'])

    def test_get_head_inventory_runs_classification_tests(self):
        jobs, blocked = verify.plan(['backend/cmd/api/testdata/get_head_permissions.json'])
        self.assertFalse(blocked)
        self.assertEqual(len(jobs), 1)
        self.assertEqual(jobs[0]['service'], 'backend')
        self.assertEqual(jobs[0]['command'], [
            'go', 'test', '-json', '-p=2', '-count=1', '-short',
            './cmd/api', '-run=^TestGETHEAD',
        ])
        self.assertTrue(jobs[0]['require_completed_test'])

    def test_other_api_testdata_stays_blocked(self):
        path = 'backend/cmd/api/testdata/unreviewed.json'
        self.assertEqual(verify.plan([path]), ([], [path]))

    def test_bootstrap_fixture_and_document_run_contract(self):
        for path in ('backend/internal/auth/testdata/first_system_admin.sql',
                     'docs/ops/deploy/FIRST_SYSTEM_ADMIN.md'):
            with self.subTest(path=path):
                jobs, blocked = verify.plan([path])
                self.assertFalse(blocked)
                self.assertEqual(len(jobs), 1)
                self.assertEqual(jobs[0]['command'], [
                    'go', 'test', '-json', '-p=2', '-count=1', '-short',
                    './internal/auth', '-run=^TestFirstSystemAdminProcedureMatchesInitSchema$',
                ])
                self.assertTrue(jobs[0]['require_completed_test'])
        unknown = 'backend/internal/auth/testdata/unreviewed.sql'
        self.assertEqual(verify.plan([unknown]), ([], [unknown]))

    def test_backend_contract_mounts_only_required_readonly_documents(self):
        for package, documents in {
            './internal/auth': ['docs/ops/deploy/FIRST_SYSTEM_ADMIN.md'],
            './internal/medicalrecord': [
                'docs/ops/deploy/LAB_DEVICE_CONNECTIVITY.md',
                'docs/architecture/adr/007-lab-device-receive-and-commit.md',
            ],
            './cmd/api': [],
        }.items():
            with self.subTest(package=package):
                command = verify.image_command('sha256:fixture', 'backend', ['go', 'test', package])
                mounts = [value for value in command if value.startswith('type=bind,')]
                self.assertEqual(len(mounts), 1 + len(documents))
                for document in documents:
                    self.assertIn('type=bind,src=' + str(verify.ROOT / document)
                                  + ',dst=/' + document + ',readonly', mounts)

    def test_backend_contract_missing_or_external_document_blocks(self):
        with tempfile.TemporaryDirectory() as directory, tempfile.TemporaryDirectory() as outside:
            root = pathlib.Path(directory)
            with mock.patch.object(verify, 'ROOT', root):
                with self.assertRaisesRegex(ValueError, 'document'):
                    verify.image_command('sha256:fixture', 'backend', ['go', 'test', './internal/auth'])
                (root / 'docs').symlink_to(outside, target_is_directory=True)
                with self.assertRaisesRegex(ValueError, 'escapes'):
                    verify.image_command('sha256:fixture', 'backend', ['go', 'test', './internal/auth'])

    def test_backend_document_change_affects_fingerprint(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            (root / 'backend/internal/auth').mkdir(parents=True)
            (root / 'backend/internal/auth/example_test.go').write_text('fixture')
            document = root / 'docs/ops/deploy/FIRST_SYSTEM_ADMIN.md'
            document.parent.mkdir(parents=True)
            document.write_text('before')
            with mock.patch.object(verify, 'ROOT', root), mock.patch.object(verify, 'git', return_value=''):
                paths = ['backend/internal/auth/example_test.go']
                before = verify.scope_fingerprint(paths)
                document.write_text('after')
                self.assertNotEqual(before, verify.scope_fingerprint(paths))

    def test_unstaged_backend_contract_document_blocks(self):
        with mock.patch.object(verify, 'git', side_effect=['', '', 'docs/ops/deploy/FIRST_SYSTEM_ADMIN.md\0']):
            with self.assertRaisesRegex(ValueError, 'unstaged or untracked'):
                verify.reject_unstaged_service_sources(['backend/internal/auth/first_system_admin_procedure_test.go'])

    def test_document_only_contract_still_checks_backend_sources(self):
        paths = ['docs/ops/deploy/FIRST_SYSTEM_ADMIN.md']
        self.assertEqual(verify.selected_services(paths), ['backend'])
        with mock.patch.object(verify, 'git', return_value='backend/internal/auth/helper.go\0'):
            with self.assertRaisesRegex(ValueError, 'unstaged or untracked'):
                verify.reject_unstaged_service_sources(paths)

    def test_backend_contract_document_symlink_blocks(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            target = root / 'source.md'
            target.write_text('fixture')
            document = root / 'docs/ops/deploy/FIRST_SYSTEM_ADMIN.md'
            document.parent.mkdir(parents=True)
            document.symlink_to(target)
            with mock.patch.object(verify, 'ROOT', root):
                with self.assertRaisesRegex(ValueError, 'symlink'):
                    verify.image_command('sha256:fixture', 'backend', ['go', 'test', './internal/auth'])

    def test_cmd_package_uses_package_tests(self):
        jobs, blocked = verify.plan(['backend/cmd/migrate/csvbundle.go'])
        self.assertFalse(blocked)
        self.assertEqual(jobs[0]['command'][-3:], ['-short', './cmd/migrate', '-skip=^TestMakefile'])

    def test_lintscan_skips_tests_outside_backend_mount(self):
        jobs, blocked = verify.plan(['backend/internal/lintscan/demo_account_label_drift_test.go'])
        self.assertFalse(blocked)
        self.assertTrue(jobs[0]['command'][-1].startswith('-skip='))
        self.assertEqual(jobs[0]['command'][-2], './internal/lintscan')

    def test_account_layout_and_worker_have_host_contracts(self):
        jobs, blocked = verify.plan(['Makefile', 'backend/worker/migrate-exec.ts', 'scripts/check-workflow-contracts.test.mjs'])
        self.assertFalse(blocked)
        commands = [tuple(job['command']) for job in jobs]
        self.assertIn(('python3', '-B', 'scripts/test_account_csv_layout.py'), commands)
        self.assertIn(('bash', 'scripts/check-test-worker-makefile.test.sh'), commands)
        self.assertIn(('node', '--test', 'scripts/check-workflow-contracts.test.mjs'), commands)

    def test_security_scan_workflow_uses_workflow_contracts(self):
        jobs, blocked = verify.plan(['.github/workflows/security-scan.yml'])
        self.assertFalse(blocked)
        self.assertEqual(jobs[0]['command'], ['node', '--test', 'scripts/check-workflow-contracts.test.mjs'])

    def test_cli_failure_has_no_pass_or_raw_output(self):
        failed = subprocess.CompletedProcess([], 1, 'sensitive stdout', 'sensitive stderr')
        with mock.patch.object(verify, 'git', return_value='fixture-head'), \
             mock.patch.object(verify, 'run', return_value=failed), \
             mock.patch('sys.argv', ['verify', '--paths', 'scripts/test_verify_agent_task.py']), \
             mock.patch('sys.stdout', new_callable=io.StringIO) as output:
            self.assertEqual(verify.main(), 1)
            self.assertIn('"status": "FAIL"', output.getvalue())
            self.assertNotIn('sensitive', output.getvalue())

    def test_frontend_package_manifest_maps_to_pnpm_audit(self):
        for path in ('frontend/package.json', 'frontend/pnpm-lock.yaml'):
            with self.subTest(path=path):
                jobs, blocked = verify.plan([path])
                self.assertFalse(blocked)
                self.assertEqual(jobs, [{
                    'service': 'host',
                    'command': [
                        'docker', 'compose', '--env-file', '.env.local',
                        'exec', '-T', 'frontend',
                        'pnpm', 'audit', '--audit-level', 'moderate',
                    ],
                }])
        jobs, blocked = verify.plan(['frontend/package.json', 'frontend/pnpm-lock.yaml'])
        self.assertFalse(blocked)
        self.assertEqual(len(jobs), 1)

    def test_unmapped_scope_never_executes(self):
        with mock.patch.object(verify, 'git', return_value='fixture-head'), \
             mock.patch.object(verify, 'run') as runner, \
             mock.patch('sys.argv', ['verify', '--paths', 'scripts/new.py']), \
             mock.patch('sys.stdout', new_callable=io.StringIO):
            self.assertEqual(verify.main(), 2)
            runner.assert_not_called()

    def test_empty_scope_skips_without_execution(self):
        with mock.patch.object(verify, 'git', return_value=''), \
             mock.patch.object(verify, 'run') as runner, \
             mock.patch('sys.argv', ['verify', '--staged']), \
             mock.patch('sys.stdout', new_callable=io.StringIO) as output:
            self.assertEqual(verify.main(), 0)
            self.assertIn('"status": "SKIP"', output.getvalue())
            self.assertIn('"executed_count": 0', output.getvalue())
            runner.assert_not_called()

    def test_dirty_staged_content_blocks(self):
        with mock.patch.object(verify, 'git', side_effect=['docs/README.md\0', 'docs/README.md\0']), \
             mock.patch.object(verify, 'run') as runner, \
             mock.patch('sys.argv', ['verify', '--staged']), \
             mock.patch('sys.stdout', new_callable=io.StringIO):
            self.assertEqual(verify.main(), 2)
            runner.assert_not_called()

    def test_missing_container_blocks(self):
        result = subprocess.CompletedProcess([], 1, '', 'sensitive detail')
        with mock.patch.object(verify, 'run', return_value=result):
            with self.assertRaisesRegex(ValueError, 'inspection failed'):
                verify.inspect_container('frontend', 'missing')

    def test_offline_image_command_cannot_pull_or_write_source(self):
        command = verify.image_command('sha256:fixture', 'frontend', ['node', 'test.js'], 'deps')
        for flag in ('--pull=never', '--network=none', '--read-only', '--cap-drop=ALL', '--security-opt=no-new-privileges'):
            self.assertIn(flag, command)
        self.assertIn('type=volume,src=deps,dst=/app/node_modules,readonly,volume-nocopy', command)
        self.assertIn('type=bind,src=' + str(verify.ROOT / 'frontend') + ',dst=/app,readonly', command)
        self.assertIn('/usr/bin/env', command)
        self.assertIn('-i', command)
        self.assertNotIn('--env-file', command)

    def test_image_inspection_is_local_only(self):
        import json
        result = subprocess.CompletedProcess([], 0, json.dumps([{'Id': 'sha256:fixture', 'Config': {}}]), '')
        with mock.patch.object(verify, 'run', return_value=result) as runner:
            self.assertEqual(verify.inspect_image('cached:tag'), 'sha256:fixture')
            runner.assert_called_once_with(['docker', 'image', 'inspect', 'cached:tag'])

    def test_external_symlink_and_control_characters_rejected(self):
        for path in ('x\x00y', 'x\ny'):
            with self.assertRaises(ValueError):
                verify.validate_path(path)
        with tempfile.TemporaryDirectory() as directory, tempfile.TemporaryDirectory() as outside:
            root = pathlib.Path(directory)
            (root / 'escape').symlink_to(outside, target_is_directory=True)
            with mock.patch.object(verify, 'ROOT', root):
                with self.assertRaises(ValueError):
                    verify.validate_path('escape/file.ts')

    def test_timeout_cleans_only_owned_ephemeral_container(self):
        timeout = subprocess.TimeoutExpired(['docker', 'run'], 300)
        with mock.patch.object(verify.subprocess, 'run', side_effect=[timeout, subprocess.CompletedProcess([], 0)]) as runner:
            with self.assertRaises(subprocess.TimeoutExpired):
                verify.run(['docker', 'run', '--rm', 'sha256:fixture'])
            created = runner.call_args_list[0].args[0]
            cleanup = runner.call_args_list[1].args[0]
            self.assertEqual(cleanup, ['docker', 'rm', '-f', created[3]])
            self.assertTrue(created[3].startswith('ae-verify-'))

    def test_implicit_image_volume_is_blocked(self):
        import json
        result = subprocess.CompletedProcess([], 0, json.dumps([{'Id': 'sha256:fixture', 'Config': {'Volumes': {'/data': {}}}}]), '')
        with mock.patch.object(verify, 'run', return_value=result):
            with self.assertRaisesRegex(ValueError, 'implicit volumes'):
                verify.inspect_image('cached:tag')

    def test_staged_service_rejects_unstaged_helpers(self):
        with mock.patch.object(verify, 'git', return_value='backend/internal/x/helper.go\0'):
            with self.assertRaisesRegex(ValueError, 'unstaged or untracked'):
                verify.reject_unstaged_service_sources(['backend/internal/x/x.go'])

    def test_service_fingerprint_changes_for_helper(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            (root / 'backend').mkdir()
            (root / 'backend/x.go').write_text('x')
            (root / 'backend/helper.go').write_text('before')
            with mock.patch.object(verify, 'ROOT', root), mock.patch.object(verify, 'git', return_value='backend/x.go\0backend/helper.go\0'):
                before = verify.scope_fingerprint(['backend/x.go'])
                (root / 'backend/helper.go').write_text('after')
                self.assertNotEqual(before, verify.scope_fingerprint(['backend/x.go']))

    def test_local_selectors_and_env_precedence(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            (root / '.claude').mkdir()
            config = root / '.claude/verification.local.json'
            config.write_text('{"backend_image":"cached:local"}')
            with mock.patch.object(verify, 'ROOT', root), mock.patch.dict(verify.os.environ, {'AGENT_VERIFY_BACKEND_IMAGE': 'cached:env'}, clear=True):
                self.assertEqual(verify.local_defaults()['backend_image'], 'cached:env')
                config.write_text('{"command":"do something"}')
                with self.assertRaisesRegex(ValueError, 'unsupported keys'):
                    verify.local_defaults()

    def test_go_zero_test_success_is_blocked(self):
        result = subprocess.CompletedProcess([], 0, '{"Action":"pass","Package":"fixture"}\n', '')
        job = {'service': 'host', 'command': ['fixture'], 'require_completed_test': True}
        with mock.patch.object(verify, 'git', return_value='fixture'), \
             mock.patch.object(verify, 'plan', return_value=([job], [])), \
             mock.patch.object(verify, 'run', return_value=result), \
             mock.patch('sys.argv', ['verify', '--paths', 'docs/README.md']), \
             mock.patch('sys.stdout', new_callable=io.StringIO) as output:
            self.assertEqual(verify.main(), 2)
            self.assertIn('"completed_tests": 0', output.getvalue())
            self.assertIn('"status": "BLOCKED"', output.getvalue())

    def test_go_tmpfs_allows_compiled_test_execution(self):
        command = verify.image_command('sha256:fixture', 'backend', ['go', 'test'])
        self.assertIn('/verify:rw,exec,nosuid,nodev,size=1073741824', command)
        self.assertIn('GOTMPDIR=/verify', command)
        self.assertIn('GOMAXPROCS=2', command)

    def test_native_frontend_reporter_and_config_mapping(self):
        jobs, blocked = verify.plan(['frontend/src/example.ts'])
        self.assertFalse(blocked)
        self.assertIn('--configLoader', jobs[0]['command'])
        self.assertIn('native', jobs[0]['command'])
        self.assertIn('--reporter=json', jobs[0]['command'])
        self.assertTrue(jobs[0]['require_frontend_tests'])
        jobs, blocked = verify.plan(['frontend/vite.config.ts'])
        self.assertFalse(blocked)
        self.assertEqual(jobs[0]['command'], ['node', '--test', 'scripts/vite-native-config.test.mjs'])

    def test_empty_mount_scaffold_preserves_existing_contents(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            (root / 'frontend').mkdir()
            with mock.patch.object(verify, 'ROOT', root), mock.patch.object(verify, 'run', return_value=subprocess.CompletedProcess([], 0)):
                verify.prepare_dependency_mountpoint('frontend', 'explicit-deps')
                target = root / 'frontend/node_modules'
                self.assertTrue(target.is_dir())
                (target / 'sentinel').write_text('preserve')
                verify.prepare_dependency_mountpoint('frontend', 'explicit-deps')
                self.assertEqual((target / 'sentinel').read_text(), 'preserve')

    def test_frontend_zero_test_report_is_blocked(self):
        result = subprocess.CompletedProcess([], 0, '{"numPassedTests":0}', '')
        job = {'service': 'host', 'command': ['fixture'], 'require_frontend_tests': True}
        with mock.patch.object(verify, 'git', return_value='fixture'), \
             mock.patch.object(verify, 'plan', return_value=([job], [])), \
             mock.patch.object(verify, 'run', return_value=result), \
             mock.patch('sys.argv', ['verify', '--paths', 'docs/README.md']), \
             mock.patch('sys.stdout', new_callable=io.StringIO):
            self.assertEqual(verify.main(), 2)

    def test_harness_configuration_has_bounded_test_mapping(self):
        jobs, blocked = verify.plan(['.gitignore', '.mcp.json', '.claude/settings.json', '.claude/scripts/mirror_safety.py'])
        self.assertFalse(blocked)
        self.assertEqual(len(jobs), 3)

    def test_hook_secret_output_is_redacted(self):
        source = pathlib.Path(__file__).resolve().parents[1].joinpath('.githooks/lib/check-secrets.sh').read_text()
        self.assertNotIn('gitleaks:latest', source)
        self.assertNotIn('git grep -E -n -I --cached -e "$PATTERN" -- "$f" || true', source)
        self.assertIn('--redact', source)

    def test_e2e_five_path_scope_maps_nonempty_offline_checks(self):
        paths = [
            'frontend/e2e/pages/accounting-page.ts',
            'frontend/e2e/pages/settings-master-page.ts',
            'frontend/e2e/s09-closing-time-boundaries.spec.ts',
            'frontend/e2e/v04-settings-master-forms.spec.ts',
            'frontend/scripts/run-e2e.sh',
        ]
        jobs, blocked = verify.plan(paths)
        self.assertFalse(blocked, blocked)
        self.assertTrue(jobs)
        services = {job['service'] for job in jobs}
        self.assertIn('frontend', services)
        self.assertIn('host', services)
        flat = [' '.join(job['command']) for job in jobs]
        self.assertTrue(any('prettier' in command for command in flat))
        self.assertTrue(any('eslint' in command for command in flat))
        self.assertTrue(any('tsc' in command or 'ae-e2e-tsconfig' in command for command in flat))
        self.assertTrue(any(job['command'][:3] == ['bash', '-n', 'frontend/scripts/run-e2e.sh'] for job in jobs))
        self.assertTrue(any('--check-e2e-scope' in job['command'] for job in jobs))
        self.assertTrue(any(job.get('require_e2e_discovery') for job in jobs))
        for job in jobs:
            joined = ' '.join(job['command'])
            self.assertNotIn('npm install', joined)
            self.assertNotIn('--network=host', joined)
            if 'cli.js' in joined and 'test' in job['command']:
                self.assertIn('--list', job['command'])
                self.assertNotIn('--headed', job['command'])
                self.assertNotIn('--debug', job['command'])

    def test_e2e_page_object_without_consumer_plans_ast_job(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            page = root / 'frontend/e2e/pages/orphan-page.ts'
            page.parent.mkdir(parents=True)
            page.write_text('export class Orphan {}\n')
            (root / 'frontend/e2e').mkdir(exist_ok=True)
            (root / 'frontend/e2e/unrelated.spec.ts').write_text(
                'import { test } from "@playwright/test";\ntest("x", async () => {});\n'
            )
            with mock.patch.object(verify, 'ROOT', root):
                jobs, blocked = verify.plan(['frontend/e2e/pages/orphan-page.ts'])
                self.assertFalse(blocked)
                self.assertTrue(jobs)
                ast_jobs = [job for job in jobs if job.get('require_e2e_page_consumers')]
                self.assertEqual(len(ast_jobs), 1)
                self.assertIn('scripts/verify-e2e-page-consumers.mjs', ast_jobs[0]['command'])
                self.assertIn('e2e/pages/orphan-page.ts', ast_jobs[0]['e2e_pages'])

    def test_e2e_deleted_or_unsupported_paths_block(self):
        missing = 'frontend/e2e/missing-spec.spec.ts'
        jobs, blocked = verify.plan([missing])
        self.assertFalse(jobs)
        self.assertEqual(blocked, [missing])
        unsupported = 'frontend/e2e/helpers/clinical-env.ts'
        jobs, blocked = verify.plan([unsupported])
        self.assertFalse(jobs)
        self.assertEqual(blocked, [unsupported])
        jobs, blocked = verify.plan(['frontend/e2e/notes.md'])
        self.assertFalse(jobs)
        self.assertEqual(blocked, ['frontend/e2e/notes.md'])

    def test_e2e_mixed_with_docs_keeps_e2e_checks_and_skips_docs(self):
        paths = [
            'docs/ops/testing/UAT-DOMAIN-STATUS.md',
            'frontend/e2e/s09-closing-time-boundaries.spec.ts',
            'frontend/scripts/run-e2e.sh',
        ]
        jobs, blocked = verify.plan(paths)
        self.assertFalse(blocked)
        self.assertTrue(jobs)
        self.assertTrue(any('eslint' in ' '.join(job['command']) for job in jobs))
        self.assertTrue(any(job['command'][:3] == ['bash', '-n', 'frontend/scripts/run-e2e.sh'] for job in jobs))

    def test_run_e2e_env_forward_contract_rejects_secret_values_on_argv(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            script = root / 'frontend/scripts/run-e2e.sh'
            script.parent.mkdir(parents=True)
            script.write_text(
                'DOCKER_ENV="-e PLAYWRIGHT_TEST_BASE_URL=${BASE_URL}"\n'
                'DOCKER_ENV="$DOCKER_ENV -e UAT_SYNTHETIC_CLOSING_PASSWORD=$UAT_SYNTHETIC_CLOSING_PASSWORD"\n'
                'DOCKER_ENV="$DOCKER_ENV -e UAT_SYNTHETIC_CLOSING_API_BASE=$UAT_SYNTHETIC_CLOSING_API_BASE"\n'
                'docker run $DOCKER_ENV image playwright test "$@"\n'
            )
            with mock.patch.object(verify, 'ROOT', root):
                with self.assertRaisesRegex(ValueError, 'name-only'):
                    verify.check_e2e_scope(['frontend/scripts/run-e2e.sh'])

    def test_check_e2e_scope_accepts_current_runner_forwarding(self):
        verify.check_e2e_scope(['frontend/scripts/run-e2e.sh'])

    def test_plan_maps_vite_native_config_script(self):
        jobs, blocked = verify.plan(['frontend/scripts/vite-native-config.test.mjs'])
        self.assertFalse(blocked)
        self.assertEqual(jobs[0]['command'], ['node', '--test', 'scripts/vite-native-config.test.mjs'])

    def test_page_only_plan_includes_all_specs_and_ast_job(self):
        jobs, blocked = verify.plan(['frontend/e2e/pages/accounting-page.ts'])
        self.assertFalse(blocked)
        all_specs = [path.removeprefix('frontend/') for path in verify.list_e2e_spec_paths()]
        self.assertGreaterEqual(len(all_specs), 1)
        tsc = next(
            job for job in jobs
            if job['command'][:2] == ['node', 'node_modules/typescript/bin/tsc']
            and 'e2e/tsconfig.json' in job['command']
        )
        self.assertIn('--noEmit', tsc['command'])
        lint = next(job for job in jobs if job['command'][:1] == ['node'] and 'eslint.js' in job['command'][1])
        fmt = next(job for job in jobs if job['command'][:1] == ['node'] and 'prettier.cjs' in job['command'][1])
        for spec in all_specs:
            self.assertIn(spec, lint['command'], spec)
            self.assertIn(spec, fmt['command'], spec)
        discovery = next(job for job in jobs if job.get('require_e2e_discovery'))
        for spec in all_specs:
            self.assertIn(spec, discovery['e2e_discovery_specs'], spec)
        ast = next(job for job in jobs if job.get('require_e2e_page_consumers'))
        self.assertEqual(ast['command'][:2], ['node', 'scripts/verify-e2e-page-consumers.mjs'])
        self.assertIn('--page', ast['command'])
        self.assertIn('e2e/pages/accounting-page.ts', ast['command'])
        self.assertIn('e2e/pages/accounting-page.ts', ast['e2e_pages'])

    def test_plan_maps_e2e_page_consumer_scripts(self):
        for path in (
            'frontend/scripts/verify-e2e-page-consumers.mjs',
            'frontend/scripts/verify-e2e-page-consumers.test.mjs',
        ):
            jobs, blocked = verify.plan([path])
            self.assertFalse(blocked, path)
            self.assertEqual(
                jobs[0]['command'],
                ['node', '--test', 'scripts/verify-e2e-page-consumers.test.mjs'],
            )

    def test_plan_does_not_execute_ast_scan(self):
        with mock.patch.object(verify, 'run') as mocked_run:
            jobs, blocked = verify.plan(['frontend/e2e/pages/accounting-page.ts'])
            self.assertFalse(blocked)
            self.assertTrue(any(job.get('require_e2e_page_consumers') for job in jobs))
            mocked_run.assert_not_called()

    def test_validate_e2e_page_consumers_accepts_valid_payload(self):
        payload = {
            'ok': True,
            'pages': [{
                'page': 'e2e/pages/accounting-page.ts',
                'consumers': [{
                    'file': 'e2e/accounting-flow.spec.ts',
                    'form': 'static-import',
                    'specifier': './pages/accounting-page',
                }],
                'blocking': [],
            }],
        }
        evidence = verify.validate_e2e_page_consumers(
            json.dumps(payload),
            ['e2e/pages/accounting-page.ts'],
        )
        self.assertEqual(len(evidence['e2e/pages/accounting-page.ts']), 1)

    def test_validate_e2e_page_consumers_fail_closed_matrix(self):
        valid_consumer = {
            'file': 'e2e/good.spec.ts',
            'form': 'static-import',
            'specifier': './pages/accounting-page',
        }
        with self.assertRaisesRegex(ValueError, 'empty output'):
            verify.validate_e2e_page_consumers('')
        with self.assertRaisesRegex(ValueError, 'malformed JSON|not JSON'):
            verify.validate_e2e_page_consumers('not-json')
        with self.assertRaisesRegex(ValueError, 'ok!=true'):
            verify.validate_e2e_page_consumers(json.dumps({
                'ok': False,
                'pages': [{'page': 'e2e/pages/accounting-page.ts', 'consumers': [valid_consumer], 'blocking': []}],
            }))
        with self.assertRaisesRegex(ValueError, 'no spec consumer'):
            verify.validate_e2e_page_consumers(json.dumps({
                'ok': True,
                'pages': [{'page': 'e2e/pages/orphan-page.ts', 'consumers': [], 'blocking': []}],
            }))
        with self.assertRaisesRegex(ValueError, 'blocking references'):
            verify.validate_e2e_page_consumers(json.dumps({
                'ok': True,
                'pages': [{
                    'page': 'e2e/pages/accounting-page.ts',
                    'consumers': [valid_consumer],
                    'blocking': [{'file': 'e2e/bad.spec.ts', 'form': 'dynamic-import'}],
                }],
            }))

    def test_validate_e2e_page_consumers_rejects_forged_forms_paths_and_set_mismatch(self):
        page = 'e2e/pages/accounting-page.ts'
        valid_consumer = {
            'file': 'e2e/good.spec.ts',
            'form': 'static-import',
            'specifier': './pages/accounting-page',
        }

        def payload(pages):
            return json.dumps({'ok': True, 'pages': pages})

        with self.assertRaisesRegex(ValueError, 'form|static-import|require'):
            verify.validate_e2e_page_consumers(payload([{
                'page': page,
                'consumers': [{**valid_consumer, 'form': 'require'}],
                'blocking': [],
            }]), [page])
        with self.assertRaisesRegex(ValueError, 'form|static-import|dynamic'):
            verify.validate_e2e_page_consumers(payload([{
                'page': page,
                'consumers': [{**valid_consumer, 'form': 'dynamic-import'}],
                'blocking': [],
            }]), [page])
        with self.assertRaisesRegex(ValueError, 'extra|set|mismatch|unexpected'):
            verify.validate_e2e_page_consumers(payload([
                {
                    'page': page,
                    'consumers': [valid_consumer],
                    'blocking': [],
                },
                {
                    'page': 'e2e/pages/extra-page.ts',
                    'consumers': [{
                        'file': 'e2e/extra.spec.ts',
                        'form': 'static-import',
                        'specifier': './pages/extra-page',
                    }],
                    'blocking': [],
                },
            ]), [page])
        with self.assertRaisesRegex(ValueError, 'duplicate'):
            verify.validate_e2e_page_consumers(payload([
                {'page': page, 'consumers': [valid_consumer], 'blocking': []},
                {'page': page, 'consumers': [valid_consumer], 'blocking': []},
            ]), [page])
        with self.assertRaisesRegex(ValueError, 'page|canonical|identity|path'):
            verify.validate_e2e_page_consumers(payload([{
                'page': 'frontend/e2e/pages/accounting-page.ts',
                'consumers': [valid_consumer],
                'blocking': [],
            }]), ['frontend/e2e/pages/accounting-page.ts'])
        with self.assertRaisesRegex(ValueError, 'page|canonical|identity|path|\\.\\.'):
            verify.validate_e2e_page_consumers(payload([{
                'page': 'e2e/pages/../secret.ts',
                'consumers': [valid_consumer],
                'blocking': [],
            }]))
        with self.assertRaisesRegex(ValueError, 'consumer|spec|file'):
            verify.validate_e2e_page_consumers(payload([{
                'page': page,
                'consumers': [{
                    'file': 'e2e/helpers/not-a-spec.ts',
                    'form': 'static-import',
                    'specifier': './pages/accounting-page',
                }],
                'blocking': [],
            }]), [page])
        with self.assertRaisesRegex(ValueError, 'specifier'):
            verify.validate_e2e_page_consumers(payload([{
                'page': page,
                'consumers': [{
                    'file': 'e2e/good.spec.ts',
                    'form': 'static-import',
                    'specifier': '',
                }],
                'blocking': [],
            }]), [page])
        with self.assertRaisesRegex(ValueError, 'importer-relative|specifier'):
            verify.validate_e2e_page_consumers(payload([{
                'page': page,
                'consumers': [{
                    'file': 'e2e/good.spec.ts',
                    'form': 'static-import',
                    'specifier': '/app/e2e/pages/accounting-page',
                }],
                'blocking': [],
            }]), [page])
        with self.assertRaisesRegex(ValueError, 'forbidden|specifier'):
            verify.validate_e2e_page_consumers(payload([{
                'page': page,
                'consumers': [{
                    'file': 'e2e/good.spec.ts',
                    'form': 'static-import',
                    'specifier': './pages/accounting-page?x=1',
                }],
                'blocking': [],
            }]), [page])
        with self.assertRaisesRegex(ValueError, 'missing|set|mismatch'):
            verify.validate_e2e_page_consumers(payload([{
                'page': page,
                'consumers': [valid_consumer],
                'blocking': [],
            }]), [page, 'e2e/pages/settings-master-page.ts'])

    def test_raw_canonical_rejects_dot_slash_and_double_slash_identities(self):
        """PurePosixPath collapses // and /. before parts checks — forged spelling must fail."""
        self.assertFalse(verify._is_canonical_e2e_page('e2e/pages/./accounting-page.ts'))
        self.assertFalse(verify._is_canonical_e2e_page('e2e/pages//accounting-page.ts'))
        self.assertTrue(verify._is_canonical_e2e_page('e2e/pages/accounting-page.ts'))

        self.assertFalse(verify._is_e2e_spec_consumer_file('e2e//forged.spec.ts'))
        self.assertFalse(verify._is_e2e_spec_consumer_file('e2e/./good.spec.ts'))
        self.assertTrue(verify._is_e2e_spec_consumer_file('e2e/good.spec.ts'))

        def payload(page_identity, file_path, specifier='./pages/accounting-page'):
            return json.dumps({
                'ok': True,
                'pages': [{
                    'page': page_identity,
                    'consumers': [{
                        'file': file_path,
                        'form': 'static-import',
                        'specifier': specifier,
                    }],
                    'blocking': [],
                }],
            })

        with self.assertRaisesRegex(ValueError, 'canonical|identity|path|raw'):
            verify.validate_e2e_page_consumers(
                payload('e2e/pages/./accounting-page.ts', 'e2e/good.spec.ts'),
            )
        with self.assertRaisesRegex(ValueError, 'canonical|identity|path|raw'):
            verify.validate_e2e_page_consumers(
                payload('e2e/pages//accounting-page.ts', 'e2e/good.spec.ts'),
            )
        with self.assertRaisesRegex(ValueError, 'canonical|consumer|spec|file|raw|identity|path'):
            verify.validate_e2e_page_consumers(
                payload('e2e/pages/accounting-page.ts', 'e2e//forged.spec.ts'),
                ['e2e/pages/accounting-page.ts'],
            )
        with self.assertRaisesRegex(ValueError, 'canonical|consumer|spec|file|raw|identity|path'):
            verify.validate_e2e_page_consumers(
                payload('e2e/pages/accounting-page.ts', 'e2e/./good.spec.ts'),
                ['e2e/pages/accounting-page.ts'],
            )

    def test_raw_canonical_rejects_nul_c0_and_del_identities(self):
        """C0 controls and DEL survive PurePosixPath; raw identities must still fail closed."""
        page = 'e2e/pages/accounting-page.ts'
        nul_file = 'e2e/good\x00.spec.ts'
        c0_file = 'e2e/good\x01.spec.ts'
        del_file = 'e2e/good\x7f.spec.ts'
        nul_page = 'e2e/pages/accounting\x00-page.ts'
        del_page = 'e2e/pages/accounting\x7f-page.ts'

        self.assertFalse(verify._is_raw_canonical_posix(nul_file))
        self.assertFalse(verify._is_raw_canonical_posix(c0_file))
        self.assertFalse(verify._is_raw_canonical_posix(del_file))
        self.assertFalse(verify._is_e2e_spec_consumer_file(nul_file))
        self.assertFalse(verify._is_e2e_spec_consumer_file(c0_file))
        self.assertFalse(verify._is_e2e_spec_consumer_file(del_file))
        self.assertFalse(verify._is_canonical_e2e_page(nul_page))
        self.assertFalse(verify._is_canonical_e2e_page(del_page))
        self.assertTrue(verify._is_raw_canonical_posix('e2e/good.spec.ts'))

        def payload(page_identity, file_path, specifier='./pages/accounting-page'):
            return json.dumps({
                'ok': True,
                'pages': [{
                    'page': page_identity,
                    'consumers': [{
                        'file': file_path,
                        'form': 'static-import',
                        'specifier': specifier,
                    }],
                    'blocking': [],
                }],
            })

        with self.assertRaisesRegex(ValueError, 'canonical|consumer|spec|file|raw|identity|path'):
            verify.validate_e2e_page_consumers(payload(page, nul_file), [page])
        with self.assertRaisesRegex(ValueError, 'canonical|consumer|spec|file|raw|identity|path'):
            verify.validate_e2e_page_consumers(payload(page, c0_file), [page])
        with self.assertRaisesRegex(ValueError, 'canonical|consumer|spec|file|raw|identity|path'):
            verify.validate_e2e_page_consumers(payload(page, del_file), [page])
        with self.assertRaisesRegex(ValueError, 'canonical|identity|path|raw'):
            verify.validate_e2e_page_consumers(payload(nul_page, 'e2e/good.spec.ts'))
        with self.assertRaisesRegex(ValueError, 'canonical|identity|path|raw'):
            verify.validate_e2e_page_consumers(payload(del_page, 'e2e/good.spec.ts'))

    def test_specifier_must_resolve_exactly_to_claimed_page(self):
        page = 'e2e/pages/accounting-page.ts'

        def payload(file_path, specifier):
            return json.dumps({
                'ok': True,
                'pages': [{
                    'page': page,
                    'consumers': [{
                        'file': file_path,
                        'form': 'static-import',
                        'specifier': specifier,
                    }],
                    'blocking': [],
                }],
            })

        with self.assertRaisesRegex(ValueError, 'specifier|resolve|mismatch|page'):
            verify.validate_e2e_page_consumers(
                payload('e2e/good.spec.ts', './pages/other-page'),
                [page],
            )
        # Nested ../ control must still pass when resolution equals the claimed page.
        evidence = verify.validate_e2e_page_consumers(
            payload('e2e/subdir/nested.spec.ts', '../pages/accounting-page'),
            [page],
        )
        self.assertEqual(len(evidence[page]), 1)
        with self.assertRaisesRegex(ValueError, 'specifier|resolve|escape|e2e|page'):
            verify.validate_e2e_page_consumers(
                payload('e2e/good.spec.ts', '../secret'),
                [page],
            )

    def test_python_has_no_regex_module_reference_authority(self):
        source = pathlib.Path(verify.__file__).read_text(encoding='utf-8')
        self.assertNotIn('E2E_MODULE_REFERENCE_PATTERNS', source)
        self.assertNotIn('aliases_page', source)
        self.assertNotIn('def e2e_module_references', source)
        self.assertNotIn('def e2e_spec_consumers', source)
        # Former FP: helper URL substring must not be Python consumer authority.
        self.assertNotIn("f'pages/{module}' in normalized", source)
        self.assertNotIn('endswith(f\'pages/{module}\')', source)

    def test_check_e2e_scope_page_only_requires_existence(self):
        verify.check_e2e_scope(['frontend/e2e/pages/accounting-page.ts'])
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            missing = 'frontend/e2e/pages/missing-page.ts'
            with mock.patch.object(verify, 'ROOT', root):
                with self.assertRaisesRegex(ValueError, 'missing'):
                    verify.check_e2e_scope([missing])



if __name__ == '__main__':
    unittest.main()
