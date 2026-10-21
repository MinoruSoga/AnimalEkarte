# rehearsals/ — 検証・リハーサル手順

> **目的**: デプロイや投入作業の検証・リハーサル（disposable 環境・スモーク確認）の手順をまとめる。
> **読者**: デプロイ・リハーサル実施者。
> **タイミング**: リハーサル・デプロイ後検証の実施時。

| ドキュメント | 内容 |
|---|---|
| [A4_UI_REHEARSAL.md](A4_UI_REHEARSAL.md) | A4 UI rehearsal 用の localhost-only disposable Compose 環境 |
| [F8_G4_FAILURE_REHEARSAL.md](F8_G4_FAILURE_REHEARSAL.md) | synthetic FK 違反による rollback 証明リハーサル |
| [CRUD-SMOKE-TEST.md](CRUD-SMOKE-TEST.md) | デプロイ直後の CRUD スモークテスト |
| [MIXED-PAYMENT-SMOKE-TEST.md](MIXED-PAYMENT-SMOKE-TEST.md) | 混在会計 (payment_splits) スモークテスト |
| [VERCEL-FRONTEND-STAGING-TEST.md](VERCEL-FRONTEND-STAGING-TEST.md) | 旧 Vercel 検証手順（履歴資料としてのみ参照可。Vercel は現行配信・rollback のいずれにも使用しない） |
