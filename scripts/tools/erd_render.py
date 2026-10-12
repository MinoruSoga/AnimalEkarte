"""Deterministic Markdown/Mermaid views of the bounded repository DDL model."""

from __future__ import annotations

import hashlib
import html
import json
import re
from pathlib import Path

from erd_schema import Column, Constraint, Schema, Table


MARKER = "<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->"
GENERATOR = "scripts/tools/generate-erd-detail.py"


def cell(value: str) -> str:
    return value.replace("\n", " ")


def code(value: str | None) -> str:
    if value is None:
        return "—"
    if "|" in value or "`" in value:
        escaped = html.escape(value.replace("\n", " ")).replace("|", "&#124;")
        return f"<code>{escaped}</code>"
    return f"`{cell(value)}`"


def source_link(source: str) -> str:
    name, line = source.rsplit(":", 1)
    return f"[{name}:{line}](../../../backend/migrations/{name}#L{line})"


def foreign_keys(table: Table) -> tuple[Constraint, ...]:
    return tuple(c for c in table.constraints if c.kind == "foreign")


def key_marks(table: Table, column: str) -> tuple[str, ...]:
    kinds = {c.kind for c in table.constraints if column in c.columns}
    marks = [mark for kind, mark in (("primary", "PK"), ("foreign", "FK"), ("unique", "UK")) if kind in kinds]
    if "UK" not in marks and any(i.unique and i.predicate is None and column in i.columns for i in table.indexes):
        marks.append("UK")
    return tuple(marks)


def attribute(table: Table, column: Column) -> str:
    declared = re.sub(r"[^a-zA-Z0-9_\[\]-]", "_", column.type_sql).strip("_")
    marks = key_marks(table, column.name)
    return f"        {declared} {column.name}" + (" " + ",".join(marks) if marks else "")


def child_cardinality(table: Table, fk: Constraint) -> str:
    keys = [c.columns for c in table.constraints if c.kind in ("primary", "unique")]
    keys.extend(i.columns for i in table.indexes if i.unique and i.predicate is None and i.columns)
    return "o|" if any(set(key) <= set(fk.columns) for key in keys) else "o{"


def diagram(schema: Schema, table: Table) -> str:
    fks = foreign_keys(table)
    parents = sorted({fk.target_table for fk in fks if fk.target_table != table.name})
    if len(parents) + 1 > 20:
        raise ValueError(f"{table.name}: ERD exceeds 20 nodes; split its diagram before generation")
    lines = ["erDiagram", f"    {table.name} {{"]
    lines.extend(attribute(table, column) for column in table.columns)
    lines.append("    }")
    for name in parents:
        parent = schema.tables[name]
        referenced = {col for fk in fks if fk.target_table == name for col in fk.target_columns}
        lines.append(f"    {name} {{")
        lines.extend(attribute(parent, col) for col in parent.columns if col.name in referenced)
        lines.append("    }")
    nullable = {c.name: c.nullable for c in table.columns}
    for fk in fks:
        nulls = [nullable[col] for col in fk.columns]
        match_full = bool(re.search(r"\bMATCH\s+FULL\b", fk.sql, re.IGNORECASE))
        parent_cardinality = "o|" if (all(nulls) if match_full else any(nulls)) else "||"
        # Solid ERD lines mean identifying relations; require FK to be part of the child's PK.
        primary = next((set(c.columns) for c in table.constraints if c.kind == "primary"), set())
        line = "--" if set(fk.columns) <= primary else ".."
        lines.append(f'    {fk.target_table} {parent_cardinality}{line}{child_cardinality(table, fk)} {table.name} : "{", ".join(fk.columns)}"')
    return "\n".join(lines)


def table_page(schema: Schema, table: Table) -> str:
    lines = [MARKER, f"# `{table.name}`", "", "[詳細ERD一覧](index.md) · [概要ERD](../erd.md)", "",
             "このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。", "",
             "## カラム", "",
             "型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。", "",
             "| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |",
             "|:---|:---|:---|:---|:---|:---|"]
    for col in table.columns:
        lines.append(f"| {code(col.name)} | {code(col.type_sql)} | {'NULL' if col.nullable else 'NOT NULL'} | {code(col.default_sql)} | {', '.join(key_marks(table, col.name)) or '—'} | {source_link(col.source)} |")
    lines.extend(["", "PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。", "",
                  "## ERD", "", "このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。", "",
                  "親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。", "",
                  "```mermaid", diagram(schema, table), "```", "", "## 外部キー（全件）", ""])
    fks = foreign_keys(table)
    if fks:
        lines.extend(["| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |", "|:---|:---|:---|:---|:---|:---|"])
        for fk in fks:
            lines.append(f"| {code(fk.name)} | {code(', '.join(fk.columns))} | [{fk.target_table}]({fk.target_table}.md) | {code(', '.join(fk.target_columns))} | {code('ON DELETE ' + (fk.on_delete or 'NO ACTION'))} | {source_link(fk.source)} |")
        lines.extend(["", "正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:", ""])
        for fk in fks:
            lines.extend([f"- {code(fk.name)}: {code(fk.sql)}"])
    else:
        lines.append("外部キーなし。")
    other = [c for c in table.constraints if c.kind != "foreign"]
    lines.extend(["", "## その他の制約", ""])
    if other:
        lines.extend(["| 制約 | 種別 | DDL | 定義元 |", "|:---|:---|:---|:---|"])
        for c in other:
            lines.append(f"| {code(c.name)} | {c.kind} | {code(c.sql)} | {source_link(c.source)} |")
    else:
        lines.append("対象の制約なし。")
    lines.extend(["", "## インデックス", ""])
    if table.indexes:
        lines.extend(["| 名前 | UNIQUE | 述語 | DDL | 定義元 |", "|:---|:---|:---|:---|:---|"])
        for index in table.indexes:
            lines.append(f"| {code(index.name)} | {'YES' if index.unique else 'NO'} | {code(index.predicate)} | {code(index.sql)} | {source_link(index.source)} |")
    else:
        lines.append("明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。")
    text = "\n".join(lines) + "\n"
    if len(text.splitlines()) >= 800:
        raise ValueError(f"{table.name}: generated page exceeds 799 lines")
    return text


def render_documents(schema: Schema, generator_dir: Path) -> dict[str, str]:
    for table in schema.tables.values():
        for name in [table.name] + [c.name for c in table.columns]:
            if not re.fullmatch(r"[a-z_][a-z0-9_]*", name):
                raise ValueError(f"Unsupported Mermaid/file identifier: {name!r}; add safe aliases before generation")
    migrations = [{"name": m.name, "sha256": m.sha256} for m in schema.migrations]
    digest = hashlib.sha256(json.dumps(migrations, sort_keys=True, separators=(",", ":")).encode()).hexdigest()
    counts = {"tables": len(schema.tables), "columns": sum(len(t.columns) for t in schema.tables.values()),
              "foreign_keys": sum(len(foreign_keys(t)) for t in schema.tables.values())}
    files = {f"{name}.md": table_page(schema, schema.tables[name]) for name in sorted(schema.tables)}
    index = [MARKER, "# 詳細ERD — 全カラム・全外部キー", "", "[概要ERD](../erd.md)", "",
             f"マイグレーション **{len(migrations)}ファイル** から **{counts['tables']}テーブル・{counts['columns']}カラム・{counts['foreign_keys']}外部キー** を静的に導出しています。", "",
             "DDLは `backend/migrations/` 直下の `.sql` をファイル名順に読み、CREATE/ALTER/DROPとレビュー済みDOブロックの変更を反映します。seed、`schema_migrations`（migrate runnerの管理テーブル）、関数・trigger・RLS・実データはこの一覧の対象外です。", "",
             "この解析器は、対応済みDDLの列・キー・indexを扱う限定的な静的解析器です。未対応のschema変更や未レビューのDOブロックはエラーにします。SQLは実行しません。DBへの適用成功・稼働DBとの一致・STG/本番受入の証明にはなりません。", "",
             "型とDEFAULTはDDLの宣言表記です。PKが暗黙に与えるNOT NULLを反映しています。複合FKの列順、NULL可否、削除動作、部分uniqueの述語は各ページを参照してください。", "",
             "## 更新・確認", "", "手編集せず、リポジトリのルートから実行します（Python標準ライブラリのみ。DBアクセス・依存導入なし）。", "",
             "```sh", "python3 -B scripts/tools/generate-erd-detail.py --write", "python3 -B scripts/tools/generate-erd-detail.py --check", "python3 -B scripts/tools/test_erd_schema.py", "python3 -B scripts/tools/test_generate_erd_detail.py", "```", "",
             "`--check` はファイルを書き換えず、カラム・FK・文書・DDL内容・生成器の更新漏れを検出します。未知のSQLが追加された場合は解析器とテストを先に更新します。`--write` は自動生成マーカーを持つ文書だけを更新・整理し、同名の手書き文書の上書きを拒否します。", "",
             "## テーブル一覧", "", "各ページはテーブルの全カラム・全FKを持ち、参照先のページへ移動できます。", "",
             "| テーブル | カラム数 | 外部キー数 |", "|:---|---:|---:|"]
    for name, table in sorted(schema.tables.items()):
        index.append(f"| [{name}]({name}.md) | {len(table.columns)} | {len(foreign_keys(table))} |")
    index.extend(["", "## 入力の照合情報", "", f"入力digest（SHA-256）: `{digest}`。各ファイルのSHA-256と生成器のSHA-256は [manifest.json](manifest.json) に記録します。", "",
                  "| migration | SHA-256 |", "|:---|:---|"])
    for m in migrations:
        index.append(f"| [{m['name']}](../../../backend/migrations/{m['name']}) | `{m['sha256']}` |")
    index.extend(["", "レビュー済みDOブロック（本文変更・新規追加は再レビューが必要）:", ""])
    index.extend(f"- {block}" for block in schema.reviewed_do_blocks)
    files["index.md"] = "\n".join(index) + "\n"
    source_paths = sorted({generator_dir / "generate-erd-detail.py", *generator_dir.glob("erd_*.py")})
    manifest = {"format_version": 1, "generator": GENERATOR, "source_digest": digest,
                "generator_sources": [{"name": p.name, "sha256": hashlib.sha256(p.read_bytes()).hexdigest()} for p in source_paths],
                "counts": counts, "migrations": migrations, "reviewed_do_blocks": list(schema.reviewed_do_blocks), "files": sorted(files)}
    files["manifest.json"] = json.dumps(manifest, ensure_ascii=False, indent=2) + "\n"
    return files
