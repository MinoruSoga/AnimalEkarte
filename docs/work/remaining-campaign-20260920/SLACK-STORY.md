# SLACK-STORY: 名前の由来と出逢い — 既存 `pets.remarks` 再利用 vs 独立フィールド（採否 UNKNOWN）

状態: **再利用/新規欄の比較 READY／製品実装・スキーマ追加 未実行（停止）**。出典は [todo-issue.md](../../../todo-issue.md) 見出し `### SLACK-STORY`（L196–200、索引 L425）。保持する現場条件:

- 「名前の由来と出逢いのストーリー欄」（出典 949–955。現行 `todo-issue.md` は要約のみ。原文行は本票では再掲しない）
- **DETAILS とは別の記録要望。** 詳細閲覧導線は [SLACK-DETAILS](../todo-campaign-20260919-ready17/SLACK-DETAILS.md)。本票は保存先の比較だけ
- 取得区分（`acquisition_type`）をストーリー全文の保存先とみなさない
- 診療記録や会計への自動転記はしない
- 独立フィールド / DB 追加は**採用まで停止**。削減工程が無い欄追加は目的を再検討する

本票は [Pet](../../../backend/internal/model/pet.go) の取得区分/備考と [PetEditModalFields](../../../frontend/src/features/owners/components/PetEditModalFields.tsx)（実入力は [PetCareSection](../../../frontend/src/features/owners/components/PetCareSection.tsx) / [PetPhysicalSection](../../../frontend/src/features/owners/components/PetPhysicalSection.tsx)）をトレースする。製品コード・テスト・migration は変更しない。

呼び出し行: **無い。** 本ファイルは製品コードから import されない。キャンペーン unit `SLACK-STORY` の owned path および人間が読む比較票である。`todo-issue.md` L196–200 は出典要約であり本票の代替ではない。ledger `owned_paths` が本パス単体のため、他シートへの追記では unit 完了にならない。既存 `docs/work/todo-campaign-20260919-ready17/` に本 unit の票は無く、[SLACK-INTAKE](../todo-campaign-20260919-ready17/SLACK-INTAKE.md) L79 / L106 は分類ポインタのみ。

採否（既存備考で足りるか / 独立欄を作るか）は **UNKNOWN**。本票は比較まで。採用を決めない。

## 医院事実（コード外・UNKNOWN）

数値・院内ルール・PO 裁定をコードから捏造しない。未採取なら該当セルは **再現 BLOCKED**。

| 項目 | コードで分かること | 医院事実 |
| --- | --- | --- |
| 報告医院・端末・ブラウザ | なし | **UNKNOWN** |
| 記録者（誰が書くか） | ペット編集は `owners` の `canEdit`（[PetEditModal.tsx](../../../frontend/src/features/owners/components/PetEditModal.tsx) L58, L196, L244–248） | **UNKNOWN**。受付か獣医師か飼主か本票で決めない |
| 読む人（誰がいつ読むか） | 飼主画面の備考列・ペット編集 textarea・飼主レポート「診療メモ」 | **UNKNOWN**。カルテヘッダーには出ない（後述） |
| 業務目的 | 「欄が欲しい」は要件ではない（[product-philosophy.md](../../product-philosophy.md) ①） | **UNKNOWN**。責任者の個人名も未記録 |
| 既存備考で足りるか | 同一 `pets.remarks` に自由文 2000 字まで入る | **UNKNOWN**。PO 採否 |
| 必要な表示先 | 編集 UI / 飼主ペット一覧 / レポート。カルテ sticky は未配線 | **UNKNOWN** |
| 上限 | BE `max=2000`。FE textarea に `maxLength` 無し | **UNKNOWN**。院内の想定長を 2000 と決めない |
| 編集権限 | `owners` 権限。fieldset `disabled={!canEdit}` | **UNKNOWN**。カルテ編集者に書くか本票で合算しない |
| 個人情報の範囲 | 由来・出逢いは飼主・第三者を含み得る。既存備考は注意文の fixture もある | **UNKNOWN**。PII 範囲を本票で採用しない |
| フロント/API revision | 本票作成時 worktree HEAD `873685b0b` | 再現セッションの SHA は **UNKNOWN** |

## 混ぜてはいけないケース

現場の「ストーリー欄」は次のどれでも同じ言葉になる。保存先と閲覧導線と enum を混ぜない。

| ID | 何か | いまのコード | 混ぜてはいけない読み替え |
| --- | --- | --- | --- |
| R — ペット備考 | `pets.remarks` text NOT NULL DEFAULT ''。[pet.go](../../../backend/internal/model/pet.go) L66。JSON `remarks`。UI「備考・特記事項」 | 飼主備考・接種備考・トリミング備考・カルテ SOAP と同一視しない |
| O — 飼主備考 | `owners.remarks`（[001_init.sql](../../../backend/migrations/001_init.sql) L309）。[OwnerBasicFields.tsx](../../../frontend/src/features/owners/components/OwnerBasicFields.tsx) L129–130 も「備考・特記事項」 | ペットの出逢いを飼主マスタへ書かない。ラベルが同じでもテーブルが違う |
| A — 入手区分 | `pets.acquisition_type` ENUM `purchased`/`transferred`/`rescued`/`other`（001_init L79, L1125）。UI「入手区分」（PetPhysicalSection L123–139） | **ストーリー全文の保存先にしない。** todo-issue L199 |
| D — 詳細導線 | SLACK-DETAILS。カルテヘッダーから OwnerForm / PetEditModal を開く話 | 保存カラム追加と詳細閲覧を同一チケットにしない（INTAKE 出典 949–955 を DETAILS+STORY に分離済み） |
| M — 診療メモ表示 | 飼主レポート [OwnerClinicalVisitPanels.tsx](../../../frontend/src/features/owner-report/components/OwnerClinicalVisitPanels.tsx) L116–119 が `pet.remarks` をラベル「診療メモ」で出す | レポート別名を新カラムの根拠にしない。同じ R の読取 |
| V — 接種/トリミング備考 | `vaccinations.remarks`（001_init L1492）、`appointment_trimming_details.remarks`（L1386） | 来院単位の備考へペット生涯ストーリーを載せない |
| C — カルテ/会計転記 | 本票作成時、カルテ sticky / PatientContextHeader に `remarks` prop は無い | ストーリーを診察記録や請求行へ自動コピーしない（todo-issue L200） |
| S2 — 独立ストーリー列 | `pets` に origin/story 列は無い | 未採用の列を「足りないから足す」と読まない。PO 採否まで停止 |

## 現行経路（ペット備考は 1 系統）

ストーリー用の第二ストアは無い。自由文の正本は `pets.remarks` だけ。入手区分は別 enum。

1. **DB:** [001_init.sql](../../../backend/migrations/001_init.sql) L1132 `remarks text NOT NULL DEFAULT ''`。COMMENT は死亡日・血液型・チップにあり、`remarks` には用途 COMMENT が無い。
2. **モデル:** [pet.go](../../../backend/internal/model/pet.go) L58–66。`AcquisitionType *AcquisitionType`（L58）と `Remarks string` `json:"remarks"`（L66）は隣接するが型が違う。
3. **作成 bind:** [pet_request.go](../../../backend/internal/pet/pet_request.go) L113 `AcquisitionType string`（enum 検証はこの struct tag には無い）。L120 `Remarks string` `binding:"omitempty,max=2000"`。飼主登録ネストも同じ max=2000（[owner/http_request.go](../../../backend/internal/owner/http_request.go) L131, L137）。
4. **更新 bind:** pet_request.go L168 `AcquisitionType *string`、L176 `Remarks *string` `omitempty,max=2000`。
5. **FE 変換:** [pet.ts](../../../frontend/src/lib/transforms/pet.ts) L56–68 が入手区分を 購入/譲渡/保護/その他へ写す。L127 `remarks: p.remarks`。create L194–196 / L203、update L245–247 / L257。
6. **フォーム型:** [types/index.ts](../../../frontend/src/features/owners/types/index.ts) L12 `ACQUISITION_TYPE_VALUES`、L48 `remarks: string`。[pet-form-data.ts](../../../frontend/src/features/owners/lib/pet-form-data.ts) L21–22 / L28 が初期値。
7. **入力 UI:** PetEditModalFields は欄を持たず 3 セクションを合成する（L98–125）。
   - 入手区分: PetPhysicalSection L123–139 `Select`（4 値のみ）
   - 備考: PetCareSection L184–194 `Textarea` id=`remarks` ラベル「備考・特記事項」`rows={3}`。**`maxLength` 無し**（上限は API 2000）
8. **保存:** [use-pet-form-list-state.ts](../../../frontend/src/features/owners/hooks/use-pet-form-list-state.ts) L161 / L219 が `remarks: petData.remarks` を update/create に載せる。[OwnersList.tsx](../../../frontend/src/features/owners/routes/OwnersList.tsx) L85 / L280 も同じフィールド。
9. **権限:** PetEditModal L196 `fieldset disabled={!canEdit}`。Save は `canEdit` のときだけ（L244–248）。
10. **一覧表示:** [OwnerPetsSection.tsx](../../../frontend/src/features/owners/components/OwnerPetsSection.tsx) L241 / L263 が備考列（`truncate max-w-[200px]`）。
11. **レポート読取:** 同じ `pet.remarks` を「診療メモ」として出す（OwnerClinicalVisitPanels L116–119）。空は「記載なし」。値ありで `alert`。
12. **テストが示す現行用途:** PetCareSection.test fixture `remarks: "咬傷注意"`（L38）。入力回帰は「備考・特記事項」へ「合成備考」（L203–231）。**由来・出逢い専用のテストは無い。**

カルテヘッダーへ備考は渡っていない。表示したい場合でも **第二カラムではなく既存 R の配線**が候補であり、本票では実装しない（SLACK-MICROCHIP と同型の「persist ではなく props」問題になり得る。表示先は PO）。

## オプション比較（実装しない）

[product-philosophy.md](../../product-philosophy.md) の順序: ①要件を疑う → ②削除 → ③簡素化。欄追加は ①② を通過するまで止める。

実践ゲート（同文書）:

- ① 責任者の個人名: **UNKNOWN**。業務目的も「欄が欲しい」のまま
- ② 削除できる工程: 独立欄は **追加だけで削除ゼロ**（二重入力の温床）
- ③–⑤: 採否前に最適化・自動化しない。診療/会計への自動転記は禁止

| ID | 案 | 保存先 | 二重入力 | 工程削除 | スキーマ | 判定 |
| --- | --- | --- | --- | --- | --- | --- |
| **O0 現状** | 備考に自由文。入手区分は enum | `pets.remarks` + `acquisition_type` | 無し | 無し | 変更なし | コード上は由来文を備考へ書ける。ラベルは「ストーリー」ではない。レポートでは「診療メモ」 |
| **O1 既存備考を再利用（推奨候補・未採用）** | 由来・出逢いを `pets.remarks` に書く。ラベル変更やプレースホルダは PO。入手区分は「買い/譲渡/保護」のまま触らない | 既存 R | **無し**（第二欄を作らない） | 新規入力工程を増やさない | **カラム追加なし** | ①の目的が「ペットマスタに自由文を残す」なら足りる可能性。注意文（咬傷）と物語が混ざるリスクは PO。上限 2000・PII・表示先も PO |
| **O2 独立ストーリー欄** | `pets.story` 等の新列 + PetEditModalFields に textarea | 新列 | **備考と二重。** どちらに書くか現場が分岐 | **削除ゼロ** | migration 必須 | **採用まで停止。** todo-issue L200。本キャンペーンは DB 列追加禁止 |
| **O3 入手区分に全文を入れる** | `acquisition_type` や `other` の横に長文 | enum | 型破壊 | 無し | 不適合 | **棄却。** L199。Select 4 値は物語を持てない |
| **O4 飼主備考へ書く** | `owners.remarks` | 別テーブル | 同居ペットで衝突 | 無し | 変更なしでも誤用 | **棄却。** ペット単位の出逢いを飼主行に畳まない |
| **O5 カルテ/会計へ自動転記** | 保存時に medical_records や請求へコピー | 複数 SoT | 二重管理 | 無し（転記工程が増える） | 対象外 | **禁止。** L200 |

**設計として先に残す比較は O0（現状）と O1（備考再利用）。** O2 は PO が「注意文と物語を分離する業務目的」と記録者/読む人/PII を確定し、かつ削減工程を示してから別 revision。本 unit では O2 を実装しない。O3–O5 は棄却。

O1 で足りるかの判定材料（PO。本票は埋めない）:

- 書く場所は既にある（PetCareSection「備考・特記事項」）
- 読む場所は飼主ペット一覧の備考列と、飼主レポートの「診療メモ」。カルテ名の近くには出ない
- 既存値は注意文用途のテストがある。物語専用ではない
- ラベルを「ストーリー」に変えると、咬傷注意の発見性が落ちる可能性（未測定）
- 独立欄は第二入力。同じペットに注意と物語の両方を残したい場合だけ O2 が①を通過し得る。その場合でも転記はしない

## ケース表（受入の読み替え防止）

| ID | 条件 | 期待（todo-issue L200） | 現行（コード） |
| --- | --- | --- | --- |
| T0 ペット編集を開く | 由来を書ける場所があるか | 二重入力なしに目的を満たせるかを検討 | 「備考・特記事項」textarea がある。ストーリー専用欄は無い |
| T1 入手区分を選ぶ | 出逢いの全文が残るか | 取得区分を全文保存先にしない | 購入/譲渡/保護/その他の Select のみ |
| T2 備考に長文 | 保存上限 | PO が上限を決める | BE 2000。FE は未制限表示 |
| T3 飼主レポート | 臨床前確認 | 転記ではなく同一 R の読取 | 「診療メモ」=`pet.remarks` |
| T4 カルテヘッダー | 名前の近くで物語を読む | 表示先は PO。新 persist ではない | remarks 未配線 |
| T5 詳細ボタン | ストーリー欄とは別 | DETAILS と分離 | 本票の対象外 |
| T6 新カラム PR | 採用前 | 独立フィールドは採用まで停止 | 本票は比較のみ。migration しない |

## 完了 / PO・停止

todo-issue L200 を本票に落とす:

- 記録者 / 読む人 / 業務目的は **UNKNOWN**。名前のない「欄が欲しい」は①未通過
- 既存備考（O1）で足りるかは **採否 UNKNOWN**。コード上は自由文を保存できる。用途ラベルは注意/診療メモ寄り
- 表示先・上限・編集権限・個人情報の範囲は PO。本票のコード値を院内ルールとして採用しない
- 削減工程が無い独立欄（O2）は目的を再検討。**DB 追加は採用まで停止**
- 診療記録や会計への自動転記はしない
- 取得区分に全文を載せない
- SLACK-DETAILS の導線実装と混ぜない

本票は再利用 vs 新規欄の比較まで。製品コード・migration は変更していない。後続実装は別 revision。採否前にカラムもラベル変更もヘッダー配線もしない。
