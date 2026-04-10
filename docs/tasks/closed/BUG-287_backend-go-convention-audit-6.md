# BUG-287: バックエンド Go コード規約準拠監査 第6回

## 概要

BUG-276（第5回最終監査）+ BUG-278/279 修正後の再監査。
CRITICAL ゼロ、HIGH 1件、MEDIUM 6件。

## 子チケット一覧

| BUG | カテゴリ | 内容 | 優先度 |
|-----|---------|------|--------|
| [BUG-288](BUG-288_timeslot-engine-raw-error.md) | Service エラー処理 | timeslot_engine.go の raw error が apperrors を通過しない | High |
| [BUG-289](BUG-289_medium-convention-violations.md) | 複合 MEDIUM | c.JSON 直接使用2件 + slog handler層1件 + 裸return 3件 + slog欠落1件 | Medium |

## 実施日

2026-04-10
