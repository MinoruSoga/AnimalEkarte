<a id="残件の検証引継ぎ2026-10-19未実行"></a>

# 残件の検証引継ぎ（2026-10-19、既存証跡照合・追加実行なし）

固定main `563353287ebb08af55056e3fbfba2aca3531e599` の調査で未検証だった31項目を、既存正本の検査へ渡す対応表。既存実行報告の限定的な結果は以下で参照し、この引継ぎ自体を新しい実行結果・新業務AC・Planeの代替台帳にはしない。列の操作は**後続の承認済み実行レーン用**で、今回の文書作業では一切実施していない。

[検証層](../../../ops/testing/TEST_ARCHITECTURE.md)、[環境準備](../../../ops/testing/UAT-ENV-SETUP.md)、[field記録](../../../ops/testing/scenarios/FIELD-LEVEL-PROTOCOL.md)を優先する。現在の実装が契約に適合する前のケースは実装差として停止し、試験の都合で正本の期待を弱めない。

## 共通の事前条件・証跡・終了条件

- 既存証跡を先に照会し、対象commit、環境、医院/fixture、操作、期待、実結果が一致する部分だけ再利用する。checkoutにreportがないことだけで未実施としない。
- 実行する場合は対象環境/clinic・fixture owner・現在grant・事前件数・期待事後件数・失敗注入方法・cleanup/復旧先・実施者と別の確認者を先に固定する。production/未承認共有医院/実患者でmutating testしない。実LINE・外部送信・実機・DB/migration・provider操作は別承認の人間レーン。
- 成功、拒否、競合、結果不明、未検出を分離する。commit後応答断は元operation照会で確認し、未検出を失敗確定として新keyのPOSTへ置き換えない。再照会にも現在認可を適用し、secretや原患者データをreceiptへ転載しない。
- `reports/uat-YYYY-MM-DD/` にcase/元AC/操作と版・期待/実結果・receipt/必要監査・件数/金額・復旧cleanup・PASS/FAIL/PARTIAL/BLOCKED/N/A理由を保存する。新しい実行結果は元reportに保存し、本表やscenario本文には対象版・範囲付きの所在と未検証部分だけを接続する。静的読取と実機合格を分離する。
- 未確認の臨床値、運用窓、backup方式/保持、実送信許可、SLO値は推測せず担当者へ限定確認。不足はBLOCKED。条件付き7項目を今回の必須試験へ増設しない。

## 項目別の対象・操作・期待

全31行について**今回の追加実行は0**。これは過去実施の不存在やユーザー未回答を意味しない。既存の部分PASS・限定観測・静的照合を下記E0–E13と31行の対応で区別する。最新ケース全体と対象版が一致する実行合格は未証明のため、各行の未検証範囲を残す。受入の正本は「接続先」である。

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


## 既存証跡の対象版・成立範囲

以下は既存report/記録の読取結果で、新しい実行receiptではない。`reports/` とrepo外ファイルは保管者へ照会する非配布locatorであり、GitHubから直接取得できるリンクと混同しない。SHA未記載は補完しない。限定PASSはその実行対象だけに適用する。E0は機械CIの報告、E8/E9は静的調査を含み、E12/E13は手順/状態記載である。

| 根拠 | 元記録・対象版/時点 | 確認できた範囲 | 最新ケースへ転用できない範囲 |
|:--|:--|:--|:--|
| E0 | [PR #698 作者返信](https://github.com/MinoruSoga/AnimalEkarte/pull/698#discussion_r4310058229) / 同PR `r4310063863`<br>47f99fbfb7194d319754f00eb99240f8b4e55e99; subsequent docs-only head563353287 | 2026-10-19作者/ユーザー報告：make ci30段階成功。今回独立再実行/生artifact取得ではない。 | 機械CIの報告。31個別ケースの実機/実DB競合/現grant/印刷/医師承認には代用不可 |
| E1 | `reports/prod-cutover-20261004/EMR-263.md`<br>2026-10-04 STG PlanetScale/PostgreSQL18.6; 対象app SHA本文未記載 | backup取得/隔離復元/件数照合PARTIAL PASS、133ファイルSHA一覧digest記録。managed backup成功の過去観測。 | 代替per-table COPYは非atomic、単一artifact restore/アプリsmoke未実施。本番method/owner/保護先/retention/receipt未承認、現状再照会なし |
| E2 | `reports/prod-cutover-20261004/EMR-129.md`<br>2026-10-04 11:10UTC STG; 4本REHEARSAL_ONLY manifestを指定(21表) | 投入manifest対STG件数、差分説明、seed照合PASS/GOの範囲を記録 | 最新formal producer/candidate22/payment正件数/全4院本番bundleへの合格ではない。manifestとrun版の違いを保持 |
| E3 | `docs/work/emr/emr-105-vaccine-multi-20261002.md`<br>2026-10-02記録。単独経路は2026-09-23 S29。実行commit SHA本文未記載 | 同日複数接種：実DBとVitest/hook/実フォーム限定PASS、単独経路既存実保存証跡への参照 | 種類別次回日設定/当時snapshot/全caller/実ブラウザ3件連続は未証明。次回日項目は補助field |
| E4 | `reports/prod-cutover-20261004/EMR-272.md`<br>2026-10-04 emr-272-cutover; 検証commit SHA本文未記載 | STG read-onlyでコース未登録を同定。トリミングVitest11files/71tests PASS等 | 新しい開始終了/再送receipt/料金D10 pull/全入口の最新実機には未適合証明 |
| E5 | `reports/prod-cutover-20261004/EMR-289.md`<br>7ff2a7e272305c44438fc5fba3c010e328e934c7、2026-10-07再実行ログ | スタッフ/アカウント無効化same-tx関連staff/reservation/一部lintscanテストPASS | 現在grant全経路/医院cache失効競合/全認証API/診療・会計新gateの検証ではない |
| E6 | `reports/prod-cutover-20261004/EMR-111.md`<br>2026-10-04 STG監査履歴。app SHA未記載、合成demoアカウント限定 | 医院別過去認証の監査記録を確認 | 現在ログインsmoke/正式名簿全員/現grantの証明はUNKNOWN。履歴の存在を現在PASSへ変換しない |
| E7 | `reports/prod-cutover-20261004/EMR-290.md`<br>2026-10-05 emr-290-cutover; 実行commit SHA本文未記載 | 死亡status OR deceased_atによる除外へ修正した報告、関連限定テスト実行記録 | LIFF新公開allowlist/訂正削除公開/単回link結果不明/全医院実機を証明しない |
| E8 | `docs/spec/lab-device-integration/README.md:56`<br>1ae07d6885863c4ab82b526f2c33acb67a6cc327 | PutWaitの生存/医院判定と当日カルテserver条件不足を限定静的照合済み | 実機PASSではない。現headで不足の修正有無/実APIの競合実行を今回試していない |
| E9 | 保護Vault `55_Handoff/BRT-39_ops.md` / [Issue #261](https://github.com/MinoruSoga/AnimalEkarte/issues/261) B/C<br>2026-08-14 main697d5c597の静的ログ。GH261最新公開status2026-09-23 | 薬量hard rejectの既存コード調査/健診provisional seed・実装テストの存在を記録 | 具体薬量bundleの医師承認や健診全段階復帰のruntimeは証明しない。10月方針を旧臨床責任者役職で覆さない |
| E10 | `reports/prod-cutover-20261004/EMR-46.md`<br>2026-10-04 STG read-only; app SHA未記載 | 締め時刻設定の保管/取得とデモ残留値を調査済み | 医院の実営業時間承認/本番投入/新設定CASや元帳票不変の実証ではない |
| E11 | `reports/prod-cutover-20261004/EMR-146.md`<br>9/22〜10/3観察。backend6191bdc14→02744ebfe等複数版、frontend c03657745まで | STG稼働の履歴観察を実施。2026-10-07訂正でBLOCKED/UNKNOWN | 単一revision5営業日受入ではない。管理loginを現場の業務稼働へ昇格しない |
| E12 | `docs/ops/testing/scenarios/S33-receipt-print-pdf.md:48`<br>手順書、実行ref/実出力reportは今回未取得 | プレビュー→window.print→PDFという既存試験範囲を定義 | これは実行証跡ではない。物理印刷/医院承認/最新全帳票のPASSは未確認 |
| E13 | `docs/ops/infra/CLOUDFLARE-EXTERNAL-INTEGRATIONS-AUDIT.md; docs/ops/agent-harness.md`<br>現在docsの状態宣言、個別運用run refは未取得 | 外部連携code側/入力parser等の静的整備を記載 | Console有効化/送信受信/cron・実hookロード/OS実効権限の実行証明ではない |

### 31項目への対応と残る検証

同じEを複数項目へ参照しても、個別ケースPASSを増やしたことにはならない。「該当証跡未取得」はこの探索範囲での状態であり、未実施・未回答の断定ではない。本文に「結果はない」「未実施」とある元調査の残件も、下表の対象版・範囲に限定して読む。

| ID | 既存根拠 | 今回なお証明できていない範囲 |
|:--|:--|:--|
| CL-R01 | E0 | 論理契約と設計案はあるが旧client/GetOrCreate/batch/機器経路を含む実API/全writer参加・lock順/必要監査/結果照会の適合は未確認。 |
| CL-R04 | E0 | 保存単位/対象世代/receiptは詳細提案まで具体化済み。予約pet変更・受付と回答競合、原文別保存、部分保存/復元の実API・旧値移行適合は未検証。 |
| CL-R05 | E0 | 通常終了後編集/元版監査/目的別訂正は既決。旧draft/finalized・旧revision/item対応・期待版なしDTO/監査best-effortと目標の適合、restore等の専用経路は未確認。 |
| CL-R06 | E9 | 犬猫の単純体重計算・目的経路パターン選択の採用は回答済み。具体値bundleの医師承認と適用版の対応、上限/単位/拒否の実行証拠は未確認。旧「臨床責任者」指定を再要求しない。 |
| CL-R08 | E4 | 途中保存/担当コース未定可・実担当/施術で終了等は既決。全入口wire/再表示、旧必須UI、共通revision・永続receipt、料金差分と会計pullの実適合は未確認。 |
| CL-R09 | E9 | 独立作成の親→健診→field段階の部分成功・値復元UI、package/field競合・期待版/全CRUD監査・旧snapshotは未確認。 |
| CL-R11 | E3 | 現行1year共通定数の観察は種類別設定/手入力目標の充足ではない。歴史的次回日/旧snapshot・全client版receipt/公開通知条件は未検証。 |
| CL-R13 | E8 | 文書はPutWaitが当日対象カルテ条件をserver検証していないと固定旧headに基づき明示済み。文書過大表現は解消済みで、最新コードでの残存/修正有無は本snapshot監査で判定しない。 |
| CL-R14 | E8 | L01/02文書断定は訂正済み。実切替/撤去日、NX/AU受入、PU4010パリティ差、IDEXX実機忠実度/医院PIMS送信運用、医院実grant/基準値承認は未検証。 |
| FIN-R12 | E0 | writer群は列挙されているが全route/service/import/repair/GET初期化/専用記録起点の実装適合証拠はない。会計権限を医師確認、旧route許可を新業務へ転用できない。 |
| FIN-R13 | E2 | E2は旧rehearsalの件数照合。最新の旧単一患者completed・method ID欠落・元資金不明・税/帳票snapshot/保険値を分類した結果は未確認。FIN001採否は再質問せず移行照合を進める。 |
| FIN-R16 | E0 | 独立フォーム明細対応は修正済み。E1全入口明細統一・D10算定・提示版保護・親確定後後継編集・確認付き未確定会計転記は、それだけで適合確認済みにならない。 |
| FIN-R17 | E2 | 請求completed総額・親完了日と実移動台帳は別基準。イベントidentity・訂正・遅着・非金銭充当・別方法返金・新旧系列切替の実consumer適合は未証明。 |
| FIN-R22 | E0 | I1は承認済み。旧絶対数量PATCH・通常治療create減算・名称同期・利用中削除と、理由付き実物台帳/版の実装適合・旧単位照合は未検証。 |
| FIN-R23 | E0 | E0のCI成功報告は確認済み。最新DC-A01〜06/S21の実DB二セッション、全intent同key/別内容/遅延commit/監査失敗の実行証跡は未確認。 |
| FIN-R24 | E12 | S33はA4明細兼領収書の既存プレビュー→PDF手順。最新の請求と実受領分離、預り受領証、訂正/返金/再発行履歴、月次未納調整、日次退院4項目を全て証明しない。物理印刷・実締め・医院運用承認は外部レーン。 |
| FIN-R25 | 該当実行証跡未取得 | 品目税務分類・適格請求書全体・保険扱い・決済規約／外部会計適合は製品算定ルールの承認から保証できない。今回原資料・実決済を検証していない。 |
| IR-08 | E7 | 保存済接種/最終来院・死亡OR除外・訂正削除公開とlink結果照会の実装/実機は未確認。2026-10-15静的読取を最新mainの適合証明にできない。 |
| IR-13 | E0 | read完全性/4分類/typed ref/prepare-read-start/end/Undo・医師確認別軸・browser overlay/遅延応答・要望readの実装実機は未検証。設計は既に具体化。 |
| IR-17 | E5, E6 | 静的台帳/seed/5分cache説明から現医院grant・API返却分離/全route認可・失効競合成立は証明できない。scope=all説明修正は完了。 |
| IR-18 | E10 | 現権限・設定writeの競合/結果不明・参照削除guard・医院/法人別値・過去帳票不変の実証未完。新医院create receipt対応も仮定不可。 |
| IR-20 | 該当実行証跡未取得 | S14/S23の目標分離・T4/T5追記は既に文書反映済み。全医院初期値/選択保持/AND結果件数/非第一医院作成帰属/入力者は今回実機未実行。 |
| REM-OPS-02 | E6, E1 | 未記入checkboxは外部状態未確認を意味する。構築未完と断定せずHOLDの解除receiptが不足。 |
| REM-OPS-04 | E1 | E1の過去STG隔離復元PARTIAL PASSは確認済み。現行で承認された単一artifactの復元、アプリ互換/所要時間、CF-only rollbackとcleanupは同一runで未確認。AWS切戻しを復活させない。 |
| REM-OPS-05 | E10, E11 | 10/20構築開始と切替日を分離済み。新window・判断者・スタッフ権限・締め時間・標準5予約区分の実適用receiptは未確認。 |
| REM-OPS-06 | E2 | 旧2院receiptは最新candidateでなく、KNJO/payment正件数・全4院bundle・source/consumerpin・6seed/target未確認。最新外部状態未照会。 |
| REM-OPS-08 | E2, E1 | candidate22単一transactionはlegacy21表別commitの実績で証明できない。COMMIT_OUTCOME_UNKNOWN/STARTED等をrollback済みとしない。 |
| REM-OPS-10 | E0, E11 | 過去V scenario-level PASSや機械check成功は全fieldのF実行証明でない。ignored reports不在を未実行とも断定しない。 |
| REM-OPS-11 | E13 | LINE allowlist/redelivery/errors・LSTEP二重gate/vendor・SMTP・LIFF現DNSはUNVERIFIED/RELEASE PENDING。採用除外のLSTEP実送信を納品の機能採用へ変えない。 |
| REM-OPS-13 | E13 | 入力parserやhookfixtureの静的修正は済み。実Claude設定合成、実効拒否、timeout/無効化、rawdevice/mount権限はUNKNOWNとして残る。 |
| REM-OPS-14 | 該当実行証跡未取得 | 測定テンプレ未記入は全世界の未実行証拠ではない。別保管receipt未照会。業務SLO採用値空欄と計測値を区別。 |


旧DB backup4件の追加資料調査は本人の対象外判断を維持し、上記REM-OPS-04等の将来の運用受入とは分離する。運用方式・作業窓・SLO採用値と、それを実測/照合する仕事は別であり、この表を「ユーザーへの未回答質問一覧」にしない。回答済み事項と調査先行の分担は[残件台帳](REMAINING-RESPONSE-20261019.md#責任者へ確認する事項)を参照。
