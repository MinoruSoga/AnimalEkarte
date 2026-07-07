# ADR-004: 健診機能の正系統 — Checkup パッケージ系に一本化

**Status**: Accepted
**Date**: 2026-07-02
**Deciders**: PO（MinoruSoga）

## Context

健康診断（健診）の記録・集計機構が2系統存在する二重管理状態が発生していた。

- **Examination 系（#160 案A）**: 既存の検査 (`exam_types` / `exam_type_fields`) 階層に健診カテゴリをマッピング。
  `003_seed_demo.sql` に exam_types 12000-12003（健康診断/一般健診/歯科健診/老齢健診）と
  exam_type_fields 45-58 を seed 投入してクローズ。
- **Checkup 系（#211）**: 健診パッケージ専用スキーマ（migration 010: `checkup_type_fields` /
  `checkup_field_results`、ENUM 型付き子フィールド・閾値機構・歯科垂直スライス）を後続実装。

#160 は案B（Checkup 拡張）を「二重管理になる」と却下したが、#211 が実質的に案Bを実装したため、
同じ「歯科健診の歯石付着度」が両系統に存在する正面衝突が現実化した（2026-07-02 クローズ済み Issue 仕様適合監査で検出）。

## Decision

**Checkup 系（#211 健診パッケージ）を健診機能の唯一の正系統とする。**

- #160 で投入した exam 系健診 seed（exam_types 12000-12003 / exam_type_fields 45-58）は
  `003_seed_demo.sql` から撤去（commit `406c6264`）。
- ID 12000-12003 / 45-58 は既適用環境に残存しうるため**再利用禁止**（seed 内の墓標コメントに明記）。
- Examination 系は本来の臨床検査（血液検査等）専用に戻る。健診の記録・集計・閾値判定は
  Checkup パッケージ（migration 010）のみが担う。

判断根拠: Checkup 系は健診専用に設計され（パッケージ・閾値・型付き結果）、実装も新しく、
migration 010 と整合する。exam 系マッピングは汎用検査機構への相乗りであり健診固有要件
（パッケージ化・種別別閾値）を表現できない。

## Consequences

**ポジティブ:**
- 健診記録の入力導線が1系統に収束し、二重入力・集計不整合・データ移行コストを回避。
- 運用開始前の決定のため、既存患者データの移行は不要。

**注意点:**
- 適用済み seed（003）の編集のため、次回 STG デプロイは `db_reset=true` 必須（checksum mismatch 回避）。
- migration 010 は未適用環境がある — db_reset と同時に適用する。
- 既適用環境の exam_types 12000-12003 / exam_type_fields 45-58 の残存行は db_reset で消える。

## References

- Issue #160（Examination マッピング採用 → 本 ADR で決定変更）
- Issue #211（Checkup パッケージ実装）
- クローズ済み Issue 仕様適合監査（2026-07-02実施）
- `backend/migrations/010_add_checkup_packages.sql`
