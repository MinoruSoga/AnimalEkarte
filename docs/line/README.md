# LINE予約システム 統合ドキュメント

本ディレクトリには、Animal Ekarte の LINE 連携機能（LIFF アプリ、Messaging API 通知、管理画面統合）に関するすべての技術仕様が集約されています。

---

## 1. ドキュメント構成

- **[architecture.md](./architecture.md)**: システム全体図、認証フロー、時間枠計算エンジン、API エンドポイント一覧。
- **[reservation-spec.md](./reservation-spec.md)**: 詳細な機能要件、画面遷移フロー、UI 仕様、DB 統合スキーマ。
- **[setup.md](./setup.md)**: LINE 公式アカウント、Messaging API、LIFF ID の取得および設定手順。
- **[lstep-integration.md](./lstep-integration.md)**: Lステップ連携実装仕様（タグ同期・セグメントDM配信基盤）。
- **[lstep-confirmation-response.md](./lstep-confirmation-response.md)**: クライアント確認事項（docx section 7）への回答書。

---

## 2. システム概要

電子カルテ本体（Go/Gin WebAPI）を共通バックエンドとし、飼い主向けの **LIFF App** と、病院スタッフ向けの **管理画面** が連携して動作します。

### 主要な特徴
- **カルテ統合**: LINE 経由の予約は自動的に `appointments` テーブルに登録され、受付画面で即座に確認可能。
- **リアルタイム時間枠計算**: スタッフの勤務シフト（`/shifts`）と休憩時間、既存予約を考慮した空き枠計算。
- **通知自動化**: 予約完了・キャンセル時に LINE Push 通知とメール通知を自動送信。

---

## 3. クイックリンク

- **管理者設定**: `/line-reservation/settings`
- **ページ編集**: `/line-reservation/page-editor`
- **マスタ設定**: `/settings/reservation-type`, `/settings/staff`
