# BUG-261: バックエンド Go コード規約準拠監査（第3回）

## 概要

第2回監査（BUG-253〜260）修正後の残存違反を洗い出す第3回監査。
前回は handler/repository の clinic_id・FromGORM・slog を主に修正したが、
service 層の naked return と slog 欠落が広範囲に残っている。

## 子チケット一覧

| BUG | 対象 | 内容 | 優先度 | 箇所数 |
|-----|------|------|--------|--------|
| [BUG-262](BUG-262_service-naked-return-errors-3.md) | Service 層 | naked return（apperrors.Wrap 欠落）第3波 | High | 41箇所/13ファイル |
| [BUG-263](BUG-263_slog-audit-log-missing-2.md) | Service 層 | slog 監査ログ欠落 第2波 | Medium | ~18箇所/8ファイル |
| [BUG-264](BUG-264_repository-inner-wrap-to-fromgorm.md) | Repository 層 | トランザクション内 Wrap → FromGORM | Medium | 5箇所/3ファイル |
| [BUG-265](BUG-265_multitenancy-clinic-id-missing-2.md) | Repository/Handler | マルチテナント clinic_id 欠落（第2波） | Critical | 6リポジトリ+8ハンドラ |
| [BUG-266](BUG-266_model-json-tag-and-secret-exposure.md) | Model | VitalRecord json タグ欠落 + LINE シークレット json:"-" 未設定 | High/Medium | 2ファイル |

## 監査範囲

- `backend/internal/repository/` — 全ファイル
- `backend/internal/service/` — 全ファイル
- `backend/internal/handler/` — 全ファイル
- `backend/internal/model/` — 全ファイル
- `backend/internal/errors/` — 全ファイル
- `backend/internal/middleware/` — 全ファイル
- `backend/internal/service/*_test.go` — 全テストファイル

## クリーン判定レイヤー

| レイヤー | 判定 |
|----------|------|
| handler/ | ⚠️ staff_handler/inquiry_handler に extractClinicID 欠落（BUG-265） |
| model/ | ⚠️ vital.go json タグ欠落、reservation_setting.go シークレット json:"-" なし（BUG-266） |
| errors/ | ✅ CLEAN — Sentinel・FromGORM・WrapXxx 完備 |
| middleware/ | ✅ CLEAN — 認証・CORS・ロギング適切 |
| *_test.go | ✅ CLEAN — 全サービスにテストファイル存在 |

## 実施日

2026-04-10
