# AnimalEkarte 受入テスト バグ報告（再検証）— 残件

- **元レポート実施日**: 2026-08-09
- **追加検出**: 2026-08-10（BUG-031 / 032 / 033）
- **実施環境**: http://localhost:3003（ローカル・seed 003_demo）
- **r2 クローズ日**: 2026-08-10
- **残タスク締め（自動再検証）**: 2026-08-11
- **main**: 本コミット時点の `origin/main`（docs 更新を含む）

## 対応状況（結論）

| 区分 | 状態 |
|------|------|
| BUG-001〜030（本レポート採番） | **FIXED on main**（r1 実装・検証・統合 + r2 再確認 PASS） |
| BUG-013 / BUG-018 | それぞれ BUG-006 / BUG-008 と同一根因として解消扱い |
| BUG-027 | 仕様判断（締め時間の境界逆転拒否は意図的 fail-closed。コード変更なし / SPEC） |
| BUG-031 / 032 / 033 | **FIXED on main**（r2 実装・検証・通常 merge+push） |
| **S05 / S06 ブラウザ E2E** | **人間向け残件（DEFER）** — 下記。コードの unit/回帰は 2026-08-11 に main 上で再 PASS |

### r2 新規実装（main 反映）

| BUG | 要約 | branch / tip | merge |
|-----|------|--------------|-------|
| 031 | 当日開始入院 → status admitted 既定 | `fix/bug-031-hosp-today-admitted` @ `817c9b448` | fe2031b89 |
| 032 | 入院 create の nested treatment_plans から care_plan_items を同一 TX seed | `fix/bug-032-hosp-care-plan-persist` @ `142675a16` | 474c1fb0a |
| 033 | 完了検査の結果編集/保存/削除ロック（FE+BE） | `fix/bug-033-exam-completed-lock` @ `5ed8a6b1b` | 5ef39c185 |

### 2026-08-11 main 上の自動再検証（Docker · migrate 未適用）

**BE** `go test ./internal/medicalrecord/`（`--entrypoint ''`）focused:

- PASS: `TestDefaultHospitalizationStatus`
- PASS: `TestHospitalizationService_Create`（BUG-031 today→admitted / future→reserved 含む）
- PASS: `TestHospitalizationService_Create_NestedPlansAtomicity`（BUG-032 seed 含む 6 subtests）
- PASS: `TestExaminationResultsLocked` / `TestExaminationLockErrorMessages`
- PASS: `TestExaminationService_UpdateUsesLockedExamStatus` / `ReplaceItemsUsesLockedExamStatus`
- PASS: `ReplaceItemsRejectsCompletedSeal` / `UpdateRejectsItemsOnCompletedSeal` / `DeleteRejectsCompletedSeal`
- PASS: `TestMedicalRecordService_Update_FinalizeAuditLog`（確定時 finalize 監査）

**FE** vitest 3 files / **78 tests PASS**:

- `examination-lock.test.ts`
- `use-examination-form.test.ts`
- `ExaminationFormFields.test.tsx`

※ DB フル migrate 依存の結合テスト（例: `vital_records` 欠落環境での FinalizeAuditFailureRollsBack）は本ラウンド対象外。`make migrate` はエージェント未実行。

Hermes ボード `animalekarte-bugmd-202608-r2`: 実装可能カード done / blocked = DEFER ブラウザ再実施 + SPEC 027。  
**staging への merge は人間担当（本キャンペーンでは実施しない）。**

---

## 残件: S05（入院サイクル）— ブラウザ E2E DEFER

- **コード**: BUG-031/032 は main 反映済み。上記 unit 再 PASS
- **未実施**: 入院管理のブラウザ E2E（一覧/詳細/退院/請求連携の通し）
- **扱い**: ブラウザ再実施まで open（コード FIXED と混同しない）

## 残件: S06（カルテロック・監査証跡）— ブラウザ/API 裏取り DEFER

- **コード**: 確定 finalize 監査 unit PASS。検査 completed ロック（BUG-033）unit+FE PASS
- **未実施**: 実カルテでの確定(Lock)・訂正追記・削除拒否・監査証跡の **ブラウザ/API 応答裏取り**
- **特に未確認（人間）**: 過去報告「カルテ確定時に身体検査所見が空・診断/方針が固定文で上書き」の **本番相当データでの再発有無**
- **扱い**: 「バグなし」ではなく **E2E 未検証**。再実施まで open

### 環境障害メモ（S04〜S06 共通 · 元レポート）

複数テスターのブラウザタブ競合、およびブラウザ操作ツール障害によりアクション系検証が不能だった記録あり。人間による S05/S06 E2E 再実施を推奨。

S04（LIFF 予約ジャーニー）手順 2〜12 も元レポートでは未実施。コード open bug としては起票しない。

---

## 削除済み（対応完了のため個別詳細を圧縮）

BUG-001〜033 の再現手順・当時エビデンスのフル本文は git 履歴（本ファイル旧版 / `29abc8963` Find bug #298）を参照。

## 人間アクション（残り）

1. 必要なら **deploy 確認**（main）
2. **staging へ main を取り込む**（人間のみ · 本作業では未実施）
3. S05 / S06 の **ブラウザ E2E 再実施**
4. BUG-027 は追加実装不要（仕様のまま）

以上。
