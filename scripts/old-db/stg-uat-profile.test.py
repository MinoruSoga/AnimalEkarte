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
import json,os,sys
with open(os.environ['MOCK_LOG'],'a') as f: f.write(json.dumps(sys.argv[1:])+'\\n')
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
        return [json.loads(line) for line in self.log.read_text().splitlines()]

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
                    call = self.calls()[-1]
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
            self.assert_profile(self.calls()[-1], payload)

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
                    self.assert_profile(self.calls()[-1], CANDIDATE)

    def test_matching_profile_reports_remain_reusable(self):
        self.fixtures()
        reports = self.root / 'reports'
        reports.mkdir()
        digest = hashlib.sha256((self.root / 'handoff/jouto/manifest.json').read_bytes()).hexdigest()
        for profile, suffix in [('legacy-21',''), (CANDIDATE,'-medicines-22-candidate')]:
            with self.subTest(profile=profile):
                (reports / ('jouto-mock-run-stg-uat-apply'+suffix+'.json')).write_text(
                    json.dumps({'status':'PASS','manifestSha256':digest}))
                env = self.env.copy()
                env[PROFILE] = profile
                result = self.run_make('stg-uat-handoff', env=env)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertIn('apply report already PASS', result.stderr)
                self.assertFalse(self.log.exists())

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
        self.assert_profile(self.calls()[-1], CANDIDATE)

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
        self.assert_profile(self.calls()[-1], CANDIDATE)


if __name__ == '__main__':
    unittest.main(verbosity=2)
