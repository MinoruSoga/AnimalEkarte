# 検査入力/結果登録 仕様書 (Examination Form)

> 2026-10-20追加目標：カルテ内の検査タブは[治療プランの同一予定追跡FR／10AC](../domains/medical-records/pages/medical-record-form.md#requirement-fr-medical-records-plan-domain-tabs)へ接続し、追加した予定を未実施として示す。保存済み実績／結果と区別し、予定追加だけで実施・結果確定・証明書発行・健診完了・機器送信・請求確定を行わない。実績の既決入力・訂正・認可は[領域正本](../domains/examinations/lifecycle.md#examinations-plan-tab-link)を維持し、別本文の二重入力を作らない。[設計差分](../../design/domains/examinations/technical-spec.md#plan-tab-link-design)は未実装／実機未確認。独立一覧・過去履歴の全件を未実施予定へ置換する意味ではない。

## 2026-10-15 D07仕様整理の優先順位

[死亡後の生前検査記録](../domains/examinations/lifecycle.md#examination-deceased-retrospective)は死亡状態を戻さず理由・履歴付きの専用後日入力／誤記訂正へ接続する。通常の新たな検査・機器待機の許可には広げず、現行フォーム／APIの専用対応は未確認。

最新の第3項合意で担当医は任意とし、未選択でも検査を保存できる。入力者の認証ログは担当医属性と別に残す。現行の担当医必須判定は解消対象の実装差で、保存／ログの適合確認は未実施。

業務・例外・GWTは[検査ライフサイクル](../domains/examinations/lifecycle.md)、現wireとの差は[技術契約](../../design/domains/examinations/technical-spec.md)を優先する。以下のUI／図は現行参照であり、保存するだけで一律に確定する意味ではない。

- confirmedは全ロック。revisionなしcompletedは結果封印、revisionありcompletedは解除後の作業版となり得る。
- 専用unconfirmは権限・理由・official版保護が必要。旧実装の親draft条件は最新D03の通常編集契約への適合差として分け、検査の解除でカルテの旧確定履歴を消去しない。
- 旧PUT /itemsは未送信／nullも全消去へ変換する現状差がある。D03継承目標は明示[]だけを意図的全消去とし、旧経路適合済みとしない。
- 日常機器受信はEMR-402のブラウザWeb Serial（NX600/AU10V、ページ表示中）。ADRの歴史的LaunchAgent説明を現在の通常経路にしない。
- HIGH/LOWは表示補助。取得不能／評価不能を正常確認済みとせず、基準値具体値の医師確認とソフト仕様整理を分ける。

## 概要
- **画面の目的**: 検査オーダーの作成、および臨床検査値（数値・判定）の精緻な記録。
- **URLパターン**: 
  - 検査管理からの新規: `/examinations/new?petId=xxx`
  - 未紐付け・検歴: `/examinations/:id`
- **アクセス権限**: `examinations` リソースの `view` 権限（ルートを `RequirePermission` でガード。新規作成ルート `/examinations/new` は `create` アクション必須）。操作ボタンの出し分けは `usePermission` で制御。

---

## 1. 画面構成

### 1.1 検査基本情報
- **検査種別**: 血液、尿、エコー等のマスタ（`exam_types`）から選択。既存の必須条件を維持する。
- **担当医**: 選択時の候補は `staffType=doctor` かつ active のスタッフ。最新合意では任意で、未選択でも保存可。入力者やカルテ担当医を自動代入せず、担当医未選択と入力者ログを別に保持する。現行必須UIの適合は未確認。
- **患者ヘッダー**: `PatientInfoCard` で種を `petName` / `petDetails` に実データで出す。特記マーク（EMR-173/231、共有 `DangerBadge`）: ペット名横に `danger_level` 高=赤い八角形アイコン・中=黄い三角形アイコン（文言なし、Popover で補足メモ）、低・未設定は非表示。飼主 `is_dangerous` は飼主名横に文言なしの赤い八角形アイコン（代替名「特記」）。採血・保定など検査実施前の安全確認用。

### 1.2 動的検査項目テーブル (`ExamItemsTable`)
選択した検査種別に基づき、測定項目が動的に生成されます。

| 項目 | 説明 |
|:---|:---|
| **項目名 / 単位** | 例：ALT (U/L), CRE (mg/dL)。マスタより自動展開。 |
| **測定値** | 実測値を数値入力。 |
| **判定 (HIGH/LOW)** | **自動判定**: マスタで定義された基準値（Min〜Max）に基づき、高値は **HIGH**（赤）、低値は **LOW**（status-blue）でハイライトされます。判定はバックエンドが導出するため、保存・再読込後に反映されます。 |
| **基準値参照** | 検査種別マスタで定義された正常範囲（`normal_value`）を横に表示し、臨床判断を助けます。 |

### 1.3 未紐付け受信バナー
選択中ペットの検査画面に、医院の未紐付け機器受信があれば表示する。1クリックで `POST /lab-imports/:job_id/attach`。値は編集しない。確認ダイアログは出さない。待機の遠隔起動はしない。

### 1.4 検査履歴パネル (`ExaminationHistoryPanel`)
画面は 2 カラム構成（左 3/5 = フォーム・右 2/5 = 履歴）で、右カラムに**同一ペットの過去の検査履歴**を常時表示します。過去の数値と見比べながら入力できるため、別画面への往復が不要です。
- **絞り込み**: 期間（開始日・終了日）、キーワード（検査種別・結果サマリーをかな正規化で部分一致）、日付の昇順/降順切替、一括クリア。
- **編集中レコードの除外**: 編集モードでは、開いている検査自身は履歴から除外されます。
- **データ源**: `useExaminationHistoryFilters` が `useGetExaminations` へ `petId` / `startDate` / `endDate` を渡し、サーバ側でペットと期間を絞る（wire は `pet_id` / `start_date` / `end_date`）。page/limit は未指定のため BE 既定の先頭20件。取得後に同一ペットの再確認、キーワード検索、編集中レコード除外、ソートを行う。`?historyView=pivot` でピボット表示。

---

## 2. 主要な臨床・安全機能

### 2.1 真正性の担保（確定ロック）
- **ステータス**: サーバ保存後の `確定 (confirmed)` は全ロック。revision 無しの `completed` は結果を封印する。
- **確定解除**: `POST /examinations/:id/unconfirm`（権限 `examination-unconfirm:edit`）。
- **印刷**: 保存済み print-snapshot。

検査レコードの確定ロック状態：

```mermaid
stateDiagram-v2
    state "入力可能" as editing
    state "確定 confirmed（全ロック）" as confirmed
    state "完了 completed" as completed
    [*] --> editing : 新規作成・結果入力
    editing --> confirmed : 保存で確定
    editing --> completed : revision 無しで保存
    confirmed --> editing : 確定解除（POST unconfirm）
    completed --> [*] : 結果を封印
```

### 2.2 保存時ハイライト
数値の判定（HIGH/LOW）はバックエンドが基準値（Min〜Max）から導出し、保存・再読込のタイミングで UI の色に反映されます。これにより、多忙な診察室でも重要値の見落としを防止します。

### 2.3 未保存データ保護（`NavigationBlocker`）
フォームに未保存の編集がある状態（`useUnsavedChanges` の dirty フラグが立ち、かつ保存処理中でない）で他画面へ遷移しようとすると、`NavigationBlocker` が遷移をブロックし確認を求めます。入力途中の検査値が黙って失われることを防ぎます。保存成功・削除確定時は dirty フラグを解除してから遷移するため、正常フローでは警告は出ません。

### 2.4 削除の保護と権限制御
- **削除**: `ConfirmDialog`（destructive）を経由し、取り消せない旨を明示してから実行します。
- **権限**: 新規ルート入口の `create` と、保存mutationの認可を区別する。新規保存は `examinations:create` と `examinations:edit` の両方、既存編集は `edit` を要求する。フォームの入力不可表示だけで保存認可を代用しない。
- **バリデーションエラー時のフォーカス移動**: 保存失敗時は最初のエラーフィールドへ自動フォーカス・スクロールします（アクセシビリティ対応）。

---

## 3. 技術仕様

### 3.1 構成コンポーネント
- **`ExaminationForm`**: 統合フォーム。
- **`ExaminationFormFields`**: 検査種別・担当医等の基本情報フィールド。
- **`ExamItemsTable`**: 選択された検査種別の項目テンプレ（`exam_type_fields`、`useGetExamTypeFields` で取得）を描画する動的テーブル。確定済み検査では入力不可。
- **`ExaminationHistoryPanel`**: 右カラムの同一ペット検査履歴（§1.4）。
- **`NavigationBlocker`** + **`useUnsavedChanges`**: 未保存データ保護（§2.3）。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/examinations/:id` | 検査基本情報の取得（項目リストは `/:id/items` で別途取得） | `examinations` | `view` |
| POST | `/api/v1/examinations` | 新規保存 | `examinations` | `create` かつ `edit` |
| PATCH | `/api/v1/examinations/:id` | 数値・判定・ステータスの更新 | `examinations` | `edit` |
| DELETE | `/api/v1/examinations/:id` | 検査記録の削除（編集時のみ削除ボタンを表示） | `examinations` | `delete` |
| GET | `/api/v1/examinations/:id/items` | 検査項目（結果）のリスト取得 | `examinations` | `view` |
| PUT | `/api/v1/examinations/:id/items` | 検査項目（結果）の一括更新 | `examinations` | `edit` |
| POST | `/api/v1/examinations/:id/unconfirm` | 確定解除 | `examination-unconfirm` | `edit` |
| GET | `/api/v1/examinations/:id/print-snapshot` | 印刷用スナップショット | `examinations` | `view` |
| GET | `/api/v1/masters/examination-types` | 利用可能な検査種別リストの取得 | `master-medical` | `view` |
| GET | `/api/v1/masters/examination-types/:id` | 検査種別詳細（項目テンプレ `exam_type_fields`）の取得 | `master-medical` | `view` |

---

## カルテへの後日結果・訂正履歴の接続

2026-10-14の[カルテ追加合意](../domains/medical-records/pages/medical-record-form.md#requirement-fr-medical-records-late-result-shared-history)により、検査画面とカルテは同じ検査結果・変更履歴を参照する。最新の[通常編集契約](../domains/examinations/lifecycle.md#examination-normal-edit-autosave)により診療終了後もカルテ本文を編集できる。原記載・参照版と変更履歴を保持し、検査自身の確定ロックは別に維持する。患者／医院／検査／カルテ対応を照合し、結果訂正の前後・理由・操作主体／時刻を履歴へ残す。

検査ownerの確定／訂正・専用unconfirm権限は本書の既存契約へ従う。検査のunconfirmは検査自身の専用操作であり、カルテの旧確定履歴の消去や患者関連の普通FK変更を許可しない。カルテ本文の通常編集は最新D03契約へ従う。[参照版と追記の技術契約](../../design/domains/medical-records/technical-spec.md#medical-record-integration-contracts)へ接続し、revisionなし旧結果・機器受信／adapter全経路はD07の文書照合に残す。今回検査内部の新しい解除手順・実装・実機受入を確定しない。

### カルテへの返却・利用開始の最終照合

[最終型／writer契約](../../design/domains/medical-records/technical-spec.md#medical-record-writer-matrix)へ接続し、院内scope・検査ID・患者／record対応・固定revision／版内item ref・変更理由／actor／時刻・取得結果を返す。原記載の確定時参照版と後日版を分離。現在のrevision／print-snapshotの存在だけでカルテadapter実装済みとしない。診療終了後のカルテ本文編集を、患者関連の普通FK変更／旧確定履歴消去で代用しない。旧結果／項目の明示対応を確認できなければ停止し、既存の検査内部確定／訂正手順を保持する。

<a id="examination-selected-print-ui"></a>
## 項目選択印刷・患者トップへの結果連動（2026-10-15追加採用）

[直接回答による追加採用](../domains/examinations/lifecycle.md#examination-print-infection-agreement-20261015)の目標。§2.1の保存済みprint-snapshotを基礎として、同じ検査・固定版内で印刷項目を選択し、プレビューと出力を一致させる。実施済みを初期選択し未実施を候補から除外する。陰性・0・不等式・定性値を空欄扱いしない。実施済み結果待ちと実施不明は別表示とし、不明行は自動選択せず理由を表示して元記録照合へ。選択0件は印刷対象なし、取消は入力保持。未保存入力は反映されないと表示し、印刷で保存・確定・解除・送信しない。

項目を絞っても患者・検査日・出典／保存版・訂正識別と保存当時の原値／単位／基準範囲を保持する。別検査日・別結果版は一つの帳票へ合成しない。新しい複数日比較帳票・全検歴UI・グラフは対象外。対象版／scopeが変われば選択を新結果へ流用せず再プレビューする。GWTはAC-examinations-print-item-selection〜single-snapshot、[技術契約](../../design/domains/examinations/technical-spec.md#examination-selected-print-contract)へ。

院内の猫FIV／FeLV・犬mf／HWの保存結果は患者トップへ自動反映する。他院はD07の結果記録として手入力し、原値・判明検査日・出典と入力者／入力時刻を保存する。未判明日付は「検査日不明」、他院は「他院・手入力」と表示する。現行フォームに日付不明／出典を表現できる保証はなく、架空日付や台帳備考への転記で代用しない。

保存結果と同じ固定refを読むので、未保存値や未紐付け受信は患者トップへ出さない。訂正／作業版の状態、同日複数候補・出典の違い、記録なし／結果待ち／取得失敗を区別する。保存成功後の再取得失敗は更新未確認として扱う。検査自身の訂正保護と担当医任意は維持し、LINE公開・送付を追加しない。受入はAC-examinations-infection-*、[結果projection契約](../../design/domains/examinations/technical-spec.md#examination-infection-summary-contract)へ。

<a id="cross-ui-search-20261019"></a>
## 10/19後続承認：同種操作・検索ボタン方式

本画面の既存キーワード欄は[適用台帳 H02](../shared/interaction-scope.md)と[共通FR/8AC](../shared/ui-contracts.md#interaction-consistency-20261019)へ従う。入力中/適用済みを分け、入力/待機/消去だけでは検索せず検索ボタンで適用。既存の明示リセット、独立filter、詳細復帰、認可・検索対象・取得上限/ページ内制限を保持し、全件検索や一律の保存方式へ変えない。取得中/0件/失敗と遅延応答を分け、検索で選択/保存/コピー/送信を確定しない。旧defer/debounceの記載は入力即適用の目標に用いず、適用後の描画最適化と区別する。[S14手順](../../ops/testing/scenarios/S14-search-and-multi-word.md#cross-ui-search-20261019)を各対象IDで実施する受入条件とし、実装適合・実機は未確認。
