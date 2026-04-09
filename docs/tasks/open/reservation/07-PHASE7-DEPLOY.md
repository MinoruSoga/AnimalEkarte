# Phase 7: 結合・デプロイ

> **状態**: 🔲 TASK-RES-071（結合テスト）・TASK-RES-072（デプロイ）が未実施
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

## TASK-RES-071: 結合テスト 🔲

**テストシナリオ（未実施）**:
- [ ] 管理者: コース・スタッフ登録 → 基本設定 → Running
- [ ] 顧客: LIFF App → 予約フロー全8ステップ → 予約確定
- [ ] LINE Push通知受信確認
- [ ] 管理者: 予約状況で確認 → キャンセル
- [ ] 顧客: マイ予約でキャンセル確認
- [ ] 管理者: 手動予約入力
- [ ] 管理者: 個人設定変更 → 空き時間への反映確認
- [ ] 同日予約制限の動作確認
- [ ] Stopped状態でメンテナンス画面表示確認
- [ ] 管理者専用コース（is_internal）が予約アプリに非表示
- [ ] 指名なし委譲の動作確認

---

## TASK-RES-072: デプロイ 🔲

**手順（未実施）**:
1. [ ] DBマイグレーション実行（003_line_reservation.sql + 003_line_reservation_seed.sql）
2. [ ] バックエンドデプロイ
3. [ ] フロントエンド（管理画面）デプロイ
4. [ ] LIFF Appデプロイ（reserve.noah-karte.com）
5. [ ] LINE Developers ConsoleでLIFF URLをデプロイ先に変更
6. [ ] リッチメニュー新規作成（予約する / 予約確認 / 電話する）

**完了条件**:
- [ ] staging環境で全フロー動作確認
- [ ] LINE公式アカウントからLIFF Appにアクセス可能
- [ ] リッチメニューが表示される
