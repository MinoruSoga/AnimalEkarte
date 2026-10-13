#!/usr/bin/env python3
"""Mock-only STG entrypoint contracts; no Docker, DB, or real CSV import."""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
PROFILE = 'STG_UAT_CSV_IMPORT_PROFILE'
CANDIDATE = 'medicines-22-candidate'


class ProfileEntrypoints(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        shutil.copy2(ROOT / 'Makefile', self.root / 'Makefile')
        script_dir = self.root / 'scripts/old-db'
        script_dir.mkdir(parents=True)
        shutil.copy2(ROOT / 'scripts/old-db/stg-uat-old-db-handoff.sh', script_dir)
        self.script = script_dir / 'stg-uat-old-db-handoff.sh'
        self.bin = self.root / 'bin'
        self.bin.mkdir()
        self.log = self.root / 'calls.jsonl'
        (self.bin / 'docker').write_text('''#!/usr/bin/env python3
import json, os, sys
argv = sys.argv[1:]
sub = ''
for i, a in enumerate(argv):
    if a.endswith('cmd/csv-import-stg-uat') and i + 1 < len(argv):
        sub = argv[i + 1]
        break
ident = {k: os.environ.get(k, '') for k in (
    'STG_UAT_CSV_IMPORT_PROFILE', 'DB_HOST', 'DB_NAME', 'TARGET_DB_NAME',
    'STG_UAT_CSV_IMPORT_CONFIRM_HOST', 'CLINIC_CODE', 'TARGET_CLINIC_ID',
    'MIGRATION_RUN_ID', 'APP_ENV')}
with open(os.environ['MOCK_LOG'], 'a') as f:
    f.write(json.dumps({'subcommand': sub, 'argv': argv, 'env': ident}) + '\\n')
if sub == 'verify':
    sys.exit(int(os.environ.get('MOCK_VERIFY_EXIT', '0')))
if sub in ('apply', 'import'):
    sys.exit(int(os.environ.get('MOCK_IMPORT_EXIT', '0')))
''')
        (self.bin / 'docker').chmod(0o755)
        self.env = dict(os.environ, PATH=str(self.bin)+os.pathsep+os.environ['PATH'],
                        MOCK_LOG=str(self.log), DB_HOST='example.psdb.cloud',
                        DB_USER='mock-user', DB_PASSWORD='mock-value',
                        TARGET_DB_NAME='mock-db', STG_UAT_CSV_IMPORT_CONFIRM_HOST='example.psdb.cloud',
                        STG_UAT_CSV_IMPORT_ALLOW_REHEARSAL='YES_I_UNDERSTAND',
                        STG_UAT_HANDOFF_SKIP_LOCAL_ENV='1', APP_ENV='staging')
        self.env.pop(PROFILE, None)
        self.env.pop('MAKEFLAGS', None)
        self.env.pop('MFLAGS', None)
        self.env.pop('MAKEOVERRIDES', None)

    def run_make(self, target, args=(), env=None):
        return subprocess.run(['make', '-s', target, *args], cwd=self.root,
                              env=env or self.env, text=True, capture_output=True)

    def calls(self):
        if not self.log.exists():
            return []
        return [json.loads(line) for line in self.log.read_text().splitlines()]

    def subcommands(self):
        return [call['subcommand'] for call in self.calls()]

    def assert_profile(self, call, profile):
        self.assertEqual(call[call.index('--cutover-profile')+1], profile)
        self.assertNotIn('--allow-local-rehearsal', call)

    def fixtures(self):
        master = self.root / 'master'
        master.mkdir()
        for name, content in {
            'clinics.csv': 'id,is_active\n2,t\n',
            'animal_species.csv': 'id,is_active\n1,t\n',
            'exam_types.csv': 'id,clinic_id,name,deleted_at\n11009,2,\u691c\u67fb,\n',
            'reservation_types.csv': 'id,clinic_id,category,is_active,deleted_at\n2,2,trimming,t,\n',
            'payment_methods.csv': 'id,clinic_id,system_key,is_active,deleted_at\n5,2,cash,t,\n6,2,credit_card,t,\n',
        }.items():
            (master / name).write_text(content)
        handoff = self.root / 'handoff/jouto'
        handoff.mkdir(parents=True)
        (handoff / 'manifest.json').write_text(json.dumps({'clinicCode':'jouto','clinicOrdinal':2,'sourceRunId':'mock-run'}))
        self.env.update(STG_UAT_HANDOFF_ROOT=str(handoff.parent), STG_UAT_HANDOFF_MASTER_DIR=str(master),
                        STG_UAT_HANDOFF_REPORT_DIR=str(self.root / 'reports'))

    def test_direct_targets_profile_environment_and_command_line(self):
        for target in ['stg-uat-csv-import-preflight','stg-uat-csv-import','stg-uat-csv-import-verify','stg-uat-import']:
            for source in ['default','environment','command-line']:
                with self.subTest(target=target, source=source):
                    env = self.env.copy()
                    args = []
                    if source == 'environment': env[PROFILE] = CANDIDATE
                    if source == 'command-line': args = [PROFILE+'='+CANDIDATE]
                    result = self.run_make(target, args, env)
                    self.assertEqual(result.returncode, 0, result.stderr)
                    call = self.calls()[-1]['argv']
                    self.assert_profile(call, 'legacy-21' if source == 'default' else CANDIDATE)
                    if target in ['stg-uat-csv-import','stg-uat-import']:
                        report = call[call.index('--report-path')+1]
                        self.assertEqual(CANDIDATE in report, source != 'default')

    def test_profile_metacharacters_remain_literal(self):
        marker = self.root / 'unexpected-execution'
        payload = '`touch '+str(marker)+'`'
        for target in ['stg-uat-csv-import-preflight','stg-uat-import']:
            env = self.env.copy()
            env[PROFILE] = payload
            result = self.run_make(target, env=env)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertFalse(marker.exists(), 'profile was executed as shell source')
            self.assert_profile(self.calls()[-1]['argv'], payload)

    def test_handoff_profiles_survive_local_env(self):
        self.fixtures()
        local_env = self.root / 'mock.local.env'
        local_env.write_text(PROFILE+'=legacy-21\n')
        local_env.chmod(0o600)
        for target in ['stg-uat-handoff-preflight','stg-uat-handoff','stg-uat-handoff-verify']:
            for source in ['environment','command-line']:
                with self.subTest(target=target, source=source):
                    env = self.env.copy()
                    env.update(STG_UAT_HANDOFF_SKIP_LOCAL_ENV='0', STG_UAT_HANDOFF_ENV_FILE=str(local_env))
                    args = []
                    if source == 'environment': env[PROFILE] = CANDIDATE
                    else: args = [PROFILE+'='+CANDIDATE]
                    result = self.run_make(target, args, env)
                    self.assertEqual(result.returncode, 0, result.stderr)
                    self.assert_profile(self.calls()[-1]['argv'], CANDIDATE)

    def test_legacy_matching_pass_report_skips_without_probe(self):
        self.fixtures()
        reports = self.root / 'reports'
        reports.mkdir()
        (reports / 'jouto-mock-run-stg-uat-apply.json').write_text(
            json.dumps({'status': 'PASS'}))
        env = self.env.copy()
        env[PROFILE] = 'legacy-21'
        result = self.run_make('stg-uat-handoff', env=env)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn('apply report already PASS', result.stderr)
        # Legacy reuse is unchanged: no verify probe is added.
        self.assertEqual(self.subcommands(), [])

    def test_candidate_matching_pass_requires_target_verify_before_skip(self):
        self.fixtures()
        reports = self.root / 'reports'
        reports.mkdir()
        digest = hashlib.sha256((self.root / 'handoff/jouto/manifest.json').read_bytes()).hexdigest()
        (reports / ('jouto-mock-run-stg-uat-apply-'+CANDIDATE+'.json')).write_text(
            json.dumps({'status': 'PASS', 'manifestSha256': digest,
                        'targetHost': 'example.psdb.cloud', 'targetDatabase': 'postgres'}))
        env = self.env.copy()
        env[PROFILE] = CANDIDATE
        result = self.run_make('stg-uat-handoff', env=env)
        self.assertEqual(result.returncode, 0, result.stderr)
        # A PASS report plus a matching manifest SHA is not enough: the
        # current target must be re-verified before the report may be reused.
        self.assertEqual(self.subcommands(), ['verify'])
        self.assert_profile(self.calls()[0]['argv'], CANDIDATE)
        self.assertIn('skip', result.stderr)

    def test_candidate_matching_pass_verify_failure_stops_and_preserves_report(self):
        self.fixtures()
        reports = self.root / 'reports'
        reports.mkdir()
        digest = hashlib.sha256((self.root / 'handoff/jouto/manifest.json').read_bytes()).hexdigest()
        (reports / ('jouto-mock-run-stg-uat-apply-'+CANDIDATE+'.json')).write_text(
            json.dumps({'status': 'PASS', 'manifestSha256': digest,
                        'targetHost': 'example.psdb.cloud', 'targetDatabase': 'postgres'}))
        env = self.env.copy()
        env[PROFILE] = CANDIDATE
        env['MOCK_VERIFY_EXIT'] = '1'
        report = reports / ("jouto-mock-run-stg-uat-apply-"+CANDIDATE+".json")
        original = report.read_bytes()
        result = self.run_make('stg-uat-handoff', env=env)
        self.assertNotEqual(result.returncode, 0)
        # No success or write attempt from stale target evidence.
        self.assertIn("preserve/archive", result.stderr)
        self.assertEqual(self.subcommands(), ['verify'])
        self.assert_profile(self.calls()[0]['argv'], CANDIDATE)

        report = reports / ("jouto-mock-run-stg-uat-apply-"+CANDIDATE+".json")
        self.assertEqual(report.read_bytes(), original)

    def test_candidate_probe_error_never_retries_import(self):
        self.fixtures()
        reports = self.root / 'reports'
        reports.mkdir()
        digest = hashlib.sha256((self.root / 'handoff/jouto/manifest.json').read_bytes()).hexdigest()
        (reports / ('jouto-mock-run-stg-uat-apply-'+CANDIDATE+'.json')).write_text(
            json.dumps({'status': 'PASS', 'manifestSha256': digest,
                        'targetHost': 'example.psdb.cloud', 'targetDatabase': 'postgres'}))
        env = self.env.copy()
        env[PROFILE] = CANDIDATE
        env['MOCK_VERIFY_EXIT'] = '17'
        env['MOCK_IMPORT_EXIT'] = '0'
        result = self.run_make('stg-uat-handoff', env=env)
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(self.subcommands(), ['verify'])

    def test_selected_db_identity_forwarded_to_verify(self):
        self.fixtures()
        reports = self.root / 'reports'
        reports.mkdir()
        digest = hashlib.sha256((self.root / 'handoff/jouto/manifest.json').read_bytes()).hexdigest()
        (reports / ('jouto-mock-run-stg-uat-apply-'+CANDIDATE+'.json')).write_text(
            json.dumps({'status': 'PASS', 'manifestSha256': digest,
                        'targetHost': 'rebuilt.psdb.cloud', 'targetDatabase': 'rebuilt-db'}))
        local_env = self.root / 'mock.local.env'
        local_env.write_text('DB_HOST=rebuilt.psdb.cloud' + chr(10) + 'DB_USER=mock-user' + chr(10) + 'DB_PASSWORD=mock-value' + chr(10) + 'TARGET_DB_NAME=rebuilt-db' + chr(10) + 'STG_UAT_CSV_IMPORT_CONFIRM_HOST=rebuilt.psdb.cloud' + chr(10))
        local_env.chmod(0o600)
        env = self.env.copy()
        env.update({PROFILE: CANDIDATE, 'MOCK_VERIFY_EXIT': '1',
                    'STG_UAT_HANDOFF_SKIP_LOCAL_ENV': '0',
                    'STG_UAT_HANDOFF_ENV_FILE': str(local_env)})
        result = self.run_make('stg-uat-handoff', env=env)
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(self.subcommands(), ['verify'])
        for call in self.calls():
            self.assertEqual(call['env']['DB_HOST'], 'rebuilt.psdb.cloud')
            self.assertEqual(call['argv'][call['argv'].index('--confirm-target-host')+1], 'rebuilt.psdb.cloud')
            self.assertEqual(call['argv'][call['argv'].index('--confirm-target-database')+1], 'rebuilt-db')
            self.assertEqual(call['env']['TARGET_DB_NAME'], 'rebuilt-db')
            self.assertEqual(call['env']['DB_NAME'], 'rebuilt-db')
            self.assertEqual(call['env']['CLINIC_CODE'], 'jouto')
            self.assertEqual(call['env']['TARGET_CLINIC_ID'], '2')

    def test_candidate_report_target_binding_must_match(self):
        self.fixtures()
        reports = self.root / 'reports'
        reports.mkdir()
        digest = hashlib.sha256((self.root / 'handoff/jouto/manifest.json').read_bytes()).hexdigest()
        report = reports / ('jouto-mock-run-stg-uat-apply-'+CANDIDATE+'.json')
        for binding in [
            {'targetHost':'different.psdb.cloud','targetDatabase':'postgres'},
            {'targetHost':'example.psdb.cloud','targetDatabase':'different-db'},
            {},
        ]:
            with self.subTest(binding=binding):
                original = json.dumps({'status':'PASS','manifestSha256':digest, **binding})
                report.write_text(original)
                env = self.env.copy()
                env[PROFILE] = CANDIDATE
                result = self.run_make('stg-uat-handoff', env=env)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn('target binding', result.stderr)
                self.assertEqual(self.subcommands(), [])
                self.assertEqual(report.read_text(), original)

    def test_candidate_stale_report_is_not_reused(self):
        self.fixtures()
        reports = self.root / 'reports'
        reports.mkdir()
        report = reports / 'jouto-mock-run-stg-uat-apply-medicines-22-candidate.json'
        report.write_text(json.dumps({'status':'PASS','manifestSha256':'stale'}))
        env = self.env.copy()
        env[PROFILE] = CANDIDATE
        result = self.run_make('stg-uat-handoff', env=env)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertNotIn('apply report already PASS', result.stderr)
        self.assert_profile(self.calls()[-1]['argv'], CANDIDATE)

    def test_unknown_handoff_profile_rejected_before_local_env(self):
        env = self.env.copy()
        env.update({PROFILE:'unknown', 'STG_UAT_HANDOFF_SKIP_LOCAL_ENV':'0',
                    'STG_UAT_HANDOFF_ENV_FILE':str(self.root / 'missing.env')})
        result = subprocess.run(['bash',str(self.script),'preflight'], env=env, text=True, capture_output=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('unknown STG_UAT_CSV_IMPORT_PROFILE', result.stderr)
        self.assertFalse(self.log.exists())

    def test_candidate_never_reuses_legacy_pass_report(self):
        self.fixtures()
        reports = self.root / 'reports'
        reports.mkdir()
        (reports/'jouto-mock-run-stg-uat-apply.json').write_text('{"status":"PASS"}')
        env = self.env.copy()
        env[PROFILE] = CANDIDATE
        result = self.run_make('stg-uat-handoff', env=env)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertNotIn('apply report already PASS', result.stderr)
        self.assert_profile(self.calls()[-1]['argv'], CANDIDATE)


if __name__ == '__main__':
    unittest.main(verbosity=2)
