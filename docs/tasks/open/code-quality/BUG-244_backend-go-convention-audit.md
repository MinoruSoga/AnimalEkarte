# BUG-244: バックエンド Go コード規約準拠監査 — 全ドメイン横断

## 概要

backend/ 全ドメイン（30+ドメイン）を対象にプロジェクトコード規約への準拠状況を監査した結果、
CRITICAL 3件・HIGH 11件・MEDIUM 多数の違反を検出した。
本チケットは **親チケット** であり、修正は子チケット BUG-245〜BUG-252 で実施する。

**実施日**: 2026-04-09
**検証方法**: Go Reviewer エージェント 5並列 → 全コード静的読み取り

---

## 子チケット一覧

| BUG | 優先度 | 内容 | 影響ドメイン数 |
|-----|--------|------|---------------|
| [BUG-245](BUG-245_price-pointer-dereference.md) | Critical | `buildXxxUpdateFields` の price ポインタ未デリファレンス | 6ファイル |
| [BUG-246](BUG-246_staff-handler-business-logic-leak.md) | Critical | staff_handler に bcrypt/Account操作が漏出 + エラー無視 + 非トランザクション | 1ドメイン |
| [BUG-247](BUG-247_clinical-plan-missing-clinic-id.md) | Critical | clinical_plan に clinic_id マルチテナント境界なし | 1ドメイン |
| [BUG-248](BUG-248_repository-fromgorm-violations.md) | High | Repository 層で `apperrors.FromGORM` 未使用 | 15+リポジトリ |
| [BUG-249](BUG-249_service-naked-return-errors.md) | High | Service 層で `apperrors.Wrap` なし naked return | 12+サービス |
| [BUG-250](BUG-250_auth-handler-direct-repo-access.md) | High | auth_handler の直接 repository アクセス | 1ドメイン |
| [BUG-251](BUG-251_gorm-error-comparison-without-errors-is.md) | High | `gorm.ErrRecordNotFound` を `==` で比較（`errors.Is` 未使用） | 2リポジトリ |
| [BUG-252](BUG-252_misc-high-medium-violations.md) | High/Medium | examination enum未検証 / slog handler層使用 / liff `_ =` エラー無視 / N+1 等 | 複数 |

## 修正優先順位

1. **BUG-245** (Critical) — price データ破壊。即修正（6箇所、各1行）— 10分
2. ~~**BUG-246**~~ ✅ FIXED — staff_handler リファクタ (`136078c6`)
3. ~~**BUG-247**~~ ✅ FIXED — clinic_id 追加 (`e680bdaa`)
4. ~~**BUG-248**~~ ✅ FIXED — FromGORM 一括置換 27ファイル ~120箇所 (`136078c6`)
5. ~~**BUG-249**~~ ✅ FIXED — naked return 修正 15ファイル (`136078c6`)
6. ~~**BUG-250**~~ ✅ FIXED — auth_handler → service 経由 (`136078c6`)
7. ~~**BUG-251**~~ ✅ FIXED — errors.Is 置換 (`5b06d2c3`)
8. ~~**BUG-252**~~ ✅ FIXED — examination enum / liff エラー / pet Count (`136078c6`)

**全チケット修正完了** (2026-04-09)

## 関連規約

- `.claude/CLAUDE.md` — エラー処理の統一 (MANDATORY)
- `.claude/rules/go-language.md` — Context伝播・エラーハンドリング
- `.claude/rules/error-handling.md` — Repository FromGORM / Service Wrap / Handler RespondError
- `.claude/rules/database-design.md` — マルチテナント clinic_id 必須
- `.claude/rules/code-style.md` — handler → service → repository レイヤー分離
