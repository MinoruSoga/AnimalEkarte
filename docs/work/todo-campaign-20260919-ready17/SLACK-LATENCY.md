# SLACK-LATENCY: 治療数量の表示 / commit / 通信 / 再取得の採時計画

状態: **計測計画 READY／実測 未実行（端末・回線・行数・IME・許容 ms はすべて UNKNOWN）**。出典は [todo-issue.md](../../../todo-issue.md) 見出し `### SLACK-LATENCY`（L112–116）。保持する現場条件:

- 「数量入力など反映に時間がかかりすぎる」（出典 883–896。現行 `todo-issue.md` は要約のみ。原文行は本票では再掲しない）
- **2回 Enter の受入とは独立**の計測課題。Enter×2 を「遅さ」の原因断定にも、廃止提案にも使わない
- 許容時間は現場の目的と実測から合意する。本票は架空の性能値・原因・debounce・楽観 persist を置かない

本票は [TreatmentQuantityCell](../../../frontend/src/features/medical-records/components/TreatmentsTab/TreatmentQuantityCell.tsx) → [use-treatments-tab](../../../frontend/src/features/medical-records/hooks/use-treatments-tab.ts) → [treatments API](../../../frontend/src/features/medical-records/api/treatments.ts) の待ちを **表示 / commit / 通信 / 再取得** に分割する。製品コード・テストは変更しない。

本ファイルは製品コードから import されない。キャンペーン unit `SLACK-LATENCY` の owned path および人間が読む計測票である。製品モジュールからの呼び出し行は無い（sibling `SLACK-STAFF-SELECT.md` と同じ）。既存 `docs/work/todo-campaign-20260918/` にも本 unit の票は無い。ledger `owned_paths` が本パス単体のため、他シートへの追記では unit 完了にならない。

## 医院事実・環境（コード外・UNKNOWN）

数値を計算で捏造しない。未採取なら該当セルは **計測 BLOCKED**。

| 項目 | コードで分かること | 医院事実 |
| --- | --- | --- |
| 対象端末クラス（`DEVICE-PC` / `DEVICE-TABLET`）・OS・ブラウザ版 | なし。クラスラベルのみプロトコルで固定（機種名は書かない） | **UNKNOWN**。採取時にクラスを選び、機種・OS・ブラウザは採取ログ側 |
| 回線（院内 LAN / モバイル / VPN） | なし | **UNKNOWN** |
| 対象カルテの治療行数帯（`BAND-FEW` / `BAND-TYPICAL` / `BAND-HEAVY`）と実数 | `sortedTreatments` は全件ソートして描画（[use-treatments-tab.ts](../../../frontend/src/features/medical-records/hooks/use-treatments-tab.ts) L91–95、[TreatmentsTabParts.tsx](../../../frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTabParts.tsx) L85–104） | **UNKNOWN**。帯名と実測行数を採取時に記録。帯の件数閾値は本票に置かない |
| 数量入力時の IME（`IME-OFF` / `IME-ON`。日本語変換中 Enter 含む） | IME Enter は arm/commit しない（後述） | **UNKNOWN**。採取時に IME-OFF/ON をランごとに固定する |
| フロント/API revision 識別方法 | 方法のみ固定: 実測セッション開始時の `git rev-parse HEAD` を必須。フロント bundle 識別が取れる場合のみ併記。本票作成時 HEAD `aac697645` は条件値に使わない | **UNKNOWN**（実測セッションの SHA / bundle）。架空 SHA 禁止 |
| 「遅すぎる」の許容 ms | なし | **UNKNOWN**。実測後に臨床 PO が合意するまで閾値を書かない |

## 混ぜてはいけない待ち

現場の「反映が遅い」は次のどれでも同じ言葉になる。段階を合算した一本のタイマーで原因を決めない。2回 Enter の受入（別課題）と本票の採時を混ぜない。

| 段階 | 何が終わるまで「反映していない」に見えるか | いまのコード | やってはいけないこと |
| --- | --- | --- | --- |
| D 表示 | キー入力が input に出るまで | `localQuantity` の `useState`。`onChange` で即 `setLocalQuantity`（TreatmentQuantityCell.tsx L39, L139–141）。debounce なし（medical-records 配下に debounce/optimistic 実装なし） | 表示遅延を persist 遅延と同一視しない。入力に debounce を足さない |
| C commit | 確定操作が `onUpdate` を呼ぶまで | 2回目 Enter または Blur。1回目 Enter は arm のみ（`reduceQuantityEnterKey`）。用量ゲートでブロック/理由待ちなら `onUpdate` しない | 1回目 Enter の「何も起きない」を通信遅延に数えない。Enter×2 を本票で廃止しない |
| N 通信 | PATCH が HTTP 完了するまで | `useUpdateTreatment` の `axios.patch`。`mutate` は fire-and-forget（handleUpdate L122–127）。`isMutating` 中は全行 `pointer-events-none` | PATCH 時間を GET 再取得に混ぜない。楽観 persist を先行実装しない |
| R 再取得 | 一覧 GET がセル表示をサーバ値に同期するまで | PATCH 成功後 `invalidateQueries` のみ。PATCH 応答の `Treatment` を cache に `setQueryData` しない。非編集時の表示は `treatment.quantity`（L162） | cache 更新待ちを「input が重い」に数えない。再取得を止める提案で整合を捨てない |

## 現行経路（数量セル）

1. **入口:** [TreatmentsTab](../../../frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTab.tsx) L36 が `useTreatmentsTab`。表は [TreatmentsTable](../../../frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTabParts.tsx) L52–108。各行 [TreatmentRow](../../../frontend/src/features/medical-records/components/TreatmentsTab/TreatmentRow.tsx) が `editField === "quantity"` のとき [TreatmentQuantityCell](../../../frontend/src/features/medical-records/components/TreatmentsTab/TreatmentQuantityCell.tsx) を編集表示にする（TreatmentRow.tsx L122–129）。
2. **表示（D）:** 編集中は `<Input type="number">` の `value={localQuantity}`。クリック開始で `treatment.quantity` を文字列化（L39, L51–58 の `treatment` 同期 effect）。キー入力は `onChange` → `setLocalQuantity` のみ。用量ゲート [useTreatmentDoseGate](../../../frontend/src/features/medical-records/hooks/use-treatment-dose-gate.ts) は `localQuantity` から `pendingGate` を再計算する（L54–58）。薬剤行の dose-params は `useMedicineDoseParams`、`staleTime: QUERY_STALE_TIMES.STATIC`（30 分、[medicine-dose-lookup.ts](../../../frontend/src/features/medical-records/api/medicine-dose-lookup.ts) L27–33）。**キーごとの dose-params 再 fetch はコード上ない**（cache miss 初回のみ N に近い待ちになり得る。実測で分離）。
3. **commit（C）:**
   - Enter: `isComposing` または `keyCode === 229` または `repeat` なら return（TreatmentQuantityCell.tsx L100–104）。それ以外は [reduceQuantityEnterKey](../../../frontend/src/features/medical-records/lib/treatment-quantity-commit.ts) L14–28。idle→armed（`action: "none"`）、armed→commit。Escape は cancel。
   - Blur: `onBlur={commitQuantity}`（L143）。Enter 回数に依存しない。テスト: [TreatmentQuantityCell.test.tsx](../../../frontend/src/features/medical-records/components/TreatmentsTab/TreatmentQuantityCell.test.tsx) L94–106。
   - [commitTreatmentQuantity](../../../frontend/src/features/medical-records/lib/treatment-quantity-commit.ts) L43–77: parseFloat、用量ブロックなら旧値へ戻して編集終了、逸脱理由が必要なら理由 UI を出して `onUpdate` しない、同一なら no-op、それ以外は `onStopEdit()` のあと `onUpdate(id, { quantity })`。
4. **通信（N）:** `handleUpdate` は `canEdit` なら `updateTreatmentFn({ treatmentId, input })`（use-treatments-tab.ts L122–127）。[useUpdateTreatment](../../../frontend/src/features/medical-records/api/treatments.ts) L72–93 が `PATCH /v1/medical-records/:id/treatments/:treatmentId`。BE は [TreatmentHandler.UpdateTreatment](../../../backend/internal/medicalrecord/treatment_handler.go) L177 以降 → [treatmentService.Update](../../../backend/internal/medicalrecord/treatment_service.go) L228–299。quantity 変更は `doseRelevant` で同一 TX 内評価（[treatment_service_tx.go](../../../backend/internal/medicalrecord/treatment_service_tx.go) L96–136）。PATCH 応答後に `FindByID` 再読込がある（treatment_service.go L295–298）。**これは N（サーバ内）であり R（FE 一覧 GET）ではない。**
5. **再取得（R）:** `onSuccess` は `queryClient.invalidateQueries({ queryKey: queryKeys.medicalRecords.treatments(medicalRecordId) })`（treatments.ts L84–87）。キーは clinic 付きなら `["treatments", id, clinicId]`、省略時 `["treatments", id]`（[query-keys.ts](../../../frontend/src/lib/query-keys.ts) L362–365）。prefix invalidate は clinic 付き query にも当たる。一覧は [useGetTreatments](../../../frontend/src/features/medical-records/api/treatments.ts) L27–42 `GET /v1/medical-records/:id/treatments`、`staleTime: QUERY_STALE_TIMES.REALTIME`（2 分）。invalidate 後は再 GET。楽観 cache なし。
6. **commit 直後の見た目:** `onStopEdit` で非編集ボタンに戻る。ボタン表示は `showDeviationReason ? localQuantity : treatment.quantity`（TreatmentQuantityCell.tsx L162）。通常 commit は `showDeviationReason` を false にする（treatment-quantity-commit.ts L72）。**PATCH/GET 完了前は旧 `treatment.quantity` が見える。** これは現行契約の観察であり、楽観 persist の実装指示ではない。
7. **表全体の操作ロック:** `isMutating = updateMutation.isPending || deleteMutation.isPending || reorderMutation.isPending`（use-treatments-tab.ts L338）。TreatmentsTab.tsx L54 が全行へ `isUpdating` を渡す。TreatmentRow.tsx L105–107 は `isUpdating` のとき行に `pointer-events-none`。**N のあいだ全行が触れない。** `isPending` は mutation 完了で落ちる。R の GET 中は必ずしもロックしない。ロック解除と数字の同期は別タイマー。

## 既存 Enter / Blur / IME（維持する制約）

計測や後続修正で壊してはいけない。根拠はテストと実装。許容時間の代わりにしない。

| 操作 | 契約 | 根拠 |
| --- | --- | --- |
| 1回目 Enter | `onUpdate` しない（arm） | TreatmentQuantityCell.test.tsx L77–87、reduceQuantityEnterKey idle→armed |
| 2回目 Enter（非 IME・非 repeat） | 数量を一度 persist | 同テスト L89–91、reduceQuantityEnterKey armed→commit |
| Blur | Enter なしで commit | 同テスト L94–106、`onBlur={commitQuantity}` |
| Escape | arm 後も `onUpdate` なし、編集終了 | 同テスト L108–123 |
| 長押し Enter `repeat` | 2回目確定に数えない | 同テスト L125–138、L102–104 |
| IME composition Enter（`isComposing` / keyCode 229） | arm にも commit にも使わない | 同テスト L140–166、L100–104 |
| スクリーンリーダ指示 | 「Enterを2回押して確定」 | TreatmentQuantityCell.tsx L165–167 |

2回 Enter は **受入仕様として残す**。本票は「反映が遅い」をこの待ちと分離して測るだけである。

## 採時プロトコル（実測前に固定する軸）

同一医院・同一ログイン・同一カルテ・同一行で、次を記録してからタイマーを置く。環境不足なら計測 BLOCKED。

### 固定条件（軸を凍結。値は採取時。未採取は UNKNOWN）

方法・プレースホルダだけを凍結する。機種名・病院行数・許容 ms・debounce・楽観 persist は捏造しない。

| 軸 | 凍結した方法 / プレースホルダ | 採取時の記入 | 捏造禁止 |
| --- | --- | --- | --- |
| revision 識別 | **方法固定:** (1) 実測セッション開始時の `git rev-parse HEAD` を必須記録 (2) フロント bundle 識別が DevTools / ビルド成果物から取れる場合のみ併記。取れなければ (1) のみでラン成立。本票作成時 HEAD は条件値に使わない | 実測時 SHA（と任意の bundle id）。未採取は UNKNOWN | 架空 SHA・旧 HEAD の流用禁止 |
| 端末クラス | **プレースホルダ固定:** `DEVICE-PC` / `DEVICE-TABLET`。機種名・OS・ブラウザ版は医院事実のまま UNKNOWN とし、本票に具体モデルを書かない | ランごとにクラスラベルを選ぶ。具体機種は採取ログ側 | 機種名・版の捏造禁止 |
| 行数帯 | **名前付き帯のみ固定:** `BAND-FEW` / `BAND-TYPICAL` / `BAND-HEAVY`。各帯の件数閾値は本票に置かない。実カルテの実数を採取時に併記。合成行で医院事実を置換しない | 帯名 + 実測行数。未採取は UNKNOWN | 病院行数の仮置き・閾値 ms 化禁止 |
| IME | **ラベル固定:** `IME-OFF`（半角数字）と `IME-ON`（対象端末の日本語 IME）を別ラン。IME composition Enter は arm/commit に使わない（既存契約） | ランラベルに IME-OFF / IME-ON | IME 条件を混ぜた平均禁止 |
| 確定操作 | **分離固定:** `OP-BLUR` と `OP-ENTER-x2` を別表・別ラン。Enter×2 では C-arm（1回目）と C1-Enter（2回目）を分ける。Blur と Enter×2 を合算しない | 操作ラベル OP-BLUR / OP-ENTER-x2 | Enter×1 化や仕様変更を本票で提案しない |
| 項目種別 | **分離固定:** 少なくとも **consultation（用量ゲートなし）** と **medicine（用量ゲートあり）** を分ける | 種別ラベル | 混ぜた平均を出さない |

タイマー（推奨。ツールが無ければ DevTools Performance + Network で同じ境界を手で印す）:

| ID | 始点 | 終点 | 段階 |
| --- | --- | --- | --- |
| D1 | `onChange`（キーが localQuantity に入る） | 次の paint（`requestAnimationFrame` 2 回、または PerformanceObserver `measure`） | D |
| C-arm | 1回目 `keydown` Enter（非 IME） | `reduceQuantityEnterKey` が armed を返した直後 | C（persist ではない） |
| C1-Enter | 2回目 `keydown` Enter（非 IME） | `onUpdate` 呼び出し | C |
| C1-Blur | `blur` | `onUpdate` 呼び出し | C |
| C0 | 確定操作だが `onUpdate` なし（同一値 / 用量ブロック / 理由待ち） | `onStopEdit` | C（通信なし）。N/R を 0 と記録し欠損扱いにしない |
| N1 | `axios.patch` 開始（Network パネル PATCH） | PATCH 応答完了（status 記録） | N |
| N2 | サーバ受領（あれば access log / trace） | handler return。TX + dose 再評価 + FindByID を内訳できれば記す | N（サーバ）。取れなければ N1 のみ |
| L1 | `isMutating` true | `isMutating` false（全行 pointer-events 解除） | 主に N。R を含めない |
| R1 | PATCH `onSuccess` / invalidate | `GET .../treatments` 開始 | R 待ち行列 |
| R2 | GET 開始 | GET 完了 | R 通信 |
| R3 | GET 完了 | 非編集セルのテキストが新数量（保存値一致） | R 描画 |
| T-user | ユーザーが「確定した」と感じる操作（2回目 Enter または Blur） | セル数字が新数量 | 合算観察。内訳 D/C/N/R なしでは原因に使わない |

保存値一致: セル表示・React Query cache・`GET` ボディの `quantity` が同じこと。一致するまでの時間が R。一致しないなら本票の「遅さ」ではなく不整合（別票）。

並行負荷: 行数帯を変えたラン（`BAND-FEW` / `BAND-TYPICAL` / `BAND-HEAVY`）を分ける。全行 `pointer-events-none` と一覧 GET は行数に感ずる可能性がある。感度は実測。行数を増やした合成データで医院事実を置き換えない。帯の件数閾値は実測後に医院事実として記録し、本票へ仮置きしない。

## 停止・PO

- 実測前に debounce・楽観更新・Enter×1 化・staleTime 変更を実装しない
- 許容 ms は本票に書かない。実測表が揃ってから臨床 PO が「業務上の目的」付きで合意する
- 対象環境が無い、行数/IME/revision を固定できない、または Network で PATCH と GET を分けられない場合、そのランは **計測 BLOCKED**
- 2回 Enter の受入は独立。本票完了は「内訳付き前後計測と保存値一致」であり、Enter 仕様変更ではない
- 用量逸脱理由 UI が出るランは C0。理由入力後の commit を別行にする

## 実測結果（未実行）

固定条件の軸列だけ揃える。実測セルは未実行のまま。閾値列・debounce 列は設けない。

| ラン | 端末クラス | 行数帯 | 行数(実測) | revision | IME | 操作 | D1 | C | N1 | L1 | R2 | R3 | T-user | 保存値一致 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| — | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN | — | 未実行 | 未実行 | 未実行 | 未実行 | 未実行 | 未実行 | 未実行 | 未実行 |

閾値列は設けない（UNKNOWN のまま空けると架空の欄になるため）。
