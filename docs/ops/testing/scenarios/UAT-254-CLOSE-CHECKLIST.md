# UAT-254 close checklist

> **目的**: #254 の stable acceptance mapping を定義する。実行結果や外部 ticket 状態はこの source に保存しない。

## Rules

1. local/mock の結果だけで #254 を close しない。
2. 実 LINE、token、別担当 sign-off は USER 管理 lane で実施する。
3. 実行時に Linear/GitHub の外部 status と acceptance owner を確認する。checkout 内の ignored report の有無から外部 status を推定しない。
4. 証跡は `reports/uat-YYYY-MM-DD/` に保存する。確認済み製品 FAIL は root `bug.md` に記録する。環境・権限・fixture BLOCKED は bug にしない。
5. scenario Markdown に PASS/FAIL/sign-off を書かない。

## Stable acceptance mapping

| Flow | Scenario | Required boundary |
|:--|:--|:--|
| 外来 → カルテ → 会計 → 監査 | [S06](S06-record-lock-audit-trail.md) | 会計(医師確認)が `confirmed` になる前は確定 disabled。その後 finalize、lock、addendum、audit を確認 |
| 入院 → care plan → 退院会計 | [S05](S05-hospitalization-cycle.md) | cage 必須、registration-time plan、二重退院拒否 |
| trimming 受付 → 診察併用精算 | [S11](S11-trimming-combined-accounting.md) | 未請求明細統合と appointment 解決 |
| LINE 予約 | [S04](S04-liff-reservation-journey.md) | mock lane と実 LINE lane を区別。stopped/error maintenance を維持 |
| LIFF health/account link | [S12](S12-liff-pet-health.md) | LIFF ID 設定/未設定の両分岐と owner isolation |

## Close gate

close 判断時に USER が次を確認する。

- 5 flow の最新 run report が同じ対象 revision/environment contract を参照する。
- 確認済み製品 FAIL が 0、または acceptance owner が納品後対応として明示承認した一覧に隔離されている。
- 実 LINE lane と token flow の必要証跡がある。mock のみでは代替しない。
- fixture と cleanup が完了し、共有/STG の既存データを変更していない。
- Linear/GitHub の外部状態を実行時に再確認した。

## 禁止

- ignored report が checkout にないことを未実施/完了の根拠にする
- 過去 comment の結果を現在の結果として転載する
- local/mock 結果だけで close する
- scenario source に dated result/sign-off を埋め込む
