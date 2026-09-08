#!/usr/bin/env python3
"""Offline regression tests; never invokes Docker or application tests."""
import importlib.util
import pathlib
import tempfile
import unittest
from unittest import mock
import subprocess
import io

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


if __name__ == '__main__':
    unittest.main()
