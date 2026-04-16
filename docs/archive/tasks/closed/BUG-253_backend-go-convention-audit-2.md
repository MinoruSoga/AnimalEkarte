# BUG-253: バックエンド Go コード規約準拠監査（第2回） — 全ドメイン横断

## 概要

BUG-244（第1回監査）の全修正完了後に実施した差分監査。
backend/ 全ドメインを5並列 Go Reviewer エージェントでスキャンした結果、
**HIGH 11件カテゴリ・MEDIUM 多数** の違反を検出した。
前回修正済みの項目（BUG-245〜252）は再発していない。

**実施日**: 2026-04-09
**検証方法**: Go Reviewer エージェント 5並列（A〜E グループ）→ 全コード静的読み取り

---

## 子チケット一覧

| BUG | 優先度 | 内容 | 影響ドメイン数 |
|-----|--------|------|---------------|
| [BUG-254](BUG-254_multitenancy-clinic-id-missing.md) | Critical | マルチテナント clinic_id 欠落（クロスクリニック参照可能） | 8ドメイン |
| [BUG-255](BUG-255_repository-fromgorm-in-reorder.md) | High | Repository Reorder/トランザクション内で `apperrors.Wrap` → `FromGORM` に統一 | 11リポジトリ |
| [BUG-256](BUG-256_service-naked-return-errors-2.md) | High | Service 層で `apperrors.Wrap` なし naked return（第2波） | 20+サービス |
| [BUG-257](BUG-257_slog-audit-log-violations.md) | High | slog 監査ログ欠落・順序不正・レイヤー違反 | 8サービス+1ハンドラ |
| [BUG-258](BUG-258_handler-direct-cjson.md) | High | Handler で `c.JSON` 直接使用（`RespondError` 迂回） | 4ファイル |
| [BUG-259](BUG-259_delete-fk-dependency-check-missing.md) | High | マスタ削除時の FK 依存チェック欠如 | 2サービス |
| [BUG-260](BUG-260_error-ignoring-and-misc.md) | Medium | Count エラー無視・liff エラー無視・税率ハードコード・重複チェック等 | 複数 |

## 修正優先順位

1. **BUG-254** (Critical) — マルチテナント漏れ。セキュリティ脆弱性。
2. **BUG-255** (High) — Repository レイヤー規約統一
3. **BUG-256** (High) — Service レイヤー規約統一
4. **BUG-257** (High) — 監査ログ整備
5. **BUG-258** (High) — Handler レスポンス統一
6. **BUG-259** (High) — FK 依存チェック追加
7. **BUG-260** (Medium) — その他

## 関連規約

- `.claude/CLAUDE.md` — エラー処理の統一 (MANDATORY)
- `.claude/rules/go-language.md` — Context伝播・エラーハンドリング
- `.claude/rules/error-handling.md` — Repository FromGORM / Service Wrap / Handler RespondError
- `.claude/rules/database-design.md` — マルチテナント clinic_id 必須
- `.claude/rules/code-style.md` — handler → service → repository レイヤー分離
