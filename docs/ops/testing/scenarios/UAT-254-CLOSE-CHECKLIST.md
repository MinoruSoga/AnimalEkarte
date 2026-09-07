# UAT-254 close checklist

> **目的**: #254 の stable acceptance mapping を定義する。実行結果や外部 ticket 状態はこの source に保存しない。

## Rules

1. local/mock の結果だけで #254 を close しない。
2. 実 LINE、token、別担当 sign-off は USER 管理 lane で実施する。
3. 実行時に Linear/GitHub の外部 status と acceptance owner を確認する。checkout 内の ignored report の有無から外部 status を推定しない。
4. 証跡は `reports/uat-YYYY-MM-DD/` に保存する。確認済み製品 FAIL は root `bug.md` に記録する。環境・権限・fixture BLOCKED は bug にしない。
5. scenario Markdown に PASS/FAIL/sign-off を書かない。

## Stable acceptance mapping

[GitHub #254 本文](https://github.com/MinoruSoga/AnimalEkarte/issues/254) が要求する 5 フローを次に対応付ける。各 scenario の個別 PASS に加え、同じ対象データを引き継ぐ通し結果を run report に残す。入院/health-card の補完確認で、この 5 フローのいずれかを置き換えない。

| Flow | Scenario / 正本 | Required boundary |
|:--|:--|:--|
| 1. 受付 → 診察 → 検査 → 会計 → 締め（AM/PM/EMG） | [V02 §8–9](V02-accounting-reservation-forms.md)、[S06](S06-record-lock-audit-trail.md)、[S02](S02-exam-abnormal-highlight-lock.md)、[S08](S08-accounting-corrections.md)、[S09](S09-closing-time-boundaries.md) | 会計(医師確認)後の finalize/lock/addendum/audit、会計完了と各締め区分まで。同時刻 fixture 不足の S09 を他の PASS で代替しない |
| 2. 予約 → 来院 → 再予約 | [V02 §7–9](V02-accounting-reservation-forms.md)、[予約からカルテの業務仕様](../../../spec/reservation-to-record-flow.md) | 予約と来院時の appointment/カルテ連携を確認し、次回の予約作成・再表示まで通す。LINE 予約のキャンセル確認だけでは再予約を証明しない |
| 3. トリミング受付 → 実施 → 精算（診察併用含む） | [S11](S11-trimming-combined-accounting.md) | 未請求明細統合、単独/併用精算、appointment 解決 |
| 4. LINE 予約 → カルテ反映 | [S04](S04-liff-reservation-journey.md)、[S06](S06-record-lock-audit-trail.md)、[予約からカルテの業務仕様](../../../spec/reservation-to-record-flow.md) | S04 の予約表示だけで止めず、病院側の来院受付・カルテ作成まで引き継ぐ。mock lane と実 LINE lane を区別し、stopped/error maintenance を維持 |
| 5. 月次集計 → 帳票出力 | [月次集計・帳票仕様](../../../spec/screens/32-accounting-reports.md)、[探索ガイド §2.2](../SECTION_14_MANUAL_TEST_GUIDE.md) | `/accounting/reports` の対象月の会計集計と帳票/出力を突合。S10 の顧客別年間 LTV/CSV は別の集計で、月次帳票の代替にならない |

補完確認: 入院サイクルは [S05](S05-hospitalization-cycle.md)（cage、registration-time plan、二重退院拒否）、LIFF health/account link は [S12](S12-liff-pet-health.md)（LIFF ID 両分岐、owner isolation）。

[2026-08-20 の owner comment](https://github.com/MinoruSoga/AnimalEkarte/issues/254#issuecomment-5352193910) は、実 LINE・token health・DB/audit・residual disposition・別 sign-off を close 条件として明示している。close 条件の参照先は [BRT-45](https://linear.app/baritechllc/issue/BRT-45)、実施レーンは [BRT-68](https://linear.app/baritechllc/issue/BRT-68)。外部の現在状態は実行時に再確認する。

## Close gate

close 判断時に USER が次を確認する。

- 5 flow の最新 run report が同じ対象 revision/environment contract を参照する。
- 臨床安全・会計金額・clinic / owner / pet / staff 分離・認証権限・データ消失の製品 FAIL は Go-live 前に解消する。その他の FAIL だけ、Linear に受容条件を記録し USER が明示受容した場合に納品後対応へ延期できる（[`todo.md` の P4 延期例外](../../../../todo.md)）。
- 実 LINE lane と token health の必要証跡がある。mock のみでは代替しない。
- DB/audit の照合、残件の disposition、実施者とは別の acceptance owner の sign-off が記録されている。
- fixture と cleanup が完了し、共有/STG の既存データを変更していない。
- Linear/GitHub の外部状態を実行時に再確認した。

## 禁止

- ignored report が checkout にないことを未実施/完了の根拠にする
- 過去 comment の結果を現在の結果として転載する
- local/mock 結果だけで close する
- scenario source に dated result/sign-off を埋め込む
