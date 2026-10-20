# カルテ詳細・入力 仕様書 (Medical Record Form)

2026-10-15の[診断2区分の各複数登録・最新バイタルの測定時刻と履歴](#medical-record-agreement-diagnosis-vitals-20261015)を追加合意として反映する。以下の自動保存・終了後通常編集・復元保護は維持し、旧単件診断APIと時刻なし表示は現実装差として扱う。実装・実機受入済みを意味しない。

<a id="medical-record-form-autosave-20261015"></a>
## 最新の操作契約（2026-10-15）

[通常編集・自動保存維持・タブクリア／複製取消・終了・履歴](../lifecycle.md#medical-record-agreement-autosave-20261015)が最新正本。医師確定ボタンを設けず、診療終了後も同じ通常カルテを編集できる。保存ボタン方式への変更は撤回され、従来のhybrid保存を維持する。以下のfinalizedロック・確定ダイアログは現実装の差分であり、現在の通常編集を制限する目標ではない。保存済み版・追記・必要監査、患者訂正や実施取消等の目的別保護は維持する。

| 操作 | 基準・結果 | 保護・失敗 |
|:---|:---|:---|
| 入力・既存保存 | 保存§2の自動／即時保存・アクティブタブ保存を維持 | 成功／送信中／失敗／不明を分離。通常保存に医師確定を移植しない |
| 入力内容クリア | 操作したタブを開いた時点へ戻す。最後の自動保存値を基準にしない | 他タブ不変。自動保存済みなら復元として保存・履歴を残す。他者版／後続作用は無警告で巻き戻さない |
| 履歴複製の取消 | 対象複製の直前へ戻す | 逐次保存の成功分・失敗／不明を照合。元記録や対象外変更は不変 |
| 診療終了 | 同じ予約を会計待ちへ。未実施／後日予定は終了時に警告し確認後終了可 | 記録・請求・入金確定なし。警告取消／取得不明は終了成功としない |
| 会計待ちから開く | 同じ予約に明示リンクされた既存カルテを開く | 新規作成／複製／再開始なし。リンク不明は理由表示・再照合 |
| スタッフ向け履歴 | 診療終了したカルテを対象に表示 | 診療中の途中保存を載せず、finalized／入金で代用しない |

クリア／複製取消の表示は対象タブ又は複製操作・復帰基準を識別できるものとする。復元中／結果不明を取り消し済みと表示せず、再読込後も成功した復帰状態を確認する。欄自体の空欄化（例：主訴区分null）、既存明細の削除、予約終了Undo、返金・在庫戻しとは異なる。詳しい型・競合・履歴は[技術契約](../../../../design/domains/medical-records/technical-spec.md#medical-record-autosave-revert-contract)へ。

[17:17合意の診療前日付訂正](../../reservations/pages/reservation-management.md#requirement-fr-reservations-preconsultation-date-record-correction)では、未記入カルテは確認して予約日へ合わせ、記入済みなら通常の日付変更を停止して内容確認の訂正へ渡す。以下の作成後日付固定・直接PATCH拒否は現行実装／通常更新の参照で、この合意の未記入・診療前の確認付き訂正を実装済みとする根拠ではない。保存ルーティング契約と予約訂正・履歴保護は維持する。通常カルテのfinalizedだけを理由とした編集禁止は[最新合意](../lifecycle.md#requirement-fr-medical-records-normal-edit-after-end)へ置換する。

17:23 JSTの[LINE予約6点合意](../../../line/reservation-spec.md#line-agreement-1723)へ従う。公開された診療・トリミング・ホテルは予約対象、入院は医院相談。ホテルには医師指名・医師シフト・通常予約の同日制約を適用しない。受付制限・入力保持・受付前取消・必要設定不足の扱いと受入条件はLINE正本を参照する。ホテル受入基準は17:35合意の医院別同時宿泊頭数・全宿泊期間判定で、実装・実機受入は未確認。

17:17 JSTの院内予約8点一括合意は[予約入口の決定一覧](../../reservations/README.md#internal-reservation-batch-1717)を参照する。誤取消／誤終了訂正、未終了予約保護、診療前の日付・カルテ整合、マスタ利用・名称履歴、休診日予約／日跨ぎ勤務、患者照合、ホテル担当時点／入院記録失敗の各正本と受入条件へ従う。LINE受付条件は17:23合意（ホテル受入基準は17:35合意、計算・保存方式は設計）、実装・実機受入は未確認。

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | page.medical-records.medical-record-form／画面仕様 |
| 決定状態 | 2026-10-17追加決定は末尾と進捗入口を参照。従来：2026-10-14初回5項目は部分合意。最新の業務正本はlifecycleの2026-10-15合意。初回合意の非変更部分は維持。対象外の既存決定状態を保持し、カルテ全体完了／実装受入にしない |
| 実装状態 | 合意と現実装は別。main ec4f7146a の必要箇所を静的照合し、差分は予約正本の残差一覧へ記録。実装変更なし |
| 実機検証状態 | 今回未実施。過去の記録と今回の文書確認を分ける |
| 文書確認日 | 2026-10-15（診断・バイタル追加合意の範囲で合意・参照・語彙を確認。実機受入日ではない） |

<a id="purpose"></a>
## 目的・対象範囲

| シナリオ | 業務上の価値 | 優先度・理由 |
|:---|:---|:---|
| 主訴区分を空欄で保存する | 任意項目を無理に選ばず記録する | 未決／要確認（既存文書に優先度の根拠なし） |

### 概要

初診のwelcome／follow-upは[D15の確認付きフォロー](../../../line/lstep-integration.md#crm-followup-confirmation)に従い、カルテ作成だけでは送信しない。スタッフが実施済み・送信対象を確認し、一般のお礼と医療的再検査／受診案内を区別する。後者は医師の要フォロー指定がある場合だけとし、担当医選択・カルテ保存／確定をその指定の代用にしない。現行create副作用の実装説明はこの目標への適合証明ではない。既存保存§2・認可／必要監査と結果不明照合を保持し、通知失敗からカルテを作り直さない。受入はD15のAC-crm-followup-record-only／medical／confirmation-failure／duplicate-unknownへ接続する。

- **画面の目的**: 臨床現場における電子診療録（SOAPS形式）の作成、検査、処置、処方、および会計連携の統合管理。

<a id="access-preconditions"></a>
## 利用者・権限・前提

- **アクセス権限**: 新規作成 (`/medical-records/new`) は `medical-records:create` を要求する `RequirePermission` ルートガードあり。編集 (`/medical-records/:id`) は親ルート `/medical-records` の `medical-records:view` ガードのみを継承し、保存可否はコンポーネント内の `usePermission`（`ResourceMedicalRecords`）で制御。

担当医以外の権限あるスタッフの入力、医院が認めた医師の目的別確認・訂正確認と実編集者ログは[初回合意](../lifecycle.md#requirement-fr-medical-records-authorized-editor-audit)。実grantと医師判定の対応は要確認。操作前提の詳細は本書の各操作・保存契約と関連正本を参照する。

<a id="navigation"></a>
## 入口・出口

- **URLパターン**: 
  - 新規作成: `/medical-records/new?petId=xxx`
  - 編集: `/medical-records/:id`

| 入口／出口 | 契機 | 引継ぎ情報 | 根拠 |
|:---|:---|:---|:---|
| URLパターン | 画面を開く | 上記URL／既存導線。呼出元の網羅は要確認 | 本節の元記載 |
| 戻り先・他画面 | 本節の既存操作 | 記載が無い引継ぎ・戻り先は要確認 | 本節／関連フロー |

診察予約の通常カルテは[受付時の作成契約](../../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-checkin-record-atomicity)へ変更する。既存の新規画面表示・保存操作は実装参照を含み、合意への追従は未確認。予約変更は[予約管理へ集約](../../reservations/pages/reservation-management.md#requirement-fr-reservations-edit-entry)し、カルテから予約を変更しない。通常新規と過去未登録診療は[初回合意の受付接続／専用入力](../lifecycle.md#requirement-fr-medical-records-reception-and-retrospective-entry)へ従う。ホテル・入院・トリミング専用記録は別ownerの内部仕様。

### 死亡登録後の生前診療記録

[死亡後適用の最新合意](../lifecycle.md#medical-record-deceased-retrospective)に従い、生前診療の後日入力・誤記訂正を死亡状態を戻さず理由と履歴付きの専用経路で扱う。既決の過去入力・医師確認・当日カード／自動課金なしを維持し、死亡後の新しい診療や未来予約／実施は許可しない。通常フォーム／選択画面の死亡拒否は現行差として区別し、専用UI・保存対応を実装済みとしない。退院・精算はD05/D10へ渡す。

<a id="display-inputs"></a>
## 表示情報・入力項目

| 項目 | 意味 | 型／形式 | 必須／任意 | 初期値 | 候補元 | 制約／根拠 |
|:---|:---|:---|:---|:---|:---|:---|
| 主訴詳細・治療方針・診断 | 問診・SOAPSの記録 | 未記載／要確認 | 未記載／要確認 | 未記載／要確認 | 入力・履歴複写 | 下記「SOAPS 形式の採用」、[保存契約](../lifecycle.md#hybrid-sync) |
| 主訴区分 | 主訴の分類 | nullable（*uint64） | 任意・空欄可 | 未記載／要確認 | 主訴区分マスタ | [画面保存操作](#22-保存プロセス) |
| 担当医・来院種別・診療日 | 患者ヘッダーの記録属性 | 未記載／要確認 | 未記載／要確認 | 未記載／要確認 | 既存値、候補の詳細は要確認 | [画面保存操作](#22-保存プロセス) |
| 次回予定 | 日付のみの未予約意図 | intended_date（日付） | 未記載／要確認 | 未記載／要確認 | 次回予定ボタン | [次回予定契約](../../../workflows/reservation-to-record-flow.md#next-visit-intent) |
| 画像・資料 | 診療資料の添付 | JPEG／PNG／GIF／PDF | 未記載／要確認 | 未記載／要確認 | ファイル選択 | 下記「画像・資料の選択制限」 |
| 訂正内容・訂正理由 | 既存追記／目的別訂正の入力。通常本文は終了後も通常編集可 | 未記載／要確認 | 未記載／要確認 | 未記載／要確認 | 入力 | 下記「臨床安全ガード」（理由500文字以内） |
| 治療明細・投与逸脱理由 | 処置・処方の記録 | 未記載／要確認 | 逸脱理由は条件付き必須 | 未記載／要確認 | 医院マスタ・履歴複写 | 下記「臨床安全ガード」、処置マスタの節 |
| 他タブの入力 | 予防接種・健診・検査・見積・会計確認 | 未記載／要確認 | 未記載／要確認 | 予防接種は種類別設定または手入力へ統一（具体間隔・初期値は追加確定しない）。現行mainのカルテ・独立画面は1year既定へ統一済みだが、目標の固定既定を採用した意味ではない。健診はパッケージ設定候補または手入力、設定・保存済み日付・手入力日がなければ空欄。他は要確認 | 下記タブ表の関連仕様 | 独立画面・API定義を複製しない。詳細は既存関連仕様／要確認 |

### 1. 画面構成（問診を治療プランへ統合）

問診は治療プランへ統合する。以下は統合後のタブ責務であり、表示順の番号は今回確定しない。

| タブ | 名称 | 内容 |
|:---:|:---|:---|
| 統合 | **治療プラン** | スタッフ問診・主訴・症状・生活環境と、別区画の飼主事前問診、移設した右側の過去カルテ履歴。身体検査所見（SOAPのS/O/A）、診断名（第1・第2を保持し、それぞれ複数登録）、今後の治療方針（Plan）。 |
| 維持 | **治療** | 実施した処置、処置項目、および処方薬の明細。 |
| 維持 | **予防接種** | [今回プランの未実施予定と同一接種追跡](#requirement-fr-medical-records-plan-domain-tabs)を表示する採用目標を追加（実装未確認）。既存の保存済み履歴は左ペインに当該ペットの接種記録一覧（`GET /vaccinations?pet_id=` + `page`/`limit`）。空なら EmptyState、「記録を追加」で入力フォーム。右カラムは `VaccinationHistory`。カルテネストの `/medical-records/:id/vaccinations` は存在せず使わない。次回日は[D08最新合意](../../vaccinations/lifecycle.md#fr-vaccinations-schedule)の種類別設定または手入力へ統一する。現行mainはEMR-430により `calculateNextDate` 配線の既定がカルテ・独立画面とも `1year` に統一された。これは実装観察で、種類別設定／手入力の目標に対する固定既定の採用ではない。[過去接種の専用後日入力](../../vaccinations/lifecycle.md#vaccination-retrospective-next-date)では、実接種日より後の当時の次回日が既に過ぎていても保持する。通常入力の制約全面解除はしない。 |
| 維持 | **定期健診** | [今回プランの未実施予定と同一健診追跡](#requirement-fr-medical-records-plan-domain-tabs)を表示する採用目標を追加（実装未確認）。既存の健康診断の結果記録を維持。`/checkups` の編集は `?tab=定期健診&checkupId=` で本タブを開き、該当行を編集状態にする。[最新第3項合意](../../checkups/lifecycle.md#checkup-next-date-approved)により、次回日はパッケージ設定候補または手入力。設定・保存済み日付・手入力日がなければ空欄とし、一律1年後を補完しない。 |
| 維持 | **検査** | [今回プランの未実施予定と同一検査追跡](#requirement-fr-medical-records-plan-domain-tabs)を表示する採用目標を追加（実装未確認）。既存の検査結果の一覧表示と既存検査の取込（カルテへの紐付け）、基準値との比較（HIGH/LOW判定ハイライト）。[D07最新合意](../../examinations/lifecycle.md#fr-examinations-record)に従い担当医未選択でも保存可、入力者ログは担当医と別に保持する。 |
| 維持 | **画像** | 患部写真、レントゲン、エコー画像、PDF資料のファイル選択によるアップロード・管理。ドラッグ&ドロップ追加は[追加3項目合意](#requirement-fr-medical-records-image-drop-selected-print)で業務採用。#235の現実装・価値実測の状態は今回更新／確認していない。 |
| 維持 | **見積書** | 提示した概算費用の管理。 |
| 維持 | **会計(医師確認)** | 診療費の最終確認と会計ステータスの送信。 |

#### 1.1 スティッキー患者ヘッダー (Sticky Patient Header)

タブの上部には `MedicalRecordStickyHeader` が画面上部に sticky 固定で常時表示され、内部の `PatientInfoCard` が患者コンテキストを保持する。タブ切替・スクロールに関係なく以下をクリック不要で確認できる。

- **今回カルテのバイタルチップ**: [最新合意](#requirement-fr-medical-records-vital-measured-time-history)に従い、体温 T / 心拍数 HR / 呼吸数 RR / 測定体重（単位付き）の最新値を、それぞれの値に対応する測定時刻とともにペット名行直下へ常時表示する。同日を含む以前の測定は履歴から確認できる。測定日も識別できる表示とし、欠測／未取得の値・時刻を補完しない。取得済みで当該カルテの測定がない、又は4項目すべて未入力の場合は値チップを表示しないが、取得失敗・不完全を測定なしへ置換しない。現行mainの `useGetVitals` → 最新1件を使う `latestVisitVitalChips` → `vitalsSummary` は `PatientInfoCard` の `petAttributes` に時刻なしで表示する実装で、測定時刻・全履歴は未追従差分。[技術契約](../../../../design/domains/medical-records/technical-spec.md#medical-record-diagnosis-vitals-contract)で照合する。

現行mainの配線では、カルテ固有属性を `PetAttributesLine` とバイタルから組み立てて `PatientInfoCard` の `petAttributes`（ReactNode）へ渡し、`vitalsSummary` propsをカードへ直接渡さない。登録体重の未登録／空は `-` とし、今回測定体重の代用にしない。担当医・来院種別は `hideStaff`／`hideReservationType` でカード内を隠し、カルテ側の専用操作として扱う。これらは実装対応の記録で、旧9タブの維持や計測時刻なし表示の目標採用ではない。
- **マイクロチップ番号**: ペットの `microchip_number`（props 名 `microchipNumber`）が存在する場合、ペット名の右隣にモノスペースのチップとして表示する（`aria-label` は「マイクロチップ番号 <値>」）。未登録または空白のみの場合はチップを表示しない。
- **次回予定ボタン（`NextVisitButton`、EMR-308）**: ヘッダーの「次回予定」はペットの tentative appointment（`status='tentative'`、日付のみ `intended_date`）を `PUT /v1/reservations/pets/:petId/next-visit-intent` で upsert、`DELETE` 同パス（`reservations:edit`）で取消する。カルテ PATCH ではない。権限は操作別（新規=`reservations:create`・既存 tentative の上書き=`create`+`edit`・確定=`edit`）で、write 契約の正本は [reservation-to-record-flow.md](../../../workflows/reservation-to-record-flow.md) §2.3、一覧・確定の管理面は [reservation-management.md](../../reservations/pages/reservation-management.md) を参照。 日付だけの予定は枠を確保せず受付ボードへ出さない。別件来院では自動消化せずスタッフが今回消化を選び、由来予約の取消では自動復活しない。必要時の手動再設定・安全な保存条件は[2026-10-15合意](../../../workflows/reservation-to-record-flow.md#next-visit-intent-agreement-20261015)へ従う。

[Backlog Phase1 F33:G33（No.32）](https://docs.google.com/spreadsheets/d/17jaFBv7NmbPzk88lZx94vFxa_G1shGNX2O_lFGt4QdQ/edit#gid=0&range=F33:G33)の回答に対応し、画面を開いた時点で飼主名・ペット名／No・生年月日／年齢・種別・性別／避妊去勢・犬種・体重・通算来院回数・保険情報を患者プロフィールとしてまとめて確認できる。体重の測定元／日時、来院回数の実来院基準、保険情報と会計の資格／適用確認の区別はそれぞれの既存正本へ従い、未取得・不明値を推定しない。

同じ飼主の同居ペットもカルテ上部に表示し、選択した子のカルテへ遷移できる。D21のレポート内ペット切替だけでこの導線を代替しない。現在の閲覧認可・医院／患者scopeと未保存入力の離脱保護を維持する。対象記録の閲覧から新規カルテ・受付・診療開始を自動作成せず、リンク不明／取得失敗は対象確認へ渡す。回答日は原セルに記載がなく、今回の実機受入も未実施。

| 条件ID | 前提／操作 | 期待結果 | 検証状態 |
|:---|:---|:---|:---|
| AC-medical-records-patient-profile-context | 通常カルテを開く | 上記プロフィールをまとめて参照し、測定元・実来院・保険表示の既存意味と不明状態を保持する | 未実施 |
| AC-medical-records-household-pet-navigation | 同じ飼主の同居ペットを選択し、未保存入力／権限不足も比較する | 対象の認可されたカルテへ移動でき、離脱保護・患者照合を保つ。表示／選択だけで受付・記録作成・診療開始を行わない | 未実施 |

<a id="requirement-fr-medical-records-input-wrap-drug-search"></a>
#### FR-medical-records-input-wrap-drug-search — 入力表示と薬名検索

長い入力内容は入力枠の幅に応じて自動折り返し表示する。表示の折り返しを保存本文への改行挿入や文字の切捨てにしない。薬名はひらがな／カタカナのどちらでも検索できる。[Backlog Phase1 No.29 F30:G30](https://docs.google.com/spreadsheets/d/17jaFBv7NmbPzk88lZx94vFxa_G1shGNX2O_lFGt4QdQ/edit#gid=0&range=F30:G30)、2026-10-15曽我回答が根拠。英語あいまい検索・頻度順ソートは今期見送りであり、この検索条件へ追加しない。

<a id="requirement-fr-medical-records-patient-history-search"></a>
#### FR-medical-records-patient-history-search — このペットのカルテ内履歴検索

対象ペットのカルテ内で治療内容・処方薬名・処置名を検索でき、例えば「OPE」で該当する履歴の日付と内容を確認できる。[Backlog Phase1 No.31 F32:G32](https://docs.google.com/spreadsheets/d/17jaFBv7NmbPzk88lZx94vFxa_G1shGNX2O_lFGt4QdQ/edit#gid=0&range=F32:G32)、2026-10-01の具体回答と10-15の再確認が根拠。カルテ一覧全体の検索述語とは別の契約で、一覧検索の対象外項目を理由にカルテ内履歴検索を省略しない。医院横断の表示は現在の閲覧認可と医院フィルタに従い、同じペット以外の履歴を混入しない。検索から記録・受付・請求を作成しない。旧UI再現・類似音検索・アニコム連携は今期見送りのまま。

| 条件ID | 前提／操作 | 期待結果 | 検証状態 |
|:---|:---|:---|:---|
| AC-medical-records-input-wrap | 長文を入力して枠幅を変える | 枠内で折り返し、保存本文の改行・内容は表示折り返しだけでは変わらない | No.29。未実施 |
| AC-medical-records-drug-kana-search | 同じ薬名をひらがな／カタカナで検索する | 双方で該当する薬を見つけられる。英語あいまい検索・頻度順ソートを今期条件に加えない | No.29。未実施 |
| AC-medical-records-patient-history-search | 対象ペットに薬名・処置名・治療内容の履歴と別ペットの履歴がある | 薬名／OPE等で当該ペットの該当履歴の日付・内容を参照できる。医院認可・絞込みを維持し、一覧検索と混同しない | No.31。未実施 |

#### 2.1 SOAPS 形式の採用

- **Subjective (主観的)**: 飼主の話。
- **Objective (客観的)**: 視診・触診・聴診結果。
- **Assessment (評価)**: 診断、病名特定。
- **Plan (計画)**: 今後の検査・治療予定。
- **Supplement (補足)**: 注意事項等。

#### 2.4 画像・資料の選択制限

ファイル選択欄の対象はJPEG/PNG/GIF/PDF。一度の選択は10件まで、1ファイル10MiB以下・合計50MiB以下とする。件数・合計サイズの超過、または1件でも個別サイズ超過がある場合、その選択分全体を受け付けずエラーを表示する（`ImageGalleryFilter`）。拡張子・形式の選択補助はファイル内容の安全性検証を保証しない。

#### 2.5 処置マスタのプロビジョニング（皮下点滴）

「治療」タブで選択する処置項目（皮下点滴を含む）は `procedures` テーブルの `clinic_id` 付きレコードであり、**医院ごとのランタイムマスタデータ**として管理される。seed 配布物（backend/migrations/seeds/002_master に procedures の CSV は存在しない）やコード内の静的定義は存在しない。各医院はマスタ設定画面 `/settings/treatment-items` の「処置」タブ（`?tab=procedure`、表示名「処置マスタ」。`ResourceMasterMedical` 権限で保護された `TreatmentPlanMaster`）から `useCreateProcedure` 等で登録・編集する。皮下点滴もこの運用経路で医院ごとに追加するものであり、デプロイ物・seed 変更を伴う導入経路は存在しない。

<a id="operations"></a>
## 操作・業務ルール・状態遷移

| 要件ID | 定義の所在 | 決定根拠 |
|:---|:---|:---|
| FR-medical-records-optional-chief-complaint-type | [既存の要件本文](#requirement-fr-medical-records-optional-chief-complaint-type) | 既存記載を継承。ID付与で合意済みにしない |
| FR-medical-records-diagnoses-multiple-per-category | [第1・第2ごとの複数診断](#requirement-fr-medical-records-diagnoses-multiple-per-category) | 2026-10-15、Sentinel_4ec55f21a17c819199d16ad1add270a4 の回答1 |
| FR-medical-records-vital-measured-time-history | [最新値・測定時刻と測定履歴](#requirement-fr-medical-records-vital-measured-time-history) | 同回答2 |

| 操作 | 実行条件 | 成功後状態／遷移 | 失敗挙動 | 根拠 |
|:---|:---|:---|:---|:---|
| 保存・クリア・複製取消・訂正 | 操作権限、対象医院／患者・版、目的別確認 | 保存・基準への復元・必要履歴 | 版競合／影響／結果不明は照合し成功を偽装しない | 最新操作契約・保存節 |
| 履歴複写 | コピー可能行・操作権限／対象照合 | 問診ローカル変更／治療明細の個別保存、操作単位の取消が可能 | 上書き確認取消は現在値保持。複製取消は直前版を照合 | 下記「履歴複写」・最新操作契約 |

### 2. 主要な臨床ロジック



#### 2.3 臨床安全ガード

最新の[入力者・編集ログ](../lifecycle.md#requirement-fr-medical-records-authorized-editor-audit)、[必須項目追加なし／未完了プラン警告](../lifecycle.md#requirement-fr-medical-records-explicit-finalize-plan-warning)、[削除・訂正](../lifecycle.md#requirement-fr-medical-records-empty-delete-reviewed-correction)、[評価不能時の医師確認・理由付き手入力](../lifecycle.md#requirement-fr-medical-records-unevaluable-dose-manual)は初回合意が正本。未完了警告は最新合意により自動保存時でなく診療終了時に表示し、確認後は終了可能。以下の確定UI・ロックAPI説明は旧実装参照で、最新合意への追従済みを表さない。

- **旧実装参照：確定ロック（最新の通常編集では置換）**: `finalized`（確定済）ステータスのカルテはバックエンドが更新を拒否し（409）、訂正は追記（addendum）のみ許可することで真正性を担保。確定への遷移は `PATCH /medical-records/:id` の `status` 指定によるもの。画面右下のフローティングアクション（`MedicalRecordFloatingActions`）に「確定する」ボタンが表示され（編集権限あり・保存済み・未確定の場合のみ）、会計(医師確認)が `confirmed` の場合だけ有効になる。未確認・差戻し・取得中・取得失敗・確認状態なしでは「会計確認が未完了です」として確定を物理ブロックする。有効時は `MedicalRecordFinalizeDialog` で不可逆であることを確認した上で確定する。確定取り消し（unfinalize）API は存在しないため、確定後の修正経路は訂正追記（addendum）のみ。会計完了時の自動確定は現状存在しない。確定済みカルテはサイドヘッダーに「確定済」バッジ（`StatusBadge`）を常時表示する。
- **旧実装参照：ロック中の画像タブ（EMR-216）**: `isLocked`（`recordFinalized || !canSubmit`）は `MedicalRecordServiceTabs` から `MedicalRecordImage` へ伝播し、ロック中は `canUpload`/`canDeleteImage` が false になる。「画像アップロード」「撮影」ボタンと file input を描画せず、ドラッグ&ドロップはドロップガイドを出さず `dropEffect=none` で拒否し、画像の削除ボタンも出さない。UI 非表示に加え callback（`handleFilesSelected`/`handleDrop`/`handleDeleteRequest`/`handleDeleteConfirm`）側でも fail-closed に拒否するため、ロック中の画像 mutation API 呼出はゼロ（外側の disabled fieldset は D&D イベントを遮断できないため、この二重防壁が必要）。画像一覧・画像リンクの閲覧は維持され、検索・期間フィルタ等のフォーム入力はタブ共通の disabled fieldset 配下で他タブと同一の境界規則に従う。ロック解除で全操作が復帰する。
- **訂正追記モーダル**: `AddendumModal` の修正内容・修正理由は controlled input。バリデーション失敗後も入力済みの値を保持する（React 19 `useActionState` の remount で消えない）。修正理由は 500 文字以内。
- **薬量自動計算と絶対上限ゲート（#201）**: 「治療」タブの処方明細（`TreatmentRow`）は、対象ペットの species と当日 vital 体重から数量を自動プリフィルする（`calculateDose`）。保存値がマスタ上限（体重連動上限 weight×max_mg/kg と体重非依存の絶対上限 absolute_max_dose の小さい方）を超える場合、フロントエンドは理由をインライン表示して追加・更新を送信せず、バックエンドも Create/Update の永続化前に 400 で拒否する。`ConfirmDialog` による解除経路は設けない。上限内でも下限未満または推奨値からの著しい乖離は、インラインの逸脱理由（`dose_deviation_reason`、空白のみ不可・500文字以内）が必須。理由を snapshot と同一 transaction の監査へ保存し、actor・監査依存の欠落や監査失敗時は保存しない（`computeDoseGate` / `ensureDoseDeviationAuditReady` / `auditDoseDeviationTx`）。現行実装は体重未記録・species 正規化不能・投与量パラメータ未設定時に評価をスキップして保存を継続する参照で、最新合意の医師確認・理由・編集履歴を満たす証拠ではない。最新契約は評価不能を明示し医師確認・理由付き手入力を認める。パラメータ取得の非 NotFound エラーと species 不一致は既存どおり fail-closed とする。権限付き例外フロー（Design B）は実装しない。
- **臨床承認の境界**: コード上のゲートと、対象薬・値・範囲・単位・出典を含む臨床承認は別。#201 の Closed は #261 への承認作業の集約であり、医師による臨床 bundle の承認完了を意味しない（[承認作業 #261](https://github.com/MinoruSoga/AnimalEkarte/issues/261)）。
- **未保存警告**: 変更がある状態でページを離れようとすると `NavigationBlocker` が警告を表示。
- **特記マーク（EMR-173/231）**: スティッキーヘッダー内の共有患者カード（`PatientInfoCard`）に共有 `DangerBadge` で表示: ペット名横に `danger_level` 高=赤い八角形アイコン・中=黄い三角形アイコン（文言なし、Popover で補足メモ）、低・未設定は非表示。飼主 `is_dangerous` は飼主名横に文言なしの赤い八角形アイコン（代替名「特記」）。医療記録中もマークを常時視認できるようにするための安全表示。


<a id="medical-record-agreement-input-four"></a>
## 入力まわり4項目の合意（2026-10-14）

質問 Sentinel_374bdacb47a0819180046f40e27a41dc と回答 Sentinel_6287d7cbcf988191aac47fbe504cfad1「ok／4 基本的にはチェックボックスとモーダルで実施済に切り替える」が根拠。1〜3の提案を承認、4は具体的操作指定を優先する。これは部分合意で、[初回5項目](../lifecycle.md#medical-record-agreement-first-five)・保存§2・予約／受付接続・編集ログ・目的別医師確認・評価不能時手入力を保持し、実装／API／実機受入済みとしない。

<a id="requirement-fr-medical-records-selective-history-copy"></a>
### 1. 選択した問診・治療明細の複写

複写の各経路で「金額・単価が変わっている可能性があるため今回の価格を確認する」注意を通知し、既存の上書き確認ダイアログにも同じ注意を表示する。根拠はBacklog Phase1 No.10／92（F11:G11、F93:G93）の回答。10/13の全経路通知・入力済み時の確認と、10/15の確認ダイアログ／通知による注意を接続し、全操作へ新しい必須ダイアログを追加する決定とはしない。対応するマスタがある明細は複写時に照合した今回価格を使う。元単価の継承は、正規の手入力「その他」行など、元々マスタ参照を持たず単価を直接記録する定義済み種別に限る。マスタ参照が必要な行の無効・欠落・取得不能・不一致、無型の項目をこの種別へ読み替えず、既存どおり再選択／拒否へ渡す。

| 条件ID | 前提／操作 | 期待結果 | 検証状態 |
|:---|:---|:---|:---|
| AC-medical-records-copy-price-notice | 各複写経路でマスタ対応行／非対応行を複写し、入力済みへの上書きも試す | 価格確認の注意を通知し既存上書きダイアログにも表示。今回価格照合／元単価継承と無効・取得不能時の保護を区別し、取消は現在入力と元記録を保持 | No.10／92、selective-history-copy。未実施 |

過去カルテから問診と治療明細を選んで複写できる。問診を上書きするときは確認し、取消なら現在入力を保持する。複写元は変更せず、問診は既存どおりローカル入力へ適用し問診保存で永続化する。治療明細は選択した項目だけを既存の個別保存へ渡し、今回の有効マスタの価格を用いる。元カルテの当時価格は元記録に保持する。

無効なマスタや対応しない項目は知らせ、再選択してもらう。旧単価・無型／不一致項目を通知なしで新しい明細へそのまま保存するfallbackは採用しない。対象と現在の医院・患者・grant・版、薬量評価／理由・確認を保存時にも検査する。旧finalizedだけで通常編集を拒否しない。

現行の治療複写は逐次POSTで、途中失敗時も成功済み行が残る。今回、全件同時成功・全体rollbackへ変更していない。成功済み／失敗／未送信の対象を区別して通知し、再試行で成功済みを重複作成しないよう照合する。結果不明を失敗確定にしない。UIの選択範囲と表示方式・再送方式は設計、全体atomic化は未承認のまま。最新合意により[複製直前へ戻す取消](../lifecycle.md#requirement-fr-medical-records-history-copy-undo)を追加し、成功分・直前版・元複製operationを照合する。取消は複製元・他者変更・別ownerの実行を巻き戻さず復元履歴を残す。下の旧履歴複写説明は現行実装参照で、無効／不一致fallback等は本合意へ未追従。

<a id="requirement-fr-medical-records-manual-template-insertion"></a>
### 2. 問診テンプレートの挿入・追記／置換

カルテで登録済みテンプレートを手動選択し、問診欄へ挿入できる。基本は追記、置き換える場合だけ確認する。置換確認の取消は既存入力を保持する。適用した内容は問診の未保存入力として扱い、テンプレート選択だけでサーバー保存しない。問診の既存保存・離脱警告・編集権限・保存履歴を継承し、旧finalizedだけで挿入を拒否しない。

自動挿入・動的サジェスト、別の診断／処方欄への自動適用は追加しない。テンプレートマスタCRUDとカルテ内の適用を別操作にし、挿入で元テンプレートを書き換えない。具体的問診欄・カテゴリ対応と連続適用の表示はUI／型設計で詰め、今回未指定の必須入力や固定カテゴリを追加しない。

<a id="requirement-fr-medical-records-master-history-snapshots"></a>
### 3. 診療マスタの権限編集・無効化・当時内容保持

権限のあるスタッフが診療マスタを編集できる。医院scope・対象masterの操作grantを保持し、カルテ編集権限だけでmaster編集を全開放しない。薬量の対象・値・範囲・単位・出典の承認は医師という初回合意を維持し、新たな臨床責任者役割を作らない。

無効化した項目は新規選択できない。既存カルテは対象項目が持つ名称・単位・価格を記録当時の内容で保持し、名称変更／価格改定・無効化で過去記録を置換しない。再診での新しい複写明細は今回価格、複写元／既存明細は当時価格と区別する。全masterへ存在しない単位／価格を新設する意味ではない。

現在master参照しかない既存データの当時内容を最新値から推測補完しない。確認できる資料・保存snapshotの調査／移行は未実施で、取得不能を当時値確定と表示しない。snapshot方式・master版と選択中の無効化／改定競合は技術設計。訂正として内容を変更するときは編集者ログ・目的別の医師確認・保存版保持と会計整合へ従う。通常本文は終了後も通常編集可能。

<a id="requirement-fr-medical-records-plan-execution-states"></a>
### 4. 治療プランの項目状態と実施済み操作

通常カルテの治療プラン項目ごとに「未実施・実施済み・後日予定・中止」を管理する。診療終了時は未実施・後日予定を未完了として警告し、終了権限のあるスタッフが確認すれば未完了でも終了可能。自動保存時にこの警告を出さない。中止は理由を残し、実施済みへの変更では誰がいつ記録したかを残す。中止と実施済みを同義にせず、確定・診療終了・精算で全項目を実施済みにしない。

実施済みへの切替はチェックボックスとモーダルを基本操作とし、項目と実施内容を識別できるよう既存の診察／治療プラン画面へ接続する。両方を必ず順番に経由する二段階業務や、二度の状態writeを必須にしない。チェック解除だけで無条件に戻さず、[追加合意の理由付き取消モーダル](../lifecycle.md#requirement-fr-medical-records-reasoned-plan-execution-undo)へ接続する。実施済みの誤記訂正は理由付き取消と保存履歴を保持し、逆向きtoggleだけで解除しない。旧finalizedだけを理由に通常編集を追記専用へ限定しない。

記録者は操作した認証済みスタッフ、記録時刻は状態を記録した時刻。実際の実施者・実施時刻と同義にせず、操作しただけで実施者／実績時刻を自動確定しない。今回指定のない実施者選択・実施時刻入力を必須にしない。実施を記録することと、治療明細・課金生成／取消は別の境界で、チェックだけで新しい請求ルールを作らない。入院・ホテルの専用記録／登録時snapshotには本項目状態を自動適用しない。

<a id="履歴複写"></a>
### 履歴複写（従来実装参照）

以下の旧コピー動作には最新合意への未追従を含む。finalized由来の編集ロックと取消なしの動作は最新目標ではない。選択複写・無効／不一致項目の通知再選択・今回価格と元記録保持は上記合意を優先し、従来のfallbackを新業務許可にしない。

- **問診履歴の前回複写（EMR-182 / EMR-219）**: `InterviewHistory` の各行は、複写可能な問診値または `recordId`（明細複写の起点）を持つ場合に限り行末尾に「コピー」ボタンを表示する（`InterviewHistoryItem` の `copySource` が存在する行のみ。複写不可の行には表示しない）。複写ペイロードは `useGetPetMedicalHistory` 内の `transformToHistoryItem` が `InquirySummaryResponse` の `inquiry.chief_complaint`（主訴詳細）・`inquiry.notes`（治療方針）・`inquiry.chief_complaint_type_id`（主訴区分）に加え、元カルテ `id` を `recordId` として保持する（feature 側の `transforms.ts` も同形）。問診値を持たなくても `recordId` のみの行は治療明細の複写が可能。「コピー」押下時、`MedicalRecordInterview` は現在の主訴詳細・治療方針・主訴区分が未編集の既定値なら即時適用し、いずれかがユーザー編集済みなら `ConfirmDialog` で上書き確認を求める（キャンセルは現在値を保持）。問診項目の複写はローカル state の setter のみを更新して `NavigationBlocker` の未保存警告を有効化し、API 送信は行わない。ボタンは行リンク（`/medical-records/:id` 詳細遷移）の sibling の `<button>` であり、確定済み・送信不可のとき `isLocked` prop で明示的に disabled にする。`InterviewHistory` パネル自体は編集ロック用の disabled fieldset の外に配置するため、確定済みでも履歴検索・行リンクは有効のまま（`ExaminationForm` が読み取り専用パネルを fieldset 外の兄弟に置くのと同じ境界規則で、`MedicalRecordMountedTab` の `internalLock` 指定タブが内部でこの境界を管理する）。
  - **治療明細の複写（EMR-219）**: 複写適用の単一経路（`applyCopySource`）が `onCopyRecordTreatments(recordId)` を発火し、`useCopyTreatmentDetails` が複写元カルテの明細を `GET /v1/medical-records/{sourceId}/treatments` で適用時に lazy fetch する（履歴一覧レスポンスに明細は含まれないため）。各明細行は `buildCopiedTreatmentPayload` で create ペイロードへ変換し、consultation/procedure/medicine 行の `unit_price` は複写元の当時価格を持ち込まず、現行マスタ（`/v1/masters/consultations`・`/v1/masters/procedures`・`/v1/masters/medicines`、キャッシュは共有 master 参照キー）の `id`・`price` で再解決する。マスタ不一致または無型（`item_type=other`）行は複写元の `unit_price`・参照をそのまま維持する。作成は既存の `POST /v1/medical-records/{currentId}/treatments` へソース `sort_order` 昇順で逐次 POST し、`sort_order` は複写先の現在最大値 +1 からのオフセットを加算する。現在カルテに `id` が無い（自動作成前）場合は問診項目の複写のみ適用され明細複写はスキップされる。fetch/create 失敗時はエラートーストを出し、複写元カルテへの書込みは一切行わない（作成済み行はロールバック API が無いため残る）。

<a id="persistence"></a>
## 保存・取消・再試行・同時編集

同一カルテを他端末で編集中なら編集を禁止し、ヘッダーに状態を表示する。[同時編集禁止・強制取得可の業務正本](../lifecycle.md#medical-record-edit-exclusivity-20261018)を先に適用する。通常の自動保存は編集権を持つ端末で維持し、以下の版競合処理だけを編集権の代用にしない。

<a id="requirement-fr-medical-records-optional-chief-complaint-type"></a>
#### 2.2 保存プロセス

最新回答で自動保存を維持する。以下の従来送信先を新しい保存ボタン待ちへ変更しない。[タブクリア](../lifecycle.md#requirement-fr-medical-records-clear-tab-open-state)／[複製取消](../lifecycle.md#requirement-fr-medical-records-history-copy-undo)は自動保存済みの対象も復元する追加操作であり、保存済み原版や監査を削除せず、復元後の状態と操作履歴を保持する。

> **追記（2026-10-07 監査）**: 保存ルーティング（アクティブタブ単位のメイン保存・ヘッダー属性の即時保存・タブ別サブデータの即時保存と各エンドポイント）の契約本文の正本は [保存ルーティング契約](../lifecycle.md#hybrid-sync)（ハイブリッド保存ロジック）。本節は統合前9タブの保存先参照を残し、統合後の画面構成は§1に従う。本画面固有の操作仕様のみを残し、ルーティングの重複掲載（旧・送信先表と mermaid 図）をやめた。

- **メイン保存の UI**: 画面右下のフローティング「保存」ボタン（`MedicalRecordFloatingActions`）。送信先はアクティブタブで切り替わる（「問診」は `PATCH /medical-records/:id/inquiries`、「診察/治療プラン」は `PATCH /medical-records/:id/clinical-plan`。切替契約の本文は [保存ルーティング契約](../lifecycle.md#hybrid-sync)）。他タブではカルテ本体の送信は行われない（`PATCH /medical-records/:id` へまとめて送信する方式ではない）。予防接種タブでは偽成功を避けるため外側の保存ボタン自体を表示せず、タブ内の接種記録追加を使う。
- **主訴区分は任意（空欄可）**: 「問診」タブの主訴区分は必須項目ではない。`InterviewChiefComplaint` 内のクリア可能な `SearchableSelect`（`clearable`）で選択し、候補は `useGetChiefComplaintTypes` の主訴区分マスタから供給される。クリア操作で空欄に戻せ、空欄時は `PATCH /medical-records/:id/inquiries` へ `chief_complaint_type_id: null` が送られる。バックエンドは nullable（`*uint64`）として受け付け、明示的なクリアとして永続化するため、保存・再読込後も空欄のまま維持される。
- **ヘッダー即時保存の UI 側挙動**: 担当医・来院種別・診療日はヘッダー変更と同時に `PATCH /medical-records/:id` する（保存ボタンを経由しない）。来院種別の成功後は `queryKeys.medicalRecords.detail` を invalidate し、再読込でラベルが戻らないようにする。失敗時はローカル state をロールバックする。appointment 紐付き通常カルテの `date` 固定（予約開始の JST 日付・変更は BE Conflict・UI は未紐付け時のみ成功）の契約本文は [保存ルーティング契約](../lifecycle.md#hybrid-sync) / [reservation-to-record-flow.md](../../../workflows/reservation-to-record-flow.md) §5。EMR-308: ヘッダーの「次回予定」（`NextVisitButton`）はカルテ PATCH ではなく `PUT /v1/reservations/pets/:petId/next-visit-intent`（取消は `DELETE` 同パス）でペットの tentative appointment（`status='tentative'`、日付のみ `intended_date`、枠・時間・担当未定）を upsert/削除する。write 契約の正本は [reservation-to-record-flow.md](../../../workflows/reservation-to-record-flow.md) §2.3、画面側の管理は [reservation-management.md](../../reservations/pages/reservation-management.md) を参照。
- **アクティブタブの追加保存**: 保存成功直後、その時点で開いているタブが「診察/治療プラン」または「見積書」の場合のみ、`useMedicalRecordPostSave` が対応する登録済みコールバックを追加実行する（他タブ在中時は発火しない。両タブを並行実行することもない）。見積タブは `items` を create/update 同一 tx で置換永続化する（独立画面 `/estimates` はヘッダ金額のみ。詳細は [23-estimate-form.md](../../../screens/23-estimate-form.md)）。
- **治療・検査等のサブリソース**: 「治療」タブの明細は行単位の追加/編集/削除操作ごとに個別・即時送信される。メイン保存とは独立しており、「バックグラウンド並行保存」ではない（タブ別サブデータの即時保存契約は [保存ルーティング契約](../lifecycle.md#hybrid-sync)）。

<a id="errors-empty"></a>
## エラー・空状態

| 状況 | 検知／表示 | 利用者の次の操作 | 根拠 |
|:---|:---|:---|:---|
| 会計確認が未完了 | 会計ownerの確認状態を表示 | 通常編集／保存・診療終了にカルテ確定ボタンを要求しない。会計の要件はD10へ | 最新操作契約 |
| 画像件数・容量超過 | 選択分全体を拒否してエラー | 制限内のファイルを選択 | 画像・資料の選択制限 |
| ヘッダー更新失敗 | 現行のローカルstate復帰と失敗表示 | クリア基準と別に、保存結果を照合して再試行。入力や結果不明を成功扱いしない | 画面保存操作 |
| 予防接種記録0件 | EmptyState | 記録を追加 | タブ構成表 |
| その他エラー・空状態 | 未記載／要確認 | 未記載／要確認 | タブ別条件の不足は要確認 |

<a id="acceptance"></a>
## 受入条件

この表は既存仕様から抽出した代表条件。すべての入力・操作・異常系への要件IDと受入条件の網羅的対応は未記載／要確認。

| 条件ID | Given（前提） | When（操作） | Then（期待結果） | 独立確認 | 成功条件 | 確認手順／証跡 |
|:---|:---|:---|:---|:---|:---|:---|
| AC-medical-records-optional-chief-complaint-type | 入力可能な記録で主訴区分を空欄にする | 保存して再読込する | 主訴区分は空欄のまま | この問診保存だけを確認 | 任意項目の空欄が保持される | [要件](#requirement-fr-medical-records-optional-chief-complaint-type)・[既存UATの所在](../README.md#受入条件の所在)。実機未確認 |
| AC-medical-records-copy-select | 問診／治療明細がある過去カルテ | 対象を選び複写 | 選択対象だけ今回入力に適用し元記録は不変 | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-selective-history-copy)。実機未確認 |
| AC-medical-records-copy-overwrite-cancel | 問診に既存入力あり | 上書き確認を取り消す | 現在入力を保持し上書きしない | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-selective-history-copy)。実機未確認 |
| AC-medical-records-copy-current-price | 過去明細と今回有効masterの価格が異なる | 治療明細を複写 | 新明細は今回価格、元の当時価格は保持 | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-selective-history-copy)。実機未確認 |
| AC-medical-records-copy-invalid-master | マスタ参照が必要な複写対象で、参照が無効／欠落／取得不能／不一致 | 複写を適用し、正規の手入力その他行の複写と比較する | 前者は知らせて再選択／拒否し、旧価格fallbackやその他行への読み替えで保存しない。元々マスタ参照を持たない定義済み手入力その他行の元単価継承は別契約として維持する | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-selective-history-copy)。実機未確認 |
| AC-medical-records-copy-partial-failure | 逐次保存で一部成功・後続失敗／不明 | 結果を確認して再試行 | 成功済みを保持・照合し重複作成しない、全件rollbackへ変えない | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-selective-history-copy)。実機未確認 |
| AC-medical-records-template-append | 問診に入力あり | templateを手動選択して追記 | 既存入力を保ち未保存問診に追記、即時API保存なし | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-manual-template-insertion)。実機未確認 |
| AC-medical-records-template-replace-cancel | 問診内容をtemplateで置換しようとする | 確認を取り消す | 既存問診を保持 | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-manual-template-insertion)。実機未確認 |
| AC-medical-records-template-no-auto | template候補がある | 画面を開く／患者を切り替える | 勝手に挿入／自動提案／元template更新をしない | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-manual-template-insertion)。実機未確認 |
| AC-medical-records-master-inactive-history | 記録済み項目を無効化／名称単位価格変更 | 新規選択と過去閲覧 | 無効化した項目は新規利用不可。有効のまま改定した項目は今回値で新規選択可。どちらも既存当時内容は保持 | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-master-history-snapshots)。実機未確認 |
| AC-medical-records-master-scope-grant | master操作権限なし／別医院 | masterを編集 | カルテ編集可だけでmasterを編集させない | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-master-history-snapshots)。実機未確認 |
| AC-medical-records-plan-checkbox-modal | 権限あるスタッフが未完了項目を確認 | チェックボックスとモーダルの導線で実施済みにする | 対象項目を実施済みにし実記録者・記録時刻を残す、二度writeしない | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-plan-execution-states)。実機未確認 |
| AC-medical-records-plan-uncheck-not-undo | 実施済み項目がある | チェックを解除しようとする | 無条件の実施取消・以前の保存版消去へ転用しない | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-plan-execution-states)。実機未確認 |
| AC-medical-records-plan-cancel-reason | 中止対象の項目がある | 中止にする | 理由を残し実施済みとは扱わない | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-plan-execution-states)。実機未確認 |
| AC-medical-records-plan-deferred-warning | 未実施／後日予定項目あり | 診療終了時に警告を確認 | 未完了でも終了可、項目を実施済みに自動更新しない。自動保存時には警告しない | 本ケースを単独確認 | 実機未確認 | [FR](#requirement-fr-medical-records-plan-execution-states) |
| AC-medical-records-plan-recorder-not-performer | 記録者と実施者が異なり得る | 実施済みを記録 | 記録者／記録時刻を残し実施者へ自動転記せず未指定項目の必須化なし | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-plan-execution-states)。実機未確認 |

<a id="decisions-evidence"></a>
## 未決事項・関連仕様・根拠

未記載の型・必須条件・初期値・エラー・受入結果は要確認。既存の未実装・判断待ち・臨床承認待ちは以下の元記載を維持する。今回の形式整理を合意や実装完了の根拠にしない。

[文書規約](../../../authoring-guide.md)・[ドメイン入口](../README.md)。根拠は整理前の既存仕様本文と、その本文が示す日付付き追記・関連資料。文書確認日は上記の構造確認日であり、過去の承認日を置き換えない。API表・コード名は参照情報で、wire型とDB定義の正本は文書規約に従う。

### 3. 技術仕様



#### 使用コンポーネント

- **`MedicalRecordForm`**: 統合フォーム。
- **`useMedicalRecordForm`**: 旧画面の状態管理（統合前9タブ ＋ 履歴引用の実装参照）を統括するカスタムフック。
- **`historyItems`**: 統合前の「問診」タブ内の右カラムから治療プランへ移設する履歴欄（`InterviewHistory`）に過去の問診履歴を表示。「予防接種」タブは左ペイン一覧と右カラム `VaccinationHistory` の両方を `useGetPetVaccinations`（`GET /vaccinations?pet_id=`）で描画する（フォーム全体で共有される単一サイドパネルではなく、タブごとに独立）。

#### API連携

| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/medical-records/:id` | カルテ本体の取得（飼主／患者／担当医・問診要約・体重投影等。治療・全バイタル・clinical-plan・画像・検査等は個別取得） | `medical-records` | `view` |
| POST | `/api/v1/medical-records` | 新規カルテレコードの作成 | `medical-records` | `create` |
| PATCH | `/api/v1/medical-records/:id` | カルテ属性の更新 | `medical-records` | `edit` |
| PATCH | `/api/v1/medical-records/:id/inquiries` | 「問診」タブの保存 | `medical-records` | `edit` |
| PATCH | `/api/v1/medical-records/:id/clinical-plan` | 「診察/治療プラン」タブ（身体検査所見・診断・治療方針）の保存 | `medical-records` | `edit` |
| POST | `/api/v1/medical-records/:id/treatments` | 処置・処方明細の追加 | `medical-records` | `create` |
| PATCH | `/api/v1/medical-records/:id/treatments/:treatmentId` | 処置・処方明細の更新 | `medical-records` | `edit` |
| DELETE | `/api/v1/medical-records/:id/treatments/:treatmentId` | 処置・処方明細の削除 | `medical-records` | `delete` |
| PUT | `/api/v1/medical-records/:id/treatments` | 明細の並び替え（`sort_order` の一括更新。内容の一括更新ではない） | `medical-records` | `edit` |
| GET | `/api/v1/estimates?medical_record_id=` | 「見積書」タブの見積データ取得 | `estimates` | `view` |
| POST | `/api/v1/estimates` | 見積の新規作成（タブからは `items` を同一 tx で保存） | `estimates` | `create` |
| PATCH | `/api/v1/estimates/:id` | 見積の更新（タブからは `items` を同一 tx で置換） | `estimates` | `edit` |
| GET | `/api/v1/vaccinations` | 予防接種タブのペット単位一覧（必須クエリ `pet_id`。`page`/`limit` 必須） | `vaccinations` | `view` |
| POST | `/api/v1/vaccinations` | 予防接種タブからの新規接種記録 | `vaccinations` | `create` |


<a id="medical-record-agreement-integration-three"></a>
## 画像・印刷・後日結果・訂正影響の追加3項目

2026-10-14、質問 Sentinel_f99722e55be0819189e5d50080928914 に対する Sentinel_59d45d62afe08191bc0b289c42486962 の「ok」に基づく3FR／12AC。最新の通常編集と目的別訂正・保存§2を継承し、実装／実機受入済みを表さない。

<a id="requirement-fr-medical-records-image-drop-selected-print"></a>
### 画像の共通入力検証と対象範囲を選ぶ印刷

撮影ボタンから写真を撮り、そのまま当該カルテへ添付できる。[Backlog Phase1 No.46 F47:G47](https://docs.google.com/spreadsheets/d/17jaFBv7NmbPzk88lZx94vFxa_G1shGNX2O_lFGt4QdQ/edit#gid=0&range=F47:G47)、2026-09-23曽我回答が根拠。撮影取消・失敗時は保存済み添付を保持し、添付の保存成功と撮影・プレビューを区別する。通常カルテの既存画像機能であり、健診専用作図・写真機能を新設する決定ではない。

画像は撮影・ファイル選択・ドラッグ＆ドロップで登録でき、同じ形式・件数・容量・操作権限・医院／患者照合・保存版／owner保護を適用する。旧finalizedだけで通常入力を拒否しない。JPEG／PNG／GIF／PDF、一度10件まで、1ファイル10MiB以下・合計50MiB以下という既存制限を保持する。選択分に超過があれば全体を受け付けず、D&Dだけで制限を迂回しない。形式の選択補助とファイル内容の検証は別で、保存時の検証も維持する。

印刷前に対象のカルテと範囲を選びプレビューする。範囲は既存の患者ヘッダー・各タブ・バイタル・画像／資料から選択する設計とし、選択した内容に関係する訂正・追記は元記載との関係と日付／記録時刻、記録者が分かる表示を含める。訂正済みを元のままの記載として印刷しない。範囲選択は未保存draftの自動保存・全タブ一括保存を発火せず、保存済みの内容を対象にする。未保存内容は印刷に反映されないことを表示し、印刷cancelは入力を保持する。

既存の閲覧／操作認可と対象情報の閲覧範囲を検査し、表示できない他ownerのデータを印刷で取得・公開しない。新しいgrant IDや飼主への自動公開・送付は追加しない。画面・帳票レイアウトや既存の印刷入口へのadapterは技術設計で、印刷成功が診療確認・確定を代行しない。

<a id="requirement-fr-medical-records-late-result-shared-history"></a>
### 後日検査結果の共通参照と履歴

通常カルテには診療終了後も後日結果を追加できる。以前の記載・保存版・既存追記を保持し、現在の記録と変更履歴を識別する。旧finalizedだけを理由に通常入力を追記専用へ限定しない。結果訂正は変更前後・理由・操作主体／記録時刻を履歴として保持し、検査画面とカルテで同じ結果とその訂正履歴を参照する。別々の結果本文を二重管理せず、患者・医院・検査・カルテの対応を照合する。

確定時に参照した結果版と後日の結果／訂正版を区別し、現在の結果だけで確定時の記載を上書きしない。旧finalizedの状態を消して履歴を失う処理や、別患者結果の取込を認めない。通常編集では現在の記録を更新できるが、以前の参照版を破壊しない。検査ownerの確定／訂正・権限手順は検査正本へ渡し、検査側の確定解除をカルテの確定解除の根拠にしない。結果待ちを保存成功・診療終了だけで完了／確定へ変換しない。結果の未取得・取得失敗・対応不明は「結果なし」「正常」へ置換しない。

<a id="requirement-fr-medical-records-correction-owner-impact"></a>
### 訂正による会計・在庫への影響表示と受渡し

カルテ訂正だけで確定済み請求・返金・在庫を自動変更しない。変更した記録／項目と理由、関連する請求・在庫等への影響を表示し、各機能の訂正操作へ案内する。対応が不明な影響を「影響なし」「訂正済み」と表示しない。手元で差額・数量を推測して確定処理せず、権限あるownerの確認・操作へ渡す。

精算確定歴がある予約の通常診療終了Undo禁止、旧確定時を含む以前の保存版・元日付の保持、当日カード／自動課金なしの過去入力、復元だけで開始／課金しない境界を維持する。締め後会計・返金・在庫増減の内部手順はD10／12／13で詰める。カルテの案内表示や訂正履歴を、返金成功・在庫増減成功の証拠にしない。閲覧権限がないownerの内容は公開せず、操作権限不足は案内先の権限不足として区別する。

### 追加受入条件（実機未確認）

| ID | Given | When | Then | 検証 | 状態 | 正本 |
|:---|:---|:---|:---|:---|:---|:---|
| AC-medical-records-camera-attach | 編集権限と患者／医院条件を満たす通常カルテ | 撮影ボタンで撮影し添付／再表示する、または撮影を取消す | 当該カルテへ写真を添付でき、共通の形式・容量・認可・対象照合を維持。取消／失敗では既存添付を保持し、健診専用機能を新設しない | 撮影・保存・再取得と取消を別確認 | 実機未確認 | [FR](#requirement-fr-medical-records-image-drop-selected-print) |
| AC-medical-records-image-drop-same-validation | 同じファイル集合 | 選択／D&Dで追加 | 形式・件数・容量・権限・scope・保存版／owner保護が同じ。親finalizedだけの通常入力拒否は置換 | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-image-drop-selected-print) |
| AC-medical-records-image-drop-reject-batch | 個別または合計サイズ／件数超過 | どちらの入口でも追加 | その選択分全体を拒否し保存済み添付を保持 | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-image-drop-selected-print) |
| AC-medical-records-print-range-corrections | 訂正／追記のある保存済み記録 | 範囲を選びプレビュー／印刷 | 選択対象と関係する訂正追記・元記載の関係を識別可能に表示 | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-image-drop-selected-print) |
| AC-medical-records-print-draft-cancel-scope | 未保存入力または閲覧不可のowner情報あり | 印刷準備／cancel | draft未反映を表示、無断保存・権限外印刷・入力破棄・飼主自動公開なし | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-image-drop-selected-print) |
| AC-medical-records-late-result-draft | 作成中で患者／医院を照合済み | 後日結果を追加 | 検査ownerの同じ結果を参照し自動確定しない | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-late-result-shared-history) |
| AC-medical-records-late-result-finalized | 旧finalizedの結果参照版あり | 後日結果／訂正版を通常カルテへ反映 | 旧finalizedだけで拒否せず、元記載・保存時参照版・履歴を保持。検査ownerの独自確定／訂正は維持 | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-late-result-shared-history) |
| AC-medical-records-late-result-revision-history | 結果が訂正される | 検査画面とカルテを開く | 同じ結果IDと履歴を参照し前後・理由・主体／時刻を識別 | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-late-result-shared-history) |
| AC-medical-records-late-result-unknown-scope | 別患者／医院または必要取得失敗 | 結果取込／再表示 | 取込を止め照合、未取得を正常／結果なしにしない | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-late-result-shared-history) |
| AC-medical-records-impact-financial-stock-no-auto | 確定請求または在庫関連あり | カルテ訂正 | 影響とowner訂正入口を表示し自動請求変更・返金・在庫増減なし | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-correction-owner-impact) |
| AC-medical-records-impact-unknown | 関連影響の取得失敗／対応未確認 | 訂正後に確認 | 未確認として表示し影響なし／訂正済みへ置換しない | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-correction-owner-impact) |
| AC-medical-records-impact-handoff-permission | owner操作権限不足 | 訂正入口へ進もうとする | 権限不足を表示し迂回操作・他owner情報公開なし | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-correction-owner-impact) |
| AC-medical-records-impact-settled-history | 精算確定歴／締め後または過去入力／復元 | カルテ側の操作 | 既決のUndo禁止・自動課金なしを保持し会計／在庫内部はowner確認 | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-correction-owner-impact) |

<a id="medical-record-shared-coverage-ac"></a>
## 全タブ・共通操作の横断受入条件

これまでの合意と保存§2に対する14ACを補う。新しい業務許可や実機PASSではない。各条件は問診・診察／治療プラン・治療・予防接種・健診・検査・画像・見積・会計医師確認、バイタル・ヘッダー・専用訂正の該当write／readに適用し、各ownerの正本を維持する。

| ID | Given | When | Then | 検証 | 状態 | 正本 |
|:---|:---|:---|:---|:---|:---|:---|
| AC-medical-records-coverage-active-tab-only | 複数タブにdraftあり | 統合先でスタッフ入力を保存 | 既存送信先を保持し他データを無断一括保存しない。飼主原文は別データで編集不可 | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#hybrid-sync) |
| AC-medical-records-coverage-header-immediate | ヘッダー変更とタブdraftあり | 担当医／来院種別／日付を許可範囲で変更 | 既存個別保存と予約整合を保持、タブdraftを保存成功扱いにしない | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#hybrid-sync) |
| AC-medical-records-coverage-child-independent | 親save前のdraftと子入力あり | 許可された子操作 | 既存子保存先へ独立送信し親save待ち／全タブ一括保存へ変更しない | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#hybrid-sync) |
| AC-medical-records-coverage-vital-read-state | バイタル履歴取得失敗／不完全または患者切替中 | 表示／入力／薬量評価 | 測定なしへ置換せず対象一致と現在情報を再照合 | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#requirement-fr-medical-records-unevaluable-dose-manual) |
| AC-medical-records-coverage-snapshot-null | 旧master参照・任意null・未取得値あり | 読込／複写／保存 | 当時値を保持し未取得を0／空欄／現masterの当時値へ推測補完しない。APIごとの未送信／null／clear・旧状態を照合 | 全該当経路を照合 | 実機未確認 | [正本](#requirement-fr-medical-records-master-history-snapshots) |
| AC-medical-records-coverage-finalized-all-writers | 旧finalized親の全子／外部owner経路 | 通常入力／編集又は目的別操作 | 親finalizedだけの通常編集拒否を置換し、以前の保存版・必要監査とowner独自の確定／削除／訂正保護を維持。GET初期化を含む全writerを照合 | 全該当経路を照合 | 実機未確認 | [最新正本](../lifecycle.md#requirement-fr-medical-records-normal-edit-after-end) |
| AC-medical-records-coverage-finalize-concurrent-deps | 終了警告の確認後に子や依存が変更 | 診療終了する | 最新依存を再照合し旧確認で終了しない。未完了は確認後終了可、取得不明は停止 | 全該当経路を照合 | 実機未確認 | [最新正本](../lifecycle.md#requirement-fr-medical-records-screen-consultation-end) |
| AC-medical-records-coverage-scope-generation | 旧医院／患者／actorの遅い取得応答あり | 画面／医院／患者を切替 | 新対象へ旧値を適用せずscopeを再取得・照合 | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#requirement-fr-medical-records-authorized-editor-audit) |
| AC-medical-records-coverage-child-audit-failure | 本体／子／専用操作の必要ログ保存が失敗 | 更新を保存 | 必要ログなしの成功を返さず同じ操作単位をrollback | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#requirement-fr-medical-records-authorized-editor-audit) |
| AC-medical-records-coverage-save-partial-dirty | 問診成功・他draft／子操作失敗あり | 保存結果を表示 | 成功対象だけdirtyを解除し未保存入力／失敗対象を保持 | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#hybrid-sync) |
| AC-medical-records-coverage-unknown-receipt | 送信後通信断で結果不明 | 再試行／取消を要求 | 同じ操作を照会／再送し重複転記・複写・復元を作らない。未対応receiptの旧APIは成功と推測せず関連照合し不明ならwrite停止 | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#requirement-fr-medical-records-empty-record-hide-restore) |
| AC-medical-records-coverage-versions-no-overwrite | 他者が同対象を変更済み | 古い版の更新 | 上書きせず再取得し自分の入力と比較・再確認 | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#requirement-fr-medical-records-authorized-editor-audit) |
| AC-medical-records-coverage-special-record-separation | 入院／ホテル／トリミング記録が併存 | 通常記録の状態／複写／受付を操作 | 専用記録へ新状態や作成契約を自動適用しない | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#requirement-fr-medical-records-reception-and-retrospective-entry) |
| AC-medical-records-coverage-owner-read-and-handoff | 別ownerへの影響・印刷対象あり | 閲覧／引継ぎ後戻る | 認可範囲と現在版を検査、遷移をowner訂正成功とみなさない。転記／過去／訂正で通常createの在庫副作用を誤流用しない | 全該当経路を照合 | 実機未確認 | [正本](#requirement-fr-medical-records-correction-owner-impact) |

## 現APIとの最終照合の入口

[実キー・型・null／clearと互換契約](../../../../design/domains/medical-records/technical-spec.md#medical-record-reconciled-contract)へ接続する。実APIの既存受理範囲を新業務許可へ拡大しない。患者転記・過去入力・復元の目的別条件を汎用PATCH／createで迂回しない。通常本文／後日結果の現在版更新は最新の通常編集契約とowner独自の確定／訂正へ接続する。取得経路にはGetOrCreate初期化があるため、印刷用read modelは副作用のない保存済みprojectionで構成する。共通14ACは全該当writer／read経路への確認条件で、現コードの適合PASSを示さない。


<a id="medical-record-agreement-diagnosis-vitals-20261015"></a>
## 診断・バイタルの追加2項目（2026-10-15）

質問 Sentinel_14ca8cf17a2881919135ac55b7828ef6 の診察中の記録2項目に対し、Sentinel_4ec55f21a17c819199d16ad1add270a4 が「1 そうですね、複数登録にしましょう、ただ第一、第二をそれぞれ複数登録で。2 yes」と回答したことを根拠とする。第1・第2を一つの診断一覧と単一の主診断指定へ置換する案は採用しない。この回答1・2は診断・バイタルを対象とし、初診判定に含める受診歴の対象は[予約／受付の別合意](../../reservations/pages/reservation-management.md#visit-history-included-types)へ従う。2FR／12ACを追加し、既存合意・保存§2と[最新の通常編集／復元契約](../lifecycle.md#medical-record-agreement-autosave-20261015)を保持する。実装・DB／API変更・実機検証は未実施。

<a id="requirement-fr-medical-records-diagnoses-multiple-per-category"></a>
### FR-medical-records-diagnoses-multiple-per-category — 第1・第2それぞれに複数診断を登録

「診察/治療プラン」タブの第1診断・第2診断の区分を保持し、各区分へ複数の診断名を登録・表示・編集できる。各診断が属する第1／第2を保存と再表示でも識別し、一方の操作で他方の診断を上書きしない。単一の主診断を選ぶ仕組みに集約せず、第1が1件だけ又は両区分合計2件までという制限にしない。今回の合意から新たな必須件数・件数上限・自動順位付け・重複禁止条件を追加しない。

第1／第2という記録上の区分と、診断病名マスタの疾患カテゴリは別概念である。個々の診断名とマスタ所属カテゴリの整合・医院／権限・有効性・当時名称の保持は[既存master契約](#requirement-fr-medical-records-master-history-snapshots)へ従う。既存の第1・第2単件データは元の区分と内容を保ち、複数化を理由に消去・改名・別区分へ移動しない。

保存先と送信境界は既存の診察／治療プラン保存に従い、自動保存を廃止したり全タブ一括保存へ変えたりしない。診療終了後も同じ通常カルテで編集でき、以前の保存版と実編集者履歴を保持する。タブクリアはタブ開始時、既存の履歴複製取消が対象とする入力の取消はその複製直前へ戻す既決保護を維持し、新しい診断複製機能の許可には広げない。旧単件APIで複数分を欠落させる保存は成功扱いにせず、未対応・競合・失敗・結果不明を区別して入力を保持する。保存済み診断の変更・復元により元版・他者変更・別ownerの確定内容を無警告で消去しない。

<a id="requirement-fr-medical-records-vital-measured-time-history"></a>
### FR-medical-records-vital-measured-time-history — 最新バイタルの測定時刻と以前の測定履歴

通常カルテの体温・体重等のバイタルは、表示する最新値にその値の測定時刻を添える。同日の測定を一つに集約せず、同日を含めた以前の測定も履歴で確認できる。値・単位・測定日時と対象患者／カルテの対応を保持し、別々の測定由来の値へ同じ時刻を一律に付けない。過去日の測定を当日の値と誤認させないよう測定日も識別できるようにする。

「最新」は実際の測定日時に基づき、登録・編集日時が新しいだけの過去測定を最新測定へ昇格させない。測定日時不明を現在時刻で補完せず、最新と特定できない場合は不明を識別する。未入力値・単位不明・取得中／失敗・不完全な履歴を、0・正常・測定なし又は全件取得済みへ置換しない。最新表示の対象値と元の測定を辿れ、複数測定や同時刻の記録を一件へ無条件に統合・削除しない。

履歴は同じ患者・カルテの認可済み記録を参照し、開く・閉じるだけで記録作成・複製・保存・診療終了を行わない。診療中でもそのカルテの以前のバイタル測定を確認でき、[終了したカルテを対象とするスタッフ診療履歴](../lifecycle.md#requirement-fr-medical-records-ended-record-history)とは区別する。測定記録・変更履歴・元版を保持し、終了後の通常編集、既存のバイタル個別保存、競合／結果不明と監査の保護を維持する。タブのクリアでタブ外の測定を巻き戻さない。体重の薬量評価への適用可否・評価不能／再評価と医師承認は既存正本に従い、最新表示から有効期限や医学的な閾値を新設しない。

<a id="medical-record-diagnosis-vitals-ac"></a>
### 追加受入条件（実装・実機未確認）

| ID | Given | When | Then | 独立確認・状態 | 正本 |
|:---|:---|:---|:---|:---|:---|
| AC-medical-records-diagnoses-both-multiple | 診察タブで第1・第2へそれぞれ2件以上の診断を入力 | 保存後に再読込 | 両区分の全件を元の区分で表示し、各1件／合計2件へ切り捨てず単一主診断への集約なし | 本ケースを単独確認。実機未確認 | [FR](#requirement-fr-medical-records-diagnoses-multiple-per-category) |
| AC-medical-records-diagnoses-category-isolation | 両区分に複数の診断がある | 第1又は第2の一件を編集 | 対象区分・項目を識別して反映し、他方の診断を上書き・移動しない | 本ケースを単独確認。実機未確認 | [FR](#requirement-fr-medical-records-diagnoses-multiple-per-category) |
| AC-medical-records-diagnoses-legacy-preservation | 既存の単件第1・第2又は未設定の区分がある | 読込・再表示 | 元区分・当時内容・未設定を保ち、架空の診断や主診断指定を補完しない | 本ケースを単独確認。実機未確認 | [FR](#requirement-fr-medical-records-diagnoses-multiple-per-category) |
| AC-medical-records-diagnoses-master-boundary | 無効masterの過去診断又は名称／疾患カテゴリ不一致の新規選択がある | 表示・選択・保存 | 過去の当時内容は保持し、新規選択は既存の有効性・関係整合・医院／権限条件へ従う。第1／第2と疾患カテゴリを混同しない | 本ケースを単独確認。実機未確認 | [FR](#requirement-fr-medical-records-diagnoses-multiple-per-category) |
| AC-medical-records-diagnoses-save-no-truncation | 複数診断と、単件のみ対応するAPI又は競合／失敗／結果不明がある | 保存・再試行 | 全件保存できない状態を成功とせず入力を保持。黙った切捨て・他者上書き・盲目的再送なし | 本ケースを単独確認。実機未確認 | [FR](#requirement-fr-medical-records-diagnoses-multiple-per-category) |
| AC-medical-records-diagnoses-ended-revert-safety | 終了済み通常カルテで診断編集・保存後、タブ開始基準がある | 通常編集又はタブクリア | 旧finalizedだけで拒否せず既存保存境界を維持。クリアは基準と現版を照合し元版・必要ログ・他者／別owner保護を保持 | 本ケースを単独確認。実機未確認 | [FR](#requirement-fr-medical-records-diagnoses-multiple-per-category) |
| AC-medical-records-vital-latest-measured-time | 同じカルテに測定値と実測日時がある | ヘッダーを表示 | 最新値・単位に対応する測定日時を識別でき、別測定の時刻や編集時刻を付けない | 本ケースを単独確認。実機未確認 | [FR](#requirement-fr-medical-records-vital-measured-time-history) |
| AC-medical-records-vital-same-day-history | 同日中に複数回の測定と過去日の測定がある | 履歴を開く | 以前の測定も値・単位・測定日時で確認でき、同日を1件に集約せず元記録を保持 | 本ケースを単独確認。実機未確認 | [FR](#requirement-fr-medical-records-vital-measured-time-history) |
| AC-medical-records-vital-measurement-order | 以前の実測日時の記録を後から登録又は編集した | 最新値と履歴を再表示 | 登録／編集日時だけで最新測定へ昇格しない。測定日時不明は現在時刻で補完せず、同時刻の別測定を消さない | 本ケースを単独確認。実機未確認 | [FR](#requirement-fr-medical-records-vital-measured-time-history) |
| AC-medical-records-vital-partial-values-time | 値・単位・測定時刻に欠損又は項目間で測定元の違いがある | 最新表示 | 不明を識別し0・正常・架空時刻へ補完しない。表示値とその測定元を対応付け、異なる測定を同一時刻に見せない | 本ケースを単独確認。実機未確認 | [FR](#requirement-fr-medical-records-vital-measured-time-history) |
| AC-medical-records-vital-history-read-scope | 履歴取得失敗／不完全又は医院・患者切替後の遅延応答がある | 最新値／履歴を表示 | 測定なし・全件取得済みとせず再照合し、旧患者／医院の値を新対象に表示しない | 本ケースを単独確認。実機未確認 | [FR](#requirement-fr-medical-records-vital-measured-time-history) |
| AC-medical-records-vital-history-edit-boundary | 診療中又は終了後の同じ通常カルテに複数測定がある | 履歴閲覧・許可された測定編集・別タブクリア | 診療中も測定履歴を閲覧でき閲覧だけのwriteなし。編集は既存個別保存と元版／認証actor履歴を保持し、別タブクリアでタブ外測定を復元しない | 本ケースを単独確認。実機未確認 | [FR](#requirement-fr-medical-records-vital-measured-time-history) |

用語と多重度は[オントロジー](../../../ontology/medical-records.md#medical-record-diagnosis-vitals-ontology)、現単件wireとの差分・全件保持・時刻と履歴の取得／表示・FR/AC対応は[技術契約](../../../../design/domains/medical-records/technical-spec.md#medical-record-diagnosis-vitals-contract)へ接続する。番号付き旧画面は互換案内のまま保持する。

<a id="medical-record-examination-read-adapter"></a>
## 検査項目印刷・患者トップへの接続（2026-10-15追加採用）

<a id="requirement-fr-medical-records-examination-read-adapter"></a>
### FR-medical-records-examination-read-adapter

[D07追加採用](../../examinations/lifecycle.md#examination-print-infection-agreement-20261015)により、カルテで検査結果を印刷する場合も、一つの保存済み検査snapshot内の項目選択・未実施除外を[D07正本](../../examinations/lifecycle.md#fr-examinations-selected-print)へ渡す。既存の患者ヘッダー／タブ等の範囲選択と項目選択は別の粒度であり、セクションを選ぶだけで項目選択を満たしたことにしない。原値・0／陰性・検査日・保存範囲・訂正識別・同じ固定版を保持し、未保存入力の自動保存・確定／解除・送付なし。新たな複数日比較帳票は設けない。

カルテ上部の患者コンテキストにも猫FIV／FeLV・犬mf／HWの原値と各検査日を表示し、D04患者トップと同じD07結果projection・版／出典・訂正／不明の識別を使う。今回カルテの日付やバイタル測定時刻を検査日へ代用せず、院内保存結果の自動反映と他院手入力はD07を通す。カルテ自動保存・タブクリア／複製取消はこの患者全体の結果を巻き戻したり複製したりしない。通常カルテの終了後編集可をD07原版保護の解除へ使わない。

患者トップは現在のD03とD07の閲覧認可内で読み、権限不足・取得失敗・不明・複数候補を記録なし／陰性へ変換しない。D20／LINE公開は追加しない。2026-10-17の直接決定で問診を治療プランへ統合する。飼主事前問診の原文保護・スタッフ記録との分離と既存保存境界を維持し、統合画面の保存単位は設計で接続する。

| ID | Given | When | Then | 検証状態 |
|:---|:---|:---|:---|:---|
| AC-medical-records-exam-item-print-adapter | 保存済み検査の実施済み・未実施・陰性／0・訂正版あり | カルテから検査項目を選択印刷 | D07と同じsnapshot／項目集合・未実施除外・訂正識別。セクション選択だけで代用せず、無断保存・複数日帳票化なし | 未実施 |
| AC-medical-records-infection-top-shared-ref | D04とカルテ患者ヘッダーを同患者で開き、結果訂正／取得失敗を起こす | 保存結果を表示／更新 | 同じD07 ref・日付・出典／版状態を表示。更新未確認を識別し、D03保存／復元から結果を変更しない。現在の両owner閲覧認可を維持 | 未実施 |
| AC-medical-records-infection-tabs-boundary | 問診と診察／治療プランがある | 追加表示・印刷を利用 | 問診を治療プランへ統合し飼主原文をスタッフクリア・複写から保護。D07保護・D20非公開を拡張しない | 未実施 |

技術追跡は[同じ結果を読むadapter](../../../../design/domains/medical-records/technical-spec.md#medical-record-exam-read-contract)。実装・実機適合や既存の全履歴取得制限の解消を表明しない。

<a id="medical-record-intake-prevention-adopted"></a>
## 初診問診項目・予防薬専用管理の追加合意（2026-10-15）

Sentinel_0adf405e47508191abf775b6c4af22a3 の回答1・2に基づき、[初診問診の画面入力・参照（2FR／13AC）](initial-intake.md)と[予防薬専用処方箋・1〜12月年間表の管理／印刷（3FR／15AC）](preventive-medication.md)を追加する。問診タブの独立維持は2026-10-17決定で失効し、治療プランへ統合する。自動保存・タブ開始時クリア・複製直前取消・診療終了後通常編集と元版保持を変更しない。自己申告を確認済み検査・接種・投薬実績へ自動昇格しない。

予防薬はD03の処方／投薬管理で、D08接種とは区別する。表示・専用2帳票の出力入口は、カルテ右上レポートから開く[既存D21](../../../screens/39-owner-report.md#owner-report-preventive-output)。D21全体の印刷・共有・臨床writeを解禁せず、一般薬歴だけで専用機能を代替しない。新しい別メニューは設ける要件にしない。旧資料・ユーザー申告と固定headの静的観察は予防薬正本で区別する。

原紙対応表・FR／AC本文は上記2正本、[語彙](../../../ontology/medical-records.md#medical-record-intake-prevention-ontology)、[保存／参照／印刷の技術trace](../../../../design/domains/medical-records/technical-spec.md#medical-record-intake-prevention-contract)へ接続。今回5FR／28ACは追加採用範囲の計数で、既存の診断／バイタル・他領域の合意件数や実装／実機PASSに加算しない。

## 薬量の自動範囲・複数パターン

[薬量追加採用の正本・追加4AC](../lifecycle.md#dose-source-adoption-20261015)へ従う。用量パターンは薬剤／動物種に加え目的・経路で区別する承認対象で、今回の明示選択と当時版を持つ。薬剤masterの1件と一対一に限定しない。単純体重計算から1回製剤量を得る範囲と、医師確認・理由付き手入力を区別する。目的／経路別選択は複雑protocol自動実行ではなく、具体値の臨床承認#261も代替しない。現行単一用量パラメータの記述は実装参照で、複数パターン対応済みの根拠ではない。

<a id="decision-alignment-20261017"></a>
## 2026-10-17確定仕様との接続

[決定根拠・適用範囲](../../SPECIFICATION_PROGRESS.md#confirmed-business-decisions-20261017)：ユーザーの2026-10-16〜17 JSTの直接決定、文書更新承認は10月17日03:20 JST。以下は目標仕様・未実行の受入条件であり、実装・実機合格を示さない。

[業務正本・受入条件](../lifecycle.md#integrated-plan-20261017)。問診統合・右側履歴移設、飼主原文保護、非表示の来院別要約保存、1回量／用法／処方総量表示、価格入り4種明細の実施前後単価変更を適用する。旧9タブ／問診単独送信の記述は統合前の保存先・実装参照で、画面を別タブへ戻す根拠としない。統合した保存単位は未選定。同一カルテの同時編集禁止・強制取得可は[2026-10-18追加合意](../lifecycle.md#medical-record-edit-exclusivity-20261018)に従い、閲覧のみでは取得せず「編集を開始」で取得し、同一カルテのタブ移動・単なる無操作では維持する。別画面離脱は保存結果確認後に解放し、切断時は入力保持と編集・送信停止、再接続時は編集権と保存結果を確認する。再読み込み後も未保存入力を復元可能とし、再取得後は最新内容と比較して明示選択適用する（04:38:21 UTC本人承認）。API・失効方式・秒数・保存場所・保持期間・比較UI等の技術詳細は未選定。通常の単価割引理由は任意、その他行・取消・提示済み見積・確定会計の必要理由と既決薬量安全条件は維持する。

## バイタル操作・チップと編集者表示の補足受入条件

Backlog No.37（F38:G38）の名前近くでの常時表示とバイタル記録ボタンによる入力／訂正は既存ヘッダー契約へ従う。原回答の時刻不要は後続の本人合意による測定時刻付き表示へ置換する。マイクロチップ番号は登録済みなら名前横、未登録なら非表示とし、登録変更はペット情報へ渡す。No.66（F67:G67）の追記には、追記した認証入力者の氏名を表示する。担当医を追記者へ代用せず、実編集者の監査・通常編集の既決を維持する。10/15の追記者表示完了報告だけを本体表示の実装確認へ代用しない。本体表示の採否は、以下の2026-10-17本人回答Aで確定した。
| 条件ID | 前提／操作 | 期待結果 | 検証状態 |
|:---|:---|:---|:---|
| AC-medical-records-vital-entry-context | 通常カルテを開き、バイタル記録ボタンから入力／訂正する | T/HR/RR/BWが名前近くにクリック不要で表示され、最新の測定時刻・履歴の既決を保持して同じカルテへ記録する | No.37、診断バイタル合意。未実施 |
| AC-medical-records-chip-context | チップ番号登録済み／未登録の患者を開き、ペット詳細で番号を変更する | 登録値は名前横、未登録は非表示。変更はペット基本情報の権限で行い別患者や既発行版を自動変更しない | No.37。未実施 |
| AC-medical-records-addendum-author-name | 担当医と異なる認可スタッフが追記し保存・再表示する | その追記の認証入力者氏名を識別でき、担当医を自動変更しない。必要監査の失敗を保存成功扱いにしない | No.66。未実施 |

## Backlog原条文の補足受入条件

保存済み原Q/Aと既存業務契約の照合による補足。対応済みという旧回答は実装・実機検証の合格証明に用いない。後続の本人合意、医院scope、権限、履歴・会計訂正の既存条件を維持する。

### 過去治療履歴／patient-history-search

同じペットの通常診療履歴とトリミング記録を時系列で併記して参照できる。元記録の種別・医院・帰属を保持し、各記録の既存履歴掲載条件と閲覧権限を適用する。画面・記録の完全統合や原記録の複製は行わない。

| 条件ID | 受入条件 | 根拠 | 検証状態 |
|:---|:---|:---|:---|
| AC-medical-records-history-includes-trimming | 同一ペットに履歴掲載対象の通常診療とトリミング記録がある／履歴を表示する／時系列で両種を確認でき元記録を区別できる。別ペットや認可外の医院履歴は混ざらない。 | Phase1 F32:G32 | 未実施 |

### 治療プランの手入力その他

| 条件ID | 受入条件 | 根拠 | 検証状態 |
|:---|:---|:---|:---|
| AC-medical-records-manual-plan-line-reason | 治療プランにマスタ参照を持たない正規の手入力その他行を追加し、名称・単価と既存必須理由を保持する。理由空白の追加は認めず、通常のマスタ行単価／割引変更の理由任意という別契約を変えない。 | Phase1 F45:G45 | 未実施 |

### 治療プランの手入力その他

| 条件ID | 受入条件 | 根拠 | 検証状態 |
|:---|:---|:---|:---|
| AC-medical-records-manual-plan-line-indicator | 保存した手入力その他行をマスタ参照行と並べると、手入力バッジで区別できる。手入力でない行を未照合のままその他へ読み替えない。原回答の「持ち物」という旧モード名は実装参照として区別し、別の業務分類へ固定しない。 | Phase1 F45:G45 | 未実施 |

## マスタ単価の初期反映の補足受入条件

保存済みSlack SLK-288〜290（2026-09-15〜16、城詰・曽我・城東受付）の金額未反映という不具合は、既存のマスタ選択・価格入り明細・会計への受渡し契約へ結ぶ。薬剤・処置・検査・物販の有効な対象マスタを選ぶ際は、その品目の単価を明細の初期値として扱い、選択しただけで実施済み・請求確定へ進めない。追加後の権限付き単価変更、現在のマスタ価格不変、未確定会計の同期3分岐と確定後訂正は既存正本へ従う。旧回答の負単価・価格未設定・0円商品保存という実装説明は、今回の実装適合の証明とは別に保持する。

| 条件ID | 前提／操作 | 期待結果 | 根拠・検証状態 |
|:---|:---|:---|:---|
| AC-medical-records-plan-master-price-initialization | 対象医院で利用可能な価格設定済みの薬剤・処置・検査・物販を、治療プランへ選択して保存・再表示する | 選んだ品目と単価を同じ明細の初期値として確認できる。別品目・別医院へ混ざらず、選択だけで実施済み・請求確定にしない。既存の単価修正・会計同期・履歴条件を維持 | SLK-288〜290、価格入り4種明細の既決。未実施 |

## 保存済み画像・PDFの閲覧の補足受入条件

保存済みSlack SLK-233（2026-09-08 黒木）の当面は画像・PDFのアップロードと閲覧でよいという範囲を、既存の添付表示へ結ぶ。現在の閲覧認可・医院/患者/添付IDと保存版を照合し、閲覧だけでOCR・検査値や診療本文への自動転記を起動しない。入力形式・件数・容量の既存検証、健診専用作図/撮影を新設しない範囲も維持する。

| 条件ID | 前提／操作 | 期待結果 | 根拠・検証状態 |
|:---|:---|:---|:---|
| AC-medical-records-saved-attachment-view | 対象カルテに保存済み画像とPDFがあり、認可スタッフが既存添付表示から開く | 各保存対象の内容を閲覧でき、別患者/医院の添付を開かない。取得失敗・未取得を正常な添付なしにせず、閲覧だけで診療値・記録を変更しない。OCRを前提にしない | SLK-233、既存添付契約。未実施 |


<a id="medical-record-main-author-name-display"></a>
## 追加決定：カルテ本体の最初の入力者・最終更新者の氏名表示

2026-10-17、質問 Sentinel_690d88ec5a108191bc72f4b73b699736「カルテ本体にも、最初の入力者・最終更新者の氏名を表示しますか？」のA「カルテ画面にも表示する」に、本人 Sentinel_c386820325948191b15281ea832aaf23 が「A」と回答した。[担当医外入力と編集者ログの既存FR](../lifecycle.md#requirement-fr-medical-records-authorized-editor-audit)を補い、カルテ本体の画面に最初の入力者・最終更新者の氏名を表示する。変更履歴を開くことでしか確認できない形には限定しない。既存の変更履歴と追記者氏名表示は保持し、今回採用を実装・実機受入済みとは扱わない。

[EMR-368](https://app.plane.so/baritechllc/browse/EMR-368/)の本人回答により、氏名は各保存が正常成立した時点のスタッフ名を保持し、スタッフIDも保持する。現在のスタッフ名と保存時名を区別し、改名だけで保存済み表示名を書き換えない。最初の入力者と最終更新者は対象カルテの実入力・編集の履歴に対応させ、担当医や予約指名医師、単なる受付／空カルテ作成者で代用しない。別の認可スタッフが編集しても最初の入力者は保持し、最終更新者をその編集に対応させる。編集だけで担当医を交代させず、氏名表示を追加しても変更履歴を省略・書換えしない。取得不能・旧履歴欠落では担当医名等を推測補完しない。具体的な配置・保存/API対応と旧データの照合は設計・実装確認の残件であり、この氏名表示の採用だけでは新たな必須入力や確定後追記のみの制限を追加しない。同一カルテの編集権・同時編集禁止は、後続の[2026-10-18追加合意](../lifecycle.md#medical-record-edit-exclusivity-20261018)へ従う。

### 本体氏名表示の受入条件

| 条件ID | 前提／操作 | 期待結果 | 検証状態 |
|:---|:---|:---|:---|
| AC-medical-records-main-author-names | 担当医と異なる認可スタッフAが初めて記載・保存し、カルテ本体を再表示する | 本体画面上で最初の入力者・最終更新者の氏名を実入力者Aとして確認でき、担当医・予約指名医師は変わらない。既存編集者ログと対応する | No.66、本人A。未実施 |
| AC-medical-records-main-author-names-update | 上記カルテを別の認可スタッフBが編集・保存し、再読込する | 最初の入力者はAの初回正常保存時のID／氏名を保持し、最終更新者はBの今回正常保存時のID／氏名を表示する。A/Bの既存変更履歴と追記者表示を保持し、最新者だけに履歴を置換しない | No.66、本人A。未実施 |
| AC-medical-records-main-author-names-unknown | 入力者履歴が欠落する旧データまたは氏名・履歴取得が失敗した対象を表示する | 不明を現在名・担当医・受付者等から推測補完せず、既存の取得異常/未確認の扱いへ渡す。表示追加で既存監査を削除せず、保存結果不明を成功扱いにしない | 既存認可・監査・保存保護の継承。未実施 |


<a id="medical-record-ipad-report-entry"></a>
## 既存操作の受入補足：iPadの右上レポート入口

10月13日議事録の「iPad(13in)にて、ヘッダーの右側のレポートなどが表示されない」は既存操作への到達不能の不具合報告として扱う。対象が明示された範囲は、10月15日の質問 Sentinel_07e69836f4b08191ba0a900874420934 の第2項に対する本人 Sentinel_0adf405e47508191abf775b6c4af22a3「yes 現状、カルテページ右上のレポートボタンを押したら開かれるところで表示するようにしています」による、既存右上レポートからの[予防薬専用処方箋・年間投薬表](preventive-medication.md)の参照／出力入口である。新機能の採否確認へ戻さない。

13インチiPadでもこの既存レポート入口を確認・操作できることを受入対象とし、ボタンが表示領域から欠落して到達できない状態を対象外扱いにしない。報告の「など」から他の全ヘッダー操作を追加採用せず、新しい独立メニュー・一般レポート全体の印刷／共有・臨床writeを設けない。患者・医院の閲覧認可、既存の選択患者の引継ぎと専用出力境界を保持する。具体的な画面配置・ブラウザ倍率・縦横・最小幅は本補足から確定せず、検証時の条件を記録する。実装・実機解消は未検証である。

| 条件ID | 前提／操作 | 期待結果 | 検証状態 |
|:---|:---|:---|:---|
| AC-medical-records-ipad-report-entry | 13インチiPadで認可内の対象カルテを開き、画面の縦横・viewport・ブラウザ／倍率を記録して、右上の既存レポート入口を操作する。対象患者の保存済み予防薬資料で確認する | 入口を確認・操作でき、同じ選択患者の既存レポートと予防薬専用処方箋／年間表の参照・出力入口へ進める。別患者へ接続せず、一般レポートの印刷・共有・臨床編集を新たに許可しない | 10/13不具合報告、10/15本人の入口指定。記録した条件の範囲だけ確認し、他端末・全条件の合格へ拡張しない。未実施 |


## 患者情報の予防接種名称ラベルの種類別色

[No.32の本人採用](../../vaccinations/lifecycle.md#vaccination-type-label-color-20261017)に従い、旧飼主・患者情報画面右下の予防接種名称／実施日／次予定一覧に相当する表示では、名称ラベルを種類別の色にし、名称文字を保持する。期限切れ等の警告と意味を分け、色だけで識別させない。現行タブ・履歴部品の説明だけで対応済みとせず、具体的な表示位置・配色・管理UIは設計とする。保存済み接種・次予定・認可・公開範囲を維持し、旧画像の全配置やLINE等への展開を追加採用しない。個別受入条件は予防接種正本へ置く。

<a id="other-pet-unpaid-header-entry-20261018"></a>
## 同じ飼主の他ペット未納のヘッダー表示・精算入口

10/13議事録の「カルテヘッダーにて、他の保有ペットの未納分もみれて、精算するできるようにする」と、2026-10-18 01:53:54 UTCの本人回答「ok」（Sentinel_2c3a5d7c10d8819190057357540b7f9a）により、カルテのヘッダーに、同じ飼主の他ペットに未納があることを表示し、そこから既存の精算画面へ進める入口を設ける。スタッフ向けの既存カルテで、当該飼主と他ペット・元請求の対応を保持する。

表示と遷移は既存の医院スコープ・閲覧／精算権限に従う。同じ飼主であることだけを理由に認可外の医院・ペット・請求情報を公開しない。遷移先では[既存の飼主単位の請求選択・元請求への回収](../../accounting/lifecycle.md#requirement-fr-accounting-owner-selected-charges)に従い、複数ペット・当日／過去分の選択、自由な支払額、元請求と医院への帰属、医師確認・請求確定・入金／訂正の既存条件を維持する。

この入口の表示・操作だけでヘッダーから直接決済したり、他ペット分を自動選択・一括回収したり、新しい請求を作ったりしない。未納の有無を知らせることと既存精算への遷移が今回の採用範囲であり、未納金額の表示追加・形式、件数、バッジ等の具体配置／表現や取得・更新方式は今回選定しない。取得失敗や確認不能を「未納なし」や精算完了として表示しない。通常のカルテ編集・自動保存と既存入力を維持する。

| 条件ID | 前提／操作 | 期待結果 | 根拠・検証状態 |
|:---|:---|:---|:---|
| AC-medical-records-other-pet-unpaid-header-entry | 同じ飼主の他ペットに、認可範囲内の未納があるカルテを開き、ヘッダーの精算入口を操作する | 他ペットに未納があることを確認でき、同じ飼主の既存精算画面へ進める。遷移先で既存の対象請求・ペット選択と元請求帰属を維持し、入口だけで決済・自動回収・新規請求・カルテ保存を発火しない | 10/13議事録、2026-10-18本人ok、D10既存精算。未実施 |
| AC-medical-records-other-pet-unpaid-header-scope | 他の飼主・認可外医院／ペットの情報を含む候補、権限不足、または未納情報の取得失敗がある状態でヘッダー／入口を確認する | 認可外情報を公開せず、別飼主の精算へ誤接続しない。失敗・確認不能を未納なしや精算完了へ置換せず、既存入力・通常自動保存と既存精算条件を維持する | 既存scope／認可と今回入口の適用境界。未実施 |

<a id="cross-ui-search-20261019"></a>
## 10/19後続承認：同種操作・検索ボタン方式

本画面の既存キーワード欄は[適用台帳 H01](../../../shared/interaction-scope.md)と[共通FR/8AC](../../../shared/ui-contracts.md#interaction-consistency-20261019)へ従う。入力中/適用済みを分け、入力/待機/消去だけでは検索せず検索ボタンで適用。既存の明示リセット、独立filter、詳細復帰、認可・検索対象・取得上限/ページ内制限を保持し、全件検索や一律の保存方式へ変えない。取得中/0件/失敗と遅延応答を分け、検索で選択/保存/コピー/送信を確定しない。旧defer/debounceの記載は入力即適用の目標に用いず、適用後の描画最適化と区別する。[S14手順](../../../../ops/testing/scenarios/S14-search-and-multi-word.md#cross-ui-search-20261019)を各対象IDで実施する受入条件とし、実装適合・実機は未確認。

### 保存時氏名の追加受入（EMR-368）

| 条件ID | 操作 | 期待結果（未実施） |
|:---|:---|:---|
| AC-medical-records-main-author-save-time-name | 初回実入力を正常保存後にスタッフ名を変更し、再表示。その後別名で実編集を正常保存 | 改名だけでは保存時の初回・最終氏名を変更しない。実編集の正常保存で最終氏名をその保存時名へ更新し、最初のID／保存時名を保持。失敗保存や空カルテ作成を初回実入力としない |

物理保存先・既存全writerでの同時成立・表示/API・旧履歴の欠落表現は実装設計の残件。今回のモデル整合から未定の氏名列を追加しない。

<a id="requirement-fr-medical-records-plan-domain-tabs"></a>
## FR-medical-records-plan-domain-tabs：治療プランの予定を検査・予防接種・定期健診タブで追跡する

| 項目 | 状態 |
|:---|:---|
| 決定状態 | 2026-10-20本人承認。検査連動の提案への「ok」と、予防接種・定期健診への同様適用を採用 |
| 要件責任者・決定者 | 依頼者 MinoruSoga。以下の本人回答に基づく。実装担当者・各writer ownerとは区別 |
| 実装状態 | 未実装差分を[技術契約](../../../../design/domains/medical-records/technical-spec.md#plan-domain-tabs-target-gap)で分離。今回コード・API・DB変更なし |
| 実機検証状態 | 未実施。下記ACは目標でありPASSではない |

根拠は、本人 Sentinel_11913f0f582481918f93b98e4f19f374 の検査表示要望、提案 Sentinel_4a0d0b7b7f348191b6af93b542806766（未実施の予定表示・同じ予定への機器／手入力結果紐付け・二重登録回避）に対する本人 Sentinel_450f505598f88191bcdea5c4ac422eb3「ok」、および本人 Sentinel_1c747dcd206c8191a23800cc36e1666c「予防接種タブ、定期健診タブも同様ですね」。目的は同じ予定の再入力と、予定・実績の別々の追跡を減らすこと。評価対象はプラン追加後に各タブで同じ予定を再登録する入力回数と、同じ結果本文を二重転記する入力回数で、いずれも不要（0回）になることを既存10ACの追跡・重複防止条件として確認する。基準状態は技術契約に示す未接続の現行経路で、現場の基準回数・所要時間や改善実測値は未計測。実装前に同じ3種別の操作を記録して比較する。測定済み・実装着手可の宣言ではない。

治療プランに追加した検査・予防接種・定期健診の予定を、それぞれのタブに初期状態の未実施として表示する。その後は同じ対応先タブで元項目の現在状態を識別し、後日予定は「後日予定」、中止は「中止」として未実施へ戻さない。後日予定は既決の未完了対象として保持し、実施への変更は既存の明示操作に従う。中止は実施候補・未完了警告の対象にせず、参照と訂正は既決owner手順へ渡す。新しい中止解除／再実施権限は追加しない。元の治療プラン項目を安定して識別し、実施／結果記録ができた後も同じ予定との対応を辿れるようにする。名称・日付・行位置の一致だけで別の予定へ自動紐付けしない。同名／同日の別予定を区別し、タブの再表示・保存再試行から予定や実績を重複登録しない。今回の予定表示と既存の保存済み履歴は識別する。

マスタから追加する業務上の予定種別は既存の[診療マスタの区分](../../../screens/settings/master-treatment.md)に従い、D07「検査」はexam_typesのID、D08「予防接種」はvaccinesのID、D09「定期健診」はcheckup_typesのIDと医院・種別を明示して選択／保持する。元プラン項目IDと元マスタIDは別で、同じ数値ID・名称でも種別が違えば別対象。現行検索のcheckupTypes由来「検査」やotherだけからD07へ推測振分けせず、元種別／参照が不明な旧行や自由記載行は確認対象とする。既存の自由記載を禁止せず、参照未解決のまま結果ownerを推測して実績生成・関連付けしない。表示selectorの配置とwireの形は技術設計へ残すが、3種別の業務的区別は本FRの前提とする。

| 対応先 | 予定から実施／結果へ渡す契約 |
|:---|:---|
| 検査 | [D07の患者帰属・結果原値](../../examinations/lifecycle.md#fr-examinations-record)、[受信・患者選択](../../examinations/lifecycle.md#fr-examinations-receive)、[同一性・attach／detach](../../examinations/lifecycle.md#fr-examinations-idempotency)を維持し、機器受信又は手入力した結果を同じ予定に関連付ける。機器受信の患者指定と予定の特定は別の照合で、同じ患者というだけで予定を自動決定しない |
| 予防接種 | [D08の一接種一記録](../../vaccinations/lifecycle.md#fr-vaccinations-record)と[訂正保護](../../vaccinations/lifecycle.md#fr-vaccinations-correction)を保ち、実接種日・名称・lot等の同じ保存済み接種を予定から参照する。予定自体を接種実績へ変換せず、[次回予定日](../../vaccinations/lifecycle.md#fr-vaccinations-schedule)・証明書発行版の既決を維持する |
| 定期健診 | [D09のCheckup正系統](../../checkups/lifecycle.md#fr-checkups-canonical)と[カルテ帰属](../../checkups/lifecycle.md#fr-checkups-record)、型付き結果・保存時snapshot・訂正保護を維持し、同じ健診予定に実施／結果を関連付ける。検査へ同じ健診を二重保存しない |

結果本文・接種実績・健診結果の正本は各ownerに置き、プラン側で別の編集可能な本文を作らない。実施状況、結果待ち／原値あり、取得失敗／対応不明、検査の確定状態を分ける。予定追加・通常保存・診療終了、結果未到着だけを根拠に、実施済み・正常・接種済み・健診完了・検査確定へ進めない。予定追加だけから機器オーダー送信・接種証明書発行・外部通知・請求確定を起動しない。予定追加・通常保存・診療終了だけで実出庫／在庫減算を起動せず、実移動は[既決の在庫owner境界](../../inventory/business-rules.md#requirement-fr-inventory-clinical-boundary)に従う明示操作へ渡す。予定表示のために物理資源を作る場合も、既存createの減算副作用を流用しない。[D10の実施対象](../../accounting/lifecycle.md#requirement-fr-accounting-billable-performance)と元明細対応を使い、プランと実績の両方を別明細として二重請求しない。見積・過去記録を無断で再生成しない。

保存済み関連付けは医院・患者・カルテ・元プラン項目と対象ownerのID／必要版を照合し、現在の閲覧／操作権限を維持する。別患者・別カルテ・古い版への遅延応答を適用しない。未保存行は保存後の同じ項目との対応を維持し、保存前に架空の実績を作らない。競合・保存結果不明・対応候補複数／不明では元操作と保存状態を照合し、無条件再createや推測紐付けをしない。誤関連と結果自体の訂正は下記の責務を区別し、変更前後／理由／actor／時刻・元版を保持して、通常FK付替えや履歴削除で代用しない。

関連そのものの誤り（同じ患者の正しい結果を別の予定へ関連付けた場合を含む）は、D03の元予定・関連を管理するwriter（共通契約窓口EMR-516、各consumer517／500）が訂正責務を持つ。既存の現在医院・患者・カルテとプラン編集の操作権限、対象ownerの参照／関連操作の保護、旧関連と新関連の必要版を照合する専用関連訂正として、元／先予定・実績ID／版・変更前後・理由・actor／時刻を履歴へ対応付ける。新grantは推定せず実装設計で既存権限との対応を確認する。関連だけの訂正で正しい臨床結果を改訂したり検査確定を解除したりしない。結果自体に誤りがある場合だけ各結果ownerの専用訂正へ渡す。関連保存の具体的API／transaction参加方式は未設計。

関連済み項目の通常削除・名称／元種別・元マスタ等の意味変更でも、元予定ID・当時の意味・関連先／版と履歴を失わない。削除や別内容への転用で関連だけを無警告に残さず、参照・請求等の影響を確認して上記専用訂正へ渡す。削除を拒否するか、取消等の履歴付き状態で元予定を引き続き参照可能にするかは技術設計で選定する。通常編集のたびに全項目をロックする新規業務制約は設けず、memo等の変更と元予定の意味変更を区別する。

プランの未実施／後日予定／実施済み／中止は[実施状態の既決](#requirement-fr-medical-records-plan-execution-states)へ従い、実施操作と結果登録を同一イベントに固定しない。中止・タブクリア／複製取消だけで受信済み結果、接種実績、健診結果を自動破棄・巻戻ししない。既に実施／請求等に使われた場合は影響を示して各ownerの訂正へ渡す。通常受付・過去専用入力・死亡後生前記録の既決条件を迂回する新しい入口は作らない。

### 追加受入条件（1FR／10AC、各条件独立、実機未実施）

| AC ID | Given | When | Then | 成功条件・証跡状態 |
|:---|:---|:---|:---|:---|
| AC-medical-records-plan-tabs-examination | D07の元種別／マスタIDを持つ今回の検査予定を追加、まだ実施しておらず結果未到着。実施済み・結果待ちの別予定もある | 検査タブを開く | 前者は同じ予定を未実施として表示し、後者は実施済み・結果待ちと区別。保存済み結果履歴と識別し、正常／確定へ補完しない | 個別対照、未実施 |
| AC-medical-records-plan-tabs-vaccination | D08の元種別／マスタIDを持つ予防接種予定を今回プランへ追加 | 予防接種タブを開き実接種を別途記録 | 予定表示から同じ接種を参照でき、追加だけでは接種実績・証明書・次回日を生成しない。実績の既決保存条件を維持 | 個別対照、未実施 |
| AC-medical-records-plan-tabs-checkup | D09の元種別／マスタIDを持つ健診予定を今回プランへ追加 | 定期健診タブを開き型付き結果を記録 | 未実施の予定から同じCheckup結果を追跡し、追加だけで完了にしない。Examinationへ二重保存しない | 個別対照、未実施 |
| AC-medical-records-plan-tabs-result-identity | 同名／同日を含む複数の検査予定 | 機器結果又は手入力結果を関連付け再表示 | 医院・患者・カルテと予定を照合し、対応した同じ結果ID／版を参照。結果本文の二重編集や別予定への推測関連なし | 各入力経路、未実施 |
| AC-medical-records-plan-tabs-scope | 別医院／患者／カルテ、権限変更、古い応答がある | 表示／関連付けを試す | 現在scope・権限・必要版を照合し対象外へ適用しない。閲覧をwrite許可にしない | 各不一致、未実施 |
| AC-medical-records-plan-tabs-unknown | 3種それぞれ保存断・結果不明又は対応候補複数 | 再表示／再試行 | 元操作・保存済みIDを照合し、重複create・推測関連・0件／正常への置換なし | 各種／不明条件、未実施 |
| AC-medical-records-plan-tabs-correction | 同患者内の誤関連、結果訂正版、又は関連済み予定の意味変更がある | D03の専用関連訂正又は結果ownerの訂正を使う | 関連だけの誤りは結果本文を改訂せず、D03が元／先予定・実績ID／版と前後・理由／actor／時刻を履歴付きで訂正。意味変更で旧意味・関連を無警告に転用せず、結果自体の訂正は結果ownerへ。通常FK付替えで代用しない | 各owner、未実施 |
| AC-medical-records-plan-tabs-cancel | 結果又は実績が既に存在する予定 | 中止／クリア／複製取消／関連済み項目の通常削除を試す | 同じ対応先タブで後日予定／中止を未実施へ戻さず識別し、中止を実施候補・未完了警告に含めない。既決owner保護に従い実績の自動破棄なし。削除は拒否又は履歴付き取消等で元予定ID・当時意味・関連を参照可能にし、通常読取から対応を消失させない。実施済みへの変更だけで結果確定なし | 各操作／3種、未実施 |
| AC-medical-records-plan-tabs-side-effects | 3種の未実施予定 | 追加／通常保存／診療終了 | 機器送信・証明書発行・外部通知・請求確定・実出庫／在庫減算・見積／過去記録再生成なし。実移動は既決ownerの明示操作へ渡し、予定用createの減算流用なし。結果待ちは完了ではない | 各操作／3種、未実施 |
| AC-medical-records-plan-tabs-billing | 同じ予定と実績の対応あり | 会計対象を取得する | D10の元明細・実施適格性で未請求取得から同じ対象を扱い、プランと実績から二重請求しない。会計確認済みでも未実施／後日予定／中止を当日請求候補へ入れない | D10対照、未実施 |

### 技術設計へ残す事項

パネル／複数結果と一つの予定の多重度、物理的な検査等の作成時点かprojectionによる表示か、関連参照のwire／schemaは技術設計へ委ねる。各ownerの既存多重度を一律1対1へ変更せず、受信結果から対応予定が一意に決まる方式も未選定。業務上の同一予定追跡・状態分離・重複防止の採用とは区別する。
