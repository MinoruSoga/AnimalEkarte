# scripts/rehearsal/lib/ — 共有 mjs モジュール

`check-a4-*` / `run-f8-g4-*` / `write-a4-*` から import される共有ライブラリ。
直接実行するスクリプトではない（エントリポイントは `scripts/` 直下側）。

## ファイル

| ファイル | 提供関数 | 用途 |
|------|------|------|
| `a4-env-file.mjs` | `validateA4EnvFile({repo, input})` | A4 env ファイル検証 |
| `a4-rehearsal-compose.mjs` | `validateA4ComposeConfig(config)` | A4 リハーサル compose 設定検証 |
| `a4-resource-boundary.mjs` | `validateA4ResourceIdentity` / `assertFreshA4Resources` | A4 リソース（container/network/volume）の境界・新規性検証 |
| `a4-runtime-report.mjs` | `buildA4RuntimeReport` | A4 ランタイムレポート生成 |
| `f8-g4-evidence.mjs` | `canonicalEvidenceBytes` / `artifactBytes` | F8-G4 証跡バイト列の正規化（再現可能な比較用） |
| `f8-g4-host-safety.mjs` | `sanitizedGitEnvironment` / `rejectDockerEnvironmentOverrides` | F8-G4 ホスト実行時の環境変数サニタイズ |

## 規約

- `*.test.mjs` が対になっているものは退行テスト。変更時はテストも更新する
- ここを変更すると呼び出し側の `check-a4-*` / `run-f8-g4-*` 全系に影響する — 影響範囲は `grep -rn "lib/" scripts/` で確認
