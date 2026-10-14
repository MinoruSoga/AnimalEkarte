"""Loopback-only contract tests for the Cloudflare demo smoke harness."""

import contextlib
import http.server
import importlib.util
import io
import json
import os
import pathlib
import sys
import threading
import unittest
from typing import Optional
from unittest import mock


SCRIPT = pathlib.Path(__file__).with_name("cf-demo-smoke.py")
SPEC = importlib.util.spec_from_file_location("cf_demo_smoke", SCRIPT)
SMOKE = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(SMOKE)


class FixtureState:
    def __init__(self) -> None:
        self.mode = "ok"
        self.trace: list[tuple[str, str, dict[str, str], bytes]] = []
        self.records: dict[tuple[str, str], dict[str, object]] = {}
        self.next_ids = {"permission-groups": "701", "staffs": "702"}
        self.delete_status = 204
        self.group_detail_reads = 0
        self.create_mode = "ok"
        self.stdout_at_first_post = ""


class Handler(http.server.BaseHTTPRequestHandler):
    state: FixtureState

    def log_message(self, format: str, *args: object) -> None:
        return

    def _reply(self, status: int, body: object, cookie: bool = False) -> None:
        raw = json.dumps(body).encode("utf-8")
        self.send_response(status)
        if cookie:
            self.send_header(
                "Set-Cookie", "access_token=dummy-cookie; Path=/; HttpOnly"
            )
            self.send_header(
                "Set-Cookie",
                "refresh_token=dummy-refresh; Path=/api/v1/auth/refresh; HttpOnly",
            )
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)

    def _handle(self) -> None:
        length = int(self.headers.get("Content-Length", "0"))
        raw = self.rfile.read(length) if length else b""
        path = self.path
        method = self.command
        headers = {key.lower(): value for key, value in self.headers.items()}
        self.state.trace.append((method, path, headers, raw))
        payload = json.loads(raw) if raw else {}

        if path == "/health":
            self._reply(200, {"status": "ok"})
        elif path == "/api/v1/login":
            if self.state.mode == "unauthorized":
                self._reply(401, {"message": "unauthorized"})
            elif self.state.mode == "login_201_cookies":
                self._reply(201, {"user": {"id": 21}}, cookie=True)
            elif self.state.mode == "login_500_cookies":
                self._reply(500, {"message": "login failed"}, cookie=True)
            elif self.state.mode == "malformed_login":
                self.send_response(200)
                self.send_header(
                    "Set-Cookie", "access_token=dummy-cookie; Path=/; HttpOnly"
                )
                self.send_header(
                    "Set-Cookie",
                    "refresh_token=dummy-refresh; Path=/api/v1/auth/refresh; HttpOnly",
                )
                self.send_header("Content-Length", "1")
                self.end_headers()
                self.wfile.write(b"{")
            elif self.state.mode == "read_failure":
                self.send_response(200)
                self.send_header(
                    "Set-Cookie", "access_token=dummy-cookie; Path=/; HttpOnly"
                )
                self.send_header(
                    "Set-Cookie",
                    "refresh_token=dummy-refresh; Path=/api/v1/auth/refresh; HttpOnly",
                )
                self.send_header("Content-Length", "128")
                self.end_headers()
                self.wfile.write(b'{"user":')
                self.close_connection = True
            else:
                self._reply(
                    200,
                    {"user": {"id": 21}},
                    cookie=self.state.mode != "missing_cookie",
                )
        elif path == "/api/v1/me":
            identity = 22 if self.state.mode == "me_mismatch" else 21
            self._reply(200, {"id": identity})
        elif path == "/api/v1/auth/refresh/logout":
            if self.state.mode == "logout_redirect":
                self.send_response(307)
                self.send_header("Location", "/api/v1/logout")
                self.send_header("Content-Length", "0")
                self.end_headers()
            else:
                self._reply(200, {"message": "logged out"})
        elif path == "/api/v1/logout":
            self._reply(200, {"message": "legacy must not be called"})
        elif path.startswith("/api/v1/masters/"):
            self._master(method, path, payload)
        else:
            self._reply(404, {})

    def _master(self, method: str, path: str, payload: dict[str, object]) -> None:
        pieces = path.split("/")
        kind = pieces[4]
        identity = pieces[5] if len(pieces) > 5 else ""
        if method == "POST":
            if not self.state.stdout_at_first_post:
                self.state.stdout_at_first_post = sys.stdout.getvalue()
            assigned = self.state.next_ids[kind]
            self.state.records[(kind, assigned)] = {**payload, "id": int(assigned)}
            if self.state.create_mode == "malformed_body":
                self.send_response(201)
                self.send_header("Content-Length", "1")
                self.end_headers()
                self.wfile.write(b"{")
            elif self.state.create_mode == "name_mismatch":
                self._reply(
                    201, {**self.state.records[(kind, assigned)], "name": "wrong"}
                )
            elif self.state.create_mode == "invalid_id":
                self._reply(201, {**self.state.records[(kind, assigned)], "id": "bad"})
            elif self.state.create_mode == "server_error":
                self._reply(500, {})
            elif self.state.create_mode == "read_failure":
                self.send_response(201)
                self.send_header("Content-Length", "128")
                self.end_headers()
                self.wfile.write(b'{"id":')
                self.close_connection = True
            elif self.state.create_mode == "known_forbidden":
                self.state.records.pop((kind, assigned), None)
                self._reply(403, {})
            else:
                self._reply(201, self.state.records[(kind, assigned)])
        elif method == "PATCH":
            record = self.state.records[(kind, identity)]
            record.update(payload)
            if kind == "permission-groups" and self.state.mode == "update_fail":
                self._reply(500, {})
            else:
                self._reply(200, record)
        elif method == "GET":
            if (
                kind == "permission-groups"
                and self.state.mode == "group_disappears_before_cleanup"
            ):
                self.state.group_detail_reads += 1
                if self.state.group_detail_reads == 3:
                    self.state.records.pop((kind, identity), None)
            record = self.state.records.get((kind, identity))
            if record is None:
                self._reply(404, {})
            elif kind == "permission-groups" and self.state.mode == "read_fail_once":
                self.state.group_detail_reads += 1
                if self.state.group_detail_reads == 2:
                    self._reply(500, {})
                else:
                    self._reply(200, record)
            elif (
                kind == "permission-groups" and self.state.mode == "ownership_mismatch"
            ):
                self._reply(200, {**record, "name": "somebody-elses-row"})
            else:
                self._reply(200, record)
        elif method == "DELETE":
            if self.state.mode == "delete_fail":
                self._reply(500, {})
            else:
                self.state.records.pop((kind, identity), None)
                self._reply(self.state.delete_status, {})
        else:
            self._reply(405, {})

    do_GET = _handle
    do_POST = _handle
    do_PATCH = _handle
    do_DELETE = _handle


class DemoSmokeTests(unittest.TestCase):
    def setUp(self) -> None:
        self.state = FixtureState()
        handler = type("BoundHandler", (Handler,), {"state": self.state})
        self.server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), handler)
        self.server.daemon_threads = True
        self.thread = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.thread.start()
        self.base = "http://127.0.0.1:%d" % self.server.server_address[1]
        self.env = {
            "WORKER_URL": self.base,
            "SMOKE_ALLOW_LOCAL": "1",
            "SEEDLOGIN_DEMO_PASSWORD": "  dummy secret with spaces  ",
        }

    def tearDown(self) -> None:
        self.server.shutdown()
        self.server.server_close()
        self.thread.join(timeout=2)

    def run_main(
        self, mode: str = "--auth", overrides: Optional[dict[str, str]] = None
    ) -> tuple[int, str, str]:
        stdout = io.StringIO()
        stderr = io.StringIO()
        environment = {**self.env, **(overrides or {})}
        with (
            mock.patch.dict(os.environ, environment, clear=True),
            mock.patch.object(SMOKE.sys, "argv", [str(SCRIPT), mode]),
            contextlib.redirect_stdout(stdout),
            contextlib.redirect_stderr(stderr),
        ):
            code = SMOKE.main()
        return code, stdout.getvalue(), stderr.getvalue()

    def test_auth_trace_uses_only_health_login_me_logout(self) -> None:
        code, output, _ = self.run_main()
        self.assertEqual(code, 0)
        self.assertEqual(
            [(method, path) for method, path, _, _ in self.state.trace],
            [
                ("GET", "/health"),
                ("POST", "/api/v1/login"),
                ("GET", "/api/v1/me"),
                ("POST", "/api/v1/auth/refresh/logout"),
            ],
        )
        self.assertIn('"accounting_smoke": "NOT_RUN"', output)
        self.assertNotIn("dummy secret", output)
        self.assertNotIn("dummy-cookie", output)
        self.assertNotIn("stg-staff-10000021@example.test", output)
        logout = self.state.trace[-1]
        self.assertIn("refresh_token=dummy-refresh", logout[2].get("cookie", ""))

    def test_invalid_targets_modes_and_missing_password_fail_before_requests(
        self,
    ) -> None:
        invalid_targets = (
            "https://example.invalid",
            "https://user:pass@animalekarte-stg-api.baritech-soga.workers.dev",
            "https://animalekarte-stg-api.baritech-soga.workers.dev/path",
            "https://animalekarte-stg-api.baritech-soga.workers.dev/?q=1",
            "https://animalekarte-stg-api.baritech-soga.workers.dev#fragment",
            "https://animalekarte-stg-api.baritech-soga.workers.dev:444",
        )
        for target in invalid_targets:
            with self.subTest(target_class="invalid origin or URL component"):
                code, output, _ = self.run_main(
                    overrides={"WORKER_URL": target, "SMOKE_ALLOW_LOCAL": ""}
                )
                self.assertEqual(code, 2)
                self.assertNotIn("PASS", output)
                self.assertEqual(self.state.trace, [])
        code, _, _ = self.run_main(
            overrides={"SEEDLOGIN_DEMO_PASSWORD": "", "STG_DEMO_PASSWORD": "legacy"}
        )
        self.assertEqual(code, 2)
        code, _, _ = self.run_main(mode="--unsupported")
        self.assertEqual(code, 2)
        self.assertEqual(self.state.trace, [])

    def test_production_host_rejected_even_with_explicit_local_flag(self) -> None:
        code, _, _ = self.run_main(
            overrides={
                "WORKER_URL": "https://api.noah-karte.com",
                "SMOKE_ALLOW_LOCAL": "1",
            }
        )
        self.assertEqual(code, 2)
        self.assertEqual(self.state.trace, [])

    def test_crud_write_flag_and_numeric_clinic_are_required_before_requests(
        self,
    ) -> None:
        code, _, _ = self.run_main("--crud")
        self.assertEqual(code, 2)
        code, _, _ = self.run_main(
            "--crud",
            {
                "SMOKE_ALLOW_FIXTURE_WRITES": "1",
                "SMOKE_FIXTURE_CLINIC_ID": "0",
            },
        )
        self.assertEqual(code, 2)
        code, _, _ = self.run_main(
            "--crud",
            {
                "SMOKE_ALLOW_FIXTURE_WRITES": "1",
                "SMOKE_FIXTURE_CLINIC_ID": "1;delete",
            },
        )
        self.assertEqual(code, 2)
        self.assertEqual(self.state.trace, [])

    def test_fixed_email_and_password_bytes_are_used_without_fallback(self) -> None:
        self.env["STG_DEMO_EMAIL"] = "untrusted@example.invalid"
        self.env["STG_DEMO_PASSWORD"] = "wrong legacy password"
        code, output, _ = self.run_main()
        login = next(entry for entry in self.state.trace if entry[1] == "/api/v1/login")
        payload = json.loads(login[3])
        self.assertEqual(code, 0)
        self.assertEqual(payload["email"], "stg-staff-10000021@example.test")
        self.assertEqual(payload["password"], "  dummy secret with spaces  ")
        self.assertIn(b'"password": "  dummy secret with spaces  "', login[3])
        self.assertNotIn("dummy secret", output)
        self.assertNotIn("wrong legacy password", output)

    def test_credentials_and_cookies_use_only_the_memory_cookie_jar(self) -> None:
        source = SCRIPT.read_text(encoding="utf-8")
        smoke = self.make_smoke()
        self.assertIsInstance(smoke.jar, SMOKE.http.cookiejar.CookieJar)
        self.assertIn("http.cookiejar.CookieJar()", source)
        self.assertNotIn("MozillaCookieJar", source)
        self.assertNotIn("NamedTemporaryFile", source)
        self.assertNotIn("open(", source.replace("self.client.open(", ""))

    def test_unauthorized_login_fails_without_retry_or_crud(self) -> None:
        self.state.mode = "unauthorized"
        code, output, _ = self.run_main()
        self.assertEqual(code, 1)
        self.assertEqual(
            [path for _, path, _, _ in self.state.trace], ["/health", "/api/v1/login"]
        )
        self.assertIn('"auth": "FAIL"', output)

    def test_unexpected_login_status_with_auth_cookies_still_logs_out(self) -> None:
        for mode in ("login_201_cookies", "login_500_cookies"):
            with self.subTest(mode=mode):
                self.state = FixtureState()
                self.state.mode = mode
                self.server.RequestHandlerClass.state = self.state
                smoke = self.make_smoke()
                code, output = self.run_with_smoke(smoke)
                self.assertEqual(code, 1)
                self.assertIn('"auth": "FAIL"', output)
                self.assertEqual(
                    [path for _, path, _, _ in self.state.trace],
                    [
                        "/health",
                        "/api/v1/login",
                        "/api/v1/auth/refresh/logout",
                    ],
                )
                self.assertIn(
                    "refresh_token=dummy-refresh",
                    self.state.trace[-1][2].get("cookie", ""),
                )
                self.assertFalse(list(smoke.jar))

    def test_malformed_login_body_fails_but_logout_runs_and_cookie_is_cleared(
        self,
    ) -> None:
        self.state.mode = "malformed_login"
        smoke = self.make_smoke()
        code, output = self.run_with_smoke(smoke)
        self.assertEqual(code, 1)
        self.assertIn('"auth": "FAIL"', output)
        self.assertTrue(smoke.session_open)
        self.assertTrue(smoke.logout_verified)
        self.assertFalse(list(smoke.jar))
        self.assertEqual(self.state.trace[-1][1], "/api/v1/auth/refresh/logout")

    def test_login_body_read_failure_fails_but_logout_is_attempted(self) -> None:
        self.state.mode = "read_failure"
        smoke = self.make_smoke()
        code, output = self.run_with_smoke(smoke)
        self.assertEqual(code, 1)
        self.assertIn('"auth": "FAIL"', output)
        self.assertTrue(smoke.session_open)
        self.assertEqual(self.state.trace[-1][1], "/api/v1/auth/refresh/logout")
        self.assertFalse(list(smoke.jar))

    def test_missing_cookie_and_me_mismatch_fail_and_logout(self) -> None:
        for mode in ("missing_cookie", "me_mismatch"):
            with self.subTest(mode=mode):
                self.state = FixtureState()
                self.state.mode = mode
                self.server.RequestHandlerClass.state = self.state
                smoke = self.make_smoke()
                code, output = self.run_with_smoke(smoke)
                self.assertEqual(code, 1)
                self.assertIn('"auth": "FAIL"', output)
                self.assertEqual(self.state.trace[-1][1], "/api/v1/auth/refresh/logout")

    def test_legacy_redirect_is_not_followed_and_logout_fails(self) -> None:
        self.state.mode = "logout_redirect"
        code, output, _ = self.run_main()
        self.assertEqual(code, 1)
        self.assertEqual(self.state.trace[-1][1], "/api/v1/auth/refresh/logout")
        self.assertNotIn("/api/v1/logout", [path for _, path, _, _ in self.state.trace])
        self.assertIn('"auth": "FAIL"', output)

    def test_crud_scopes_every_master_call_and_deletes_only_returned_ids(self) -> None:
        self.env.update(
            {"SMOKE_ALLOW_FIXTURE_WRITES": "1", "SMOKE_FIXTURE_CLINIC_ID": "1"}
        )
        code, output, _ = self.run_main("--crud")
        self.assertEqual(code, 0)
        master = [entry for entry in self.state.trace if "/masters/" in entry[1]]
        self.assertTrue(master)
        self.assertTrue(all(entry[2].get("x-clinic-id") == "1" for entry in master))
        created = [entry for entry in master if entry[0] == "POST"]
        self.assertEqual(len(created), 2)
        self.assertEqual(
            [entry[1] for entry in created],
            [
                "/api/v1/masters/permission-groups",
                "/api/v1/masters/staffs",
            ],
        )
        group_body = json.loads(created[0][3])
        staff_body = json.loads(created[1][3])
        self.assertEqual(
            set(group_body), {"name", "description", "color", "is_active", "sort_order"}
        )
        self.assertEqual(
            {key: value for key, value in group_body.items() if key != "name"},
            {
                "description": "Synthetic smoke fixture",
                "color": "#6366F1",
                "is_active": True,
                "sort_order": 99,
            },
        )
        self.assertEqual(set(staff_body), {"name", "sort_order"})
        self.assertEqual(staff_body["sort_order"], 99)
        self.assertTrue(group_body["name"].startswith("ci-smoke-"))
        self.assertTrue(staff_body["name"].startswith("ci-smoke-"))
        patches = [entry for entry in master if entry[0] == "PATCH"]
        self.assertEqual(len(patches), 2)
        self.assertEqual(
            json.loads(patches[0][3]),
            {
                "description": "Updated synthetic smoke fixture",
                "name": group_body["name"],
            },
        )
        self.assertEqual(
            json.loads(patches[1][3]),
            {"sort_order": 100, "name": staff_body["name"]},
        )
        deletes = [entry[1] for entry in master if entry[0] == "DELETE"]
        self.assertEqual(
            set(deletes),
            {
                "/api/v1/masters/permission-groups/701",
                "/api/v1/masters/staffs/702",
            },
        )
        for path in deletes:
            delete_index = next(
                index
                for index, entry in enumerate(master)
                if entry[0] == "DELETE" and entry[1] == path
            )
            self.assertEqual(master[delete_index + 1][0:2], ("GET", path))
        forbidden = ("/pets", "/owners", "/accountings", "?search=", "?limit=")
        self.assertFalse(
            any(
                any(part in path for part in forbidden)
                for _, path, _, _ in self.state.trace
            )
        )
        self.assertIn('"master_crud": "PASS"', output)
        self.assertIn('"accounting_smoke": "NOT_RUN"', output)
        self.assertNotIn("Synthetic smoke fixture", output)
        self.assertTrue(
            all(
                method != "GET" or path.count("/") >= 5 for method, path, _, _ in master
            )
        )

    def test_crud_progress_marker_precedes_mutations_and_has_only_scope_tag(
        self,
    ) -> None:
        self.env.update(
            {"SMOKE_ALLOW_FIXTURE_WRITES": "1", "SMOKE_FIXTURE_CLINIC_ID": "1"}
        )
        code, output, _ = self.run_main("--crud")
        self.assertEqual(code, 0)
        marker = json.loads(output.splitlines()[0])
        self.assertEqual(marker["phase"], "fixture_start")
        self.assertEqual(marker["clinic_id"], "1")
        self.assertTrue(marker["run_tag"].startswith("ci-smoke-"))
        self.assertLess(
            output.index('"phase": "fixture_start"'),
            output.index('"master_crud": "PASS"'),
        )
        self.assertIn('"phase": "fixture_start"', self.state.stdout_at_first_post)
        self.assertNotIn("dummy secret", output)
        self.assertNotIn("dummy-cookie", output)
        self.assertNotIn("stg-staff-10000021@example.test", output)

    def test_ambiguous_fixture_create_fails_cleanup_without_deleting_unknown_id(
        self,
    ) -> None:
        for mode in (
            "malformed_body",
            "name_mismatch",
            "invalid_id",
            "server_error",
            "read_failure",
        ):
            with self.subTest(create_mode=mode):
                self.state = FixtureState()
                self.state.create_mode = mode
                self.server.RequestHandlerClass.state = self.state
                self.env.update(
                    {
                        "SMOKE_ALLOW_FIXTURE_WRITES": "1",
                        "SMOKE_FIXTURE_CLINIC_ID": "1",
                    }
                )
                code, output, _ = self.run_main("--crud")
                self.assertEqual(code, 1)
                self.assertIn('"master_crud": "FAIL"', output)
                self.assertIn('"fixture_cleanup": "FAIL"', output)
                self.assertFalse(
                    any(method == "DELETE" for method, _, _, _ in self.state.trace)
                )
                self.assertIn(
                    ("POST", "/api/v1/auth/refresh/logout"),
                    [(method, path) for method, path, _, _ in self.state.trace],
                )
                self.assertNotIn("dummy secret", output)
                self.assertNotIn("dummy-cookie", output)

    def test_known_forbidden_create_is_not_reported_as_unresolved_cleanup(self) -> None:
        self.state.create_mode = "known_forbidden"
        self.env.update(
            {"SMOKE_ALLOW_FIXTURE_WRITES": "1", "SMOKE_FIXTURE_CLINIC_ID": "1"}
        )
        code, output, _ = self.run_main("--crud")
        self.assertEqual(code, 1)
        self.assertIn('"master_crud": "FAIL"', output)
        self.assertIn('"fixture_cleanup": "PASS"', output)
        self.assertFalse(
            any(method == "DELETE" for method, _, _, _ in self.state.trace)
        )

    def test_failed_update_still_cleans_created_group(self) -> None:
        self.state.mode = "update_fail"
        self.env.update(
            {"SMOKE_ALLOW_FIXTURE_WRITES": "1", "SMOKE_FIXTURE_CLINIC_ID": "1"}
        )
        code, output, _ = self.run_main("--crud")
        self.assertEqual(code, 1)
        self.assertIn('"master_crud": "FAIL"', output)
        self.assertIn(
            ("DELETE", "/api/v1/masters/permission-groups/701"),
            [(method, path) for method, path, _, _ in self.state.trace],
        )

    def test_failed_post_update_read_still_cleans_owned_group(self) -> None:
        self.state.mode = "read_fail_once"
        self.env.update(
            {"SMOKE_ALLOW_FIXTURE_WRITES": "1", "SMOKE_FIXTURE_CLINIC_ID": "1"}
        )
        code, output, _ = self.run_main("--crud")
        self.assertEqual(code, 1)
        self.assertIn('"master_crud": "FAIL"', output)
        self.assertIn(
            ("DELETE", "/api/v1/masters/permission-groups/701"),
            [(method, path) for method, path, _, _ in self.state.trace],
        )

    def test_ownership_mismatch_prevents_deletion_and_delete_failure_fails(
        self,
    ) -> None:
        self.state.mode = "ownership_mismatch"
        self.env.update(
            {"SMOKE_ALLOW_FIXTURE_WRITES": "1", "SMOKE_FIXTURE_CLINIC_ID": "1"}
        )
        code, output, _ = self.run_main("--crud")
        self.assertEqual(code, 1)
        self.assertFalse(
            any(method == "DELETE" for method, _, _, _ in self.state.trace)
        )
        self.assertIn('"master_crud": "FAIL"', output)

        self.state = FixtureState()
        self.state.mode = "delete_fail"
        self.server.RequestHandlerClass.state = self.state
        code, output, _ = self.run_main("--crud")
        self.assertEqual(code, 1)
        self.assertIn('"fixture_cleanup": "FAIL"', output)

    def test_missing_owned_fixture_before_delete_fails_but_cleans_other_owned_fixture(
        self,
    ) -> None:
        self.state.mode = "group_disappears_before_cleanup"
        self.env.update(
            {"SMOKE_ALLOW_FIXTURE_WRITES": "1", "SMOKE_FIXTURE_CLINIC_ID": "1"}
        )
        code, output, _ = self.run_main("--crud")
        trace = [(method, path) for method, path, _, _ in self.state.trace]
        group = "/api/v1/masters/permission-groups/701"
        staff = "/api/v1/masters/staffs/702"
        self.assertEqual(code, 1)
        self.assertIn('"master_crud": "FAIL"', output)
        self.assertIn('"fixture_cleanup": "FAIL"', output)
        self.assertNotIn(("DELETE", group), trace)
        self.assertIn(("DELETE", staff), trace)
        staff_delete = trace.index(("DELETE", staff))
        self.assertEqual(trace[staff_delete + 1], ("GET", staff))
        self.assertIn(("POST", "/api/v1/auth/refresh/logout"), trace)

    def make_smoke(self) -> object:
        return SMOKE.Smoke(self.base, self.env["SEEDLOGIN_DEMO_PASSWORD"], None)

    def run_with_smoke(self, smoke: object) -> tuple[int, str]:
        stdout = io.StringIO()
        stderr = io.StringIO()
        with (
            mock.patch.dict(os.environ, self.env, clear=True),
            mock.patch.object(SMOKE.sys, "argv", [str(SCRIPT), "--auth"]),
            mock.patch.object(SMOKE, "Smoke", return_value=smoke),
            contextlib.redirect_stdout(stdout),
            contextlib.redirect_stderr(stderr),
        ):
            code = SMOKE.main()
        return code, stdout.getvalue()


if __name__ == "__main__":
    unittest.main()
