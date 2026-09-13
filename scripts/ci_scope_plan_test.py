#!/usr/bin/env python3
"""Offline unit tests for ci_scope_plan."""
import importlib.util
import pathlib
import unittest

SPEC = importlib.util.spec_from_file_location(
    'ci_scope_plan', pathlib.Path(__file__).with_name('ci_scope_plan.py')
)
plan_mod = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(plan_mod)


class CiScopePlanTests(unittest.TestCase):
    def test_reservation_only_is_partial_backend(self):
        plan = plan_mod.plan_scope(['backend/internal/reservation/reservation_service.go'])
        self.assertEqual(plan['mode'], 'partial')
        self.assertEqual(plan['backend_domains'], ['reservation'])
        self.assertEqual(plan['coverage_ratchet'], 'skip')
        self.assertTrue(plan['backend'])
        self.assertFalse(plan['frontend'])
        self.assertEqual(len(plan['backend_matrix']), 1)
        self.assertEqual(plan['backend_matrix'][0]['packages'], './internal/reservation')
        self.assertTrue(plan['run_backend_tests'])
        self.assertEqual(plan['backend_shard_count'], 1)

    def test_ui_dialog_forces_full_frontend(self):
        plan = plan_mod.plan_scope(['frontend/src/components/ui/dialog.tsx'])
        self.assertEqual(plan['mode'], 'full')
        self.assertEqual(plan['coverage_ratchet'], 'run')
        self.assertTrue(plan['frontend'])
        self.assertEqual(len(plan['frontend_matrix']), 2)

    def test_model_package_forces_full_backend(self):
        plan = plan_mod.plan_scope(['backend/internal/model/reservation.go'])
        self.assertEqual(plan['mode'], 'full')
        self.assertEqual(plan['coverage_ratchet'], 'run')
        self.assertTrue(plan['backend'])

    def test_two_features_partial(self):
        plan = plan_mod.plan_scope([
            'frontend/src/features/reservations/api/transforms.ts',
            'frontend/src/features/owners/components/OwnerCard.tsx',
        ])
        self.assertEqual(plan['mode'], 'partial')
        self.assertEqual(plan['frontend_features'], ['owners', 'reservations'])
        self.assertEqual(plan['coverage_ratchet'], 'skip')
        self.assertEqual(
            [row['feature_path'] for row in plan['frontend_matrix']],
            ['src/features/owners', 'src/features/reservations'],
        )
        self.assertEqual(plan['frontend_shard_count'], 2)

    def test_migration_forces_full(self):
        plan = plan_mod.plan_scope(['backend/migrations/002_medical_records_entered_by_staff_fk.sql'])
        self.assertEqual(plan['mode'], 'full')
        self.assertEqual(plan['coverage_ratchet'], 'run')

    def test_httpapi_is_shared_not_domain(self):
        plan = plan_mod.plan_scope(['backend/internal/httpapi/response.go'])
        self.assertEqual(plan['mode'], 'full')

    def test_docs_only_skips(self):
        plan = plan_mod.plan_scope(['docs/ops/ci-policy.md'])
        self.assertEqual(plan['mode'], 'skip')
        self.assertEqual(plan['coverage_ratchet'], 'skip')
        self.assertEqual(plan['backend_matrix'], [])
        self.assertEqual(plan['frontend_matrix'], [])

    def test_domain_plus_shared_is_full(self):
        plan = plan_mod.plan_scope([
            'backend/internal/reservation/x.go',
            'backend/internal/sharedkernel/y.go',
        ])
        self.assertEqual(plan['mode'], 'full')
        self.assertIn('reservation', plan['backend_domains'])


if __name__ == '__main__':
    unittest.main()
