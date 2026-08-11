# AnimalEkarte 受入テスト バグ報告 — クローズ

- **元レポート実施日**: 2026-08-09
- **r2 クローズ**: 2026-08-10
- **r3 コード land**: 2026-08-11（main 通常 merge+push）
- **r3 環境反映 + ブラウザ再確認**: 2026-08-12
- **main tip（本更新）**: 下記 docs コミット時点の `origin/main`
- **実施環境**: http://localhost:3003（bind-mount main · `docker compose up -d --force-recreate frontend backend` 実施済み）

## 結論

| 区分 | 状態 |
|------|------|
| BUG-001〜026, 028〜032 | **FIXED on main** |
| BUG-001 / 002 | **FIXED** — r3 FE 直URLガード + 2026-08-12 ブラウザ PASS |
| BUG-027 | **SPEC**（実装不要） |
| BUG-033 | **FIXED** — コード r2 + 2026-08-12 ブラウザで完了検査の結果値/削除/追加が disabled（対象 ID 1014565。旧 1014563 は DB に無し） |
| BUG-034 / 035 | **FIXED** — r3 + 2026-08-12 ブラウザ/API 再確認 PASS |
| S05 ブラウザ E2E | **PASS**（2026-08-11） |
| S06 ブラウザ E2E | **PASS（ガード・ロック・notes 保持）** — 監査ログの DB 生 SQL 裏取りのみ任意未実施 |

**open / NG のコード欠陥は残っていない。**

## r3 実装（main）

| BUG | 要約 | tip | merge |
|-----|------|-----|-------|
| 034 | InquirySummary に notes 往復 | `fec644128` | 64cada44a |
| 035 | 確定済 fieldset lock（contents 撤去） | `a3643a58b` | d96ea911c |
| 002 | medical-records/new 死亡 hard-stop | `d9c9b193b` | 4d3c214de |
| 001 | accounting/new 死亡 FE ブロック | `6370be3ea` | c8623130b |

## 2026-08-12 再確認ログ（migrate 未適用）

**Unit（Docker `--entrypoint ''`）**
- BE billing/medicalrecord focused deceased + exam seal: **PASS**
- FE vitest 7 files / **145 passed** | 3 skipped

**環境**
- frontend/backend は host `main` を bind-mount。image rebuild は no-op（CACHED）のため **force-recreate** でプロセス更新。

**ブラウザ（八王子・一般スタッフ）**
| 項目 | 結果 |
|------|------|
| BUG-002 `/medical-records/new?petId=1000003` | 「死亡したペットは新規カルテを作成できません」表示 · 編集フォーム無し |
| BUG-001 `/accounting/new?petId=1000003` | 「死亡したペットは会計を作成できません」· 閲覧専用 · 確定操作無し |
| BUG-033 `/examinations/1014565`（status=completed） | 結果 textbox・項目削除・項目追加 disabled。保存ボタン無し |
| BUG-035 `/medical-records/1080036`（確定済） | バナー + 主訴/治療方針等 disabled。保存無し。追記のみ |
| BUG-034 | API GET `medical-records/1425546` の `inquiry.notes` = `UAT-r3 治療方針 keep`（確定後も保持）。画面 1080036 でも治療方針が `# 治療方針` に戻っていない |

## 人間側の任意・別レーン

1. **staging へ main を取り込む**（本レポート範囲外 · 人間のみ）
2. S06 監査ログの DB 生裏取り（任意）
3. BUG-027 は変更しない

以上。
