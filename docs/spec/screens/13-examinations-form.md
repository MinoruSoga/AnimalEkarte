# 検査入力/結果登録 仕様書 (Examination Form)

## 概要
- **画面の目的**: 検査オーダーの作成、および臨床検査値（数値・判定）の精緻な記録。
- **URLパターン**: 
  - 検査管理からの新規: `/medical-records/new?petId=xxx&tab=検査`（当日カルテの検査タブ）
  - 未紐付け・検歴: `/examinations/:id`
- **アクセス権限**: `examinations` リソースの `view` 権限（ルートを `RequirePermission` でガード。新規作成ルート `/examinations/new` は `create` アクション必須）。操作ボタンの出し分けは `usePermission` で制御。

---

## 1. 画面構成

### 1.1 検査基本情報
- **検査種別**: 血液、尿、エコー等のマスタ（`exam_types`）から選択。
- **担当医**: `staffType=doctor` かつ active のスタッフ。検査種別とともに保存時の必須項目。
- **患者ヘッダー**: 種を `petName` / `petDetails` に実データで出す。

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
- **データ源**: `useGetExaminations`（page/limit 未指定 → BE 既定 20 件）をクライアント側でペット ID フィルタする。同一ペットの履歴は最大 20 件。`?historyView=pivot` でピボット表示。

---

## 2. 主要な臨床・安全機能

### 2.1 真正性の担保（確定ロック）
- **ステータス**: サーバ保存後の `確定 (confirmed)` は全ロック。revision 無しの `completed` は結果を封印する。
- **確定解除**: `POST /examinations/:id/unconfirm`（権限 `examination-unconfirm:edit`）。
- **印刷**: 保存済み print-snapshot。

### 2.2 保存時ハイライト
数値の判定（HIGH/LOW）はバックエンドが基準値（Min〜Max）から導出し、保存・再読込のタイミングで UI の色に反映されます。これにより、多忙な診察室でも重要値の見落としを防止します。

### 2.3 未保存データ保護（`NavigationBlocker`）
フォームに未保存の編集がある状態（`useUnsavedChanges` の dirty フラグが立ち、かつ保存処理中でない）で他画面へ遷移しようとすると、`NavigationBlocker` が遷移をブロックし確認を求めます。入力途中の検査値が黙って失われることを防ぎます。保存成功・削除確定時は dirty フラグを解除してから遷移するため、正常フローでは警告は出ません。

### 2.4 削除の保護と権限制御
- **削除**: `ConfirmDialog`（destructive）を経由し、取り消せない旨を明示してから実行します。
- **権限**: 編集権限（新規は `create`・編集は `edit`）が無い場合はフォーム全体が入力不可になります。
- **バリデーションエラー時のフォーカス移動**: 保存失敗時は最初のエラーフィールドへ自動フォーカス・スクロールします（アクセシビリティ対応）。

---

## 3. 技術仕様

### 3.1 構成コンポーネント
- **`ExaminationForm`**: 統合フォーム。
- **`ExaminationFormFields`**: 検査種別・担当医等の基本情報フィールド。
- **`ExamItemsTable`**: 選択された検査種別の項目テンプレ（`exam_type_fields`、`useGetExamTypeFields` で取得）を描画する動的テーブル。確定済み検査では入力不可。
- **`ExaminationHistoryPanel`**: 右カラムの同一ペット検査履歴（§1.3）。
- **`NavigationBlocker`** + **`useUnsavedChanges`**: 未保存データ保護（§2.3）。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/examinations/:id` | 検査基本情報の取得（項目リストは `/:id/items` で別途取得） | `examinations` | `view` |
| POST | `/api/v1/examinations` | 新規保存 | `examinations` | `create` |
| PATCH | `/api/v1/examinations/:id` | 数値・判定・ステータスの更新 | `examinations` | `edit` |
| DELETE | `/api/v1/examinations/:id` | 検査記録の削除（編集時のみ削除ボタンを表示） | `examinations` | `delete` |
| GET | `/api/v1/examinations/:id/items` | 検査項目（結果）のリスト取得 | `examinations` | `view` |
| PUT | `/api/v1/examinations/:id/items` | 検査項目（結果）の一括更新 | `examinations` | `edit` |
| POST | `/api/v1/examinations/:id/unconfirm` | 確定解除 | `examination-unconfirm` | `edit` |
| GET | `/api/v1/examinations/:id/print-snapshot` | 印刷用スナップショット | `examinations` | `view` |
| GET | `/api/v1/masters/examination-types` | 利用可能な検査種別リストの取得 | `master-medical` | `view` |
| GET | `/api/v1/masters/examination-types/:id` | 検査種別詳細（項目テンプレ `exam_type_fields`）の取得 | `master-medical` | `view` |

---

