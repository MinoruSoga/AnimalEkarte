# AnimalEkarte 受入テスト バグ報告（再検証）— 残件

- **元レポート実施日**: 2026-08-09
- **追加検出**: 2026-08-10（BUG-031 / 032 / 033）· 2026-08-11（BUG-034 / 035 · 001/002/033 ブラウザ再NG）
- **実施環境**: http://localhost:3003（ローカル・seed 003_demo）
- **r2 クローズ日**: 2026-08-10
- **r3 クローズ日**: 2026-08-11（UAT 残バグ対応 · main 通常 merge+push）
- **main tip（本更新時点）**: `c8623130b`（`origin/main` 反映済み）

## 対応状況（結論）

| 区分 | 状態 |
|------|------|
| BUG-001〜030（本レポート採番） | **FIXED on main**（r1 + r2） |
| BUG-013 / BUG-018 | それぞれ BUG-006 / BUG-008 と同一根因として解消扱い |
| BUG-027 | 仕様判断（締め時間の境界逆転拒否は意図的 fail-closed。コード変更なし / **SPEC**） |
| BUG-031 / 032 / 033 | **FIXED on main**（r2 実装・検証・通常 merge+push）。033 の unit/FE は main 上 PASS |
| BUG-001 / 002 | **FIXED on main（r3 FE 直URLガード追加）**。BE 拒否は r1 時点で ALREADY。ブラウザ再NG は stale コンテナ + FE 表示漏れ |
| BUG-034 / 035 | **FIXED on main（r3）** |
| **S05 ブラウザ E2E** | **PASS（2026-08-11）** |
| **S06 ブラウザ E2E** | **PARTIAL** — コード上 034/035 は r3 で修正。**ローカル再ビルド後のブラウザ再確認は人間**。監査DB裏取りは未実施 |

### r3 新規実装（main 反映）

| BUG | 要約 | branch / tip | merge |
|-----|------|--------------|-------|
| 034 | 問診「治療方針」= inquiry.notes の GET 往復欠落を修正 | `fix/bug-034-treatment-policy-reload` @ `fec644128` | 64cada44a |
| 035 | 確定済カルテ: fieldset `display:contents` 撤去 + 保存非表示 | `fix/bug-035-finalized-mr-lock` @ `a3643a58b` | d96ea911c |
| 002 | `/medical-records/new` 死亡ペット hard-stop | `fix/bug-002-deceased-mr-new-block` @ `d9c9b193b` | 4d3c214de |
| 001 | `/accounting/new` 死亡ペット FE ブロック | `fix/bug-001-deceased-accounting-new-fe` @ `6370be3ea` | c8623130b |

### Phase 0 分類（r3 · 根拠）

| ID | 分類 | 根拠 |
|----|------|------|
| 001 | BE ALREADY + FE STILL_OPEN→FIXED + ENV_STALE | BE unit PASS; FE accounting に死亡ガード無し; 実行 image が 7月 |
| 002 | BE ALREADY + FE display STILL_OPEN→FIXED + ENV_STALE | BE unit PASS; new でフォーム表示が残存 |
| 033 | **ENV_STALE**（コード ALREADY） | main に exam lock + unit PASS。ブラウザ NG は古い frontend/backend イメージが主因 |
| 034 | STILL_OPEN→FIXED | InquirySummary が notes を落とす |
| 035 | STILL_OPEN→FIXED | fieldset + contents で disabled 非伝播 |
| 027 | SPEC | 触らない |

### 2026-08-11 r3 検証（kanban · Docker · migrate 未適用）

- ボード: `animalekarte-bugmd-202608-r3`
- 実装+検証 4 本すべて reviewer **APPROVE (PASS)**
- 新規 migration ファイル: **なし**（`make migrate` 不要）

## 削除済み（対応完了のため個別詳細を圧縮）

BUG-001〜035 の再現手順・当時エビデンスのフル本文は git 履歴（本ファイル旧版 / `8077f00f3` 以前）を参照。

## 人間アクション（残り）

1. **ローカル Docker 再ビルド / 再起動**（UAT が乗っている image が古い）:
   ```bash
   cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte
   docker compose build frontend backend
   docker compose up -d frontend backend
   ```
2. ブラウザ再確認: 死亡 petId=1000003 の `/medical-records/new` `/accounting/new` · 完了検査 1014563 · 確定カルテ 治療方針保持 · 確定後 input disabled
3. **staging へ main を取り込む**（人間のみ · 本キャンペーンでは未実施）
4. S06 監査ログ DB 裏取り（任意）
5. BUG-027 は追加実装不要（仕様のまま）
6. `make migrate` — 本 r3 では新 migration なし

**staging への merge は人間担当（本作業では実施しない）。**

以上。
