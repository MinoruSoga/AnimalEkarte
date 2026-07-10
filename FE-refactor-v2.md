# FE-refactor-v2.md — frontend/ リファクタリング完了記録

- **作成日**: 2026-07-10 / **完了日**: 2026-07-11
- **ステータス**: **全項目 CLOSED（未対応 0 件）**
- **対象**: `frontend/`（main / liff / line-reserve）— knip 棚卸し 233 件解消 + エピック後発見の構造負債是正 + a11y 残債 + knip gating 化

旧 `FE-refactor.md`（FD1〜FD12 / R-F1〜R-F25、2026-07-10 CLOSED, commit `2a1ef3ad`）が「別チケット」として先送りした残債（knip 棚卸し・a11y 残差・R-F19 未分割ファイル）と、エピック完了後の全体掃き出しで新規検出した負債（レイヤ逆転 import・定数複製・死にシム）を対象に、全 17 項目（FE-R1〜FE-R17）を実行し完了した。挙動保存を原則とし、挙動変更を要する項目は着手せず §別トラックへ記録した。

---

## 完了サマリー

| FE-R | 内容 | commit |
|------|------|--------|
| R1 | knip 誤検知の恒久除去（knip.json ignore 設定） | `c8962b49` |
| R2 | 死にファイル3件 + UnifiedTabs 重複 export 削除 | `d2e25916`, `476e7591` |
| R3 | 検証済み死にシンボルの削除（DELETE-SYMBOL バッチ） | `4d1ac654` |
| R4 | 未使用 barrel re-export 行の除去（約55行） | `338d6bbd` |
| R5 | 未使用 export キーワード除去 — features/*/api 生 fetcher 群（約90件） | `e38bfadd` |
| R6 | 未使用 export キーワード除去 — components/constants/hooks/types/liff/line-reserve（約35件） | `a1b302e3` |
| R7 | cascade 第2波の回収（knip 再実行→新規 unused files 削除） | `e4afac37` |
| R8 | 未使用依存 jsonwebtoken の削除 | `f2ab71cf` |
| R9 | accounting のレイヤ逆転 import 解消 | `9672aac3`, `36f57366` |
| R10 | aggregation のレイヤ逆転 import 解消（AggregationTab 型移動） | `93bb3275` |
| R11 | medical-records の lazy レジストリを components/ へ移動 | `78eea810` |
| R12 | PaymentCard の PAYMENT_METHOD_LABELS 複製を共有定数 import に置換 | `fd355b92` |
| R13 | liff/line-reserve の use-liff 互換シムを inline 化して削除 | `4307d822` |
| R14 | PetEditModalFieldSections.tsx（469行）の3セクション分割 | `52bd3781` |
| R15 | フォーム input ラベル付与 — settings/lstep | `cb9dcb68` |
| R16 | フォーム input ラベル付与 — 残り全 feature | `247e243b` |
| R17 | knip の gating 化（`--no-exit-code` / `continue-on-error` 除去） | `5568a0f0` |

knip 棚卸し（R0-2 ベースライン 2026-07-10 実測: Unused files 3 / dependencies 1 / devDependencies 3 / exports 147 / exported types 78 / duplicate exports 1）は R7 完了時点で **Unused files 0 / exports 0 / exported types 0 / duplicate exports 0** に到達し、R8 で dependencies も 0 化。R17 で CI の knip ステップを non-gating から gating へ切り替え、再蓄積を防止した。

R17 の申し送り: **push 後に frontend job の Knip ステップが green になることを確認**（本計画のローカル実行では push 済み CI 結果を検証していない）。

---

## 別トラック（本計画では実行しなかった・記録のみ）

| 項目 | 内容 | 実行しない理由 |
|---|---|---|
| useMasterSave validationError fix | `toast.error(error)` 追加（WIP 3 ファイル: use-master-save.ts / use-master-save.test.ts / StaffSettings.test.tsx） | 別セッションが作業中の挙動変更。本計画では触っていない |
| AccountingListTable の支払方法表記 | `credit_card: "クレジットカード"`（AccountingListTable.tsx:25）vs 他画面 "カード" の表記分岐 | どちらが正かは PO 判断のユーザー可視変更。統一する場合は FE-R12 の共有定数（daily-accounting-utils.ts）に寄せる |
| `{cond && <JSX>}` の機械強制 | `eslint-plugin-react`（jsx-no-leaked-render）等の導入 | devDependency 追加はオーナー判断。現状違反0件のため導入即 green |
| R-F17 統合フェーズ（calcAgeAt 共通化） | pets の年齢計算統合 | 挙動変更を伴う（旧 FE-refactor.md で `fix:` 分離確定済み） |
| dangerouslySetInnerHTML / PrintPortal XSS 監査 | セキュリティレビュー | 別途レビュー枠 |
| React Query の clinic_id キャッシュ境界 | クリニック切替時のキャッシュ分離検証 | 調査タスク（結果次第で挙動変更） |
| FE zod ↔ Backend バリデーション二重管理 | 設計判断 | architect 判断が必要 |
| OwnerInfoFieldSections.tsx (441行) / MedicalRecordForm.tsx (411行) の再分割 | R-F19 分割後も400行超の残差 | 800行ハード上限内で優先度 LOW。JSX 移動リスクが効果に見合わない |
| shadow rgba 1件（ShiftTemplateSettingsParts.tsx:225） | design-audit の C1/C3/C5 対象外の cosmetic nit | 既存 STYLE 定数と rounded 指定が異なり単純置換不可 |
| `src/types/` 手書き型 vs `types/generated/` 重複棚卸し | FE-R6 で個別シンボルは処理済みだが全量監査は未実施 | tygo 生成側の設計に関わるため別途 |

---

## 教訓

- FE-R4 の barrel 行削除で feature 外からの deep import（Feature Indexing 違反）が見つかった場合は import を書き換えず file:line を報告に留める運用にした（本件では違反0件）。
- knip の「同ファイル内参照あり→export キーワードのみ除去 / 参照なし→grep 検証の上シンボル削除」ルールで機械的に処理し、型チェック（`docker compose exec frontend pnpm run type-check`）を正本ガードとして各項目のコミット前に必ず通した。
