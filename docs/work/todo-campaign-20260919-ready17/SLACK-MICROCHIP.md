# SLACK-MICROCHIP: 既存 `pet.microchip_number` を患者名近くへ表示（第二ストア禁止）

> Task migrated to Plane `EMR-185`. This file remains supporting acceptance/evidence material; use Plane for current status.

状態: **表示経路のマッピング READY／製品実装・実機受入 未実行**。出典は [todo-issue.md](../../../todo-issue.md) 見出し `### SLACK-MICROCHIP`（L130–134、索引 L422）。保持する現場条件:

- マイクロチップ番号を**名前の近くへ表示**したい（出典 938–943。現行 `todo-issue.md` は要約のみ。原文行は本票では再掲しない）
- **既存フィールドを一度だけ表示**する。ヘッダー用の別 persist / 架空番号を作らない
- 番号有無・長い表示・患者切替・1366×625 で**正しい患者の値**が欠けず読めること
- 実装後は API 再取得 / 切替時の**古い値残留**を受入で確認する

本票は [pets.microchip_number](../../../backend/internal/model/pet.go) → [transformBackendPetToFrontend](../../../frontend/src/lib/transforms/pet.ts) → [MedicalRecordStickyHeader](../../../frontend/src/features/medical-records/components/MedicalRecordStickyHeader.tsx) → [PatientContextHeader](../../../frontend/src/components/shared/PatientContextHeader/PatientContextHeader.tsx) の props 欠落をトレースする。製品コード・テストは変更しない。

呼び出し行: **無い。** 本ファイルは製品コードから import されない。キャンペーン unit `SLACK-MICROCHIP` の owned path および人間が読む調査票である（sibling `SLACK-VITALS.md` L13 と同じ）。既存 `docs/work/todo-campaign-20260918/` に本 unit の票は無く、`todo-issue.md` L130–134 は出典要約であり本票の代替ではない。ledger `owned_paths` が本パス単体のため、他シートへの追記では unit 完了にならない。

## 医院事実（コード外・UNKNOWN）

数値・院内ルールをコードから捏造しない。未採取なら該当セルは **再現 BLOCKED**。

| 項目 | コードで分かること | 医院事実 |
| --- | --- | --- |
| 報告医院・端末・ブラウザ | なし | **UNKNOWN** |
| 「名前の近く」が飼主名かペット名か | ヘッダーは飼主ボタン + ペット名 span が並ぶ（後述） | **UNKNOWN**。PO が対象ラベルを確定するまで片方へ実装しない |
| 表示目的（識別・照会・転記防止）と権限 | カルテ sticky は既にペット属性を出している。飼主レポートは別画面 | **UNKNOWN**。todo-issue L134 が PO 確認を要求 |
| 未記録時に `-` を出すか行を隠すか | 変換層コメントは `-`、体重チップは falsy で非表示（後述） | **UNKNOWN**。採用値を本票で決めない |
| 現場の番号桁・区切り・先頭ゼロ | BE `max=64`、FE Input `maxLength={64}`。テストfixture に 15 桁例がある | **UNKNOWN**。fixture を院内規格として採用しない |
| 1366×625 実機 CSS viewport | ローカル基準は [CHART-FIT票](../todo-campaign-20260918/UAT-R2-CHART-FIT.md) と同じ 1366×625 CSS px | **UNKNOWN**。実機 `innerWidth`/`innerHeight` は採取前 |
| フロント/API revision | 本票作成時 worktree HEAD `aac697645` | 再現セッションの SHA は **UNKNOWN** |

## 混ぜてはいけないケース

現場の「チップ番号を名前の近くに出したい」は次のどれでも同じ言葉になる。表示と persist を混ぜない。

| ID | 何か | いまのコード | 混ぜてはいけない読み替え |
| --- | --- | --- | --- |
| P — ペットマスタの番号 | `pets.microchip_number`（NULL=未記録）。[pet.go](../../../backend/internal/model/pet.go) L55。JSON `microchip_number` | カルテ行・バイタル・飼主レポート専用カラムと同一視しない |
| H — カルテ sticky ヘッダー | [MedicalRecordStickyHeader](../../../frontend/src/features/medical-records/components/MedicalRecordStickyHeader.tsx) L194–210 が `selectedPet` の名前/体重等を [PatientContextHeader](../../../frontend/src/components/shared/PatientContextHeader/PatientContextHeader.tsx) へ渡す。**`microchipNumber` prop は無い** | 「フィールドがある」=「ヘッダーに出ている」と読まない |
| R — 飼主レポート詳細 | [OwnerClinicalBasicPanel](../../../frontend/src/features/owner-report/components/OwnerClinicalBasicPanel.tsx) L40 が `DetailField` で「マイクロチップ」を出す。空は `-`（[ClinicalBriefingFields](../../../frontend/src/features/owner-report/components/ClinicalBriefingFields.tsx) L66 `value \|\| "-"`） | レポート別窓をカルテ sticky の「名前の近く」と同一視しない |
| E — ペット編集入力 | [PetPhysicalSection](../../../frontend/src/features/owners/components/PetPhysicalSection.tsx) L78–86。保存は既存 pet create/update | ヘッダー表示のために入力欄を複製しない。ヘッダーから PATCH しない |
| C — 同居チップ切替 | StickyHeader L54–60 の Link は `medicalRecords?pet_id=`（**新規カルテ入口**） | 既存カルテの `pet_id` を付け替える操作と混ぜない |
| S2 — 第二ストア | `backend/internal/medicalrecord` に `microchip` 記号は無い（本票作成時 `rg` 0 件） | ヘッダー用カラム・localStorage・カルテ JSON へのコピーを「表示」と呼ない |

## 現行経路（保存は pets 1 系統）

ヘッダーへ **第二ストアを作らない**。入力・保存・再読込の正本は `pets.microchip_number` だけ。

1. **DB:** [001_init.sql](../../../backend/migrations/001_init.sql) L1138 `microchip_number text NULL`。COMMENT L1151「NULL=未記録」。
2. **モデル:** [pet.go](../../../backend/internal/model/pet.go) L55 `MicrochipNumber *string` `json:"microchip_number,omitempty"`。
3. **作成 bind:** [pet_request.go](../../../backend/internal/pet/pet_request.go) L110 `binding:"omitempty,max=64"`（string）。飼主登録ネストも同じ max=64（[owner/http_request.go](../../../backend/internal/owner/http_request.go) L125）。
4. **作成 mapper:** [mapper.go](../../../backend/internal/pet/mapper.go) L45–52。空文字は NULL のまま（「記録済みと誤表示しない」）。[mapper_test.go](../../../backend/internal/pet/mapper_test.go) L34 未指定は `Nil`、L51–63 指定時は `"392140000123456"`。
5. **更新:** [pet_request.go](../../../backend/internal/pet/pet_request.go) L165 `*string` `omitempty,max=64`。[service.go](../../../backend/internal/pet/service.go) L131–132 は `input.MicrochipNumber != nil` のとき `fields[colPetMicrochipNumber] = *input.MicrochipNumber`。**空文字ポインタを送ると NULL ではなく空文字で上書きし得る。** ヘッダー表示案はこれを直さない。クリア契約は PO / 既存 pet 編集の範囲。
6. **応答:** [pet_response.go](../../../backend/internal/pet/pet_response.go) L66（detail）と L121（list）が `microchip_number,omitempty`。
7. **FE 変換:** [pet.ts](../../../frontend/src/lib/transforms/pet.ts) L105–107 `microchipNumber: p.microchip_number ?? undefined`。コメントは「未記録は undefined → 表示側で `-`」。create/update request は L187 / L238。
8. **変換テスト:** [pet.test.ts](../../../frontend/src/lib/transforms/pet.test.ts) L85–98 が値あり / 未設定（捏造しない）。L101–121 が create/update request へ載せる。
9. **入力 UI:** PetPhysicalSection L78–86 `maxLength={64}`。保存は [useUpdatePet](../../../frontend/src/features/pets/api/update-pet.ts) L20–22 が `queryKeys.pets.list()` と `queryKeys.pets.detail(id)` を invalidate。
10. **カルテが読む値:** [use-medical-record-form-helpers.ts](../../../frontend/src/features/medical-records/hooks/use-medical-record-form-helpers.ts) L147–148 `useGetPet(resolvedPetId)`。[use-pet.ts](../../../frontend/src/hooks/use-pet.ts) L53–64 `GET /v1/pets/${petId}` → `transformBackendPetToFrontend`。query key は [query-keys.ts](../../../frontend/src/lib/query-keys.ts) L329 `["pet", id]`。`staleTime` は [QUERY_STALE_TIMES.STATIC](../../../frontend/src/lib/react-query.ts) L48–49 **30 分**。`keepPreviousData` は detail に無い。

飼主レポート GET も同じカラムを別 DTO で読む（[get-owner-report-pets.ts](../../../frontend/src/features/owner-report/api/get-owner-report-pets.ts)）。**別テーブルではない。** ヘッダーがレポート API を追加購読すると第二 read 経路になるので、カルテ sticky は `useGetPet` の `selectedPet` を再利用する。

## ヘッダー props マッピング（欠落）

[MedicalRecordFormReadyPanels](../../../frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) L204–205 が `selectedPet` を StickyHeader へ渡す。本番の PatientContextHeader マウントは StickyHeader のみ。

| selectedPet フィールド | StickyHeader → PatientContextHeader | いまの表示 |
| --- | --- | --- |
| `ownerName` | L195 `ownerName=` | 飼主名（L117–131）。クリックは飼主差替え（SLACK-DETAILS 対象） |
| `name` | L196 `petName=` | ペット名 span（L132–137）。Tooltip で全文 |
| `petNumber` / `id` | L197 `petNumber=` | `#番号`（L148–154） |
| `weight` | L198 `weight={selectedPet.weight ?? undefined}` | **truthy のときだけ** Weight アイコン（L181–188）。空は行ごと出さない |
| `status` / `birthDate` / `species` / `gender` / `neuteredDate` / `breed` / 保険 / `visitCount` | L199–207 | 属性行 |
| **`microchipNumber`** | **渡していない。** PatientContextHeaderProps（L34–52）にフィールド無し | **名前の近くに出ない** |

`rg microchip` を StickyHeader / PatientContextHeader に掛けてもヒットしない（本票作成時）。値は `selectedPet` 上に既にある（Pet 型 = transform 戻り値、[pet.ts](../../../frontend/src/lib/transforms/pet.ts) L135–138）。**不足は persist ではなく header props 配線。**

体重と同じ形の最小案（実装しない。比較用）:

1. `PatientContextHeaderProps` に optional `microchipNumber?: string` を足す
2. StickyHeader が `microchipNumber={selectedPet.microchipNumber}` を渡す（`?? undefined` は変換済み）
3. ペット名隣または属性行に read-only 表示。クリック・input・PATCH を付けない
4. 飼主レポート / PetPhysicalSection / 新規 GET を増やさない

## 空欄 / 長い値 / 患者切替 / 再取得の古い値

| ID | 条件 | 期待（受入・todo-issue L134） | 現行（コード） | ヘッダー案での注意 |
| --- | --- | --- | --- | --- |
| M0 未記録 | DB NULL / JSON 欠落 | 欠けて読める（空と他ペットの番号を混同しない） | transform は `undefined`（pet.test.ts L94–98）。レポートは `-`。体重チップは非表示 | **PO:** `-`（変換コメント）か hidden（体重）。未記録を `"0"` や仮番号で埋めない |
| M1 空文字 | update が `""` を書いた行 | 未記録と同じ扱いか、空文字として残るかは persist 契約 | create mapper は空→NULL。update は nil でなければ空文字を書き得る | 表示は falsy なら M0 と同じに倒す。空文字を「番号あり」としない |
| L0 長い値 | 最大 64 文字（BE bind + FE maxLength） | 1366 幅でも正しい患者の全文が欠けず読める（truncate するなら Tooltip） | 名前は `truncate` + Tooltip（PatientContextHeader.test.tsx L177–201）。体重は truncate 無し | 64 文字をヘッダーに生置きすると保険カード・contextControls と幅競争。**実機は UNKNOWN**。DOM に全文を残す |
| L1 15 桁 fixture | テストの `"392140000123456"` | 値ありの表示 | mapper_test / pet.test の例 | **院内規格ではない。** 15 桁以外を不正としない |
| SW0 既存カルテの患者 | `resolvedPetId = existingRecord?.petId`（helpers L147） | 開いているカルテのペットの番号 | `useGetPet(petId)` の query key が pet id 単位 | カルテ A の番号をカルテ B に残さない |
| SW1 新規 `?pet_id=` | `resolvedPetId = petId`（helpers L147） | URL のペット | 同上 | 新規と既存でソースが違う。同居 Link（C）は SW1 |
| SW2 同居チップ | StickyHeader L57–59 は一覧 `?pet_id=` | **別カルテ入口。** 今の記録のペットは変わらない | チップ label は name/species のみ | 「切替」受入を同居クリックだけで済ませない。SW0 は別記録を開き直す |
| SW3 ロード中 | detail に `keepPreviousData` 無し | 前のペットの番号が一瞬残らない | 新 key の fetch 中は `selectedPet` が undefined → [MedicalRecordForm.tsx](../../../frontend/src/features/medical-records/routes/MedicalRecordForm.tsx) L71–72 が `return null` | 空白フラッシュは番号取り違えより安全。placeholder で前ペットを残す変更は禁止 |
| RF0 同一ペット再取得 | `staleTime` 30 分 | 実装後受入: 再取得/切替で古い値が残らない | 30 分以内は fresh。window focus でも refetch しない（stale でないため） | **残留ポイント。** 別端末で番号変更後、開いたままのカルテ sticky は最大 30 分古い |
| RF1 自画面の pet PATCH | `useUpdatePet` が `["pet", id]` を invalidate | 成功後は新しい番号 | invalidate 後に refetch | ヘッダーが同じ query key を読んでいれば追従する |
| RF2 レポート別窓だけ更新 | レポート経路の invalidate 範囲は別 | カルテ sticky は RF0 のままになり得る | レポート表示は sticky を更新しない | 別窓成功を sticky 更新済みと書かない |
| W0 1366×625 | CHART-FIT ローカル基準 | 名前・番号・9 タブ・保存到達。文字縮小やタブ削除で逃さない | sticky は `flex-wrap`（PatientContextHeader L103）。タブは overflow-x | 実機 inner は UNKNOWN。合格は到達であり、常時 1 行固定ではない |

## 配置オプション比較（実装しない。ヘッダーは表示専用）

| ID | 配置 | 名前の近く | 第二ストア | 1366×625 | 備考 |
| --- | --- | --- | --- | --- | --- |
| **O0 現状** | レポート詳細 / ペット編集にだけある | いいえ | なし | 影響なし | 現場条件を満たさない |
| **O1 selectedPet を sticky へ配線（推奨候補）** | ペット名隣または属性行。`selectedPet.microchipNumber` のみ | はい（ペット名隣が候補。飼主名隣は PO） | **なし。** GET `/v1/pets/:id` 再利用 | wrap + truncate/Tooltip | persist・API・権限モデルを増やさない |
| **O2 飼主レポートを sticky から開くだけ** | 既存レポートボタン（StickyHeader L175–188） | クリック必須 | なし | 別窓 | 「クリックなしで名前の近く」を満たさない。案内には使える |
| **O3 ヘッダーから PATCH** | 名前隣に input | はい | ヘッダー state が第二入力になりやすい | 幅不足 | **棄却。** 編集は PetPhysicalSection |
| **O4 medical_records に番号を複製** | カルテ JSON / 新列 | 出せる | **第二ストア** | 不問 | **禁止。** 切替・再取得で必ず腐る |
| **O5 新規マイクロチップ API** | 別 resource | 出せる | 第二ストア | 不問 | **禁止。** フィールドは既にある |

**設計として先に残すのは O1。** O0 は現状。O2 は補助案内。O3–O5 は第二ストアまたは編集導線の複製。

## 完了 / PO・停止

todo-issue L134 を本票に落とす:

- 対象ヘッダーはカルテ sticky の PatientContextHeader（本番マウントは 1 箇所）
- 表示目的 / 権限 / 飼主名隣かペット名隣か / 未記録を `-` か非表示かは PO
- 二重保存・架空番号・カルテへのコピーは停止
- 受入: 未記録、64 文字、SW0 患者切替、RF0 再取得の古い値、1366×625 で正しい患者の値が欠けず読める
- 実機 viewport・院内桁数は UNKNOWN のまま採取。本票で数値を埋めない

本票は経路トレースと O1 比較まで。製品コードは変更していない。
