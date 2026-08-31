# ADR-004: 健診機能の正系統 — Checkup パッケージ系に一本化

**Status**: Accepted
**Date**: 2026-07-02（2026-07-22 改定）
**Deciders**: PO（MinoruSoga）

## Context

健康診断（健診）の記録・集計機構が2系統存在する二重管理状態が発生していた。

- **Examination 系（#160 案A）**: 既存の検査 (`exam_types` / `exam_type_fields`) 階層に健診カテゴリをマッピング。
  `003_seed_demo.sql`（現行は削除。運用上は `backend/migrations/seeds/` 配下）に `exam_types` 12000-12003（健康診断/一般健診/歯科健診/老齢健診）と
  exam_type_fields 45-58 を seed 投入してクローズ。
- **Checkup 系（#211）**: 健診パッケージ専用スキーマ（当時は migration 010 として扱われた
  `checkup_type_fields` / `checkup_field_results`、ENUM 型付き子フィールド・閾値機構・歯科垂直スライス）を後続実装。現行は `001_init.sql` へ統合済み。

#160 は案B（Checkup 拡張）を「二重管理になる」と却下したが、#211 が実質的に案Bを実装したため、
同じ「歯科健診の歯石付着度」が両系統に存在する正面衝突が現実化した（2026-07-02 クローズ済み Issue 仕様適合監査で検出）。

## Decision

**Checkup 系（#211 健診パッケージ）を健診機能の唯一の正系統とする。**

- #160 で投入した exam 系健診 seed（exam_types 12000-12003 / exam_type_fields 45-58）は旧 `003_seed_demo.sql` から撤去済み（commit `406c6264`）。SQL-to-CSV 移行後に存在した `seeds/003_demo` bundle も commit `09d2c9e2b` で退役し、HEAD の active seed bundle は `seeds/002_master` だけである。
- ID 12000-12003 / 45-58 は既適用環境に残存しうるため**再利用禁止**。現行 seed に tombstone comment はないため、この ADR を reservation record とする。新しい割当時は当該 ID 範囲を再利用しないことを review で確認する。
- Examination 系は本来の臨床検査（血液検査等）専用に戻る。健診の記録・集計・閾値判定は
  Checkup パッケージ（`001_init.sql`）が担う。

判断根拠: Checkup 系は健診専用に設計され（パッケージ・閾値・型付き結果）、実装も新しく、
`001_init.sql` と整合する。exam 系マッピングは汎用検査機構への相乗りであり健診固有要件
（パッケージ化・種別別閾値）を表現できない。

## Consequences

**ポジティブ:**
- 健診記録の入力導線が1系統に収束し、二重入力・集計不整合・データ移行コストを回避。
- 運用開始前の決定のため、既存患者データの移行は不要。

**注意点:**
- active seed は `backend/migrations/seeds/002_master` のみ。seed / migration の変更は `backend/migrations/CLAUDE.md` と現行 migration policy に従い、適用済み環境の checksum と反映経路を変更ごとに評価する。退役した `003_demo` を編集対象として扱わない。
- checkup パッケージ関連は `001_init.sql` に統合されており、未適用環境では新規デプロイまたは必要な再構築時に適用される。
- 既適用環境の exam_types 12000-12003 / exam_type_fields 45-58 を除去する場合は、影響件数を確認したappend-only migrationとして扱う。暗黙のDB resetに依存しない。

## References

- Issue #160（Examination マッピング採用 → 本 ADR で決定変更）
- Issue #211（Checkup パッケージ実装）
- クローズ済み Issue 仕様適合監査（2026-07-02実施）
- `backend/migrations/001_init.sql`（Checkup 系 DDL：`checkup_type_fields`, `checkup_field_results`）
