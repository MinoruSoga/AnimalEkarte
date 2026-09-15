# STG UAT 医院フィードバック Q1–Q4（回答原文と修正タスク）

作成: 2026-09-15。回答§1 最終更新: 2026-09-15（STG デプロイ後）。  
入口: [docs/work/README.md](./README.md) の補助表。  
範囲: 医院からの4問への回答、および各問に対して実施すべき修正タスク。  
制約: 推測で「デモだから」と説明しない。Linear 新規 Issue は作らない。本ファイルは実行 SoT ではない。  
反映範囲: Q1 / Q4保険 / Q2履歴ナビは **STG デプロイ済み**（PR [#411](https://github.com/MinoruSoga/AnimalEkarte/pull/411)、merge `d337f016`。Backend Deploy #176 / Frontend Deploy #51 success）。production 未反映。ブラウザ確認は未実施。

根拠（コード・仕様。PHI は含めない）:

| 題材 | 根拠 |
|------|------|
| 飼主・ペット検索 | `docs/spec/screens/03-owners-list.md`、`backend/internal/pet/repository.go` `applyPetListSearch` |
| 過去カルテ UI | `frontend/src/features/medical-records/components/InterviewHistory.tsx`、`frontend/src/hooks/use-medical-records.ts` `useGetPetMedicalHistory` |
| 移行21表 | `backend/internal/csvimport/cutover_contract.go` `CutoverTableSpecs`、`docs/ops/deploy/CLINIC_CSV_IMPORT.md` |
| ワクチン種 | `old_db/sql/migration/030_stage.sql` vaccines.species = NULL |
| 性別コード | `old_db/docs/migration/propose-review-decisions.md` `PetSeibt_Kbn`、`old_db/sql/migration/030_stage.sql` pets.gender CASE |
| 未納 | `docs/spec/screens/30-unpaid-list.md`、`old_db/sql/migration/030_stage.sql` billings.status |
| 保険割合 | `frontend/src/features/accounting/components/InsuranceCard.tsx`、`docs/spec/screens/11-accounting-detail.md` |
| デモ退役 | `docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md`（`003_demo` 退役。臨床データは handoff/import） |

---

## 1. 医院向け回答（送付用）

ご確認いただきありがとうございます。ご指摘の4点について、現状と対応を分けて回答します。  
検索・保険割合・過去カルテの導線は、**いまお使いの検証環境に反映済み**です。本番環境にはまだ出ていません。

**Q1. 名字＋ペット名での検索**

ご指摘のとおり、名字だけだと件数が多く探せませんでした。飼主・ペット一覧の検索欄はそのまま1つで、**空白で区切った語をすべて満たす行だけ**出すようにしました。検証環境でお試しいただけます。

- 「小林 ポチ」→ 飼主名側に小林、ペット名側にポチがある行に絞る
- 「小林」だけ → 従来どおり同姓がすべて出る
- 電話番号・ペット番号での検索は従来どおり使えます

**Q2. 過去カルテが参照できない／猫に犬用注射が出る**

**デモデータではありません。** 検証環境には、旧システムから移行した医院データが入っています。

過去カルテについて:

- カルテ本体、問診、診察・方針の文章は移行対象です。
- カルテ入力の右側は **問診（主訴）の抜粋**です。行を開くと、その日のカルテ詳細へ進めます（直近50件まで）。
- 旧システムの**処置・処方明細は、今回の移行対象に含まれていません。** 旧画面の注射・処置一覧が治療タブに無いことがあります。こちらは別途、移行範囲の判断が必要です。

猫に Proheart や6種混合が出る件は、ワクチンマスタに「犬／猫」の種別が移行されておらず、接種記録がペットにそのまま紐づいているためです。件数の切り分けはこれから行います。該当個体の例をいただければ、個別に追います。

**Q3. 避妊去勢済みが性別不明になる**

性別と避妊去勢は、新システムでは別項目です。

- 性別：雄／雌／不明
- 避妊去勢：手術日（別欄）

旧システムでは、性別と避妊去勢が1つの区分に入っていました。移行時に、去勢・避妊済みのコードを「不明」に落としているため、避妊去勢済みの個体が性別不明に見えています。避妊去勢の欄（手術日）は別経路で入っています。

方針は、**オス／メスは復元し、避妊去勢は今どおり別項目で持つ**ことです。まだデータ側の訂正前です。

**Q4. 会計画面**

**未納残高もデモ用のダミーではありません。**  
会計の未納は、支払待ち（未精算）の請求から集計しています。旧システムで未精算だったものが、移行後も未納として出ます。全件が実未納か、一部が移行上の未紐付けかは、件数を切り分けます。実運用上すでに回収済みのものがあれば、例をください。

保険の負担割合（保険会社が支払う割合）は、ご指摘を受けて **新規の選択肢を 50% と 70% のみ**にしました（検証環境に反映済み）。すでに 90% / 100% で保存されている会計は、その割合を消して 50% に置き換えません。マスタの補償率からの自動セットはありません。50% と 70% 以外に必要な割合があれば、その値をください。

---

## 2. 判断原則（本フィードバックへの適用）

`docs/product-philosophy.md` の順（疑う → 削除 → 簡素化 → 速度 → 自動化）。

| 問 | 残す理由 | 削除・しないこと |
|----|----------|------------------|
| Q1 | 同姓400件超は個体特定のサイクルを止める。検索欄を増やすのではなく、既存1欄の意味を「語をすべて満たす」へ変える | 第2検索ボックス、種別名検索の復活 |
| Q2 | 「過去のカルテ」と書いて問診抜粋しか出さないのは工程を増やす。パネルを正直にするか、既存カルテ詳細へ通す | 新規の巨大ビューア、処置未移行をUIだけで偽装 |
| Q3 | 性別と避妊去勢は既に別項目。壊れているのは移行のコード解釈 | 性別enumに「去勢オス」を足して二重管理しない |
| Q4 保険 | 使わない90%/100%は選択肢から消す | 保険マスタ自動連携の新規実装（仕様上未実装のまま） |
| Q4 未納 | 未納一覧の定義自体は業務機能 | 未納をデモ扱いして一括消去しない |

医院コメントの責任者名は本ファイル作成時点で未記載。送付・改修着手時に受付／会計の個人名を付ける。

---

## 3. 修正タスク一覧

優先は現場で毎日止まる順。P0はデータ誤りまたは使えない検索。P1は画面の約束と実データのギャップ。P2は移行範囲の拡張（PO承認が先）。

| ID | 問 | 優先 | 種別 | 状態 |
|----|----|------|------|------|
| UAT-Q1-SEARCH-AND | Q1 | P0 | アプリ（BE、仕様、テスト） | STG デプロイ済み（PR #411）。ブラウザ未確認。production 未反映 |
| UAT-Q4-INSURANCE-RATES | Q4 | P0 | アプリ（FE、仕様） | STG デプロイ済み（PR #411）。ブラウザ未確認。production 未反映 |
| UAT-Q2-HISTORY-NAV | Q2 | P1 | アプリ（FE） | STG デプロイ済み（PR #411）。ブラウザ未確認。production 未反映 |
| UAT-Q3-GENDER-MAP | Q3 | P0 | 移行SQL＋既存データの訂正 | 未着手。old_db + STG 運用承認 |
| UAT-Q2-VACCINE-SPECIES | Q2 | P1 | 調査→マスタ種 | 未着手。STG 件数調査が先 |
| UAT-Q4-UNPAID-TRIAGE | Q4 | P1 | 調査（集計のみ） | 未着手。STG read 承認 |
| UAT-Q2-TREATMENTS-IMPORT | Q2 | P2 | 移行21表の外 | deferred。PO 未決 |

---

### UAT-Q1-SEARCH-AND: 空白区切りAND検索（名字＋ペット名）

- **問題**: 飼主・ペット一覧の `search` は飼主名・ペット名などを **OR** 部分一致する。同姓（小林・渡辺等）が400件超になり個体を特定できない。医院は「名字＋ペット名」を1回の検索で使いたい。
- **根拠**: `applyPetListSearch` は `pets.name` / `owners.name` / 電話 / `owners.id` / `pets.pet_number` を単一パターンで OR。空白除去は **飼主フルネーム**（姓と名の空白差）用であり、飼主名とペット名の連結検索ではない。`docs/spec/screens/03-owners-list.md` もその定義。
- **修正方針**: 既存の1欄を維持する。空白（半角・全角・連続）で分割した **各語が、既存対象フィールドのいずれかにヒットすること（AND）** を追加する。1語のときは現行ORのまま（「小林」は同姓全員、「ポチ」はペット名、「090-…」は電話）。2語「小林 ポチ」は、一方が飼主名側・他方がペット名側でも、両語が同一行で満たされればヒット。飼主フルネーム無空白一致（現行 compact）は1語扱いに残す。カルテ一覧検索はこの単位に含めない（医院の問は動物名検索）。
- **受け入れ条件**:
  1. 飼主「小林太郎」＋ペット「ポチ」、別行で飼主「小林」＋ペット「シロ」があるとき、`search=小林 ポチ` はポチ行のみ。`search=小林` は両行。
  2. `search=ポチ`・電話・ペット番号・空白のみ0件は現行テストと同じ。
  3. clinic 外の同名はヒットしない（既存 isolation テスト維持）。
  4. 仕様 `03-owners-list.md` に「複数語は AND」を追記。
- **検証**: `docker compose exec backend go test ./internal/pet/...`（`repository_test.go` / `repository_space_search_test.go` / `repository_kana_search_test.go` を拡張）。全件 `go test ./...` は自動実行しない。
- **やらないこと**: 新しい検索API、クライアント全件取得、種別名の検索対象化。
- **状態**: STG デプロイ済み（PR #411、`d337f016`）。ブラウザ未確認。production 未反映。

---

### UAT-Q3-GENDER-MAP: 旧性別コード 3/4 を雄/雌へ直す

- **問題**: 避妊去勢済み個体が性別「不明」。医院はオス/メスを残したい。避妊去勢欄はある。
- **根拠**: 承認済みデコードは `PetSeibt_Kbn` `{1,3}=male, {2,4}=female, {5,0}=unknown`。去勢事実はコードから `neutered_date` を捏造しない（`propose-review-decisions.md`）。一方 stage SQL は `sex_kbn IN ('1','01')→male`、`('2','02')→female`、**else unknown**。コード3/4が unknown になる。`neutered_date` は `PetOpe_Date` から別列。AE の性別は `male|female|unknown`、UI は雄/雌/不明。去勢日は別フィールド。
- **修正方針**:
  1. `old_db/sql/migration/030_stage.sql` の gender CASE を承認デコードに合わせる（1/01/3/03→male、2/02/4/04→female、それ以外 unknown）。
  2. 対応する stage テストがあれば更新。
  3. **既に STG へ入った行**は SQL 再実行または `pets.gender` の訂正バッチ。再取込は `make stg-uat-*` の運用承認が必要。エージェントは migrate/STG 書き込みを自動実行しない。
  4. AE の性別enumは増やさない。画面は「性別」＋「去勢・避妊手術日」の併記を維持。一覧に性別列は仕様上無い（`03-owners-list.md`）。詳細フォームで確認できることを受入にする。
- **受け入れ条件**:
  1. 新規 stage で sex_kbn=3 のペットは `gender=male`、4 は `female`。5 と 0 は `unknown`。
  2. `neutered_date` はコード3/4からは作らない（日付ソースは従来どおり `PetOpe_Date`）。
  3. 飼主詳細のペット編集で、去勢日がある個体が雄または雌で表示される（不明に落ちない）。
- **検証**: old_db の stage SQL テスト（当該 repo）。AE 側は性別表示の既存 transform テストで十分（enum 変更なし）。STG 訂正後の件数は集計のみ（PHIなし: gender×neutered_date IS NOT NULL のクロス集計）。
- **やらないこと**: 性別に「去勢済オス」を足す。去勢日未入力をコード3/4から捏造する。
- **状態**: マッピング承認と SQL の不一致は文書上確定。STG 適用は運用ゲート。

---

### UAT-Q4-INSURANCE-RATES: 会計の負担割合から 90%/100% を外す

- **問題**: 医院「保険の負担割合は90%と100%は無い」。
- **根拠**: `InsuranceCard` の選択肢は 50/70/90/100。ラベルは「負担割合（保険会社が支払う割合）」。既定は `use-accounting-detail-state.ts` の `"0.5"`。保険マスタ `coverage_rate` からは自動セットしない（`11-accounting-detail.md` / `master-insurance.md`）。操作マニュアル草稿 `29-master-insurance.md` は 50/70/100 と食い違う。
- **修正方針**: 会計詳細の Select を **50% と 70% のみ**にする（医院が他割合を指定したらそれに合わせる）。保存済み `insurance_ratio` が 0.9 / 1.0 の会計は、選択肢に無い値でも表示を落とさず、編集時は値を維持するか明示的に 50/70 へ変更させる（サイレント丸め禁止）。仕様 `11-accounting-detail.md` を選択肢に合わせて直す。マスタ補償率の自動連携はこの単位に含めない。
- **受け入れ条件**:
  1. 新規会計で保険ONのとき、Select に 90% / 100% が無い。50% / 70% はある。
  2. 既定は 50%。
  3. 既存で 90% が入っている詳細を開いても計算に使った割合が分かる（消えて 50% に置き換わらない）。
- **検証**: `docker compose exec frontend npx vitest run src/features/accounting`（InsuranceCard / use-accounting-detail-state）。ブラウザ確認は会計新規の保険スイッチ。
- **やらないこと**: 保険マスタ画面の補償率 0–100 整数を会計 Select に結合する（未実装のまま）。
- **状態**: STG デプロイ済み（PR #411）。新規 50/70、既存 0.9/1.0 は Select に残す。ブラウザ未確認。production 未反映。

---

### UAT-Q2-HISTORY-NAV: 「過去のカルテ」を実際に参照できる状態にする

- **問題**: 医院「過去カルテが参照できない」。カルテフォーム右の見出しは「過去のカルテ」だが、中身は問診主訴の抜粋（最大50件、page=1）。クリックは展開のみ。`引用` ボタンは `stopPropagation` だけで引用しない。処置明細は21表外。
- **根拠**: `InterviewHistory.tsx`（見出し・展開・引用no-op）、`useGetPetMedicalHistory`（`GET /v1/medical-records?pet_id=&limit=50&page=1`、`inquiry.chief_complaint`）。一覧は `05-medical-records-list.md`。21表に `treatments` なし。
- **修正方針（簡素化）**: 新ビューアを足さない。
  1. パネル各行から **同一ペットのそのカルテ詳細**（`/medical-records/:id`）へ遷移する。
  2. 見出しまたは空状態で、ここが問診抜粋であること、全文はカルテ詳細であることを書く。
  3. `引用` が未配線なら、動かないボタンは削除する（存在しない機能を残さない）。引用を残すなら、主訴テキストを当日問診へ入れる実装を同じ単位で完了させる。中途のボタン禁止。
  4. 50件上限は、詳細への導線があればこの単位では維持してよい。全件無限スクロールは範囲外。
- **受け入れ条件**:
  1. 過去行を開くと、その `id` のカルテ詳細が表示される（問診・診察方針があるレコードでは本文が見える）。
  2. 処置が無い移行カルテでは治療タブが空でも、詳細自体は開ける。
  3. 引用ボタンを残す場合: クリックで当日問診に主訴が入る。残さない場合: ボタンが無い。
- **検証**: `npx vitest run src/features/medical-records/components/InterviewHistory.test.tsx` および Form の履歴導線テスト。
- **やらないこと**: この単位で treatments を移行する。デモデータのせいにするコピー。
- **状態**: STG デプロイ済み（PR #411）。見出し「問診抜粋」、詳細へ Link、引用ボタン削除。ブラウザ未確認。production 未反映。

---

### UAT-Q2-VACCINE-SPECIES: 猫に犬用ワクチンが出る件

- **問題**: 猫の履歴に Proheart・6種混合など犬用接種が載る。
- **根拠**: stage の `vaccines.species` はソース列なしで **常に NULL**。`vaccinations` は pet_id＋vaccine_id で載せる。AE のワクチンマスタは `dog|cat|both` を持てるが、移行行は種なし。接種一覧はペット種でワクチンを除外しない。
- **修正方針**:
  1. **先に調査（PHIなし）**: 猫（species=猫）に紐づくワクチン名の件数上位。旧マスタ名のままか、vaccine_id 取り違えかを見る。個体名は出さない。
  2. 調査後: ワクチンマスタへ種を付ける（名称ルールまたは医院確認リスト）。**新規接種の選択肢**はペット種で `dog/cat/both` を絞る（既存のマスタ API `species` クエリを使う）。
  3. **過去履歴の非表示は別判断**。旧システムに本当にその接種があれば消さない。取り違えなら vaccine_id 訂正。調査前に猫の Proheart を一括削除しない。
- **受け入れ条件（調査単位）**: 猫×犬用ワクチン名の件数表（院内限り、名前なし）。Class A（マスタ種欠損）/ B（行の取り違え）/ C（旧データのまま）を文書化。
- **受け入れ条件（実装単位・調査後）**: 猫の新規接種UIに Proheart / 犬6種が選択肢に出ない。`both` は出る。履歴の扱いが調査メモの方針と一致する。
- **やらないこと**: ソース列が無いのに species を推測で埋める。調査なしの一括 DELETE。
- **状態**: 医院例（Proheart、6種）は手がかり。件数未計測。実装は調査の次。

---

### UAT-Q4-UNPAID-TRIAGE: 未納の切り分け（消さない）

- **問題**: 医院「未納残高が入っているのはデモ版か」。
- **根拠**: デモ臨床シードは退役。未納は waiting 全額＋クレジット訂正残（`30-unpaid-list.md` / `S08`）。stage は `settle_flg IN ('0','00')` を `waiting`、支払突合できたものを `completed`。窓口0精算は completed へ寄せた記録がある。残 waiting は旧未精算または突合失敗の可能性がある。
- **修正方針**: アプリで未納定義を変えない。STG で **集計のみ**: status 別件数、waiting かつ payment 有無、金額帯。PHI・飼主名は出さない。結果を医院に「旧未精算の移行」と説明し、回収済みの例があればその billing だけ個別確認。一括 completed 化は禁止。
- **受け入れ条件**: 集計表（件数・合計金額、status 内訳）。「デモではない」を集計で裏付ける。個別訂正ルールが必要なら別単位。
- **やらないこと**: 未納タブの隠蔽、waiting の一括完了。
- **状態**: 読み取り調査。STG 接続は運用承認。

---

### UAT-Q2-TREATMENTS-IMPORT: 処置・処方の移行（今期外候補）

- **問題**: 旧カルテの注射・処置が一覧に無いと「カルテが空／参照できない」に見える。
- **根拠**: `CutoverTableSpecs` に `treatments` も `prescriptions` も無い。`030_stage.sql` に treatments INSERT は無い。
- **修正方針**: 21表の契約変更＋old_db stage＋F6。PO が「処置明細を移行する」と決めるまで実装しない。決まるまでは Q2-HISTORY-NAV で詳細へ通し、治療タブ空は移行範囲と明示する。
- **受け入れ条件**: 未設定（対象外のまま）。
- **状態**: 今期外候補。本ファイルでは追跡するが READY にしない。`phase2-deferred.md` へは、再開条件・named owner・Linear URL が揃ってから移す。

---

## 4. 推奨実施順（残）

AE のみの3単位は STG デプロイ済み（§6、PR #411）。production 未反映。残る順:

1. STG 画面確認（飼主一覧「名字 ペット名」、会計保険 50/70、カルテ右パネル→詳細）
2. **UAT-Q3-GENDER-MAP**（old_db SQL＋STG 訂正。書き込みは人間）
3. **UAT-Q4-UNPAID-TRIAGE** と **UAT-Q2-VACCINE-SPECIES** 調査（並列可、read-only）
4. ワクチン種の実装は 3 の結論後
5. 処置移行は PO 決定後

---

## 5. 本ファイルの位置づけ

- 医院回答の保管と、修正単位の分解。
- **実行 SoT ではない。** 着手する単位は承認後 `todo-issue.md` に短い節で移す。
- Linear 新規作成はしない。既存ハブ [BRT-4](https://linear.app/baritechllc/issue/BRT-4) へのコメントは明示承認後。
- 秘密・個体名・行値は書かない。

---

## 6. キャンペーン結果（2026-09-15）

キャンペーン `stg-uat-clinic-feedback-20260915`（`.planning/agent-fast-campaign/`、gitignore）。`outcome=COMPLETE`、`fixed_agent_target_set` は Q1 / Q4保険 / Q2履歴の3本。

local `main` 統合（この tree）:

- cherry-pick: `8c590134a` / `a49ef4e48` / `86a5c436d`
- merge: `f622f8312` feat/UAT-Q1-SEARCH-AND、`a1267bdc4` feat/UAT-Q4-INSURANCE-RATES、`e6eff68a0` feat/UAT-Q2-HISTORY-NAV
- AI claim 3本は統合後に `git branch -d` 済み
- `feat/UAT-Q1-SEARCH-AND` / `feat/UAT-Q4-INSURANCE-RATES` / `feat/UAT-Q2-HISTORY-NAV` と対応 worktree は、この時点でリポジトリに無い（統合済みのため再作成しない）
- `feat/R09-C-PRIME-CSVIMPORT-ENTERED-BY-FK` は別作業。削除しない
- 2026-09-15 STG: PR [#411](https://github.com/MinoruSoga/AnimalEkarte/pull/411) を `staging` に merge（`d337f016`）。[Backend Deploy #176](https://github.com/MinoruSoga/AnimalEkarte/actions/runs/34923018516) success（Worker、migrate、`/health`、CRUD smoke）。[Frontend Deploy #51](https://github.com/MinoruSoga/AnimalEkarte/actions/runs/34923018544) success。production 未反映。Linear なし

| ID | Mode 3 | 照合 |
|----|--------|------|
| UAT-Q1-SEARCH-AND | COMPLETE | `applyPetListSearch` が語ごと AND。仕様「複数語は AND」 |
| UAT-Q4-INSURANCE-RATES | COMPLETE | 新規 50/70。0.9/1.0 は legacy Select |
| UAT-Q2-HISTORY-NAV | COMPLETE | `Link` → カルテ詳細。見出し「問診抜粋」。引用なし |

検証:

- キャンペーン: isolated pet test / accounting 290 / medical-records 469
- 統合後 compose: `go test ./internal/pet/...` PASS、`InsuranceCard` 4 + `InterviewHistory` 9 PASS
- ブラウザ: SKIP（ログインが必要）

残差: Q3 gender map、ワクチン種調査、未納 STG 集計、処置移行。
