# ダッシュボード 仕様書

## 概要
- **画面の目的**: 病院全体の当日受付状況をリアルタイムで把握するホーム画面。カンバンボード形式で患者フローを管理する。
- **URLパターン**: `/`
- **アクセス権限**: 認証済ユーザー全員

## 画面レイアウト
```
┌──────────────────────────────────────────────────────────────┐
│ 当日の受付  [yyyy年M月d日 (E)]  [フィルター] [新規予約ボタン] │
├──────────────────────────────────────────────────────────────┤
│ [フィルターパネル（折り畳み、animate-in）]                    │
│  診察区分: □初診 □再診（NotionCheckbox）                    │
│  指名: [ドクター選択 Select]                                  │
│  種類: □トリミングのみ表示                                   │
├────────────┬────────────┬────────────┬────────────┬──────────┤
│ 受付予約   │ 受付済     │ 診療中     │ 会計待ち   │ 会計済   │
│ [+追加]   │ [+追加]   │            │            │          │
│            │            │            │            │          │
│ カード1    │ カード2    │ カード3    │ カード4    │ カード5  │
│ カード...  │ カード...  │ カード...  │ カード...  │カード... │
└────────────┴────────────┴────────────┴────────────┴──────────┘
```

## カンバンカラム構成

| カラム名 | 説明 | 追加ボタン | 対応ステータス |
|---------|------|-----------|--------------|
| 受付予約 | 本日来院予定の予約 | あり（`/reservations` へ遷移） | `pending`, `confirmed` |
| 受付済 | 来院済み・待合室待機中 | あり（`/reservations` へ遷移） | `checked_in` |
| 診療中 | 診察室・処置室で対応中 | なし | `in_consultation` |
| 会計待ち | 診察終了・会計計算待ち | なし | `accounting` |
| 会計済 | 会計完了・帰宅 | なし | `completed` |

## 表示項目（AppointmentCard）

| フィールド名 | 型 | 説明 | 備考 |
|------------|-----|------|------|
| 時刻 | time | 予約時刻（Clock アイコン付き、等幅フォント） | `appointment.time` |
| 次回予約バッジ | enum | 「次回予約済」=secondary / 「精算未確認」=destructive+AlertCircle | `nextAppointment` |
| 飼い主名 | string | 飼い主氏名（太字） | `ownerName` |
| ペット名 | string | `petType - petName`（Dog アイコン付き） | `petName`, `petType` |
| 初診/再診 | string | visitType バッジ（初診=青背景 / 再診=スレート背景） | `visitType` |
| 診療区分 | string | serviceType バッジ（自動アイコン: トリミング→Scissors / 予防接種→Syringe / 手術→Activity / 診療→Stethoscope） | `serviceType` |
| 担当医 | string | 担当医バッジ（指名時はオレンジ背景＋「指」ラベル、無効スタッフ時は赤背景＋AlertCircle） | `doctor`, `isDesignated` |

## 表示項目（DashboardDetailModal）

| セクション | 項目 |
|-----------|------|
| ヘッダー | 初診/再診アイコン（初/再）、診療区分名、ステータスバッジ（カラム別カラー） |
| 時間カード | 時刻（等幅・大文字）、nextAppointment バッジ |
| 患者情報 | ペット名、ペット種、飼い主名 |
| 診療詳細 | 担当医（未定表示あり）、指名バッジ |

## DashboardDetailModal ステータス別アクション

| ステータス | アクションボタン |
|-----------|---------------|
| 受付予約 | 取消、編集、飼主詳細、「受付済にする」 |
| 受付済 | 飼主詳細、「カルテ作成」（診療時、同時に診療中へ移動）/ 「診察を開始する」（非診療時） |
| 診療中 | 飼主詳細、「診察を終了する」、「カルテ入力」、「検査」（診療時）/ 「施術記録」（トリミング時） |
| 会計待ち | 飼主詳細、「会計へ進む」（精算未確認時は disabled） |
| 会計済 | 飼主詳細、「完了/リストから削除」 |

## UI コンポーネント

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `Dashboard` | `[R]` | メインページ |
| `KanbanColumn` | `[C]` | カラムコンテナ（ドラッグ&ドロップ対応） |
| `AppointmentCard` | `[C]` | 患者カード（`react-dnd` によるDnD対応） |
| `DashboardDetailModal` | `[C][M]` | 患者詳細モーダル（ステータス別アクション） |
| `DashboardSummaryWidget` | `[C]` | 統計サマリーウィジェット（カラム別件数カード＋ネイティブSVGミニスパークライン、ホバーツールチップ） |
| `HospitalizationAlertWidget` | `[C]` | 入院アラートウィジェット（入院中件数・退院超過アラート・入院管理へのクイックリンク） |
| `ReservationFormModal` | `[S][M]` | 予約編集モーダル（カード編集時に利用） |
| `FormHeader` | `[S]` | ページヘッダー（タイトル・日付・アクションボタン） |
| `ConfirmDialog` | `[S][M]` | 予約取消確認ダイアログ |
| `useDashboardKanban` | `[H]` | カンバン状態管理（フィルタ・moveCard・advanceStatus・cancelAppointment・updateAppointment） |
| `useDashboardWeeklyStats` | `[H]` | 週次統計データ収集（localStorage永続化、スパークラインデータ提供） |

## ユーザーアクション

| アクション | トリガー | 処理内容 | 遷移先 |
|-----------|---------|---------|--------|
| カード詳細表示 | カードクリック | `DashboardDetailModal` を開く | モーダル表示 |
| ステータス進行 | 詳細モーダルのアクションボタン | カードを次のカラムに移動 | 同画面 |
| カード編集 | 詳細モーダル「編集」ボタン | `ReservationFormModal` を開く | モーダル表示 |
| 予約取消 | 詳細モーダル「取り消し」ボタン | `ConfirmDialog` 後、カード削除 | 同画面 |
| カード移動 | ドラッグ&ドロップ | ステータス変更（`react-dnd` 使用） | 同画面 |
| 新規予約 | ヘッダー「新規予約」ボタン | 予約管理画面へ遷移 | `/reservations` |
| フィルター | 「フィルター」ボタン | フィルターパネル開閉（slide-in アニメーション） | 同画面 |
| 受付済に追加 | 受付予約/受付済カラム「+」ボタン | `/reservations` へ遷移 | `/reservations` |
| カルテ作成 | 詳細モーダル「カルテ作成」ボタン | カルテフォームへ遷移 | `/medical-records/new?petId=xxx` |

## 画面遷移

| 遷移元 | 遷移先 | 条件 |
|--------|--------|------|
| サイドバー「ダッシュボード」 | `/` | 常時 |
| 「新規予約」ボタン | `/reservations` | ボタンクリック |
| 受付予約/受付済カラム「+」ボタン | `/reservations` | ボタンクリック |
| 詳細モーダル「カルテ作成」 | `/medical-records/new?petId=xxx` | 診療系サービス |
| 詳細モーダル「施術記録」 | `/trimming/new?petId=xxx` | トリミング |
| 詳細モーダル「飼主詳細」 | `/owners/:ownerId` | ボタンクリック |

## フィルター機能

| フィルター | 状態 | 動作 |
|-----------|------|------|
| 診察区分（初診/再診） | `selectedVisitTypes: AppointmentVisitType[]` | チェックを外したvisitTypeのカードを非表示 |
| 指名 | `selectedDoctor: string` | "all"=全表示 / 医師名=指名のみ / "医師指名なし"=指名なしのみ |
| トリミングのみ | `isTrimmingOnly: boolean` | トリミング以外のカードを非表示 |

## 状態管理

| 状態 | 型 | 説明 |
|------|-----|------|
| `columns` | `ColumnData[]` | 5カラム分の予約データ（`useDashboardKanban` 管理） |
| `filteredColumns` | `ColumnData[]` | フィルタ適用後のカンバンデータ（useMemo） |
| `modalOpen` | `boolean` | 詳細モーダル表示フラグ |
| `selectedAppointment` | `Appointment \| null` | 選択中の予約 |
| `isEditModalOpen` | `boolean` | 編集モーダル表示フラグ |
| `isFilterOpen` | `boolean` | フィルターパネル表示フラグ |
| `cancelConfirmOpen` | `boolean` | 取消確認ダイアログ表示フラグ |

## バリデーション・制約
- 「受付済」から「診療中」へのドラッグ&ドロップは禁止（トーストエラー: 「カルテ作成が必要です」）
- フィルター適用時、条件に合致しないカードは非表示（フロントエンドフィルタリング）

## データ型
`Appointment`, `ColumnData`, `DashboardColumnTitle`, `AppointmentVisitType`, `WeeklyChartPoint`, `WeeklyStatsResult`

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/dashboard/kanban` | 当日のカンバンデータ取得 | 未実装 |
| PUT | `/api/v1/dashboard/kanban/:appointmentId` | ステータス更新 | 未実装 |
| GET | `/api/v1/dashboard/stats` | 統計情報 | 未実装 |

## 実装状況
- フロントエンド(ui-sample): 実装済（モックデータ使用）
- バックエンドAPI: 未実装

## 備考
- スタッフ一覧は `useMasterItems("staff")` で取得し、`status === "active"` のみフィルター
- 旧仕様にあった `DashboardSummaryWidget`（統計ウィジェット）・`HospitalizationAlertWidget`（入院アラート）が追加されている
