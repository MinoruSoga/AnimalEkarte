# 過去のseed医院UATソース

既存の医院ID、seedアカウント、臨床fixture、一時権限変更に依存する過去の試験を保全する。通常のE2Eへ混在させず、実行は停止する。

| ファイル | 当時のパスと用途 |
|---|---|
| `seeded-clinic-master-forms.archived.ts` | `v04-settings-master-forms-retest.spec.ts`。V04マスタフォームの2026-09-23再確認 |
| `seeded-clinic-uat-residuals.archived.ts` | `uat-2026-09-24-residuals.spec.ts`。予約・印刷等の2026-09-24再確認 |

`.archived.ts` はPlaywrightと検証用inventoryの `*.spec.ts` 発見対象に入らず、Playwrightの `testIgnore` でもこのフォルダを除外する。現在のE2E用tsconfigの型検査対象からも外れる。過去の結果や旧パスは履歴として保持し、現行手順・現在のPASS判定へ転用しない。

再利用する場合は、必要なケースを使い捨てfixtureの通常specへ移植する。承認済みの対象環境、合成アカウント、隔離医院、失敗時も完了するcleanupを確認し、固定IDや既存医院への一時権限変更に依存させない。拡張子の復帰、除外の解除、独自configによる実行再開は行わない。

通常の `run-e2e.sh --v04` は `v04-settings-master-forms.spec.ts` の既存fixture経路を使う。本整理ではブラウザ・backend・DBへの要求を実行していない。
