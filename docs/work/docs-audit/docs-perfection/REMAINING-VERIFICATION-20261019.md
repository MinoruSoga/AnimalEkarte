# 残件の検証引継ぎ（2026-10-19、未実行）

固定main `563353287ebb08af55056e3fbfba2aca3531e599` の調査で未検証だった31項目を、既存正本の検査へ渡す対応表。これは実行結果・新業務AC・Planeの代替台帳ではない。列の操作は**後続の承認済み実行レーン用**で、今回の文書作業では一切実施していない。

[検証層](../../../ops/testing/TEST_ARCHITECTURE.md)、[環境準備](../../../ops/testing/UAT-ENV-SETUP.md)、[field記録](../../../ops/testing/scenarios/FIELD-LEVEL-PROTOCOL.md)を優先する。現在の実装が契約に適合する前のケースは実装差として停止し、試験の都合で正本の期待を弱めない。

## 共通の事前条件・証跡・終了条件

- 既存証跡を先に照会し、対象commit、環境、医院/fixture、操作、期待、実結果が一致する部分だけ再利用する。checkoutにreportがないことだけで未実施としない。
- 実行する場合は対象環境/clinic・fixture owner・現在grant・事前件数・期待事後件数・失敗注入方法・cleanup/復旧先・実施者と別の確認者を先に固定する。production/未承認共有医院/実患者でmutating testしない。実LINE・外部送信・実機・DB/migration・provider操作は別承認の人間レーン。
- 成功、拒否、競合、結果不明、未検出を分離する。commit後応答断は元operation照会で確認し、未検出を失敗確定として新keyのPOSTへ置き換えない。再照会にも現在認可を適用し、secretや原患者データをreceiptへ転載しない。
- `reports/uat-YYYY-MM-DD/` にcase/元AC/操作と版・期待/実結果・receipt/必要監査・件数/金額・復旧cleanup・PASS/FAIL/PARTIAL/BLOCKED/N/A理由を保存する。本表やscenario本文へ実行結果を書かない。静的読取と実機合格を分離する。
- 未確認の臨床値、運用窓、backup方式/保持、実送信許可、SLO値は推測せず担当者へ限定確認。不足はBLOCKED。条件付き7項目を今回の必須試験へ増設しない。

## 項目別の対象・操作・期待

全31行とも実行状態は**未実施**。受入の正本は「接続先」であり、この表は試験対象の割当と準備を具体化する。

| ID・接続先 | fixture・前提 | 後続で行う操作 | 照合する期待 |
|:---|:---|:---|:---|
| [CL-R01](../../../design/shared/integration-contracts.md) | 同一record・臨床子owner・旧client/取込の入口台帳 | 更新先行／強制取得先行、子保存と親変更、必要監査失敗・commit応答断 | 元成功保持／旧epoch拒否・子owner独立保存・結果照会収束を各入口で確認 |
| [CL-R04](../../../design/domains/medical-records/technical-spec.md) | 複数頭予約と別owner入力を持つカルテ、タブ開始版/複製前版 | 片側owner保存失敗、送信中clear、copy後Undo、旧回答の遅延取得 | 成功subsetを保持し未保存だけ残す。原文・他者/他ownerを消さず未回答予約は通す |
| [CL-R05](../../../design/domains/medical-records/technical-spec.md) | 旧finalized・専用訂正対象・生前日付き死亡患者の合成fixture | 通常編集と専用訂正を各許可入口から実施、版競合/失権/応答断 | 旧版履歴・目的別理由/医師確認保持。通常編集可から専用記録の無制限変更を導かない |
| [CL-R06](../../../design/domains/medical-records/technical-spec.md) | 医師承認値/単位/適用版が揃う犬猫の対象パターン | 欠損・単位違い・上限直前/一致/超過・未承認種/経路・評価不能 | 1回量/用法/総量を分離し上限等を既決通り停止。数値未承認ならBLOCKED |
| [CL-R08](../../../design/domains/trimming/technical-spec.md) | 単独/診察併用、担当未定、既存割引、終了結果不明 | 開始/終了/請求引継ぎを再送、遅延応答・二端末、確定会計参照 | 同一終了に予約detailを重複作成せず割引一回、終了と精算を分離 |
| [CL-R09](../../../design/domains/checkups/technical-spec.md) | 保存済親のみ／健診あり結果なし／一部結果あり、旧field | 親・健診・結果の各段階で失敗/応答消失して再開、定義変更後再取得 | 同じ親/健診IDへ再開し旧値を保持。空配列全削除と未送信を区別 |
| [CL-R11](../../../design/domains/vaccinations/technical-spec.md) | 単独画面とカルテタブ、旧接種snapshot・手入力次回日 | 種類/接種日/間隔変更、複製、遅延応答、設定読取失敗 | 当時値・手入力・種類別計算と日付検証を分離し黙って消去/上書きしない |
| [CL-R13](../../../spec/lab-device-integration/README.md) | 当日/別日、患者違い、医院違い、取消/終了のカルテ | 候補表示後に対象状態/日付/権限を変え、待機登録を両順で試す | 明示当日カード条件をserver再照合。UI候補表示だけでは合格にしない |
| [CL-R14](../../../design/domains/examinations/technical-spec.md) | 医院/機種/配線/PIMS・承認範囲を記入した専用試験計画 | 受信/患者対応/旧ID/欠落・重複・単位/参照範囲・停止復旧 | 医院と機器別の医師承認・実受信証跡。未承認送信/PU日常選択は実施しない |
| [FIN-R12](../../../design/domains/accounting/technical-spec.md) | 会計/予約/退院/回収/返金/在庫等の実symbol台帳 | 各入口で現在認可失効、期待版変更、必要監査失敗、並行二要求 | 同じ共通gate/owner契約に参加。separate-txをsame-tx済みとしない |
| [FIN-R13](../../../design/domains/accounting/technical-spec.md) | 承認済み旧金銭fixture・未分類/NULL/欠損を含む対照集合 | 移行前後の請求/受領/充当/返金・元日付・税snapshotを照合 | 件数/金額/出所一致、未分類は停止台帳。不明を0/現金へ補完しない |
| [FIN-R16](../../../design/domains/finance-operations/technical-spec.md) | 未提示/提示済返答待ち/承認済/期限切れの原本と後継 | 提示と編集、旧版返答、期限跨ぎ転記、draft保存の応答消失 | 原本不変・後継への旧回答継承なし、提示単価と保険再確認、二重転記なし |
| [FIN-R17](../../../design/domains/finance-operations/technical-spec.md) | 部分入金/後日回収/預り受領・充当解除/取消再請求/別方法返金 | 同一元eventを締め営業区分とJST月次で再集計、訂正後再出力 | 請求と実資金移動を二重加算せず、元日付/資金/訂正関係で差額説明可 |
| [FIN-R22](../../../design/domains/finance-operations/technical-spec.md) | 小数使用・棚卸・返却済/未返却・使用中名称の合成在庫 | 棚卸と出庫、返却二要求、同key再送、名称更新/削除の両順 | 数量版・元出庫返却累計・履歴一致。原物連動やlot新設を要件にしない |
| [FIN-R23](../../../design/domains/accounting/technical-spec.md) | 最新DC-A01〜06等の各金銭intent専用fixture | 医師確認と価格/数量変更の両順、失権、監査失敗、commit後応答断 | 金額のみ再確認不要／臨床変更再確認、確定精算停止、結果照会・監査を照合 |
| [FIN-R24](../../../ops/testing/scenarios/S33-receipt-print-pdf.md) | 医院許可済み帳票・対象版、各出力項目matrix | PDFと実プリンタ別に改頁/欠落/訂正再印刷/基本情報変更後出力 | 全欄証跡と医院承認、印刷から会計/受領/在庫writeなし。PDFだけで物理合格にしない |
| [FIN-R25](../../../spec/domains/accounting/lifecycle.md) | 適用品目/法人登録情報/決済契約/外部会計出力の対象表 | 税務/決済責任者が必要記載・分類・実返金方法・出力を原契約へ照合 | 確認者/対象版/根拠/判定を保護記録へ。文書だけで法令/契約適合を断定しない |
| [IR-08](../../../spec/screens/38-liff-pet-health.md) | 公開allowlist・LIFF本人・単回link両分岐・失効token | 同時消費、commit応答断、他本人/別医院/認可失効、親未確定 | 公開範囲を保ち一度だけ効果、原子失敗、元結果照会。新たな親確定条件なし |
| [IR-13](../../../design/domains/reception/technical-spec.md) | 別軸医師確認、選択日/医院、表示設定保持の受付fixture | pollと日付/医院切替、複数ページ不足、状態更新応答断 | 遅延表示で旧context復活なし。設定変更が予約/会計writeを生まない |
| [IR-17](../../../spec/shared/access-control.md) | 非所属/操作不足/失効/最後の管理権限/lookup障害 | 読取後grant変更、保存競合、cache旧版、復旧とlogoutの両順 | 現在grantで拒否・対象不変を確認。新grant自動付与やpollを追加しない |
| [IR-18](../../../design/domains/clinic-settings/technical-spec.md) | 医院設定/法人設定の各保存単位と既発行版 | 同時編集・変更応答断・旧version・設定変更後の旧帳票再取得 | 操作別元結果と現設定を区別、既発行snapshot不変。全未終了業務解消を新条件にしない |
| [IR-20](../../../ops/testing/scenarios/S23-multi-clinic-scope.md) | S14 T1〜T5、S23 T-A1〜4、検索条件の現行fixture | 権限内全医院初期値/選択保持、自院利用、別人重複、新pet、詳細復帰 | 現目標の画面/API対応と非機密証跡。旧観察PASSを現版合格へ流用しない |
| [REM-OPS-02](../../../ops/infra/production/setup.md) | 承認済本番account/zone/DBrole/対象Worker・責任者 | 担当者がconsole・保護設定・DB-backed health・実フロントの対象を照合 | 対象ID/時刻/版/責任者の非機密receipt。health単体をDB成立としない |
| [REM-OPS-04](../../../ops/infra/production/runbook.md) | 承認方式による一つのbackup artifactと隔離復元先 | 復元・件数/整合照合、CF-only rollback・再復旧・後片付けを責任者実施 | 同artifactのhash/復元結果/所要時間/判断/cleanup。共有DBへ復元しない |
| [REM-OPS-05](../../../delivery/GOLIVE_RUNBOOK.md) | 承認済作業窓・判断者・連絡先・医院権限・初期master | 事前gate確認と当日切替gateを別記して手順通過 | 必要receiptを責任者が確認しGo/No-Go。構築開始日を切替日としない |
| [REM-OPS-06](../../../ops/deploy/seeds/CLINIC_CSV_IMPORT.md) | formal producer入力・payment source・固定candidate/6seed | 同一candidateのmanifest/正件数paymentgraph/source/consumerpinsを照合 | 完全性と一致receiptが揃う。旧candidate PASSを新入力へ転用しない |
| [REM-OPS-08](../../../ops/deploy/seeds/CLINIC_CSV_IMPORT.md) | 承認実規模・固定target/profile/manifest・隔離実行先 | 長時間移行、途中失敗、commit応答消失、結果照合、復旧を実施 | 同runの所要時間/件数/原子性/復旧cleanup receipt。部分完了を成功表示しない |
| [REM-OPS-10](../../../ops/testing/scenarios/FIELD-LEVEL-PROTOCOL.md) | 承認inventory×適用F、C1〜C3、同一revisionの5flow | 既存runの版を照合し不足だけ実行、実LINE/token/audit/別owner承認 | 全適用条件PASS/理由付きN/A、未完/FAILを別記。禁止系FAILはGo前解消 |
| [REM-OPS-11](../../../ops/infra/CLOUDFLARE-EXTERNAL-INTEGRATIONS-AUDIT.md) | 承認済testaccount/医院/targetref、停止と再送責任者 | console配線と実送受信、重複/失敗監視、cron/pause/revision/restart | 実受信・停止・結果不明処置のreceipt。PROD userへの試験や未承認稼働再開なし |
| [REM-OPS-13](../../../ops/agent-harness.md) | 非危険fixture、launcher/root/backend/空白worktree設定 | 設定合成・親override・hookロード/失敗/timeoutと実OS権限を検査 | 実runtime/refと各拒否証跡。JSON exit2だけでホスト上の拒否としない |
| [REM-OPS-14](../../../ops/testing/STG-PERFORMANCE-CHECKLIST.md) | 採用SLO・対象API/データ量/負荷/時間/中止基準 | 承認負荷でAPI別件数/p95/p99/失敗率・回復・停止cleanupを計測 | 同runで採用基準比較。health/clinicsだけの短時間結果を全業務保証にしない |

## 既存実行報告の扱い

[PR #698](https://github.com/MinoruSoga/AnimalEkarte/pull/698)の作者・ユーザー報告では `47f99fbfb7194d319754f00eb99240f8b4e55e99` で `make ci` 全30段階成功。基準mainまでは#722/#725のdocsのみ11パスであり、「make ci未実行」へ戻さない。これは作者報告の再利用で、今回の独立再実行・生artifact確認ではない。STG migration037/038適用、ブラウザ受入、R07/R10、人のChromatic視覚承認は別の未検証として残す。
