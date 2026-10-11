#!/usr/bin/env python3
"""Safety policy propagation, drift and foreign-work protection contracts."""
import importlib.util
import pathlib
import tempfile
import unittest

SCRIPT = pathlib.Path(__file__).with_name('sync-agent-trust-boundaries.py')
spec = importlib.util.spec_from_file_location('trust', SCRIPT)
trust = importlib.util.module_from_spec(spec)
spec.loader.exec_module(trust)


class TrustBoundaryTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = pathlib.Path(self.tmp.name)
        source = self.root / trust.SOURCE
        source.parent.mkdir(parents=True)
        source.write_text('Canonical constraints\n')
        self.target = self.root / 'CLAUDE.md'
        self.original = b'---\ntools: [Read, Bash]\n---\n\nOwner text\r\n'
        self.target.write_bytes(self.original)

    def test_preserves_original_bytes_frontmatter_and_is_idempotent(self):
        trust.sync(self.root, ['CLAUDE.md'])
        first = self.target.read_bytes()
        self.assertTrue(first.startswith(self.original))
        self.assertEqual(trust.sync(self.root, ['CLAUDE.md'], check=True), 0)
        trust.sync(self.root, ['CLAUDE.md'])
        self.assertEqual(first, self.target.read_bytes())

    def test_readonly_check_detects_missing_and_stale_policy(self):
        self.assertEqual(trust.sync(self.root, ['CLAUDE.md'], check=True), 1)
        self.assertEqual(self.target.read_bytes(), self.original)
        trust.sync(self.root, ['CLAUDE.md'])
        (self.root / trust.SOURCE).write_text('Updated constraints\n')
        self.assertEqual(trust.sync(self.root, ['CLAUDE.md'], check=True), 1)
        trust.sync(self.root, ['CLAUDE.md'])
        self.assertIn('Updated constraints', self.target.read_text())

    def test_preflight_preserves_all_files_when_owner_edits_block(self):
        trust.sync(self.root, ['CLAUDE.md'])
        self.target.write_text(self.target.read_text().replace('Canonical constraints', 'Foreign WIP'))
        other = self.root / 'AGENTS.md'
        other.write_text('Other owner text\n')
        before = self.target.read_bytes(), other.read_bytes()
        with self.assertRaises(ValueError):
            trust.sync(self.root, ['AGENTS.md', 'CLAUDE.md'])
        self.assertEqual(before, (self.target.read_bytes(), other.read_bytes()))

    def test_refuses_malformed_and_duplicate_blocks(self):
        valid = trust.render('Owner\n', 'Policy\n')
        for content in ('Owner\n' + trust.START, valid + valid, valid.replace(trust.END, '')):
            with self.subTest(content=content), self.assertRaises(ValueError):
                trust.render(content, 'Policy\n')

    def test_refuses_symlink_and_outside_root_targets(self):
        link = self.root / 'link.md'
        link.symlink_to(self.target)
        for path in ('link.md', '../outside.md'):
            with self.subTest(path=path), self.assertRaises(ValueError):
                trust.sync(self.root, [path])

    def test_repository_blocks_are_current_and_cover_instruction_surfaces(self):
        root = SCRIPT.parents[2]
        paths = trust.targets(root)
        self.assertIn('AGENTS.md', paths)
        self.assertIn('backend/CLAUDE.md', paths)
        self.assertIn('.claude/agents/implementer.md', paths)
        self.assertIn('.claude/rules/worktree-runtime.md', paths)
        self.assertNotIn(trust.SOURCE, paths)
        self.assertEqual(trust.sync(root, paths, check=True), 0)


if __name__ == '__main__':
    unittest.main()
