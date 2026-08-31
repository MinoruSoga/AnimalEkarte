# SEC-SECRETS-5 — gitleaks 全履歴スキャン インベントリ（2026-07-15）

> 値は記載しない。パスと件数のみ。ローテーション（[外部資格情報オペレーション](./BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) §1）完了まで Issue #89/#97 はクローズしない。
> filter-repo は禁止方針。正攻法はローテーション。

## スキャン条件

- ツール: `gitleaks detect --log-opts='--all'`
- 作業ツリー上の seed LINE 平文は本セッションで除去済み（`003_demo` CSV）

## 履歴ヒット（パスのみ・値なし）

| パス（履歴上） | 備考 |
|----------------|------|
| `.env.local` | ローカル env（現行 tree は gitignore 想定） |
| `.env.staging` | untrack 済み（`6e34e684`）— 履歴残存 |
| `CODING_RULES.md` | 要確認（ダミー可能性） |
| `backend/.env.production` | 履歴上の env |
| `backend/docs/README.md` | 要確認 |
| `backend/migrations/001_init.sql` | 要確認（ダミー可能性） |
| `backend/migrations/002_seed.sql` | 旧ファイル名・履歴 |
| `backend/migrations/003_line_reservation_seed.sql` | 旧ファイル名・履歴 |
| `backend/migrations/003_seed_demo.sql` | 旧 seed SQL・履歴（現行は CSV） |

**findings_count**: 22（ユニークパス 9）

## エージェント完了 / USER 残

- [x] 作業ツリー seed / テストから LINE 実平文除去
- [x] 本インベントリ記録
- [ ] USER: 4 系統ローテーション（[外部資格情報オペレーション](./BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) §1）
- [ ] USER: current secret registration sources（[`infra/cloudflare/README.md`](../../../../infra/cloudflare/README.md) と target Wrangler file）に従って GitHub/Cloudflare secret names を登録
- [ ] USER: #97 本文マスク（ローテーション後）
