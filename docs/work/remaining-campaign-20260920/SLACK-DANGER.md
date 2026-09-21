# SLACK-DANGER: 既存「高」危険表示 vs 要望の赤/黄（色意味は PO）

状態: **既存-versus-要望の調査票 READY／製品実装・色意味裁定 未実行**。出典は [todo-issue.md](../../../todo-issue.md) 見出し `### SLACK-DANGER`（L190–194、索引 L423）。保持する現場条件:

- 危険度に応じ **赤/黄で分かる表示** を求める（出典 944–948。現行 `todo-issue.md` は要約のみ。原文行は本票では再掲しない）
- **既存の高危険表示は未実装と扱わない。** 飼主一覧の `⚠ 危険` バッジと理由 Popover を保持する
- **危険度「高」の理由必須** を保持する（入力・サーバ双方）
- **黄色の意味を推測して 中/低へ自動割当しない。** 未裁定なら色意味の変更は停止

本票は [Pet danger_level / danger_reason](../../../backend/internal/model/pet.go) → [DANGER_LEVEL_MAP](../../../frontend/src/lib/transforms/pet.ts) → [OwnersListTable 高バッジ](../../../frontend/src/features/owners/components/OwnersListTable.tsx) の現行経路を引用する。製品コード・テスト・色トークンは変更しない。

呼び出し行: **無い。** 本ファイルは製品コードから import されない。キャンペーン remaining-ops-20260920 revision 1 unit `SLACK-DANGER` の owned path および人間が読む調査票である。`todo-issue.md` L190–194 は出典要約であり本票の代替ではない。`docs/work/todo-campaign-20260919-ready17/` に本 unit の票は無い。ledger `owned_paths` が本パス単体のため、他シートへの追記では unit 完了にならない。

HEAD at write: `873685b0bea3692c2f8100b19ded660c8357f2b0` (`feat/rem-slack-danger-20260920`)。

## 医院事実（コード外・UNKNOWN）

数値・院内ルール・色の臨床意味をコードから捏造しない。未採取なら該当セルは **UNKNOWN**。

| 項目 | コードで分かること | 医院事実 |
| --- | --- | --- |
| 報告医院・端末・ブラウザ | なし | **UNKNOWN** |
| 要望の「赤」が指す画面 | 一覧バッジは既に `C.danger`（`#C0392B`） | **UNKNOWN**。既存一覧かカルテか受付かは PO |
| 要望の「黄」が指す危険度 | FE ラベルは `低` / `中` / `高`。一覧バッジは `高` のみ | **UNKNOWN。黄 ≠ 中、黄 ≠ 低。自動割当しない** |
| 不明（未設定・未知 wire）の色 | BE enum は `low`/`medium`/`high` のみ。default `'low'` | **UNKNOWN**。不明を黄や灰へ決めない |
| 理由の閲覧場所 | 一覧は 1 操作 Popover。カルテ sticky に危険記号なし | **UNKNOWN**。追加箇所は PO |
| 文字/アイコンでも区別できること | 既存は `⚠ 危険` テキスト + 赤系クラス。中/低は一覧にバッジ無し | 色以外の区別案は **PO**。本票で新ラベルを作らない |
| フロント/API revision | 本票作成時 worktree HEAD `873685b0b` | 再現セッションの SHA は **UNKNOWN** |

## 混ぜてはいけないケース

「危険が色で分かればよい」は次のどれでも同じ言葉になる。既存「高」を欠落と読まない。黄を中/低へ割り当てない。

| ID | 何か | いまのコード | 混ぜてはいけない読み替え |
| --- | --- | --- | --- |
| P — ペット取扱注意 | `pets.danger_level` / `danger_reason`。[pet.go](../../../backend/internal/model/pet.go) L33–38, L59–60 | 飼主の `owners.is_dangerous`（DEC-14）と混用しない |
| H — 一覧の既存「高」 | [OwnersListTable.tsx](../../../frontend/src/features/owners/components/OwnersListTable.tsx) L203–226。`dangerLevel === "高"` のときだけ `⚠ 危険` + `C.danger` | 「赤/黄が無い」=「高表示が無い」と読まない |
| M — 中/低 | 同じテーブルで `中`/`低` はバッジ非表示。[OwnersListTable.report.test.tsx](../../../frontend/src/features/owners/components/OwnersListTable.report.test.tsx) L179–207 | 黄を `中` や `低` に自動マップしない |
| R — 理由必須 | FE [PetEditModal.tsx](../../../frontend/src/features/owners/components/PetEditModal.tsx) L127–128。BE [normalizeDangerReason](../../../backend/internal/pet/owner_registration.go) L219–235 は high で空/空白を拒否。medium は理由なし可（[danger_reason_test.go](../../../backend/internal/pet/danger_reason_test.go) L126–146） | 色追加のために理由必須を外さない。中にも理由必須を勝手に広げない |
| C — カルテヘッダー | `frontend/src/features/medical-records` に `dangerLevel` / `危険` ヒット無し。PatientContextHeader は microchip 等のみ | 一覧バッジがカルテ sticky にも出ていると書かない |
| O — owner-facing | 飼主レポートは danger を非描画（owner-report tests）。仕様 [04-owners-form.md](../../../docs/spec/screens/04-owners-form.md) L37 | 院内バッジを Owner Report / LIFF へ出さない |
| Y — 黄色トークン | `C.danger` は `#C0392B`（赤系）。design-tokens の warning box コメントも red-50/300/700 | 既存 warning クラスを「黄」と見なして中危険へ流用しない |

## 現行経路（表示は「高」だけ。色意味は未裁定）

1. **DB enum:** `danger_level AS ENUM ('low', 'medium', 'high')`（[testdb.go](../../../backend/internal/testdb/testdb.go) L302）。default はモデル上 `'low'`。
2. **モデル:** [pet.go](../../../backend/internal/model/pet.go) L33–38 `DangerLevelLow/Medium/High`、L59 `DangerLevel`、L60 `DangerReason *string`。
3. **高の理由必須（サーバ）:** [owner_registration.go](../../../backend/internal/pet/owner_registration.go) L219–235。`level == DangerLevelHigh` かつ reason nil / 空白 trim 後空 → InvalidInput「危険度がhighの場合は危険理由を入力してください」。medium は理由なしで persist 可。
4. **FE ラベル変換:** [pet.ts](../../../frontend/src/lib/transforms/pet.ts) L70–80。`low→低` / `medium→中` / `high→高`。逆写像も同じ3値。未知 wire は `?? p.danger_level` のまま通し、黄へ落とさない。
5. **入力 UI:** [PetPhysicalSection.tsx](../../../frontend/src/features/owners/components/PetPhysicalSection.tsx) L143–178。選択肢は [DANGER_LEVEL_VALUES](../../../frontend/src/features/owners/types/index.ts) L15 `["低", "中", "高"]`。理由欄の `*` / `aria-required` は `高` のときだけ。
6. **入力 validation:** [PetEditModal.tsx](../../../frontend/src/features/owners/components/PetEditModal.tsx) L127–128 `dangerLevel === "高" && !trimmedDangerReason`。
7. **飼主一覧バッジ（既存「高」表示）:** OwnersListTable.tsx L203–226。条件は **`pet.dangerLevel === "高"` のみ**。trigger 文言 `⚠ 危険`。クラス `C.bgDanger10` `C.danger` `C.borderDanger20`（[design-tokens.ts](../../../frontend/src/lib/design-tokens.ts) L59 `#C0392B`, L280–288）。Popover 1 操作（click / テストは Enter/Space も）で `dangerReason?.trim() || "理由未登録"`。
8. **一覧テストが規定する範囲:** OwnersListTable.report.test.tsx L179–207。high だけ trigger + 視覚クラス。medium/low はボタン無し。`getAllByText("⚠ 危険")` は length 1。
9. **同型の高専用バッジ:** [PetSelectionResultsTable.tsx](../../../frontend/src/components/shared/PetSelection/PetSelectionResultsTable.tsx) L141 付近も `dangerLevel === "高"` のみ。中/未判定は非表示（同ディレクトリ test L390–410）。
10. **仕様書:** [03-owners-list.md](../../../docs/spec/screens/03-owners-list.md) L24 飼主名横に `⚠ 危険`（`danger_level=high`）。[04-owners-form.md](../../../docs/spec/screens/04-owners-form.md) L34–60 は `高 / 中 / 低` 設定と、高のとき一覧警告・理由 Popover。中の理由は任意。

**結論（既存）:** 「高」の赤系バッジと理由開示は実装済み。欠落しているのは **中/低/不明を赤/黄のどれで出すか、どこに出すか** であり、既存「高」ではない。

## 要望との差分（実装しない。PO 待ち）

| 要望（todo-issue L192–194） | 現行 | 本票の扱い |
| --- | --- | --- |
| 危険度に応じ赤/黄で分かる | 高だけ赤系 `C.danger`。中/低は一覧で無バッジ | 追加色は **PO**。黄の割当は **UNKNOWN** |
| 既存高危険表示の保持 | OwnersListTable + PetSelection の `⚠ 危険` | **保持。未実装と書かない** |
| 高/中/低・不明と赤/黄の対応 | ラベル3値。色は高のみ。不明の独立 enum なし | 対応表を本票で埋めない |
| 表示場所 | 飼主一覧の飼主名横、ペット選択結果。カルテ sticky なし | 追加場所は **PO**。推測でヘッダーへ足さない |
| 理由の閲覧 | 高バッジ Popover。空は `理由未登録` | 高の必須と Popover を保持 |
| 色以外の文字/アイコンでも区別 | 高は `⚠ 危険` テキストあり。中/低は文字も無し | 中/低用の新文言は作らない |
| 黄色を中/低へ自動割当しない | コードに yellow→medium/low マップ無し | **禁止を維持。本票も割当しない** |

配置・色の比較（**採用しない。記録のみ**）:

| ID | 案 | 既存「高」 | 黄の意味 | 判定 |
| --- | --- | --- | --- | --- |
| **O0 現状** | 高だけ `⚠ 危険`（赤系） | 保持 | 未使用 | 要望の赤/黄は未充足。高は充足 |
| **O1 高を赤・他を黄** | 中または低または両方を黄 | 保持し得る | **未裁定** | **停止。** 黄=中とも黄=低とも書かない |
| **O2 カルテ sticky へ複製** | PatientContextHeader に危険バッジ | 一覧は残る | 場所が未裁定 | 表示場所 PO 待ち。第二ストアは禁止 |
| **O3 高の理由必須を緩める** | 色だけで足りるとみなす | 壊す | 無関係 | **棄却。** L194 の保持条件に反する |

設計として残すのは **O0 の事実記載** と、PO が赤/黄・場所・不明・文字代替を決めるまで **O1/O2 を実装しない** こと。O3 は禁止。

## 完了 / PO・停止

todo-issue L194 を本票に落とす:

- 既存一覧の「高」`⚠ 危険`（OwnersListTable.tsx L203–226）は **ある**。色追加の前提は保持であり欠落補完ではない
- 高/中/低・不明、赤/黄の対応、表示場所、理由の閲覧範囲は **臨床 PO**
- 色以外の文字/アイコン区別も PO。本票で新ラベルを決めない
- **黄色の意味を推測して 中/低へ自動割当しない**
- 高の理由必須（FE + `normalizeDangerReason`）を保持
- 患者切替と狭い画面の確認は実装後受入。本票では実機 **UNKNOWN**
- 未裁定なら色意味の変更・新バッジ・カルテヘッダー追加は停止

本票は既存-versus-要望の引用まで。製品コードは変更していない。色マッピングは **UNKNOWN**。
