# INVESTIGATION: Lステップ連携＆DM機能の調査と実装方針確定

**作成日**: 2026-04-16  
**Status**: CLOSED（2026-04-20 クライアント要件書により調査完了）  
**Priority**: **HIGH**  
**目的**: 機能定義 No.22 / No.23 にある「Lステップとの連携＆DM機能」について、現行 `EkarteSprint` の対象範囲を確認し、実装要否・実装時期・実装方式を確定する。  
**前提**: このタスクでは実装は行わない。まず調査と方針整理のみを実施する。

---

## 背景

現行の LINE予約実装は、予約フォーム、予約状況連携、LIFF フローまでは整理されているが、Lステップ連携や公式LINEでのDM送信に関する実装方針がタスク上で明示されていない。

当院では Lステップを日常運用で利用しているため、カルテ登録情報と Lステップのタグ紐づけ、対象を絞った DM 配信は運用上重要な機能である。

---

## 確認したい論点

1. **No.22 の実装範囲確認**
   - 「LINE予約の実装完了」が以下のどこまでを指すかを確定する
   - ① 予約システムの連携（フォーム → カルテ自動入力）
   - ② 予約状況の連携（予約 on LINE との連携）
   - ③ Lステップと連携＆DM機能

2. **No.23 の扱い確認**
   - Lステップ連携＆DM機能が EkarteSprint のスコープ外なのか
   - スコープ外であれば、別タスク化するのか、次フェーズへ送るのか
   - スコープ内であれば、実装時期と依存関係を明確化する

3. **連携方式の確認**
   - カルテ側の患者情報を Lステップのどの情報に紐づけるか
   - タグ付与 / 剥奪 / セグメント配信 / 手動配信のどこまで対応するか
   - 公式LINE API、Lステップ API、既存の LINE連携機構のどこを経由するか

4. **運用要件の確認**
   - どのイベントをトリガーに DM 配信するか
   - 既存飼い主 / 新規飼い主 / 予約完了 / 予約変更 / 来院後 などの分類
   - 既存の予約・カルテ・顧客データとの整合性要件

---

## 調査対象

- `docs/line/reservation-spec.md`
- `docs/tasks/closed/reservation/00-OVERVIEW.md`
- `docs/tasks/closed/reservation/03-PHASE3-LIFF-API.md`
- `docs/tasks/closed/reservation/05-PHASE5-LIFF-APP.md`
- `docs/tasks/closed/reservation/06-PHASE6-LINE.md`
- `backend/internal/handler/liff_*`
- `backend/internal/service/liff_*`
- `backend/internal/service/appointment_notification_service.go`
- `frontend/line-reserve/src/`
- 既存の Lステップ / 公式LINE 運用資料

---

## 期待する成果

1. No.22 / No.23 の解釈を文章で確定する
2. Lステップ連携の要否と優先度を整理する
3. 実装する場合の方式案を 1 つ以上に絞る
4. 実装しない場合は、スコープ外として明文化する
5. その結果をもとに、別途実装タスクを起票できる状態にする

---

## 関連チケット

- No.22: 「LINE予約の実装完了」の範囲確認
- No.23: Lステップとの連携＆DM機能
- 既存の LINE予約実装タスク群

---

## 備考

- このタスクではコード修正を行わない
- 調査完了後に、必要であれば実装タスクへ分割する

---

## 調査完了（2026-04-20）

クライアント（ノア動物病院）から要件書 `lstep_karte_spec.docx` が提供され、上記の確認論点がすべて回答された。

**成果物**: `docs/line/lstep-integration.md`（全仕様を統合済み）

**主な確定事項:**
- 連携方式: Ekarte → Lステップ API（一方向同期、イベント駆動）
- スコープ: タグ連携（最優先）+ CPMステージ管理 + ②スタッフ個別送信
- No.23 スコープ: 実装対象（Phase 1〜5 で段階実装）
- 確認事項: `docs/line/lstep-integration.md` の section 12 に社内運用の追加確認事項、section 13 に要望書 section 7 の Q&A を統合済み
