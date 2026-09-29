#!/usr/bin/env python3
# scripts/manual-sync.py
#
# マニュアル記事の DB オーバーライドとリポジトリ MD ファイルを同期するツール。
#
# 正本モデル:
#   frontend/src/features/manual/content/{screens,workflows}/*.md … canonical baseline
#   DB manual_articles … 環境ごとの下書きオーバーライド（閲覧時に MD を覆い隠す）
#
# オーバーライドをベースラインへ取り込む運用:
#   export → git diff 確認 → commit/PR → デプロイ → prune（stale override の後始末）
#
# STG 再構築時の退避・復元（DB_RESET で manual_articles は消える）:
#   rebuild 前:  export --env stg --content-dir <backup-dir>
#   rebuild 後:  push   --env stg --content-dir <backup-dir> --execute
#
# Verbs:
#   status   DB override と MD ファイルの差分一覧（read-only）
#   export   DB override を MD ファイルへ書き出す（既定: リポジトリ content 配下）
#   push     --content-dir 配下の MD ファイルを全て PUT /manual/articles へ送る
#   prune    MD ファイルと内容一致する DB override を DELETE する
#
# 認証: POST /api/v1/login（Cookie セッション）+ X-Requested-With ヘッダ。
#   資格情報は MANUAL_SYNC_EMAIL / MANUAL_SYNC_PASSWORD 環境変数、
#   または --email / --password（password 未指定時は対話プロンプト）。
#
# Usage:
#   python3 scripts/manual-sync.py status --env local
#   python3 scripts/manual-sync.py export --env stg
#   python3 scripts/manual-sync.py push --env stg --content-dir /path/to/backup --execute
#   python3 scripts/manual-sync.py prune --env stg --execute
#
# Exit codes: 0 OK / 1 実行エラー / 2 引数・認証エラー
from __future__ import annotations

import argparse
import getpass
import http.cookiejar
import json
import os
import re
import sys
import tempfile
import urllib.error
import urllib.request
from dataclasses import dataclass
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
DEFAULT_CONTENT_DIR = REPO_ROOT / "frontend" / "src" / "features" / "manual" / "content"
CATEGORIES = ("screens", "workflows")

ENV_BASE_URLS = {
    "local": "http://localhost:8080",
    "stg": "https://animalekarte-stg-api.baritech-soga.workers.dev",
}

# frontend/src/lib/manual-index.ts の parseFrontmatter と同じ規則
FRONTMATTER_RE = re.compile(r"^---\n(.*?)\n---\n(.*)$", re.DOTALL)


@dataclass
class Article:
    category: str
    slug: str
    title: str
    order: float
    section: str
    body: str


def eprint(msg: str) -> None:
    print(msg, file=sys.stderr)


def fail(msg: str, code: int = 1) -> None:
    eprint(f"FAIL  {msg}")
    sys.exit(code)


def parse_md(raw: str, slug: str) -> Article:
    """MD ファイル文字列を Article へ分解（frontmatter 欠落時は FE 既定値に倒す）。"""
    meta: dict[str, str] = {}
    body = raw
    m = FRONTMATTER_RE.match(raw)
    if m:
        fm_block, body = m.group(1), m.group(2)
        for line in fm_block.split("\n"):
            idx = line.find(":")
            if idx < 0:
                continue
            key = line[:idx].strip()
            value = line[idx + 1 :].strip().strip("\"'")
            meta[key] = value
    try:
        order = float(meta["order"]) if "order" in meta else 9999.0
    except ValueError:
        order = 9999.0
    return Article(
        category="",
        slug=slug,
        title=meta.get("title") or slug,
        order=order,
        section=meta.get("section") or "その他",
        body=body,
    )


def render_md(a: Article) -> str:
    """canonical な MD ファイル文字列を生成する。"""
    order = int(a.order) if float(a.order).is_integer() else a.order
    return (
        "---\n"
        f"title: {a.title}\n"
        f"order: {order}\n"
        f"section: {a.section}\n"
        "---\n"
        "\n"
        f"{a.body.strip()}\n"
    )


def norm_body(body: str) -> str:
    return body.strip()


def article_from_db(row: dict) -> Article:
    return Article(
        category=str(row["category"]),
        slug=str(row["slug"]),
        title=str(row["title"]),
        order=float(row["order_value"]),
        section=str(row["section"]),
        body=str(row["body_markdown"]),
    )


def articles_equal(a: Article, b: Article) -> bool:
    return (
        a.title == b.title
        and a.section == b.section
        and float(a.order) == float(b.order)
        and norm_body(a.body) == norm_body(b.body)
    )


def load_file_articles(content_dir: Path) -> dict[tuple[str, str], Article]:
    out: dict[tuple[str, str], Article] = {}
    for cat in CATEGORIES:
        cat_dir = content_dir / cat
        if not cat_dir.is_dir():
            continue
        for path in sorted(cat_dir.glob("*.md")):
            a = parse_md(path.read_text(encoding="utf-8"), path.stem)
            a.category = cat
            out[(cat, a.slug)] = a
    return out


class ApiClient:
    def __init__(self, base_url: str, timeout: float) -> None:
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.cookies = http.cookiejar.CookieJar()
        self.opener = urllib.request.build_opener(
            urllib.request.HTTPCookieProcessor(self.cookies)
        )

    def _request(self, method: str, path: str, payload: dict | None = None) -> tuple[int, bytes]:
        data = None if payload is None else json.dumps(payload).encode("utf-8")
        req = urllib.request.Request(
            f"{self.base_url}{path}",
            data=data,
            method=method,
            headers={
                "Accept": "application/json",
                "Content-Type": "application/json",
                # middleware/csrf.go RequireXRequestedWith 契約
                "X-Requested-With": "XMLHttpRequest",
            },
        )
        try:
            with self.opener.open(req, timeout=self.timeout) as res:
                return res.status, res.read()
        except urllib.error.HTTPError as e:
            return e.code, e.read()
        except urllib.error.URLError as e:
            fail(f"{method} {path}: {e.reason}")

    def login(self, email: str, password: str) -> None:
        code, raw = self._request("POST", "/api/v1/login", {"email": email, "password": password})
        if code != 200:
            fail(f"login failed: HTTP {code} {raw[:200]!r}", code=2)

    def list_articles(self) -> list[Article]:
        code, raw = self._request("GET", "/api/v1/manual/articles")
        if code != 200:
            fail(f"GET /api/v1/manual/articles: HTTP {code} {raw[:200]!r}")
        try:
            rows = json.loads(raw).get("data") or []
        except json.JSONDecodeError:
            fail("GET /api/v1/manual/articles: invalid JSON response")
        return [article_from_db(r) for r in rows]

    def put_article(self, a: Article) -> tuple[int, bytes]:
        return self._request(
            "PUT",
            f"/api/v1/manual/articles/{a.category}/{a.slug}",
            {
                "title": a.title,
                "order_value": a.order,
                "section": a.section,
                "body_markdown": a.body,
            },
        )

    def delete_article(self, a: Article) -> tuple[int, bytes]:
        return self._request("DELETE", f"/api/v1/manual/articles/{a.category}/{a.slug}")


def resolve_base_url(args: argparse.Namespace) -> str:
    if args.base_url:
        return args.base_url.rstrip("/")
    env = args.env or "local"
    if env in ENV_BASE_URLS:
        return ENV_BASE_URLS[env]
    if env.startswith("http://") or env.startswith("https://"):
        return env.rstrip("/")
    fail(f"unknown --env '{env}' (use local|stg|https://...) or pass --base-url", code=2)
    raise AssertionError


def resolve_credentials(args: argparse.Namespace) -> tuple[str, str]:
    email = args.email or os.environ.get("MANUAL_SYNC_EMAIL")
    if not email:
        fail("email is required (--email or MANUAL_SYNC_EMAIL)", code=2)
    password = args.password or os.environ.get("MANUAL_SYNC_PASSWORD")
    if not password:
        password = getpass.getpass(f"password for {email}: ")
    if not password:
        fail("password is required", code=2)
    return email, password


def classify(db: dict[tuple[str, str], Article], files: dict[tuple[str, str], Article]):
    same, diverged, db_only = [], [], []
    for key, a in sorted(db.items()):
        f = files.get(key)
        if f is None:
            db_only.append(key)
        elif articles_equal(a, f):
            same.append(key)
        else:
            diverged.append(key)
    file_only = [k for k in sorted(files) if k not in db]
    return same, diverged, db_only, file_only


def cmd_status(client: ApiClient, args: argparse.Namespace) -> int:
    db = {(a.category, a.slug): a for a in client.list_articles()}
    files = load_file_articles(Path(args.content_dir))
    same, diverged, db_only, file_only = classify(db, files)
    print(f"DB overrides: {len(db)}  files: {len(files)}")
    for key in db_only:
        print(f"  db-only   {key[0]}/{key[1]}  (MD ファイルなし → export で新規作成)")
    for key in diverged:
        print(f"  diverged  {key[0]}/{key[1]}  (DB override が MD と差異あり)")
    for key in same:
        print(f"  same      {key[0]}/{key[1]}  (prune 可能)")
    for key in file_only:
        print(f"  file-only {key[0]}/{key[1]}  (override なし・MD 版が表示中)")
    if diverged:
        print("diverged が残っています: export で取り込むか、override を個別 DELETE してください")
    return 0


def cmd_export(client: ApiClient, args: argparse.Namespace) -> int:
    content_dir = Path(args.content_dir)
    db_articles = client.list_articles()
    if not db_articles:
        print("DB overrides: 0 件（書き出すものなし）")
        return 0
    files = load_file_articles(content_dir)
    created = updated = unchanged = 0
    for a in db_articles:
        key = (a.category, a.slug)
        if a.category not in CATEGORIES:
            eprint(f"WARN  unknown category '{a.category}' for slug '{a.slug}' — skipped")
            continue
        new_text = render_md(a)
        path = content_dir / a.category / f"{a.slug}.md"
        existing = files.get(key)
        if existing is not None and articles_equal(a, existing):
            unchanged += 1
            continue
        action = "create" if existing is None else "update"
        if args.dry_run:
            try:
                shown = path.relative_to(REPO_ROOT)
            except ValueError:
                shown = path
            print(f"  [dry-run] {action}: {shown}")
            if existing is None:
                created += 1
            else:
                updated += 1
            continue
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(new_text, encoding="utf-8")
        if existing is None:
            created += 1
        else:
            updated += 1
        print(f"  {action}: {a.category}/{a.slug}")
    suffix = " (dry-run)" if args.dry_run else ""
    print(f"export{suffix}: created={created} updated={updated} unchanged={unchanged}")
    if not args.dry_run and (created or updated):
        print("次: git diff → commit/PR → デプロイ → 'prune' で override を後始末")
    return 0


def cmd_push(client: ApiClient, args: argparse.Namespace) -> int:
    content_dir = Path(args.content_dir)
    files = load_file_articles(content_dir)
    if not files:
        fail(f"no MD files under {content_dir}/{'|'.join(CATEGORIES)}")
    if not args.execute:
        print(f"[dry-run] {len(files)} 件を PUT /api/v1/manual/articles へ送ります:")
        for (cat, slug), a in sorted(files.items()):
            print(f"  would PUT {cat}/{slug}  (title='{a.title}' order={a.order:g} section='{a.section}')")
        print("実行するには --execute を付けてください")
        return 0
    ok = failed = 0
    for (cat, slug), a in sorted(files.items()):
        code, raw = client.put_article(a)
        if code in (200, 201):
            ok += 1
            print(f"  PUT {cat}/{slug} → {code}")
        else:
            failed += 1
            eprint(f"  FAIL PUT {cat}/{slug} → HTTP {code} {raw[:200]!r}")
    print(f"push: ok={ok} failed={failed}")
    return 1 if failed else 0


def cmd_prune(client: ApiClient, args: argparse.Namespace) -> int:
    db = {(a.category, a.slug): a for a in client.list_articles()}
    files = load_file_articles(Path(args.content_dir))
    same, diverged, db_only, _ = classify(db, files)
    for key in diverged:
        print(f"  skip diverged {key[0]}/{key[1]} (MD と不一致 — 削除しません。export で取り込むか個別判断を)")
    for key in db_only:
        print(f"  skip db-only  {key[0]}/{key[1]} (比較対象の MD なし — 削除しません)")
    if not same:
        print("prune: MD と一致する override なし")
        return 0
    if not args.execute:
        print(f"[dry-run] MD と一致する override {len(same)} 件を DELETE します:")
        for key in same:
            print(f"  would DELETE {key[0]}/{key[1]}")
        print("実行するには --execute を付けてください")
        return 0
    ok = failed = 0
    for key in same:
        code, raw = client.delete_article(db[key])
        if code in (200, 204):
            ok += 1
            print(f"  DELETE {key[0]}/{key[1]} → {code}")
        else:
            failed += 1
            eprint(f"  FAIL DELETE {key[0]}/{key[1]} → HTTP {code} {raw[:200]!r}")
    print(f"prune: deleted={ok} failed={failed}")
    return 1 if failed else 0


def _self_test() -> bool:
    """ネットワーク不要な pure 関数の回帰チェック（scoped verify の契約）。"""
    # frontmatter `order: 0` は有効値として 0 を保持する（既定値 9999 へ丸めない）。
    a = parse_md("---\ntitle: 概要\norder: 0\nsection: 基本\n---\n\n# 概要\n本文\n", "00-overview")
    assert a.order == 0.0 and a.title == "概要" and a.section == "基本", a
    # frontmatter 欠落時は FE 既定値（title=slug, order=9999, section=その他）。
    b = parse_md("# body only\n", "raw")
    assert (b.title, b.order, b.section) == ("raw", 9999.0, "その他"), b
    # render → parse の往復でフィールドが保存される（order=0 含む）。
    c = parse_md(render_md(a), a.slug)
    assert articles_equal(a, c), (a, c)
    # 本文差分は equal ではない。
    assert not articles_equal(a, Article(a.category, a.slug, a.title, a.order, a.section, "別文"))
    # DB 行 → Article（order_value=0 が float 0 になる）。
    d = article_from_db({
        "category": "screens", "slug": "s1", "title": "t",
        "order_value": 0, "section": "sec", "body_markdown": "body",
    })
    assert d.order == 0.0 and d.category == "screens", d
    # classify の4分類。
    same_k = ("screens", "s1")
    files = {same_k: d, ("screens", "f-only"): b}
    db = {same_k: d}
    db[("workflows", "db-only")] = Article("workflows", "db-only", "t", 1, "s", "y")
    same, diverged, db_only, file_only = classify(db, files)
    assert same == [same_k] and diverged == [] and db_only == [("workflows", "db-only")], (same, diverged, db_only)
    assert file_only == [("screens", "f-only")], file_only
    div_db = dict(db)
    div_db[same_k] = Article("screens", "s1", "t", 0, "sec", "changed")
    _, diverged, _, _ = classify(div_db, files)
    assert diverged == [same_k]
    # load_file_articles はカテゴリ dir 配下の *.md のみ拾う。
    with tempfile.TemporaryDirectory() as tmp:
        cat_dir = Path(tmp) / "screens"
        cat_dir.mkdir()
        (cat_dir / "a1.md").write_text("---\ntitle: A\norder: 1\nsection: S\n---\nx\n", encoding="utf-8")
        (cat_dir / "note.txt").write_text("skip", encoding="utf-8")
        loaded = load_file_articles(Path(tmp))
    assert set(loaded) == {("screens", "a1")} and loaded[("screens", "a1")].order == 1.0, loaded
    return True


def main(argv: list[str] | None = None) -> int:
    argv = sys.argv[1:] if argv is None else argv
    if argv == ["--self-test"]:
        return 0 if _self_test() else 1
    p = argparse.ArgumentParser(
        prog="manual-sync.py",
        description="マニュアル DB override ⇔ リポジトリ MD の同期（正本はリポジトリ）",
    )
    sub = p.add_subparsers(dest="verb", required=True)
    for name, helptext in (
        ("status", "DB↔MD 差分一覧（read-only）"),
        ("export", "DB override → MD ファイル書き出し"),
        ("push", "--content-dir の MD 全件 → DB PUT"),
        ("prune", "MD と一致する DB override を DELETE"),
    ):
        sp = sub.add_parser(name, help=helptext)
        sp.add_argument("--env", default="local",
                        help="接続先 (local|stg|https://... の直接指定も可。default: local)")
        sp.add_argument("--base-url", help="API ベース URL（--env より優先）")
        sp.add_argument("--email", help="ログイン email（or MANUAL_SYNC_EMAIL）")
        sp.add_argument("--password", help="ログイン password（or MANUAL_SYNC_PASSWORD）")
        sp.add_argument("--content-dir", default=str(DEFAULT_CONTENT_DIR),
                        help="MD ルート（default: リポジトリ content）")
        sp.add_argument("--timeout", type=float, default=30.0)
        if name in ("export",):
            sp.add_argument("--dry-run", action="store_true", help="書き込まず計画だけ表示")
        if name in ("push", "prune"):
            sp.add_argument("--execute", action="store_true", help="実際に変更を適用（既定は dry-run）")
    args = p.parse_args(argv)

    base_url = resolve_base_url(args)
    email, password = resolve_credentials(args)
    client = ApiClient(base_url, args.timeout)
    client.login(email, password)

    return {
        "status": cmd_status,
        "export": cmd_export,
        "push": cmd_push,
        "prune": cmd_prune,
    }[args.verb](client, args)


if __name__ == "__main__":
    sys.exit(main())
