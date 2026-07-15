# UI改善タスク: 受付ヘッダー テレメトリ表示エリア

> **完了（2026-07-15）**: Phase 1（本日受付件数）＋ Phase 2（`checked_in_at` 待ち時間統計）は実装済み。
> **アーカイブ日**: 2026-07-15（ルート `change-ui.md` から移動）
>
> **アクティブ仕様の正本**: [`docs/screens/01-reception.md`](../../../screens/01-reception.md) §3 受付テレメトリ
>
> **実装ポインタ**:
> - FE: `ReceptionTelemetryStrip` / `use-reception-telemetry.ts`（`RECEPTION_TELEMETRY_PHASE2_ENABLED = true`）/ `Reception.tsx`
> - BE: `reservation_service.go`（`checked_in` 遷移時スタンプ）/ migrations `checked_in_at`（`001_init.sql` 統合済み）
>
> **次期判断（未決）**: キルスイッチ `RECEPTION_TELEMETRY_PHASE2_ENABLED` の恒久 true 化後始末 → [`FE-refactor.md`](../../../../FE-refactor.md)
>
> 詳細な導入経緯・受け入れ基準は [`CHANGE-UI-reception-telemetry.md`](./CHANGE-UI-reception-telemetry.md) を参照。
