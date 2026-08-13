# AnimalEkarte 受入テスト バグ報告 — クローズ（r4）

- **r3 コード land**: 2026-08-11
- **r3 環境反映 + ブラウザ**: 2026-08-12
- **r4 コード land**: 2026-08-13（main 通常 merge+push）
- **main tip**: `48a17ae68`（origin/main）
- **migrate in r4 land**: **なし**

## 結論

| 区分 | 状態 |
|------|------|
| BUG-001〜026, 028〜032 | **FIXED on main**（r3 以前） |
| BUG-027 | **SPEC**（実装不要） |
| BUG-033 | **FIXED on main**（r4）— 完了検査シール EN/JA normalize harden。UAT 1014563 は soft-deleted pending（404）で residual 対象外 |
| BUG-034 | **FIXED on main**（r3） |
| BUG-035 | **FIXED on main**（r4）— 確定済カルテ臨床欄の explicit disabled + status OR ロック |
| BUG-036 | **FIXED on main**（r4）— 勤務シフトの開始/終了必須（FE+BE） |
| BUG-037 | **FIXED on main**（r4）— 入院ケージ必須（FE+BE create） |
| BUG-038 | **FIXED on main**（r4）— clinics `scope=all` を hospital-settings.view で一覧可 + FE エラー表示 |
| S04 予約枠・全日付 disabled | **DEFER / 環境・データ前提**（コード sprint 外） |
| S12 有効 LINE token / S13 複数医院リンク実操作 | **DEFER / 前提不足** |

**r4 対象の open コード欠陥は main 上 FIXED。** ブラウザ再確認は人間推奨（特に 033/035/038）。

## r4 実装（main）

| BUG | 要約 | tip | merge |
|-----|------|-----|-------|
| 033 | completed exam seal status normalize | `a433d3e43` | `ca9e706ce` |
| 035 | finalized MR clinical lock residual | `7595b0083` | `0cc6eb739` |
| 036 | shift required start/end | `86d09676b` | `5b22180d5` |
| 037 | hospitalization cage required | `56538fbf0` | `97ea499d8` |
| 038 | clinic master list scope=all | `393b49205` | `48a17ae68` |

Board: `animalekarte-bugmd-202608-r4` — impl+verify 10 done · APPROVE (PASS)×5 · SPEC/Phase0 blocked.

## 人間側の任意・別レーン

1. **staging へ main を取り込む**（本レポート範囲外 · 人間のみ）
2. デモユーザで 033/035/038 のブラウザ再確認（force-recreate 後推奨）
3. S04 枠データ整備後の予約 E2E
4. BUG-027 は変更しない
5. Linear Done / VERIFIED_FIXED は人間のみ

## ローカル再反映（任意）

```bash
cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte
git checkout main && git pull --ff-only
docker compose up -d --force-recreate frontend backend
# migrate は r4 差分なし — 不要
```

以上。
