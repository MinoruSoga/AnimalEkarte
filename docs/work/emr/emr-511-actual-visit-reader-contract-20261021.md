# EMR-511 実来院 reader 公開接続の採用契約

2026-10-21。既決業務仕様に基づく root 採用済み技術契約。新たな業務決定・実装済み・DB適用済みを意味しない。本文の「案」「提案」は未実装の採用技術契約を指す。今回の変更は本書を含む新規2文書のみで、コード変更・試験・API登録・DB適用は行わない。実装時の試験は NOT_RUN。

## 1. 根拠と現状

確認 main: `7a57b7bfebacdf212bc9555ac6622c0df8f3af99`。Plane EMR-511 は 2026-10-21T01:03:49Z 更新の FIRST-SLICE-20261021-511 を読み直した。第一PRは private reader/projectionのみであり、本案はその後の公開接続を具体化する。

- [EMR-511](https://app.plane.so/baritechllc/browse/EMR-511/)
- [業務正本 D16](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/docs/spec/domains/analytics/business-rules.md#requirement-fr-analytics-owner-counts)：実来院日distinct、未知を0にしない、旧記録日数、歴史的飼主帰属。冒頭「今期の提供範囲」はランキング・犬種・CPM人数/一覧を対象外とする。
- [既存選定read契約](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/docs/design/domains/finance-operations/technical-spec.md#analytics-actual-visit-source)：受付/過去入力のsource owner、未知母集団の非除外、観測窓・閾値不変。
- [現handler/route/DTO](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/backend/internal/lstep/aggregation_handler.go)：GET `/api/v1/clinics/:clinic_id/owners/aggregations`、owners:view＋selected clinic grant。total/annual_visit_countはint64、日付はomitempty。
- [現SQL](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/backend/internal/owner/ltv_repository_query.go)：DISTINCT mr.date、現在pet.owner_id、COALESCE(...,0)、NULL日付→no_visit。
- [既存repository](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/backend/internal/owner/ltv_repository.go)、[service/CPM](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/backend/internal/lstep/aggregation_service.go)。SQL→service→DTOすべて旧int64を保持している。
- [fact最終DDL静的展開](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/docs/architecture/erd-detail/actual_visit_facts.md)、[保存model](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/backend/internal/model/storage_clinical.go)：14列、kind=checkin/retrospective/correction、source_version>0、訂正参照FK。corrected_fact_idにUNIQUEなし、evidenceはobject以外の意味制約なし。coverage台帳・正当な無来院の証明はこの14列にない。
- [FE API型](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/frontend/src/features/aggregation/api/get-aggregations.ts)、[CSV](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/frontend/src/features/aggregation/lib/aggregation-csv.ts)、[表示列](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/frontend/src/features/aggregation/components/AggregationOwnerTableColumns.tsx)。旧FEはnumber前提、CSVはString(total_visit_count)であり、単純nullable化不可。
- [LSTEP観測SQL](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/backend/internal/lstep/lstep_delivery_trigger_log_repository.go)：firedログのfired_at、非抑止母集団、JST開始日とdays後の日を両端包含。今回365日metricと観測窓を混同しない。

## 2. 選定案の要点

1. 既存endpointにopt-in `visit_basis=actual_visit_v1` を追加。未指定/`legacy_record_days_v1`は旧互換。既存フィールド名の意味・型を静かに入れ替えない。
2. 新値は必須nullableメンバーを持つ `actual_visit` オブジェクトに隔離。旧値は `legacy_record_days` に明示名でまとめ、旧トップレベルキーは互換aliasとして残す。旧値を新値のfallbackにしない。
3. 新modeの抽出・ソートはactualだけで行う。unknown行を無断除外せず、判定不能行を一覧に残し不完全性と件数を明示する。
4. sourceの完全性証拠が未提供でも契約/adapter/unknown経路は実装可能。実データのcompleteを宣言するには468/D03の証拠が必要。リリース日やテーブル空をcoverageの証明にしない。
5. 既存最終来院抽出とその必要な表示/CSVに限定。ランキング・犬種・CPM UIを復活させず、配信write、金額集計移行、backfillは行わない。

## 3. Requestと互換境界

Routeと認可は既存のまま。新しいendpoint/権限は不要。

追加query:
- `visit_basis`: `legacy_record_days_v1`（未指定時）/`actual_visit_v1`。その他400。
- `visit_snapshot`: 一覧の後続ページ取得時だけ送るopaque token。別医院・別認可scope・別条件には使えない。tokenは認可ではなく毎回現在認可を再検査。

既存の検索、page/per_pageと上限制約を維持。legacy modeのparser/resolver/default sortは変更しない。actual modeは独立したparser/resolverで1回だけ取得した `as_of_instant` をAsia/Tokyoへ変換し、`as_of_date_jst` を作る。OS time.Localや下位関数のtime.Nowを参照しない。日付は厳密なYYYY-MM-DD、実在暦日。未来factは予定来院として数えずsource不整合とする。

### actual modeの期間解決（採用案）

優先順位は既存意図と同じ `from/to > year > period_preset > 全期間`。ただし全入力の型/enum/範囲は優先度によらず検証する。from/toの片方だけ、from>toは400。yearは既存1900〜2100。未来だけの期間は400。要求末尾が未来に達する場合、実来院のresolved末尾は必ずas_of翌日へ切り詰め、応答にその実際の範囲を返す。

- from/to: 両端包含入力を `[from, min(to+1日, as_of+1日))` へ解決。
- year: `[year-01-01, min((year+1)-01-01, as_of+1日))`。未来yearは400。
- last_3_months / last_6_months / last_12_months: `[as_ofをN暦月戻した日, as_of+1日)`。日が移動先月にない場合はその月末にclampする。Go AddDateの月越し正規化をそのまま使わない。3/6/12ヶ月は90/180/365日固定ではない。既存FEの「直近3ヶ月/6ヶ月/12ヶ月」のラベルに対応し、開始日・当日を含む。
- calendar_year: `[as_of年01-01, as_of+1日)`。既存FE「今年度」の実装上の意味は暦年であり4月年度に変えない。必要なら表示を「今年（暦年）」として誤読を防ぐ。
- 未指定: 通算（from/to_exclusiveともnull）。通算coverageは全過去からas_ofまでを要求。
- rolling365はselected periodと独立して常に `[as_of−364日, as_of+1日)`。

例 as_of=2026-10-21: 3ヶ月は2026-07-21〜2026-10-22未満、6ヶ月は2026-04-21〜同末尾、12ヶ月は2025-10-21〜同末尾、暦年は2026-01-01〜同末尾、rolling365は2025-10-22〜同末尾。2026-05-31から3ヶ月前は2026-02-28、2024-02-29から12ヶ月前は2023-02-28。明示2025年は2025-01-01〜2026-01-01未満。current yearの未来日を完全性の未達要件にしない。

これは新actual基準だけの具体化で、legacy calculateDateRangeのNow/time.Local/timestamp挙動を修正するものではない。actual mode内の既存金額条件も同じresolved期間を渡すが、金額source・式・請求日意味を新台帳へ移行しない。

### actual modeのmetric/sort

- metric未指定は `last_visit`。明示 `last_visit` のみ許可。それ以外（annual_sales/visit_countを含む）は400 `unsupported_visit_basis_metric`。旧modeの既定annual_salesは不変。
- sort未指定は `last_visit_date`、order未指定は `desc`。明示sortは `last_visit_date`（last_visitは同値alias）、`days_since_last_visit`、`owner_name` のみ。その他（旧既定total_amount・annual_amount・回数ランキングsortを含む）は400 `unsupported_visit_basis_sort`。orderはasc/desc以外400。旧resolveAggregationSortのdefault total_amountへ落とさない。
- 既存回数min/max filterはselected_periodのactual countに適用。last_visit_bucketは既存5enum、未知enumは400。
- `cpm_stage` filterはactual modeでは400 `unsupported_visit_basis_filter`。旧CPMをactual抽出へ混ぜない。旧modeの既存互換動作は不変。
- 新ランキング/CPM/breed UIへの接続なし。金額条件は既存式のまま、新金銭sourceに適合済みとは表示しない。

## 4. Exact response shape

既存listのtotal/page/per_page/owners/cpm_stage_countsを残す。下記は追加メンバー。actual modeではactual_visitとvisit_queryを必須、nullメンバーをomitemptyにしない。

```json
{
  "visit_query": {
    "basis": "actual_visit_v1",
    "legacy_basis": "legacy_record_days_v1",
    "as_of_date_jst": "2026-10-21",
    "selected_period": {"from": "2026-10-01", "to_exclusive": "2026-10-22"},
    "rolling365": {"from": "2025-10-22", "to_exclusive": "2026-10-22"},
    "snapshot": "opaque-server-token",
    "complete": false,
    "matched_total": 8,
    "unknown_total": 2
  },
  "total": 10,
  "page": 1,
  "per_page": 20,
  "owners": [{
    "owner_id": "42",
    "owner_name": "表示名",
    "actual_visit": {
      "basis": "actual_visit_v1",
      "filter_match": null,
      "lifetime": {"count": null, "complete": false, "reasons": ["coverage_unproven"]},
      "selected_period": {"count": 1, "complete": true, "reasons": []},
      "rolling365": {"count": null, "complete": false, "reasons": ["coverage_gap"]},
      "first_date_jst": null,
      "last_date_jst": null,
      "days_since_last_visit": null,
      "last_visit_bucket": null,
      "dates_complete": false,
      "date_reasons": ["coverage_unproven"]
    },
    "legacy_record_days": {
      "basis": "legacy_record_days_v1",
      "total_record_day_count": 12,
      "rolling365_record_day_count": 4,
      "selected_period_record_day_count": 2,
      "first_record_date_jst": "2020-01-02",
      "last_record_date_jst": "2026-10-01"
    }
  }]
}
```

この例は追加メンバーの説明用で、実際の既存必須金額/互換メンバーを削除しない。旧top-level total_visit_count等は旧record-day値のaliasでありactual値へコピーしない。互換メンバーのdeprecationとbasisをapi.yamlに明記する。旧値を表示するcallerのラベルは「旧記録日数」「旧記録最終日」。

型:
- count: integer >=0 またはnull。新countは日数でJS安全整数を超えない。ID/source_versionは正規形十進string、DB bigint上限を守る。内部idをNumberへ変換しない。
- complete: boolean。falseなら対象countは必ずnull。true/0だけが確認済み0。証明のない部分観測件数をcountに入れない。
- reasons/date_reasons: unique enum配列、completeなら空。`coverage_unproven`, `coverage_gap`, `source_unresolved`, `source_version_mismatch`, `correction_conflict`, `invalid_scope`, `invalid_visit_date`。DB障害を理由配列で隠さない。
- dates_complete: 初期接続では通算coverageと有効graphが全証明された時だけtrue。通算未確認時に「確認できた中の最大日」を真の最終日と表示しない。最小/最大の片側証明による最適化は本PRでは行わない。
- complete通算0: first/last/daysはnull、bucket=`no_visit`、dates_complete=true。
- complete通算>0: first/lastはYYYY-MM-DD、daysはJST暦日差、bucketは既存の<90/<180/<365/それ以降。unknown bucketはnullでありno_visitやcpm_unclassifiedではない。
- rolling365: as_of日を含む365JST暦日 `[as_of−364日, as_of+1日)`。既存legacy SQLのNOW−365days結果を黙って書き換えない。新基準の日付粒度を明示しfixtureで両端を固定。
- selected periodは§3のactual専用resolverの結果を返す。無期間ならfrom/to_exclusiveともnullで通算と同じ。片側だけnullは不可。

## 5. unknownを消さないfilter/sort/pagination

1. clinic権限、owner名等の非来院条件を先に適用し候補集合を確定。actual経路は旧SQL `buildLTVHaving` のvisit min/max predicateと、旧 `filterLTVRows` のshouldExcludeNoVisit/last_visit_bucket predicateの両方をバイパスする。既存repositoryへ通常paramsのまま入れて後付けactual計算する方式は禁止。金額/検索等の非来院条件だけを分離したcandidate readerを使う。
2. 同じsnapshotでactual projectionを計算。来院predicateはtrue/false/unknownの3値。falseのみ落とす。unknownは候補一覧に残しfilter_match=null。条件なしはfilter_match=trueでも指標の不完全性は残す。
3. bucket条件の既存OR意味を保持する。選択bucketがno_visit以外なら、`bucket一致 OR (include_no_visit=true AND 証明済みno_visit)`。no_visit選択は証明済みno_visitだけtrue。bucket未選択でinclude_no_visit=falseは証明済みno_visitだけfalse。未知bucketは全ケースunknown。回数min/max等の別条件とは3値AND（いずれかfalseならfalse、falseなしでunknownがあればunknown）。未知はno_visitに合流させない。
4. total=既知match＋判定不能。matched_totalとunknown_totalは来院predicate判定後・ページ分割前。visit_query.completeも返却対象全体からページ分割前に計算する。後ろのページに未知があっても先頭ページをcompleteと偽装しない。未知件数は既知0に加算しない。
5. 指定sortの既知値を昇降順、当該sort値が未知なら最後、同値はowner_idの数値順で安定化。owner_name sortは名称の既存照合とID tie-break。来院なしと未知の表示区別は維持。
6. snapshot bindingのnormalized queryは `page` と `visit_snapshot` 自身を除外する。clinic/現在認可scope、visit_basis、per_page、実効sort/order/metric、正規化された検索・LINE/金額/回数/bucket/include条件、as_of日、resolved selected/rolling窓を含む。優先順位で無効になった入力でなく解決後の意味をbindingする。次pageは同じtokenで取得可能。per_pageや条件変更は初回pageから新tokenを取得。
7. snapshot fingerprintにはsource事実・全訂正閉包・source verification/coverage証拠だけでなく、候補owner集合、表示/順序に使うowner名、LINE条件の値、非削除/医院scope、返却金額と金額filterの値・依存元版を含む。結果・順序・totalに影響する変更はすべて検出する。単一read snapshotで取得し、次pageでfingerprintが変われば409 `visit_snapshot_changed`。MAX(id)単独は世代証明ではない。tokenは現在認可の代用不可。
8. FEは未知行に「実来院情報の照合未完了」を表示。フィルタ結果が完全でないことと既知一致/未確定件数を表示し、未知を勝手に隠すtoggleは追加しない。
9. CSVは既存どおり「現在ページで選択した行だけ」。選択0件はボタンdisabled＋handler早期return。ページ跨ぎ収集・全件取得・全件exportは実装しない。既存許可済み最終来院CSVの選択行のみ、新基準値を空欄＋basis/complete/reason列で出力し、未知を0/no_visitにせずformula escapingを維持。取得済み同一responseから出力し、医院/条件/ページ/response snapshotが変わったら選択を解除する。照合未完了の行を選択したCSVは不完全性を明記し、完全な抽出結果と表示しない。

## 6. 468/D03 source handoff（内部interface案。既存APIの存在主張ではない）

`ActualVisitSourceReader.ReadSnapshot(ctx, authorizedClinic, scope, windows)` → `SourceSnapshot`。writer直呼びHTTPは新設しない。既存のDB読取transactionへ渡すinternal adapter。

SourceSnapshot:
- `clinic_id`, `scope`（owner/pet識別、必要な旧/新帰属集合）, `as_of`, `snapshot_fingerprint`
- `facts[]`: final model14列に対応。source_kind, appointment_change_event_id / record_correction_link_id, source_version, corrected_fact_id, owner_id, pet_id, visit_date_jst, evidence, operation_id, recorded_by/atを欠落なく渡す。
- `source_verifications[]`: fact_id、根拠owner=`D01`/`D03`、source_ref/source_version、同医院・患者・当時owner照合、`verified|unresolved|contradictory`、evidence_ref。
- `coverage[]`: source ownerごとに `scope`, `from_date_jst|null`, `to_exclusive_jst`, `all_history_before_end_verified`, `status=verified|unproven|gap`, `evidence_ref|null`, `revision_token|null`, `unresolved_source_refs[]`。null開始はall_history=trueの場合だけ全過去の証明。それ以外は不明であり無限過去保証ではない。
- coverage証拠は当該scope/期間の対象sourceが欠落なく照合されたこと、全訂正の閉包、同一read世代を示す。アプリ導入日・最初のfact日・最後のid・空table・新writer配備だけでは証明不可。
- 現在その証拠を永続管理/生成する仕組みは本調査で確認できていない。初期providerは正直にunprovenを返す。テストfixtureだけverifiedを与え、実DBに架空のstart日やverified値をseedしない。coverage保存が必要なら別途DBownerへ具体設計を渡すが本PRで新DDL/backfillしない。

468 handoff: `reservation_guarded_checkin.go`のguardedCheckIn/checkInConfirmedを主候補とし、Updateのchecked_in遷移を含む全確定受付入口を棚卸し。成功source event・actual_visit_fact・必要監査/489 receiptは同じ最外transactionに参加。再送は同source/refへ収束、失敗/取消予定だけではfact生成なし。戻り値にcommit済みfact ID/版/根拠refを出す内部契約。既存上書きchecked_in_atから過去を再構成しない。

D03 handoff: 明示過去来院確認済み入力と患者/日付訂正のownerが、record_correction_link_id、確認した実日、当時owner/pet、根拠版を渡す。外部資料・転記・単なる医師確認は未適格。専用実装hook名は未確認なので架空の関数を既存と記載しない。writer担当が実在経路とcommit境界を同定する作業を残す。

訂正: 全graphをclosureとして取得し、期間/owner filterより先に解決。旧日→新日や旧owner→新ownerを跨ぐ訂正を途中で切らない。各rootから一意で整合するterminalを採用し、元factは残す。fork/cycle/missing/cross-clinic/版不一致は影響するscope/windowをunknown。影響範囲さえ証明できなければ要求clinic全体を不完全とする。MAX(id)/recorded_at順で勝者を発明しない。取消しが実来院そのものの否認を意味するか不明なsourceはsource owner照合待ちとし、予約cancel/statusだけで元来院を消さない。新しい取消業務を本readerから作らない。

509には必要なsource_ref/版/clinic/owner/pet/operation_idの照合subsetだけを引き渡す。finance_source_events / ingestionを実来院の根拠にしない。売上返金や請求取消は来院訂正ではない。

## 7. LSTEP/CPM接続境界

- 新response内の旧cpm_stage/cpm_stage_countsはlegacy基準の互換情報。actual値と同基準だと表示しない。新actual modeをCPM公開UIへ接続しない。
- private readerは将来LSTEP/CPMに同じbasis/completeを渡せるが、本PRではタグ同期・自動配信・母数・CPM版/閾値を変更しない。
- 観測consumer引継ぎではログ母数/fired_at/非抑止、days=1〜365、既存JST両端包含を維持。ログ1件の来院観測はtrue/false/unknown、証明された不存在だけfalse。不明ログを母数から除外して率を良くしない。新観測wireの未接続をEMR-511の残工程として明記し、owners APIだけで全consumer完成にしない。

## 8. エラー/認可

400: query/date/basis不正。401/403:従来の認証/owners:view/現在clinic grant不足。409:後続pageのsnapshot変化。503 aggregation_timeoutは既存の自身deadlineのみ。source/DB通信障害は既存error分類の非200で返す。成功応答内のunknownは取得できた根拠の不足/不整合であって、通信失敗の成功偽装ではない。

404/空ownerリストだけで来院0を主張しない。current owner一覧の権限を維持しつつ、歴史factsのowner_idはcurrent pet.owner_idへrewriteしない。削除済みownerを新規公開したり権限範囲を広げたりしない。死亡は履歴計数の除外条件ではなく通知抑止だけ別扱い。

## 9. 最小作業単位と完了判定

第一slice（既存計画）は維持。本案を採用後の「公開読取接続」1PR:
1. `actual_visit_reader.go` / `actual_visit_projection.go`の内部成果を再利用し、source verification/coverage provider interfaceとunproven実装を追加。private fixtureでcompleteとunknownの双方を記述。
2. owner LTV adapter→lstep service→handler/query→api.yamlを一括変更。旧mode維持、新mode3値filter/snapshot/nullable DTO。shared OpenAPI/DIはownerへ調整し、無断並行編集なし。
3. aggregation FE API型・query key（basis/clinic/query/snapshot）・既存最終来院表示/filters・許可済みCSVのbasis/unknownを同期。新ランキング/CPM/breed UIを作らない。
4. 合成受入fixture、旧API互換例、NOT_RUNと未接続consumer/source一覧をdocs/Planeに記録。

このPRは「unknownを正しく公開できるreader接続」で閉じられる。実データcomplete、writer全入口実装、LSTEP観測切替が未完了ならEMR-511全体のUAT/完成を主張しない。全会計完成や実DBbackfillを着手gateにはしない。

最小handoff: 468は受付source/証拠、D03は過去来院/訂正source/証拠、511はprojection/wire/必要FE、489はwriterのreceipt既存契約、509は別のfinance source。新しい重複ticketは不要。rootが各既存票へこのinterfaceと残完了条件を追補すればよい。

## 10. 固定fixture（作成対象、実行NOT_RUN）

- 同医院owner2pets同日→owner1、pet各1。同pet再受付同日1、別日2、別医院非混在。
- 確定受付＋draftchartは1。予約/カルテ/会計単独・将来予定・外部転記は0を作らず非根拠。
- unknown coverage＋空facts→null。証明済全source空→0/no_visit。期間completeで通算unknownの例。
- 主飼主変更後も旧fact owner帰属維持。日/患者訂正の旧/新scope両方、期間外から内/内から外の訂正。
- chain一意leaf、fork/cycle/missing source/版不一致はunknown。元fact保持。DB障害は非200。
- actual resolverの明示from/to > year > preset、片側日付拒否、未来末尾clamp、月末/閏年clamp、calendar_year、無指定通算。legacy挙動は不変。actual未指定sort=last_visit_date、明示非対応sort/metric拒否。
- rolling365の開始前/開始日/as_of日/翌日。JST日界、90/180/365分類境界。first/lastとcountは同有効集合。
- SQL HAVINGとfilterLTVRowsの旧visit条件を通らずunknown残存。bucket OR 証明済みno_visit、false AND unknown=false、true AND unknown=unknown。pageだけ変更した同token成功、per_page/条件変更は新query。
- unknown filter行残存、total=matched+unknown、pagination安定、page間source/owner名/金額変化409、先頭page既知のみでも後続未知があればcomplete=false。CSVは現在page選択行だけ、0選択disabled、条件/ページ変更で選択解除。
- 旧クライアント互換、explicit legacyラベル、新FEはactual absent/nullを旧countへfallbackしない。
- 認可喪失/医院切替で旧cache表示不可。死亡で歴史回数不変、通知write0。

## 11. 本人への追加質問

本案を進めるための新しい業務判断は見つからない。来院意味、旧値区別、今期除外、historical owner保持は既決。未確認なのはcoverage証拠の実在・writer全入口の実装状態・D03 hookで、ユーザーに開始日を推測させる質問ではなく調査/実装handoff事項。rootは技術契約案として採用/修正できる。実データ完全性を宣言する前にはsource ownerの検証可能な証拠が必要。

## 12. 独立レビュー反映（2026-10-21）

旧visit条件のSQL/Go二段漏れ、snapshotのpage自己bindingと結果依存不足、既存resolverの独立Now/time.Local、CSVの選択範囲拡張を修正した。追加根拠：[query parser](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/backend/internal/lstep/aggregation_request.go)、[期間ラベル](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/frontend/src/features/aggregation/lib/aggregation-filter-panel-model.ts)、[現在page選択CSV](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/frontend/src/features/aggregation/routes/AggregationDashboardPage.tsx)。owner_idは既存からstringなので修正不要。
