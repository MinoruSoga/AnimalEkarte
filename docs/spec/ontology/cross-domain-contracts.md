# 横断概念・識別・関係・状態のモデル

<a id="cross-domain-model"></a>
## 適用と意味の正本

2026-10-18の構築承認に基づくローカルの概念設計。業務語彙は既存の[領域別ontology](README.md#ontology-current-20261018)を再利用し、本書は領域をまたぐ関係・識別・制約の正本とする。操作の可否と受入条件は参照先FR／ACが優先する。ここでいうイベントは「起きた業務上の事実」の説明で、event sourcing・message bus・DB schema・API追加の採用ではない。

安定した意味アンカーを概念の参照に用い、既存FR／ACを再採番しない。表の識別要素は業務上取り違えないための区別で、複合PK・外部キー・global IDを選ぶものではない。記載のない多重度や履歴件数は未選定とし、0..1を必須1へ読み替えない。

<a id="identity-axes"></a>
## エンティティと識別軸

| 対象／意味の正本 | 識別して保つもの | 同一性として使わないもの／制約の追跡 |
|:---|:---|:---|
| [飼主・患者](owners-pets.md#common-patient-concept) | 共通の人/個体と、利用医院・主副飼主・院別業務帰属 | 氏名・電話・同日・ペット名の一致だけでID同一としない。FR-owners-pets-common-use、AC-owners-pets-duplicate-confirm |
| [予約・受付](reservations.md#reservation-ontology) | 同じ予約と対象ペット、医院、業務分類、予定時間、実進行 | カード表示位置・一覧の日付・カルテ番号を予約IDの代用にしない。FR-reservations-checkin-record-reuse |
| [飼主事前回答](reservations.md#revisit-intake-concept) | 予約×対象ペット、原文の現回答/旧回答と日時 | スタッフ聞取/通常カルテ/同じ飼主の別頭の回答と別。FR-reservations-line-revisit-intake |
| [通常カルテ・編集権](medical-records.md#record-edit-right-concept) | 記録の同一性、保存版、現在の編集権、未保存入力、実編集者 | 担当医/確認医師/閲覧者/権保持端末を同一にしない。FR-medical-records-single-editor |
| [通常診療の治療プラン項目集合](medical-records.md#outpatient-treatment-plan-concept) | 今回項目、元マスタ/複写元、項目状態、当時の値 | 統合画面タブと保存単位を同一視しない。FR-medical-records-integrated-plan |
| [滞在](hospitalization.md#stay-identity-and-time-concept) | 入院/宿泊の実滞在、予定・実開始終了、日次記録 | 同じペット/日付だけで別滞在を統合しない。FR-hospitalization-entry-and-start |
| [トリミング記録](trimming.md) | 予約の施術detail、実担当、コース/option、要望/結果 | 通常カルテ・予約notes・希望/完成写真を同義にしない。FR-trimming-reservation-and-detail |
| [検査結果・健診](examinations.md#clinical-results-and-certificates-boundary) | 検査ID/固定結果版/版内項目、出典、健診の独立result set | 同名項目・受信日時・表示先・自己申告と確認結果を同一にしない。FR-examinations-revision、FR-checkups-canonical |
| [証明書発行版](medical-records.md#certificate-edition-concept) | 様式/発行内容/訂正前後の関係 | 接種記録ID、共通基本情報の現在値、D21専用帳票と別。FR-vaccinations-certificate-editions |
| [請求・資金](accounting.md#accounting-ontology) | 請求/元明細/対象飼主ペット、実受領/実返金、配分/預り残 | 予約ID・領収書印刷回数・充当操作を実入金の同一性にしない。FR-accounting-receivable-partial-zero |
| [見積・締め・在庫](finance-operations.md) | 提示版/後継、営業日区分/元締め/追記、物品/実移動 | 見積承認=売上、請求取消=在庫返却とはしない。各owner正本参照 |
| [スタッフ・認可・公開](access-and-publication.md) | 職員/認証主体/所属/勤務/許可/選択医院、公開目的 | 職種/担当/共有患者の閲覧を書込や公開の許可にしない。FR-access-control-independent-authority |

<a id="cross-domain-relations"></a>
## 関係と多重度

以下は概念関係で、矢印は参照・派生・受渡しの方向を示す。各正本にない上限や必須性を補わない。

| 関係 | 方向・多重度／任意性 | 制約・FR／AC追跡 |
|:---|:---|:---|
| 飼主→ペット、ペット→主飼主 | 一飼主に複数ペット、現在主飼主は一人、副飼主は複数可 | [飼主正本](owners-pets.md)。FR-owners-pets-primary-transfer、過去請求/預り帰属を現在主飼主へ自動移動しない |
| 共通患者↔利用医院 | 複数医院での利用可という業務関係 | [共通患者](owners-pets.md#common-patient-concept)。AC-owners-pets-other-clinic-use／AC-owners-pets-common-denied。共有ID/物理統合方式は未選定 |
| 予約→区分→固定業務分類 | 通常の予約登録は1区分、新規利用する自由登録区分は固定業務分類1つが必須。未設定の旧区分は個別正本の内容確認・補完/新規利用停止と履歴保持へ従い、架空の分類を補わない。表示グループは0..1で別 | [予約関係](reservations.md#reservation-concept-relations)。分類ごとの記録ownerへ接続し、グループ名/色でroutingしない |
| 通常診療の予約→active通常カルテ | 作成前/失敗は0、active最大1。正常受付は同じ予約の受付と作成/再利用の成功が整合 | [check-in](../workflows/reservation-to-record-flow.md#checkin-contract)、FR-reservations-checkin-record-reuse。全分類/過去専用入力へ一律適用しない |
| 予約×ペット→飼主原文 | 任意のペット別回答、回答訂正履歴は複数あり得る | [問診関係](reservations.md#revisit-intake-concept)、AC-reservations-revisit-multi-pet。物理予約モデル全体の多重度をこの問診契約から決めない |
| 通常カルテ→編集権 | ある時点の有効保持0..1、閲覧者人数の上限は新設しない | [編集権](medical-records.md#record-edit-right-concept)、AC-medical-records-single-editor-header。強制取得でも同時に2権利を作らない |
| 通常カルテ→治療項目/測定/保存版 | 複数項目/複数測定/複数保存版を取り得る | [D03](medical-records.md)。未実施等の項目状態、最新測定と履歴を分け、同日の旧測定を失わない |
| 通常カルテ→会話要約項目 | 来院カルテごとの専用項目。履歴版の件数は未選定 | FR-medical-records-conversation-summary、AC-medical-records-summary-hidden。全文/手動貼付/画面表示を追加しない |
| 予定入院/診察後決定→入院 | 二つの入口を同じ入院へ合流。物理対応の多重度は未選定 | [入院](hospitalization.md#stay-identity-and-time-concept)、FR-hospitalization-entry-and-start。二重作成/通常カルテの自動変換なし |
| 滞在→ケア項目/日次記録 | 複数の独立入力、プラン/実績を区別 | FR-hospitalization-distinct-plans-and-records。ケアプラン保存を実施済み記録にしない |
| トリミング予約→detail | 作成前はなし、同一予約のdetailを重複作成しない | [D06](trimming.md)、FR-trimming-reservation-and-detail。独立の通常カルテへ統合しない |
| 予約→コース/追加option | コース0..1、異種option0..複数、同一1回・数量なし | [コース関係](reservations.md#booking-course-option-concept)、AC-reservations-single-course-limit。内包重複禁止、courseなし候補別設定 |
| 検査→結果版→版内項目 | 一検査に複数版、一版に複数項目を取り得る | [D07](examinations.md)。AC-examinations-cross-version。版内refを名称/順序だけで別版へ移さない |
| 元プラン項目→会計明細 | 元項目を追跡。分割/集約の物理上限は未選定 | [同期](accounting.md#doctor-review-and-charge-concept)、AC-accounting-plan-sync-three-branches。未確定/手修正/確定の3分岐、二重請求を作らない |
| 飼主→請求、請求→元明細/入金 | 一請求の飼主は一人、複数ペット可、確定時元明細1件以上、入金0..複数 | [D10関係](accounting.md)。0円で架空入金なし、共通料金はペット参照なし |
| 実入金→元請求への配分 | 一実入金を複数請求へ配分し得る | 同上。配分回数で受領額を増やさず、過去未納回収を新売上にしない |
| 預り受領→飼主/予約/充当 | 元飼主必須、予約0..1任意、充当/解除/返金履歴を保持 | FR-accounting-reservation-advance。同医院同飼主、別ペット利用可、他医院/他飼主へ自動移行なし |
| 元締め→訂正差分 | 元を保持し追記0..複数 | FR-closing-append-only。締め訂正を実入出金/元締め書換えとしない |
| 保存済み接種→証明書発行版 | 訂正後の新発行と旧版の関係を保持、物理件数は未選定 | [証明書](medical-records.md#certificate-edition-concept)。AC-vaccinations-certificate-correction |
| 接種→証明書の適格性・発行前入力 | 狂犬病様式は適格フラグと接種時スナップショットが真のときだけ。発行前入力10欄はすべて任意、通常様式は空のみ。1接種に発行版0..n、1発行版に様式artifact1..n | [適格性・入力・権限](../domains/vaccinations/lifecycle.md#certificate-eligibility-inputs-20261021)、[表示規則](../domains/vaccinations/lifecycle.md#certificate-display-rules-20261021)。AC-vaccinations-certificate-eligibility-gate／AC-vaccinations-certificate-inputs-optional／AC-vaccinations-certificate-normal-null-only |
| 証明書の発行版→表示発行日／実発行時刻 | 表示発行日は初回はプレビュー日に固定、再発行は当初の発行日が初期値で変更可。実発行時刻は別に保持 | [再発行日](../domains/vaccinations/lifecycle.md#fr-vaccinations-reissue-date-20261021)。同版再印刷は版を増やさず、新版再発行は旧版を保持。AC-vaccinations-reissue-default／AC-vaccinations-reprint-fixed |
| 帳票セット（RV4種類）→発行版・artifact | 出力対象の構成。診療項目パッケージとは別概念で、パッケージ適用から自動発行しない | [RV一括出力](../domains/vaccinations/lifecycle.md#fr-vaccinations-rv-batch-20261021)。欠損・重複・一部失敗を4種類完了としない。AC-vaccinations-rv-partial |
| 健診Checkup→外注PDF原本→文字起こし結果→職員確認済み結果 | 原本・候補・確定の順で、確認前を確定扱いしない。原本1..複数、候補は再抽出ごと、確定版は確認保存ごと。D07検査とは別 | [外注PDF](../domains/checkups/lifecycle.md#fr-checkups-outsourced-pdf-20261021)、[語彙](examinations.md#outsourced-checkup-pdf-concepts)。AC-checkups-pdf-review／AC-checkups-pdf-shared-result |
| 機器ジョブ状態→再評価可否 | 再評価はreceived／needs_reviewで未persistのジョブのみ。persisted後は再解決を拒否しD07訂正経路へ | [FR-examinations-mapping](../domains/examinations/lifecycle.md#fr-examinations-mapping)、[語彙](examinations.md#examination-state-and-receive-concepts) |
| サポート報告→外部送信（スクリーンショット） | 送信前の内容・送信先確認は維持し、マスキングを必須条件にしない | [D23](../domains/support/README.md#support-screenshot-personal-information-20261020)、[公開・読取](access-and-publication.md#read-and-publication-boundaries) |
| ペット死亡登録（D04）→既存の未来予約（D01） | 既存の未来予約は取消（削除ではない）。誤登録の取消でも自動再開しない | [D04](../domains/owners-pets/lifecycle.md#pet-death-future-reservations-cancel-20261022)、語彙は[飼主・ペット](owners-pets.md)と[予約](reservations.md) |
| 入院明細の税情報未確認（D05）→会計取込（D10） | 税情報が未確認の明細は概算のまま保持し、会計取込は税基準未確認として停止。退院のみは独立に継続可 | [入院税情報](../domains/hospitalization/lifecycle.md#hospitalization-tax-unknown-20261020)。既定値の仮明細を作らない |
| 月次売上の主表示（D16） | 主表示は税抜（純売上・部門別）で、消費税と税込売上合計を併記 | [集計の税基準](../domains/analytics/business-rules.md#analytics-sales-tax-basis-20261022)。表示規則で金額・入金を変更しない |

<a id="state-and-event-axes"></a>
## 状態遷移とイベントの独立軸

保存・業務進行・医師確認・請求確定・資金・在庫は直交する事実である。例えば保存成功から診療終了、診療終了から全明細実施済み、医師確認から入金済みを導出しない。以下はイベント名の説明であり、新status enumを指定しない。

| 対象／イベント | 前後・結果 | 保持するもの／制約の追跡 |
|:---|:---|:---|
| 受付/診療開始/診療終了 | 同じ予約が受付済→診療中→会計待ちへ明示進行 | [予約正本](reservations.md)。終了後も通常カルテ編集可。FR-medical-records-screen-consultation-end |
| 取消/未来院/再受付/誤開始終了Undo | 同じ予約を個別の承認済み専用操作で扱う | [逆遷移](../workflows/reservation-to-record-flow.md#requirement-fr-reservations-controlled-reverse-transition)。会計確定歴があれば通常Undo停止、返金/取消で自動再開なし |
| 編集開始/タブ移動/離脱 | 明示取得、同カルテで維持、保存結果確認後に解放 | [編集権](medical-records.md#record-edit-right-concept)。単なる無操作解放なし |
| 強制移転/切断/復帰/reload | 移転は旧編集保存停止、切断は入力保持停止、復帰は権/保存結果照合、reloadは入力回復可能 | 同正本。再取得後最新比較・明示選択、旧入力自動再送なし |
| プラン項目の状態変更 | 未実施/実施済み/後日予定/中止の4状態、実施取消は専用理由付き | [D03](medical-records.md)。中止は原記録削除でなく、請求/在庫/通知の逆実行なし |
| 医師確認/確認後の変更 | 会計準備は先行可、確定精算は確認後。単価/割引等の金額のみは警告了承、薬剤・処置の追加削除/数量変更は再確認まで確定・精算停止 | FR-accounting-doctor-review。予約に独立の医師確認待ちstatusを足さない |
| プラン更新→会計同期 | 未確定未手修正は自動、未確定手修正は差異判断、確定後訂正 | FR-accounting-plan-sync。会計→プランの逆自動更新なし |
| 実施/準備/引渡し/出庫 | 院内投薬注射の実施後、持帰薬物販の内容数量確定/準備後に請求対象 | FR-accounting-billable-performance。実引渡し/出庫と別、予定額を売上にしない |
| 入院開始/退院 | 担当医決定済み明示開始と記録作成整合、実退院は予定終了と別 | FR-hospitalization-entry-and-start／FR-hospitalization-discharge-and-billing。会計なし退院可の既存条件、入金は別 |
| トリミング開始/終了/実施なし中止 | 専用記録で進行、終了で会計待ち、実施なし中止は理由と請求判断を区別 | FR-trimming-safety-and-status、FR-trimming-zero-performance-abort。中止料の個別契約維持 |
| 検査結果保存/封印/訂正 | D07自身の確定/公式版/訂正作業版を区別 | FR-examinations-revision。D03旧finalized撤回でD07封印を解除しない |
| 請求確定/部分入金/後日回収 | 請求確定と受領は別、入金0又は部分でも確定請求あり | FR-accounting-receivable-partial-zero。残額を表示、同元請求へ配分 |
| 請求取消/減額/充当解除/実返金 | 請求と配分/預り残の訂正、実返却を別々に追跡 | FR-accounting-reasoned-financial-correction、FR-accounting-reservation-advance。原請求/入金/確定歴を保持 |
| 過受領の整合 | 入力訂正、実返金又は了承付き明示預り | [過受領](accounting.md#excess-receipt-concept)。同受領は1回のみ、未確認差額のまま完了しない |
| 締め/締め後訂正 | 実際に締めた区分の原snapshotを保護し追記 | [D12](finance-operations.md)。未締め別区分まで一律保護しない、差額メモと架空移動は別 |
| 共通情報変更/証明訂正/再印刷 | 基本情報の現在値更新と新発行版作成、既存版再現は別 | FR-owners-pets-common-use、FR-vaccinations-certificate-editions。旧訂正版の有効出力不可 |

<a id="temporal-and-history-axes"></a>
## 時点・版・履歴

| 時点／版 | 意味と混同しないもの | 正本・受入 |
|:---|:---|:---|
| 予定日時/実施測定日時/入力保存日時 | 計画、実際の事象、記録操作の異なる時間。不明を現在時刻で補わない | [D03測定](medical-records.md#medical-record-diagnosis-vitals-ontology)、AC-medical-records-vital-latest-measured-timeの定義は画面正本を参照 |
| 保存済み版/現在の未保存入力 | 成立済みの保存事実と未送信/送信中/失敗/競合/結果不明の操作状態 | FR-medical-records-preserve-autosave。結果不明は同一操作照合、別createや無条件再送なし |
| タブ開始版/複製直前版/最新保存版 | クリア基準、複製取消基準、再取得後比較相手である別の基準 | FR-medical-records-clear-tab-open-state、FR-medical-records-history-copy-undo、AC-medical-records-editor-reacquire-select |
| マスタ現在値/採用時snapshot | 予約時間、臨床用量、会計価格/割引/保険/税等で当時値を保持 | [予約](reservations.md#booking-course-option-concept)、[会計](accounting.md)。現マスタへ無断追従せず、明示再適用は正本条件 |
| 現回答/旧回答、公式結果版/訂正作業版、発行版 | 元対象・出典・actor・時点・訂正関係を保持する異なる履歴 | [問診](reservations.md#revisit-intake-concept)、[検査](examinations.md)、[証明書](medical-records.md#certificate-edition-concept) |
| 実退院JST暦日/金銭営業日/請求日/実入出金日 | 退院一覧の[D00:00,翌D00:00)と、締めの営業日区分・金銭移動日の帰属 | [日次退院](finance-operations.md#discharge-report-and-accounting-time)。AC-closing-discharge-calendar-day-not-closing-period |
| 検索日付/医院filter | 画面表示条件であり業務発生日時/認可の代用ではない | FR-common-ui-list-filter。詳細復帰/当日初期値/ペット全期間/医院保持を区別 |

## 残件の詳細設計案への対応（2026-10-18）

[C06の責任ownerと提案／未回答境界](../../design/shared/integration-contracts.md#detail-design-proposals-20261018)へ接続する。業務正本・既存FR/ACを置換せず、版と操作結果の混同を防ぐ照合である。

| 関係 | 今回具体化した制約・正本 |
|:---|:---|
| 統合タブ→各owner保存、予約×pet→回答世代 | 画面一つ≠tx一つ。原文は独立reader、旧targetの遅延送信を新petへ適用しない。[D03/D01](medical-records.md#record-technical-version-relations) |
| record→編集権、保存operation→初回結果／現在状態 | 保持0..1と保存版・actorを分離。強制取得前成功は保持、取得後旧世代write拒否。成功receiptで現在状態を巻き戻さない。logout下書きは本人専用で同本人再ログイン後だけ回収し、別主体／患者へ適用しない |
| common subject→医院利用、LINE binding→本人確認 | 共通IDは他院認可・LINE本人確認ではない。[D04](owners-pets.md#common-subject-technical-relations)の条件付き移行Q-I1と本人確定の独立解除Q-I2を区別。解除しても予約／患者関連は保持、再連携後はstaffが選択確認 |
| 臨床対象集合→確認receipt→会計確定 | 金額変更と臨床変更を別世代へ。未手修正／手修正／確定後の3分岐と確定歴保持。[D10](accounting.md#accounting-review-version-relations) |
| 接種→証明系譜→発行版→artifact | 保存事実・発行・描画・物理印刷を分離。旧訂正版は有効出力不可。Q-V2は既決適用として訂正保存成功で停止、失敗時は保持し競合版を照合。Q-V1は本人確定で通常縦横の同じ最新版を共有し印刷時に選択、固有項目保持。[D08](../../design/domains/vaccinations/technical-spec.md#certificate-edition-proposal) |
| 退院事実→日次帳票行 | 元stayと有効事実／版を一度だけ参照し金銭へwriteしない。[D12](../../design/domains/finance-operations/technical-spec.md#discharge-report-projection-proposal) |

<a id="cross-domain-invariants"></a>
## 横断不変条件と設計を残す境界

| 不変条件 | 確認する反例／追跡 |
|:---|:---|
| 同一性・scopeを保存/訂正/復元時にも照合する | 同名患者、別院ID、画面切替前の遅延応答で他対象へ適用しない。AC-owners-pets-common-denied、AC-access-control-04 |
| 一つの実効果を再送/配分/表示回数で増殖させない | 受付再試行のカルテ重複、受領と充当の二重入金、取消の二重在庫戻しを禁止。AC-owners-pets-unknown-result、AC-accounting-excess-consented-deposit-single-receipt |
| 確認表示だけで現在の状態/認可/版検証を代替しない | 確認後に権利/明細/依存が変化した場合は古い確認で確定しない。編集権/会計/主飼主移管の各正本 |
| 不明と無し/0/陰性/成功を分ける | 失敗した取得を空記録・空ケージ・未検査・保存済みにしない。AC-owners-pets-infection-ambiguity、各領域失敗条件 |
| 表示/出力と業務writeを混同しない | 証明書印刷で接種登録、退院行表示で金銭移動、要望表示で施術記録作成をしない。各projectionの正本 |
| 原本履歴と現在値を両立する | 非表示/取消/クリア/訂正で旧保存版・原画像・請求確定歴を消去しない。FR-common-retention-visibility／FR-common-retention-trace |

内部共有ID・unlink、統合プラン保存単位、編集権API/失効/保持形式、帳票の固定版保存/render、同期/補償のトランザクション境界などは設計。現在のコード・テーブル名は参考対応であり、未承認の多重度/業務状態を補完する根拠にはしない。用語整理は実装/実機/医療値妥当性の合格ではない。

<a id="search-operation-context-concept"></a>
## 検索操作と表示文脈（10/19横断合意）

[同種操作の正本](../shared/ui-contracts.md#interaction-consistency-20261019)と[適用台帳](../shared/interaction-scope.md)へ接続。操作の同一性は名前/部品ではなく用途・対象・効果で判断する。検索の入力中値、明示適用条件、取得/絞込み範囲、表示結果と要求世代、選択済み業務IDを別概念とする。医院/pet/日付/filter/pageは適用文脈であり、検索語だけで結果を識別しない。

入力→検索適用→結果→明示選択→保存は別の関係で、検索は書込・患者同定・link/コピー/請求/通知の成功を含まない。clear入力/検索reset/フォームclear/業務取消も別の効果。未取得/失敗は0件を意味せず、現在ページ内の0件は全期間全件の不存在を意味しない。E用途の候補補完とL/M/Hの保存済み一覧検索、Sのowner/pet選択を区別する。物理schema/新enum/認可モデルの追加ではない。
