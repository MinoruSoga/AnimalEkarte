# #254 / BRT-45 — 納品 UAT close チェックリスト

> **目的**: GitHub [#254](https://github.com/MinoruSoga/AnimalEkarte/issues/254) の 5 業務フローを、既存シナリオ（S / V / SECTION_14）へ対応付け、**不足だけ**をチェックリスト化する。
> **読者**: 検証実施者・PO。**実施レーン**は Linear [BRT-68](https://linear.app/baritechllc/issue/BRT-68)。本ファイルは **close 条件の棚卸し**であり、実施そのものではない。
> **棚卸し日**: 2026-08-20
> **Linear**: [BRT-45](https://linear.app/baritechllc/issue/BRT-45)

## 不変条件

1. **local PASS を #254 close にしない。** `LIFF_MOCK` の local 証跡は進捗であり、受け入れ条件の充足ではない。
2. **実 LINE / token 実測 / 別 sign-off は USER（BRT-68）。** 本ファイルでは実施しない。
3. **結果列は発明しない。** 手順があることと、通し結果が記録されていることは別である。作業ツリーに `reports/uat-*` は無い（gitignore / 削除済み）。GitHub コメント上の 2026-08-14 local FAIL0 は close 根拠にしない。
4. **検査（BRT-41 / #249）は別セッション。** S02 は対応表に載せるだけ。検査 feature は編集しない。
5. シナリオ md に PASS/FAIL を書かない（[TEST_ARCHITECTURE.md](../TEST_ARCHITECTURE.md) §4）。

---

## 1. #254 受け入れ条件（原文の分解）

| AC | 意味 | 現状（2026-08-20） |
|---|---|---|
| 全シナリオの通し結果（PASS/FAIL・証跡）が記録されている | 下記 5 フロー + 発見不具合の記録 | **未充足**。現行作業ツリーに通し証跡ディレクトリが無い |
| FAIL 項目がゼロ、または納品後対応として合意済みリストに隔離 | 合意リストの存在と承認 | **未充足**。合意済み隔離リストは repo に無い |

PO-06 コメント（#254, `opaque_ref=2026-08-08-PO-uat-TASK-023`）は overall=PARTIAL（f1/f2 PASS, f3 FAIL, f4 UNTESTED, f5 BLOCKED）。これは close でも現在の通し結果でもない。

---

## 2. 5 フロー × 既存シナリオ対応

| #254 フロー | 既存の手順正本 | 通し 1 本のシナリオか | 結果（本書では未記入） | 不足 |
|---|---|---|---|---|
| 1. 受付 → 診察 → 検査 → 会計 → 締め（AM/PM/EMG） | [SECTION_14](../SECTION_14_MANUAL_TEST_GUIDE.md) §2.1–2.2 / [S06](S06-record-lock-audit-trail.md) / [S09](S09-closing-time-boundaries.md) / [S02](S02-exam-abnormal-highlight-lock.md)（検査は参照のみ） / V01・V02 | **無い**（断片の接続） | **未記入** | 5 画面を 1 日サイクルとして記録する実施。検査実施は BRT-41。締め 3 区分は S09 手順あり・結果なし |
| 2. 予約 → 来院 → 再予約 | SECTION_14 §2.1 / V02（予約・受付） | **無い**（院内予約の専用 S は無い。S04 は LIFF） | **未記入** | 院内予約 → カンバン来院 → 再予約の通し記録 |
| 3. トリミング受付 → 実施 → 精算（診察併用） | [S11](S11-trimming-combined-accounting.md) | 手順あり | **未記入** | S11 の現行通し結果。歴史的コメントの f3=FAIL を消したことにしない |
| 4. LINE 予約 → カルテ反映 | [S04](S04-liff-reservation-journey.md) / [S12](S12-liff-pet-health.md) / V05 | 手順あり（**local は mock**） | **未記入** | **実 LINE**（STG。mock 禁止）。カルテ反映の実機証跡。本セッションでは実施しない |
| 5. 月次集計 → 帳票出力 | [S10](S10-customer-aggregation-consistency.md) / SECTION_14 §2.2 | 手順あり | **未記入** | S10 + `/accounting/reports` CSV の通し結果 |

---

## 3. close ゲート（不足チェックリスト）

実施レーンは **BRT-68**。本表の結果はすべて **未記入**（代行しない）。

| ID | ゲート | 手順の所在 | 誰 | 結果 |
|---|---|---|---|---|
| C-LINE | 実 LINE 予約 → 病院側反映 | S04（STG 節）・S12 | USER / BRT-68 / BRT-52 | **未記入**（本セッション未実施） |
| C-TOKEN | token health（LINE / LIFF / 関連 secret の健全性） | repo に専用シナリオ無し。値は書かない | USER | **未記入** |
| C-AUDIT | カルテ確定 Lock と `audit_logs` 目視 | S06 | BRT-68 | **未記入** |
| C-RESIDUAL | FAIL の納品後隔離リストと承認 | 未作成 | USER | **未記入** |
| C-SIGNOFF | 別 USER の close 承認 | #254 コメント欄 | USER | **未記入**（代行しない） |
| C-MATRIX | 5 フロー通し証跡（PASS/FAIL） | 本ファイル §2 + `reports/uat-YYYY-MM-DD/`（gitignore） | BRT-68 | **未記入** |

---

## 4. 既存シナリオ索引との関係

業務ギャップ用の S01–S13 とフォーム V01–V05 の索引は [README.md](README.md) が正本。本ファイルはそれを **#254 の 5 フローに再配置**しただけであり、S/V 本文の重複定義ではない。

#254 本文に無いが受入層にあるもの（close 条件の必須ではない）:

| ID | 役割 | #254 5 フローとの関係 |
|---|---|---|
| S01 | 死亡ペット誤操作ブロック | 臨床安全。5 フロー外 |
| S03 | ワクチン次回予定 | 臨床安全。5 フロー外 |
| S05 | 入院サイクル | 5 フロー外 |
| S07 / S08 | 見積・会計訂正 | フロー 1 の会計枝 |
| S13 | identity 手動訂正 | 5 フロー外 |
| V03 / V04 | 組織・マスタ項目単位 | 納品前フル受入（TEST_ARCHITECTURE §6）だが #254 本文の 5 フロー外 |

---

## 5. やらないこと（本チケット）

- 実 LINE 送信・実機 LIFF
- local / mock 結果を PASS として #254 を close
- 別 sign-off の代筆
- 検査 feature の編集（BRT-41）
- シナリオ md への結果転記
