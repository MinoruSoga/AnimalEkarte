# 証明書・院内マニュアルの限定静的監査（2026-10-23）

**計画整備。製品実装Ready・公開契約採択・実行受入ではない。**

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.audit-certificates-manual-20261023 |
| 種別 | 限定静的監査の受領記録 |
| 決定状態 | 既存採択仕様を参照。新業務・技術契約の採択なし |
| 実装状態 | 以下の固定範囲だけを分類。全票・全ACの実装完了認定なし |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23（資料受領・静的整形） |

[今回の索引・判定軸](emr-limited-audit-plans-index-20261023.md#limited-audit-status)を参照。以下の監査範囲・件数は受領した固定基準の報告であり、今回再度の全件照会を行ったという意味ではない。

基準main: `171b56575e7ee028e54358dad22ae28574853125`（2026-10-23 読取）。製品実装・試験・DB・Plane書込なし。

<a id="audit-certificates-manual-conclusion"></a>
## 結論
486/487は読取3操作・未接続FE subsetまで。writer・artifact結合・適合renderer・回復fence・入口接続・RV4/再発行日の全体は未完。502は既存editと公開readerを受領済み、公開操作・復元操作は親の既知残scope。新欠陥や重複票として扱わない。

<a id="audit-certificates-manual-evidence"></a>
## 証拠と確認境界
- Plane全コメントを取得、最新回答を旧質問より優先。コメント件数9/15/11・next_page_results=false。
- PR772 edf71515 / PR791 e71ddf03 / PR800 d5d5b767 / PR778 eb137881 はmerged。PRの限定scopeを親完了へ拡張しない。
- 既存後続525/526/513/527確認。検索は題名中心のため未発見を全workspace不存在の証明にしない。
- 公開READの将来string-ID wire、hide suppression、維持/移行実証は#778に分離済み。既知残差であり新欠陥を増やさない。
- 実bucket/role・鍵・DB適用状態・実フォント資産導入は未確認。ソース上のnil/falseと実環境receipt欠如を混同しない。

<a id="audit-certificates-manual-handoffs"></a>
## 有限成果と受領先
<a id="audit-certificates-manual-c1-v1-context-preview-content-missing"></a>
### C1 既存v1のcontext・発行前preview・発行・再印刷・content・結果照会を縦切り接続 [missing]
- owner: EMR-486
- 成果: 未登録6操作は必要順の成功単位へ分ける。A=context＋発行前previewの固定snapshot/eligibility/10入力、B=issue＋元key結果照会のclaim anchor/CAS/同TX edition・履歴・監査・receipt、C=reprintの不変edition/現在有効性/元key receipt、D=contentのprivate store読取/現在認可・版・digest照合。read3操作を再実装せず、登録した操作に対応するphantom entryだけ除去する。B/Cは必要準備/lookup/回復fenceを受領し、Dはstore受領後に独立して閉じられる範囲で先行できる。安全上不可分な必要subsetだけ同PRへ含め、6操作全体を機械的に一PRへ束ねない。
- 依存: 489必要subsetのみ。DB既存保存形を再利用。実bucketと描画資産の設定は別条件。
- 既存後続: 既存親486の残scope。独立後続票は証明書検索範囲では未発見。新欠陥として再起票しない。
- 根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/routes_records.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/apicontract/openapi_route_drift_test.go#L195 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/cmd/api/composition_medicalrecord.go#L105

<a id="audit-certificates-manual-c2-private-artifact-adapter-dependency"></a>
### C2 private artifact adapter・環境結合・候補照合を受領 [dependency]
- owner: EMR-486
- 成果: Get/PutIfAbsent/StatDigest/DeleteCandidateの専用adapter、決定的key、DB参照保護、候補閉鎖とreconciliationを実装。実環境bucket/role deployment receiptを別受領。本番未結合をnilのまま完成としない。
- 依存: Cloudflare R2専用private方針は10/22回答済み。実bucket/資格情報/権限の作成・設定は別承認。unknown/busy候補は自動削除しない。
- 既存後続: 486既存質問への回答76c90766が現在正本。再質問・新規独立票は不要。
- 根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/cmd/api/composition_medicalrecord.go#L105 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_handler.go#L305

<a id="audit-certificates-manual-c3-font-renderer-missing"></a>
### C3 採用埋込fontに適合した決定的rendererを受領 [missing]
- owner: EMR-487
- 成果: Noto Sans CJK JP Regular 2.004の実資産digest/ライセンス/manifest、決定的subset、実幅・Unicode・欠字停止、16MiB制約を実装し3既知様式を照合。flagだけ変更しない。
- 依存: 元font名/版/ライセンス選定済み。実asset採用receipt・library/subset実装は未受領。RV4見本は別。
- 既存後続: 487既存回答6733def8の技術残。
- 根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_render.go#L29

<a id="audit-certificates-manual-c4-rv4-master-4-batch-normal-missing"></a>
### C4 RV4 master・全4原子batch・normal再発行日を段階接続 [missing]
- owner: EMR-486 / EMR-487
- 成果: 採択v2契約のmaster revisions/4member guard、batch/member保存・候補準備・全4commit、normal表示日当初default/変更/再preview/新不変版を有限PRへ分離。FEとserver renderer各受領SHAを記録。
- 依存: wire/storage方式は選定済み。4名称/見本/紙寸法/field mapping未受領のみ資料依存。4種類を4ページ・旧RV4複製としない。
- 既存後続: 486/487既存追加要件。別票は検索で未発見。
- 根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/routes_records.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_render.go#L29 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/apicontract/openapi_route_drift_test.go#L195

<a id="audit-certificates-manual-c5-discovery-edition-slice-b-fence-missing"></a>
### C5 回復discovery・非表示edition再印刷探索・Slice B fenceを接続 [missing]
- owner: EMR-487 / EMR-486 / EMR-489必要subset
- 成果: 既存queryと接種単位reprints/queryのAPI収録・BE/FE adapter、固定upper/keyset、元edition認可、fence/世代CASを接続。0件/completeから新key許可なしを維持。
- 依存: PR804で接種単位queryは技術選定済み。PR800 Slice Aだけで完了不可。fence本体のowner受領が必要。
- 既存後続: PR800 r4335183217既知未解決、487コメント91182220で対応契約あり。独立新欠陥にしない。
- 根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/vaccinations/lib/certificate-recovery-gate.ts / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/apicontract/openapi_route_drift_test.go#L195

<a id="audit-certificates-manual-c6-panel-missing"></a>
### C6 保存済み接種詳細の証明書入口と共通panelを結合 [missing]
- owner: EMR-487
- 成果: C1/C2/C3/C5を受領した限定接続PR。既存FE mock/内部版検証/CSPを再実装せず、保存済み接種詳細から接続し独立routeは増やさない。
- 依存: C1/C2/C3/C5。結合/実印刷検証は本人指示待ちNOT_RUN。
- 既存後続: 既存487。525/526は別完結済み小修正で代用不可。
- 根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/vaccinations/lib/certificate-recovery-gate.ts / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/vaccinations/api/certificate-http-client.ts#L115

<a id="audit-certificates-manual-m1-closed-wire-missing"></a>
### M1 公開操作のclosed wire選定と縦切り実装 [missing]
- owner: EMR-502
- 成果: 選択content版を固定して公開pointer・公開履歴・監査・receiptを同TXにする。具体endpoint/DTO/エラー/同key lookupを先に確定し、既存editor→公開→公開readerまで接続。
- 依存: 業務・原子性・CAS・receiptは採択済み。既存edit限定契約はpublish APIを定義していない。全体仕様待ちではなくclosed wire選定残。
- 既存後続: 502親scopeとして既記載。検索「公開」「復元」では別のEMR実装票なし。
- 根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/work/emr/emr-502-manual-edit-contract-20261021.md#L5 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/design/domains/internal-manual/technical-spec.md#L33 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/handler.go#L373

<a id="audit-certificates-manual-m2-closed-wire-missing"></a>
### M2 過去版復元のclosed wire選定と新内容版作成を接続 [missing]
- owner: EMR-502
- 成果: 明示元versionをContentに拘束し、新revision/current pointer/復元元履歴/監査/receiptを原子的保存。旧版上書きなし、公開と復元を混同しない。既存history UIから明示操作を接続。
- 依存: 業務意味はAC05で採択済み。具体wire/participant/routeは未選定・未実装。
- 既存後続: 502既知残scope。独立後続票は検索で未発見。
- 根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/work/emr/emr-502-manual-edit-contract-20261021.md#L5 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/design/domains/internal-manual/technical-spec.md#L33 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/handler.go#L373

<a id="audit-certificates-manual-m3-dbortx-2-missing"></a>
### M3 返却済みDBOrTx在庫2メソッドの証跡を回収 [missing]
- owner: EMR-502（共通operation ownerと調整）
- 成果: RecordUnknownUnderLock / MarkApplySavepointのambient TX原子性fixture対応を確認し、必要差分とallowlistを同時に回収。allowlist追記だけで閉じない。
- 依存: 10/22 23:45コメントd811f052で502へ返却済み。現mainでも2メソッドはDBOrTx使用、allowlistなし。テスト未実行につき新しいFAIL実測とは呼ばない。
- 既存後続: 502既存返却コメント。共通型は別527、重複票を作らない。
- 根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/operation/store.go#L260 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/lintscan/dbortx_inventory_lint_test.go#L870

<a id="audit-certificates-manual-m4-dependency"></a>
### M4 共通生成型と稼働設定を別受領 [dependency]
- owner: EMR-527 / deployment owner
- 成果: 527の正規生成・caller整合成果を受領。manual edit HMAC keyringの環境receiptを確認し、未設定ならfail-closedと記録。鍵作成/設定をこの監査で行わない。
- 依存: 527既存管理票。設定済み実環境は未確認で欠落断定しない。
- 既存後続: EMR-527あり。PR778共通型分離を解消済みと推測しない。
- 根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/cmd/api/composition_runtime.go#L750

<a id="audit-certificates-manual-m5-hide-override-wire-missing"></a>
### M5 hide/override解除と公開wire移行の既知分離境界を保持 [missing]
- owner: EMR-502（既存scopeの後段）
- 成果: 既知の後段操作について履歴保持・公開reader suppression・FE builtin復帰条件をclosed contractにし、明示hide/remove_overrideと結果照合を縦切り回収。旧DELETEを新保持契約完成としない。
- 依存: PR778で数値ID/平坦公開wireを維持する限定採択あり。新wireへ無断変更せず別限定単位として選定する。実DB内容・削除結果は未確認。
- 既存後続: PR778 r4334962851で既知deferred。独立後続票は今回未確認。新欠陥として追加しない。
- 根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/work/emr/emr-502-manual-edit-contract-20261021.md#L5 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/design/domains/internal-manual/technical-spec.md#L33 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/handler.go#L373

<a id="audit-certificates-manual-ac"></a>
## AC個別判定
statusは全ACを静的完了とする意味ではない。全行の実行検証はNOT_RUN。
- **AC-vaccinations-certificate-paper-size — dependency / NOT_RUN**: 既知3様式の描画コードはあるが非埋込font・production=false。本番描画と見本/紙寸法受入を満たさない。 回収単位: C3。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_render.go#L29
- **AC-vaccinations-certificate-three-layouts — dependency / NOT_RUN**: 既知3様式の描画コードはあるが非埋込font・production=false。本番描画と見本/紙寸法受入を満たさない。 回収単位: C3。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_render.go#L29
- **AC-vaccinations-certificate-layout-shared-edition — missing / NOT_RUN**: read/FEの同版layout選択subsetはある。発行・再印刷writer未接続で版増分なしの全経路未成立。 回収単位: C1, C6。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/routes_records.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/apicontract/openapi_route_drift_test.go#L195
- **AC-vaccinations-certificate-layout-fields-fixed — dependency / NOT_RUN**: 固定snapshot reader/rendererは実装。artifact nilとrenderer停止により実出力未成立。 回収単位: C2, C3。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_reader.go#L194 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/cmd/api/composition_medicalrecord.go#L105 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_render.go#L29
- **AC-vaccinations-certificate-correction — missing / NOT_RUN**: source_changed等validity projectionと保存preview旧版拒否は実装。新版発行/再印刷/直接content/訂正との原子境界は未回収。 回収単位: C1。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_reader.go#L194 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_handler.go#L305 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/apicontract/openapi_route_drift_test.go#L195
- **AC-vaccinations-certificate-basic-update — missing / NOT_RUN**: 保存snapshot readerは現在値で補完しない。新発行と固定版再印刷を含む全経路は未接続。 回収単位: C1, C2。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_reader.go#L194 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/apicontract/openapi_route_drift_test.go#L195
- **AC-vaccinations-certificate-no-side-effect — missing / NOT_RUN**: 保存previewはread-only。正式発行/再印刷全経路がないため機能として完了とはしない。 回収単位: C1, C6。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_handler.go#L305 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/apicontract/openapi_route_drift_test.go#L195
- **AC-vaccinations-certificate-eligibility-gate — missing / NOT_RUN**: eligibilityは公開選定・FE mockのみ。master flag/接種snapshot供給とcontext/write検査が未回収。 回収単位: C1。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/apicontract/openapi_route_drift_test.go#L195
- **AC-vaccinations-certificate-inputs-optional — missing / NOT_RUN**: FE/保存読取subsetあり。ただし発行前preview/issue等未登録。書込時検証・現在認可・snapshot固定の受入は残る。 回収単位: C1, C6。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/routes_records.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/apicontract/openapi_route_drift_test.go#L195
- **AC-vaccinations-certificate-normal-null-only — missing / NOT_RUN**: FE/保存読取subsetあり。ただし発行前preview/issue等未登録。書込時検証・現在認可・snapshot固定の受入は残る。 回収単位: C1, C6。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/routes_records.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/apicontract/openapi_route_drift_test.go#L195
- **AC-vaccinations-certificate-issue-date-initial — missing / NOT_RUN**: FE/保存読取subsetあり。ただし発行前preview/issue等未登録。書込時検証・現在認可・snapshot固定の受入は残る。 回収単位: C1, C6。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/routes_records.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/apicontract/openapi_route_drift_test.go#L195
- **AC-vaccinations-certificate-permissions — missing / NOT_RUN**: FE/保存読取subsetあり。ただし発行前preview/issue等未登録。書込時検証・現在認可・snapshot固定の受入は残る。 回収単位: C1, C6。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/routes_records.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/apicontract/openapi_route_drift_test.go#L195
- **AC-vaccinations-rv-four-batch — missing / NOT_RUN**: v2採択契約あり、runtime/master/batch/再発行日writer未接続。見本資料依存とは分離し共通構造から進行可能。 回収単位: C4。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/routes_records.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_render.go#L29
- **AC-vaccinations-rv-missing — missing / NOT_RUN**: v2採択契約あり、runtime/master/batch/再発行日writer未接続。見本資料依存とは分離し共通構造から進行可能。 回収単位: C4。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/routes_records.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_render.go#L29
- **AC-vaccinations-rv-duplicate — missing / NOT_RUN**: v2採択契約あり、runtime/master/batch/再発行日writer未接続。見本資料依存とは分離し共通構造から進行可能。 回収単位: C4。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/routes_records.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_render.go#L29
- **AC-vaccinations-rv-partial — missing / NOT_RUN**: v2採択契約あり、runtime/master/batch/再発行日writer未接続。見本資料依存とは分離し共通構造から進行可能。 回収単位: C4。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/routes_records.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_render.go#L29
- **AC-vaccinations-reissue-default — missing / NOT_RUN**: v2採択契約あり、runtime/master/batch/再発行日writer未接続。見本資料依存とは分離し共通構造から進行可能。 回収単位: C4。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/routes_records.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_render.go#L29
- **AC-vaccinations-reissue-change — missing / NOT_RUN**: v2採択契約あり、runtime/master/batch/再発行日writer未接続。見本資料依存とは分離し共通構造から進行可能。 回収単位: C4。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/routes_records.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_render.go#L29
- **AC-vaccinations-rv-history-scope — missing / NOT_RUN**: v2採択契約あり、runtime/master/batch/再発行日writer未接続。見本資料依存とは分離し共通構造から進行可能。 回収単位: C4。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/routes_records.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccination_certificate_render.go#L29
- **AC-vaccinations-reprint-fixed — missing / NOT_RUN**: 保存版read/同版layout subsetはある。reprint writer/content未登録。 回収単位: C1, C2。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/apicontract/openapi_route_drift_test.go#L195
- **AC-vaccinations-rv-package-boundary — unverified / NOT_RUN**: 独立概念の契約は採択。RV4自体未接続でありpackage apply経路全体の副作用否定は本監査未確認。 回収単位: C4。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/routes_records.go
- **AC-internal-manual-01 — implemented / NOT_RUN**: 公開readerはpublished版のみ、全認証staffが取得。FE strict読取は失敗時本文/検索/印刷を停止。01のスクリーンショット・実印刷までは未検証。 回収単位: 追加単位なし／限定確認。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/repository.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/handler.go#L373 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/manual/routes/ManualPage.tsx#L60
- **AC-internal-manual-02 — implemented / NOT_RUN**: 公開readerはpublished版のみ、全認証staffが取得。FE strict読取は失敗時本文/検索/印刷を停止。01のスクリーンショット・実印刷までは未検証。 回収単位: 追加単位なし／限定確認。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/repository.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/handler.go#L373 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/manual/routes/ManualPage.tsx#L60
- **AC-internal-manual-03 — implemented / NOT_RUN**: 公開readerはpublished版のみ、全認証staffが取得。FE strict読取は失敗時本文/検索/印刷を停止。01のスクリーンショット・実印刷までは未検証。 回収単位: 追加単位なし／限定確認。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/repository.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/handler.go#L373 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/manual/routes/ManualPage.tsx#L60
- **AC-internal-manual-06 — implemented / NOT_RUN**: 公開readerはpublished版のみ、全認証staffが取得。FE strict読取は失敗時本文/検索/印刷を停止。01のスクリーンショット・実印刷までは未検証。 回収単位: 追加単位なし／限定確認。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/repository.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/handler.go#L373 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/manual/routes/ManualPage.tsx#L60
- **AC-internal-manual-04 — unverified / NOT_RUN**: 既存edit FEは採用済み、draft保持コメントあり。保存失敗時の全UI状態を今回静的に再監査し切っていない。試験待ちと実装欠落を区別。 回収単位: 追加単位なし／限定確認。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/manual/routes/ManualPage.tsx#L60 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/participant.go#L230
- **AC-internal-manual-05 — missing / NOT_RUN**: 復元は新内容版という採択scope内だがedit限定PRから除外、専用routeなし。 回収単位: M2。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/handler.go#L373 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/design/domains/internal-manual/technical-spec.md#L33
- **AC-internal-manual-07 — implemented / NOT_RUN**: 既存editはglobal row lockと期待Version/current content CAS、公開pointer不変。runtime競合実証はNOT_RUN。 回収単位: 追加単位なし／限定確認。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/participant.go#L230
- **AC-internal-manual-10 — implemented / NOT_RUN**: 既存editはglobal row lockと期待Version/current content CAS、公開pointer不変。runtime競合実証はNOT_RUN。 回収単位: 追加単位なし／限定確認。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/participant.go#L230
- **AC-internal-manual-08 — missing / NOT_RUN**: hide/remove_override履歴移行は#778対象外。現DELETEは旧repository.Delete→handler別audit。保持を壊す実行結果は未検証で断定しない。 回収単位: M5。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/work/emr/emr-502-manual-edit-contract-20261021.md#L5 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/repository.go / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/handler.go#L373
- **AC-internal-manual-09 — missing / NOT_RUN**: editは版/current/履歴/監査/receipt同TXのコードあり。ただしAC対象の公開操作そのものが未接続。 回収単位: M1。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/participant.go#L230 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/handler.go#L373
- **AC-internal-manual-11 — missing / NOT_RUN**: edit同key lookup・原receipt・unknown停止は実装。publish/restore同key照合は操作未実装のため全AC完了不可。 回収単位: M1, M2。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/handler.go#L373 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/participant.go#L230
- **AC-internal-manual-12 — missing / NOT_RUN**: edit出所不明停止・公開reader未公開除外は実装。公開/復元の旧出所確認は操作未実装。 回収単位: M1, M2。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/participant.go#L230 / https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/repository.go
- **AC-internal-manual-multipage-print — unverified / NOT_RUN**: 今回の502限定監査ではprint CSS/ブラウザ複数ページを検証していない。新欠陥認定なし。 回収単位: 追加単位なし／限定確認。根拠: https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/manual/routes/ManualPage.tsx#L60

<a id="audit-certificates-manual-ticket-acceptance"></a>
## 票の5条件との対応
- 486 AC1: read/snapshot追跡はimplemented、発行/監査writerはC1 missing。AC2: shared layout readerはimplemented、writerはC1。AC3: validity/preview拒否implemented、訂正/新版/直接contentはC1。AC4: readerは固定snapshot、全発行・再印刷はC1/C2。AC5:対象fixture/同時性/DB/実印刷はNOT_RUN、runtime不足はC1/C2/C5。
- 487 AC1:描画コードsubsetあり、本番埋込/見本受入C3 dependency。AC2:同版出力C1/C2/C6 missing。AC3:保存preview拒否のみimplemented、content/reprintはC1。AC4:read-only subset、全体C1/C6。AC5:golden/render/upload/応答断/実印刷はNOT_RUN、回復C5 missing。
- 502 AC1:既存edit保存→readerのコードはimplemented、publish/restoreはM1/M2。AC2:current/published分離implemented、復元はM2 missing。AC3:edit認可/CAS/receipt implemented、公開/復元はM1/M2、試験NOT_RUN。AC4:edit C06接続implemented、公開/復元missing。AC5:静的path/限定PRは提供済み、M3/527受領と実行検証が残る。
