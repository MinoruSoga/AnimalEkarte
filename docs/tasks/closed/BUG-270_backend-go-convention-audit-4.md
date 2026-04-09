# BUG-270: バックエンド Go コード規約準拠監査（第4回）

## 概要

第3回監査（BUG-261〜266）修正後の残存違反を洗い出す第4回監査。
前回までの3回で主要な multitenancy / FromGORM / naked return / slog 違反は修正済み。
残存は handler→repo 直接アクセス、slog 追加漏れ、テストmock不備、コード品質改善。

## 子チケット一覧

| BUG | 対象 | 内容 | 優先度 | 箇所数 |
|-----|------|------|--------|--------|
| [BUG-271](BUG-271_handler-direct-repo-access.md) | Handler 層 | handler → repository 直接アクセス（service 迂回） | High | 9箇所/3ファイル |
| [BUG-272](BUG-272_slog-audit-log-missing-3.md) | Service 層 | slog 監査ログ欠落 第3波 | High | 8箇所/5ファイル |
| [BUG-273](BUG-273_repository-reorder-outer-double-wrap.md) | Repository 層 | Reorder/Transaction 外側二重ラップ `Wrap` → `return err` | Medium | 11箇所/8ファイル |
| [BUG-274](BUG-274_test-mock-source-param-missing.md) | Test | reservation mock の `source` パラメータ欠落 | High | 2ファイル |
| [BUG-275](BUG-275_misc-medium-violations.md) | 複合 | swaggerignore 残存 / liff_auth URLエンコード / audit_log 型 / liff NOTE コメント / FK チェック | Medium | 複数 |

## クリーン判定レイヤー

| レイヤー | 判定 |
|----------|------|
| model/ | ⚠️ swaggerignore 残存タグ、audit_log.go OldValue/NewValue 型 |
| errors/ | ✅ CLEAN |
| middleware/ | ⚠️ liff_auth.go URL エンコード問題 |
| handler/ | ⚠️ handler→repo 直接アクセス 9箇所 |
| service/ | ⚠️ slog 欠落 8箇所、FK チェック欠落 1箇所 |
| repository/ | ⚠️ Reorder 外側二重ラップ 11箇所 |
| test/ | ⚠️ reservation mock source 欠落 |

## 実施日

2026-04-10
