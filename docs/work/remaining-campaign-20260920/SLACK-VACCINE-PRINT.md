# SLACK-VACCINE-PRINT: 一般カルテ印刷 vs ワクチン証明書（renderer 欠落）

状態: **経路分離 READY／専用証明書 renderer 未確認／製品実装・PDF 未実行（停止）**。出典は [todo-issue.md](../../../todo-issue.md) 見出し `### SLACK-VACCINE-PRINT`（L203–206、索引 L426）。保持する現場条件:

- 出典 956–960 の紙のワクチン証明書の質問。専用証明書の対応範囲確認が必要
- [カルテ print hook](../../../frontend/src/features/medical-records/hooks/use-medical-record-form-modals.ts) と [診療カルテ印刷](../../../frontend/src/features/medical-records/components/MedicalRecordPrintView.tsx) は **一般カルテ用**
- 調査範囲では専用証明書 route / renderer を確認できていないため **「既に印刷可」と回答しない**
- [接種フォーム仕様](../../../docs/spec/screens/15-vaccinations-form.md) と保存項目を、承認済みの証明書見本/必要項目へ対応づける
- 公的文書要件や未収録の見本を推測しない。専用仕様が決まるまで実装は停止
- **一般カルテ印刷と証明書を同一視しない**

本ファイルは製品コードから import されない。キャンペーン unit `SLACK-VACCINE-PRINT`（`remaining-ops-20260920` revision 1）の owned path および人間が読むギャップ票である。照合 revision `873685b0bea3692c2f8100b19ded660c8357f2b0`（`feat/rem-slack-vaccine-print-20260920`）。製品コード・テスト・PDF は変更しない。

呼び出し行: **無い。** 既存 `docs/work/todo-campaign-20260919-ready17/` に本 unit の票は無く、[SLACK-INTAKE](../todo-campaign-20260919-ready17/SLACK-INTAKE.md) L80 は分類ポインタのみ。`todo-issue.md` L203–206 は出典要約であり本票の代替ではない。ledger `owned_paths` が本パス単体のため、他シートへの追記では unit 完了にならない。

見本（用紙・レイアウト・記載欄）は repo に無い。**specimen UNKNOWN**。本票は見本を描かない。

## 実践ゲート（実装しない理由）

[product-philosophy.md](../../product-philosophy.md) の 5 ステップをこの要望に当てると、①要件を疑う段階で止まる。

1. **要件を疑う:** 「紙のワクチン証明書が欲しい」は画面要望であり、用途・種類・発行者・用紙が未裁定。責任者の個人名が無い。一般カルテの「印刷」ボタンがあることを証明書の成立根拠にしない。
2. **削除:** 接種記録の保存と証明書発行は別工程。カルテ印刷で代用できるかは PO。本票で代用を採用しない。
3. **簡素化 / サイクル短縮 / 自動化:** 見本未承認のまま PDF を足す最適化は禁止。自動化は手動検証済みプロセスのみ。

公的文書のレイアウトを invent するのは安全境界。本票はギャップの並置まで。

## 医院事実（コード外・UNKNOWN）

数値・院内ルール・公的様式をコードから捏造しない。未採取なら該当セルは **UNKNOWN**。

| 項目 | コード/文書で分かること | 医院事実 |
| --- | --- | --- |
| 報告医院・端末・ブラウザ | なし | **UNKNOWN** |
| 紙の見本（specimen） | repo に用紙スキャン・寸法・欄配置は無い | **UNKNOWN**。未収録見本を推測しない |
| 証明書の用途 | 院内控え / 飼主交付 / 行政・宿泊・渡航 の区別なし | **UNKNOWN**。todo-issue 完了条件どおり PO |
| 種類（狂犬病 / 混合 / その他） | 接種行は `vaccine_id` 1 件。証明書種別 enum は無い | **UNKNOWN**。マスタ名から公的種別を決めない |
| 記載内容 | 接種フォームは日付・ワクチン・LOT・次回予定・備考 | **UNKNOWN**。必要項目は承認済み見本との対応づけ |
| 発行者 / 発行日 | カルテ印刷は担当医名と診療日。証明書の発行者フィールドは無い | **UNKNOWN** |
| 用紙 / レイアウト | カルテ印刷は A4 portrait / `window.print()` | **UNKNOWN**。専用用紙を invent しない |
| フロント/API revision | 本票作成時 worktree HEAD `873685b0bea3692c2f8100b19ded660c8357f2b0` | 再現セッションの SHA は **UNKNOWN** |

## 混ぜてはいけない読み替え

現場の「ワクチン印刷」「証明書」は次のどれでも同じ言葉になる。一般カルテ印刷・接種記録・共有ファイル・会計印刷を混ぜない。

| ID | 何か | いまのコード | 混ぜてはいけない読み替え |
| --- | --- | --- | --- |
| C — 一般カルテ印刷 | `MedicalRecordPrintView` 見出し「診療カルテ」。主訴・所見・診断・方針・選択済み処方薬。`window.print()` | 「印刷ボタンがある = 証明書が出せる」と回答しない |
| V — 接種記録 | `/vaccinations` とカルテ「予防接種」タブ。保存は `POST/PATCH /api/v1/vaccinations` | 記録できることを証明書発行と同一視しない |
| F — 接種フォーム仕様 | [15-vaccinations-form.md](../../../docs/spec/screens/15-vaccinations-form.md)。LOT・次回予定。印刷/証明書の節は無い | 仕様書の保存項目を証明書欄と決めない |
| U — 共有ファイル `vaccine_cert` | LSTEP 共有ファイルの purpose 定数。アップロード用途 | 既存 PDF をアップロードできることを **生成 renderer** と同一視しない |
| A — 会計 / 月次印刷 | 会計明細・月次レポートの `window.print()` / 「PDFとして保存」 | 領収書印刷をワクチン証明書の実装と読まない |
| M — 同日複数接種 | [SLACK-VACCINE-MULTI](../todo-campaign-20260919-ready17/SLACK-VACCINE-MULTI.md) | 複数本の保存問題と証明書用紙を混ぜない |
| S — 種別マスタ | [UAT-Q2-VACCINE-SPECIES](../todo-campaign-20260918/UAT-Q2-VACCINE-SPECIES.md) | 種の候補絞り込みを証明書種別としない |

## 現行経路 C — 一般カルテ印刷（証明書ではない）

入口は保存済みカルテのフローティング「印刷」だけ。新規カルテでは出さない。予防接種タブでも **同じカルテ印刷**が残る。

1. **ボタン:** [MedicalRecordFormActions.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordFormActions.tsx) L87–96。`!isNewRecord` のとき「印刷」。タブ条件なし。
2. **予防接種タブ:** 同ファイル L114–116 とテスト L156–161。「予防接種タブでは保存を出さず確定する・印刷は残す」。残るのはカルテ印刷であり証明書ではない。
3. **配線:** [MedicalRecordFormReadyPanels.tsx](../../../frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) L361 `onPrintClick={ready.modals.handlePrintClick}`。
4. **hook:** [use-medical-record-form-modals.ts](../../../frontend/src/features/medical-records/hooks/use-medical-record-form-modals.ts) L30–35。`setIsPrinting(true)` → 100ms 後 `window.print()` → `setIsPrinting(false)`。PDF ライブラリは呼ばない。
5. **マウント:** 同 [MedicalRecordFormActions.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordFormActions.tsx) `MedicalRecordPrintArea` L221–244。`isPrinting && !isNewRecord && pet` のときだけ。`hidden print:block`。`@page { size: A4 portrait; margin: 15mm; }`。
6. **見出し:** [MedicalRecordPrintView.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordPrintView.tsx) L38–40 コメント「カルテ印刷レイアウト」。L66 `<h2>診療カルテ</h2>`。L67 診療日。L68 カルテ番号。
7. **載るもの:** 医院名/住所/TEL、ペット名/種別/飼主名、担当医、主訴、身体検査所見、診断、治療方針、`item_type === "medicine" && is_selected` の処方薬（L54, L139–163）。
8. **載らないもの:** ワクチン名、接種日、LOT1–4、次回予定日、`vaccinations.remarks` / `supplemental`、証明書番号、発行者印、公的様式。props に接種レコードは無い（L22–33）。

この経路の成功は「診療カルテがブラウザ印刷できる」まで。証明書 preview → PDF → 実印刷の受入条件を満たさない。

## 現行経路 V — 接種記録（印刷 UI なし）

独立画面とカルテ埋め込みは保存用。証明書 renderer ではない。

### 仕様（15-vaccinations-form.md）

[15-vaccinations-form.md](../../../docs/spec/screens/15-vaccinations-form.md) SHA-256 `eae64e78170196890d724ae600c18da65ec94f1f6fbde12688fbf63ac722ea08`。

| 節 | 内容 | 証明書との関係 |
| --- | --- | --- |
| 概要 | ロット番号の編集と次回予定日。URL `/vaccinations/new?petId=` / `/vaccinations/:id`。カルテタブは別実装 | 印刷 URL は無い |
| 1.1 | 接種日必須、ワクチンマスタ必須、補助説明、備考 | 保存項目。証明書欄ではない |
| 1.2 | LOT 最大 4 | トレーサビリティ記録。用紙上の印字位置は UNKNOWN |
| 1.3 | 次回予定日（3週/4週/1年/手動） | リマインド用。証明書の「有効期限」と決めない |
| 1.4 | 過去履歴の右カラム | 画面一覧。証明書履歴面ではない |
| 1.5 | 削除 | 発行取り消しではない |
| API | GET/POST/PATCH/DELETE `/api/v1/vaccinations` | print / PDF / certificate エンドポイントは無い |

`rg print` / `rg 証明書` をこの仕様ファイルに対して実行した結果、一致なし。

### フロント（vaccinations feature）

[paths.ts](../../../frontend/src/config/paths.ts) L148–155: `/vaccinations`、`/vaccinations/select-pet`、`/vaccinations/new`、`/vaccinations/:id`。print path は無い。

[VaccinationForm.tsx](../../../frontend/src/features/vaccinations/routes/VaccinationForm.tsx) は保存・削除・履歴。印刷ハンドラなし。

[VaccinationFormPagePanels.tsx](../../../frontend/src/features/vaccinations/routes/VaccinationFormPagePanels.tsx) のフォーム状態は `date` / `vaccineId` / `supplemental` / `lot1–4` / `nextScheduleType` / `nextDate` / `remarks`（L80–92）。印刷ボタンなし。

`frontend/src/features/vaccinations/**` に対する `rg 'print|証明書|window\.print'` は一致なし。

### 保存項目（対応づけ用・採用しない）

承認済み見本が無いので、下表は **記録側に何があるか** だけ。証明書の必須欄にはしない。

| 記録フィールド | 出典 | 証明書への写し |
| --- | --- | --- |
| `date` | 仕様 1.1、[vaccination_record.go](../../../backend/internal/model/vaccination_record.go) L24 | UNKNOWN（接種日として使うかは見本） |
| `vaccine_id` → マスタ名 | 仕様 1.1、モデル L23 | UNKNOWN。公的ワクチン名との対応は PO |
| `lot1`–`lot4` | 仕様 1.2、モデル L29–32 | UNKNOWN |
| `next_date` / `next_schedule_type` | 仕様 1.3、モデル L26–27 | UNKNOWN。有効期限と同一視しない |
| `supplemental` / `remarks` | 仕様 1.1、モデル L28 / L33 | UNKNOWN。証明書備考に出さない |
| `doctor_id` | モデル L25 | UNKNOWN。発行者かは PO |
| `pet_id` / `medical_record_id` | モデル L21–22 | 対象動物の紐付け。用紙レイアウトではない |
| 医院名・住所 | カルテ印刷の clinic props。接種 API の応答ではない | UNKNOWN |

[buildCreateVaccinationRequest](../../../frontend/src/features/vaccinations/hooks/use-vaccination-form-model.ts) L138–155 は上記の保存 body。print payload は無い。

## 欠落 — 専用証明書 renderer

本 worktree HEAD `873685b0bea3692c2f8100b19ded660c8357f2b0` の調査範囲:

| 探したもの | 結果 |
| --- | --- |
| `VaccinationPrintView` / certificate コンポーネント | 無し（`MedicalRecordPrintView` のみカルテ） |
| `/vaccinations/.../print` route | 無し（paths.ts） |
| 接種フォームの印刷ボタン | 無し |
| 仕様の証明書節 | 15-vaccinations-form.md に無し |
| 生成 PDF API | 接種 CRUD のみ |
| 紙見本・寸法 | repo に無し → **specimen UNKNOWN** |

近傍で混同しやすいが **renderer ではない** もの:

- `SharedFilePurposeVaccineCert = "vaccine_cert"`（[shared_file.go](../../../backend/internal/model/shared_file.go) L37、[models.ts](../../../frontend/src/types/generated/models.ts) L3331）。LSTEP 共有ファイルの purpose。JPEG/PNG/PDF の **アップロード**（同ファイル L41–48）。生成しない。
- 会計の「印刷 / PDF出力」は月次・明細。ワクチン証明書ではない。

## 停止条件

todo-issue L206 を本票の完了ゲートとして再掲する。本票は調査まで。次を満たすまで製品コード・PDF を足さない。

1. 臨床 PO が証明書の **用途・種類・記載内容・発行者/日付・用紙/レイアウト** を承認する
2. 承認済み見本（specimen）が特定される。未収録見本を invent しない
3. 正しい接種記録だけから **preview → PDF → 実印刷** できる受入条件が書ける
4. 一般カルテ印刷（経路 C）を証明書の代替にしない、という判定が残る

PO への限定質問は既存本文どおり（INTAKE: 用途/記載/用紙）。本票で新しい質問を増やさない。公的文書レイアウトの invention が必要になった時点で停止。

## 検証メモ（本票作成時）

- Binding SHA-256: `todo-issue.md` `395c2588da867276a0b0c503f6ebcae792b9afad070947ebaf4dcbf949f2cb0b`（読み取りのみ。編集しない）
- `MedicalRecordPrintView.tsx` `bfbfe3269283fe7536071de8b382ee50f08618111d361855d7c624ec10812f7f`
- `15-vaccinations-form.md` `eae64e78170196890d724ae600c18da65ec94f1f6fbde12688fbf63ac722ea08`
- 本票は docs-only。Docker アプリテストは走らせない
