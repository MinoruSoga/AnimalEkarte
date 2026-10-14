#!/usr/bin/env python3
import http.cookiejar
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request
import uuid
from typing import Optional

DEMO_EMAIL = "stg-staff-10000021@example.test"
STAGING_HOST = "animalekarte-stg-api.baritech-soga.workers.dev"
MAX_RESPONSE_BYTES = 1024 * 1024


class SmokeError(Exception):
    pass


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(
        self,
        req: urllib.request.Request,
        fp: object,
        code: int,
        msg: str,
        headers: object,
        newurl: str,
    ) -> None:
        return None


def positive_id(value: object) -> str:
    if isinstance(value, bool):
        raise SmokeError("FAIL invalid resource identity")
    text = str(value)
    if not text.isascii() or not text.isdecimal() or not 0 < int(text) < 2**64:
        raise SmokeError("FAIL invalid resource identity")
    return str(int(text))


def settings(mode: str) -> tuple[str, str, Optional[str]]:
    base = os.environ.get("WORKER_URL", "https://" + STAGING_HOST).rstrip("/")
    try:
        url = urllib.parse.urlsplit(base)
        staging = (
            url.scheme == "https"
            and url.hostname == STAGING_HOST
            and url.port in (None, 443)
        )
        local = (
            os.environ.get("SMOKE_ALLOW_LOCAL") == "1"
            and url.scheme == "http"
            and url.hostname in ("127.0.0.1", "localhost", "::1")
            and url.port is not None
        )
        if (
            not (staging or local)
            or url.username
            or url.password
            or url.query
            or url.fragment
            or url.path
        ):
            raise ValueError()
    except ValueError:
        raise SmokeError(
            "BLOCKED target must be the approved staging origin or explicit loopback fixture"
        )
    password = os.environ.get("SEEDLOGIN_DEMO_PASSWORD", "")
    if not password.strip():
        raise SmokeError("BLOCKED SEEDLOGIN_DEMO_PASSWORD is not configured")
    clinic = None
    if mode == "crud":
        if os.environ.get("SMOKE_ALLOW_FIXTURE_WRITES") != "1":
            raise SmokeError("BLOCKED fixture CRUD requires explicit approval")
        clinic = positive_id(os.environ.get("SMOKE_FIXTURE_CLINIC_ID", ""))
    return base, password, clinic


class Smoke:
    def __init__(self, base: str, password: str, clinic: Optional[str]) -> None:
        self.base = base
        self.password = password
        self.clinic = clinic
        self.jar = http.cookiejar.CookieJar()
        self.client = urllib.request.build_opener(
            NoRedirect(), urllib.request.HTTPCookieProcessor(self.jar)
        )
        self.session_open = False
        self.auth_verified = False
        self.logout_verified = False
        self.owned: list[tuple[str, str, str]] = []
        self.cleaned = True
        self.creation_unresolved = False
        self.tag = "ci-smoke-" + uuid.uuid4().hex[:16]

    def request(
        self,
        method: str,
        path: str,
        payload: Optional[dict[str, object]] = None,
        scoped: bool = False,
    ) -> tuple[int, dict[str, object]]:
        headers = {
            "Accept": "application/json",
            "User-Agent": "AnimalEkarte-STG-Demo-Smoke/1.0",
        }
        if method not in ("GET", "HEAD"):
            headers["X-Requested-With"] = "XMLHttpRequest"
        if scoped:
            if self.clinic is None:
                raise SmokeError("FAIL missing fixture clinic scope")
            headers["X-Clinic-ID"] = self.clinic
        data = None
        if payload is not None:
            headers["Content-Type"] = "application/json"
            data = json.dumps(payload).encode("utf-8")
        request = urllib.request.Request(
            self.base + path, data=data, headers=headers, method=method
        )
        try:
            with self.client.open(request, timeout=30) as response:
                if (
                    method == "POST"
                    and path == "/api/v1/login"
                    and response.status == 200
                ):
                    self.session_open = True
                raw = response.read(MAX_RESPONSE_BYTES + 1)
                if len(raw) > MAX_RESPONSE_BYTES:
                    raise SmokeError("FAIL oversized API response")
                body = json.loads(raw.decode("utf-8")) if raw else {}
                if not isinstance(body, dict):
                    raise SmokeError("FAIL unexpected API response shape")
                return response.status, body
        except urllib.error.HTTPError as failure:
            failure.close()
            return failure.code, {}
        except (urllib.error.URLError, OSError, ValueError):
            raise SmokeError("FAIL API transport or JSON response")

    def expect(
        self,
        method: str,
        path: str,
        expected: int,
        payload: Optional[dict[str, object]] = None,
        scoped: bool = False,
    ) -> dict[str, object]:
        status, body = self.request(method, path, payload, scoped)
        if status != expected:
            raise SmokeError(
                "FAIL "
                + method
                + " "
                + path.split("/", 4)[-1].split("/")[0]
                + " HTTP "
                + str(status)
            )
        return body

    def authenticate(self) -> None:
        health = self.expect("GET", "/health", 200)
        if health.get("status") != "ok":
            raise SmokeError("FAIL health status")
        status, body = self.request(
            "POST", "/api/v1/login", {"email": DEMO_EMAIL, "password": self.password}
        )
        self.session_open = status == 200
        self.password = ""
        if status != 200:
            raise SmokeError("FAIL demo login HTTP " + str(status))
        user = body.get("user")
        if not isinstance(user, dict):
            raise SmokeError("FAIL login identity shape")
        identity = positive_id(user.get("id"))
        if not any(
            cookie.name == "access_token" and cookie.value for cookie in self.jar
        ):
            raise SmokeError("FAIL login access cookie absent")
        me = self.expect("GET", "/api/v1/me", 200, scoped=self.clinic is not None)
        if positive_id(me.get("id")) != identity:
            raise SmokeError("FAIL login and me identities differ")
        self.auth_verified = True

    def verify_owned(self, item: tuple[str, str, str]) -> bool:
        kind, identity, name = item
        path = "/api/v1/masters/" + kind + "/" + identity
        status, body = self.request("GET", path, scoped=True)
        if status == 404:
            return False
        if (
            status != 200
            or positive_id(body.get("id")) != identity
            or body.get("name") != name
        ):
            raise SmokeError("FAIL fixture ownership verification")
        return True

    def fixture(
        self,
        kind: str,
        payload: dict[str, object],
        update: dict[str, object],
        field: str,
        value: object,
    ) -> None:
        name = self.tag + ("-group" if kind == "permission-groups" else "-staff")
        payload = dict(payload, name=name)
        path = "/api/v1/masters/" + kind
        self.creation_unresolved = True
        status, created = self.request("POST", path, payload, scoped=True)
        if status != 201:
            self.creation_unresolved = not 400 <= status < 500
            raise SmokeError("FAIL fixture creation HTTP " + str(status))
        identity = positive_id(created.get("id"))
        if created.get("name") != name:
            raise SmokeError("FAIL created fixture ownership; inspect run tag manually")
        item = (kind, identity, name)
        self.owned.append(item)
        self.creation_unresolved = False
        if not self.verify_owned(item):
            raise SmokeError("FAIL created fixture not readable")
        self.expect(
            "PATCH", path + "/" + identity, 200, dict(update, name=name), scoped=True
        )
        changed = self.expect("GET", path + "/" + identity, 200, scoped=True)
        if (
            positive_id(changed.get("id")) != identity
            or changed.get("name") != name
            or changed.get(field) != value
        ):
            raise SmokeError("FAIL fixture update verification")

    def crud(self) -> None:
        print(
            json.dumps(
                {
                    "phase": "fixture_start",
                    "run_tag": self.tag,
                    "clinic_id": self.clinic,
                }
            ),
            flush=True,
        )
        self.fixture(
            "permission-groups",
            {
                "description": "Synthetic smoke fixture",
                "color": "#6366F1",
                "is_active": True,
                "sort_order": 99,
            },
            {"description": "Updated synthetic smoke fixture"},
            "description",
            "Updated synthetic smoke fixture",
        )
        self.fixture(
            "staffs", {"sort_order": 99}, {"sort_order": 100}, "sort_order", 100
        )

    def cleanup(self) -> None:
        if self.creation_unresolved:
            self.cleaned = False
            print("FAIL unverified fixture creation; inspect run tag", file=sys.stderr)
        for item in reversed(self.owned):
            try:
                if not self.verify_owned(item):
                    raise SmokeError(
                        "FAIL fixture missing before deletion verification"
                    )
                kind, identity, _ = item
                path = "/api/v1/masters/" + kind + "/" + identity
                status, _ = self.request("DELETE", path, scoped=True)
                if status not in (200, 204):
                    raise SmokeError("FAIL fixture deletion")
                status, _ = self.request("GET", path, scoped=True)
                if status != 404:
                    raise SmokeError("FAIL fixture deletion confirmation")
            except Exception:
                self.cleaned = False
                print("FAIL fixture cleanup; inspect run tag", file=sys.stderr)
        if self.session_open or any(
            cookie.name in ("access_token", "refresh_token") and cookie.value
            for cookie in self.jar
        ):
            try:
                status, _ = self.request("POST", "/api/v1/auth/refresh/logout")
                self.logout_verified = status == 200
            except Exception:
                self.logout_verified = False
            if not self.logout_verified:
                print("FAIL logout cleanup", file=sys.stderr)
        self.jar.clear()
        self.password = ""


def main() -> int:
    if len(sys.argv) > 2 or (
        len(sys.argv) == 2 and sys.argv[1] not in ("--auth", "--crud")
    ):
        print("BLOCKED supported modes are --auth and --crud", file=sys.stderr)
        return 2
    mode = "crud" if len(sys.argv) == 2 and sys.argv[1] == "--crud" else "auth"
    try:
        base, password, clinic = settings(mode)
    except SmokeError as failure:
        print(str(failure), file=sys.stderr)
        return 2
    smoke = Smoke(base, password, clinic)
    verified = False
    try:
        smoke.authenticate()
        if mode == "crud":
            smoke.crud()
        verified = True
    except SmokeError as failure:
        print(str(failure), file=sys.stderr)
    except Exception:
        print("FAIL unexpected smoke error", file=sys.stderr)
    finally:
        smoke.cleanup()
    auth_ok = smoke.auth_verified and smoke.logout_verified
    result = {
        "run_tag": smoke.tag,
        "auth": "PASS" if auth_ok else "FAIL",
        "master_crud": ("PASS" if verified and smoke.cleaned else "FAIL")
        if mode == "crud"
        else "NOT_RUN",
        "fixture_cleanup": "PASS" if smoke.cleaned else "FAIL",
        "accounting_smoke": "NOT_RUN",
    }
    print(json.dumps(result, sort_keys=True))
    return 0 if verified and auth_ok and smoke.cleaned else 1


if __name__ == "__main__":
    sys.exit(main())
