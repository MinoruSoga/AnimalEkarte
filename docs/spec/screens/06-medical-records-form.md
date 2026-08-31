# カルテ詳細・入力 仕様書 (Medical Record Form)

## 概要
- **画面の目的**: 臨床現場における電子診療録（SOAPS形式）の作成、検査、処置、処方、および会計連携の統合管理。
- **URLパターン**: 
  - 新規作成: `/medical-records/new?petId=xxx`
  - 編集: `/medical-records/:id`
- **アクセス権限**: 新規作成 (`/medical-records/new`) は `medical-records:create` を要求する `RequirePermission` ルートガードあり。編集 (`/medical-records/:id`) は親ルート `/medical-records` の `medical-records:view` ガードのみを継承し、保存可否はコンポーネント内の `usePermission`（`ResourceMedicalRecords`）で制御。

---

## 1. 画面構成 (9 タブ構成)

膨大な診療情報を整理するため、以下の 9 つのタブで構成されています。

| タブ | 名称 | 内容 |
|:---:|:---|:---|
| 1 | **問診** | 飼主からの主訴、現在の症状、生活環境のヒアリング記録。 |
| 2 | **診察/治療プラン** | 身体検査所見（SOAPのS/O/A）、診断名（第1・第2）、今後の治療方針（Plan）。 |
| 3 | **治療** | 実施した処置、処置項目、および処方薬の明細。 |
| 4 | **予防接種** | 左ペインに当該ペットの接種記録一覧（`GET /vaccinations?pet_id=` + `page`/`limit`）。空なら EmptyState、「記録を追加」で入力フォーム。右カラムは `VaccinationHistory`。カルテネストの `/medical-records/:id/vaccinations` は存在せず使わない。次回予定ラジオは `calculateNextDate` に配線済み（既定 `4weeks`。独立画面の既定は `1year`）。 |
| 5 | **定期健診** | 健康診断の結果記録。`/checkups` の編集は `?tab=定期健診&checkupId=` で本タブを開き、該当行を編集状態にする。 |
| 6 | **検査** | 検査結果の一覧表示と既存検査の取込（カルテへの紐付け）、基準値との比較（HIGH/LOW判定ハイライト）。 |
| 7 | **画像** | 患部写真、レントゲン、エコー画像、PDF資料のアップロード・管理。 |
| 8 | **見積書** | 提示した概算費用の管理。 |
| 9 | **会計(医師確認)** | 診療費の最終確認と会計ステータスの送信。 |

---

## 2. 主要な臨床ロジック

### 2.1 SOAPS 形式の採用
- **Subjective (主観的)**: 飼主の話。
- **Objective (客観的)**: 視診・触診・聴診結果。
- **Assessment (評価)**: 診断、病名特定。
- **Plan (計画)**: 今後の検査・治療予定。
- **Supplement (補足)**: 注意事項等。

### 2.2 保存プロセス
- **メイン保存**: 画面右下のフローティング「保存」ボタン（`MedicalRecordFloatingActions`）は、アクティブタブに応じて送信先を切り替える。「問診」タブは `PATCH /medical-records/:id/inquiries`（主訴・主訴区分・治療方針）、「診察/治療プラン」タブは `PATCH /medical-records/:id/clinical-plan`（治療方針・診断詳細・診断名）と `PATCH /medical-records/:id`（次回来院推奨日）を送信する。他タブではカルテ本体の送信は行われない（`PATCH /medical-records/:id` へまとめて送信する方式ではない）。予防接種タブでは偽成功を避けるため外側の保存ボタン自体を表示せず、タブ内の接種記録追加を使う。
- **ヘッダー即時保存**: 担当医・来院種別・診察日・次回予定はヘッダー変更と同時に `PATCH /medical-records/:id` する（保存ボタンを経由しない）。来院種別の成功後は `queryKeys.medicalRecords.detail` を invalidate し、再読込でラベルが戻らないようにする。失敗時はローカル state をロールバックする。appointment 紐付き通常カルテの `date` は予約開始の JST 日付に固定され、変更は BE Conflict（UI は未紐付け時のみ成功する。正本は [99-medical-record-flow.md](./99-medical-record-flow.md) / [reservation-to-record-flow.md](../reservation-to-record-flow.md) §5.5）。
- **アクティブタブの追加保存**: 保存成功直後、その時点で開いているタブが「診察/治療プラン」または「見積書」の場合のみ、`useMedicalRecordPostSave` が対応する登録済みコールバックを追加実行する（他タブ在中時は発火しない。両タブを並行実行することもない）。見積タブは `items` を create/update 同一 tx で置換永続化する（独立画面 `/estimates` はヘッダ金額のみ。詳細は [23-estimate-form.md](./23-estimate-form.md)）。
- **治療・検査等のサブリソース**: 「治療」タブの明細は行単位の追加/編集/削除操作ごとに `/medical-records/:id/treatments...` へ個別・即時送信される（メイン保存とは独立しており、「バックグラウンド並行保存」ではない）。

### 2.3 臨床安全ガード
- **確定ロック**: `finalized`（確定済）ステータスのカルテはバックエンドが更新を拒否し（409）、訂正は追記（addendum）のみ許可することで真正性を担保。確定への遷移は `PATCH /medical-records/:id` の `status` 指定によるもの。画面右下のフローティングアクション（`MedicalRecordFloatingActions`）に「確定する」ボタンが表示され（編集権限あり・保存済み・未確定の場合のみ）、会計(医師確認)が `confirmed` の場合だけ有効になる。未確認・差戻し・取得中・取得失敗・確認状態なしでは「会計確認が未完了です」として確定を物理ブロックする。有効時は `MedicalRecordFinalizeDialog` で不可逆であることを確認した上で確定する。確定取り消し（unfinalize）API は存在しないため、確定後の修正経路は訂正追記（addendum）のみ。会計完了時の自動確定は現状存在しない。確定済みカルテはサイドヘッダーに「確定済」バッジ（`StatusBadge`）を常時表示する。
- **訂正追記モーダル**: `AddendumModal` の修正内容・修正理由は controlled input。バリデーション失敗後も入力済みの値を保持する（React 19 `useActionState` の remount で消えない）。修正理由は 500 文字以内。
- **薬量自動計算と絶対上限ゲート（#201）**: 「治療」タブの処方明細（`TreatmentRow`）は、対象ペットの species と当日 vital 体重から数量を自動プリフィルする（`calculateDose`）。保存値がマスタ上限（体重連動上限 weight×max_mg/kg と体重非依存の絶対上限 absolute_max_dose の小さい方）を超える場合、フロントエンドは理由をインライン表示して追加・更新を送信せず、バックエンドも Create/Update の永続化前に 400 で拒否する。`ConfirmDialog` による解除経路は設けない。下限未満または推奨値からの著しい乖離は保存を許可して audit に記録する。体重未記録・species 正規化不能・投与量パラメータ未設定時は評価をスキップして従来どおり保存を継続する。パラメータ取得の非 NotFound エラーと species 不一致は既存どおり fail-closed とする。権限付き例外フロー（Design B）は実装しない。
- **未保存警告**: 変更がある状態でページを離れようとすると `NavigationBlocker` が警告を表示。

---

## 3. 技術仕様

### 使用コンポーネント
- **`MedicalRecordForm`**: 統合フォーム。
- **`useMedicalRecordForm`**: 複雑な状態管理（9タブ ＋ 履歴引用）を統括するカスタムフック。
- **`historyItems`**: 「問診」タブ内の右カラム（`InterviewHistory`）に過去の問診履歴を表示。「予防接種」タブは左ペイン一覧と右カラム `VaccinationHistory` の両方を `useGetPetVaccinations`（`GET /vaccinations?pet_id=`）で描画する（フォーム全体で共有される単一サイドパネルではなく、タブごとに独立）。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/medical-records/:id` | カルテ本体の取得（Treatments・Vitals・Owner・Pet 等を同梱。clinical-plan・画像・検査等は個別エンドポイント） | `medical-records` | `view` |
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

---
