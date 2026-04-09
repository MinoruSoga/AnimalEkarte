# Phase 7: 結合・デプロイ

> **状態**: ✅ TASK-RES-071（結合テスト）完了 / TASK-RES-072（デプロイ）PR作成済み・LINE Console設定は手動
>
> TASK-RES-069（CORS）・TASK-RES-070（Docker Compose）は実装済み。

## TASK-RES-069: CORS設定 ✅

**実装済みファイル**: `backend/internal/middleware/cors.go`

**設定内容**:
```
許可オリジン:
  - https://reserve.noah-karte.com（本番）
  - http://localhost:3001（開発）
  - https://liff.line.me（LIFF内ブラウザ）
```

**完了条件**:
- [x] LIFF AppからAPIリクエストが成功（設定済み）
- [x] 不正なオリジンからのリクエストが拒否される

---

## TASK-RES-070: Docker Compose更新 ✅

**実装済みファイル**:
- `docker-compose.yml`（liff-app コンテナ追加済み）
- `liff-app/Dockerfile`（作成済み）

**完了条件**:
- [x] `docker compose up` で backend, frontend, liff-app, db 全起動

---

## TASK-RES-071: 結合テスト ✅

**テスト実施日**: 2026-04-09

- [x] 管理者: 基本設定ページ表示（Running/Stopped）
- [x] 顧客: LIFF App 表示（`http://localhost:3001/3`、Stopped → メンテナンス画面）
- [x] 管理者: 予約管理カレンダー（週表示）
- [x] 管理者: 手動予約入力（2段階モーダル）
- [x] 管理者: 予約顧客管理ページ
- [ ] 顧客: LIFF App → 予約フロー全8ステップ → 予約確定（LINE IDトークン必要・staging環境待ち）
- [ ] LINE Push通知受信確認（LINE Messaging API 設定後）
- [ ] 管理者専用コース（is_internal）非表示確認（Running状態でのみ確認可能）
- [ ] 指名なし委譲の動作確認（Running状態でのみ確認可能）

---

## TASK-RES-072: デプロイ 🔲

**PR**: MinoruSoga/AnimalEkarte#27（main → staging）

**手順**:
1. [x] PR #27 作成済み（2026-04-09）
2. [ ] PR レビュー・マージ
3. [ ] DBマイグレーション実行（**db_reset=true** 必須 — 001_init.sql に LINE予約テーブル統合済み）
4. [ ] バックエンド・管理画面フロントエンドは CI/CD で自動デプロイ
5. [ ] LIFF Appデプロイ（reserve.noah-karte.com）— 別途 CI/CD 設定または手動
6. [ ] LINE Developers Console: LIFF URL を `https://reserve.noah-karte.com/{clinicId}` に変更（手動）
7. [ ] リッチメニュー作成: 「予約する / 予約確認 / 電話する」（手動）

**完了条件**:
- [ ] staging環境で全フロー動作確認
- [ ] LINE公式アカウントからLIFF Appにアクセス可能
- [ ] リッチメニューが表示される
