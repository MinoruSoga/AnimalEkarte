# FE-007: Dashboard カンバン ドラッグ&ドロップ時に API 呼び出しが行われない

## 概要

Dashboard の Kanban カンバンでカードをドラッグして別のカラムにドロップした際、UI上ではドラッグのビジュアルフィードバックが表示されるが、バックエンド API への状態更新リクエストが送信されない。

## 問題の詳細

### 実装ファイル
- `frontend/src/features/dashboard/routes/Dashboard.tsx`
- `frontend/src/features/dashboard/hooks/useDashboardKanban.ts`
- `frontend/src/features/dashboard/api/update-appointment-status.ts`

### 現象
1. 予約カード (例: 「14:00 猫 - ソラ 再診」) を「受付予約」カラムから「受付済」カラムへドラッグ
2. ドラッグのビジュアルフィードバック (ドラッグ中の表示) は正常に機能
3. **ドロップ後**: カード表示位置は変わらず、元のカラムに留まる
4. **ネットワークパネル確認**: PATCH リクエストが送信されていない

### 期待値
ドロップ後に以下の API リクエストが送信される:
```
PATCH /v1/reservations/:id/status
Body: { "status": "checked_in" }  // または該当する次のステータス
```

### 実際の動作
- 予約カード位置は変わらない（UI状態がロールバック）
- ネットワークにリクエスト痕跡なし
- ブラウザコンソールにエラー出力なし

## 影響範囲

- Dashboard カンバンボード全機能
- 予約ステータスの運用フロー

## 原因の推測

1. `useDashboardKanban.ts` で `onDragEnd` ハンドラが実装されていない、または条件不足
2. `update-appointment-status.ts` が正しくインポート・呼び出されていない
3. dnd-kit のドラッグ終了イベント処理が未実装

## テスト方法

1. Dashboard ページを開く（2026-03-16の予約データ表示）
2. 医師フィルターを「高橋 健一」に設定
3. 「受付予約」カラムに表示される予約カード (例: ソラ) をドラッグ
4. 「受付済」カラムにドロップ
5. ネットワークパネルで PATCH リクエストが送信されていることを確認
6. 予約カードが「受付済」カラムに移動

## 優先度

- **中** (UI は正常に見えるが、実際のステータス更新ができない。運用上、手動更新ぬけの原因になる可能性)

## 関連API

- `PATCH /v1/reservations/:id/status` — 予約ステータス更新エンドポイント

## 修正方針

1. `useDashboardKanban.ts` の `onDragEnd` ハンドラで `update-appointment-status()` を呼び出す
2. ドラッグ終了時にドロップ先カラムの ID から新ステータスを決定
3. API 呼び出し後、React Query キャッシュを無効化して最新データを再取得

## ✅ 修正完了（2026-03-16）

### 根本原因（最終特定）

dnd-kit の `closestCorners` 衝突検出戦略により、ドラッグ終了時に最も近いドロップ可能要素を検出した結果、ドラッグされたカード自身（overId="10"）がドロップターゲットとして選択され、ターゲットカラムの検出に失敗していた。

### 修正内容

1. **Dashboard.tsx**: 衝突検出戦略を `closestCorners` → `pointerWithin` に変更
   - `pointerWithin` はマウスポインタを含むドロップ可能要素を検出するため、カラムコンテナを正確に識別可能

2. **Dashboard.tsx の handleDragEnd/handleDragOver**: `over.data?.columnTitle` を優先的に使用
   - `useDroppable` で設定した `data: { columnTitle }` メタデータから直接ターゲットカラムを取得
   - ID マッチングより信頼性が高い

3. **useDashboardKanban.ts**: デバッグ用 console.log を削除

### コード変更

```typescript
// Dashboard.tsx の import
import { DndContext, PointerSensor, useSensor, useSensors, DragOverEvent, DragEndEvent, pointerWithin } from "@dnd-kit/core";

// DndContext の collisionDetection
<DndContext sensors={sensors} collisionDetection={pointerWithin} onDragOver={handleDragOver} onDragEnd={handleDragEnd}>

// handleDragEnd での column detection
const targetTitle = (over.data?.columnTitle as string) || (over.id as string).replace("column-", "");
```

### 検証結果（2026-03-16 ブラウザテスト実施）

✅ **受付予約 → 受付済**: カード移動 + PATCH /v1/reservations/10 {"status":"checked_in"} [200] 送信確認
✅ **受付済 → 診療中**: 制限が機能（カルテ作成が必要なため直接ドラッグ禁止）
✅ **データ永続化**: リロード後も受付済カラムに残存、バックエンド更新の永続性を確認
