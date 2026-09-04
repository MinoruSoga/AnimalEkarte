#!/usr/bin/env python3
from pathlib import Path
import tempfile
import unittest

from merge_go_coverprofiles import merge_profiles, read_profile, write_profile


class MergeGoCoverprofilesTest(unittest.TestCase):
    def write(self, directory: Path, name: str, body: str) -> Path:
        path = directory / name
        path.write_text(body, encoding="utf-8")
        return path

    def test_merges_overlapping_atomic_blocks_and_keeps_union(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            first = self.write(
                root,
                "first.out",
                "mode: atomic\nexample/a.go:1.1,2.2 2 1\nexample/b.go:3.1,4.2 1 0\n",
            )
            second = self.write(
                root,
                "second.out",
                "mode: atomic\nexample/a.go:1.1,2.2 2 3\nexample/c.go:5.1,6.2 4 1\n",
            )

            mode, blocks = merge_profiles([first, second])

            self.assertEqual("atomic", mode)
            self.assertEqual(4, blocks["example/a.go:1.1,2.2"].count)
            self.assertEqual(0, blocks["example/b.go:3.1,4.2"].count)
            self.assertEqual(1, blocks["example/c.go:5.1,6.2"].count)

    def test_set_mode_uses_max_instead_of_sum(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            first = self.write(root, "first.out", "mode: set\na.go:1.1,1.2 1 1\n")
            second = self.write(root, "second.out", "mode: set\na.go:1.1,1.2 1 1\n")

            _, blocks = merge_profiles([first, second])

            self.assertEqual(1, blocks["a.go:1.1,1.2"].count)

    def test_rejects_mode_mismatch(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            first = self.write(root, "first.out", "mode: atomic\n")
            second = self.write(root, "second.out", "mode: count\n")

            with self.assertRaisesRegex(ValueError, "coverage mode mismatch"):
                merge_profiles([first, second])

    def test_round_trip_is_deterministic(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            source = self.write(
                root,
                "source.out",
                "mode: atomic\nz.go:1.1,1.2 1 2\na.go:1.1,1.2 1 1\n",
            )
            mode, blocks = read_profile(source)
            output = root / "merged.out"

            write_profile(output, mode, blocks)

            self.assertEqual(
                "mode: atomic\na.go:1.1,1.2 1 1\nz.go:1.1,1.2 1 2\n",
                output.read_text(encoding="utf-8"),
            )


if __name__ == "__main__":
    unittest.main()
