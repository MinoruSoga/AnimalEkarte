# BE: BUG-086 シフト時刻逆転で 500 → 400 に修正

## 概要

シフト作成・更新で `start_time > end_time`（開始時刻 > 終了時刻）の場合に
バックエンドが 500 Internal Server Error を返す。
フロントエンドのバリデーションは修正済みだが、API 直接呼び出しには対応できていない。

## 期待する動作

- `start_time >= end_time` の場合は 400 Bad Request を返す
- エラーメッセージ: `"start_time must be before end_time"`

## 実装場所

- `backend/internal/service/shift_service.go` または `handler/shift_handler.go`
- バリデーション追加

## 優先度

Medium

## 関連

- `docs/tasks/closed/BUG-086_shift_time_reverse_500.md`（FE部分は修正済み）
- FUNCTIONAL_TEST_REPORT.md line 8832
