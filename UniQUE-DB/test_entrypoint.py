"""Exercise migration startup ordering without a running database."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest


class EntrypointTest(unittest.TestCase):
    def run_entrypoint(self, action, responses):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "21_settings.up.sql").touch()
            (root / "responses").write_text("\n".join(responses) + "\n")
            migrate = root / "migrate"
            migrate.write_text('''#!/bin/sh
printf '%s\\n' "$*" >> "$TEST_ROOT/calls"
response=$(head -1 "$TEST_ROOT/responses")
sed '1d' "$TEST_ROOT/responses" > "$TEST_ROOT/remaining"
mv "$TEST_ROOT/remaining" "$TEST_ROOT/responses"
if [ "$response" = error ]; then exit 1; fi
printf '%s\\n' "$response" >&2
''')
            migrate.chmod(0o755)
            result = subprocess.run(
                ["sh", str(Path(__file__).with_name("entrypoint.sh")), action],
                env={**os.environ, "PATH": f"{root}:{os.environ['PATH']}",
                     "TEST_ROOT": str(root), "MIGRATIONS_DIR": str(root),
                     "DATABASE_URL": "mysql://synthetic", "MAX_RETRIES": "3", "SLEEP": "0"},
                capture_output=True, text=True, timeout=5,
            )
            calls = (root / "calls").read_text().splitlines()
            return result.returncode, calls

    def test_up_retries_until_database_is_available(self):
        code, calls = self.run_entrypoint("up", ["error", "error", "ok"])
        self.assertEqual(code, 0)
        self.assertEqual(len(calls), 3)
        self.assertTrue(all(call.endswith(" up") for call in calls))

    def test_up_failure_stays_failed(self):
        code, calls = self.run_entrypoint("up", ["error"] * 3)
        self.assertEqual(code, 1)
        self.assertEqual(len(calls), 3)

    def test_wait_blocks_until_clean_target_version(self):
        code, calls = self.run_entrypoint("wait", ["error", "20", "21"])
        self.assertEqual(code, 0)
        self.assertEqual(len(calls), 3)
        self.assertTrue(all(call.endswith(" version") for call in calls))

    def test_wait_rejects_dirty_schema(self):
        code, calls = self.run_entrypoint("wait", ["21 (dirty)"] * 3)
        self.assertEqual(code, 1)
        self.assertEqual(len(calls), 3)

    def test_wait_rejects_older_schema(self):
        code, _ = self.run_entrypoint("wait", ["20"] * 3)
        self.assertEqual(code, 1)

    def test_wait_allows_newer_clean_schema(self):
        code, calls = self.run_entrypoint("wait", ["22"])
        self.assertEqual(code, 0)
        self.assertEqual(len(calls), 1)


if __name__ == "__main__":
    unittest.main()
