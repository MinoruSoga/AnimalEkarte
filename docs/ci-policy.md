# CI ポリシー — GitHub Actions ピン記法

> **目的**: GitHub Actions のピン記法(バージョン指定)方針を定義する(#195 決定記録)。
> **読者**: CI ワークフローを変更する開発者。
> **タイミング**: `.github/workflows/*.yml` を編集する時。

（#195 の積み残し「ピン記法ポリシー決定＋文書化」の決定記録）

## ピン記法の基準

| 対象 | 記法 | 理由 |
|---|---|---|
| GitHub 公式 actions（`actions/*`） | メジャータグ（例: `actions/setup-node@v6`） | 公式の改竄リスクは低く、パッチ追従の利便を優先 |
| ベンダー公式 actions（`aws-actions/*`、`golangci/*`、`pnpm/*` 等） | メジャータグまたは完全 semver（例: `@v6.1.0`） | 準公式。完全 semver 併用可（再現性優先の場合） |
| サードパーティ actions（個人・小規模組織） | **コミット SHA ピン**＋バージョンをコメント明記 | サプライチェーン対策（タグは付け替え可能、SHA は不変） |
| シェルからの外部スクリプト取得（curl 等） | **バージョンタグ固定必須**（`main`/`master` 参照禁止） | actionlint.yml の前例（`ba8cecea` で v1.7.12 固定）。未ピンは改竄・非再現の両リスク |

## 運用ルール

1. **新規追加・既存変更時に本基準を適用する**（ratchet 方式）。既存の `uses:` を基準準拠のためだけに一括書き換えるスイープは行わず、該当ワークフローを触る際に合わせる。
2. 同一 action は全ワークフローで**単一バージョンに統一**する（#195 で達成済みの状態を維持）。ドリフト検出は `scripts/check-actions-version-drift.sh`（actionlint.yml の CI job 内で実行・混在を fail）で行う。**actionlint 自体はバージョンドリフトを検出しない**（構文・式チェックのみ — backend-deploy.yml 新設時に @v4 が混入し #195 が回帰した実績あり）。env/with/working-directory のドリフトはスクリプト対象外のため PR レビューで見る（既知の盲点）。
3. バージョン更新は四半期ごとに棚卸しする（actionlint の固定バージョン、SHA ピンの追従を含む）。

## 現状の準拠状況（2026-07-07 更新）

- 2026-07-05 の backend-deploy.yml 新設（Cloudflare Phase 5）で `setup-node@v4` / `pnpm/action-setup@v4` が混入し #195 が一時回帰 → 2026-07-07 に @v6 へ是正し、`check-actions-version-drift.sh` を CI 追加（再発は CI が fail させる）。

### 2026-07-02 時点の記録

- 統一済み: `actions/checkout@v7.0.0` / `actions/setup-node@v6` / `actions/setup-go@v6` / `actions/upload-artifact@v7` / `pnpm/action-setup@v6` / `aws-actions/configure-aws-credentials@v6.1.0` / `aws-actions/amazon-ecs-render-task-definition@v1.8.5` / `dorny/paths-filter@v4` / `golangci/golangci-lint-action@v9`
- SHA ピン済み: security-scan.yml の agentshield
- 既知の未収束: `dorny/paths-filter@v4`（サードパーティ）は本基準では SHA ピン対象 — 次回 paths-filter を触る PR で SHA 化する
