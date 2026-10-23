# 見積・締め・在庫・分析の技術契約と現行差

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | design.finance-operations.technical-spec／技術整理 |
| 決定状態 | 業務決定は各spec正本。現行読取根拠と設計要求を区別し、API／DB変更を承認しない。 |
| 実装状態 | 現行差を分離。今回実装変更・網羅監査なし |
| 実機検証状態 | 未実施。受入条件は検査定義 |
| 文書確認日 | 2026-10-20 |

## 対象と参照

D11 [見積](../../../spec/domains/estimates/business-rules.md)、D12 [締め](../../../spec/domains/cash-register/business-rules.md)、D13 [在庫](../../../spec/domains/inventory/business-rules.md)、D16 [分析](../../../spec/domains/analytics/business-rules.md)の保存・読取境界。API wireは[api.yaml](../../../../backend/docs/api.yaml)、DBはmigrationsが正本。現行API／DBと、本書で選定する目標の意味型・操作結果を区別する。物理名は現行実装の存在を主張しない。実装・設定・移行・実機・外部送信は今回未実施。

<a id="estimate-current-gap"></a>
## D11 見積：現行値と承認済み目標の差

| 事実／根拠 | 現行契約 | 目標／検証対象 |
|:---|:---|:---|
| [estimate_service.go](../../../../backend/internal/billing/estimate_service.go) | 明細の数量×単価丸め、行割引、税とヘッダ割引を計算。独立ヘッダ値・insurance_amountと明細計算は別扱い | 全入口の明細化、D10の明細→全体・税率別税・端数・除外・概算保険表示へ適合するか検査。既存算式の流用を適合済みとしない |
| [estimate_service_tx.go](../../../../backend/internal/billing/estimate_service_tx.go) | 親カルテdraft保護、関連FKと医院所属検査、採番・作成・明細置換・再取得をtransactionで行う | 原見積snapshotと新計算根拠の保存を分離。既存ヘッダのみの履歴を無断で再計算・明細化しない |
| 同update path | LockEditableByID後にitemsを再取得し、明細省略でも保存明細があればヘッダ再計算。items空配列は全行置換／0合計。省略時と同じでない | 明細方式の旧値移行、競合、部分更新を試験。discount権限がない操作で権限対象値を黙って初期化しない |
| [estimate_service_successor.go](../../../../backend/internal/billing/estimate_service_successor.go)・[tx](../../../../backend/internal/billing/estimate_service_successor_tx.go) | 承認／却下のみ後継作成。理由trim後必須・500文字以内、所属検査、新番号、元参照と明細コピー、監査失敗rollback | 下記選定契約の操作ID／receiptで同じ後継へ収束。原番号保持と新番号付与を再発行と混同しない |
| 親カルテ確定と後継 | 後継createは親draftロックを例外的に省略する一方、作った後継の通常updateは親draftを要求 | 旧親draft更新gateは現行差。未提示下書きは親確定だけで通常編集を禁止せず、見積自身の状態・版・関係・権限で検査する。後継は下記専用条件で編集し、提示原本保護を維持。通常削除は[2026-10-23本人判断](../../../spec/domains/estimates/business-rules.md#estimates-delete-parent-state-20261023)に従い親未確定条件だけを外し、見積自身の削除条件/履歴を保持する |
| 通常CRUD監査 | best-effort。後継のみfail-closed | 現行保証の差を維持し、下記目標では操作変更と必要監査／receiptを原子的に保存 |
| 会計受渡し | 見積serviceのCRUD／successorだけではD10新契約の確認付き転記を証明できない | 元見積ID／版・選択行・飼主／ペット・価格税割引根拠・概算保険の再確認、先draft会計ID／結果を対応させる。下記選定契約で受渡しの必須意味・原子保存・結果照会を確定。物理wire名は実装選択 |

検査群：確定と明細置換の競合、別医院参照、discount権限、0行／省略、後継監査失敗、親確定後の専用訂正、再送による重複転記、期限切れ後継、既存ヘッダだけの履歴保全。FR/ACは業務正本に一本化する。

<a id="closing-consumer-gap"></a>
## D12 締め：D10新イベントのconsumer

2026-10-23選定の[A1/A2限定契約](cash-closing-contract.md#cash-closing-public-wire)が、有限wire/C06/source完全性・旧締め履歴・遅着/cutoffの詳細を固定する。以下の旧候補field名は当該範囲の新wireを再定義しない。非R1 producer、opening writer、B grant等は未選定/未供給を分け、共有供給/実装/試験/activationは未完。

旧cash-register.mdはcompleted_at／completed状態の会計総額を集計し、前回繰越・入金／支出項目なし。旧payment_splits／refundsの読取は[画面29](../../../spec/screens/29-closing-aggregation.md#closing-allocation-contract)に契約がある。これをD10の部分入金・後日回収・預り金受領／充当／復元・別方法返金にそのまま流用しない。

### adapterが必要な意味

| D10から受ける意味 | D12の処理 | 禁止される読み替え |
|:---|:---|:---|
| 実受領／実返金ID、元資金参照、実日時、医院、方法、実額 | 実日時を営業日／区分へ解決して1回集計 | 請求完了日時へ付け替え、請求総額を実受領額にする |
| 未納回収と元請求別配分 | 受領イベントは1回、配分は複数元請求を参照 | 請求数分だけ受領を複製、過去売上を再生成 |
| 預り金受領／充当／解除・復元 | 受領／実返金だけ移動、充当等は金銭移動0 | 同じ資金を受領と充当の両方で加算 |
| 請求差分・取消 | 元締めへの訂正差分と監査を保持 | 自動現金返金・締め再開 |
| 返金元資金内訳と実返金方法 | 元資金の整合を検査し実際の返金方法を集計 | 元カード決済だから現金返金をカード減額へ計上 |
| C1のレジ補充／回収・開始残 | 承認済み目標として現金実査の別内訳 | 売上明細や飼主入金へ変換 |

ID／版／source参照・再送・最終再集計・snapshot・後日差分・権限／監査は下記選定契約で確定する。D10をsource-of-truthとし、物理schemaの実装と適合試験は未実施。

実査誤入力の訂正は2026-10-15追加確認8の目標で、D12がwrite owner。現行Create／List／Get／Previewと会計起因adjustmentだけでは専用訂正を満たさない。原closeのUPDATE/DELETE禁止は維持し、下記専用intentの追記として実現する。

### 既存保証を保持する範囲

- cash_register_closes はclinic×date×periodの完全UNIQUE、UPDATE／DELETE拒否のappend-only。訂正もappend-onlyで元close参照を保持する。
- 会計編集・締め差分・必要監査は同transaction fail-closed。accounting_deltaとcash_movement_amountを混同しない。
- AM／PM／EMGは半開区間。既存snapshotは設定変更で再計算しない。営業時間例・#252裁定・DB既定・環境実値は別。
- 旧『同日のいずれか区分が締まれば日単位で会計保護』と、越日EMGの実日付／営業日は検査対象。日付ラベル一致だけで締め後編集判定を流用しない。
- matrixは返金発生日、旧KPIは親会計完了日。旧KPIの理論現金控除を新実査の唯一の基準にしない。

検査群：境界ぴったり／越日、部分入金、後日回収、現金釣銭、預り受領→充当→取消復元→実返金、非現金→現金返金、0円会計、source重複／欠落、同時締め、監査失敗、締め後記録、画面と印刷／履歴の対象差。

<a id="inventory-current-gap"></a>
## D13 在庫：現在値編集と実績台帳を分ける

| 根拠 | 現行で静的確認した事実 | 目標／残検査 |
|:---|:---|:---|
| [model/inventory.go](../../../../backend/internal/model/inventory.go) | quantity・minimumはint、数量0以下でout_of_stock、正数量で閾値以下ならlow、他はsufficient | 旧『min0なら常に十分』を是正。0／閾値ぴったりのAPI／UI整合 |
| [inventory_service.go](../../../../backend/internal/inventory/inventory_service.go) | CRUD、現在庫の絶対値PATCH、statusは入力から書かない | I1の理由付き実績・前後差分・版検査は未実装／未確認。下記選定契約で属性編集から数量writeを分離 |
| [repository.go](../../../../backend/internal/inventory/repository.go) | updateと再取得をtx内実行。DecreaseStockは正整数限定・clinic／id／active／quantity>=qtyを同UPDATEで検査 | 負数在庫と小数切捨てを防ぐ既存保証。絶対値編集と同時出庫のlost-update防止を確認 |
| 同DeleteIfUnused | treatments／vaccines／medicinesのactive参照なしを同一削除条件で要求 | 使用中削除の競合を検査。FE削除UI未実装を操作可能と書かない |
| 同薬剤補助経路 | DeleteByNameAndMedicineCategory／UpdateNameByMedicineCategoryは医院＋同名medicineで検索し、IDによる一般DeleteIfUnusedと同じ経路ではない | 同名品目・rename・削除時の依存保持／履歴に差がないかD03と照合。薬剤マスタ無効化を在庫削除に置換しない |
| 通常治療create | inventory ID指定時減算の旧実装副作用がD03 writer表で確認済み | [最新の実施／実引渡し時の明示出庫](../../../spec/domains/inventory/business-rules.md#inventory-outflow-confirmation-20261020)へ追従する差分。保存だけの減算を目標にせず、過去入力・転記・訂正へ無条件流用しない。臨床数量と管理単位の対応を検査 |
| 物販master | 在庫参照なし | 名称一致による自動連動を追加しない。将来の参照設計は別決定 |

I1に対応する実績入力意味：実績種別、対象医院・品目・管理単位、元関連記録、前版、実日時、理由、担当、操作identity、前後数。下記選定契約で原子性・再送・版・旧値停止を確定する。物理table／endpointは実装選択、保持はC05へ従う。

検査群：同時減算、棚卸と出庫、二重返却、医薬品名同期、使用中削除競合、正整数／小数／不足、参照不明、他院ID、取得失敗、薬剤masterと在庫masterの権限境界。

<a id="analytics-current-gap"></a>
## D16 集計：各メトリックの実際の取得根拠

| 根拠 | 現行の計数 | 対応時の注意 |
|:---|:---|:---|
| [owner/ltv_repository_query.go](../../../../backend/internal/owner/ltv_repository_query.go) | completed billingsのみ。期間式はCOALESCE(bmr.date,b.scheduled_date)、paidは親請求のpayments合計、netは親請求のrefunds控除 | completed_at期間・実入金日期間と同義にしない。D10新台帳adapterが必要 |
| 同visit集計 | pet.owner_idに帰属するDISTINCT mr.date、annualはNOW−365days。status確定条件はこのSQLにない | 旧値は未削除カルテのある日数と明示。承認6の実来院日集計へは下記source契約で切り替え、旧SQLを適合済みとしない。D03の登録日／実診療日、転記と新来院を区別 |
| 同総額／期間額 | total_amountは累計請求額、annual_amountは選択期間・amount_basis式 | amount_basisを変えたから全表示列が一律同じ基準になると推定しない |
| 同絞込 | include_zeroは売上軸中心、include_no_visitは最終来院軸、明示from/to→year→preset | 指標ごとの母集団・条件順を保持し、画面ラベルと実際のSQLを照合 |
| [lstep_analytics_service.go](../../../../backend/internal/lstep/lstep_analytics_service.go) | JST月、来院判定daysは1〜365、母数0なら率0 | 期間不正・失敗と0を区別。UIの30日固定とAPIの任意値は別 |
| [lstep_delivery_trigger_log_repository.go](../../../../backend/internal/lstep/lstep_delivery_trigger_log_repository.go) | 配信状態件数はscheduled_at、来院率母数はfired_at／fired／非抑止のログ、来院はカルテ日をJST日付両端包含で照合 | ログ件数をユニーク飼主数や外部最終到達数へ読み替えない。実来院sourceへの適合は別で、旧カルテ日観測を実来院確定にしない |
| [LstepAnalyticsPage](../../../../frontend/src/features/lstep/routes/LstepAnalyticsPage.tsx) | 月選択、配信統計、30日来院率、CSV取込部品を併置 | CSV取込は別write責務D15。分析viewから許可を派生しない |
| 月次画面32／締め画面29 | 暦日と営業日、matrixの返金発生日と旧KPI親会計帰属が異なる | 同一金額／同一日と偽装せず、出力に期間・基準を保存。EMGを含む件数の網羅を検査 |

D10連携は下記D16選定adapter契約に従う。必須検査：未入金請求・部分入金・後日回収・取消・再請求・別方法返金・預り金受領／充当／復元。請求額／回収額／実移動のsource参照を保ち、旧集計結果を適合済みとして移行しない。


<a id="analytics-actual-visit-source"></a>
### D16：承認6の実来院sourceと旧値の区別

主たるwrite ownerはD01の受付、D03等の臨床記録で、D16はその確定結果を読む。通常は[D01受付の同時成功・結果照会](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-checkin-record-atomicity)と[目的別操作／履歴契約](../reservations/technical-spec.md)の医院・予約／患者・実受付結果・日付／履歴参照を照合する。予定日時や現status一覧の単純抽出を実来院sourceにしない。

現head da38895d21d60085b3190c785660a2867b81be26の[reservation_service_update.go](../../../../backend/internal/reservation/reservation_service_update.go)はchecked_in遷移でchecked_in_atを更新し、再受付時には最後の遷移時刻で上書きする。したがってchecked_in_at単独では元の実来院日の履歴を復元できず、UpdatedAtや予約予定日への置換も不可。確定受付結果／履歴・訂正関係を参照し、上書き済みで由来不明なら旧データの照合要として残す。新しい物理イベントtable／endpointが存在するとの主張ではない。

過去診療は[D03専用入力](../../../spec/domains/medical-records/lifecycle.md#requirement-fr-medical-records-retrospective-no-card-no-billing)の由来・患者・医院・実診療日と、当該医院への実来院事実を確認できる場合だけ入力候補となる。既存の医師確認を来院確認と無条件に同一視せず、外部資料・転記・単なる日付入力は除外する。すべての通常来院へ新しい医師承認を要求しない。他領域の入院／宿泊の日次記録だけから毎日の新規来院を作らない。

読取adapterはsource owner／ID／版、医院・確認済み対象飼主／ペット、実来院日（JST暦日）、根拠となる受付／過去入力・訂正参照、照合状態を保持する。同じsourceを一度だけ解決したうえで、飼主回数は医院＋飼主＋日、ペット回数は医院＋ペット＋日でdistinct。同日の2ペットを飼主2回にせず、1ペットを複数カルテ・再受付の数で増やさない。訂正は元sourceとの関係をたどって有効事実へ投影し、元・転記先を新規来院として二重計上しない。飼主帰属はD04の確認・履歴契約と照合し、現在のpet.owner_idだけで過去を移す方針を追加しない。

sourceが不足・矛盾・取得失敗なら実来院の値を未確定として返す。必要な分類が不明の飼主を母集団から無断除外せず、不完全応答として示す。旧記録日数を新値に代入したり、不明を0回／来院なしに変換したりしない。期間内・累計・rolling365日、同指標からの最終来院日、画面／CSV／CPM入力と来院観測は同じsource基準を識別する。CPM版・閾値・配信ログ母数・観測窓や通知送信許可をこのadapterで変更しない。FR-analytics-owner-counts／FR-analytics-lstep-observationとAC-analytics-same-day-pets／record-days-not-visits／actual-source／visit-recheckin／retrospective-visit／visit-unknown／visit-consumer-basisへ追跡する。

<a id="finance-selected-operation-contract"></a>
## 選定済みの操作・結果契約（D11／D12／D13／D16）

本節はE1／C1／I1およびD10の合意を実現する今回の技術決定で、上の現行実装表に優先する目標契約。API名・JSON綴り・DB物理名・lock手段は実装選択だが、以下の必須意味・成功条件・停止条件は後決めにしない。既存APIが満たさない場合は互換adapterまたは専用経路が必要で、旧APIの成功を新契約成立の証拠にしない。

### 共通要求・応答・照会

- write要求は操作ID、intent、医院ID、対象資源IDと期待版、入力内容、要求される理由／確認を持つ。actorは認証から確定し、クライアントの自己申告を信用しない。新規資源はID未採番でも操作IDが必須。関連する飼主・ペット・元記録／元明細は型付きIDと医院を検証し、氏名・行順・同日だけで対応させない。
- 操作の一意性は医院＋owner領域＋操作ID。同じkeyの内容digestはintent・対象・入力・確認根拠を含み、期待版は操作別に業務内容とCAS/fenceへ分類する。EMR-483のC06参加ではCAS/fenceだけを除外し、了承before/after計算revisionは含める（[483の限定接続](../accounting/technical-spec.md#price-acknowledgement-handoff-20261020)）。他financial operationの意味を一律変更しない。同key同内容は元の結果へ収束し、別内容は拒否する。新しい業務操作を同じkeyへ混ぜない。重複判定と業務変更／履歴／必要監査／結果receiptを同じtransactionで成立させる。途中失敗で業務変更だけを残さない。
- 成功receiptは操作ID、intent、対象ID、確定した版、新規ID／原本参照、結果状態、保存時刻、actor、保存された数量／金額snapshotと副作用区分を返す。金銭移動なし・在庫移動なしも明示する。receiptは飼主向け領収書とは別。権限失効後の再送では変更を再実行せず、保存済み結果の閲覧も現在のread権限で判定する。
- 結果照会は同医院・領域・操作IDで行い、保存済み成功／確定拒否／処理中／未検出を区別する。照会失敗や未検出は初回失敗の証明ではない。処理中・未検出・unknownは同一keyで結果を再照会し、自動再実行しない。業務失敗が確定して再試行可能な場合だけ、各operation契約と人の再確認に従う。別keyの作り直しや関連一覧0件による再作成を禁止する。サーバーの確定拒否後に業務内容を修正する場合は旧結果との対応を残して新操作にする。483ではCAS前提のみの更新は人の再確認後に同keyを維持し、了承版の変更と分ける。
- 更新は現在版・状態・認可・依存資源版を保存直前に再検証する。競合は変更せず、最新値と差分を再読してスタッフが再確認する。黙った上書き・自動再配賦・勝手な価格再適用をしない。監査には前後値／参照版・理由・actor・実日時と記録日時・操作IDを残す。必要履歴の削除やreceiptの期限切れを重複防止解除として使わず、保持はC05に従う。
- 新契約の操作は各領域の既存操作権限と依存write権限を現在検査する。旧routeの認可だけで後継訂正・会計転記・締め差分・実物調整を開放しない。必要な操作と既存grantの対応が未設定なら拒否する。新職種・全スタッフ許可は追加しない。

<a id="estimate-selected-contract"></a>
### D11：原本／後継／会計転記

| intent・owner | 必須意味と保存前検査 | 原子的な保存と結果 |
|:---|:---|:---|
| 見積新規／下書き編集・D11 | 見積期待版（新規はなし）、飼主／ペット／任意親記録、明細ID・数量単価・税区分・割引／除外snapshot・概算保険、有効期限。通常編集は未提示の編集可能状態・期待版・医院／親子関係・権限を再検査する。旧親draft更新gateは現行差で、目標では親確定だけで未提示下書きを拒否しない。提示済み金額／明細の変更は後継へ分離する。明細省略は保持、明示[]は全行削除とし、下書き0行を請求確定と混同しない | 見積・明細・計算根拠版・履歴／監査・receiptを同時保存。D10算定契約に従うサーバー再計算で返し、クライアント合計を正本にしない |
| 承認／却下・D11 | 見積と全明細の期待版、選択状態、現在の内容。編集と同じ資源境界で直列化 | 状態と固定snapshot・履歴／receiptを同時保存。先に確定した後の通常編集／削除は拒否。外部送信・診療・会計・在庫は生成しない |
| 後継作成・D11 | 元見積ID／版、理由、同医院／飼主／ペット。提示済み・返答待ち原本の金額／明細変更、承認／却下原本の変更、または期限切れ再提示という専用intentを検証 | 新ID／番号・元ID／版・行の由来・後継draft・履歴／監査・receiptを同時保存。原本不変。同一操作の再送は同じ後継ID。別intentの複数後継を同じ原本という理由だけで統合しない |
| 後継専用編集・D11 | 後継ID／期待版・元見積ID／版・後継関係・編集可能状態・変更項目／理由。親確定だけでは拒否せず、後継の未提示状態・権限とこの専用後継条件を検査する。原見積や親記録への変更は要求として拒否 | 後継だけの計算snapshot／明細・変更履歴／必要監査・receiptを同時保存。専用経路でも承認済み後継を通常編集せず、さらに後継へ |
| 確認付き会計転記・D11→D10 | 見積ID／版・選択行ID／由来・飼主／ペット・実際の請求対象確認・価格税割引差分確認・概算保険再確認、あれば元臨床実績の型付きID。期限切れは後継で再提示し、古い版のまま進めない | D10が未確定会計を保存し、先会計ID／版と元見積／行対応・履歴／receiptを同一transactionで確定。D11側の転記対応だけ成功／先draft不在を作らない。請求確定・受領・在庫は0 |
| 削除・D11 | 対象期待版・見積自身の編集可能状態・医院/関連整合・明細／受渡し依存なしと削除権限を現在検査。親draftは要求せず、診療終了/医師確認/会計確定で代替しない | 削除と履歴／receiptを同時保存し、承認済み原本・転記元を失わせない |

後継作成・後継編集・会計転記は別intent／別操作ID／別receiptであり、順に後継ID・保存版・先会計IDを確認する。後継作成成功を残る二操作の成功とみなさない。

転記はD10を受け手write ownerとし、見積の固定版と元明細の現在の請求利用状態を同じ保存境界で照合する。同じ操作は同じ先draftへ収束。別操作でも同じ元実績／選択見積行の由来に有効な転記対応がある場合は重複draftを作らず既存先を提示し、必要な変更はその先の期待版付きD10編集へ戻す。後継のコピーで由来IDを失わせない。取消元からの再請求はD10の専用再請求intentであり、見積転記を重複防止の迂回に使わない。

旧ヘッダのみ見積は読取・旧帳票を維持し、架空明細へ分解しない。新方式への編集／転記にはスタッフが明細・税／割引根拠・宛先を明示して新draftまたは適格な後継へ作り直し、旧IDを参照する。根拠不明なら金額を推測せず転記停止。旧確定原本を現在masterで再計算しない。新規方式の切替後に旧ヘッダwriteへ逃げて制約を回避する経路は閉じる。

<a id="closing-selected-contract"></a>
### D12：イベント受入・現金実査・確定／差分

1. D10イベント入力は不変のイベントID／版、医院、種別（請求差分／実受領／実返金／非金銭配分）、元請求・元資金・必要な元イベント参照、実日時、記録日時、方法、実額と請求差分を区別して持つ。D12は医院＋source owner＋イベントIDを一意に受け入れる。同ID同内容はno-op、同ID別内容は停止してD10へ照会する。訂正は元IDを改変せず新訂正ID＋元参照。複数請求への配分で実受領イベントを複製しない。
2. イベント日時から営業日／区分へ解決した結果と適用時刻設定版を保存する。医院／元資金／実日時／方法／金額の欠落・不整合は隔離して当該集計を未完了にし、0円・現金・今日へ補完しない。非金銭eventは追跡するが現金合計へ加えない。D10の確定済み実移動は通知失敗で削除せず、同IDの再読取で投影を復元する。
3. 開始残入力は医院・営業日／区分・開始残額・根拠（前締め参照またはスタッフ確認）・期待版・操作IDを持つ。同一対象へ一つの有効開始残。前締めの実査額を自動で再加算しない。補充／回収は別の現金移動ID・正額・実日時・理由・actor・操作IDを持ち、符号はintentから決める。これらと履歴／監査／receiptを原子的に保存し、診療売上／飼主入金を作らない。誤記は元参照付き差分訂正で残す。
4. プレビューは医院・営業日／区分・境界設定版、対象source ID／版集合の識別値、開始残版、補充回収版、理論額・内訳を返す。締め要求はこれらの期待値・既存正本の有限かつ0以上という条件に従う実査額・操作ID・既存正本が要求する場合の差異理由を持つ（本節で全締め差異への理由必須化は追加しない）。保存時に同じ区分への受入と締めを直列化し、source完全性と各版を再検査して再集計する。差があれば締めず最新プレビューで再確認させる。取得失敗／欠落sourceは確定不可。
5. 締め成功は医院×営業日×区分の一意な原締めID、source集合／設定／開始残snapshot、実査／理論／差額、履歴／監査／receiptを同時保存。同一操作は同じclose ID、別操作の同一区分締めは既存締めを示して拒否する。原締めのUPDATE／DELETE・通常再締めをしない。
6. 締めと同時のイベントは区分の直列化順により原締めへ含むか、締め後差分へ含むかのいずれか一方にする。遅着・後日訂正は実日時／元請求日を変えず、元close IDと元event IDを参照する差分IDを作る。原snapshotは不変、修正後投影は原本＋有効差分と明示。D10会計訂正に必要な締め差分・監査はD10変更と同一transactionに参加し、参加できない既存経路は適合まで拒否する。受領済み事実の遅着投影は事実自体を取消さず、不完全状態から同ID照合で復旧する。
7. 訂正は元close／差分・期待版・理由・修正対象・操作IDを持ち、差分・履歴／必要監査・receiptが同時成功する。二重訂正・再送は元event／操作の一意性で防ぐ。現金の実移動と方法誤記・請求差分を区別し、元締めを開き直さない。応答不明は共通操作照会とclose ID照合へ進み、新key再締めをしない。

旧completed請求総額から部分入金・預り受領等の実移動を逆算しない。旧締めは旧計算基準の固定履歴として閲覧可能にする。新集計対象の旧データは元支払／返金の実日時・額・方法・帰属の照合が揃ったものだけ移行参照IDを付けて受入れ、未知は『照合要／確定不可』にする。新旧基準の混在を一つの理論現金と表示しない。

<a id="closing-actual-input-correction-contract"></a>
### D12：実査誤入力訂正の専用intent

- 要求は医院・元close ID、現在の有効実査の期待版（原本と有効訂正列の版）、訂正後実査額、trim後非空の理由、操作IDを持つ。actorは認証から確定し、元営業日／区分と実査時点は元closeから取得する。実査額は有限・0以上の既存条件、現在の当該操作認可、医院と対象参照を保存直前に検査する。既存view/createまたはaccounting-post-close-editという名称だけで実査訂正許可を派生しない。具体的grant対応が未設定の経路は拒否し、新職種や全員許可を作らない。
- 同じcloseの訂正と直列化し、最新有効実査額・期待版が一致した場合だけ、元close／直前訂正への参照、訂正前後額、元実査時点・訂正記録日時、理由・actor・操作ID、必要監査とreceiptを原子的に追記する。原closeと既存訂正はUPDATE／DELETEしない。保存失敗は全体rollback、同じ操作ID／同内容は同じ訂正receiptへ収束し、別内容の同keyは拒否する。結果不明は共通照会へ戻す。
- receiptは原close ID、訂正ID／版、前後実査額、採用理論基準／source版、訂正後差額と現金移動0を返す。実査額は絶対値の後継として投影し、複数訂正後の額を合算しない。過不足はその表示が採用する同じ理論基準との比較で、原snapshotを現在の理論値へ置換しない。会計由来の理論差分がある場合も実査訂正と別の由来・版で表示する。
- 読取は原実査／原差額、訂正列と最新有効実査／差額を区別する。後続区分の開始残に元closeを採用済みでも、自動上書き・再締め・架空補充をしない。既存会計／現金eventや必要な別訂正へ混ぜず、実査誤入力だけで会計ロック解除も行わない。
- FR-closing-actual-input-correctionとAC-closing-actual-input-correction／actual-correction-guard／actual-correction-conflict／actual-correction-retry／actual-correction-separation／actual-correction-historyへ追跡する。専用route・物理schema・grant設定・実機は今回未変更／未検証。

<a id="inventory-selected-contract"></a>
### D13：実物移動と数量版

| intent | 必須意味・現在検査 | 同時保存と拒否条件 |
|:---|:---|:---|
| 品目属性更新 | 品目ID／医院・期待版、名称／カテゴリ／最小管理単位／閾値、現在権限 | 属性・履歴／receipt。数量を暗黙変更しない。利用履歴がある管理単位を換算根拠なしに変更しない |
| 入荷／廃棄／出庫 | 品目ID／医院・期待数量版・正整数量・管理単位・実日時・理由・元関連参照・操作ID | 入荷は加算、廃棄／出庫は減算。現在数量・active・単位を再検査し、不足は全件拒否。前後数量・実績ID・履歴／監査／receiptを数量と同時保存 |
| 棚卸訂正 | 同じ参照・期待版と実査非負整数数量・理由 | 最新数量と期待版一致時のみ、実査−現在の差を実績として保存。古い絶対値で新しい出庫を消さない。差0も照合記録として保持し架空移動は作らない |
| 実物返却 | 同じ参照・期待版、元出庫実績／行、今回返却正整数量、実物確認・理由 | 元出庫の累計実返却量と競合検査し、残未返却量を超えず一度だけ増加。取消だけでは実績を作らない。元出庫不明は自動対応せず確認停止 |
| 実績訂正 | 元実績ID／期待版、訂正理由・前後根拠・差分、品目期待数量版 | 元履歴は保持し補正実績を追記。数量と累計返却制約を再検査し負在庫／過返却を拒否。原実績削除で帳尻を合わせない |
| 品目削除 | 品目ID／期待版・現在削除権限・依存の不存在 | 通常の利用中削除拒否を維持。依存追加との競合も同じ境界で検査し、原実績や参照を失わせない |

旧属性PATCHの絶対数量入力は新目標の数量変更経路として許可しない。既存画面を互換維持する場合も、数量変更は明示した棚卸訂正intent・期待版・理由へ変換して同じ履歴契約を通し、変換根拠がなければ拒否する。

数量変更の直列化単位は医院＋品目ID。複数品目が同一intentに必要なら安定したID順で全対象を検査し、全数量・実績・監査・receiptを一体で保存する。同操作再送は同じ実績ID、別操作でも同じ元出庫の返却済み量を超えない。臨床量の小数を切捨て／四捨五入してこの契約へ押し込まない。最小管理単位との対応を確認できない場合は在庫移動を停止して明示し、臨床量自体を書き換えない。

[2026-10-20の出庫確定回答](../../../spec/domains/inventory/business-rules.md#inventory-outflow-confirmation-20261020)を目標とする。院内投薬の実施時、持帰薬・物販の実引渡し時に担当スタッフが出庫を明示確定し、その在庫移動intentで数量・実績・監査・receiptを同時保存する。カルテ保存・請求確定・準備readyだけでは減算せず、出庫確定が未成立なら出庫成功を返さない。従来の通常治療createと減算の同一transactionは旧実装観察として保持し、保存だけで出庫する新目標や同じ実施／引渡しの二重減算に流用しない。独立保存された臨床記録の過去入力／訂正は別結果であり、同じcreateを呼び直して出庫を再発火させない。会計取消・薬剤master名称変更・飼主変更は在庫移動intentではない。名称同期は対象品目の明示IDと医院・版を照合し、同名だけの一括変更を新目標の識別根拠にしない。

旧数量を架空の入荷履歴へ変換しない。新台帳への開始時点は品目／単位・確認した数量・確認actor／日時・旧参照を持つ基準snapshotとして記録する。旧版がないwriteは無条件上書きせず最新snapshot版取得を要求する。単位／数量／元出庫が不明な返却や移行は照合停止。過去履歴・品目を消して移行完了にしない。

<a id="analytics-selected-contract"></a>
### D16：読取adapter・期間／基準の固定

D16はD10の請求snapshot／差分と実移動event、D12の締めsnapshotを別projectionとして消費する。入力の一意性・同ID別内容拒否・欠落時不完全表示はD12と同じsource契約に従うが、D16の読取からD10/D12へwriteしない。投影結果には医院・対象期間・JST暦日または営業日の別・amount_basis・source世代／集計時点・完了／不完全状態を付ける。更新途中の別世代を一つの確定合計に混ぜず、取得失敗／欠落を0円へ変換しない。

| D16 source | 日付／帰属と投影 | 訂正・重複防止 |
|:---|:---|:---|
| 請求snapshot／請求差分 | 元請求ID／版・元医院／飼主／ペット／明細参照。請求軸の原請求日と、旧顧客指標のCOALESCE期間は別基準として保存 | 取消／訂正の元参照を保持。新番号再請求は旧番号と対応し、旧売上と新請求を無条件加算しない |
| 実受領（当日／過去未納回収） | 実受領ID・元資金・実日時／方法／額。元請求別配分は請求の回収額へ、現金軸は受領一回 | 同IDは一度。配分訂正で元受領額／日を改変しない |
| 預り金受領 | 元飼主・受領ID・実日時／方法／額。実移動へ計上、請求売上へ加えない | 元資金参照を保ち、後の充当を二度目の受領にしない |
| 充当／解除／復元／再配分 | 元資金・元請求・非金銭event ID／版と関連snapshot。残高／回収配分の更新で、現金移動0 | 元参照に対する有効配分の変化を投影し、元実受領日へ新しい現金額を追加しない |
| 実返金 | 返金ID・元受領／元資金内訳・元請求・実返金日時／方法／額。実移動軸は実返金日 | 元の方法と異なっても実方法で計上。請求訂正だけから返金を推測しない |
| D12締め | 元close ID・営業日区分・固定snapshot・差分ID | 元締めと差分を区別し、月次暦日合計と同じ基準だと表示しない |

請求軸は元請求とその版／差分、受領／返金軸は実イベントと実日時を用い、同一eventの複数請求配分を現金回数に重複計上しない。取消後新番号再請求の旧入金再配分は実移動0、預り金の充当／復元も0として区別する。旧顧客金額指標のCOALESCE(カルテ日,請求予定日)と旧記録日distinctは旧基準と明示し、実入金日や実来院回数へ無条件に移し替えない。来院回数の承認済み目標は[実来院source契約](#analytics-actual-visit-source)へ切り替える。新台帳の照合がない旧期間は旧基準と明示し、新実移動集計と同一連続系列に偽装しない。

画面の再読は医院・権限・期間・基準・対象飼主の組を照合し、古い応答は捨てる。0件／権限不足／未集計／失敗を区別し、CSV取込やCPM配信writeの権限を分析readから派生させない。Lstepの既存scheduled_at／fired_at・ログ母数・V1/V2の意味は現行根拠表と業務正本を保持する。

### 受入との対応と実装検査

上記は既存ACと追加承認の実査誤入力訂正FR／ACについて、保存／再送／競合における具体的期待結果を閉じる。D11はlock-race／successor-audit／successor-edit-handoff／handoff-confirm、D12はevent-repeat／adjust-audit／zero-v-error／no-reopen／opening-cash、D13はinsufficient／concurrent-count／adjust-retry／actual-return、D16はold-ar／deposit／midnight-basis／record-days-not-visitsと実来院source節の追加ACへ追跡する。実査誤入力訂正は専用intent節の6 ACへ追跡する。

実装受入では同key同内容／別内容、初回commit後応答消失、照会未検出中の遅延commit、期待版競合、認可失効、監査保存失敗、旧値不明を各intentで試し、本節の成功または停止結果を確認する。新endpoint／DDL／実grant設定・既存writerの改修・移行実行・実機試験は未実施だが、これらを理由に意味契約自体を未決へ戻さない。

<a id="discharge-report-projection-proposal"></a>
## D12退院帳票の行粒度・固定読取の詳細設計案

[実退院JST日・掲載4項目の正本](../../../spec/domains/cash-register/business-rules.md)と[C06提案境界](../../shared/integration-contracts.md#detail-design-proposals-20261018)に接続する。正本が設計事項として残した行粒度は、同名患者の重複排除ではなく**clinic＋入院滞在ID＋有効な退院事実の1行**を提案する。再入院は別滞在、同じ退院の再読取・同操作再送は同じ事実。既存正本で許可される訂正だけをD05の現在有効性・履歴からprojectionし、現在statusだけで架空の退院日時を作らない。本案で退院取消／再開という新writer・業務遷移は採用しない。取消事実を扱うのは許可されたsourceが実際に提供する場合だけで、未提供はUNKNOWNとして照合へ残す。

| owner／読取段階 | 提案契約 |
|:---|:---|
| D05 source | 型付きstay/退院事実ID・各版、実退院日時、入院日、元owner/pet refと当該表示値の版、有効／訂正関係を返す。期間内全件取得の完全性を返し、ページ先頭だけで完成としない |
| D12抽出 | 認可されたclinic＋JST対象日について、実退院instantをJST日へ写像して抽出。予定退院日・請求日・受領日を使わない。滞在/退院identityで一度だけ選択 |
| 表示・順序 | 退院日時・飼主名・ペット名・入院日の4項目。内部identity/版は照合用で、新しい表示列・担当医・金額列を追加しない。順序案は実退院日時→滞在ID→事実IDの安定順 |
| プレビュー／出力 | query（clinic/日付/権限scope）と取得したsource版集合・projection版を固定し、同じ4項目を画面／印刷へ。版が進めば更新ありを返して再取得し、混在版の帳票を成功にしない |
| 訂正後の再集計 | 元退院事実と訂正後有効版を追跡し新projectionへ。以前の締め・出力履歴を上書きせず、D05/D10の原実績・請求・実入金にwriteしない |
| 失敗 | 読取不能／ページ欠落／source関連不明は部分取得を明示して出力停止。確認済み0件は空一覧として出力可。別日／別医院のcacheへfallbackしない |

本adapterはread-onlyで、退院・請求・入金・在庫・通知・証明書発行を起動しない。`backend/internal/medicalrecord/hospitalization_discharge_tx.go`の請求用ケア明細（quantity=1等）は別の金銭projectionであり、退院一覧の行粒度・金額の根拠へ流用しない。既存の締め営業区分や繰越表示を日付抽出から書換えない。

DC-F01：JST 00:00前後、予定日と実退院日が異なる、同名別pet・同日再入院。DC-F02：同じ退院の再送・既存許可済み訂正・日付を跨ぐ訂正で重複／架空行なし。未提供の取消sourceから架空の取消・再開を生成しない。DC-F03：複数ページ・読取失敗・0件・医院切替遅延・preview後版変更。DC-F04：再印刷／再集計でも売上／入金の二重加算・新writeなし。これらは既存の帳票ACの技術検証案で未実行。

## 完了判断と未実行

仕様整理の候補と静的読取根拠であり、全API網羅・DB整合・税務適合・権限実配賦・本番値・make ci・実機受入を実施済みとはしない。API／DB／コード・環境は未変更。各FR→ACはspec正本で追跡し、業務の未承認提案を設計都合で承認済みにしない。


### D03の通常編集／自動保存／復元と財務owner

[D03会議追加合意](../../../spec/domains/medical-records/lifecycle.md#medical-record-agreement-autosave-20261015)のタブ復元結果をD11の原見積／後継、D12の原締め／訂正、D13の実在庫イベントへ共通のrollback命令として配信しない。各ownerの状態・元ID／版・現在権限・必要監査・操作receiptを維持する。カルテの旧finalizedを解除したからといって、見積承認済み原本の通常updateや確定済み会計・締めのUPDATEを許可しない。D16実来院sourceは受付／実来院の保存結果であり、診療終了やカルテ本文の自動保存・復元で来院回数を増減させない。既存writerがカルテ側フラグを機械的に共用する差は実装適合で分離して検査する。

<a id="estimate-price-presentation-selected-contract"></a>
## D11追加合意：提示原本の競合保護と会計単価

[追加業務合意](../../../spec/domains/estimates/business-rules.md#estimate-price-presentation-agreement-20261015)に従う。提示操作は対象見積・全明細の期待版を検証し、提示内容snapshot／提示版と履歴・receiptを同じ保存境界で固定する。通常編集と提示を直列化し、編集が先なら旧版での提示を再確認し、提示が先ならその原本の金額・明細編集を後継へ誘導する。返答は対象提示版を識別して記録する。状態を下書きへ戻す、カルテタブから保存する等の別writerでも原本保護を迂回しない。提示というローカル記録だけでは外部送信成功を主張しない。具体的なwire／保存フィールドは実装設計で適合を確認する。

返答待ち原本からの改訂も既存の理由付き後継・新ID／番号・元版参照・同医院／飼主／ペット・監査・操作receipt契約を使う。原本への回答を別内容の後継へ継承せず、改訂版の再提示とその版への返答を分ける。現行successorが承認／却下だけを許す点、送付済みのupdateを許す点は実装差であり、適合済みとはしない。

D10転記保存時に承認状態・有効期限・提示版・明細内容／数量の一致を再照合する。適格なら提示単価を使い、現在マスタとの差を表示しても自動置換しない。前提が変われば旧確認のまま保存せず再読取・再説明へ戻す。追加処置／内容変更は後継再提示、期限切れは既存の再確認・後継条件に従う。D10の税／割引計算と窓口保険再確認は維持し、単価と各適用条件の出所を分離して保存する。元見積snapshotは不変、転記成功は先draft保存結果で判定し、請求確定・実入金・在庫実績は生成しない。

### 追加AC追跡（D11）

- [AC-estimates-quoted-price](../../../spec/domains/estimates/business-rules.md#requirement-fr-estimates-validity)：提示単価を使い、元版・行参照と単価の出所を保持する
- [AC-estimates-changed-content-explain](../../../spec/domains/estimates/business-rules.md#requirement-fr-estimates-successor)：改めて説明し、提示済み見積の変更は後継を作り再提示する。旧原本は不変
- [AC-estimates-presented-pending-original](../../../spec/domains/estimates/business-rules.md#requirement-fr-estimates-state-protection)：原本を保持して直接変更を拒否し、後継改訂の経路を示す
- [AC-estimates-presented-pending-successor](../../../spec/domains/estimates/business-rules.md#requirement-fr-estimates-successor)：別ID／番号・原版参照を持つ後継を作り再提示。旧承認や返答を新内容へ自動継承しない
- [AC-estimates-presentation-edit-race](../../../spec/domains/estimates/business-rules.md#requirement-fr-estimates-state-protection)：提示した版を固定し、提示後の変更が原本へ混入しない。競合は再読取・再確認する
- [AC-estimates-quoted-price-other-rules](../../../spec/domains/estimates/business-rules.md#requirement-fr-estimates-billing-handoff)：提示単価を保持しつつD10適用計算と保険再確認を維持。単価保証を総額や保険額の無条件固定としない
- [AC-estimates-quoted-price-no-approval](../../../spec/domains/estimates/business-rules.md#requirement-fr-estimates-validity)：承認済み限定の単価保証条件を満たしたと推定しない。原本保護は維持する
- [AC-estimates-presented-successor-retry](../../../spec/domains/estimates/business-rules.md#requirement-fr-estimates-successor)：結果照会／同一操作IDで同じ後継へ収束し、重複後継や外部自動送信を作らない

<a id="closing-segment-difference-selected-contract"></a>
## D12追加合意：区分判定と差額メモ

締め済み保護は実際に締めた区分だけに適用し、午前締めだけで未締めの午後・夜間を保護対象にしない。確定済み会計の訂正条件は維持する。現金差額が残っていても締められるが、非0差額ではメモ必須とし「原因不明・確認中」も許容する。差額合わせの架空入出金は作らない。 [承認根拠](../../../spec/domains/cash-register/business-rules.md#closing-segment-difference-agreement-20261015)。保護判定は対象医院・営業日／区分・その区分で採用した実時間範囲とclose IDへ解決する。単純な同日close有無を全区分の保護判定に使わず、越日夜間の実日時と営業日を取り違えない。会計自体の確定状態による保護と締め区分の保護を独立検査する。過去原締めを後の設定値で再計算しない。

締め保存時の最新理論額・実査額から差額を算出し、非0なら空白だけでないメモを検証する。原因不明／確認中の記載を拒否せず、原因特定や調整入出金を保存の前提にしない。差額0はメモ任意。プレビュー後の変化は再確認対象とし、変化後の差額にメモ条件を適用する。差額・メモ・対象区分・原snapshot・必要履歴／receiptを同じ保存境界で確定し、応答不明では同一結果照会／再送へ収束する。原締め直接更新・取消や架空金銭イベントは生成しない。既存実査訂正の理由・権限・原値保持は継続する。

### 追加AC追跡（D12）

- [AC-closing-closed-segment-only](../../../spec/domains/cash-register/business-rules.md#requirement-fr-closing-period-boundary)：未締め区分を締め済み扱いにしない。午前原締め保護は維持
- [AC-closing-closed-segment-finalized-invoice](../../../spec/domains/cash-register/business-rules.md#requirement-fr-closing-period-boundary)：区分が未締めでも会計自体の確定後訂正条件を維持する
- [AC-closing-cash-difference-note](../../../spec/domains/cash-register/business-rules.md#requirement-fr-closing-cash-reconciliation)：不足100円とメモを保持して締められる。架空入出金は生成しない
- [AC-closing-cash-difference-note-empty](../../../spec/domains/cash-register/business-rules.md#requirement-fr-closing-cash-reconciliation)：保存を拒否しメモ入力を求める。原因判明までは要求しない
- [AC-closing-cash-difference-zero-note](../../../spec/domains/cash-register/business-rules.md#requirement-fr-closing-cash-reconciliation)：他の締め条件を満たせば、差額メモ不足を理由に拒否しない
- [AC-closing-cash-difference-repreview](../../../spec/domains/cash-register/business-rules.md#requirement-fr-closing-cash-reconciliation)：再集計と差額再確認を行い、保存時の非0差額にメモを必須とする


<a id="analytics-scope-exclusion-trace"></a>
## D16の提供範囲・月次出力の制約

D16のread adapterがD12締めsnapshotを受け取れることと、月次帳票へ全フィールドを出せることを分ける。下記は[既存分析正本](../../../spec/domains/analytics/business-rules.md)のconsumer制約で、集計指標や画面を新設しない。

| 正本AC | 既存契約への技術接続・確認点 | 業務正本 |
|:---|:---|:---|
| AC-analytics-current-scope | 既存の今期提供対象である月次レポートと、対象外のランキング／犬種別画面・CSV、CPM人数／一覧を出力先として区別する。共通read/export adapterの存在や旧対応済み報告から、対象外の公開を許可しない。 | [正本](../../../spec/domains/analytics/business-rules.md) |
| AC-analytics-monthly-no-cash-difference | 月次画面／PDF／CSVのprojectionから実現金・差額を除外し、締め／実査訂正の差額を売上補正へ使わない。締め・履歴・日次締め帳票での区別表示は保持する。元締め／snapshotや実金銭eventへwriteしない。 | [正本](../../../spec/domains/analytics/business-rules.md) |

各行は既存ACの接続先を明確にするもので、新しい業務要件・実装完了・実機合格を表さない。具体的な既存wire／取得元との適合は実装工程で照合する。

<a id="finance-operations-candidates-20261019"></a>
## 財務運用の保存資源・投影候補（2026-10-19・内部提案）

以下はFIN-R16〜R19・R22の既存意味契約を実装へ接続する候補。実在table/APIや採用済み物理設計と扱わない。列名、wire、索引、transaction実装、実grant・データ適合は未確認。全writeは既存共通gate/取得順、操作key同内容収束、現在認可、期待版、監査/receipt同時保存へ参加する。新業務要件、実送信、DB変更、実機合格を追加しない。

### D11・FIN-R16：提示原本・後継・会計転記

| 候補資源 | キーと制約 | 操作・結果/失敗 |
|:--|:--|:--|
| `estimate_presentation` | `(clinic_id,estimate_id,presentation_version)`、全明細版集合/内容digest・単価・数量・税/割引根拠・有効期限・表示宛先snapshot | 提示と編集が同じ見積資源で直列化。提示先行なら金額明細は後継、編集先行なら旧版提示を止める。提示記録は外部送信成功ではない |
| `estimate_reply` | 対象提示版・返答状態・既存actor/time、元見積ref | 旧提示への承認を後継へコピーしない。提示snapshotと返答を対応させ、旧原本不変 |
| `estimate_successor_link` | 元ID/提示版・後継ID/番号・理由・操作ref。操作key一意で同key→同後継 | 別intentの後継を元IDだけで統合しない。後継作成/編集/再提示/転記の結果を別receiptへ |
| `estimate_handoff` | 提示版・選択行ID/元由来・受け手会計ID/行ID・各版・操作ref。実績がある行はtyped元実績IDも保持 | D10が先draftと対応を同時保存。元実績の有効請求帰属制約を後継コピーで失わず、同keyまたは同じ有効転記の別keyで重複draftなし |

転記previewは現提示版・承認・期限・内容/数量一致・今回保険再確認・選択元明細の未請求性を返す。commitはこれらをD10側で再検査し、適格なら提示単価、税/割引はD10契約のsnapshotへ。承認後の数量変更・期限切れは古いpreviewで転記せず後継再提示へ。旧headerだけの見積を架空行へ自動分解しない。検査は提示/編集、返答/後継、転記/別請求、転記成功応答断を両順で実施し、親カルテ確定だけで未提示下書きを一律拒否しない。

### D12・FIN-R17：実移動イベント受入と区分締め

[A1/A2の後発選定](cash-closing-contract.md#cash-closing-source-completeness)を当該有限範囲の詳細正本とする。sourceの実在と意味/完全性はproducer別に受領し、旧候補名や既存modelの存在だけで接続済みとしない。

候補`finance_source_event`の一意キーを `(clinic_id,source_owner,event_id)` とし、版・内容digest、種別、元請求/資金/元event、実日時/記録日時、実方法、実移動額、請求差分を別列に持つ。訂正は新event ID+元ref。D12受入表も同じsourceキーを一意とし、同内容はno-op/別内容は照合停止。通知再送で新eventを作らず、欠落した投影だけを同IDから再建する。

`cash_segment_projection`候補は医院/営業日/区分/採用境界設定版/source世代をキー参照とし、開始残・現金実受領・現金実返金・補充・回収を元ID付きで集計する。理論現金＝開始残＋現金実受領−現金実返金＋補充−回収。非現金受領、請求金額差分、預り充当/解除/再配分は現金増減0。元カードを現金で実返金した時は実返金方法で減少。過受領は実釣銭後の受領を一度だけ用い、了承預りと請求配分の二重現金計上をしない。

preview応答候補は`boundary_version/source_generation/source_set_digest/opening_version/cash_adjustment_version/theoretical_amount`。締めcommitは全版と完全性を再検査し、その区分へのイベント受入と直列化して原close snapshotを一度保存。イベント先行なら原締め、締め先行なら後日差分の一方へ。遅着を実日時変更で原締めへ押し込まない。非0差額のメモ必須・0差額任意、原因不明/確認中可の既決を維持する。原close/訂正列は追記、実査訂正は絶対値の後継なので複数実査額を合算しない。

検査：8,000請求に5,000実受領、預り5,000受領→充当→取消復元→実返金、元カード/現金返金、AMのみ締め、越日EMG、source重複/欠落/同ID異内容、preview後イベント、二重締め、監査失敗。原締め不変、receipt同一、差分帰属と各実方法の一度計上を照合する。

### FIN-R18：月次未納のcutoff・調整内訳

候補queryは`clinic, authorized_scope, period_start_JST, period_end_exclusive_JST, owner_filter, source_generation`。未納は原請求IDと元owner/petに帰属し、今月売上0の繰越請求を母集団から落とさない。各請求への元受領配分と非金銭調整を別eventで投影する。現残だけのAPIは過去cutoffを証明しない。

月次投影行の候補は `owner/pet群、前月末残、当月発生、当月実受領の請求帰属、非入金調整行[]、翌月繰越、cutoff/世代、complete`。調整行は元event ID/版・請求/資金ref・発生日・種別・符号付き残高影響・根拠を保持。同じ減額を当月発生へ控除し調整でも控除しない。正本の限定対象だけを含む月は、前月末残＋当月未払い−当月入金＋調整増減＝翌月繰越。次月の前月末残は同じ境界へ連続する。

実返金・未充当受領・非現金超過を含む月は、実受領総額と請求に帰属した額の照合差を元資金/eventから説明する。正本が決めていない返金の当月入金相殺方式を候補schemaから決めない。未説明差は`complete=false`で理由を返し、表示writeや架空調整で閉じない。画面/印刷は同じquery/世代/調整内訳を用い、確認済み0件と欠落/失敗を区別。

検査：10,000＋5,000−4,000−2,000−1,000＝8,000、充当済み5,000の同月発生/解除/取消→残0/実入金0、月末JST直前直後、翌月連続性、future訂正、ページ欠落と医院切替遅延。実返金等の表示が未確認なら限定式の成功を全月へ拡張しない。

### FIN-R19：退院4項目・固定出力

候補read adapter応答は `query_ref, source_generation, complete, next_cursor, items[{clinic_id,stay_id,discharge_fact_id,fact_version,effective_relation,actual_discharge_at,admission_date,owner_ref,pet_ref,owner_display_snapshot,pet_display_snapshot}]`。これはD05に提供能力を求める内部案で、現在APIの存在説明ではない。完全性には全ページの同一世代・欠落検出を含め、最後のpageだけで成功にしない。

D12はJSTの実退院日で抽出し、内部キー`clinic+stay+有効退院事実`、安定順`実退院時刻→stay ID→fact ID`を使う候補。紙面では既存日次帳票内に対象日見出しと4列「退院日時/飼主名/ペット名/入院日」を同じ行で配置し、改頁ごとに見出しを繰返す案。金額/担当医等の新列や単独帳票を追加しない。長い名はwrapし、別petへ省略統合しない。

preview版を出力時にも照合し、版変化は再取得、0件確認は空表示可、不完全は出力停止。許可済み訂正を現在有効事実へ投影し、未知の退院取消/再開writerを作らない。DC-F01〜04で同名別pet/再入院/JST境界/全ページ/遅延応答/preview後変更/再印刷write0を確認。請求用ケアquantity=1を退院一覧粒度や実退院日の代用にしない。

### D13・FIN-R22：数量版・元出庫返却と旧値

候補`inventory_quantity_head`は`clinic+inventory_item_id`一意、管理単位/quantity/version/基準snapshot_ref。`inventory_movement`は実績ID・操作ref・種別・正整数量・前後数量/版・実日時/記録日時・理由/actor・元臨床ref・元出庫refを保持する。物理単位との換算を確認できない臨床小数量は在庫移動停止で、臨床量を丸めない。

棚卸は実査絶対値を入力しても、保存時には期待版の現数量との差分と理由を持つ実績へ変換する候補。出庫/返却/棚卸/訂正は同じ数量headを直列化し、CAS更新＋実績＋必要監査＋receiptを同時保存する。返却は元出庫headも同順で保護し、有効返却累計≤元出庫の未返却量。数量非負、同操作→同実績ID、元実績訂正は追記で累計再検査。属性PATCHの絶対quantityを裏口にしない。

限定静的根拠：`backend/internal/inventory/repository.go:132–150`は更新と再読取のtransaction、`:161–209`は通常DeleteIfUnusedの依存条件、`:211–233`は正整数/不足同UPDATE減算、`:236–260`は名称/medicine分類による別削除・名称更新。既存正整数保護は維持し、通常削除と名称補助経路を同じ保証と断言しない。理由実績・数量版・元出庫返却・receiptの全経路成立は未検証。

旧数量は品目/単位/確認数量/旧ref/確認actor日時の基準snapshotにし、架空入荷実績を作らない。旧版なしwriteは最新head取得から開始。元出庫不明の返却、単位不明・名称だけの対応は停止。物販名称一致による自動在庫連動、ロット管理等を追加しない。検査は棚卸/出庫両順、同元出庫二重返却、同操作再送、属性変更/出庫、名称同期/使用中削除競合、旧値不明停止。

### FIN-R23〜R25：実行待ちの証跡単位

実API/DB二session試験は各intentに`環境/版/clinic/fixture/必要grant/操作key/前後残・数量/監査/receipt/両先行順/結果`を保存する。PDFと実プリンタ、帳票項目と税務適用、手動返金記録と決済事業者の実返金は別レーン・別判定。帳票matrixは請求、実受領領収、預り受領、最終会計充当、訂正/実返金/再発行、月次未納4列+調整、日次退院4列ごとに元取引/版・宛名/登録番号/税内訳・pet別/共通・今回差分・期間を対応させる。未実行・fixture不足・実設定未確認はPARTIAL/BLOCKEDで、本文精査やCI成功からPASSを転記しない。

## 10/19横断検索・同種操作の設計接続

[共通adapter](../../shared/integration-contracts.md#cross-ui-search-adapter-20261019)を本領域の[適用台帳](../../../spec/shared/interaction-scope.md)L/M/H/S consumerへ接続する。draft/applied・独立filter・取得scope/世代・復帰/失敗状態を分離し、検索ボタンだけでキーワードを適用。既存API検索と取得済み内絞込みの責務、ページング/上限、認可・保存方式を維持する。E/N操作へ部品単位で一括適用しない。実装/全caller/実機未確認。

<a id="doctor-confirm-gate-20261019"></a>
## 医師確認 gate と金額了承の設計接続（EMR-462 案・2026-10-19・2026-10-20 訂正）

Completeの医師確認gateは[medical-records側の本人決定・最終DDL接続](../medical-records/technical-spec.md#doctor-confirm-gate-and-price-ack-20261019)へ従う。廃止済み`billing_confirmations.status`を残す旧最小gate案は使用せず、`doctor_review_sets`・`doctor_review_records`・会計の確認参照と現在の対象集合／臨床世代を照合する。金額了承は既存`billing_price_acknowledgements`＋`billing_change_events`へ接続する。確認後の金額変更は警告了承で可能・医師再確認不要とし、薬剤処置追加削除／数量変更は医師再確認まで請求確定・精算を停止する。確認前の準備／金額未確定表示は可。会計からプランを自動書換えしない。旧調査案の再掲は本人回答の撤回ではない。保存形状の存在をComplete・全writer・画面の実装／実機合格とせず、API／業務処理・DDLは今回変更しない。

具体的なgateのtransaction参加・operation_bindings/operation_resultsへの接続設計はEMR-462の着手前に確定する。保存形状の選定と、後続のgate実装方式が未確定であることを分けて扱う。
