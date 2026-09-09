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

    def test_e2e_page_object_without_consumer_is_blocked(self):
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
                self.assertFalse(jobs)
                self.assertEqual(blocked, ['frontend/e2e/pages/orphan-page.ts'])

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

    def test_page_only_plan_includes_real_consumer_specs_in_tsc(self):
        jobs, blocked = verify.plan(['frontend/e2e/pages/accounting-page.ts'])
        self.assertFalse(blocked)
        tsc = next(job for job in jobs if job['command'] and job['command'][0] == 'sh' and 'ae-e2e-tsc' in job['command'])
        self.assertIn('e2e/pages/accounting-page.ts', tsc['command'])
        self.assertIn('e2e/s09-closing-time-boundaries.spec.ts', tsc['command'])
        self.assertIn('e2e/accounting-flow.spec.ts', tsc['command'])
        self.assertIn('e2e/accounting-smoke.spec.ts', tsc['command'])
        discovery = next(job for job in jobs if job.get('require_e2e_discovery'))
        self.assertIn('e2e/s09-closing-time-boundaries.spec.ts', discovery['e2e_discovery_specs'])
        self.assertIn('e2e/accounting-flow.spec.ts', discovery['e2e_discovery_specs'])

    def test_comment_only_page_import_is_not_a_consumer(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            page = root / 'frontend/e2e/pages/orphan-page.ts'
            page.parent.mkdir(parents=True)
            page.write_text('export class Orphan {}\n')
            (root / 'frontend/e2e').mkdir(exist_ok=True)
            (root / 'frontend/e2e/fake.spec.ts').write_text(
                '// import { Orphan } from "./pages/orphan-page";\n'
                'import { test } from "@playwright/test";\n'
                'test("x", async () => {});\n'
            )
            with mock.patch.object(verify, 'ROOT', root):
                self.assertEqual(verify.e2e_spec_consumers('frontend/e2e/pages/orphan-page.ts'), [])
                jobs, blocked = verify.plan(['frontend/e2e/pages/orphan-page.ts'])
                self.assertFalse(jobs)
                self.assertEqual(blocked, ['frontend/e2e/pages/orphan-page.ts'])

    def test_discovery_rejects_missing_or_zero_registered_tests(self):
        payload = {
            'config': {'rootDir': '/app/e2e'},
            'suites': [{
                'specs': [{
                    'file': 'valid.spec.ts',
                    'tests': [{'title': 'ok'}],
                }],
            }],
        }
        counts = verify.validate_playwright_discovery(
            json.dumps(payload),
            ['e2e/valid.spec.ts'],
        )
        self.assertEqual(counts['frontend/e2e/valid.spec.ts'], 1)
        with self.assertRaisesRegex(ValueError, 'no registered tests'):
            verify.validate_playwright_discovery(
                json.dumps(payload),
                ['e2e/valid.spec.ts', 'e2e/empty.spec.ts'],
            )
        with self.assertRaisesRegex(ValueError, 'malformed|empty|not JSON'):
            verify.validate_playwright_discovery('', ['e2e/valid.spec.ts'])

    def test_discovery_rejects_same_basename_aliasing(self):
        payload = {
            'config': {'rootDir': '/app/e2e'},
            'suites': [{
                'specs': [{
                    'file': 'group-a/shared.spec.ts',
                    'tests': [{'title': 'valid'}],
                }],
            }],
        }
        with self.assertRaisesRegex(ValueError, 'group-b/shared.spec.ts'):
            verify.validate_playwright_discovery(
                json.dumps(payload),
                ['e2e/group-a/shared.spec.ts', 'e2e/group-b/shared.spec.ts'],
            )
        both = {
            'config': {'rootDir': '/app/e2e'},
            'suites': [{
                'specs': [
                    {'file': 'group-a/shared.spec.ts', 'tests': [{'title': 'a'}]},
                    {'file': 'group-b/shared.spec.ts', 'tests': [{'title': 'b'}]},
                ],
            }],
        }
        counts = verify.validate_playwright_discovery(
            json.dumps(both),
            ['e2e/group-a/shared.spec.ts', 'e2e/group-b/shared.spec.ts'],
        )
        self.assertEqual(counts['frontend/e2e/group-a/shared.spec.ts'], 1)
        self.assertEqual(counts['frontend/e2e/group-b/shared.spec.ts'], 1)

    def test_string_literal_is_not_an_import_specifier(self):
        text = 'const message = "import { AccountingPage } from \'./pages/accounting-page\'";\n'
        self.assertEqual(verify.e2e_import_specifiers(text), [])
        text_real = 'import { AccountingPage } from "./pages/accounting-page";\n'
        self.assertEqual(verify.e2e_import_specifiers(text_real), ['./pages/accounting-page'])

    def test_regex_literal_is_not_an_import_specifier(self):
        text = 'const pattern = /import { AccountingPage } from ".\\/pages\\/accounting-page"/;\n'
        self.assertEqual(verify.e2e_import_specifiers(text), [])

    def test_discovery_rejects_dot_prefix_alias(self):
        payload = {
            'config': {'rootDir': '/app/e2e'},
            'suites': [{
                'specs': [{
                    'file': '.group-a/shared.spec.ts',
                    'tests': [{'title': 'valid'}],
                }],
            }],
        }
        with self.assertRaisesRegex(ValueError, 'ambiguous|no registered tests|out of root'):
            verify.validate_playwright_discovery(
                json.dumps(payload),
                ['e2e/group-a/shared.spec.ts'],
            )

    def test_nested_relative_import_resolves_against_importer(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            page = root / 'frontend/e2e/pages/accounting-page.ts'
            page.parent.mkdir(parents=True)
            page.write_text('export class AccountingPage {}\n')
            nested = root / 'frontend/e2e/subdir/foo.spec.ts'
            nested.parent.mkdir(parents=True)
            nested.write_text(
                'import { AccountingPage } from "./pages/accounting-page";\n'
                'import { test } from "@playwright/test";\n'
                'test("x", async () => {});\n'
            )
            with mock.patch.object(verify, 'ROOT', root):
                self.assertEqual(verify.e2e_spec_consumers('frontend/e2e/pages/accounting-page.ts'), [])

    def test_mixed_supported_and_unsupported_page_reference_blocks(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            page = root / 'frontend/e2e/pages/accounting-page.ts'
            page.parent.mkdir(parents=True)
            page.write_text('export class AccountingPage {}\n')
            good = root / 'frontend/e2e/good.spec.ts'
            bad = root / 'frontend/e2e/bad.spec.ts'
            good.write_text(
                'import { AccountingPage } from "./pages/accounting-page";\n'
                'import { test } from "@playwright/test";\n'
                'test("x", async () => {});\n'
            )
            bad.write_text(
                'import { AccountingPage } from "@/e2e/pages/accounting-page";\n'
                'import { test } from "@playwright/test";\n'
                'test("y", async () => {});\n'
            )
            with mock.patch.object(verify, 'ROOT', root):
                with self.assertRaisesRegex(ValueError, 'unsupported import topology'):
                    verify.e2e_spec_consumers('frontend/e2e/pages/accounting-page.ts')


if __name__ == '__main__':
    unittest.main()
