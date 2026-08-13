# AnimalEkarte 受入バグ台帳（bug.md）

- **更新**: 2026-08-13
- **main tip**: 下記 docs コミット時点の `origin/main`
- **方針**: **未対応のコード欠陥のみ**を残す。対応済み BUG 行は削除（履歴は git）。

## Open（コード欠陥）

**なし。**

## クローズ済み（参照のみ・本文削除済）

| 範囲 | 結果 | 根拠 |
|------|------|------|
| BUG-001〜026, 028〜038 | FIXED on main | r3/r4 land · 2026-08-13 ブラウザ/API 再確認 |
| BUG-027 | SPEC（実装不要） | 製品判断。コード変更なし |
| S04 LIFF 予約（全日付 disabled） | **解消** | 原因はコードではなく **シフト未登録**。デモ用に staff_id=1 の 2026-08-15〜28 シフトを API 投入後、日付 14 日選択可・時間枠表示・予約確定 `R-20260815-0001`・キャンセル済まで通し |
| S12 mock ヘルスカード | **解消（mock）** | `/liff/health-card?clinic_id=1` で飼主「テストユーザー」表示。実 LINE token 連携は STG/本番の人間レーン（ローカル mock 対象外） |
| S13 identity-links | **解消** | `/identity-links` 到達・飼主/ペットリンク UI 表示（identity-links 権限ありユーザ）。複数医院の実リンク操作は任意の深い UAT |

## 2026-08-13 再確認ログ（要約）

- FE/BE: `docker compose up -d --force-recreate frontend backend` 実施
- BUG-033 `/examinations/1014565`: 完了ロック文言 + 結果テーブル入力 disabled
- BUG-035 `/medical-records/1080036`: 確定済バナー + 保存非表示 + textarea disabled
- BUG-038 `/settings/clinic` + `GET /clinics?scope=all`: 医院 4 件
- S04: 上記シフト投入後 E2E 通し（確定→キャンセル）
- 証跡: `reports/r4-closeout/`（未追跡）

## 人間レーン（bug.md 外・任意）

1. staging へ main 取り込み
2. 実 LINE token での S12/S04 通知受信（STG）
3. Linear Done / VERIFIED_FIXED は人間のみ
4. migrate: r4 差分なし

以上。
