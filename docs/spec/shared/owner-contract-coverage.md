# 共通契約の全owner適用照合（C01／C02／C03／C05／C06）

| 項目 | 内容 |
|---|---|
| 文書ID | shared.owner-contract-coverage |
| 確認範囲 | D01〜D23の重要な識別・語彙・保存単位・安全・履歴・画面失敗と既存正本の対応 |
| 決定状態 | 個別正本の既決／継承／現行差を参照する索引。新業務の承認ではない |
| 基準 | PR #670 head 41fcf3ebと2026-10-15の統合文書候補。最新承認は各ownerの根拠節 |
| 実装／実機 | 未実施。コード全writer・実grant・旧データ・外部処理の網羅適合を宣言しない |

本表は[UI C01](ui-contracts.md)、[概念・文書状態 C02](document-contracts.md)、[安全 C03](clinical-safety.md)、[保持 C05](retention-audit.md)、[接続 C06](../../design/shared/integration-contracts.md)を具体的なowner操作へ結び付ける。認可は[C04](access-control.md)、横断接続は[X01〜X08](../workflows/reservation-to-record-flow.md#cross-domain-review-20261015)も参照する。リンク先の要件／ACを第二定義としてコピーせず、どの条件をどの操作に使うかを示す。

## A. 識別・語彙・保存単位・画面失敗（C01／C02／C06）

| 領域 | 具体的なID・概念境界 | 保存／画面失敗・再送の適用先と根拠 |
|---|---|---|
| D01 予約 | appointment ID／版、操作intent／キー、business_class、指名と実担当、予約とtentativeを分離。JST時刻や氏名でIDを代替しない | [閉じた操作契約](../../design/domains/reservations/technical-spec.md#reservation-api-closed-contract)の同scope同key／fingerprint、成功結果と現在状態、warning_changed／版競合。未確認結果に別key作成なし |
| D02 受付 | 予約ID単位のカード、専用owner/kind/id/revision、表示列と実status、本人browser非表示を区別 | [snapshot](../../design/domains/reception/technical-spec.md#reception-read-scope)は全件性・total一致を確認。[操作](../../design/domains/reception/technical-spec.md#reception-operation-contract)はprepare→read→start、結果不明は元context照合。更新失敗を0件とせず古いsnapshotからwriteしない |
| D03 通常カルテ | record／患者／子ID・固定結果版、確認医師と編集actor、項目実施状態・自動保存・診療終了と履歴付き通常編集を分離 | [実型・全writer](../../design/domains/medical-records/technical-spec.md#medical-record-reconciled-contract)と[専用intent](../../design/domains/medical-records/technical-spec.md#medical-record-legacy-and-special-contract)。保存§2の個別送信、成功対象のみdirty解除、GET初期化もwriter、未送信/null/clearは型別 |
| D04 台帳 | owner／petと主飼主1人／副飼主複数、D19group・D20LINE linkは別。副飼主追加で請求先・LINE通知先を自動変更せず、主飼主移管と過去資金移管も別 | [登録scope](../domains/owners-pets/lifecycle.md#requirement-fr-owners-pets-register-scope)は新規飼主＋複数pet同時成功。副飼主は[全置換・版](../domains/owners-pets/lifecycle.md#requirement-fr-owners-pets-primary-secondary)。予約の段階患者登録とは異なる保存単位。結果不明のメール／電話一致をreceiptにしない |
| D05 入院・ホテル | stay・hospitalization・通常record・日次record・care item・billingを分離。同一宿泊を二重計数しない | [開始](../domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-entry-and-start)、[ケージ期待割当](../domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-cage-allocation)、[退院](../domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-discharge-and-billing)。日次三入力は独立、交換は両側同時、退院会計は状態と明細同時成功。取得不明を空きにしない |
| D06 トリミング | appointmentと専用detail一件、course/type/option、予約価格と請求snapshotを分離 | [予約detail](../domains/trimming/lifecycle.md#requirement-fr-trimming-reservation-and-detail)は既存を再利用。[画像](../domains/trimming/lifecycle.md#requirement-fr-trimming-image-lifecycle)はupload成功keyとrecord保存別、送信済key保護。[atomicity](../domains/trimming/lifecycle.md#requirement-fr-trimming-delete-and-atomicity)は予約/detail/options／必要監査 |
| D07 検査・受信 | job指紋・source・検体ID、患者attach、exam／revision／版内itemを区別。検体IDは患者自動確定に使わない | [再送・Undo](../domains/examinations/lifecycle.md#fr-examinations-idempotency)と[回復](../domains/examinations/lifecycle.md#fr-examinations-recovery)。receivedとpersistedを分け、未知コードを残す。旧PUT全消去とPATCH不変更を同一化しない |
| D08 接種 | 単一vaccination ID、一接種一保存、同日複数、実績と次回予定、親record帰属を区別 | [二入口同資源](../domains/vaccinations/lifecycle.md#fr-vaccinations-record)は外側カルテ保存と別。[訂正](../domains/vaccinations/lifecycle.md#fr-vaccinations-correction)は成功時だけ初期化、対象ID一行削除、通信不明で再createなし。pointer nullを解除済と誤表示しない |
| D09 健診 | Checkup系のpackage／field／result、旧Examination seedと区別、false/0/未入力を区別 | [型・一括置換](../domains/checkups/lifecycle.md#fr-checkups-results)は未送信/null拒否、[]のみ明示全消去。[Path B](../domains/checkups/lifecycle.md#fr-checkups-protection)は親→子→fieldの段階保存、前段IDを照合し再開。独立フォームに離脱保護が実装済みとしない |
| D10 会計 | 請求／確定歴・実移動event・元資金・元明細・預り金充当／復元を分離。会計医師確認IDと精算IDは別 | [第十群](../domains/accounting/lifecycle.md#accounting-empty-rounding-master)は明細なし確定不可／行別丸め／明示再適用。[受け手](../../design/domains/accounting/technical-spec.md#accounting-handoff-contract)は正確な参照と現在版／receipt。成功／不明・部分実受領を偽装しない |
| D11 見積 | 元見積ID／番号／版／明細と後継、見積承認と未確定会計転記を区別 | [後継](../domains/estimates/business-rules.md#requirement-fr-estimates-successor)は原本不変・新ID、確定親で専用経路。[金額](../domains/estimates/business-rules.md#requirement-fr-estimates-money-source)は新規明細方式、旧ヘッダのみ見積を偽の明細へ移行しない。再送は[選定済操作契約](../../design/domains/finance-operations/technical-spec.md#estimate-selected-contract)の操作IDで結果照会し、後継関係を補助照合。一覧0件だけで失敗認定しない |
| D12 締め | 営業日／AM/PM/EMG、原締めsnapshot／追記差分、実査／理論、支払配賦と割引配賦を区別 | [期間](../domains/cash-register/business-rules.md#requirement-fr-closing-period-boundary)と[append-only](../domains/cash-register/business-rules.md#requirement-fr-closing-append-only)。同医院・日・区分の原締め一件。取得失敗を0円へせず、結果不明は[締め操作契約](../../design/domains/finance-operations/technical-spec.md#closing-selected-contract)の操作IDで結果照会し履歴を補助照合。一覧0件だけで失敗認定しない。開始残と前区分終了残を二重加算しない |
| D13 在庫 | inventory品目／管理単位、臨床投与単位、物販価格masterは別。数量由来statusと保存statusを区別 | [数量状態](../domains/inventory/business-rules.md#requirement-fr-inventory-quantity-status)と[実物調整](../domains/inventory/business-rules.md#requirement-fr-inventory-adjustment)。数量／閾値0の意味、正整数、元操作参照、現在版。[在庫操作契約](../../design/domains/finance-operations/technical-spec.md#inventory-selected-contract)に従い、棚卸の旧絶対値で別出庫を消さない |
| D14 職員・勤務 | staff本人／account／医院所属、職種／grant／個人対応能力、勤務とtemplateを分離 | [本人](../domains/staff-workforce/README.md#requirement-fr-staff-workforce-identity)と[勤務](../domains/staff-workforce/README.md#requirement-fr-staff-workforce-workforce-maintenance)。templateは入力コピーで既存勤務を変更しない。未登録勤務を休日や他院勤務へ補完しない |
| D15 CRM | delivery log／claim／tagcache／external result、import ID／snapshot、shared file／別原本を分離 | [CSV](../line/lstep-integration.md#requirement-fr-crm-csv-snapshot)は新POST=new import、受理数≠挿入数、snapshot rollbackとfailed更新は別。[手動send](../line/lstep-integration.md#requirement-fr-crm-manual-send)は外部成功とlog失敗別。preview件数は最大100の返却集合 |
| D16 集計 | 会計distinct／部門行件数、飼主記録日数／実来院、配信ログ／unique飼主、V1分析／V2配信を分離 | [期間](../domains/analytics/business-rules.md#requirement-fr-analytics-period-basis)、[母集団](../domains/analytics/business-rules.md#requirement-fr-analytics-owner-counts)、[観測](../domains/analytics/business-rules.md#requirement-fr-analytics-lstep-observation)。[集計選定契約](../../design/domains/finance-operations/technical-spec.md#analytics-selected-contract)の旧COALESCE日付、選択ページCSV、取得失敗と率0／7区分0を区別 |
| D17 医院・設定 | 選択clinic／法人global情報、医院固有番号／法人番号、入口閲覧と個別master操作を分離 | [医院](../domains/clinic-settings/README.md#requirement-fr-clinic-settings-clinic-directory)、[入口認可](../domains/clinic-settings/README.md#requirement-fr-clinic-settings-portal-authorization)。院切替中の古い入力を別医院に保存しない。設定ポータルviewで個別writeを許可しない |
| D18 LINE予約 | LINE customer／本人owner・pet、設定公開／枠／確定予約、医師指名とホテル条件を分離 | [保存再判定](../screens/37-line-reserve-owner-flow.md#requirement-fr-line-channel-confirm-recheck)・[入力保持](../screens/37-line-reserve-owner-flow.md#requirement-fr-line-channel-input-retention)・[設定write](../screens/28-line-reservation.md#line-settings-write-contract)。空き表示だけで成功にせず、設定保存結果と実チャネル稼働を別にする |
| D19 identity | group ID／member逆引き、親owner group／pet group、関係と台帳mergeを分離 | [関係](../screens/40-identity-links.md#requirement-fr-identity-manual-relation)と[逆引き](../screens/40-identity-links.md#requirement-fr-identity-member-group-resolution)。同集合createは同groupへ、残一名維持／零名groupのみ削除。unlink NotFoundだけで初回失敗としない |
| D20 健康表示 | 本人LINE認証／単回link token／owner、保存接種・親draftと限定投影を分離 | [単回link](../screens/38-liff-pet-health.md#requirement-fr-liff-health-single-use-link)・[同時成功](../screens/38-liff-pet-health.md#requirement-fr-liff-health-atomic-link)・[公開事実](../screens/38-liff-pet-health.md#requirement-fr-liff-health-published-facts)。使用済tokenは初回失敗の証明でなく、連携状態照会へ |
| D21 院内レポート | 1頭contextと飼主、page grantとsection grant、診療終了した通常カルテと他履歴、raw件数と可視件数を分離 | [1頭](../screens/39-owner-report.md#requirement-fr-owner-report-single-pet-context)、[部分取得](../screens/39-owner-report.md#requirement-fr-owner-report-partial-data)、[閲覧のみ](../screens/39-owner-report.md#requirement-fr-owner-report-read-only)。範囲外petは正規同居petへ、0頭は専用表示。印刷・共有機能を新設しない |
| D22 manual | 静的記事／DB override／公開版／過去版、製品記事と仕様管理文書を分離 | [公開](../domains/internal-manual/README.md#requirement-fr-internal-manual-publication)・[版保持](../domains/internal-manual/README.md#requirement-fr-internal-manual-version-retention)。保存失敗は公開成功にしない、旧版復元は新しい版、override削除と追加記事非表示を区別 |
| D23 support | 共有ROOM表示単位と永続session、共有READ／報告医院mutation／本人history、院内報告と外部issueを分離 | [共有](../domains/support/README.md#requirement-fr-support-shared-read)・[mutation](../domains/support/README.md#requirement-fr-support-report-mutation)・[外部結果](../domains/support/README.md#requirement-fr-support-external-send)。本人history削除は共有元にも反映。storage先行削除後DB失敗を完了にせず再照会 |

## B. 臨床安全・履歴／監査・旧値（C03／C05／C06）

| 領域 | 保護する原事実・具体的安全条件 | 停止／現行差と確認条件 |
|---|---|---|
| D01/D02 | [同時成功・同ID](../workflows/reservation-to-record-flow.md#checkin-contract)、[確定歴Undo](../workflows/reservation-to-record-flow.md#accounting-finalization-undo-boundary)。受付表示片付けは履歴削除でない | 旧auto-create／cancel-cleanup／自動未来院を最新仕様に混ぜない。[参加規約](../../design/domains/reservations/technical-spec.md#reservation-global-lock-contract)は旧writer／backfillもgate参加または切替停止。unknown確定歴は通常Undo不許可 |
| D03 | [専用intent](../../design/domains/medical-records/technical-spec.md#medical-record-legacy-and-special-contract)の元本文／日付／元先対応・固定版、必須編集ログ、empty/entered/unknown | 取得障害を評価不能手入力へ落とさず、絶対上限超過は解除不可。旧プラン状態／master当時値を推測補完しない。転記／過去入力に通常create在庫減算を発火しない |
| D04 | [安全特記](../domains/owners-pets/lifecycle.md#requirement-fr-owners-pets-safety-and-private-notes)と[参照削除](../domains/owners-pets/lifecycle.md#requirement-fr-owners-pets-delete-dependencies)。飼主向けDTOへ内部特記を漏らさない | 死亡専用writer、カルテ依存拒否、主飼主移管は過去請求／預り金を変えない。全専用FK参照／一般PATCHの共通版は実装適合未確認。依存不明で削除救済しない |
| D05 | [planと記録](../domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-distinct-plans-and-records)、[専用訂正](../domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-special-correction)は原記録保持 | ホテル任意担当を必須へ戻さず、臨床入院開始の医師決定と区別。退院時固定税率／取消会計soft-delete置換はD10目標との差。死亡状態を維持した生前記録・誤記訂正・退院・精算は理由／履歴付き専用経路へ。新規未来の臨床行為を許可せず、死亡状態の汎用PATCH迂回はしない |
| D06 | [安全状態](../domains/trimming/lifecycle.md#requirement-fr-trimming-safety-and-status)、[画像](../domains/trimming/lifecycle.md#requirement-fr-trimming-image-lifecycle)、[訂正](../domains/trimming/lifecycle.md#requirement-fr-trimming-special-correction)。原施術／画像参照・履歴保持 | 患者読込失敗を生存にしない。保存送信済keyは未保存cleanup対象外、参照中削除拒否。完了訂正から自動返金／在庫返却なし。価格時点・取込後変更はD10とのadapterで照合 |
| D07 | [revision](../domains/examinations/lifecycle.md#fr-examinations-revision)、[対応mapping](../domains/examinations/lifecycle.md#fr-examinations-mapping)。原値／単位／未対応item・official版を保持 | 親カルテの診療終了／旧finalizedだけでは通常編集を拒否しない。検査自身のconfirmed／official版の通常上書き禁止と、usage receiptありdetach拒否は維持する。生bytes非保存は保存済み臨床結果の消去ではない。旧completed/revision有無を一律unconfirm可へ変換しない |
| D08 | [訂正保護](../domains/vaccinations/lifecycle.md#fr-vaccinations-correction)、[master／配信](../domains/vaccinations/lifecycle.md#fr-vaccinations-master-delivery)。保存済み各接種IDの履歴保持 | 同日複数を再送重複許可にしない。親カルテの診療終了／旧finalizedだけでは、同じ患者・カルテに対する権限付き通常入力／編集を拒否せず、変更履歴と現在版を保持する。削除・患者変更・死亡後処理の個別保護は維持する。旧snapshot欠損を現masterで補完しない。種filter未実装・既定間隔差を医師の数値承認済みとしない |
| D09 | [正系統](../domains/checkups/lifecycle.md#fr-checkups-canonical)、[結果](../domains/checkups/lifecycle.md#fr-checkups-results)、[保護](../domains/checkups/lifecycle.md#fr-checkups-protection) | 旧seed ID再利用なし、旧field FK欠損も保存snapshot保持。false/0を未入力にしない。Path B後段失敗で前段を自動削除しない。通知失敗から健診再作成なし |
| D10 | [訂正](../domains/accounting/lifecycle.md#requirement-fr-accounting-reasoned-financial-correction)、[元資金](../domains/accounting/lifecycle.md#accounting-rebill-snapshot-payment-source)、[帳票](../domains/accounting/lifecycle.md#requirement-fr-accounting-document-and-aged-debt) | 旧取消後も確定歴・元番号／参照保持。資金元内訳不明なら返金確定停止。受領／充当／復元を重複計上せず、旧請求回収で新売上なし。薬量／臨床確定を精算成功へ転用しない |
| D11 | [提示済み原本（返答待ち含む）](../domains/estimates/business-rules.md#requirement-fr-estimates-state-protection)、[後継](../domains/estimates/business-rules.md#requirement-fr-estimates-successor) | 後継は原本を改変せず、親finalized解除なし。通常CRUD監査best-effortと後継必須監査の差は実装事実として保持。旧ヘッダ見積を自動請求せず確認転記のみ |
| D12 | [原締め](../domains/cash-register/business-rules.md#requirement-fr-closing-append-only)、[実移動](../domains/cash-register/business-rules.md#requirement-fr-closing-movement-consumer) | 原snapshotを再計算／再締めで置換せず訂正差分。実移動event一度消費、原請求日と実日時別。開始残・補充は売上でなく、過不足で架空入金生成なし |
| D13 | [臨床境界](../domains/inventory/business-rules.md#requirement-fr-inventory-clinical-boundary)、[実物調整](../domains/inventory/business-rules.md#requirement-fr-inventory-adjustment) | 取消だけで返却しない。前後数量・理由・実日時・元参照を保持。在庫不足拒否と臨床記録結果を分離。薬剤削除の在庫連動が依存保護済みかは適合確認 |
| D14 | [履歴](../domains/staff-workforce/README.md#requirement-fr-staff-workforce-history-and-access)、[既指名保護](../domains/staff-workforce/README.md#requirement-fr-staff-workforce-booking-contract-index) | 一医院終了と全体退職を区別、過去勤務訂正で臨床／会計変更なし。名前一致統合なし。旧医院能力保存／日跨ぎ／実権限は設計・適合確認 |
| D15 | [file](../line/lstep-integration.md#requirement-fr-crm-shared-file)、[保持分類](retention-audit.md#requirement-fr-common-retention-delivery-copy)と送信log／claim | 原本所在・一時分類不明は143h cleanup対象外。URL期限は原本削除でない。外部成功後log/cache失敗、死亡再読直後の競合を隠さず、絶対遮断／既読保証をしない |
| D16 | [金銭](../domains/analytics/business-rules.md#requirement-fr-analytics-money-separation)、[計数](../domains/analytics/business-rules.md#requirement-fr-analytics-owner-counts) | read-only集計を履歴writerにしない。原請求・実移動・実来院日を保存事実に従って投影。実来院回数は医院内の飼主／日1回、ペット／日各1回とし、旧カルテ記録日数と分離。現SQLと新D10adapterを同一視しない。CSV出力とD15取込認可別 |
| D17 | [医院・法人](../domains/clinic-settings/README.md#requirement-fr-clinic-settings-clinic-directory)、[無効化](../domains/clinic-settings/README.md#requirement-fr-clinic-settings-clinic-retirement) | 現在設定で発行済み帳票snapshotを置換せずD10へ。医院無効化を全未終了業務の新しい強制解消条件にしない。scope=allを全write開放としない |
| D18 | [本人scope](../screens/37-line-reserve-owner-flow.md#requirement-fr-line-channel-auth-scope)、[取消](../screens/37-line-reserve-owner-flow.md#requirement-fr-line-channel-cancel-boundary)、[通知](../screens/37-line-reserve-owner-flow.md#requirement-fr-line-channel-notification-boundary) | 同医院本人owner/pet・死亡を再照合、院内警告overrideをLINEへ適用しない。通知失敗で予約再作成なし。設定／公開保存と外部enableは別 |
| D19 | [scope](../screens/40-identity-links.md#requirement-fr-identity-authorized-scope)、[必要監査](../screens/40-identity-links.md#requirement-fr-identity-audit-atomicity) | hidden/他医院混在は全体拒否、監査失敗rollback。監査はIDのみ・氏名電話なし。unlinkで元台帳・臨床履歴・元資金を削除／移管しない |
| D20 | [限定項目](../screens/38-liff-pet-health.md#requirement-fr-liff-health-limited-fields)、[公開事実](../screens/38-liff-pet-health.md#requirement-fr-liff-health-published-facts) | 保存済み接種の限定表示と親draft条件は最新承認。接種単体finalizedを新設せず、診断・処方・検査全面公開なし。死亡status OR deceased_at除外。本人認証にstaff grantを代用しない |
| D21 | [履歴意味](../screens/39-owner-report.md#requirement-fr-owner-report-history-meaning)、[section認可](../screens/39-owner-report.md#requirement-fr-owner-report-page-section-auth) | 実効的な診療終了結果に対応する通常カルテを対象とし、診療中の途中保存は除外。取得上限・部分失敗を明示。一般履歴はスタッフ閲覧限定で印刷／共有なし。追加採用の予防薬専用処方箋・年間投薬表の印刷だけを別契約とし、全sectionへ広げない。D20公開条件をこの画面の認可へ流用しない |
| D22 | [履歴・復元](../domains/internal-manual/README.md#requirement-fr-internal-manual-version-retention)、[公開](../domains/internal-manual/README.md#requirement-fr-internal-manual-publication) | 旧版保持・復元は新しい版、一般スタッフREADと編集権限別。未保存／未公開案を全スタッフへ出さず、記事を患者原記録の新保管場所としない |
| D23 | [医院mutation](../domains/support/README.md#requirement-fr-support-report-mutation)、[外部送信](../domains/support/README.md#requirement-fr-support-external-send) | 共有READを他医院mutation許可へ拡張しない。本人history削除範囲と共有元への反映を説明し、外部コピー削除を保証しない。送信前確認と院内保持を現通知先行実装から区別 |

## 仕様照合の残差と判定

- 前段階の横断監査ではD21への包括的な印刷認可記述をD07〜D09／C02／X05の閲覧契約へ同期した。今回の予防薬専用2帳票の明示採用だけを限定例外として追加し、一般履歴の印刷／共有は解禁しない。D03・D07自領域の印刷は維持する。
- ホテルの旧必須care_assignment gateは任意担当医師へ修正し、臨床入院開始時の担当医決定条件は保持する。
- D15任意ファイルを全件一時コピーと推定できないため、原本不明cleanupを拒否する選定条件をC05へ具体化した。実装の143h一律対象、分類／元原本証跡の保存対応、実file判定は未適合／未検証。
- 全操作の実schema・lock・監査sink・実grant・旧データ適合はこのリンク表の存在だけで完了としない。各技術正本が示す拒否・不明時停止・成功保存単位を、全writer／旧client／直接writerに実装する後続検証が必要。
- 表に具体条件のない新しいconsumerや操作が見つかった場合は、共通原則だけで補完せず該当owner正本とACへ追記する。仕様不足と実装不足を分離し、未詳査を対象外へ隠さない。

### 自動／手動の実体削除境界（C05）

D06保存済み画像・D15アップロードと手動delete・D23添付について、業務原本／必要履歴／原本所在不明は通常操作で実体を消さず非表示／論理削除とする。未保存作業copyと確認済み一時配送copyは区別する。分類・別原本保持が確認できない対象をcleanupへ入れず、手動deleteにも同じ保護を適用する。DB soft deleteと添付blobの削除は別の副作用であり、前者だけで後者の保持を保証したと判定しない。現在の物理削除経路との差は各ownerの操作／ACへ接続し、未適合を文書完了で隠さない。

2026-10-15の最終独立レビューで、上記23ownerと各正本の文書接続を照合し重大指摘を解消した。明示手動削除の非表示／論理操作と、自動cleanupが原本／分類不明を対象除外し論理状態・実体とも保持する条件を区別する。表の存在だけで完了にせず、各操作本文・AC・技術traceを再確認した。

### 最新第3〜5項の接続

D09の[パッケージ設定／手入力](../domains/checkups/lifecycle.md#checkup-next-date-approved)は設定・保存済み日付・手入力日がなければ空欄とし、一律1年後を補完しない。D08の[過去接種後日入力](../domains/vaccinations/lifecycle.md#vaccination-retrospective-next-date)は実接種日より後の当時の次回日を期限経過後も保持し、通常入力の制約を全面解除しない。D06の[未定の途中保存](../domains/trimming/lifecycle.md#trimming-partial-save-undecided)は担当のみ・コースのみ・両方未定に対応する。終了時は[追加承認](../domains/trimming/lifecycle.md#trimming-end-recorded-performance)の実担当と記録済み施術内容を確認し、コースなしの既存オプション明細を認める。終了条件を途中保存／開始へ流用しない。各ownerのFR／AC／画面・技術traceへ接続し、実装・実機受入とは分ける。

### 追加9点の接続再確認

[追加9点の承認状態](../domains/SPECIFICATION_PROGRESS.md#additional-business-decisions-nine)を各owner本文へ反映し、追加差分の独立内容・構造レビューを通過。D04の重複警告／別登録と移管前影響確認、D05の履歴付きプラン変更、D07担当医任意／入力者ログ、D08次回日設定／手入力、D12実査誤入力訂正、D16実来院日、C03の死亡状態を保つ後日記録・退院・精算、D15の確認付きフォロー通知が対象。元の7群承認を維持する。文書接続の検査と全業務承認の確認を区別し、この表の掲載だけで承認根拠未照合を解消しない。

D04の[連絡先重複](../domains/owners-pets/lifecycle.md#requirement-fr-owners-pets-contact-duplicate-warning)は権限内候補と別owner登録を許す条件で、同一人物・自動統合・receiptの証明ではない。[主飼主影響確認](../domains/owners-pets/lifecycle.md#requirement-fr-owners-pets-primary-transfer)では未来予約とLINE通知先を一覧確認し、保存前変化を再確認する。未連携の自動リンクと元飼主の過去資金移管を追加しない。


### 後続第1・2・6項の接続

D01／D18の[分類別・担当者別開始時刻](../line/reservation-spec.md#requirement-fr-line-reservations-business-start-times)は、通常診療の営業時間生成とトリミング登録時刻限定を分離し、時刻変更・担当者別設定・既存予約保護を扱う。[指名なし](../line/reservation-spec.md#line-no-nomination-unassigned)は担当者別設定と既存能力制約で判定しても実担当未定を保持し、仮候補を保存・表示・通知上の実担当にしない。D10の[停止支払方法](../domains/accounting/lifecycle.md#accounting-inactive-payment-method)は新受領のみを制限し、D12／D16の過去支払履歴・集計を保持する。旧資金の再配分・預り充当／解除・返金へ新受領gateを一律適用しない。各ownerのID・版・認可・保存／結果照会と接続し、全過去業務承認や実装適合の網羅確認を表さない。文書候補の状態は[全体進捗](../domains/SPECIFICATION_PROGRESS.md#additional-business-decisions-six-followup)へ。

## 2026-10-15会議合意の接続：trimming-4

D06は型付き施術record ref＋同じappointment ID／版を保持し、開始後はD02診療中列、専用画面終了は同じ予約の会計待ちへ。会計待ちから既存記録を開き、リンク不明はowner確認で停止、推測・架空予約なし。終了receiptと記録保存／請求・入金は分離する。コース／option料金の手修正・行値引きをD10へpullし、元値・調整値・対象scope／版・計算根拠を保持。0円実在行を落とさず、二重適用・master／確定請求の自動上書きなし。追加後は差分確認・明示再適用、確定後は既決訂正へ。

## 2026-10-15会議合意の接続：shifts-3

D14の複製／CSV入力は[新操作の意味契約](../../design/domains/staff-workforce/technical-spec.md#shift-bulk-operation-contract)へ。元／先の医院・スタッフ・期間・勤務ID／旧値、確認差分、現在権限・期待版・保護予約を拘束する。勤務／休憩と必要履歴・監査は一回に確認して確定する範囲全体で原子的に保存し、結果不明は元操作照会、部分結果は対象別に保持する。明示選択した置換だけを対象とし、エラー／未解決衝突は確認範囲全体を止める。範囲選び直しは再プレビュー・再確認とし、自動部分採用・無断上書き・同一効果の再実行を許可しない。CSV形式・一貫性制御・照会wire等の具体設計と実装適合は別工程。

## 2026-10-15の通常編集と復元操作の受け手

[D03](../domains/medical-records/lifecycle.md#medical-record-agreement-autosave-20261015)のタブクリア／複製取消は、現在医院・患者・対象タブ／記録・元操作と基準版・現在版を照合する履歴付き復元。既に保存済みの変更も対象になり、元版を消して戻さない。D07確定結果、D10資金・確定請求、D11見積原本、D13実物入出庫、D15送信済通知は逆実行しない。D08／D09現在版の更新、D20限定表示、D21終了履歴の各consumerは自身の認可・保存版を再取得して参照する。

## 2026-10-15 D03復元snapshotの共通対応

[最新D03技術](../../design/domains/medical-records/technical-spec.md#medical-record-autosave-revert-contract)のタブ開始snapshot／複製直前snapshot、元操作・成功subset・期待版・復元receiptを照合する。旧finalizedだけの通常編集拒否を置換し、元保存版・同ID・既存追記／監査を保持。別ownerの実行や他者版を無警告で巻き戻さず、保存§2の個別送信先と必須ログを維持する。

<a id="joint-walkthrough-owner-boundaries-20261015"></a>
## 共同確認した追加境界（2026-10-15）

- D01／D02の[ペット別実受診歴と次回予定](../workflows/reservation-to-record-flow.md#next-visit-intent-agreement-20261015)は、候補とスタッフ選択、未予約の日付、明示消化、取消後の明示再設定を区別する。D04登録モード・会員区分だけで初診／再診や予約可否を決めない。
- D03の[第1・第2各複数診断／時刻付きバイタル履歴](../domains/medical-records/pages/medical-record-form.md#medical-record-agreement-diagnosis-vitals-20261015)は、全診断の区分・元測定／実測日時を既存の参照・印刷へ渡す。旧単件型で切り捨てず、飼主向け公開や医学的閾値を追加しない。
- D04の[会員区分・DM希望](../domains/owners-pets/lifecycle.md#requirement-fr-owners-pets-membership-dm)は独立値。退亡者は死亡ではなく、区分変更だけで料金・予約受付・LINE可否を自動変更しない。D15の[目的別配信](../line/lstep-integration.md#crm-dm-purpose-contract)は未設定／不要の案内・販促を抑止し、予約確認・変更・予約リマインドにも現在の本人／宛先・独立停止等の安全条件を適用する。
- D11の[提示価格と原本保護](../domains/estimates/business-rules.md#estimate-price-presentation-agreement-20261015)は、承認済み・期限内・同内容数量の単価と、提示済み返答待ちを含む原版を保持する。D10の税・割引・保険確認、確認転記、実授受とは別。D12の[締め区分と現金差額](../domains/cash-register/business-rules.md#closing-segment-difference-agreement-20261015)は締めた区分だけを保護し、非0差額はメモ付きで残す。原因不明可とし架空入出金を作らない。

今回の合意と独立レビュー済み文書の接続を示す表であり、全既存業務の共同確認完了ではない。文書整合・独立レビュー・公開結果・実装適合はそれぞれ[全体進捗](../domains/SPECIFICATION_PROGRESS.md#joint-walkthrough-batch-20261015)で区別する。

### 追加採用の検査・初診問診・予防帳票の接続

[D07の保存結果・選択印刷・感染症投影](../domains/examinations/lifecycle.md#fr-examinations-infection-summary)では、D04／D03患者トップが同じ固定結果版・項目ref・検査日・出典・訂正／取得状態を読む。台帳PATCHや問診自己申告で原結果を上書き・確認済みに昇格させず、陰性／0・未実施・結果待ち・不明／取得失敗を区別する。D20へ公開しない。D03は初診問診と予防薬記録のowner、D21は既存レポート入口から許可された専用2帳票だけを参照・印刷する。印刷による処方、投薬、会計、通知の副作用を追加しない。

## 原資料追加採用6点のowner接続

[D14職員](../domains/staff-workforce/README.md#staff-source-adoption-20261015)の属性と休業期間はD01の新規指名候補／既存予約保護へ渡し、認可roleや既存予約を自動変更しない。[D10月次未納](../domains/accounting/pages/unpaid-list.md#unpaid-monthly-source-adoption)は元請求・実入金・預り金充当／解除・訂正を区別し、2026-10-16承認の実受領のみの当月入金と別掲調整内訳を受け渡す。元ID／発生日・符号付き残高影響とcutoffを揃え、充当と受領を重複させない。実返金の新規方針・税額変更は含めない。[D12日次内訳](../screens/29-closing-aggregation.md#closing-source-breakdown)はD10の保存済み会計額から参照し、税再計算writeを行わない。[D03薬量](../domains/medical-records/lifecycle.md#dose-source-adoption-20261015)は承認済み単純計算範囲と今回選択版を保持し、原資料OCRや一般yesを臨床値承認へ変換しない。

今回の追加ACは12、既存FR／ACの保護を継承する。新しい全owner再レビュー完了を意味しない。
