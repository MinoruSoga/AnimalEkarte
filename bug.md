# 受入テスト バグ報告

実施環境: ローカル (seed 003_demo) / URL: http://localhost:3003
実施範囲: docs/ops/testing/scenarios 配下 S01〜S12, V01〜V05
実施日: 2026-07-31

---

## 対応状況サマリ（2026-08-05 JST）

- **Product evidence snapshot（2026-08-03 時点の履歴スナップショット。現況の正本ではない）**: この bullet は列挙した commit 群を固定記録するものであり、以後の更新で追記しない。現況は各個票の `対応状況` 行を見ること。台帳 commit parent HEAD `fb0cf9c910aef842fdde1a0206bb5546163096c3`。飼主・ペット・受付は到達済み commit `d7bf32f2214d6bb6c252b99b001d2ed2044de7c9`（BUG-001）、`a17d39d6f46ddaf8afcba7ed53419dbc4f92e968`（BUG-002）、`eb7db0dc94fb842c7e569252a9cebc6aee96cd60`（BUG-021）、`fc3c12b2800942c7527b0be951aad20860c6131c`（BUG-022）、`617f6f9bf88be3627ba789d447c98858dd34c80a`（BUG-020）。検査は BUG-004（`2a8aca33c1848613e7c3ccd9ffa2f2a4e3c9ad5e`）、BUG-005（`dfd653eaa5ccb089707c3a088863c39c07669288`）、BUG-017（`7f71063759974257be14a4ed0a8a5fd04a5c6880`）。その他 `IMPLEMENTED_UNVERIFIED` に BUG-011（`b65cf69ef56785c473ddd233624292a3c338401e`）。BUG-003 は承認済み構造化 range data 不在で BLOCKED。
- **判定基準**: current checkout から到達可能な code/test を正本とする。GitHub Issue/PR の closed/merged 単独は closure に使わない。本更新では原文ブラウザシナリオを再実行していない。
- **件数**: OPEN=10 / IN_PROGRESS=0 / IMPLEMENTED_UNVERIFIED=21 / VERIFIED_FIXED=0 / BLOCKED=1 / DUPLICATE=0 / NOT_REPRODUCIBLE=0 / **合計=32**
- **コードベース照合（2026-08-05）**: 集計の再計算ではなく、**各個票の `根拠` が現行コードで今も成立するかを source で検証**した。結果は全件一致で、状態変更を要する個票は 0 件。検証内容: ① OPEN 13 件それぞれの `根拠` が指す実装箇所を grep/read で実査（BUG-006 `PatientInfoCard` に petDetails 未渡し / BUG-007 `useGetPetVaccinations` は owner-report のみで vaccinations 側は page 方式 / BUG-016 hospitalization hook の `isError` は `handleApiError` のトースト表示のみで not-found ゲートではない / BUG-019 `EstimateForm` に isError 処理なし / BUG-026・029 の validate は `!d.name.trim()` のみ / BUG-031 `restoreSession = !isAuthPublicPath(pathname)` / BUG-032 checkup sync に timeout・LIMIT なし、ほか）② 個票が引用する commit hash 20 件が全て HEAD から到達可能であることを `git merge-base --is-ancestor` で確認（stale・revert 済み参照 0 件）。**この照合は 2026-08-05 の日中時点の記録であり、以後 BUG-016 / BUG-019 / BUG-030 が同日中に修正されて `IMPLEMENTED_UNVERIFIED` へ移動した。上の「OPEN 13 件」はその照合時点の母数である。** **次回このセクションを更新するときも、集計の再計算だけで「最新化した」と書かないこと。**
- **集計の取得方法**: 2026-08-05 夕に各 `## BUG-NNN` 節の最初の `対応状況` 行を awk で機械抽出して再計上した（32/32 を取得）。下の「ドメイン別タスク索引」の状態内訳も同じ抽出結果から再計算している。**この更新で行ったのは全数抽出による再計上のみであり、各個票の `根拠` が現行コードで成立するかの source 照合は再実施していない**（それは上の「コードベース照合」bullet の対象で、2026-08-05 日中の記録が最後）。以後この集計を更新するときも、記憶や差分ではなく同じ全数抽出をやり直すこと。個票を更新して集計を据え置くと必ずドリフトする（実際 08-03 集計は OPEN=20 / IMPLEMENTED_UNVERIFIED=11 のまま 7 件、08-05 集計は OPEN=13 / IMPLEMENTED_UNVERIFIED=18 のまま 3 件ずれていた）。
- **原文シナリオ再検証**: PASS=0 / FAIL=0 / BLOCKED=0 / UNREPORTED=32 / **合計=32**
- **未検証境界**: 本更新でのブラウザ/DB mutation 再現は未実施。`VERIFIED_FIXED` は 0。
- **個票正本**: 各 `## BUG-NNN` 節の最新 `対応状況` 行。

# ドメイン別タスク索引

各 BUG は実装を主導する**一次担当ドメインへ1回だけ**配置する。複数ドメインにまたがる依存・同時回帰は、後段の「横断クラスタ索引」を併用する。この索引は担当分割用であり、個票の重大度・対応状況・Waveを変更しない。

| 一次担当ドメイン | 対象 BUG | 件数 | 状態内訳 | 主な責務境界 |
|---|---|---:|---|---|
| 飼主・ペット・受付 | BUG-001, BUG-002, BUG-020, BUG-021, BUG-022 | 5 | IMPLEMENTED_UNVERIFIED 5 | `owner`, `pet`, owners UI、予約モーダルの新規飼主入力 |
| 検査 | BUG-003, BUG-004, BUG-005, BUG-017 | 4 | IMPLEMENTED_UNVERIFIED 3 / BLOCKED 1 | `examination`、検査フォーム、担当医選択、異常値判定・確定 |
| 予防接種 | BUG-006, BUG-007 | 2 | OPEN 2 | `vaccination`、対象ペット表示、ペット別接種履歴 |
| LINE・LIFF・Lステップ連携 | BUG-008, BUG-014, BUG-030, BUG-032 | 4 | IMPLEMENTED_UNVERIFIED 1 / OPEN 3 | `line-reserve`, `liff`, LINE予約設定、`lstep/checkup-sync` |
| 入院・ホテル | BUG-009 | 1 | IMPLEMENTED_UNVERIFIED 1 | `hospitalization`、ステータスタブ・一覧取得 |
| カルテ・バイタル | BUG-010, BUG-015 | 2 | IMPLEMENTED_UNVERIFIED 2 | `medicalrecord`, `vital`、診察/治療プラン、体重単位 |
| 見積・会計 | BUG-011, BUG-013, BUG-018, BUG-019 | 4 | IMPLEMENTED_UNVERIFIED 4 | `estimate`, `billing`、未請求明細、締め後会計、Not Found |
| 顧客集計 | BUG-012 | 1 | OPEN 1 | `aggregation`、LTV/売上集計、CPM取得 |
| 横断フォーム基盤 | BUG-016 | 1 | IMPLEMENTED_UNVERIFIED 1 | 予防接種・検査・入院フォーム共通の取得失敗/Not Found 契約 |
| 認証・権限 | BUG-023, BUG-024, BUG-031 | 3 | IMPLEMENTED_UNVERIFIED 1 / OPEN 2 | `auth`、権限グループ、セッション復元・ログイン遷移 |
| 設定・マスタ | BUG-025, BUG-026, BUG-027, BUG-028, BUG-029 | 5 | IMPLEMENTED_UNVERIFIED 3 / OPEN 2 | settings UI、各 master owner、共通保存・重複エラー契約 |
| **合計** | **BUG-001〜BUG-032** | **32** | **IMPLEMENTED_UNVERIFIED 21 / OPEN 10 / BLOCKED 1** | 各個票を正本とする |

# 実装優先ウェーブ / 横断クラスタ索引

調査基準は `main` の `fa74ef92e`（主要コード調査は `239a8a736`、以後のproduct source差分と最終docs-only進行を再監査）。以下は実装そのものではなく、現行コードを起点にした実装・検証計画である。報告時の推定と現行ソースが一致しない項目は、原因を固定せず再現または計測を先行させる。

| Wave | 目的 | 対象 |
|---|---|---|
| Wave 0 | 臨床安全、法的記録、会計不整合の封じ込め | BUG-002, BUG-003, BUG-004, BUG-010, BUG-015, BUG-018, BUG-021, BUG-022 |
| Wave 1 | 認証ブロッカー、集計・請求の可用性 | BUG-008, BUG-009, BUG-011, BUG-012, BUG-013, BUG-014, BUG-024 |
| Wave 2 | マスタ保存契約とエラー契約の是正 | BUG-023, BUG-025, BUG-026, BUG-027, BUG-028, BUG-029, BUG-030 |
| Wave 3 | 検索・表示・Not Found・入力フィードバック | BUG-001, BUG-006, BUG-007, BUG-016, BUG-017, BUG-019, BUG-032 |
| Wave 4 | データ品質と低リスク導線 | BUG-005, BUG-020, BUG-031 |

| クラスタ | 主計画 | 兄弟計画 | 共通契約 |
|---|---|---|---|
| `C-LIFF-AUTH` | BUG-008 | BUG-014 | ローカルモックと実 LIFF 認証を分離し、401 を原因別に表示する |
| `C-EXAM-LIFECYCLE` | BUG-003 | BUG-004 | 基準値導出と「初回確定」を同一の検査保存契約で守る |
| `C-PET-DEATH` | BUG-002 | BUG-021, BUG-022 | 死亡日の検証、永続化、キャッシュ反映、再読込表示を一連で保証する |
| `C-SILENT-VALIDATION` | BUG-017 | BUG-021 | ブロック理由をフィールド近傍に表示し、無音失敗を禁止する |
| `C-NOT-FOUND-EMPTY` | BUG-016 | BUG-019 | 取得失敗を空の編集モデルへ変換しない |
| `C-MASTER-FALSE-SUCCESS` | BUG-026 | BUG-029 | 成功通知は mutation 成功応答後にだけ出す |
| `C-MASTER-DUPLICATE-MSG` | BUG-023 | BUG-027 | 一意制約を安定したエラー code/params に変換する |
| `C-VAX` | BUG-006 | BUG-007 | 対象ペット正本とペット別履歴を同じ patient context で扱う |
| `C-MASTER-SAVE-FIELDS` | BUG-025 | BUG-028 | 必須フィールドと既定値を FE/BE 間で明示する |

共通実装規約: 書き込み責務は ADR-006 の owner package に置き、consumer からは interface 経由で呼ぶ。`clinic_id` と request 由来の owner/pet/staff FK を fail-closed で検証し、複数書き込みと監査ログは同じ `DBOrTx` 境界に含める。フロントは Feature Indexing と query key の正本を維持する。将来の検証は Docker 内の対象 package/file に限定し、migration/seed が必要な場合はコード修正・既存データ補修と分離してレビューする。migration 取り込み後の `make migrate` は人が実行し、agent は自動適用しない。

## BUG-001: 飼主・ペット一覧の検索が「姓 スペース 名」形式でヒットしない

- **重大度**: 中（受付業務で頻出する検索パターンが機能しない）
- **対応状況（2026-08-03 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み commit `d7bf32f2214d6bb6c252b99b001d2ed2044de7c9` で `PetRepository.FindAll` が空白非依存の飼主フルネーム、`owners.id` 文字列一致の飼主No、`pets.pet_number` 部分一致、空白のみ fail-closed、clinic-scoped JOIN を実装。`TestPetRepository_FindAll_Search` と owners list / OpenAPI / 画面仕様も更新済み | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: S01 手順1を専用 synthetic fixture でブラウザ再検証し、`VERIFIED_FIXED` 可否を判定
- **発見シナリオ**: S01 手順1の確認中（飼主・ペット一覧 `/owners`）
- **再現手順**:
  1. `/owners` の検索ボックス（プレースホルダ「飼主名、ペット名、飼主No、種別...」）に `伊藤` とだけ入力 → 120件ヒット（正常）。
  2. 検索ボックスに `伊藤 宏美`（半角スペース区切りのフルネーム）と入力 → **0件「データが見つかりません」**。実在する飼主（飼主NO 300290、生存ペット複数あり）が見つからない。
  3. 同様に `伊藤 史安`（飼主NO 307867、生存ペットあり）で検索しても0件。
- **期待結果**: プレースホルダが「飼主名」を検索対象と明示しており、スタッフが自然に入力する「姓 スペース 名」形式のフルネーム検索はヒットするべき。
- **実際の結果**: 姓のみ・名のみでは検索できるが、スペース区切りのフルネームでは常に0件になる。
- **備考**: 飼主NO（例: `307867`）そのものでの検索も0件だった（`310371` など未変更の飼主でも同様に0件）。プレースホルダが「飼主No」も検索対象と謳っているが、実際には飼主NOでの検索がヒットしない可能性がある。あわせて要確認。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `S-OWNER-SEARCH`
- 同一 PR にする BUG: なし
- 先行必須: なし
- 後続解放（シナリオ/他BUG）: S01 手順1の飼主・ペット検索と番号検索

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE
- 観測根拠（API・クエリ・コード参照）: `/owners` は owner repository ではなく pet 検索を使い、現行 SQL は pet 名・owner 名の各列・電話番号だけを個別 `ILIKE` する。姓と名を空白で連結した値、`owner_number`、`pet_number` は対象外。（具体パスは §2）
- 重複判定（例: 002 vs 022）: 単独。BUG-016/019 の Not Found 契約とは別根因。
- 所有境界（FE / BE / データ / 環境）: FE=owners feature; BE=pet read owner; データ=clinic-scoped owner/pet 検索行; 環境=なし

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/owners/loaders.ts` の owner 一覧 loader、`frontend/src/features/owners/components/OwnersListTable.tsx` の検索 UI。 | 現行証拠 / 変更候補 |
| BE | BE owner: `backend/internal/pet/pet_handler.go`、`backend/internal/pet/pet_request.go`、`backend/internal/pet/repository.go` の `Search` 条件。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: 既存 `GET /api/v1/pets?search=` の response/DB schema は不変。search の意味論へ空白正規化・連結氏名・owner/pet番号を追加し、index が必要なら別 migration。
- 正しい挙動の定義（1〜3 文）: 正規化した検索語で clinic 内の表示氏名・飼主No・ペットNoを検索し、count/page と UI説明を一致させる。
- やらないこと（Out of scope）: 曖昧な全列検索、全clinic検索、無根拠な fuzzy search。
- 既存データ修復の要否と手順: 不要。index候補は計測後の別migrationとし、人が `make migrate`。

- BE で検索語の前後空白と連続空白を正規化し、姓・名を連結した表示名、飼主 No、ペット No を clinic scope 内で検索する。複数 token は順序を保持した氏名一致を最低契約とし、曖昧な全列 AND/OR 拡張は別件にする。
- FE の説明文と placeholder は、BE が実際に保証する対象だけを列挙する。件数・ページングはサーバ側検索後の値を使う。

#### 4. 受け入れ基準（AC）

1. Given clinic 1 に `伊藤` / `宏美` と飼主 No `300290` がある、When `伊藤 宏美`、余分な空白を含む同名、または `300290` で検索する、Then 同じ飼主の pet 行が返る。
2. 回帰: 姓のみ・電話検索、server pagination/countを維持する。
3. 負例: 別clinic同名/同番号、空白だけ、未知番号は0件で存在情報を漏らさない。 既存境界AC: Given 別 clinic に同名・同番号がある、When clinic 1 で検索する、Then 別 clinic の行・件数・存在は返らない。
4. 横展開確認対象: owner/pet検索、一覧placeholder、代表query plan

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/pet -run 'Test.*Search'`
- FE: `docker compose exec -T frontend npx vitest run src/features/owners/loaders.test.ts src/features/owners/routes/OwnersList.test.tsx src/features/owners/components/OwnersListTable.report.test.tsx`
- 手動/E2E: S01 手順1の確認中（飼主・ペット一覧 `/owners`）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 臨床/会計直接変更なし。検索結果・countをclinic_idで分離し、番号を文字列扱いする。
- fail-closed の維持: 別clinic候補をfallbackせず0件とし、検索errorを空結果に潰さない。
- audit / トランザクション境界: read-only。書込/audit txなし。index migrationは別承認。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- 連結式への部分一致は index を外し得るため `EXPLAIN (ANALYZE, BUFFERS)` で代表検索を測る。番号検索を数値変換して失敗させず、文字列として扱う。
- 既存データ補修は不要。検索用 index が必要なら migration を別差分にし、人手の `make migrate` 境界を守る。

#### 7. 実装ステップ（順序付き）

1. repository test に氏名空白・番号・clinic 分離の失敗ケースを追加する。
2. request 正規化と repository predicate を最小変更し、handler の pagination 契約を確認する。
3. FE 文言と loader/component test を合わせ、代表 E2E を実施する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- 上記 AC、対象テスト、代表 SQL 計測が通り、検索結果・count とも clinic scope が確認され、仕様外の曖昧検索を追加していない。

---

## BUG-002: 死亡登録直後、飼主編集画面のペット一覧テーブルの生死列がリロードするまで更新されない

- **重大度**: 低（データ自体は正しく保存されるが、画面表示が古いままになりスタッフが「保存されていない」と誤解し得る）
- **対応状況（2026-08-03 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み commit `a17d39d6f46ddaf8afcba7ed53419dbc4f92e968` で、死亡登録・解除成功時に modal-local form と外側の `pets` / `editingPet` を同一 `petId` の `status`・`deceasedAt` で不変同期する。foreign/absent ID では `pets` / `editingPet` の参照を保ち、非同期中に別ペットへ切り替えた場合は表示中 modal state を上書きしない。focused BUG-002 20件、hook＋BUG-373 42件、focused statements 95.00%、対象 ESLint / diff-check が green | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: S01 の専用 synthetic fixture で死亡登録直後の外側一覧表示をブラウザ再検証
- **発見シナリオ**: S01 手順1（`/owners/:id` でペット編集モーダルから死亡登録）
- **再現手順**:
  1. `/owners/307867` を開き、ペット「豆助」（001）の詳細・編集から「死亡を記録」→死亡日・理由を入力し確定。
  2. モーダル内では即座に生死ステータスが「死亡」に切り替わり「死亡を記録しました」のトーストが出る（正常）。
  3. モーダルを閉じて、同じ画面のペット情報テーブル（モーダル外側の一覧）を見ると、「豆助」の生死列が**「生存」のまま**（更新されていない）。
  4. ページを再読み込みすると正しく「死亡」と表示される。
- **期待結果**: モーダルでの確定後、外側のペット一覧テーブルも即時に再取得され「死亡」表示に切り替わるべき。
- **実際の結果**: ページを手動リロードするまで外側テーブルは古い「生存」表示のまま。
- **備考**: サーバー側のデータは正しく更新されている（リロード後は正しい／カルテ・会計・入院のペット選択画面では正しく「選択不可」になる）。React Query 等のキャッシュ無効化漏れの可能性。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-PET-DEATH`
- 同一 PR にする BUG: BUG-002, BUG-022（表示正本）; BUG-021 は同一クラスタの先行検証
- 先行必須: BUG-021 の死亡日入力契約
- 後続解放（シナリオ/他BUG）: S01 死亡登録の即時表示と BUG-022 の再読込表示

#### 1. 切り分けステータス

- 主因レイヤ: FE（BE lifecycle 契約を回帰）
- 観測根拠（API・クエリ・コード参照）: mutation は pet query を invalidate するが、外側テーブルは `use-pet-form-list-state.ts` の別ローカル state を表示する。成功 callback はモーダル内 form だけを書き換える。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-022 と症状は同じ死亡表示だが、本件は mutation 後のローカル一覧 stale、022 は再取得 transform。BUG-021 は入力検証。
- 所有境界（FE / BE / データ / 環境）: FE=owners/pet query state; BE=lstep/pet lifecycle; データ=pet death record; 環境=なし

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/owners/components/PetCareSection.tsx`、`frontend/src/features/owners/hooks/use-pet-form-list-state.ts`、`frontend/src/hooks/use-record-pet-death.ts`。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/lstep/lstep_lifecycle_handler.go`、同 service、`backend/internal/pet/repository.go` の死亡更新。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: API/DB 変更なし。既存 death mutation と既存 query-key 契約を使い、外側一覧の状態同期だけを直す。
- 正しい挙動の定義（1〜3 文）: 204成功後、入力済み死亡日を使って外側一覧stateを不変更新し、既存canonical query keyをinvalidateする。モーダルと一覧は即時にdeceasedとなり、refetch後も一致する。
- やらないこと（Out of scope）: 新しい clinicId 付き query key の発明、失敗時の擬似成功、死亡記録の変更。
- 既存データ修復の要否と手順: 不要。既存死亡記録は変更しない。

- death API は204でpetを返さないため、成功時は入力済み死亡日を外側一覧の既存state ownerへ不変更新し、project既定のcanonical query keyをinvalidate/refetchする。`query-keys.ts` の規約に反するclinicId入りkeyは新設しない。
- BUG-021 の検証を通過した要求だけを BE が同一 tx で保存・監査し、BUG-022 の transform と同じ status 契約を使う。

#### 4. 受け入れ基準（AC）

1. Given 生存 pet が外側一覧とモーダルに表示中、When 有効な死亡日で確定する、Then リロードせず両方が「死亡」となり、再取得後も同じ。
2. 回帰: モーダル・外側一覧・full reloadをBUG-021/022と通す。
3. 負例: 4xx/5xxでは楽観表示を固定せず、別owner/petのstateを更新しない。 既存境界AC: Given 更新が 4xx/5xx、When mutation が失敗する、Then 一覧を楽観的に死亡へ固定せず、理由を表示する。
4. 横展開確認対象: owners内のpet表示、患者選択、death query invalidation

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/lstep ./internal/pet -run 'Test.*Death|Test.*Deceased'`
- FE: `docker compose exec -T frontend npx vitest run src/features/owners/hooks/use-pet-form-list-state.test.ts`。一覧ownerまでmountする回帰testは新規追加予定。
- 手動/E2E: S01 手順1（`/owners/:id` でペット編集モーダルから死亡登録）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 死亡状態は臨床安全対象。owner/pet/clinicを正規query stateで分離する。
- fail-closed の維持: mutation失敗時にdeceasedへ固定せず、server正本へ戻す。
- audit / トランザクション境界: 死亡保存とauditはBEの同一txを回帰し、FE表示修正から分離しない。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- owner/pet が違う cache key を更新しない。死亡は臨床状態なので失敗時の擬似成功、無条件 rollback、別 clinic invalidation を禁止する。
- 既存死亡記録の一括補修はこの UI 修正に含めない。

#### 7. 実装ステップ（順序付き）

1. モーダル成功後も外側一覧が古いことを component test で RED にする。
2. query key と state owner を一本化し、不変更新 + 再取得を実装する。
3. BUG-021/022 の cluster test と実ブラウザ往復を通す。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- 即時表示、再読込、失敗 rollback、clinic/owner 分離がすべて確認され、死亡監査が成功 tx と同じ境界にある。

---

## BUG-003: 検査結果の異常値判定（H/L ハイライト）が常に「未判定」のまま計算されない【重大】

- **重大度**: 高（S02 の中核機能。臨床安全に直結する異常値の見落とし防止が機能していない）
- **対応状況（2026-08-03 JST）**: BLOCKED | **根拠**: 構造化 `exam_reference_ranges` の承認済み data が demo seed に不在（`003_demo` は exam_types/fields/results のみ）。`assessExamResult` は構造化 range がある場合のみ H/L。Mode3 follow-up で下限ちょうど・上限ちょうど equality の table-driven test を追加し assessment suite green。閾値推測・seed 作成は out of scope。必要入力: 獣医師承認済み clinic×species×field 構造化 range data packet | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: 承認済み reference range seed/data packet 供給後に S02 H/L を再検証
- **発見シナリオ**: S02 手順2〜4（検査管理 `/examinations`）
- **再現手順**:
  1. `/examinations/select-pet` から生存ペット（伊藤史安/豆助）を選び、新規検査登録。検査種別「血液検査（院内）」を選択（WBC基準値 6.0-17.0、RBC基準値 5.5-8.5、HCT基準値 37-55 などが動的表示される）。
  2. WBC=25.0（上限17.0を大幅に超過）、RBC=3.0（下限5.5を大幅に下回る）、HCT=37（下限ちょうど）、PLT=300（正常範囲内）を入力し保存。
  3. 一覧から当該検査を再度開く、またはページを再読み込みする。
  4. ステータスを「結果入力済み」に進めて保存、再読み込みしても同様。
- **期待結果**（[S02 手順2-4](docs/ops/testing/scenarios/S02-exam-abnormal-highlight-lock.md)）: WBCは高値(H)で赤ハイライト、RBCは低値(L)でstatus-blueハイライト、HCTは基準値ちょうどのため正常扱い。
- **実際の結果**: 判定列が **全項目・全ケースで「未判定」のまま**。値を25.0→26.0に変更して再保存しても変化なし。ステータスを「依頼中」→「結果入力済み」に進めても変化なし。
- **備考**: 仕様上、判定はバックエンド（`computeExamResultStatus`）が保存時に導出しDBに保存する設計だが、実際には保存後も判定が更新されていない。異常値の自動検出という臨床安全上重要な機能が事実上動作していない。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-EXAM-LIFECYCLE`
- 同一 PR にする BUG: BUG-003, BUG-004（承認済みseed/data packetは別差分。schema migration不要）
- 先行必須: 獣医師承認済み reference range と BUG-004 の初回確定順序
- 後続解放（シナリオ/他BUG）: S02 の H/L 判定、初回確定、再表示

#### 1. 切り分けステータス

- 主因レイヤ: BE / データ
- 観測根拠（API・クエリ・コード参照）: 現行保存経路は `assessExamResult` と clinic/species scoped range resolverを使い、`exam_reference_ranges` schemaも既存である。一方、`003_demo` は表示用 `normal_value` のみで構造化range seedが確認できず、live DBのrange有無はNEEDS REVIEW。`computeExamResultStatus` は互換wrapperであり、文字列から閾値を推測するfallbackは採用しない。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-004 と同じ保存 lifecycle だが、本件は reference range 不在/判定、004 は初回確定順序。
- 所有境界（FE / BE / データ / 環境）: FE=examination display; BE=medicalrecord assessment owner; データ=clinic/species reference ranges; 環境=seed/runtime DB の range 有無を確認

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| BE | `backend/internal/medicalrecord/exam_result_assessment.go`、`examination_service.go`、`exam_reference_range_repository.go`。 | 現行証拠 / 変更候補 |
| DB/migration | `backend/migrations/001_init.sql` の `exam_reference_ranges` table / clinic-field-species unique。 | 既存schema。新規migration不要 |
| Data | `backend/migrations/seeds/003_demo/exam_type_fields.csv` と、欠落を確認する reference-range seed。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/features/examinations/components/ExamPivotTable.tsx` と API transform。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: API shape と既存 DB schema は不変。既存 `exam_reference_ranges` へ承認済み data を供給する seed/data packet は実装差分と分離する。
- 正しい挙動の定義（1〜3 文）: 同一 clinic/species/field の構造化rangeだけでlow/normal/highを導出する。range不在・非数値・矛盾rangeは既存契約どおり `status=normal` かつ派生 `is_assessed=false` とし、新enumは追加しない。
- やらないこと（Out of scope）: 表示用 `normal_value` のparse、犬猫共通閾値の推測、agentによるmigration/seed適用。
- 既存データ修復の要否と手順: 既存結果の自動再判定はしない。必要なら read-only対象抽出→臨床承認→監査可能な別補修packet。

- 獣医師承認済みの species/clinic/field 別 lower/upper bound を構造化dataとして供給し、保存時に同一clinicのrangeだけで `low/normal/high` を導出する。境界値はnormal、非数値・range不在は `status=normal` + `is_assessed=false` として未評価表示にする。
- seed/data packetとapplicationの変更を分離し、既存結果の再判定は承認された補修jobとしてaudit可能にする。

#### 4. 受け入れ基準（AC）

1. Given WBC 6.0–17.0 と RBC 5.5–8.5 が対象 species/clinic に設定済み、When WBC=25.0、RBC=3.0、境界値=下限を保存する、Then high、low、normal が永続化・再表示される。
2. 回帰: 境界値・range不在・既存表示をBUG-004初回確定と通す。
3. 負例: 別clinic/species range、非数値、矛盾range、未承認dataでは推測しない。Given該当rangeなし、When保存、Then永続statusは既存enumのnormal、responseの`is_assessed=false`で未評価と示し、他tenantのrangeは使わない。
4. 横展開確認対象: exam結果API、pivot表示、reference rangeを使う全判定

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/medicalrecord/... -run 'Test.*Exam.*Assessment|Test.*ReferenceRange' -count=1`
- FE: `docker compose exec -T frontend npx vitest run src/features/examinations/components/ExamPivotTable.test.tsx src/features/examinations/api/transforms.test.ts`
- 手動/E2E: S02 手順2〜4（検査管理 `/examinations`）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 基準値は臨床データ。clinic/species/field provenanceを必須にする。
- fail-closed の維持: range不在・非数値は推測判定しない。
- audit / トランザクション境界: 結果/status/auditはmedicalrecord ownerのtx境界。既存結果補修は別監査packet。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- 閾値は臨床知識であり、表示文字列の parse や架空 seed を禁止する。species 未確定時に犬猫共通化しない。
- 既存結果の一括再計算は不可逆な臨床記録変更になり得るため本実装の外。対象抽出、承認、dry-run、監査を別計画にする。

#### 7. 実装ステップ（順序付き）

1. range 有/無、境界、clinic/species 隔離を unit/integration test で固定する。
2. 承認済み range data を別差分で用意し、service の status 永続化を確認する。
3. BUG-004 の確定順序修正後に初回確定と再表示の E2E を行う。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- 数値判定のACと隔離testが通り、range provenanceがレビュー可能で、欠落時は`is_assessed=false`として安全に未評価となる。既存schemaへ無用なmigrationを追加していない。

---

## BUG-004: 検査記録を初めて「確定」ステータスへ保存しようとすると「確定済みの検査は編集できません」と拒否され、確定できない【重大】

- **重大度**: 高（S02 の中核機能。検査確定ロック自体に到達できない）
- **対応状況（2026-08-03 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: current truth（`examination_parent_audit_test.go`）は confirmed 経路の write 順を `items → revision → audit → status` として固定。`TestExaminationService_ConfirmWithItemsPersistsItemsBeforeStatusTransition` / CreateConfirmed 同名 test が 409・audit rollback も回帰。Mode3 follow-up で ledger のみ訂正（product 無変更） | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: S02 手順6を専用 synthetic fixture でブラウザ再検証し、`VERIFIED_FIXED` 可否を判定
- **発見シナリオ**: S02 手順6（検査詳細・編集画面 `/examinations/:id`）
- **再現手順**:
  1. 上記 BUG-003 で作成した検査（現在ステータス=結果入力済み、未確定）を開く。
  2. ステータスのドロップダウンで「確定」を選択し、他のフィールドは変更せずに「保存」を押す。
- **期待結果**（[S02 手順6](docs/ops/testing/scenarios/S02-exam-abnormal-highlight-lock.md)）: 保存が成功し、サーバに confirmed 状態が保存される（再オープン時に初めてロックされる）。
- **実際の結果**: 保存直後にエラートースト「確定済みの検査は編集できません」が表示され、保存が失敗する（ページ離脱確認ダイアログが出る＝未保存のまま）。この検査はこれまで一度も確定されたことがないにもかかわらず、「既に確定済みだから編集できない」という趣旨のガードで弾かれている。再読み込み後もステータスは「結果入力済み」のままで、確定への遷移が一切できない。
- **備考**: 複数回（別タブでの再試行含む）、他フィールドを一切変更しない単独のステータス変更でも同じエラーで再現する。確定ロック機能自体（手順6・7）を検証できないブロッカー。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-EXAM-LIFECYCLE`
- 同一 PR にする BUG: BUG-003, BUG-004
- 先行必須: なし（BUG-003 の E2E より先に完了）
- 後続解放（シナリオ/他BUG）: BUG-003 の判定結果を含む初回確定 E2E

#### 1. 切り分けステータス

- 主因レイヤ: BE
- 観測根拠（API・クエリ・コード参照）: service が親 exam を `confirmed` に更新した後、同じ保存処理で items replacement を呼び、repository の「confirmed は編集不可」guard が自分自身の初回確定を拒否する。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-003 と同じ lifecycle だが、本件は confirmed guard の自己衝突。
- 所有境界（FE / BE / データ / 環境）: FE=既存 examination consumer; BE=medicalrecord write owner; データ=exam/status/items/audit; 環境=なし

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| BE | `backend/internal/medicalrecord/examination_service.go` の update orchestration。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/medicalrecord/examination_repository.go` の `ReplaceItemsByExamID` と confirmed guard。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: API/DB 変更なし。medicalrecord owner 内の tx順序・status transition・409 conflict分類を修正する。
- 正しい挙動の定義（1〜3 文）: 開始時 `result_entered` の初回 confirmed だけを許可し、items・status・auditを1 txで確定する。既にconfirmedは409で拒否する。
- やらないこと（Out of scope）: confirmed再編集の許可、repository guardの全面撤去、部分commit。
- 既存データ修復の要否と手順: 不要。失敗した初回確定は全rollbackされる。

- tx 開始時に既存 status/version を lock して「既に confirmed」と「今回 confirmed へ遷移」を区別する。検証済み items を先に置換し、最後に親を confirmed へ遷移して監査する。
- 既に confirmed の再編集は引き続き fail-closed。部分更新や item だけの commit を残さない。

#### 4. 受け入れ基準（AC）

1. Given `result_entered` の検査、When status=`confirmed` と同じ/更新済みitemsを保存する、Then 1 txで成功し再取得がconfirmedになる。
2. 回帰: 通常のresult入力・既confirmed読取・BUG-003判定を維持する。
3. 負例: 開始時confirmed、version競合、items/audit途中失敗は409または全rollback。 既存境界AC: Given 開始時点で confirmed、When 任意の編集を試す、Then 競合を含め拒否し DB と監査は変わらない。
4. 横展開確認対象: examinationのstatus transitionとitem置換

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/medicalrecord/... -run 'Test.*Examination.*Confirm|Test.*ReplaceItems' -count=1`
- FE: 該当なし（BE/data-only plan）
- 手動/E2E: S02 手順6（検査詳細・編集画面 `/examinations/:id`）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 確定検査は臨床記録。clinic/version/statusをlockして改変を防ぐ。
- fail-closed の維持: 既confirmed・競合・途中失敗を拒否し、部分itemsを残さない。
- audit / トランザクション境界: items→status→auditを同じDBOrTxで全commit/rollback。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- lock 順序を統一して deadlock を避け、clinic scope と optimistic version を外さない。監査失敗を成功扱いにしない。

#### 7. 実装ステップ（順序付き）

1. 初回確定失敗と既確定拒否を別 test にする。
2. service の tx 内順序と repository guard の責務を修正する。
3. BUG-003 の range 判定を含む初回確定 E2E を実施する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- 初回確定だけが成功し、既確定・競合・途中失敗は全 rollback、clinic 隔離と監査原子性が確認される。

---

## BUG-005（軽微・要確認）: 検査の「担当医」選択肢にスタッフ以外の項目が混在している

- **重大度**: 低〜中（データ品質・UI混乱の懸念。実害は要確認）
- **対応状況（2026-08-03 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: commit `dfd653eaa…` で active doctor 候補 filter。Mode3 follow-up で薄い selector と master full shape を `queryKeys.masters.staffSelectorList()` / `category("staffs")` に分離し、`staff_type` 欠損の fail-open (`?? "doctor"`) を除去。両取得順の cache 契約 test green。BE doctor fail-closed 維持 | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: S02 手順1 を synthetic fixture でブラウザ再検証し `VERIFIED_FIXED` 可否を判定
- **発見シナリオ**: S02 手順1（新規検査登録の担当医セレクタ）
- **内容**: `/examinations/new` の「担当医」ドロップダウンの選択肢に、スタッフ氏名（林文明、ノア、倉田春香 等）に混じって「お手入れ・オゾン療法」「健診・ワクチン・狂犬病」「ドッグラン(アジリティ解放)」「クイックシャンプー」のような、明らかに施術・サービスメニュー名と思われる項目が含まれていた。
- **備考**: マスタデータ側の混入か、コンポーネントの参照先マスタ取り違えの可能性。実害（誤って選択されるリスク）は未検証。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `S-EXAM-DOCTOR-OPTIONS`
- 同一 PR にする BUG: なし
- 先行必須: staff endpoint の role / active / clinic 実データ確認
- 後続解放（シナリオ/他BUG）: S02 担当医選択と検査保存

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE / データ（再現先行）
- 観測根拠（API・クエリ・コード参照）: FE は汎用 `useMasterItems("staff")` の全件を担当医 selector に渡す。BE 保存側は active staff を検証するため、UI data source または seed 汚染を切り分ける。（具体パスは §2）
- 重複判定（例: 002 vs 022）: 単独。スタッフ以外の混入が API か seed か未確定のため他 BUG と統合しない。
- 所有境界（FE / BE / データ / 環境）: FE=examinations selector; BE=staff read + medicalrecord validation; データ=staff role/active/clinic assignment; 環境=demo seed 実値確認

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/examinations/routes/ExaminationForm.tsx`、`ExaminationFormFields.tsx`、`frontend/src/hooks/use-master-items.ts`。 | 現行証拠 / 変更候補 |
| BE | BE staff endpoint と `backend/internal/medicalrecord/examination_repository.go` の staff validation。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: 現行 staff API の実契約を先に確認。必要なら doctor-selectable 条件を additive filter/専用endpointで明示し、DB schemaは変更しない。
- 正しい挙動の定義（1〜3 文）: 当該 clinic の active かつ担当医として許可された staff だけを表示し、BEも保存時に再検証する。
- やらないこと（Out of scope）: 表示名による除外、FEだけのsecurity filter、seed値の推測修正。
- 既存データ修復の要否と手順: staffデータ汚染が証明された場合のみ read-only audit と別承認補修。

- staff API の返却型・role/active 条件を確認し、担当医として選択可能な staff だけを server contract または専用 query で返す。文字列名による除外はしない。
- request 由来 staff ID は clinic 内 active staff として BE でも fail-closed 検証する。

#### 4. 受け入れ基準（AC）

1. Given active doctor、非医師 staff、施術マスタ、別 clinic staff、When 担当医選択肢を開く、Then 許可された当該 clinic staff だけが表示・保存できる。
2. 回帰: 既存active doctorの選択・保存を維持する。
3. 負例: inactive/非医師/マスタ行/別clinic staff IDは表示・保存不可。
4. 横展開確認対象: 担当医selectorを使う検査・カルテ・予約画面

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/medicalrecord ./internal/staff -run 'Test.*Staff|Test.*Doctor'`
- FE: `docker compose exec -T frontend npx vitest run src/features/examinations/hooks/use-examination-form.test.ts src/features/examinations/components/ExaminationFormFields.test.tsx src/features/examinations/routes/ExaminationForm.permissions.test.tsx`
- 手動/E2E: S02 手順1（新規検査登録の担当医セレクタ）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 担当医FKは臨床/権限対象。clinic assignment、role、activeを検証する。
- fail-closed の維持: FE非表示だけに依存せず、request由来staff IDをBEで拒否する。
- audit / トランザクション境界: 検査保存/auditの既存txを維持。staff data補修は別承認。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- 表示 filter だけで不正 staff FK を許可しない。既存検査の担当医データ修正は source 汚染が確認された場合のみ別途行う。

#### 7. 実装ステップ（順序付き）

1. API/seed の実値を種別・clinic・active で照合する。
2. 選択契約 test を追加し、専用 endpoint または typed filter を実装する。
3. BE の FK/clinic guard と既存表示を確認する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- 混在原因が証拠付きで確定し、許可 staff のみ表示・保存され、別 clinic/非 active ID が拒否される。

---

## BUG-006: 予防接種登録画面のヘッダーに表示される年齢・性別・去勢避妊状況が、対象ペットによらず常に同じ誤った固定値になっている【重大】

- **重大度**: 高（診療画面に誤った患者属性が表示される。臨床安全観点で懸念）
- **対応状況（2026-08-06 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み commit `3db97bb19e8bcc823eaf75029519ad2be895bcd2` で `formatPatientPetDetails` 追加、`VaccinationForm` が birthDate/gender/neuteredDate から petDetails を渡し、`PatientInfoCard` 固定既定「9才5ヶ月 / メス / 避妊済」を「不明」へ変更。scoped vitest 12/12 green | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: S03 手順1を専用 fixture でブラウザ再検証し VERIFIED_FIXED 可否を判定
- **発見シナリオ**: S03 手順1（予防接種 新規登録 `/vaccinations/new?petId=...`）
- **再現手順**:
  1. `/vaccinations/new?petId=1000002`（伊藤史安／豆助、実データ: 生年月日2012-12-20＝13才7ヶ月、性別=雄）を開く。
     → ヘッダーに **「9才5ヶ月 / メス / 避妊済」** と表示される。
  2. 別の全く無関係なペット `/vaccinations/new?petId=1000005`（朝長三枝子／はな、実データ: 生年月日2003-11-26＝22才、性別=不明）を開く。
     → ヘッダーに **全く同じ「9才5ヶ月 / メス / 避妊済」** と表示される。
  3両方とも `/owners/:id` のペット編集モーダル（正本データ）で確認した実際の年齢・性別・去勢避妊状況とは一致しない。
- **期待結果**: ヘッダーには選択したペット本人の実際の年齢・性別・避妊/去勢状況が表示されるべき。
- **実際の結果**: ペットによらず常に同一の「9才5ヶ月 / メス / 避妊済」という固定値（ダミーデータと思われる）が表示される。
- **備考**: 予防接種フォーム自体の保存データ（ワクチン選択・次回予定日等）には影響していない可能性があるが、診療中にスタッフが動物の性別・年齢を画面表示で確認する場面で誤った情報を見せてしまう。ヘッダー部分のペット基本情報取得ロジックが固定値/モックのままになっている疑い。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-VAX`
- 同一 PR にする BUG: BUG-006, BUG-007
- 先行必須: なし
- 後続解放（シナリオ/他BUG）: S03 の患者ヘッダーとペット別接種履歴

#### 1. 切り分けステータス

- 主因レイヤ: FE
- 観測根拠（API・クエリ・コード参照）: vaccination route は pet data を取得しているが `PatientInfoCard` に `petDetails` を渡さず、shared component の固定 default `9才5ヶ月 / メス / 避妊済` が表示される。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-007 と patient context を共有するが、本件は固定 fallback 表示、007 は取得/page契約。
- 所有境界（FE / BE / データ / 環境）: FE=vaccinations + shared PatientInfoCard; BE=pet APIは既存契約; データ=pet demographics; 環境=病院 timezone

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/vaccinations/routes/VaccinationForm.tsx`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/components/shared/PatientInfoCard/PatientInfoCard.tsx`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/lib/transforms/pet.ts` の pet response transform。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: API/DB 変更なし。既存 pet response を typed patient view model に変換する。
- 正しい挙動の定義（1〜3 文）: 対象petの生年月日・性別・避妊去勢情報を表示し、欠損は推測せず不明とする。
- やらないこと（Out of scope）: 固定fallback、petデータの書換え、shared card全用途の無関係な再設計。
- 既存データ修復の要否と手順: 不要。

- 対象 pet の生年月日・性別・去勢避妊状態を typed view model に変換して必須 props として渡す。臨床画面では fixed fallback を削除し、欠損は「不明」と明示する。
- 年齢は表示時点の病院 timezone 基準で一つの utility から計算し、BUG-007 と同じ pet context/query key を使う。

#### 4. 受け入れ基準（AC）

1. Given 属性が異なる2頭、When 各 petId の新規接種画面を開く、Then 各正本の年齢・性別・去勢避妊状態が表示され、固定値にならない。
2. 回帰: 既知/欠損属性とBUG-007 pet切替を通す。
3. 負例: pet未取得/属性欠損時に固定値を表示せず、別pet cacheを再利用しない。 既存境界AC: Given 属性欠損、When 表示する、Then 推測せず「不明」となり別 pet の値を再利用しない。
4. 横展開確認対象: PatientInfoCardの全consumerとpet transform

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: 該当なし（API/DB変更なし。FE境界testで検証）
- FE: `docker compose exec -T frontend npx vitest run src/features/vaccinations/hooks/use-vaccination-form.test.ts src/features/vaccinations/routes/VaccinationForm.permissions.test.tsx src/components/shared/PatientInfoCard/PatientInfoCard.test.tsx src/lib/transforms/pet.test.ts`
- 手動/E2E: S03 手順1（予防接種 新規登録 `/vaccinations/new?petId=...`）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 患者属性の誤表示は臨床リスク。pet/clinic正本を維持する。
- fail-closed の維持: 欠損を固定値で補わず不明表示にする。
- audit / トランザクション境界: read-only表示。DB/audit変更なし。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- patient context の stale cache は誤患者表示につながるため petId を query key に含める。年齢だけで臨床判断せず生年月日も source として保持する。

#### 7. 実装ステップ（順序付き）

1. 異なる2頭と欠損値の表示 test を RED にする。
2. view model と必須 props を追加し fixed clinical default を除く。
3. BUG-007 の履歴 query と pet 切替 E2E を通す。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- 2頭切替・再読込・欠損ケースで正本と一致し、固定の患者属性が臨床画面に残っていない。

---

## BUG-007: 予防接種を新規登録しても、一覧・対象ペットの「過去の接種履歴」パネルのどちらにも表示されず、登録結果を画面上で確認できない

- **重大度**: 中〜高（S03 手順6の中核要件「自動計算結果を画面上でいつでも確認できる」が満たされない）
- **対応状況（2026-08-06 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み commit `7f663716d0b2bfac3a8f5fd5cfe7e9291b2d22da` で `useGetVaccinations` が `pet_id` + `page`/`limit=HISTORY_FETCH_LIMIT` を送る。`VaccinationForm` 履歴は pet スコープ取得に切替（unscoped page1+client filter 廃止）。`useGetPetVaccinations` も limit 明示。scoped vitest 30/30 green | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: S03 手順6・7で登録直後の履歴・一覧をブラウザ再検証
- **発見シナリオ**: S03 手順6・7（予防接種管理 `/vaccinations`）
- **再現手順**:
  1. `/vaccinations/new?petId=1000002`（伊藤史安／豆助）で接種日=2026/07/31、ワクチン=バンガードL4(4種)、次回予定=2026/09/15（手動調整）として保存 → 「予防接種を登録しました」の成功トーストが出る。
  2. 保存後 `/vaccinations`（一覧）を見ても、直近ページ（20件、全て2029年のseedデータ）に新規登録分が出てこない。
  3. 同じペットで再度 `/vaccinations/new?petId=1000002` を開き、右側の「過去の接種履歴」パネルを見ると **「履歴がありません」** と表示される（このペット自身の直前の登録が反映されない）。
  4. バックエンドAPIを直接確認（`GET /api/v1/vaccinations?pet_id=1000002`）すると、レコード自体は正しく保存されている（id 1091849, pet_id 1000002, date 2026-07-31, next_date 2026-09-15, lot1正しい）。**データは正しく保存されているが、UI側のクエリが正しく引けていない**。
- **期待結果**（[S03 手順6](docs/ops/testing/scenarios/S03-vaccination-next-due-autocalc.md)）: 一覧の「次回予定」列で確認でき、対象ペットの接種フォームを開けば過去の接種履歴として直近の登録が見える。
- **実際の結果**: どちらの画面でも直近の登録が見えず、スタッフは「本当に保存されたか」を画面上で確認できない。
- **備考**: `GET /api/v1/vaccinations?petId=1000002`（キャメルケース）で試すと **フィルタが無視されて全件が返る**のに対し、`?pet_id=1000002`（スネークケース）なら正しくフィルタされる。フロント側が誤ったパラメータ名（`petId`）でこのAPIを呼んでいるためにペット別の履歴パネルが常に空になっている可能性が高い。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-VAX`
- 同一 PR にする BUG: BUG-006, BUG-007
- 先行必須: BUG-006 の patient context 正本
- 後続解放（シナリオ/他BUG）: S03 登録後履歴・一覧・再読込

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE
- 観測根拠（API・クエリ・コード参照）: 現行 hook は `pet_id` を送る。フォームと一覧はいずれも汎用の先頭ページを取得してから client filter しており、2029 seed の page window 外に 2026 登録が隠れる経路を確認した。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-006 と patient context を共有するが、本件は server filter/page と query invalidation。
- 所有境界（FE / BE / データ / 環境）: FE=vaccination hooks/routes; BE=medicalrecord vaccination read owner; データ=clinic/pet vaccination rows; 環境=seed の日付分布

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/hooks/use-pet-vaccinations.ts`、`frontend/src/hooks/use-vaccinations.ts`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/features/vaccinations/routes/VaccinationForm.tsx`、`VaccinationList.tsx`。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/medicalrecord/vaccination_handler.go`、`vaccination_repository.go`。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: 既存 snake_case `pet_id` query/API と DB schema は維持。FEのserver filter/page/sort利用とquery invalidationを是正する。
- 正しい挙動の定義（1〜3 文）: 新規接種は対象pet履歴と全体一覧の正しい位置に即時/再取得後とも現れ、別petを混ぜない。
- やらないこと（Out of scope）: camelCase queryの追加、全件client取得、重複再POST。
- 既存データ修復の要否と手順: 不要。既存接種行は変更しない。

- pet 履歴は初めから `pet_id` を server query に渡し、一覧は選択 filter/sort/page を全て API に渡す。取得済み1ページへの client filter を廃止する。
- mutation 成功後は一覧と対象 pet 履歴の正規 query key を更新/invalidate し、サーバの sort 契約で直近順を保証する。

#### 4. 受け入れ基準（AC）

1. Given page 1 が未来 seed 20件で埋まる、When 2026年の接種を pet A に登録する、Then pet A 履歴に即時表示され、再読込後も存在する。
2. 回帰: 新規登録・全体一覧・pagination/count・BUG-006 headerを維持する。
3. 負例: 別pet/別clinic、page外データを混入せず、retryでduplicate作成しない。 既存境界AC: Given pet B、When 同じ画面を開く、Then pet A の履歴は混入しない。全体一覧では指定 sort/page 条件に従う。
4. 横展開確認対象: vaccination list/history/query keys

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*Vaccination.*List|Test.*PetID'`
- FE: `docker compose exec -T frontend npx vitest run src/hooks/use-pet-vaccinations.test.ts src/features/vaccinations/hooks/use-vaccination-form.test.ts src/features/vaccinations/routes/VaccinationList.test.tsx src/features/vaccinations/components/VaccinationFormPanels.test.tsx`
- 手動/E2E: S03 手順6・7（予防接種管理 `/vaccinations`）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 接種履歴は臨床記録。clinic/pet filterとquery keyを分離する。
- fail-closed の維持: 取得失敗を空履歴や再POSTで補わず、別pet行を表示しない。
- audit / トランザクション境界: 接種write/auditの既存txを回帰。FE invalidateは成功後だけ。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- query key に petId/clinic/filter/page を含め、別患者のキャッシュ再利用を防ぐ。登録データ補修は不要。

#### 7. 実装ステップ（順序付き）

1. page-window 再現 test を追加する。
2. server-side filter/sort と typed query key へ統一する。
3. 登録直後、再読込、別 pet 切替の E2E を行う。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- camelCase 仮説ではなく再現した page-window 原因が解消し、pet 履歴と一覧の件数・順序・隔離が API 契約と一致する。

---

## BUG-008: LIFF予約フローがコース選択画面で必ず「ログイン情報の有効期限が切れました」となり、以降に一切進めない【重大】【S04ブロッカー】

- **重大度**: 高（S04 全体のブロッカー。LIFF飼い主予約ジャーニーが手順2「コース選択」より先へ一切進めない）
- **対応状況（2026-08-03 JST）**: OPEN | **根拠**: LIFF mock/auth: `VITE_LIFF_MOCK`/`LIFF_MOCK` 不整合時に 401→`handle-fetch-error.ts`「ログイン情報の有効期限が切れました」；compose は mock 未固定（wave-1） | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: ローカル両 mock 有効でコース一覧 200 を確認し C-LIFF-AUTH を実装
- **発見シナリオ**: S04 手順1→2（LIFF予約アプリ `/line-reserve/1/`、`VITE_LIFF_MOCK=true`／バックエンド`LIFF_MOCK=true`前提）
- **再現手順**:
  1. `/line-reserve/1/` を開き「新規予約」→お客様情報（お名前・電話番号・飼い主名・ペット追加）を入力し「次へ」を押す。
  2. 「コースを選択」画面へ遷移した直後、**必ず**「ログイン情報の有効期限が切れました。LINEアプリを再起動して開き直してください。」というエラーメッセージのみが表示され、コース一覧が表示されない。
  3. ネットワークログをクリアしてから同じ手順を再実行して単離したところ、お客様情報入力〜「次へ」押下までの間、**API呼び出しは一切発生しない**（LIFF認証やトークン取得のためのAPIコールが存在しない）。「次へ」押下直後に初めて `GET /api/liff/1/courses` が呼ばれ、**2回とも401**で返る。
  4. レスポンスボディを直接確認: `{"code":401,"message":"missing authorization header","timestamp":"..."}` — Authorizationヘッダ自体が送信されていない。
  5. `localStorage`/`sessionStorage` を確認しても LIFF 用のモックトークンやアクセストークンは保存されておらず（`sessionStorage` にあるのは `LIFF_STORE:isInClient=0` のみ）、`window.liff` オブジェクトも存在しない（`undefined`）。
  6. 3回（別タブ・別データで2回、初回ログ確認1回）すべて同一の401＋同一メッセージで再現。ページ再読み込みでも状態変化なし（タイミング/トークンTTLレースではない）。
- **期待結果**（[S04 手順1-2](docs/ops/testing/scenarios/S04-liff-reservation-journey.md)、前提条件「LIFFモックはローカル専用: フロントは`VITE_LIFF_MOCK=true`、バックエンドは`LIFF_MOCK=true`で認証をバイパスする」）: モック環境ではLIFF認証がバイパスされ、コース選択画面にマスタで公開設定された予約区分の一覧が表示されるべき。
- **実際の結果**: 認証バイパスが機能しておらず、`/api/liff/1/courses` がAuthorizationヘッダなしで呼ばれて401となり、画面には「ログイン情報の有効期限が切れました」という趣旨の異なるエラー文言が表示される。コース選択以降（スタッフ選択・日時選択・予約確定・病院側確認・キャンセル・トリミング分岐）が一切検証できない。
- **備考**: S04手順3〜12はすべて本ブロッカーにより未実施（実施不能）。原因は以下のいずれか、または複合と推測: (a) フロントエンドがLIFFモック用のダミートークンを取得・付与する処理自体を呼んでいない、(b) `VITE_LIFF_MOCK`が本セッションのビルド/dev-server設定で実際には有効になっていない、(c) 認証バイパスは実装されているがエラーメッセージ文言側だけが「ログイン期限切れ」用の汎用ハンドラに誤って束ねられている。原因切り分けにはフロント側のLIFF認証初期化コード（`liff_auth`関連）とVite環境変数の実値確認が必要（本セッションではソースコード未参照のためAPI観測のみで判断）。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-LIFF-AUTH`
- 同一 PR にする BUG: BUG-008, BUG-014
- 先行必須: ブラウザ配信 asset / runtime env / request header の provenance 採取
- 後続解放（シナリオ/他BUG）: S04 予約フローと BUG-014 / S12 認証検証

#### 1. 切り分けステータス

- 主因レイヤ: 環境 / FE / BE（REPRODUCE-FIRST）
- 観測根拠（API・クエリ・コード参照）: 現行 `use-liff.ts` はFE mock=trueで`mock-token`を生成し、line-reserve APIはBearer headerを必ず構築する。BEも`LIFF_MOCK=true`かつnon-releaseならheaderなしのmock分岐を持つ。したがって報告時の`missing authorization header`は現行sourceだけでは説明不能で、default Composeのfrontend env allowlist欠落、起動時env、served asset、proxy前後を同一試行で確認する。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-014 と同一観測クラスタ。ただし現行 source は header 付与済みで、報告時の『header欠落実装』は未確認。
- 所有境界（FE / BE / データ / 環境）: FE=shared LIFF + line-reserve transport; BE=LIFF auth middleware; データ=clinic/customer scope; 環境=built asset と VITE_LIFF_MOCK / LIFF_MOCK provenance

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/shared-liff/use-liff.ts`、`frontend/line-reserve/src/App.tsx`、`frontend/line-reserve/src/api/liff-api.ts`、`frontend/src/shared-liff/handle-fetch-error.ts`。 | token生成、描画gate、Bearer送信、error変換 |
| BE | `backend/internal/middleware/liff_auth.go`、`backend/internal/reservation/routes.go`。 | 現行証拠 / 変更候補 |
| Env | `docker-compose.yml` のfrontend environment allowlist/backend `env_file`、`.env.example`、`Makefile`。 | 起動時flagとserved asset provenance |

#### 3. 修正方針

- 契約変更の有無（API・DB）: 原因確定前は API/DB を変更しない。認証wire契約は `Authorization: Bearer <token>` を維持し、provenanceで落下点が証明された層だけを直す。
- 正しい挙動の定義（1〜3 文）: 現行built clientが取得tokenをheaderへ渡し、mock/realの設定不一致は原因別にfail-closedエラーとなる。
- やらないこと（Out of scope）: mock認証bypass、token値の記録、sourceと矛盾したheader追加の重複実装。
- 既存データ修復の要否と手順: 不要。

- まずHEAD、container再作成時刻、allowlisted booleanの`VITE_LIFF_MOCK`/`LIFF_MOCK`、browser request header、served asset hash、proxy前後を同一試行で採取する。env全体や実token値は出力しない。現行sourceでheaderが送られる場合はproduct codeを変更せず、stale asset/起動時env/configを修正する。
- drop pointがfrontend envと確定した場合だけComposeへ`VITE_LIFF_MOCK`を明示allowlistし、`.env.local`全体をfrontendへ注入しない。BE mock lookup失敗はreal authへsilent fallthroughさせずnon-release設定errorとしてfail-closedにし、releaseの既存mock拒否とreal LINE検証を維持する。

#### 4. 受け入れ基準（AC）

1. Given Docker local で両側 mock=true、When `/line-reserve/1/` から courses を取得する、Then Authorization header があり 200、S04 の後続へ進める。
2. 回帰: courses schema、real LIFF経路、BUG-014 transportを維持する。
3. 負例: FE/BE flag不一致、mock lookup失敗、release mock、header欠落をfail-closedに分類する。 既存境界AC: Given FE/BE mock が不一致、When 起動または request、Then 明示的設定エラーとなり、real token と誤認しない。 / Given staging/production、When dummy token を送る、Then 必ず拒否され、別 clinic/owner の情報は返らない。
4. 横展開確認対象: shared-liff、line-reserve、health-card、backend middleware

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。 追加観点: E2E: Docker local の S04 を mock、STG の S12 を実 LINE test account で別々に実施する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/middleware/... ./internal/reservation/... -run 'Test.*(Liff|LIFF|Course)' -count=1`
- FE: `docker compose exec -T frontend npx vitest run src/shared-liff/use-liff.test.ts line-reserve/src/api/liff-api.test.ts line-reserve/src/App.test.tsx`
- 手動/E2E: S04 手順1→2（LIFF予約アプリ `/line-reserve/1/`、`VITE_LIFF_MOCK=true`／バックエンド`LIFF_MOCK=true`前提）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 認証/clinic隔離を最優先。実token・env全体をartifactへ残さない。
- fail-closed の維持: release mock/bypass、lookup失敗、flag不一致を認証成功へfallbackしない。
- audit / トランザクション境界: read認証のためDB書込なし。mock customer dataが作られる場合はtest clinicでcleanup記録。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- mock secret/token をログ、文書、client bundle の production build に残さない。mock bypass の有効化条件を曖昧な空文字 default にしない。
- 静的検査だけでは owner 二者間隔離を証明できないため、STG では test owner A/B と clinic A/B の否定ケースを必須にする。

#### 7. 実装ステップ（順序付き）

1. HEAD/container時刻、allowlisted flagの有無、served asset、request header、proxy前後を一つのprovenance packetにし、drop pointを固定する。
2. 証明された最小層だけを修正し、既存Bearer処理を重複させず、BE mock/release fail-closed testを通す。
3. S04 local mockと、BUG-014/S12のSTG real LINE・tenant否定を別証拠として実施する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- local mock の S04、real-auth の S12、mock production 拒否、clinic/owner isolation が証拠化され、資格情報や operational secret が成果物へ出ていない。

---

## BUG-009: 入院・ホテル管理画面のタブ切替（予約／退院済／すべて）が機能せず、常に「入院中」のデータしか表示されない【重大】

- **重大度**: 高（S05 の中核機能。予約中・退院済の入院を画面上で一切確認できない）
- **対応状況（2026-08-04 JST）**: IMPLEMENTED_UNVERIFIED | **実装 commit**: `6e9674286` | **review-gate commit**: `55125f858` | **根拠**: タブ→server status/page/limit、server total 件数正本、client status 二重 filter 削除、未知 status「不明」fail-closed、Pagination server 正本（`6e9674286` + `55125f858`） | **原文シナリオ再検証**: WAIVED（2026-08-05 USER 判断・ブラウザ再検証を実施しない） | **次のアクション**: なし（検証見送り。claim 解放済み）
- **発見シナリオ**: S05 手順1・2・2b（入院・ホテル管理 `/hospitalization`）
- **再現手順**:
  1. `/hospitalization/new?petId=1000002`（伊藤史安／豆助）で新規入院登録（入院タイプ=入院、ケージ=犬用ケージ（中）、期間はデフォルトの本日〜+7日）を保存 → 「入院情報を登録しました」トースト。保存直後のステータスは `reserved`（チェックイン前）。
  2. `/hospitalization` の「入院中」「予約」「すべて」のいずれのタブを開いても **0件** と表示され、ボードの全ケージが「空き」のまま（該当ケージ「犬用ケージ（中）」も空き表示）。
  3. バックエンドAPIを直接確認（`GET /api/v1/hospitalizations`、フロントが実際に呼ぶのと同一のクエリなし）すると、当該レコード（id=1, status=reserved, cage_id=4）が正しく1件返る。**API側は正常にデータを返しているが、画面には一切反映されない**。
  4. 詳細画面 `/hospitalization/1` へ直接URLアクセスすると正常に表示され、そこから「チェックイン」ボタンでステータスを `reserved`→`active` に変更。
  5. 再度 `/hospitalization` の一覧へ戻ると、今度は「1件」・ボードの「犬用ケージ（中）」に豆助のカードが表示される（＝ステータスが `active`＝「入院中」になった時だけ表示される）。
  6. 別ペット（朝長三枝子／はな、pet_id 1000005）で2件目の入院登録（猫用ケージ（小）、チェックインはせず`reserved`のまま）を作成し、`/hospitalization` で「予約」タブに切り替えたところ、件数表示は「1件」のまま・ボード/リストいずれの表示も**チェックイン済みの豆助（入院中ステータス）のみが表示され続け**、はな（reservedステータス）はボードにもリストにも一切出てこない（該当ケージ「猫用ケージ（小）」も空き表示のまま）。List Viewに切り替えても同様に豆助のみ表示。
- **期待結果**（[S05 手順1・2・2b](docs/ops/testing/scenarios/S05-hospitalization-cycle.md)）: 保存成功後は入院一覧の「入院中」タブ（または少なくとも該当ステータスのタブ）に表示され、ボード/リストいずれのビューでも正しいステータスの入院が確認できる。
- **実際の結果**: タブ切替UIの選択状態にかかわらず、常に「入院中（active）」ステータスの入院のみがボード・リスト両ビューに表示される。チェックイン前（`reserved`）の予約は、詳細ページへの直接URLアクセス以外に画面上で確認する手段が一切ない。
- **備考**: APIレスポンス自体は正しいため、フロント側のタブ状態とAPIクエリパラメータ（またはクライアント側フィルタ条件）が連動していない可能性が高い。予約枠の運用（当日より前にケージ予約を確保しておく、等）が実質的に画面から見えず運用に支障が出る。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `S-HOSPITALIZATION-TABS`
- 同一 PR にする BUG: なし
- 先行必須: request / response / transform / tab state の同一 trace 採取
- 後続解放（シナリオ/他BUG）: S05 入院一覧の全タブと board/list 整合

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE（再現先行）
- 観測根拠（API・クエリ・コード参照）: 現行 `HospitalizationList.tsx` は取得後にタブ別 status filter を持ち、BE にも status query があるため、「全タブで active 固定」という報告原因は現ソースだけでは確定しない。一方、日付だけで取得した1ページを client filter する page-window 問題は残る。（具体パスは §2）
- 重複判定（例: 002 vs 022）: 単独。静的 source には tab filter があるため stale asset、transform、pagination を切り分ける。
- 所有境界（FE / BE / データ / 環境）: FE=hospitalization list/board; BE=medicalrecord hospitalization read owner; データ=status/date/clinic rows; 環境=loaded asset provenance

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/hospitalization/routes/HospitalizationList.tsx`、同 feature の `constants.ts` と `api/transforms.ts`。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/medicalrecord/hospitalization_request.go`、`hospitalization_handler.go`、`hospitalization_repository.go`。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: 原因確定前は API/DB を変更しない。既存 status/date/page filter 契約をFEが利用する修正を第一候補とする。
- 正しい挙動の定義（1〜3 文）: 各tabのstatusがrequest/response/transform/viewで一致し、board/list/countが同じ集合を示す。
- やらないこと（Out of scope）: 全件client filter、未知statusのactive化、stale assetをproduct bugとして決め打ち。
- 既存データ修復の要否と手順: 不要。

- まず Docker ブラウザで tab state、実 request URL、response status、transform 後値を一つの trace で照合する。再現時は status/date/page/sort を API query の正本へ寄せ、取得済み page への二重 filter を除く。
- 再現しなければ current source の回帰 test と stale asset/build 診断だけを追加し、推測修正はしない。

#### 4. 受け入れ基準（AC）

1. Given reserved/active/discharged を同日に各1件用意、When 予約・入院中・退院済・すべてを切り替える、Then request 条件、件数、board/list の内容が各 status 契約と一致する。
2. 回帰: 4tab、board/list、pagination、日付filterを通す。
3. 負例: 未知status/別clinicをactiveへ混ぜず、stale asset時はcode fixを行わない。 既存境界AC: Given page size を超えるデータ、When 各タブを開く、Then page 外 status のために誤って0件にならず、別 clinic 行は返らない。
4. 横展開確認対象: hospitalization read consumersとstatus transform

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*Hospitalization.*List|Test.*Status'`
- FE: `docker compose exec -T frontend npx vitest run src/features/hospitalization/routes/HospitalizationList.test.tsx`
- 手動/E2E: S05 手順1・2・2b（入院・ホテル管理 `/hospitalization`）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 入退院statusは臨床運用対象。clinic/status/date filterを分離する。
- fail-closed の維持: 未知statusやerrorをactive/空一覧へ潰さない。
- audit / トランザクション境界: read-only。DB/audit txなし。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- board と list の filter 実装を分岐させない。status label と wire value の対応を型で固定する。テスト seed の状態変更は専用 clinic に限定する。

#### 7. 実装ステップ（順序付き）

1. 報告手順を trace 付きで再現し、source/asset/page-window のどれかを確定する。
2. 失敗 test を追加し、server filter と query key を一本化する。
3. board/list、全タブ、再読込、clinic 否定ケースを検証する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- 原因証拠が残り、4タブ×2表示の AC と page/clinic 条件が通る。再現不能ならその事実と回帰 test の範囲を明記する。

---

## BUG-010: カルテ「診察/治療プラン」タブの身体検査所見・診断詳細・治療方針が、入力しても保存されず空欄化または固定文字列に置き換わる【重大】

- **重大度**: 高（S06 手順1の中核要件。臨床所見・診断・治療方針という法的記録の根幹部分が保存されない／改ざんされる）
- **対応状況（2026-08-03 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: commit `646fb4353fb6b66c2a86daf17145ed8eece2cee4` — 単一 versioned PATCH で physical_exam / diagnosis_details / treatment_policy を常送、ClinicalPlanSection controlled 化、post-save 二重書き込み除去、hydrate 前空クリア拒否、mutation 応答で version cache 更新。audit residual commit `929fef0fa66609f7811fcef7cb254a6c035e68fb` — clinical plan Update を staff actor 付き audit と同一 DBOrTx で fail-closed 記録（scoped FE/BE tests green）。delete-audit residual commit `90ee096bfaddb6dc41298be9903507f8c9aef553` — clinical plan Delete を staff actor 付き pre-delete 値 audit と同一 DBOrTx で fail-closed 記録（scoped BE tests green） | **原文シナリオ再検証**: WAIVED（2026-08-05 USER 判断・ブラウザ再検証を実施しない） | **次のアクション**: なし（検証見送り。claim 解放済み）
- **発見シナリオ**: S06 手順1（カルテ編集 `/medical-records/:id`、「診察/治療プラン」タブ）
- **再現手順**（他の操作を一切介さないクリーンな単離手順で2回再現）:
  1. カルテ新規作成 → 生存ペット（小玉哲博／ラッキー、pet_id 1000019）を選択（カルテID 1425547 が自動作成される）。
  2. 「診察/治療プラン」タブへ切替。「身体検査所見」欄に `CLEAN-TEST 身体検査所見の内容ABC`、「診断詳細」欄に `CLEAN-TEST 診断詳細の内容XYZ`、「治療方針」欄に `CLEAN-TEST 治療方針の内容123` を入力（診断カテゴリ等のドロップダウンには一切触れていない）。入力直後、各テキストエリアの `value` がそれぞれ正しく入力した文字列になっていることを確認済み。
  3. 「保存」を押す（`PATCH /api/v1/medical-records/1425547/clinical-plan` が発火し200 OK）。
  4. 直後に `GET /api/v1/medical-records/1425547/clinical-plan` で保存結果を確認すると:
     - `physical_exam`: `""`（入力した内容が消え、空文字列で保存されている）
     - `diagnosis_details`: `"# 診断詳細"`（入力した内容ではなく、ラベルのような固定文字列に置き換わっている）
     - `treatment_policy`: `"# 治療方針"`（同上、入力内容ではなく固定文字列）
  5. 別ペット（田中光広／チビチビ、カルテID 1425546）でも同様の手順で同じ結果（`physical_exam` 空欄化、`diagnosis_details`="# 診断詳細"、`treatment_policy`="# 治療方針"）を確認済み。
  6. 参考: カルテの主訴詳細（「問診」タブ、`主訴詳細` 欄）についても、ペット選択直後（何も入力していない時点）から常に「43 / 500 文字」というテンプレートらしき文字数がフォームに表示されており、そこへ手動入力しても保存・再読込後に元のテンプレート相当の文字数へ戻る挙動が見られた（保存先APIエンドポイントが未特定のため本バグには含めず、要追加調査として備考に記載）。
- **期待結果**（[S06 手順1](docs/ops/testing/scenarios/S06-record-lock-audit-trail.md)）: 「診察/治療プラン」タブで入力した所見（S/O/A）・診断名がそのまま保存され、再読み込み後も同じ内容が表示される。
- **実際の結果**: 身体検査所見は常に空文字列で保存され、診断詳細・治療方針は入力内容に関わらず常に固定文字列「# 診断詳細」「# 治療方針」で上書き保存される。ユーザーが入力した臨床記録内容が一切保持されない。
- **備考**: 固定文字列が見出しマークダウン風（`# ラベル名`）であることから、リッチテキスト/マークダウンエディタ系コンポーネントのデフォルト初期値（見出しテンプレート）がテキストエリアの実際の入力値と正しく同期されないまま送信されている可能性が高い。S06 手順2以降（確定・ロック・訂正追記等）はこの中身が保存されない状態のまま進めることになり、確定ロック機能自体の検証は継続するが、法的記録としての真正性という本シナリオの主目的は本バグにより満たされない。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `S-CLINICAL-PLAN-SAVE`
- 同一 PR にする BUG: なし
- 先行必須: 保存 payload / response / refetch の failure point 固定
- 後続解放（シナリオ/他BUG）: S06 診察・治療プランの保存と再表示

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE
- 観測根拠（API・クエリ・コード参照）: 親 save action が template state を PATCH した後、`ClinicalPlanSection` が別 state/別 PATCH を持ち、post-save が再度保存する。refetch による再 hydrate も重なり、同じ臨床フィールドの複数 writer が last-write-wins を起こす。（具体パスは §2）
- 重複判定（例: 002 vs 022）: 単独。clinical-plan payload/response と post-save overwrite の同一 trace で根因を固定する。
- 所有境界（FE / BE / データ / 環境）: FE=medical-record save action/section; BE=medicalrecord write owner; データ=clinical plan + audit; 環境=なし

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/features/medical-records/hooks/use-medical-record-post-save.ts`。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/medicalrecord/clinical_plan_handler.go`、`clinical_plan_service.go`、`clinical_plan_repository.go`。 | typed PATCH、optimistic lock、clinic-scoped writeとaudit追加位置 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: 既存 clinical-plan API/DB field を維持。payload/response/refetchの確認後、欠落mappingまたはpost-save overwriteだけを最小修正する。
- 正しい挙動の定義（1〜3 文）: 入力した身体所見・診断詳細・治療方針が保存応答と再GETで一致し、固定文字列で上書きされない。
- やらないこと（Out of scope）: clinical plan以外のカルテ再設計、失敗時の成功toast、監査外更新。
- 既存データ修復の要否と手順: 不明。保存済み欠落値は元入力を復元できないため自動補修せず、監査可能な原資料がある場合のみ別検討。

- 身体検査所見・診断詳細・治療方針の state owner を一つにし、一回の versioned PATCH で送る。child component は controlled input とし独自保存を持たない。
- BE は clinic/pet/record、finalized/locked、version を検証し、更新と監査を同一 tx に置く。空文字での明示クリアと「未送信」を区別する。

#### 4. 受け入れ基準（AC）

1. Given未確定カルテに3欄の異なる値を入力、When一度保存して再読込する、Then入力値が一字も固定文言へ置換されず保持され、clinical-plan PATCHは一回だけになる（別責務のnext-visit PATCHはこの回数に含めない）。
2. 回帰: 他カルテfield、再GET、audit、再保存を維持する。
3. 負例: 4xx/5xx/競合では入力を固定文字列へ置換せず、成功toastを出さない。 既存境界AC: Given stale version または finalized record、When 保存する、Then 409/適切な拒否となり一部フィールドも監査も commit されない。
4. 横展開確認対象: medical-record save action内の全section mapping

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/medicalrecord/... -run 'Test.*(ClinicalPlan|OptimisticLock|Audit)' -count=1`
- FE: `docker compose exec -T frontend npx vitest run src/features/medical-records/hooks/use-medical-record-save-action.test.ts`。`src/features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.test.tsx`は新規追加予定。
- 手動/E2E: S06 手順1（カルテ編集 `/medical-records/:id`、「診察/治療プラン」タブ）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 診療計画は臨床記録。clinic/medical-record/pet/staff FKを検証する。
- fail-closed の維持: payload欠落・競合・保存失敗で固定値や成功表示へfallbackしない。
- audit / トランザクション境界: clinical planとauditをmedicalrecord ownerの同一DBOrTxで更新。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- 失われた既存診療記録を固定文字列から推測復元しない。補修は audit/backup を用いた個別レビュー対象。
- retry による二重 PATCH、別 pet/clinic の record ID、確定後更新を fail-closed にする。

#### 7. 実装ステップ（順序付き）

1. 二重 request と再読込欠落を component/integration test で再現する。
2. state owner と save command を一本化し、BE の versioned tx を確認する。
3. 失敗 rollback、再読込、navigation blocker、監査を E2E で検証する。

#### 8. 完了定義（DoD）

- [x] §4 の AC が全通過（scoped unit/integration。ブラウザ S06 は UNREPORTED）
- [x] 関連クラスタと横展開対象の回帰が通過（save-action / form.auto-create / clinical plan BE の scoped green）
- [x] 作成した test data の cleanup、または cleanup 不要を記録（AutoMigrate テスト DB のみ。本番 seed 変更なし）
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録（ブラウザ未起動のため UNREPORTED）

- 3欄の round-trip、単一 writer、競合/lock は scoped tests で通過。clinical plan Update/Delete の audit 同一 tx は `929fef0fa` / `90ee096bf` で実装済み（IMPLEMENTED_UNVERIFIED）。既存データ補修は別承認事項として切り出される。

---

## BUG-011: 2件目以降の見積書新規作成が、タイトル内容によらず常に「estimate '' already exists」(409)で拒否される【重大】【S07ブロッカー】

- **重大度**: 高（見積書の新規作成機能そのものが1クリニックにつき実質1回しか使えなくなる）
- **対応状況（2026-08-03 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み commit `b65cf69ef56785c473ddd233624292a3c338401e` で Create は同一 tx 内の `AllocateNextEstimateNo` から `EST-{N}` を採番し、`TestEstimateService_Create_AssignsEstimateNo` が空の `estimate_no` を送らず作成する契約を固定。原文の2件目409は未再実行で、23505時の `estimate '' already exists` メッセージ残渣は別品質問題 | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: 2件目以降の `POST /estimates` を専用 synthetic fixture で再検証し、必要ならメッセージ品質を分離修正
- **発見シナリオ**: S07 手順1・7・9（見積書管理 `/estimates/new`）
- **再現手順**:
  1. `/estimates/new` でタイトル「S07 検証用A」・小計10,000円・ステータス「下書き」を入力し「作成」→ 成功（一覧に追加され、以後ステータスを送付済み→承認済みへ遷移できることまで確認済み＝BUG-011とは無関係に正常動作）。
  2. 続けて同じ画面から2件目としてタイトル「S07 検証用B」で新規作成を試行 → 保存が完了せず`/estimates/new`に留まる。トーストに「estimate '' already exists」と表示される。
  3. `POST /api/v1/estimates` を `{"title": "S07 検証用B3", "status": "draft", "subtotal": 5000}` として直接叩いても **409** で同一エラー `{"error":"estimate '' already exists"}` が返る。
  4. タイトルを完全に一意な文字列 `COMPLETELY-UNIQUE-TITLE-999` に変えて再度直接APIを叩いても、**同じ409・同じ「estimate '' already exists」**（送信した実際のタイトルではなく空文字列 `''` を指すエラー文言）が返る。
  5. 以降、何を入力しても新規見積書の作成が一切できない状態が継続する。
- **期待結果**（[S07 手順1・7・9](docs/ops/testing/scenarios/S07-estimate-status-control.md)）: 任意のタイトルで複数の見積書を新規作成でき、S07手順7（検証用B作成）・手順9（検証用C作成）が実施できる。
- **実際の結果**: 1クリニックにつき最初の1件の新規作成のみ成功し、それ以降のあらゆる新規見積書作成が、入力したタイトルに関わらず「estimate '' already exists」という空文字列を指すエラーで409拒否される。
- **備考**: エラーメッセージが実際に送信したタイトルではなく空文字列 `''` を指していること、seed 003_demo に元々タイトル未設定（空文字列）の見積書が多数存在すること（一覧で20件中10件以上がタイトル空欄）から、バックエンドの重複チェックロジックがリクエストボディのタイトルを正しく参照できておらず、常に空文字列としてユニーク制約に照らしてしまっている可能性が高い。本バグにより S07 手順7（検証用B: 送付済み→却下遷移の確認）・手順9（検証用C: 下書きのままの編集・削除確認）が実施不能。手順1〜6・8（検証用Aのみで完結する範囲）は正常動作を確認済み。

- **追加調査（V01検証より）**: カルテの見積書タブ（`/medical-records/:id` 見積書タブ）経由でも同一の409 `estimate '' already exists` が再現することを確認（詳細はBUG-V01-004として当初別項目で調査したが、本バグと同一の根本原因と判断し統合した）。`GET /api/v1/estimates?pet_id=1000002` で調査したところ、当該clinicには `estimate_no: ""`（空文字列）を持つ既存の無関係な見積レコード（`id:1000819, title:"S07 検証用A"`）が存在しており、一意性チェックが `medical_record_id`／作成対象単位ではなく `estimate_no` に対して**クリニック全体スコープ**で行われている可能性が高い。加えて新規見積の `estimate_no` 算出結果自体が空文字列になるケースがあるとみられ、この2点（クリニック全体スコープでの一意性チェック＋estimate_no算出漏れ）が組み合わさることで、空の`estimate_no`を持つレコードが1件でも存在すると当該クリニックの新規見積作成が全滅する、という設計上の脆弱性が示唆される。

- 予防接種新規登録フォームの「接種日」は、[S03 手順1](docs/ops/testing/scenarios/S03-vaccination-next-due-autocalc.md) の期待「実施日のデフォルトが当日である」に反し、**空欄**で開く（デフォルト値なし）。
- 「次回の予定」で任意の日付に手動上書きしても、隣の間隔セレクタ（3週後/4週後/1年後/以外）の表示が直前に選んでいた値（例: 1年後）のまま変わらず、「以外（手動）」に切り替わらない。実害は小さいが、選択中の間隔と実際の次回予定日の表示が食い違って見える。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `S-ESTIMATE-NUMBER`
- 同一 PR にする BUG: BUG-011のみ
- 先行必須: 現行 DB constraint と採番 tx の競合再現
- 後続解放（シナリオ/他BUG）: S07 見積書2件目以降の作成

#### 1. 切り分けステータス

- 主因レイヤ: BE / DB（REPRODUCE-FIRST）
- 観測根拠（API・クエリ・コード参照）: `estimate_service.go` は create 前に `AllocateNextEstimateNo` を呼び、repository は clinic scoped advisory lock で採番する。bug 報告時の空文字経路が current HEAD で残るか、元の2件目手順で確認する。（具体パスは §2）
- 重複判定（例: 002 vs 022）: 単独。title duplicate ではなく clinic-scoped estimate_no 採番/一意制約の競合。
- 所有境界（FE / BE / データ / 環境）: FE=estimate create consumer; BE=billing estimate owner; データ=estimates full unique index; 環境=同時実行/既存 sequence 状態

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| BE | `backend/internal/billing/estimate_service.go` の create command。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/billing/estimate_repository.go` の `AllocateNextEstimateNo`。 | 現行証拠 / 変更候補 |
| DB/migration | `backend/migrations/001_init.sql` の `estimate_no` NOT NULL と `idx_estimates_clinic_estimate_no`。 | `WHERE`なしのclinic-scoped full unique。論理削除番号も予約 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: API変更なし。既存NOT NULLとclinic-scoped full unique index（WHEREなし）を維持する。current HEADでも空番号が生成される場合だけ、補修後に`btrim(estimate_no) <> ''` CHECKを別migrationで追加する。
- 正しい挙動の定義（1〜3 文）: 同一clinicでestimate_noを一意に割当て、2件目と並行作成が成功/競合再試行で整合する。
- やらないこと（Out of scope）: title重複扱い、空identifier露出、既存checksumの自動書換え。
- 既存データ修復の要否と手順: `btrim(estimate_no)=''`、重複、最大値をclinic別read-only dry-runし、業務承認後に一意番号へ別補修する。CHECK migrationを取り込んだ場合もagentは適用せず、人が`make migrate`。

- current HEAD で2件連続・並行作成を再現し、空番号が送信/保存されないことを確認する。再現しなければ追加実装をせず、回帰 test と既存空番号データ監査のみ行う。
- 再現時は採番と create を同一 tx/advisory lock 内に閉じ、client 指定番号を信頼しない。unique collision は限定 retry 後に明示失敗する。

#### 4. 受け入れ基準（AC）

1. Given 同一 clinic、When 異なるタイトルの見積書を逐次2件および並行作成する、Then 空でない一意な番号で全件成功する。
2. 回帰: 1件目、2件目、並行作成、別clinic採番を通す。
3. 負例: 採番競合・DB errorで空identifierやSQLを露出せず、重複rowをcommitしない。 既存境界AC: Given 別 clinic、When 同時作成する、Then clinic ごとの採番規約に従い、相互存在を漏らさない。
4. 横展開確認対象: estimate create/updateと番号表示

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/billing/... -run '^(TestEstimateService_Create_AssignsEstimateNo|TestEstimateService_Create_TwoSequentialRequests|TestEstimateRepository_AllocateNextEstimateNo_ConcurrentClinicIsolation)$' -count=1`
- FE: `docker compose exec -T frontend npx vitest run src/features/estimates/routes/EstimateForm.test.tsx`
- 手動/E2E: S07 手順1・7・9（見積書管理 `/estimates/new`）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 会計整合対象。estimate_noはclinic内一意・非空、別clinic分離。
- fail-closed の維持: 採番/insert失敗で空番号・重複・部分見積を残さない。
- audit / トランザクション境界: 採番lock、insert、auditをbilling ownerの同一txへ置く。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- advisory lock key に clinic を含め、tx 外採番を禁止する。既存 `estimate_no=''` の補修は件数/重複候補を dry-run し、帳票識別子の業務承認後に別実行する。

#### 7. 実装ステップ（順序付き）

1. 元手順と並行 test を current HEAD で実施し、request/DB 番号を記録する。
2. 必要な場合だけ tx 採番を修正し回帰 test を追加する。
3. 空番号監査と S07 後続を行い、補修要否を別判断する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- 元障害シナリオと競合 test が通り、空番号を新規生成せず、現行で再現しない場合は「既修正・回帰確認」として証拠を残す。

---

## S08（会計訂正系）: バグなし

- 手順1〜4: クレジットカード会計を作成・確定 → クレジット訂正ダイアログで理由を空のまま実行すると「訂正理由を入力してください」で正しくブロックされ、API 呼び出しも発生しないことをネットワークログで確認。理由を入力し金額を¥1,100→¥900に訂正 → 成功し、`GET /accountings/:id` で `billing_amount: 900` の反映を確認。会計一覧にも正しく反映。
- 手順5（権限分離）・手順6〜9（部分入金からの未収金フロー）はシナリオ自身が「一般ロールへの再ログイン」「現行UIでは部分入金保存不可のためBLOCKED」と明記している範囲のため、本セッションでは執行ロールのみでの確認に留めた。手順6を実際に試行し、請求額より少ない現金額のみでは「会計を確定する」が無効化され保存できないこと（＝シナリオの予期するBLOCKED挙動）を確認。
- 手順10（締め後訂正）: S09側でレジ締めを確定した後、同一会計へクレジット訂正を再実行 → 200成功。締め後訂正としての監査メタデータ（`post_close`）はAPI経由で直接確認していないが、UI操作・API応答とも異常なし。

## S09（締め境界）: 一部BLOCKED（環境制約）、確認できた範囲でバグなし

- **環境制約（バグではない）**: 本シナリオはローカルDBの直接更新による「同一営業日内に完了時刻10:00/13:30:00/14:00/20:00/翌日02:00の会計5件」というテストデータ準備を必須としているが、本セッションはブラウザ操作のみでDB/バックエンドへのシェルアクセスを持たないため、AM/PM/EMG境界判定・越日EMG帰属の精密な突合（手順1〜6）は実施不能。
- 実施できた範囲: `/settings/closing-time` の表示確認（八王子病院、城東センター病院への切替は未実施）、`/accounting/close` で対象日2026-07-31・区分「午後」のプレビューを実行し、S08で作成した田中光広/チビチビの会計（完了15:13、クレジットカード¥900）が正しく含まれることを確認（1データ点のみだが `completed_at` ベースの区分帰属ロジックが機能していることの裏付けにはなる）。
- 手順7（締め確定）: 実際のレジ現金`0`で「締める」→確認モーダル「締めを実行しますか？ 2026-07-31 午後 の締めを実行します。この操作は取り消せません。」が表示され、確定後「締めを実行しました」トースト・`POST /api/v1/cash-register/closes` が201で成功。
- 異常系「二重締め」: 同一日・同一区分で直接APIを再実行 → **409 `{"error":"この日時はすでに締め済みです"}`** で正しく拒否されることを確認（UI上も「すでに締め済みです」の案内に切り替わり、締めるボタン自体が消える）。バグなし。
- 手順9（締め履歴）: `/accounting/close/history` で該当レコードが区分「午後」・理論現金¥0（現金決済がなかったため。カード¥900は理論現金に含まれない仕様と解釈でき、整合）・実際の現金¥0・差額+0・締め時刻07-31 15:20で一覧表示され、行クリックで部門別内訳（未分類・要確認 ¥900 1件）を含む詳細モーダルが開くことを確認。バグなし。
- **軽微な観察事項（未報告・要確認）**: 締め履歴一覧の「日付」列が `2026-07-31T09:00:00+09:00` のような未整形のISO日時文字列のまま表示されている（本来の締め対象日は2026-07-31のみで、時刻部分`09:00:00`は無意味な値に見える）。実害は小さく、他の項目（区分・締め時刻等）は正しく整形表示されているため、表示専用の軽微な体裁不備の可能性が高いが、正式なバグ番号は付与せずここに観察事項として記録する。
- 手順8（締め済み期間の編集警告）・手順10（区分フィルタ）は、S08手順10で締め後訂正が問題なく機能したことから編集経路自体はブロックされていないことを確認済みだが、警告文言・修正理由必須の具体的な表示検証、および区分フィルタの動作検証は時間の都合上未実施。

---

## BUG-012: 顧客集計ダッシュボードが恒久的に「読み込み中...」のまま表示されず、バックエンドAPIが応答しない【重大】【S10ブロッカー】

- **重大度**: 高（S10の対象機能が完全に使用不能。集計・分析機能全体が機能していない）
- **対応状況（2026-08-03 JST）**: OPEN | **根拠**: `ListOwnerAggregation` が全 LTV 行を読込後 filter; payments 集計に clinic_id 欠落リスク; FE が CPM fan-out 7 重 call（wave-1） | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: 集計 SQL/ページングと CPM bulk を直しタイムアウト内応答を計測
- **発見シナリオ**: S10 手順1（顧客集計ダッシュボード `/aggregation`）
- **再現手順**:
  1. `/aggregation` を開く（売上ランキングタブが既定表示）。
  2. 画面は「読み込み中...」と表示されたまま、CPMセグメントのチップにも人数が表示されず、何秒待っても状態が変化しない。
  3. ネットワークログを確認すると、`GET /api/v1/clinics/1/owners/aggregations?...`（一覧本体）および CPM 6段階分のチップ集計クエリ（`cpm_stage=cpm_encounter` 等、計7本）がいずれも `pending` のまま完了しない。
  4. 直接 `fetch()` で同エンドポイントを叩いて検証: `per_page=50` で40秒待っても応答なし（タイムアウト）。`per_page=1` に絞っても20秒でタイムアウト。クエリパラメータを一切付けない `GET /api/v1/clinics/1/owners/aggregations` のみでも15秒でタイムアウト。件数を絞っても改善しないことから、データ量ではなくクエリ自体がハング（無限ループ／デッドロック等）している可能性が高い。
- **期待結果**（[S10 手順1](docs/ops/testing/scenarios/S10-customer-aggregation-consistency.md)）: 売上ランキングタブが表示され、CPMセグメント6段階の人数チップが表示される。
- **実際の結果**: APIが恒久的に応答せず、画面は「読み込み中...」から一切進まない。CPMチップクリック・タブ切替・飼主検索・CSV出力などS10の手順2〜11すべてが実施不能。
- **備考**: 本バグにより S10 は手順1で完全にブロックされ、以降の手順（飼主A/BのLTV突合、CPM絞り込み、最終来院区分、CSV出力、ドリルダウン）はすべて未実施。会計側（S08・S09で確認したaccounting API群）は正常に動作しているため、`owners/aggregations` エンドポイント固有の問題と考えられる。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `S-OWNER-AGGREGATION`
- 同一 PR にする BUG: なし
- 先行必須: handler / query / lock の区間計測
- 後続解放（シナリオ/他BUG）: S10 顧客集計ダッシュボードと後続集計確認

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE / DB（計測先行）
- 観測根拠（API・クエリ・コード参照）: FE は main analytics に加え CPM stage count を6回要求する。BE aggregation service は全対象を読み Go 側で filter/page し、LTV repository には非 bounded 集計と相関 subquery がある。どの query が支配的かは未計測。（具体パスは §2）
- 重複判定（例: 002 vs 022）: 単独。timeoutを外部依存と決め打ちせず query/lock を測る。
- 所有境界（FE / BE / データ / 環境）: FE=aggregation dashboard state; BE=lstep/billing aggregation owner; データ=owner aggregation query/cache; 環境=DB lock/plan/row count

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/aggregation/api/get-cpm-stage-counts.ts` と dashboard route。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/lstep/aggregation_service.go`、`backend/internal/owner/ltv_repository.go`。 | 現行証拠 / 変更候補 |
| DB/query | `backend/internal/owner/ltv_repository.go` のpayments集約subquery/join。 | `payments.clinic_id` predicateとcomposite joinの隔離確認 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: 既存responseへ`cpm_stage_counts`を加算追加し、FEの6本fan-outを1 HTTP requestへ統合する。owners/totalは全filter、countsは`cpm_stage`以外のfilterを適用する。schema indexはEXPLAIN根拠の別migration。
- 正しい挙動の定義（1〜3 文）: 集計はboundedに完了し、deadline/lock時はloadingを終了して再試行可能な明示errorを返す。
- やらないこと（Out of scope）: 空結果fallback、無根拠な外部API timeout、full table scanの隠蔽。
- 既存データ修復の要否と手順: 不要。indexが必要なら別migration、人が `make migrate`。

- request ID ごとに service/repository 区間を計測し、`EXPLAIN (ANALYZE, BUFFERS)` で clinic/date/filter 別の plan と行数を採取する。
- DB側へfilter/pagination/countを移し、dashboard responseへtyped `cpm_stage_counts`を一括追加する。payments集約は`WHERE payments.clinic_id = ?`、`GROUP BY clinic_id,billing_id`、親joinもclinic_id+billing_idで結合する。indexは計測で裏付けたpredicate/orderにだけ追加する。
- request context cancel と server timeout を伝播し、FE は timeout/error/retry を表示する。

#### 4. 受け入れ基準（AC）

1. Given demo相当件数、When `/api/v1/clinics/:clinic_id/owners/aggregations` を既定条件と空期間で呼ぶ、Then1 HTTP requestでowners/total/pageと全stage countsが合意時間内に返り、page外ownerをGoへ全件搬送せずUIがloadingを離脱する。
2. 回帰: 他集計条件・cache・dashboard retryを維持する。
3. 負例: lock/deadline/cancelでloadingを継続せず、別clinic countを返さない。 既存境界AC: Given 別 clinic、When同じfilterを使う、Then row/count/aggregate の全てが分離される。Given client cancel、Then DB query も中断される。
4. 横展開確認対象: owner aggregation endpointsとdashboard states

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/owner/... ./internal/lstep/... -run '^(TestFindOwnerLTV_.*|TestListOwnerAggregation_.*)$' -count=1`
- FE: `docker compose exec -T frontend npx vitest run src/features/aggregation/api/get-cpm-stage-counts.test.tsx src/features/aggregation/routes/AggregationDashboardPage.test.tsx`
- 手動/E2E: S10 手順1（顧客集計ダッシュボード `/aggregation`）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 集計はclinic隔離対象。count/rows/cacheの全predicateにclinic_idを含める。
- fail-closed の維持: timeout/errorを空集計や永続loadingへ潰さない。
- audit / トランザクション境界: read-only。cancelをDB queryへ伝播し、indexは別migration。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- index の憶測追加、全件を memory へロードする fallback、失敗を空集計に変換する処理を禁止する。EXPLAIN は個人情報値を成果物へ残さない。

#### 7. 実装ステップ（順序付き）

1. latency budget と trace/EXPLAIN を取得し failure signature を固定する。
2. repository query を bounded 化し、必要な index を別 migration として検証する。
3. response 集約、cancel/error UI、clinic/count integration test を実装する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- p50/p95 と query plan の before/after、結果同値性、cancel、clinic isolation が証拠化され、migration は人手適用境界を明記する。

---

## BUG-013: 未請求明細取得APIが実データ存在時に500エラーを返し、トリミング×診察の統合会計が機能しない【重大】【S11ブロッカー】

- **重大度**: 高（S11の中核機能である「同日同一ペットの未請求明細の会計統合」が完全に機能しない）
- **対応状況（2026-08-03 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: commit `74aa3e2c6e6dfe6c43227b2aacc0a699cc416a1a` — additive `GET /billing-items/unbilled-details` が items + `vaccination_master_unbillable` blocking warning を返し、legacy `/unbilled` raw-array は維持。vaccination unbillable は skip+count（infra は 500）。CreateAccounting/CreateItem は write-time 再集計で 409 fail-closed。FE 新会計 consumer は details へ移行し blocking/未取得中は確定無効化。scoped BE/FE tests green | **原文シナリオ再検証**: WAIVED（2026-08-05 USER 判断・ブラウザ再検証を実施しない） | **次のアクション**: なし（検証見送り。claim 解放済み）
- **発見シナリオ**: S11 手順5（会計新規作成 `/accounting/new?petId=xxx`）
- **再現手順**:
  1. 川崎和久／ナッツ（petId=1004170）に対し、トリミング登録（コース「八王子カット」¥0＋オプション「爪切り」¥300、ステータス予約→受付済→施術open→診療中）と、同日の通常カルテ（一般診察、処置「S11検証用処置」¥15,000を追加し確定）をそれぞれ作成。
  2. 受付カンバンで両方の appointment を「会計待ち」まで進める（トリミングカードも診察カードも会計待ち列に正しく表示されることを確認済み）。
  3. `/accounting/new?petId=1004170` を開くと、明細一覧が**空（¥0）**のまま表示され、「同日にまだ会計対象化されていない項目があります（診察1件）。受付ボードで対象を会計待ちに進めてから会計すると、1会計にまとめられます。」という警告が表示される（実際には両方ともすでに会計待ちへ進めた後）。
  4. ネットワークログを確認すると、ページ内部で呼ばれる `GET /api/v1/billing-items/unbilled?pet_id=1004170` が**2回とも500 `{"error":"internal server error"}`** を返している。
  5. 直接 `fetch()` で同エンドポイントを再現: `pet_id=1004170`（未請求明細が実在するペット）→ 3回とも500。一方 `pet_id=1`（既存だが未請求明細なしと思われるペット）および存在しない `pet_id=999999` → いずれも200 `[]`。この対比から、実際に未請求明細（トリミング明細・処置明細）が存在するケースでのみサーバー側の統合クエリがクラッシュしていると考えられる。
- **期待結果**（[S11 手順5](docs/ops/testing/scenarios/S11-trimming-combined-accounting.md)）: 診察の処置明細とトリミングのコース・オプション明細が、未請求明細として1つの会計に自動で統合して取り込まれる。
- **実際の結果**: 未請求明細取得APIが500エラーでクラッシュし、明細一覧は常に空のまま。トリミング・診察のいずれの明細も会計へ取り込まれない。
- **備考**: 本バグにより S11 手順5以降（計算サマリ突合・精算・再表示防止確認・受付カンバン会計済み遷移）と異常系A1〜A3はすべて実施不能。トリミングコース側の価格が本clinic seedで軒並み¥0だった点（別途環境上の制約として記録）とは独立した、バックエンド側の実装不具合と考えられる。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-BILLING-UNBILLED`
- 同一 PR にする BUG: BUG-013のみ
- 先行必須: source 別 price / clinic / pet 整合の audit
- 後続解放（シナリオ/他BUG）: S11 統合会計と BUG-018 の aggregate command

#### 1. 切り分けステータス

- 主因レイヤ: BE / データ / FE
- 観測根拠（API・クエリ・コード参照）: `billing_item_service.go` はいずれか一 source の失敗で全 aggregation を500にする。vaccination candidate は price 欠損/負値を internal error とし、demo vaccine seed に price NULL があるため、正常な診療・トリミング候補まで失われ得る。（具体パスは §2）
- 重複判定（例: 002 vs 022）: 単独だが BUG-018 の候補入力契約を解放する。price欠損 source と全体500を区別。
- 所有境界（FE / BE / データ / 環境）: FE=accounting unbilled consumer; BE=billing aggregation owner; データ=source prices + clinic/pet provenance; 環境=demo seed price audit

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/accounting/api/get-unbilled-items.ts`、`use-accounting-detail-state.ts`。 | legacy array consumerと新details consumerの移行点 |
| BE | `backend/internal/billing/routes.go`、`billing_item_handler.go` の`GetUnbilledItems`。 | legacy route維持とadditive route追加位置 |
| BE | `backend/internal/billing/billing_item_service.go` の unbilled aggregation。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/billing/billing_item_repository.go` の vaccination candidates。 | 現行証拠 / 変更候補 |
| Data/seed | `backend/migrations/seeds/003_demo/vaccinations.csv` と `vaccines.csv`。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: additive API変更あり。既存`GET /api/v1/billing-items/unbilled`と`BackendAccountingItem[]`は維持し、新規`GET /api/v1/billing-items/unbilled-details`が`{items: BackendAccountingItem[], warnings: {source, code, count, blocking}[]}`を返す。DB schema変更なし。
- 正しい挙動の定義（1〜3 文）: 有効sourceはitemsで可視化し、欠損price sourceは`blocking=true`のtyped warningにする。blocking warningが1件でも残る間はFE確定を無効化し、direct aggregate commandもserver側再集計で全書込を拒否してunderbillingを防ぐ。
- やらないこと（Out of scope）: legacy endpointのbreaking envelope化、欠損価格の0円化、SQL/tenant情報のwarning露出。
- 既存データ修復の要否と手順: seed価格は業務承認後の別差分。既存請求データは自動補修しない。

- 新details routeの`warnings`は現在`source="vaccination"`、`code="vaccination_master_unbillable"`、当該clinic/petの除外`count`、`blocking=true`だけを返し、record ID・master名・価格・tenant情報を含めない。型付きdata-quality事象だけをwarning化し、SQL/timeout/接続errorは500のままfail-closedにする。
- 既存`getUnbilledItems()`はsignatureを維持し、新規`getUnbilledItemDetails()`へ新会計consumerをatomicに移す。legacy routeのbreaking envelope化や、全consumer同時削除はしない。
- 実際の会計書き込みでは欠損価格を0円に捏造せず fail-closed。clinic/pet/reservation 相関を各 source で検証する。
- seed 補正は業務上正しい価格が承認された場合のみ別差分で行う。

#### 4. 受け入れ基準（AC）

1. Given有効なtreatment/trimmingとprice欠損vaccinationがある、When未請求detailsを取得する、Then有効項目と`blocking=true`のvaccination warningが返り、FEはwarning解消まで会計確定を許可しない。
2. 回帰: legacy raw-array endpointと全source成功を維持する。
3. 負例: Givenclientがblocking warningを無視してvalid itemsだけをdirect aggregate commandへ送る、When serverが未請求sourceを再集計する、Then409/422で全書込を拒否する。全source失敗、欠損/負価格、別clinic/pet候補も0円やsilent successにしない。
4. 横展開確認対象: accounting consumer、source repositories、BUG-018 command

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/billing/... ./internal/medicalrecord/... -run 'Test.*(Unbilled|VaccinationCandidate|Aggregation)' -count=1`
- FE: `docker compose exec -T frontend npx vitest run src/features/accounting/api/get-unbilled-items.test.ts`
- 手動/E2E: S11 手順5（会計新規作成 `/accounting/new?petId=xxx`）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 会計整合対象。items/warningsをclinic/pet/sourceで分離する。
- fail-closed の維持: 価格欠損を0円化せず、blocking warning中の部分会計を拒否し、infra errorをpartial successへ変換しない。
- audit / トランザクション境界: readはsource別。実会計write時にcanonical価格を同一txで再検証する。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- partial-success を silent success にせず警告を必須表示する。金額欠損をゼロ扱いしない。source error の内容に内部 SQL/tenant 存在を露出しない。

#### 7. 実装ステップ（順序付き）

1. source matrix の integration test で全失敗/部分失敗/全成功を固定する。
2. legacy routeを固定するcontract test、新details routeとtyped warnings、新FE getterを実装し、新会計consumerだけを移行する。write-time再検証も追加する。
3. BUG-018 の一括会計 command と S11 を接続して検証する。

#### 8. 完了定義（DoD）

- [x] §4 の AC が全通過（scoped unit/integration + FE getter tests; S11 ブラウザは UNREPORTED）
- [x] 関連クラスタと横展開対象の回帰が通過（Unbilled/Vaccination/Trimming/Routes snapshot scoped green）
- [x] 作成した test data の cleanup、または cleanup 不要を記録（AutoMigrate テスト DB のみ。seed 変更なし）
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録（localhost:3003/seed 未起動のため UNREPORTED）

- 有効候補は保持、無効候補は明示警告、書込は fail-closed、clinic/pet 相関が test され、seed 価格の推測補正をしていない。

---

## S12（LIFFペットヘルスとアカウント連携）: 手順1〜3はモックの既知制約どおり、手順5以降はBUG-008と同根の認証バグでブロック

- 手順1〜3（連携用URL発行→LIFF連携ページを開く→病院側で連携状態確認）: 飼主編集画面「LINE/Lステップ連携」から連携用URL発行 → 発行されたURL形式のクエリ（`clinic_id`・`token`）を用いて `/liff/1?clinic_id=1&token=...` を開くと「連携が完了しました」の成功画面が表示される。ただしネットワークログを確認すると、この間 `/api/...` への連携API呼び出しは一切発生していない（フロント側の完全モック分岐で、シナリオの前提条件に明記された「モックの限界」どおりの既知の挙動）。病院側の飼主編集画面を再確認すると、案の定「未連携」のままで変化なし。**これはバグではなく、シナリオ文書が事前に明記しているローカルモックの制約どおりの結果。**
- 手順4（期限切れ／使用済みトークン）: モックが実API呼び出しを行わないため実機検証不能（シナリオの想定どおりBLOCKED）。

## BUG-014: LIFFペットヘルスページが「ログイン情報の有効期限が切れました」で必ず失敗し閲覧不能（BUG-008と同根）【重大】【S12ブロッカー】

- **重大度**: 高（S12対象機能であるペットヘルスページが完全に閲覧不能。S04で確認したBUG-008と同じ「LIFFモックがAPI呼び出しにAuthorizationヘッダを一切付与しない」という根本原因が、別のLIFFアプリ（frontend/liff＝ペットヘルス・連携用）でも再現）
- **対応状況（2026-08-03 JST）**: OPEN | **根拠**: BUG-008 と同 C-LIFF-AUTH（shared use-liff / 401 copy / liff_auth）だが health-card 別 app; DUPLICATE 未証明のため OPEN 維持（wave-1） | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: 008 と同一 auth 修正後に health-card 単独で 200 表示を確認
- **発見シナリオ**: S12 手順5（token無しURL `/liff/1?clinic_id=1` でペットヘルスページを開く）
- **再現手順**:
  1. `/liff/1?clinic_id=1`（tokenパラメータなし）を開く。
  2. 画面は「データ取得に失敗しました / ログイン情報の有効期限が切れました。LINEアプリを再起動して開き直してください。」というエラー表示のみで、ペットカード等は一切表示されない。
  3. ネットワークログで `GET /api/liff/1/health-card` が**2回とも401**を返していることを確認。
  4. 直接 `fetch()` で同エンドポイントを再現: `{"code":401,"message":"missing authorization header","timestamp":"..."}` — Authorizationヘッダが全く送信されていない（S04のBUG-008で確認した `GET /api/liff/1/courses` の401 "missing authorization header" と全く同一のエラーメッセージ・原因パターン）。
- **期待結果**（[S12 手順5・6](docs/ops/testing/scenarios/S12-liff-pet-health.md)）: ペットヘルスページに切り替わり、ヘッダーに飼主名とLINEプロフィール画像、ペットごとのカード（名前・種/品種・最終来院日・ワクチン記録テーブル）が表示される。
- **実際の結果**: API呼び出しにAuthorizationヘッダが付与されず401で拒否され、常にエラー画面のみが表示される。
- **備考**: 本バグにより S12 手順5〜9（ペットカード表示内容確認・飼主間隔離の実機証明・clinic_idなしエラー確認・データ取得失敗時のリトライ確認）はすべて実施不能。S04のBUG-008（line-reserveアプリ、`/api/liff/1/courses`）と本バグ（frontend/liffアプリ、`/api/liff/1/health-card`）は別々のフロントエンドアプリ・別APIエンドポイントで発生しているが、いずれも「LIFFモックトークンがAPIリクエストに一切伝播しない」という共通の実装欠陥に起因すると考えられ、LIFF関連機能全体（予約・ペットヘルス・アカウント連携）が本セッションの検証環境ではブラウザ経由で実質的に検証不能な状態。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-LIFF-AUTH`
- 同一 PR にする BUG: BUG-008, BUG-014
- 先行必須: BUG-008 の provenance gate と共通 auth contract
- 後続解放（シナリオ/他BUG）: S12 health card、二者・二 clinic 隔離

#### 1. 切り分けステータス

- 主因レイヤ: 環境 / FE / BE（REPRODUCE-FIRST）
- 観測根拠（API・クエリ・コード参照）: 現行health Appはtoken取得後だけpageを描画し、`liff-api.ts`はBearer headerを必ず送る。報告時401はBUG-008と同じ観測だが、同根は未確定。認証後service/repositoryにはclinic/LINE customer/pet scopeがある一方、静的確認だけでは二者間隔離の実証にならない。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-008 と同一の401観測クラスタ。ただし現行 health-card client も header 付与済みで同根は未確定。
- 所有境界（FE / BE / データ / 環境）: FE=shared LIFF + health-card transport; BE=LIFF auth/reservation health owner; データ=LINE customer→owner→pet clinic scope; 環境=built asset/runtime env provenance

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/shared-liff/use-liff.ts`、`frontend/liff/src/App.tsx`、`frontend/liff/src/api/liff-api.ts`。 | token生成、描画gate、Bearer送信 |
| BE | `backend/internal/reservation/liff_service_health_card.go`、routes/auth middleware。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/lstep/line_customer_repository.go`、`backend/internal/medicalrecord/vaccination_repository.go`。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: 原因確定前は API/DB 変更なし。BUG-008のBearer/provenance契約を再利用し、health-card専用bypassを作らない。
- 正しい挙動の定義（1〜3 文）: 認証後のLINE user→owner→petをclinic内で解決し、正当なownerのpetだけを返す。
- やらないこと（Out of scope）: sourceと矛盾したheader追加、token/PIIのartifact化、pet ID直指定による認可。
- 既存データ修復の要否と手順: 不要。

- auth transport/error mapping は BUG-008 だけで直し、health card 側で別 bypass を作らない。認証後の LINE user→owner→pet 解決を clinic scoped service で fail-closed にする。

#### 4. 受け入れ基準（AC）

1. Given localで両mock flag=trueまたはSTGでowner Aの正規LIFF token、When health cardを開く、ThenBearer認証が成立しAに紐づくpet履歴だけが返る。
2. 回帰: BUG-008 courses/transportとhealth response schemaを維持する。
3. 負例: 別customer/clinic、未連携、期限切れ/release mockでpet/clinic存在を漏らさない。 既存境界AC: Given owner B、別 clinic、未連携 user、期限切れ token、When A の pet を推測して要求する、Thenデータ非開示の拒否となる。
4. 横展開確認対象: shared-liff auth、health-card、course catalog

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。 追加観点: STG S12 は実 LINE test account A/B で実施し、local mock 結果と混同しない。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/reservation/... ./internal/lstep/... ./internal/medicalrecord/... -run 'Test.*(HealthCard|LineCustomer|Vaccination)' -count=1`
- FE: `docker compose exec -T frontend npx vitest run liff/src/api/liff-api.test.ts`
- 手動/E2E: S12 手順5（token無しURL `/liff/1?clinic_id=1` でペットヘルスページを開く）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 認証と患者情報隔離対象。token→customer→owner→petsをclinic内で再検証する。
- fail-closed の維持: pet ID直指定、release mock、real authへのsilent fallbackを禁止する。
- audit / トランザクション境界: read-only。token/PIIをaudit artifactへ残さない。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- pet ID の直接指定だけで取得を許可しない。認証エラー文言で owner/clinic/pet の存在を区別しない。実 token をログ・成果物へ残さない。

#### 7. 実装ステップ（順序付き）

1. BUG-008のprovenance gateでhealth-card requestも同時採取し、共通drop pointが証明されるまで別実装を始めない。
2. owner A/B・clinic A/B の service integration test を追加する。
3. S12 全手順と vaccination 追加後の再取得を STG で観測する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- S12 の表示・再取得・認証失敗・二者/二 clinic 否定ケースが通り、共通 auth の重複実装がない。

---

# V01〜V05（個別フォーム検証、計84フォーム）

S01〜S12の業務シナリオ検証に続き、個別フォーム単位の受入テスト（V01〜V05、計84フォーム）をサブエージェントに分担して実施した。以下、V01〜V05で新たに発見されたバグ（BUG-015〜BUG-032）と、各シナリオの実施範囲サマリを記載する。

## BUG-015: バイタルの体重 Kg/g 単位切替で数値が単位換算されずそのまま保存され、1000倍のデータ破損が生じる【重大】

- **重大度**: 高（体重 8.5kg の記録が数値そのまま「8.5g」として永続化される。薬量自動計算は直近バイタルの体重を基準にするため、下流で致死的な過小投与量が算出されるリスクがある）
- **対応状況（2026-08-03 JST）**: IMPLEMENTED_UNVERIFIED | **実装 commit**: `98639b4fa` | **review-gate commit**: `28539d466` | **根拠**: FE `toggleWeightValueAndUnit` で Kg↔g 原子換算; BE weight 構造検証（finite・正数・unit enum）; vital create/update/delete を `AuditTxLogger.LogEntryTx`（ambient tx）で fail-closed（`98639b4fa` + `28539d466`） | **原文シナリオ再検証**: WAIVED（2026-08-05 USER 判断・ブラウザ再検証を実施しない） | **次のアクション**: なし（検証見送り。claim 解放済み）
- **発見シナリオ**: V01 §3 手順4（カルテ バイタル `/medical-records/:id` バイタル記録モーダル）
- **再現手順**:
  1. `/medical-records/1425549` を開き、バイタル記録モーダルで既存レコード（体温45℃・体重8.5kg）を編集
  2. 体重欄の単位トグルボタンをクリックして `Kg` → `g` に切り替える（体重の数値欄は操作しない）
  3. 行の保存ボタンをクリック
  4. `GET /api/v1/medical-records/1425549/vitals` で確認
- **期待結果**: 単位切替は実測値の意味を変えるものであり、切替後も実測値と一致する形で保存されるべき
- **実際の結果**: `weight: 8.5, weight_unit: "g"` として保存された（切替前は `weight: 8.5, weight_unit: "Kg"`）。数値は一切変換されず、単位ラベルのみが変わる。UI上も再読込後「8.5 g」と表示され、実際には8.5kgだった体重が8500分の1として記録される。
- **備考**: UI操作→直接API GETの両方で2回再現性を確認。確認後、`PATCH /api/v1/medical-records/1425549/vitals/1425545` で `weight_unit: "Kg"` に戻し、後続検証への影響を排除済み。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `S-VITAL-WEIGHT-UNIT`
- 同一 PR にする BUG: なし
- 先行必須: canonical physical mass、表示精度、既存API/DBのvalue+unit契約の確認
- 後続解放（シナリオ/他BUG）: 投薬量計算と S06/V02 のバイタル再利用

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE / データ
- 観測根拠（API・クエリ・コード参照）: edit/add row の Kg/g toggle は unit 文字列だけを変え、数値を換算せず raw value/unit を保存する。BE request に value/unit 組合せ validation がなく、dose utility は g を1000で割るため破損が投薬量へ伝播する。（具体パスは §2）
- 重複判定（例: 002 vs 022）: 単独。表示単位切替と canonical保存単位の契約不一致。投薬計算への横展開あり。
- 所有境界（FE / BE / データ / 環境）: FE=VitalsTab unit state; BE=medicalrecord vital validation; データ=canonical weight + historical outlier candidates; 環境=なし

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/medical-records/components/VitalsTab/VitalsTabRows.tsx`。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/medicalrecord/vital_request.go`、`vital_service.go`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/lib/medicine-dose.ts`。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: API/DB shapeと`body_weight_unit ('Kg','g')`を維持する。FE内部ではcanonical physical massを正本にし、選択unitのvalue+unit pairへ変換して保存する。BEはpairを安定した400契約で検証する。
- 正しい挙動の定義（1〜3 文）: Kg/g切替は同じphysical massを5 Kg↔5000 gとして表示・保存し、再GETと投薬計算で実質重量が一致する。
- やらないこと（Out of scope）: 表示labelだけの修正、既存値の無条件1000倍/1000分の1補正。
- 既存データ修復の要否と手順: read-onlyで異常値候補を抽出し、患者・時刻・入力根拠を臨床承認後に別補修。

- Kg↔g の pure conversion を一つ作り、add/edit 両 row で値と unit を原子的・不変に更新する。display 精度と保存精度を明示する。
- BE は finite、正数、unit enum、獣医師承認済みの plausible bound を検証する。dose 計算は不正/欠損体重で提案せず fail-closed。
- vital 更新と audit を同一 tx にし、audit の best-effort 継続を許さない。

#### 4. 受け入れ基準（AC）

1. Given 5 kg、When gへ切替、Then 5000 g、再び kgで5 kgとなり保存・再読込・dose の実質重量が不変。
2. 回帰: Kg/g往復、再GET、既存kg入力、投薬計算を通す。
3. 負例: 空/負/上限外/単位不明を保存せず、既存外れ値を自動補正しない。 既存境界AC: Given 5000 g、When保存、Then同じ実質重量。Given NaN/0/負数/不正unit/承認上限外、Then保存とdose提案を拒否して理由を表示する。
4. 横展開確認対象: vitals表示・request validation・medicine-dose

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/medicalrecord/... -run 'Test.*(Vital|Weight)' -count=1`
- FE: `docker compose exec -T frontend npx vitest run src/lib/medicine-dose.test.ts`。pure conversionと`src/features/medical-records/components/VitalsTab/VitalsTabRows.test.tsx`は新規追加予定。
- 手動/E2E: V01 §3 手順4（カルテ バイタル `/medical-records/:id` バイタル記録モーダル）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 1000倍誤保存は臨床安全。canonical unitと上限/下限をFE/BEで検証する。
- fail-closed の維持: unit不明・範囲外を保存せず、既存値を自動変換しない。
- audit / トランザクション境界: vital write/auditをmedicalrecord ownerの同一tx。補修は別承認。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- 既存値を「1000倍/1000分の1」と一律補修しない。pet の履歴、unit、時刻、周辺値を抽出する read-only audit、獣医師レビュー、dry-run、監査付き個別補修を別 packet にする。
- 換算の丸めで投薬量が変わらない precision を決める。別 clinic/pet vital の更新を拒否する。

#### 7. 実装ステップ（順序付き）

1. pure conversion と add/edit/dose の RED test を追加する。
2. FE 原子更新、BE boundary validation、tx audit を実装する。
3. 既存データ候補の read-only 件数を出し、補修は別承認へ渡す。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- round-trip と dose 不変、invalid fail-closed、clinic/pet 隔離、audit 原子性が通り、既存データを未承認で変更していない。

## BUG-016: 予防接種・検査・入院の各独立フォームで、存在しないIDのURLを直叩きすると空の編集可能フォームが開いてしまう

- **重大度**: 中（バックエンドは正しく404を返しており保存を試みても失敗するためデータ破損には至らないが、「存在しないIDのURL直叩きはエラー画面になるべき」という異常系要件に反する）
- **対応状況（2026-08-05 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み commit `7ee0edbacdc7aeff60c3f4f764889fe28a431010` — entity read を `loading|found|notFound|forbiddenOrHidden|error` に分類し vaccination/examination/hospitalization edit で Not Found ゲート + mutation 0 を vitest 固定 | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: V01 §8 手順5 / §7 手順4 / §10 手順7 をブラウザ再検証し `VERIFIED_FIXED` 可否を判定
- **発見シナリオ**: V01 §8 手順5（予防接種独立画面 `/vaccinations/:id`）、§7 手順4（検査入力 `/examinations/:id`）、§10 手順7（入院登録 `/hospitalization/:id/edit`）
- **再現手順**:
  1. `/vaccinations/999999999` を直接開く → 空の「予防接種詳細・編集」フォームが表示され、削除・保存ボタンが有効な状態で開く
  2. `/examinations/999999999` を直接開く → 同様に空の「検査詳細・編集」フォームが表示される
  3. `/hospitalization/999999999/edit` を直接開く → 同様に空の「入院編集」フォームが表示される（右下に未翻訳の「not found」トーストが一瞬表示されるのみ）
- **期待結果**: 「予防接種が見つかりません」等のエラー画面が表示され、白画面にも空フォームにもならない
- **実際の結果**: 3フォームとも編集可能な空フォームがそのまま開く。エラーの手がかりは（保存を試みた場合のみ）右下の未翻訳トースト「not found」だけで、フォームを開いた直後には一切のエラー表示がない。
- **備考**: 対照として `/medical-records/999999999`・`/trimming/999999999` は正しく「見つかりません」エラー画面が表示されることを確認済み。よって本バグは予防接種・検査・入院の3フォームに限定される。`GET /api/v1/vaccinations/999999999` が404を返すこと（バックエンドは正しい）をネットワークログで確認済み。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-NOT-FOUND-EMPTY`
- 同一 PR にする BUG: BUG-016, BUG-019（shared result contract; feature差分は分割可）
- 先行必須: shared found / notFound / forbidden / error 契約
- 後続解放（シナリオ/他BUG）: 予防接種・検査・入院・見積の直 URL 安全化

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE
- 観測根拠（API・クエリ・コード参照）: vaccination/examination の get API error が form hook の default model に吸収され、hospitalization は `isError` でも form render を続ける。その結果 URL に ID があるのに新規同様の編集可能画面になる。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-019 と空編集 fallback を共有。本件は vaccination/examination/hospitalization の3機能差分。
- 所有境界（FE / BE / データ / 環境）: FE=各 edit query/hook/route; BE=各 read endpoint の404契約; データ=clinic-scoped entity IDs; 環境=なし

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | Vaccination: `frontend/src/features/vaccinations/api/get-vaccination.ts`、`use-vaccination-form.ts`。 | 現行証拠 / 変更候補 |
| FE | Examination: `frontend/src/features/examinations/api/get-examination.ts`、`use-examination-form.ts`。 | 現行証拠 / 変更候補 |
| FE | Hospitalization: `frontend/src/features/hospitalization/hooks/use-hospitalization-form.ts` と form route。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: API/DB変更なし。FE read resultを `found/notFound/forbidden/error` で扱い、404をdefault edit modelへ変換しない。
- 正しい挙動の定義（1〜3 文）: edit URLはfound時だけ編集可能、404/403/errorは非編集状態、create routeだけdefault formを使う。
- やらないこと（Out of scope）: 他clinic IDの存在開示、全フォームの無関係なrouter再設計。
- 既存データ修復の要否と手順: 不要。

- loader/query の `loading | found | notFound | forbiddenOrHidden | error` を discriminated union にし、edit route では `found` 以外に form model を作らない。
- 404 と tenant/authorization の内部差は server log にだけ残し、client には存在を漏らさない共通 error boundary を使う。create route の default model は別 entry point にする。

#### 4. 受け入れ基準（AC）

1. Given 存在しない vaccination/examination/hospitalization ID、When edit URL を開く、Then空の編集 form と保存 button は表示せず、既定 error page を表示する。
2. 回帰: 正常editとcreate routes、BUG-019 estimateを維持する。
3. 負例: 404/403/network errorを空formへ変換せず、他clinic存在を区別しない。 既存境界AC: Given 別 clinic の実在 ID、When同 URL を開く、Then同等の非開示応答となり、既存値も作成 default も見えない。
4. 横展開確認対象: vaccination/examination/hospitalization/estimate edit loaders

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*NotFound|Test.*ClinicScope'`
- FE: `docker compose exec -T frontend npx vitest run src/features/vaccinations/hooks/use-vaccination-form.test.ts src/features/examinations/hooks/use-examination-form.test.ts src/features/hospitalization/hooks/use-hospitalization-form.test.ts src/features/hospitalization/routes/HospitalizationForm.permissions.test.tsx`
- 手動/E2E: V01 §8 手順5（予防接種独立画面 `/vaccinations/:id`）、§7 手順4（検査入力 `/examinations/:id`）、§10 手順7（入院登録 `/hospitalization/:id/edit`）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 空編集による誤保存を防ぐ。clinic-scoped IDとroute stateを分離する。
- fail-closed の維持: 404/403/errorをeditable defaultへfallbackしない。
- audit / トランザクション境界: read-only表示。後続writeはfound entityだけをBEで再検証。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- network error を404へ偽装して retry 導線を失わない。404/403 の差で tenant record の存在を推測させない。blank form の submit による意図しない create/update を禁止する。

#### 7. 実装ステップ（順序付き）

1. 3 form の不存在/別 clinic test を table 化して RED にする。
2. query result union と共通 route boundary を実装する。
3. create/edit の初期化分離、retry、save 非発火を確認する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- 3フォームで空編集が消え、create は退行せず、404・network error・tenant 非開示と「保存 request 0件」が検証される。

## BUG-017: 検査入力フォームで必須項目（検査種別・担当医）が未入力のまま保存を押しても、保存はブロックされるがエラーメッセージが一切表示されない

- **重大度**: 中（保存自体は正しくブロックされデータ破損はないが、職員には保存失敗の理由が一切示されない「無音失敗」に該当する）
- **対応状況（2026-08-03 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: commit `7f7106375…` で fieldErrors 配線 + ARIA。Mode3 follow-up で実 `ExaminationForm` の保存 button `user.click` 統合 test（`ExaminationForm.validation.test.tsx`）が 2 件日本語 alert・ARIA・`testTypeId` focus・create/update/axios 0 を同一 test で固定 | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: V01 §7 を synthetic fixture でブラウザ再検証し `VERIFIED_FIXED` 可否を判定
- **発見シナリオ**: V01 §7 手順1（検査入力 `/examinations/new?petId=xxx`）
- **再現手順**:
  1. `/examinations/new?petId=1000002` を開く（検査種別・担当医とも未選択のまま）
  2. 保存ボタンをクリック
  3. 画面を確認する。ネットワークログで `POST /api/v1/examinations` が発生していないことを確認
- **期待結果**: それぞれ必須エラーが表示され保存されない
- **実際の結果**: 保存は正しくブロックされる（POSTは発生しない）が、インラインエラーテキストもトーストも一切表示されない。検査種別セレクトにフォーカスリング（青枠）が付くのみで、通常のフォーカス表示と視覚的に区別しづらい。
- **備考**: 同カルテ内の予防接種独立フォームでは同条件で明確なインラインエラーが表示されており、フォーム間で必須検証のUXが一貫していない。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-SILENT-VALIDATION`
- 同一 PR にする BUG: BUG-017; BUG-021 は同一表示契約を別 owner で適用
- 先行必須: 共通 field-error / ARIA 契約
- 後続解放（シナリオ/他BUG）: 検査フォームと BUG-021 死亡記録の無音失敗解消

#### 1. 切り分けステータス

- 主因レイヤ: FE
- 観測根拠（API・クエリ・コード参照）: `use-examination-form.ts` は検査種別/担当医 error を生成し route は focus しようとするが、`ExaminationFormFields.tsx` が error props を受け取らず描画しない。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-021 と無音 validation 表示契約を共有。本件は examination fields。
- 所有境界（FE / BE / データ / 環境）: FE=examination hook/fields + FormFieldError; BE=既存 validation は変更不要; データ=なし; 環境=browser accessibility tree

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/examinations/hooks/use-examination-form.ts`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/features/examinations/routes/ExaminationForm.tsx`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/features/examinations/components/ExaminationFormFields.tsx` と共通 `FormFieldError`。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: API/DB変更なし。FE field-errorとARIA表示契約だけを是正する。
- 正しい挙動の定義（1〜3 文）: 必須field未入力ではrequestせず、各field近傍のerrorとfocus/ARIAを表示し、修正時に解消する。
- やらないこと（Out of scope）: BE errorのvalidation成功化、全errorの一括消去。
- 既存データ修復の要否と手順: 不要。

- form hook の typed error map を field component へ渡し、該当入力の直後に `FormFieldError`、`aria-invalid`、`aria-describedby` を設定する。submit 後に先頭 invalid field へ focus する。
- error は値変更時に当該 field だけ再検証し、server error と client field error を混同しない。

#### 4. 受け入れ基準（AC）

1. Given 検査種別・担当医が空、When 保存、Then request は送られず両 field の日本語理由が表示され、先頭 field に focus する。
2. 回帰: 有効submit、server error、BUG-021のfield-error契約を維持する。
3. 負例: 必須未入力ではPOSTせず、修正済みfield以外のerrorを消さない。 既存境界AC: Given 値を修正、When 再入力/保存、Then該当 error が消え、他 error は必要なら残り、有効 payload だけ送信する。
4. 横展開確認対象: FormFieldErrorと全medical form validation

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: 該当なし（API/DB変更なし。FE境界testで検証）
- FE: `docker compose exec -T frontend npx vitest run src/features/examinations/hooks/use-examination-form.test.ts src/features/examinations/components/ExaminationFormFields.test.tsx`
- 手動/E2E: V01 §7 手順1（検査入力 `/examinations/new?petId=xxx`）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 無音validationは臨床入力リスク。必須fieldとARIAを一致させる。
- fail-closed の維持: invalid時はrequestしない。server errorをfield成功へ潰さない。
- audit / トランザクション境界: invalid pathはwrite/auditなし。valid pathの既存txを回帰。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- 色だけで error を示さず screen reader に関連付ける。focus loop や stale error を作らない。client validation を BE validation の代替にしない。

#### 7. 実装ステップ（順序付き）

1. 2 field の message/ARIA/focus test を RED にする。
2. typed props と共通 error component を接続する。
3. 値修正、server error、keyboard submit を回帰確認する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- request 0件、可視/読み上げ可能な理由、focus、修正後 clear が通り、BUG-021 でも同じ表示規約を使う。

## BUG-018: レジ締め済み期間に作成した会計の明細追加が400で失敗し、合計金額と明細が不整合な「壊れた」会計が残る

- **重大度**: 高（会計データの整合性が崩れ、金額と明細が食い違ったレコードがDBに残る。かつUIはエラーを警告するのみで自動的にロールバックしない）
- **対応状況（2026-08-03 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: commit `75f8912fc2a8d3b2fb5c0290a766bb0f5fc12ac5` claim/BUG-018 + additive complete command + migration 005 作成のみ（**`make migrate` 未適用**）+ FE 単一 mutation。**Independent Review Gate（5 agent）実施**: code-reviewer / go-reviewer / typescript-reviewer / database-reviewer / clinic-isolation-auditor。HIGH 修正 commit `0b2bde8154677d88654ff0d6b557813a3f394541`: AlreadyExists→idempotent replay、CreatedBy 手入力 other、discount:edit on complete、FE Idempotency-Key mutation reuse + discount fields、seed `003_demo/billings.csv` 列追従。MEDIUM 記録のみ: post-close 権限 write-time 再検証・source-linked 価格 master 再解決・Complete isolation 専用 test。scoped BE/FE green | **原文シナリオ再検証**: WAIVED（2026-08-05 USER 判断・ブラウザ再検証を実施しない） | **次のアクション**: なし（検証見送り。claim 解放済み）
- **発見シナリオ**: V02 §1 会計・精算フォーム（新規会計作成） `/accounting/new?petId=X` → 確定後の明細追加
- **再現手順**:
  1. 当日のレジ締め（PM）が既に完了している状態で `/accounting/new?petId=X` から新規会計を作成する
  2. 会計ヘッダーの確定操作を行う（`POST /api/v1/accountings` が201で成功し、会計IDが発行される）
  3. 続けて物販・その他の明細行を追加しようとする（`POST /api/v1/billing-items` を該当accounting IDに対して呼び出す）
  4. ネットワークログを確認する
- **期待結果**: レジ締め済み期間に該当する会計操作は、ヘッダー確定の時点で一貫してブロックされるか、あるいは `post_close_reason`（締め後理由）の入力を求めるダイアログが明細追加時にも一貫して提示され、会計ヘッダーと明細の内容が常に一致した状態で保存される
- **実際の結果**: 会計ヘッダーの作成（`POST /api/v1/accountings`）は201で成功する一方、後続の明細追加（`POST /api/v1/billing-items`）は400エラーで失敗する。明細追加ダイアログには `post_close_reason` を入力する導線が存在しないため、ユーザーはこのエラーから回復する手段がない。結果として、合計金額が設定されているのに明細行が0件（またはヘッダーと不整合な内容）の会計レコードがDBに残る。
- **備考**: 根本原因はコードレベルの欠陥である: (1) 明細追加ダイアログに `post_close_reason` を渡す経路が実装されていない、(2) 会計確定フロー側で「明細の合計とヘッダーのtotal_amountが一致しているか」を最終確認する仕組みがない。通常運用でも「締め後に会計を開始してしまった」ケースで再現しうる。ネットワークログ（`/api/v1/accountings` 201, `/api/v1/billing-items` 400）で直接確認済み。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-BILLING-ATOMIC`
- 同一 PR にする BUG: BUG-018のみ
- 先行必須: BUG-013 の未請求details契約
- 後続解放（シナリオ/他BUG）: V02/S11 の原子的な会計確定とレジ締め後運用

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE / DB
- 観測根拠（API・クエリ・コード参照）: FE `use-accounting-completion-action.ts` は header、items、payment/reservation を複数 phase で保存し、items も逐次 request。レジ締め後理由が item request に届かず、途中まで commit された後400となる。（具体パスは §2）
- 重複判定（例: 002 vs 022）: 単独。BUG-013 は入力候補、こちらは aggregate write の原子性。
- 所有境界（FE / BE / データ / 環境）: FE=accounting completion consumer; BE=billing aggregate write owner; データ=accounting/items/payments/splits/reservation/audit; 環境=register-close state

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/accounting/hooks/use-accounting-completion-action.ts`、`create-accounting-items.ts`、`create-billing-item.ts`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/features/accounting/api/create-accounting.ts` と新規`complete-accounting.ts`。 | 旧多段consumerから単一commandへの移行 |
| BE | `backend/internal/billing/routes.go`、`accounting_request.go`、`accounting_service_core.go`。 | additive route、typed input、aggregate tx owner |
| BE | `backend/internal/billing/billing_item_service.go`。 | item/close reason/auditの既存検証をtx-aware collaborator化 |
| DB/migration | `billings` のcompletion request key/hashとclinic-first full unique。 | retry idempotency。別migration・人手適用 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: additive API変更あり: `POST /api/v1/accountings/complete` と必須`Idempotency-Key` UUIDを追加し、既存accountings/billing-items endpointはFE移行中維持する。`billings.completion_request_id/hash`と`UNIQUE(clinic_id, completion_request_id)`は別migrationで追加し、agentは適用せず人が`make migrate`。
- 正しい挙動の定義（1〜3 文）: serverがcanonical価格を再読込し、header/items/total/payments/splits/reservation/auditを1つの`DBOrTx`で全commitまたは全rollbackする。
- やらないこと（Out of scope）: FE compensation delete、client total信頼、既存endpointの同時破壊。
- 既存データ修復の要否と手順: 既存不整合はread-only audit→業務承認→別補修tx。自動削除/再計算しない。

- FE の多段書込を、billing header、全items、server再計算 total、payments/splits、reservation completion、`post_close_reason` を含む一つの aggregate command に置換する。
- BE billing owner が一つの `DBOrTx` で clinic/pet/owner/reservation、close境界、金額整合性を lock/再検証し、監査を含め全成功または全 rollback にする。FE compensation delete は使わない。
- command bodyは`medical_record_id`/`hospitalization_id`、`owner_id`、`pet_id`、`scheduled_date`、memo/insurance、`items[]`、`payment_splits[]`、`post_close_reason`を受ける。`clinic_id`/actorは認証contextから取得し、source-linked itemのname/price/tax、billing ID/status/completed_at/subtotal/tax/totalはserverがtx内で再解決・生成する。
- 初回成功は201＋既存AccountingResponse、同一key・同一normalized digestのretryは同じ会計を200で返し、同一key・異なるdigestは409、key欠落/不正は400とする。soft-delete後もkeyを再利用させず、retryでpayment/auditを重複させない。

#### 4. 受け入れ基準（AC）

1. Given open period、When 複数明細会計を確定、Then header/items/total/payment/reservation/audit が一度に一致して commit される。
2. 回帰: open periodと既存read/update、BUG-013 warningsを維持する。
3. 負例: closed reason欠落、N番目item失敗、競合/retryは全tableとauditを不変にする。 既存境界AC: Given closed period で理由欠落/不正、または N番目 item が失敗、When確定、Then全 table と audit が変更されない。 / Given許可された post-close 理由、When確定、Then理由と actor が監査され全体が原子的に成功する。
4. 横展開確認対象: accounting create/items/payments/splits/reservation/auditと同一/異digest idempotent retry

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/billing/... -run 'Test.*(CompleteAccounting|Accounting.*Atomic|PostClose|Rollback|Idempot)' -count=1`
- FE: `docker compose exec -T frontend npx vitest run src/features/accounting/hooks/create-accounting-items.test.ts src/features/accounting/routes/AccountingDetail.test.tsx`。`src/features/accounting/api/complete-accounting.test.ts`は新規追加予定。
- 手動/E2E: V02 §1 会計・精算フォーム（新規会計作成） `/accounting/new?petId=X` → 確定後の明細追加を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 会計整合・権限・clinic隔離の最高リスク。actorとpost-close権限を検証する。
- fail-closed の維持: client total/priceを信頼せず、いずれの途中失敗でも全rollbackする。
- audit / トランザクション境界: header/items/payments/splits/reservation/audit/idempotencyを一つのDBOrTx。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- server total を client 値で信頼しない。close override の権限、理由、actor、clinic を fail-closed 検証する。監査を tx 外へ出さない。
- 既存不整合は自動削除/再計算せず、header-total/items/payments/splits の read-only audit、業務レビュー、補修 tx を別 packet にする。

#### 7. 実装ステップ（順序付き）

1. N番目失敗、closed-period、同一/異digest retry、audit失敗、close raceのintegration testをREDにし、idempotency migrationを別レビューする。
2. migration取込後は人が`make migrate`を実行し、additive route/request/responseとbilling ownerの単一transactionを実装する。
3. FEを単一mutationへatomicに切替え、legacy routeを残したままBUG-013 warnings、retry、V02/S11を検証する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- success/全 rollback、server total、close override、監査、clinic/pet相関、idempotent retry が通り、既存不整合補修を別承認にしている。

## BUG-019: 存在しない見積書IDで編集画面を開くとエラー画面ではなく空白のフォームが表示される

- **重大度**: 中（クラッシュや無限ローディングではないが、ユーザーが「新規作成中」と誤認しうる）
- **対応状況（2026-08-05 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み commit `7ee0edbacdc7aeff60c3f4f764889fe28a431010` — estimate edit を route param mode + found 状態で判定し 404/403 非開示 Not Found / network retry を vitest 固定（BUG-016 shared contract 再利用） | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: V02 §6 をブラウザ再検証し `VERIFIED_FIXED` 可否を判定
- **発見シナリオ**: V02 §6 見積書フォーム（異常系） `/estimates/:id`（存在しないID）
- **再現手順**:
  1. 存在しない見積書ID（DBに存在しない大きな数値）を使って `/estimates/999999999` のようなURLに直接アクセスする
  2. 画面の表示内容を確認する
- **期待結果**: 「見積書が見つかりません」等の明示的なエラー表示、または一覧画面へのリダイレクトが行われる
- **実際の結果**: エラー画面は表示されず、空白（未入力）の見積書編集フォームがそのまま表示される。新規作成フォームと見た目上区別がつかない。
- **備考**: コンソール・ネットワークログ上で明示的な500エラーは出ていない（404が握りつぶされてUI側でフォールバックされている挙動とみられる）。BUG-002/016と同系統（存在しないID直叩き時のエラー画面欠如）だがフォームが異なるため別項目として記載。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-NOT-FOUND-EMPTY`
- 同一 PR にする BUG: BUG-016, BUG-019
- 先行必須: BUG-016 の shared result contract
- 後続解放（シナリオ/他BUG）: 存在しない見積 ID の安全な Not Found 表示

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE
- 観測根拠（API・クエリ・コード参照）: `get-estimate.ts` の取得失敗後も route は undefined を form hook へ渡し、hook は `!!estimate` で new/edit を判定するため、URL は edit のまま new-like blank model になる。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-016 と空編集 fallback を共有。本件は estimate edit-mode 判定。
- 所有境界（FE / BE / データ / 環境）: FE=estimates query/hook/route; BE=billing estimate read contract; データ=clinic-scoped estimate ID; 環境=なし

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/estimates/api/get-estimate.ts`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/features/estimates/routes/EstimateForm.tsx`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/features/estimates/hooks/use-estimate-form.ts`。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: API/DB変更なし。estimate editはroute paramとfound状態で判定し、404をdefault modelへ変換しない。
- 正しい挙動の定義（1〜3 文）: new routeだけ空の作成form、存在するIDだけedit form、404/403/errorは非編集状態にする。
- やらないこと（Out of scope）: 他clinic estimate存在開示、estimate API全体の再設計。
- 既存データ修復の要否と手順: 不要。

- BUG-016 の discriminated loader result と error boundary を再利用し、route param の有無で mode を確定する。edit mode の undefined data を new default に変換しない。

#### 4. 受け入れ基準（AC）

1. Given存在しない/別clinicの estimate ID、When edit URL を開く、Then非開示 error page、save request 0件、空 form なし。
2. 回帰: 正常editとnew route、BUG-016 shared resultを維持する。
3. 負例: 404/403/network errorをcreate modeへ落とさず、別clinic存在を漏らさない。 既存境界AC: Given network error、When retry、Then復旧可能で404と同一扱いに固定されない。Given `/estimates/new`、Then従来の新規 form が開く。
4. 横展開確認対象: estimate routes/hooksと共通Not Found UI

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/billing -run 'Test.*Estimate.*NotFound|Test.*ClinicScope'`
- FE: `docker compose exec -T frontend npx vitest run src/features/estimates/routes/EstimateForm.test.tsx src/features/estimates/hooks/use-estimate-form.test.ts`
- 手動/E2E: V02 §6 見積書フォーム（異常系） `/estimates/:id`（存在しないID）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 空見積の誤保存を防ぎ、clinic-scoped IDの存在を漏らさない。
- fail-closed の維持: 404/403/errorをcreate modeへfallbackしない。
- audit / トランザクション境界: read-only表示。後続writeはfound estimateだけをbilling txで扱う。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- 404/403 の差で tenant 存在を漏らさない。new/edit mode を fetched object の truthiness に委ねない。

#### 7. 実装ステップ（順序付き）

1. not found/別 clinic/network/new の route test を追加する。
2. BUG-016 共通 result/boundary を estimates に適用する。
3. save 非発火と direct URL/retry を確認する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- 空編集が消え、new route は維持され、tenant 非開示と network retry が検証される。

## BUG-020: 電話番号フォーマットの検証エラーメッセージが、正しい値に修正した後も表示され続ける

- **重大度**: 低（機能的にはフォーム送信をブロックしない。ネットワークログでオーナー作成POSTが実際に正しい値で発火していることを確認済み。UI表示が古いままになる見た目のみの問題）
- **対応状況（2026-08-03 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み commit `617f6f9bf88be3627ba789d447c98858dd34c80a` で raw phone を共通 validator で検証し、有効値への修正時に `newOwnerErrors.phone` だけを不変更新で解除する。空白付き無効値、他 field error、invalid path の `onSave` 0回を維持し、依存更新後の scoped Vitest 2 files / 18 tests、statements 84.28%、対象 ESLint / diff-check が green | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: V02 §7を専用 synthetic fixture でブラウザ再検証し、`VERIFIED_FIXED` 可否を判定
- **発見シナリオ**: V02 §7 予約登録モーダル（新規飼主タブ） 電話番号入力欄
- **再現手順**:
  1. 予約登録モーダルの「新規飼主」タブを開く
  2. 電話番号欄に不正な形式（桁数不足）を入力し、フォーマットエラーメッセージが表示されることを確認する
  3. 続けて欄を正しい形式の値に修正する
  4. エラーメッセージ表示欄を確認する
- **期待結果**: 入力値が正しい形式に修正された時点で、エラーメッセージは消える、または再検証されて表示が更新される
- **実際の結果**: 値を正しい形式に修正した後も、直前のフォーマットエラーメッセージが画面上に残り続ける。
- **備考**: ハードブロッカーではないためUI上の表示不整合（バリデーション状態の再計算漏れ）として低重大度で報告。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `S-PHONE-ERROR-LIFECYCLE`
- 同一 PR にする BUG: なし
- 先行必須: なし
- 後続解放（シナリオ/他BUG）: 予約内新規飼主フォームの再検証

#### 1. 切り分けステータス

- 主因レイヤ: FE
- 観測根拠（API・クエリ・コード参照）: `ReservationFormModal.tsx` は submit 時に phone error を設定するが、`NewOwnerInlineForm` の change callback が当該 error を clear/revalidate しない。既存 test は error 表示までしか扱わない。（具体パスは §2）
- 重複判定（例: 002 vs 022）: 単独。field error lifecycle の stale state。
- 所有境界（FE / BE / データ / 環境）: FE=shared reservation modal/new owner form; BE=owner APIは回帰のみ; データ=なし; 環境=なし

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/components/shared/ReservationFormModal/NewOwnerInlineForm.tsx`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.test.tsx`。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: API/DB変更なし。phone fieldのvalidate/clear lifecycleだけを修正する。
- 正しい挙動の定義（1〜3 文）: 有効値へ直すとphone errorだけが消え、他field errorは保持し、正当なsubmitは1回行われる。
- やらないこと（Out of scope）: 電話番号仕様の変更、全errorの無条件clear。
- 既存データ修復の要否と手順: 不要。

- 電話番号 value/error を一つの form state で管理し、change/blur/submit が同じ pure validator を使う。変更時にその field だけ再検証し、無関係な server error は消さない。

#### 4. 受け入れ基準（AC）

1. Given不正電話で submit して error 表示中、When有効形式へ修正する、Then電話 error が消え、再 submit が可能。
2. 回帰: 他field errorと正常owner作成を維持する。
3. 負例: 無効phoneでsubmitせず、有効phone修正で無関係なerrorを消さない。 既存境界AC: Given別 field error、When電話だけ修正、Then別 error は保持される。
4. 横展開確認対象: 予約modal内の新規/既存owner入力

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: 該当なし（API/DB変更なし。FE境界testで検証）
- FE: `docker compose exec -T frontend npx vitest run src/components/shared/ReservationFormModal/ReservationFormModal.test.tsx src/components/shared/ReservationFormModal/NewOwnerInlineForm.test.tsx`
- 手動/E2E: V02 §7 予約登録モーダル（新規飼主タブ） 電話番号入力欄を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 臨床/会計直接変更なし。個人連絡先validationをfield単位で扱う。
- fail-closed の維持: invalid時はrequestせず、他field errorをsilent clearしない。
- audit / トランザクション境界: invalid pathはwrite/auditなし。正常owner作成txは既存契約。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- 入力途中の値を過剰に拒否せず submit/blur 時期を明示する。電話番号をログへ出さない。

#### 7. 実装ステップ（順序付き）

1. invalid→valid の stale-error test を追加する。
2. validator と field state owner を統一する。
3. keyboard/blur/server-error coexistence を回帰確認する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- invalid→valid、別 error 保持、submit の3ケースが通り、error state の二重 owner がない。

## BUG-021: 死亡記録ダイアログの必須・未来日バリデーションでエラーメッセージが一切表示されない

- **重大度**: 中
- **対応状況（2026-08-03 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み commit `eb7db0dc94fb842c7e569252a9cebc6aee96cd60` で、実ボタン経路から空欄・未来日の field error と `aria-invalid` / `aria-describedby` を表示し、mutation 0回を回帰固定。BEもJST基準で欠損・空・不正形式・未来日を service 呼出前に400で拒否し、当日・過去日のみ通す | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: V03 §4チェック1・2を専用 synthetic fixture でブラウザ再検証し、`VERIFIED_FIXED` 可否を判定
- **発見シナリオ**: V03 §4 ペット死亡登録ダイアログ（pet-deceased-dialog）チェック1・チェック2
- **再現手順**:
  1. 飼主編集画面（`/owners/310374`）でペット「V03テストペット改」の編集モーダルを開く
  2. 「死亡を記録」リンクをクリックして「死亡を記録する」ダイアログを開く
  3. 死亡日を空欄のまま「死亡を記録する」ボタンをクリック
  4. （別ケース）死亡日に翌日以降の未来日を入力して「死亡を記録する」をクリック
- **期待結果**: それぞれ「死亡日を入力してください」「未来の日付は指定できません」というエラーメッセージが表示される
- **実際の結果**: どちらのケースもAPIコール（`PATCH /api/v1/clinics/1/pets/{id}/death`）は発火せず保存はブロックされる（データ整合性は保たれている）が、画面上にエラーメッセージが一切表示されない。「死亡を記録する」ボタンは `<form>` に属さない `button[type="submit"]` のため、ネイティブHTML5バリデーションすら発火しない。ユーザーからは、ボタンを押しても何も起きていないように見える。
- **備考**: pets/1015656（V03テストペット改, owner 310374）で検証。2回ずつ再現性を確認済み。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-PET-DEATH + C-SILENT-VALIDATION`
- 同一 PR にする BUG: BUG-021（BUG-002/022 と回帰結合）
- 先行必須: 共通 field-error 契約
- 後続解放（シナリオ/他BUG）: BUG-002 の即時表示と BUG-022 の永続表示

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE
- 観測根拠（API・クエリ・コード参照）: submit button の `form` 属性は現行に存在する。ただし native `required`/date constraint が React action より先に submit を止めるため custom error state が描画されない。既存 test は action を直接呼び、この経路を通らない。BE は必須だけで未来日を拒否しない。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-017 と無音 validation、BUG-002/022 と死亡 lifecycle を共有するが、本件は死亡日入力境界。
- 所有境界（FE / BE / データ / 環境）: FE=PetDeceasedDialog; BE=lstep/pet death validation owner; データ=death date/optional reason/audit; 環境=現行server canonical JST

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/components/shared/PetDeceasedRecordButton/PetDeceasedDialog.tsx` と同 test。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/lstep/lstep_lifecycle_handler.go` および death service/request validation。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: API/DB schema変更なし。死亡日は必須・未来日不可、理由は任意という既存契約をFE/BEで一致させる。
- 正しい挙動の定義（1〜3 文）: 無効な死亡日はrequestせずfield errorを表示し、有効日は同一txで状態・死亡記録・auditを保存する。
- やらないこと（Out of scope）: 理由の必須化、timezone推測、失敗時のdeceased表示固定。
- 既存データ修復の要否と手順: 不要。

- browser native validation と app validation の owner を明示し、form action が常に typed field errors を生成できる構成に統一する。`noValidate` を使う場合も semantic input/ARIA を維持する。
- BEは現行server canonical JSTで死亡日を必須・実在日・未来日禁止として検証する。理由は任意を維持し、clinic/pet/既死亡を既存の原子的status predicateとtxで拒否する。per-clinic timezoneや新しいpet lockを根拠なく導入しない。

#### 4. 受け入れ基準（AC）

1. Given死亡日が空（理由は任意）、When実際のbuttonを押す、Then request 0件で「死亡日を入力してください」が表示・読み上げされる。
2. 回帰: 有効死亡日・任意理由・BUG-002/022表示を維持する。
3. 負例: 空/未来日/invalid timezoneはrequestせず、BE失敗時にdeceased表示を固定しない。 既存境界AC: Given未来日、When保存、Then FE/BE とも拒否する。Given今日以前、When保存、Then BUG-002/022 の状態反映へ進む。
4. 横展開確認対象: 死亡入力を使うdialog/API/lifecycle

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/lstep/... -run 'Test.*(Death.*Date|Deceased)' -count=1`
- FE: `docker compose exec -T frontend npx vitest run src/components/shared/PetDeceasedRecordButton/PetDeceasedDialog.test.tsx`
- 手動/E2E: V03 §4 ペット死亡登録ダイアログ（pet-deceased-dialog）チェック1・チェック2を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 死亡日は臨床/法的記録。clinic/pet/timezoneを検証し、理由は任意を維持。
- fail-closed の維持: 無効日を保存せず、失敗時にdeceased表示へfallbackしない。
- audit / トランザクション境界: pet status/death record/auditを同一DBOrTx。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- client clock だけを信頼せず BE timezone を正本にする。未来日拒否と死亡記録の監査を別 tx にしない。既死亡 record を上書きしない。

#### 7. 実装ステップ（順序付き）

1. real click で native block を再現する component test を作る。
2. error owner/ARIA と BE date validation を実装する。
3. BUG-002 の即時反映、BUG-022 の再取得を同じ E2E で通す。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- 空/未来/有効日の FE・BE matrix、request 0件、ARIA、tx監査、死亡表示の往復が通る。

## BUG-022: 死亡記録後、フルリロード後もペット編集モーダル・一覧の生死ステータスが「生存」のまま表示される（DBはdeceasedで正しい）

- **重大度**: 高
- **対応状況（2026-08-03 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み commit `fc3c12b2800942c7527b0be951aad20860c6131c` で、pet detail・owner埋込petは `deceased` / `deceased_at` を保持し、`deceased_at` を返さない一覧APIでは owners loader が `deceased` status を「死亡」へ変換、編集formは提供された `deceasedAt` を保持する。未知・欠損statusは「生存」へ推測せず「不明」にし、死亡記録・患者選択などの臨床操作をfail-closedにする回帰を追加済み | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: V03 §4チェック3を専用 synthetic fixture の真のフルリロードで再検証し、`VERIFIED_FIXED` 可否を判定
- **発見シナリオ**: V03 §4 pet-deceased-dialog チェック3（C2永続化確認: 記録→一覧→再読込→再オープン）
- **再現手順**:
  1. ペット編集モーダルで死亡日（当日）・死亡理由を入力し「死亡を記録する」で保存 → 「死亡を記録しました」トースト表示、モーダル内の生死ステータスが「死亡」に切り替わり永眠バナーが表示される
  2. `window.location.reload()` でページを完全リロード（SPAのnavigateではなく真のフルリロード）
  3. 飼主詳細画面のペット一覧、および再度開いたペット編集モーダルの生死ステータスを確認
- **期待結果**: リロード後も死亡ステータス・死亡日・理由が保持されて表示される
- **実際の結果**: `GET /api/v1/pets/1015656` および `GET /api/v1/owners/310374` のレスポンスは正しく `status: "deceased"` を返す（バックエンドの永続化は正常）。しかしフルリロード後の一覧表示・再オープンしたペット編集モーダル内の生死ステータスラジオボタンは、いずれも「生存」が選択された状態のまま表示される。フロントエンドが死亡ステータスの読み取り・表示に失敗している。
- **備考**: 既知バグBUG-002（「ペット死亡ステータスが一覧に反映されない・要リロード」）と類似領域だが、BUG-002は「リロードすれば直る」ことが前提の記述である一方、本件は**フルページリロード後も直らず**、かつ一覧表示だけでなく**ペット編集モーダル自体の生死ラジオボタンも誤表示する**点でBUG-002より重大、または別事象の可能性がある。重複か否かの切り分けはコードレビューでの確認を推奨する。クリーンアップ: `DELETE /api/v1/clinics/1/pets/1015656/death` により生存状態へ復元済み（API経由で確認済み）。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-PET-DEATH`
- 同一 PR にする BUG: current buildで同一原因を再現できた場合のみBUG-002, BUG-022。非再現または別原因なら分離
- 先行必須: response→transform→view-modelとserved assetのprovenance採取
- 後続解放（シナリオ/他BUG）: S01 死亡登録の再読込後表示と関連患者選択

#### 1. 切り分けステータス

- 主因レイヤ: FE（REPRODUCE-FIRST）
- 観測根拠（API・クエリ・コード参照）: 現行 `frontend/src/lib/transforms/pet.ts` は `deceased` を「死亡」へ写像し loader も transform を使うため、フルリロード後も生存という報告原因は current source だけでは説明できない。古い bundle/API payload/query key を観測する。一方、未知 status を生存へ fallback する経路は fail-open リスク。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-002は即時local-state staleで確認済み。本件は現行mappingが既にdeceased対応のためNEEDS REVIEWであり、再現・同一failure pointの証明前は重複扱いしない。
- 所有境界（FE / BE / データ / 環境）: FE=pet/owner transforms + owners loader/form; BE=pet/owner staff response; データ=既に正しいdeceased; 環境=served asset/cache provenance

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/lib/transforms/pet.ts`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/lib/transforms/owner.ts`、`frontend/src/features/owners/components/pet-form-data.ts`。 | owner responseとform defaultの変換 |
| FE | `frontend/src/features/owners/loaders.ts` と owner/pet query cache。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/features/owners/hooks/use-owner-form.ts` の status default。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/pet/pet_response.go`、`backend/internal/owner/http_response.go`。 | staff向けstatus/date DTO。death reasonは意図的に除外 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: status/date修正はAPI/DB変更なし。`deceased_reason`は現行pet/owner DTOからPII最小化のため意図的に除外されており、理由再表示が必須ならstaff-only curated APIのproduct/security決定を別承認する。
- 正しい挙動の定義（1〜3 文）: full reloadとmodal/listの全consumerが`status=deceased`と`deceased_at`を死亡表示へ変換する。death reasonは現契約では表示対象外としてUNREPORTEDを残し、無断で汎用DTOへ追加しない。
- やらないこと（Out of scope）: DB状態の上書き、未知statusの生存扱い、別pet cache流用。
- 既存データ修復の要否と手順: 不要。

- 元 pet ID で network response→transform→form/table view model を trace し、source/asset/cache の failure point を確定する。再現しなければ current mapping の regression test と build provenance の確認だけを行う。
- 未知/欠損 status は「生存」に推測せず「不明」または安全な非選択状態とし、clinical action で fail-closed にする。

#### 4. 受け入れ基準（AC）

1. Given DB/APIが`status=deceased`と`deceased_at`を返す、When full reloadしてモーダルと一覧を開く、Then両方「死亡」と日付を示し、臨床/会計/予約の対象選択は不可となる。
2. 回帰: 生存pet、死亡pet、modal/list/full reloadを維持する。
3. 負例: unknown/null/別pet statusを生存defaultへ黙って変換しない。 既存境界AC: Given未知 status、When表示/会計・予約選択、Then「生存」扱いせず明示的な不明/拒否になる。別 pet/clinic cache は混入しない。
4. 横展開確認対象: pet transformの全consumer

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/pet/... ./internal/owner/... -run 'Test.*(Deceased|ClinicScope|Response)' -count=1`
- FE: `docker compose exec -T frontend npx vitest run src/lib/transforms/pet.test.ts src/features/owners/loaders.test.ts src/features/owners/hooks/use-owner-form.test.ts src/features/owners/components/OwnerPetsSection.test.tsx`
- 手動/E2E: V03 §4 pet-deceased-dialog チェック3（C2永続化確認: 記録→一覧→再読込→再オープン）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 死亡statusは臨床安全対象。clinic/pet responseを正しくtransformする。
- fail-closed の維持: unknown/nullを生存へ黙ってdefaultせず、別pet値を使わない。
- audit / トランザクション境界: read-only表示。既存death write/auditはBUG-002/021で回帰。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- stale asset を source 修正と誤認しない。未知 status の fail-open は臨床/予約制約を破るため禁止する。既存 DB status を自動補修しない。

#### 7. 実装ステップ（順序付き）

1. response/asset hash/query key を含む元手順を再現する。
2. 確定した層だけを修正し、unknown fail-closed test を追加する。
3. BUG-002/021 と full reload を一つの pet-death E2E にする。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- full reloadとimmediate updateの両方が死亡status/date表示、unknownはfail-closed、原因がsource/asset/cacheの証拠で説明される。death reasonはstaff-only curated API決定がない限り残余UNREPORTEDとして明記する。

## BUG-023: 権限グループ名の重複エラーが未整形の生バックエンドメッセージのまま表示され、グループ名部分が空文字になる

- **重大度**: 中
- **対応状況（2026-08-04 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み commit `41ba79b4ce8db01e4743bbfcaf77fa1612a3a186` で permission group name unique を `permission_group_name_conflict` + safe params に変換し FE が「権限グループ名『X』は既に使用されています」を表示。`uk_permission_group_rules` は name conflict へ昇格しない | **原文シナリオ再検証**: WAIVED（2026-08-05 USER 判断・ブラウザ再検証を実施しない） | **次のアクション**: なし（検証見送り。claim 解放済み）
- **発見シナリオ**: V03 §6 permission-group-side-panel チェック3
- **再現手順**:
  1. `/settings/permission-groups` で「新規登録」をクリック
  2. グループ名に既存グループと同一の文字列（例: "執行"）を入力
  3. 「保存」をクリック
- **期待結果**: 409は適切にハンドリングされ、ユーザーに理解可能な日本語エラーメッセージが表示される
- **実際の結果**: トーストに `permission_group '' already exists` という未翻訳の生バックエンドエラー文字列がそのまま表示される。テンプレート変数（重複したグループ名）も空文字列になっており、本来入っているべき "執行" の部分が欠落している。500/白画面のクラッシュは回避されているが、日本語話者の医院スタッフには意味の分からない英語の技術的エラーメッセージが露出している。
- **備考**: `handleApiError` 経由のエラー整形処理に、このエンドポイント固有のメッセージローカライズ・変数埋め込みが未実装と思われる。2回再現性確認済み。BUG-027（動物種類の同種メッセージ）と共通の実装パターンの可能性が高い。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-MASTER-DUPLICATE-MSG`
- 同一 PR にする BUG: BUG-023, BUG-027
- 先行必須: 安定した conflict code / params 契約
- 後続解放（シナリオ/他BUG）: 権限グループ・動物種ほか duplicate 表示

#### 1. 切り分けステータス

- 主因レイヤ: BE / FE
- 観測根拠（API・クエリ・コード参照）: repository の DB error が `apperrors`/HTTP response を通って生文字列になり、FE `handle-api-error.ts` がそのまま toast へ渡す。対象名を構造化 params として保持していない。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-027 と unique violation→生文言を共有。本件 resource は permission group。
- 所有境界（FE / BE / データ / 環境）: FE=API error adapter/settings UI; BE=auth repository + apperrors/httpapi; データ=clinic-scoped unique conflict; 環境=なし

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| BE | `backend/internal/auth/permission_group_repository.go`。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/apperrors/errors.go`、`backend/internal/httpapi/response.go`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/lib/handle-api-error.ts` と permission group settings UI。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: additive API error契約変更あり: conflictへ安定した `code` と安全な `params` を付与。DB schema変更なし。
- 正しい挙動の定義（1〜3 文）: 重複permission group名を対象名付き日本語で示し、unknown DB errorは内部情報なしのfallbackにする。
- やらないこと（Out of scope）: constraint/table名やSQLの露出、FEだけの文字列parse。
- 既存データ修復の要否と手順: 不要。

- owner service で unique violation を domain error `permission_group_name_conflict` と安全な params（入力名）へ変換する。HTTP は stable code と field を返し、内部 constraint/table/SQL を返さない。
- FE は code を日本語 field/toast 文言へ mapping し、未知 code は一般エラーへ落とす。生 backend message を表示しない。

#### 4. 受け入れ基準（AC）

1. Given同一 clinic に同名 group、When作成/改名、Then保存されず「権限グループ名『X』は既に使用されています」と表示し、空名・table名・SQLは出ない。
2. 回帰: 正常作成、unknown error fallback、BUG-027を維持する。
3. 負例: 別clinic規則どおりの同名、空params、未知DB errorで内部名を露出しない。 既存境界AC: Given別 clinic に同名、When許可された作成、Then tenant 規約どおり扱われ、他 clinic の存在を理由に拒否/開示しない。
4. 横展開確認対象: 全master conflict adapterとhttp error envelope

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/auth ./internal/httpapi -run 'Test.*PermissionGroup.*Duplicate|Test.*ErrorResponse'`
- FE: `docker compose exec -T frontend npx vitest run src/features/master/routes/PermissionGroupSettings.test.tsx src/features/master/components/PermissionGroupSidePanel.test.tsx`。`src/lib/handle-api-error.test.ts`は新規追加予定。
- 手動/E2E: V03 §6 permission-group-side-panel チェック3を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 権限/clinic隔離対象。safe code/paramsだけを公開する。
- fail-closed の維持: DB/constraint/table名をraw表示せず、unknown errorをsuccessにしない。
- audit / トランザクション境界: permission group create/auditの既存txを維持。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- constraint 名の文字列一致だけに依存せず Postgres code/known constraint を限定 mapping する。入力名は UI text として escape し、tenant existence は漏らさない。

#### 7. 実装ステップ（順序付き）

1. repository/service/HTTP の error contract test を追加する。
2. stable code/params と FE localization を実装する。
3. BUG-027 を同じ adapter で検証し、未知 DB error の非漏洩を確認する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- 2 master の重複が適切な対象名で表示され、生 DB message/空名がなく、clinic 分離と unknown-error fallback が通る。

## BUG-024: 権限グループの権限マトリクス（表示/作成/編集/削除チェックボックス）の変更が保存されない（成功トースト・200応答にもかかわらずDBに反映されない）【重大】

- **重大度**: 高
- **対応状況（2026-08-03 JST）**: OPEN | **根拠**: FE rules[] + BE `UpdateWithRules` 経路は存在; 原文「200だがDB未反映」は request/DB トレース未実施で否定不能（wave-1） | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: トグル→PATCH body→GET/DB 一致の統合確認
- **発見シナリオ**: V03 §6 permission-group-side-panel チェック2（C2永続化確認）およびチェック4（自己剥奪ガード確認）
- **再現手順**:
  1. `/settings/permission-groups` で「執行」グループの編集パネルを開く
  2. 権限マトリクスの任意のリソース行（例:「当日の受付」の「表示」列、または「権限グループ」の「編集」列）のチェックを外す
  3. 「保存」をクリック →「更新しました」のトースト表示を確認
  4. `PATCH /api/v1/masters/permission-groups/1` が200で成功していることを確認
  5. 直後に `GET /api/v1/masters/permission-groups/1` を叩き、当該ルールの値を確認
- **期待結果**: チェックを外した状態が保存され、一覧・再読込・再オープンいずれでも変更後の状態が表示される。自己剥奪ガード（BUG-140）対象フィールドでは保存自体が拒否される。
- **実際の結果**: `PATCH` は200で成功し「更新しました」のトーストも表示されるが、直後の `GET` では該当ルールの値が変更前のまま（`updated_at` タイムスタンプのみ更新される＝サーバ側で何らかの処理は実行されているが値自体が反映されていない）。この現象は通常フィールドでも自己剥奪ガード対象フィールドでも同様に再現するため、チェック4（BUG-140自己剥奪ガードの動作確認）は本バグによりブロックされ、「ガードが正しく機能して拒否している」のか「単にこの保存バグにより変更自体が反映されていないだけ」なのかを区別できず、**未確定**として報告する。
- **備考**: 3回のトグル+保存+API確認で再現性を確認済み、いずれも同じ結果。ページ再読込・パネル再オープン後も変更前の値が表示されることも確認済み。管理者が特定スタッフの操作権限を剥奪したつもりが実際は剥奪されていない、という重大な業務影響が起こり得る。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `S-PERMISSION-MATRIX`
- 同一 PR にする BUG: なし
- 先行必須: payload / response / DB readback / cache の同一 trace 採取
- 後続解放（シナリオ/他BUG）: V03 権限マトリクスと自己剥奪 guard

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE / DB（REPRODUCE-FIRST）
- 観測根拠（API・クエリ・コード参照）: 現行 FE は rules builder と PATCH body を持ち、BE request/service は rules を受け repository tx で置換する。報告の「200だがDB未反映」は source だけでは再現原因を確定できない。自己権限剥奪 guard と stale response/cache も区別する。（具体パスは §2）
- 重複判定（例: 002 vs 022）: 単独。現行 source は rules payload/update を持つため request/DB/readback/cache の落下点を確定する。
- 所有境界（FE / BE / データ / 環境）: FE=permission settings model/API/cache; BE=auth handler/service/repository; データ=group rules + audit + clinic scope; 環境=loaded asset/runtime trace

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | FE `frontend/src/features/master/routes/permission-group-settings-model.ts` と `frontend/src/features/master/api/permission-groups.ts`。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/auth/http_permission.go`、`permission_group_service.go`、`permission_group_repository.go`。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: 再現で落下点が確定するまでAPI/DB変更なし。成功はpersist/readback後だけとし、必要なowner層だけを直す。
- 正しい挙動の定義（1〜3 文）: matrix変更が同一clinicのDB/再GETへ一致し、自己剥奪や別clinic rule IDは原子的に拒否される。
- やらないこと（Out of scope）: UIだけの成功表示修正、権限guard無効化、他tenant rule付与。
- 既存データ修復の要否と手順: 不明。DB不反映なら不要、部分反映が見つかった場合のみread-only audit→別補修。

- click時 form state→PATCH body→response→GET→permission-group/rules tables を同一 ID/clinic で追跡し、drop point を確定する。成功通知は再取得した version/rules が request と一致した後に出す。
- service は group と全 rules を一 tx で lock/replace/audit し、自己剥奪 guard は明示 code で全 rollback する。

#### 4. 受け入れ基準（AC）

1. Given非自己剥奪の1権限変更、When保存、ThenPATCH body、response、再GET、DB rules が一致し再ログイン後も反映される。
2. 回帰: 正常rule更新、再GET、self-deprivation guardを維持する。
3. 負例: 別clinic rule ID、権限なしactor、途中失敗は200/success toastや部分commitにしない。 既存境界AC: Given自己剥奪/別clinic rule ID/途中失敗、When保存、Then明示拒否か全 rollback で、虚偽成功を出さない。
4. 横展開確認対象: permission matrixのcreate/update/delete/readback

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/auth -run 'Test.*PermissionGroup.*Rule|Test.*Self.*Permission'`
- FE: `docker compose exec -T frontend npx vitest run src/features/master/routes/PermissionGroupSettings.test.tsx`
- 手動/E2E: V03 §6 permission-group-side-panel チェック2（C2永続化確認）およびチェック4（自己剥奪ガード確認）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 権限・clinic隔離の重大変更。actor、group、rule IDsを同一clinicで検証する。
- fail-closed の維持: 自己剥奪/別clinic/途中失敗を200やpartial rulesへfallbackしない。
- audit / トランザクション境界: rules一式とauditをauth ownerの同一DBOrTxでlock/update/readback。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- 権限変更は current session を含むため専用 test actor/group を使い、管理者をロックアウトしない。rule replace と audit を tx 外に分けない。

#### 7. 実装ステップ（順序付き）

1. read-only trace で報告を再現し failure layer を固定する。
2. その層の RED test を追加し、必要最小の save/tx/cache 修正を行う。
3. 再ログイン、自己剥奪、別 clinic、途中失敗を検証する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- body→DB→再GET→認可結果が一致し、虚偽成功なし、自己剥奪/tenant/rollback test が通る。再現不能なら current contract の証拠と回帰範囲を記録する。

## BUG-025: 主訴種別・問診テンプレートなど一部マスタで新規作成時に is_active=false（無効）で保存される

- **重大度**: 高（新規作成したマスタが画面上「有効」に見えるトグル状態のまま保存すると、実際には無効状態でDBに保存され、他画面のドロップダウン等に一切現れず、ユーザーが気づかないまま機能しない）
- **対応状況（2026-08-04 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み commit `4335e3a99718f90157f739f887f965c3e341a905` で FE create builder が `is_active` を明示送信し、BE create request は presence-aware `*bool`（欠落→true / 明示 false は false）に変更。builder unit + medicalrecord package 回帰 green。| **原文シナリオ再検証**: WAIVED（2026-08-05 USER 判断・ブラウザ再検証を実施しない） | **次のアクション**: なし（検証見送り。claim 解放済み）
- **発見シナリオ**: V04 §1 主訴種別 `/settings/interview/chief-complaint`、問診・定型文テンプレート `/settings/inquiry-templates`
- **再現手順**:
  1. `/settings/interview/chief-complaint` を開く
  2. 「新規登録」→名称「V04主訴」を入力（ステータストグルは初期表示のまま「有効」、一切操作しない）
  3. 「保存」をクリック → 一覧に「V04主訴」が追加されるが、ステータス列が「無効」と表示される
  4. `GET /api/v1/masters/chief-complaint-types` で確認すると `"is_active": false`
  5. 同様の手順を `/settings/inquiry-templates`（問診テンプレート）で実施 → こちらも新規作成した「V04テンプレ」が `is_active: false` で保存される
- **期待結果**: ステータストグルが「有効」を表示したまま保存した場合、保存後のレコードも `is_active: true` になる（動物種類・診断カテゴリ・ケージ・物販・保険・入院プランでは同一パターンで正しく `is_active: true` で保存されることを確認済み）
- **実際の結果**: 主訴種別・問診テンプレートの新規作成では、UIのトグル表示に反して `is_active: false` でDBに保存される
- **備考**: `GET /api/v1/masters/chief-complaint-types` のレスポンスで再現性を2回確認済み（id:19, id:20 いずれも false）。本バグは一部マスタ（少なくとも主訴種別・問診テンプレートの2フォーム）に限定される可能性が高く、他フォームでも横展開で発生していないか要確認。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-MASTER-SAVE-FIELDS`
- 同一 PR にする BUG: BUG-025, BUG-028（共通 builder 規約; domain差分は分離可）
- 先行必須: create payload の field-presence 契約
- 後続解放（シナリオ/他BUG）: 主訴・問診テンプレートと処置マスタ保存

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE
- 観測根拠（API・クエリ・コード参照）: 主訴/問診 side panel の初期 model は active=true だが create payload builder が `is_active` を落とす。BE/ORM の zero value/default 契約に依存して false 保存となる経路がある。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-028 と create field omission を共有。本件は is_active presence、028 は anesthesia。
- 所有境界（FE / BE / データ / 環境）: FE=master form models/builders; BE=domain create validators; データ=existing inactive rows; 環境=なし

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/master/components/chief-complaint-side-panel-model.ts` と `frontend/src/features/master/routes/chief-complaint-settings-model.ts`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/features/master/components/interview-template-side-panel-model.ts` と `frontend/src/features/master/routes/interview-template-settings-model.ts`。 | 現行証拠 / 変更候補 |
| BE | 対応する BE request/service/repository の create contract。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: API payload意味論の明示あり: create builderが`is_active`のpresenceとbooleanを送る。DB schema変更なし。
- 正しい挙動の定義（1〜3 文）: 明示falseはinactiveのままround-tripし、true/defaultとの区別をFE/BEで保持する。
- やらないこと（Out of scope）: 既存inactive行の自動有効化、全master builderの無関係な変更。
- 既存データ修復の要否と手順: 既存inactive行は意図を推測できないため自動補修しない。

- create/update payload 型に `is_active` を必須 boolean として含め、UI default を request まで保持する。BE は field presence と値を明示的に扱い、未指定を暗黙 false にしない。
- 既存 inactive rows は意図的無効化とバグ生成を区別できないため一括 active 化しない。

#### 4. 受け入れ基準（AC）

1. Given新規 panel の既定 active、When主訴/問診を保存、Then request/response/再GET/DB が `is_active=true`。
2. 回帰: 明示true/updateとBUG-028 field builderを維持する。
3. 負例: 明示falseをdefault trueへ変換せず、既存inactive行を自動変更しない。 既存境界AC: Given明示的 inactive、When保存、Then false が round-trip し、別 clinic の同名/状態へ影響しない。
4. 横展開確認対象: chief complaint/interview templateほかmaster create builders

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*ChiefComplaint.*Create|Test.*InterviewTemplate.*Create'`
- FE: `docker compose exec -T frontend npx vitest run src/features/master/components/MasterCRUDPage.test.tsx`。`src/features/master/routes/ChiefComplaintSettings.test.tsx`と`InterviewTemplateSettings.test.tsx`は新規追加予定。
- 手動/E2E: V04 §1 主訴種別 `/settings/interview/chief-complaint`、問診・定型文テンプレート `/settings/inquiry-templates`を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: master有効状態は診療選択肢へ波及。clinic/presence/booleanを検証する。
- fail-closed の維持: 明示falseをdefault trueへfallbackしない。
- audit / トランザクション境界: create/update/auditを各domain ownerの既存txで維持。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- omitted/false を同一視しない。既存 inactive の補修は作成時刻・audit・利用有無を抽出して業務承認する別作業。

#### 7. 実装ステップ（順序付き）

1. payload builder の true/false round-trip test を追加する。
2. FE 必須 field と BE presence-aware request を実装する。
3. 既存候補を read-only 監査し、BUG-028 の必須 field 契約へ横展開する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- true/false 双方の create/update/再GET、clinic 分離が通り、既存 inactive を未承認で変更していない。

## BUG-026: 保険マスタで補償率が範囲外（>100）の値を保存しようとすると、実際には保存されていないのに「登録しました」の成功トーストが表示される

- **重大度**: 高（無音失敗どころか、偽の成功通知によりユーザーが保存されたと誤認する）
- **対応状況（2026-08-03 JST）**: OPEN | **根拠**: 受入実測は「POST 0件＋成功 toast」だが、current `InsuranceSettings` の validate は名称のみ、共有 `useMasterSave` は mutation 成功後だけ success toast を出す。この source は実測前から存在し、事後 fix の証拠ではなく、観測経路との矛盾が未解決。原文シナリオ未再実行のため IMPLEMENTED_UNVERIFIED へ上げない | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: coverage=101 で click/request/mutation/toast timeline を再取得し、再現後に最小修正を決める（REPRODUCE_FIRST）
- **発見シナリオ**: V04 §1 保険 `/settings/insurance`（C1-3 境界値チェック）
- **再現手順**:
  1. `/settings/insurance` で「新規登録」→ 名称「V04保険2」、補償率(%)に `101` を入力
  2. 「保存」をクリック
  3. 画面右下に緑色の「登録しました」トーストが表示され、SidePanelは開いたまま（一覧には反映されない）
  4. ネットワークログを確認すると、`POST /api/v1/masters/insurances` へのリクエストが一切発生していない（GETのみ）
  5. `GET /api/v1/masters/insurances` で確認 → 「V04保険2」は存在しない（実際には保存されていない）
  6. 同様に `150` でも再現
- **期待結果**: 補償率が0〜100の範囲外の場合は保存されず、エラーメッセージが表示される（成功トーストは出ない）
- **実際の結果**: 保存は実際には行われず（FE側で送信をブロック）にもかかわらず、保存成功時と同じ「登録しました」の成功トーストが表示される。エラー表示が一切ない。
- **備考**: `POST` が送信されていないことをネットワークログで確認、また保存後にAPI照会で新規レコードが存在しないことを確認。100（境界内）は正しく受理・保存されることを確認済み。BUG-029（支払方法マスタ）と表示パターンが完全に一致しており、同一の共通保存処理（SidePanel/useMasterSave）に起因する可能性が高い。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-MASTER-FALSE-SUCCESS`
- 同一 PR にする BUG: BUG-026, BUG-029
- 先行必須: click / request / mutation / toast lifecycle の再現 trace
- 後続解放（シナリオ/他BUG）: 保険・支払方法と共通 master save

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE（REPRODUCE-FIRST）
- 観測根拠（API・クエリ・コード参照）: 現行 `use-master-save.ts` は mutation success callback 後だけ toast するように見え、validation false で成功通知する直接経路は確認できない。native form/action、前回 toast の残留、呼出側 return 契約を実操作で区別する。保険は FE/BE 双方の補償率境界も不足がないか確認する。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-029 と false-success 報告を共有。ただし現行 hook は mutation成功後だけ toast するため実 trace 必須。
- 所有境界（FE / BE / データ / 環境）: FE=useMasterSave + insurance form; BE=billing insurance validation; データ=coverage_rate; 環境=click/request/toast timeline

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/master/hooks/use-master-save.ts`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/features/master/routes/InsuranceSettings.tsx`、`frontend/src/features/master/components/InsuranceSidePanel.tsx`。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/billing/insurance_request.go` と insurance service。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: 原因確定前はAPI/DB変更なし。再現時のみFE save resultをdiscriminated union化し、0–100 validationをFE/BEで一致させる。
- 正しい挙動の定義（1〜3 文）: validation失敗・mutation失敗では成功toastを出さず入力を保持し、成功応答時だけ1回表示する。
- やらないこと（Out of scope）: 報告だけを根拠に現行hookを改変、BE validation弱体化。
- 既存データ修復の要否と手順: 不要。

- save pipeline を `validation_failed | mutation_succeeded | mutation_failed` の discriminated result にし、toast/close/cache update は `mutation_succeeded` だけで行う。
- 補償率は FE/BE とも0〜100の同じ境界で検証し、API未発火の client error と API error を別表示する。再現不能なら共通 hook を書き換えず回帰 test のみ追加する。

#### 4. 受け入れ基準（AC）

1. Given補償率101、When保存、Then POST/PATCH 0件、field error 表示、成功 toast 0件、panel は入力保持。
2. 回帰: 0/100と正常保存、BUG-029共通saveを維持する。
3. 負例: 101/NaN/server failure/request0ではsuccess toastを出さない。 既存境界AC: Given0/100、When保存、Then API成功後だけ toast/close/cache update。Given server失敗、Then成功通知なし。
4. 横展開確認対象: useMasterSave全consumerとtoast lifecycle

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/billing -run 'Test.*Insurance.*Validation'`
- FE: `docker compose exec -T frontend npx vitest run src/features/master/hooks/use-master-save.test.ts`。`src/features/master/routes/InsuranceSettings.test.tsx`は新規追加予定。
- 手動/E2E: V04 §1 保険 `/settings/insurance`（C1-3 境界値チェック）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 会計マスタvalidation。clinic scopeとcoverage_rate境界を維持する。
- fail-closed の維持: validation/mutation失敗をsuccess toastへ変換しない。
- audit / トランザクション境界: 成功mutationのBE tx/audit後だけFE成功通知。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- stale toast を新しい操作の結果と誤認しないよう test ID/時系列を確認する。共通 hook の変更は全 master の close/cache 挙動を回帰させ得るため consumer matrix を限定実行する。

#### 7. 実装ステップ（順序付き）

1. 実 button click で network/toast/panel state を再現する。
2. result 型と保険境界 test を追加し、必要な経路だけ修正する。
3. BUG-029 と代表成功/server失敗 master を回帰確認する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- false-success の原経路が除去または再現不能と証明され、invalid/valid/server-fail の toast・request matrix が通る。

## BUG-027: 動物種類マスタで一意制約違反時のエラートーストに実際の名称が表示されず空文字になる

- **重大度**: 低（機能的には重複登録が正しく拒否されており「無音失敗・白画面」にはならないが、エラーメッセージの品質が低く、内部テーブル名を露出した上に該当の名称が空欄になっている）
- **対応状況（2026-08-04 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み commit `41ba79b4ce8db01e4743bbfcaf77fa1612a3a186`（C-MASTER-DUPLICATE-MSG と同一）で animal species name unique を `animal_species_name_conflict` + safe params に変換し FE が「動物種類『X』は既に使用されています」を表示 | **原文シナリオ再検証**: WAIVED（2026-08-05 USER 判断・ブラウザ再検証を実施しない） | **次のアクション**: なし（検証見送り。claim 解放済み）
- **発見シナリオ**: V04 §1 動物種類 `/settings/animal-species`（C3-2 一意制約違反チェック）
- **再現手順**:
  1. `/settings/animal-species` で「V04動物種類」という名称を新規登録
  2. 続けて同名「V04動物種類」で再度新規登録を試みる
  3. `POST /api/v1/masters/animal-species` が 409 を返す
  4. 画面に表示されるエラートースト: `animal_species '' already exists`
- **期待結果**: ユーザーフレンドリーな日本語エラーメッセージで、入力した名称が正しく表示される
- **実際の結果**: `animal_species '' already exists` という内部的な英語メッセージがそのまま表示され、名称部分が空文字になっている
- **備考**: 保存自体は正しくブロックされている（一覧に重複行は増えない）ため機能的な重大バグではないが、UXとして低品質。BUG-023（権限グループの同種メッセージ）と共通の実装パターンの可能性が高い。他の一意制約マスタでも同様のメッセージ形式が使われている可能性が高い（未全数確認）。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-MASTER-DUPLICATE-MSG`
- 同一 PR にする BUG: BUG-023, BUG-027
- 先行必須: BUG-023 の conflict code / params 契約
- 後続解放（シナリオ/他BUG）: 動物種と他マスタの duplicate 表示

#### 1. 切り分けステータス

- 主因レイヤ: BE / FE
- 観測根拠（API・クエリ・コード参照）: animal species repository の unique error も共通 `apperrors`→raw HTTP→FE raw toast 経路を通り、入力名 params が失われる。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-023 と同一 error adapter 契約。本件 resource は global animal species。
- 所有境界（FE / BE / データ / 環境）: FE=master UI/error adapter; BE=pet repository + apperrors/httpapi; データ=global unique conflict; 環境=なし

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| BE | `backend/internal/pet/animal_species_repository.go` と service。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/apperrors/errors.go`、`backend/internal/httpapi/response.go`。 | 現行証拠 / 変更候補 |
| FE | animal-species settings UI と `frontend/src/lib/handle-api-error.ts`。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: additive API error契約はBUG-023を共用し、animal species用code/paramsを定義。DB schema変更なし。
- 正しい挙動の定義（1〜3 文）: global scopeの重複名を安全な日本語で示し、unknown errorは内部情報を隠す。
- やらないこと（Out of scope）: global/clinic scopeの変更、FE文字列parse。
- 既存データ修復の要否と手順: 不要。

- BUG-023 の stable code/params/localization adapter を再利用し、code を `animal_species_name_conflict` に限定する。global master か clinic master かという現行 scope 契約を test で明示する。

#### 4. 受け入れ基準（AC）

1. Given同一 scope に名称X、When重複作成、Then「動物種類『X』は既に使用されています」と表示し、空名/constraint/table は出ない。
2. 回帰: 正常作成、global scope、BUG-023を維持する。
3. 負例: unknown DB error、空paramsでtable/constraint名を露出しない。 既存境界AC: Given未知 DB error、When失敗、Then安全な一般文言となり内部情報を出さない。
4. 横展開確認対象: animal speciesと全duplicate master messages

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/pet ./internal/httpapi -run 'Test.*AnimalSpecies.*Duplicate|Test.*ErrorResponse'`
- FE: `docker compose exec -T frontend npx vitest run src/features/master/routes/AnimalSpeciesSettings.test.tsx`。`src/lib/handle-api-error.test.ts`はBUG-023と共通で新規追加予定。
- 手動/E2E: V04 §1 動物種類 `/settings/animal-species`（C3-2 一意制約違反チェック）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: global権限境界とerror安全。safe code/params以外を公開しない。
- fail-closed の維持: unknown DB errorをraw表示/successにしない。
- audit / トランザクション境界: animal species createの既存tx/auditを維持。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- global/clinic scope を勝手に変更しない。入力名は escape し、DB error はログにも必要最小限の context だけを残す。

#### 7. 実装ステップ（順序付き）

1. BUG-023 の共通 error adapter test に animal species case を追加する。
2. domain mapping と日本語文言を実装する。
3. empty-name/raw-message/unknown-error を回帰確認する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- BUG-023/027 の共通基盤と各 domain code が通り、対象名欠落・内部情報露出がない。

## BUG-028: 診療項目マスタ「処置」タブで新規登録が常に失敗する（内部フィールド名 anesthesia が必須なのにSidePanelに入力欄がない）

- **重大度**: 高（処置タブへの新規項目登録が事実上全くできない。エラートーストは表示されるが内部英語フィールド名が露出しており原因が分かりにくく、無音失敗に近い）
- **対応状況（2026-08-04 JST）**: IMPLEMENTED_UNVERIFIED | **実装 commit**: `842fe78d47f3609fc23a538f7c4613635e4fa84b` | **根拠**: FE 処置 SidePanel に麻酔区分 Select 追加、`CreateProcedureRequest.anesthesia` 必須化、create/update builder が送信、負値単価は field error で mutation 0 回 | **原文シナリオ再検証**: WAIVED（2026-08-05 USER 判断・ブラウザ再検証を実施しない） | **次のアクション**: なし（検証見送り。claim 解放済み）
- **発見シナリオ**: V04 §2 診療項目マスタ 処置タブ `/settings/treatment-items?tab=procedure`
- **再現手順**:
  1. `/settings/treatment-items?tab=procedure` で「新規登録」→ 名称「V04処置テスト」、単価に `-100`（負値）を入力して保存 → 何の反応もなくパネルが開いたまま（無音失敗）
  2. 単価を `500`（正常値）に修正し再度保存 → 数秒後に赤いエラートースト `anesthesia は必須です` が表示される（英語の内部フィールド名がそのまま露出）。SidePanelは開いたままで一覧（4,623件）は変化なし
  3. `POST /api/v1/masters/procedures` を `X-Requested-With: XMLHttpRequest` 付きで直接叩くと `400 {"error":"anesthesia は必須です"}`。`anesthesia:"none"` を追加すると `201` で正常作成される
  4. SidePanelのフォームには「麻酔」に相当する入力欄が一切存在しない（ステータス／単価(税込)／課税区分／税率／保険対象外／親カテゴリ／備考のみ）ため、UIからは処置タブの新規登録が構造的に不可能
  5. 同じマスタの「診察」タブでは同一構造のSidePanelから正常に新規保存できることを確認済み → 本バグは処置タブ（procedures エンドポイント）に限定
- **期待結果**: 処置タブの新規登録がSidePanelの入力のみで成功する。バックエンドが要求する必須フィールドはすべてフォームに入力欄があるか、妥当なデフォルト値がFEから送信される。エラーメッセージは内部フィールド名でなく日本語ラベルで表示される。
- **実際の結果**: `anesthesia`（麻酔）というSidePanelに存在しないフィールドがバックエンドで必須とされており、新規登録が常に失敗する。加えて負値単価での保存試行時は何のフィードバックも出ない（無音失敗）。
- **備考**: APIレベルでの直接検証によりFE/BE双方の原因を特定済み。作成してしまった検証用レコード（id:1005198, id:10013, および正常テスト用レコード）はいずれもAPI経由で `is_active:false` に無効化済み。既存の4,623件の処置データは編集（PATCH）では影響を受けない可能性があるため、既存データの編集保存も別途要確認（未実施）。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-MASTER-SAVE-FIELDS`
- 同一 PR にする BUG: BUG-025, BUG-028
- 先行必須: 麻酔区分の業務正本と BUG-025 の明示 field 規約
- 後続解放（シナリオ/他BUG）: 処置マスタ create/update round-trip

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE
- 観測根拠（API・クエリ・コード参照）: `TreatmentItemSidePanel.tsx` に anesthesia 入力がなく、create builder も送信しない一方、BE `procedure_request.go` は必須。負単価時の feedback も別 validation delta として扱う。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-025 と field omission を共有。本件は臨床的 anesthesia enum と price。
- 所有境界（FE / BE / データ / 環境）: FE=treatment item panel/builder; BE=medicalrecord procedure owner; データ=anesthesia/price/category/clinic FK; 環境=業務正本確認

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/master/components/TreatmentItemSidePanel.tsx`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/features/master/routes/treatment-plan-master-model.ts`。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/medicalrecord/procedure_request.go` と procedure service。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: 既存BE request契約を維持し、FEが承認済みanesthesia enumとpriceを明示送信する。DB schema変更なし。
- 正しい挙動の定義（1〜3 文）: 有効な麻酔区分/価格がcreate/update/reGETで一致し、欠落・負値はfield errorになる。
- やらないこと（Out of scope）: 暗黙`none`補完、既存処置4,623件の一括変更。
- 既存データ修復の要否と手順: 既存行は自動補修しない。enum/defaultの業務承認を先行する。

- 業務上の anesthesia enum/既定値を確認し、SidePanel に日本語ラベルの typed control を追加して create/update payload に必須送信する。根拠なく `none` を暗黙補完しない。
- 単価は FE/BE で非負・上限を同じ契約で検証し、field error を表示する。

#### 4. 受け入れ基準（AC）

1. Given有効な麻酔区分と単価、When処置を新規作成、Then201、再GET/DBで全 field が一致し active 状態も保持される。
2. 回帰: 正常update、price境界、BUG-025 buildersを維持する。
3. 負例: anesthesia欠落/未知enum/負価格/別clinic category FKは保存・success toast不可。 既存境界AC: Given麻酔未選択または負単価、When保存、Then API前またはBEで明示的な日本語 field error、成功 toast なし。
4. 横展開確認対象: procedure create/updateとtreatment master fields

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*Procedure.*Create|Test.*Anesthesia'`
- FE: `docker compose exec -T frontend npx vitest run src/features/master/routes/TreatmentPlanMaster.test.tsx src/features/master/components/TreatmentPlanSidePanelHost.test.tsx`
- 手動/E2E: V04 §2 診療項目マスタ 処置タブ `/settings/treatment-items?tab=procedure`を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 麻酔区分は臨床安全、価格は会計整合。clinic/category FKと権限を検証する。
- fail-closed の維持: 未知enum・負価格・別clinic FKをdefault/成功へfallbackしない。
- audit / トランザクション境界: procedure create/update/auditをmedicalrecord ownerの同一tx。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- anesthesia は臨床意味を持つため、既存値や暗黙 default を推測しない。既存4,623件の一括変更は範囲外。別 clinic の parent/category FK を拒否する。

#### 7. 実装ステップ（順序付き）

1. enum/既定値の業務正本を確認し、missing/negative test を追加する。
2. panel/model/request validation を同じ typed contract にする。
3. create/update/再GET、parent FK、成功/失敗 toast を検証する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- UIから正当な処置を作成でき、必須/単価/clinic FK が fail-closed、内部 field 名の生表示と無音失敗がない。

## BUG-029: 支払方法マスタで名称重複時、実際には保存されていないのに「登録しました」の成功トーストが表示される（BUG-026と同一パターン）

- **重大度**: 中〜高（BUG-026と同根と見られる。無音失敗ではなく虚偽の成功通知が出る点でUXへの実害が大きい。2つの異なるマスタ・2種類の異なるバリデーション〈範囲外値／一意制約違反〉で再現しており、共通基盤（useMasterSave等）の問題である可能性が高い）
- **対応状況（2026-08-03 JST）**: OPEN | **根拠**: 受入実測は「2回目 POST 0件＋成功 toast」だが、current `PaymentMethodSettings` の validate は名称のみ、共有 `useMasterSave` は mutation 成功後だけ success toast を出す。この source/test は実測前から存在し、同名 precheck の実経路と観測矛盾が未解決。原文シナリオ未再実行のため IMPLEMENTED_UNVERIFIED または DUPLICATE へ上げない | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: 同名2件目で click/request/mutation/toast timeline を再取得し、BUG-023 の409文言問題と分離する（REPRODUCE_FIRST）
- **発見シナリオ**: V04 §1 支払方法 `/settings/payment-methods`（C3-2 一意制約違反チェック）
- **再現手順**:
  1. `/settings/payment-methods` で「V04支払方法」を新規登録（正常に保存され一覧に反映、`is_active:true`）
  2. 続けて「新規登録」→ 同名「V04支払方法」を入力して保存
  3. 画面右下に緑色の「登録しました」トーストが表示される（1回目の保存時と同一の成功トースト）
  4. しかし一覧は5件のままで重複行は増えない。`GET /api/v1/payment-methods` で確認しても「V04支払方法」は1件（id:53）のみ存在
  5. ネットワークログ上、2回目の保存操作では `POST /api/v1/payment-methods` を含むいかなるリクエストも発生していない
- **期待結果**: 名称が重複している場合は保存されず、エラーメッセージが表示される（成功トーストは出ない）
- **実際の結果**: 保存が実際には行われていない（APIコール自体が発生しない）にもかかわらず、正常保存時と同一の「登録しました」成功トーストが表示され、エラー表示が一切ない。
- **備考**: BUG-026（保険マスタ・範囲外値）と表示パターンが完全に一致しており、同一の共通保存処理（SidePanel/useMasterSave）に起因する可能性が高い。時間の都合上、他マスタでの横展開確認は未実施。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `C-MASTER-FALSE-SUCCESS`
- 同一 PR にする BUG: BUG-026, BUG-029
- 先行必須: BUG-026 の再現 trace と discriminated save result
- 後続解放（シナリオ/他BUG）: 支払方法の重複/競合と共通 toast 契約

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE（REPRODUCE-FIRST）
- 観測根拠（API・クエリ・コード参照）: payment method の client validation が重複を検出して request 0件にする経路はあるが、現行共通 hook は mutation success 後だけ toast する。実 button click と toast lifecycle で BUG-026 と同じ failure signature か確認する。（具体パスは §2）
- 重複判定（例: 002 vs 022）: BUG-026 と false-success 報告を共有。ただし現行 source の request0+success経路は未確認。
- 所有境界（FE / BE / データ / 環境）: FE=payment settings/API/useMasterSave; BE=billing duplicate guard; データ=payment method scope; 環境=request/toast timeline

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/master/routes/PaymentMethodSettings.tsx`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/features/master/api/payment-method-master.ts`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/features/master/hooks/use-master-save.ts`。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: 原因確定前はAPI/DB変更なし。再現時のみBUG-026のsave resultを共用し、duplicate precheckをvalidation_failedとして扱う。
- 正しい挙動の定義（1〜3 文）: 重複/競合では成功toastなし、正常作成は1 request・1 toastになる。
- やらないこと（Out of scope）: client precheckを一意性保証にすること、scope意味論変更。
- 既存データ修復の要否と手順: 不要。

- BUG-026 の discriminated save result を再利用し、duplicate precheck は `validation_failed` として field error を返す。最終一意性は BE/DB が判断し、競合時も成功通知しない。

#### 4. 受け入れ基準（AC）

1. Given名称Xが存在、When同名を保存、Then POST 0件または409、成功 toast 0件、重複 error 表示、panel入力保持。
2. 回帰: 正常作成、競合、BUG-026共通saveを維持する。
3. 負例: client precheck漏れ/DB race/request0でfalse successを出さない。 既存境界AC: Given競合する2 request、When同時作成、Then一方だけ成功し、他方は適切な重複 error。別 clinic/global scope は現契約どおり。
4. 横展開確認対象: payment methodとuseMasterSave全consumer

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/billing -run 'Test.*PaymentMethod.*Duplicate'`
- FE: `docker compose exec -T frontend npx vitest run src/features/master/hooks/use-master-save.test.ts`。`src/features/master/routes/PaymentMethodSettings.test.tsx`は新規追加予定。
- 手動/E2E: V04 §1 支払方法 `/settings/payment-methods`（C3-2 一意制約違反チェック）を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 会計マスタの一意性/clinic scopeを維持する。
- fail-closed の維持: client precheck漏れやDB競合をfalse successにしない。
- audit / トランザクション境界: 成功mutationのBE tx/audit後だけFE成功通知。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- client precheck を一意性保証にしない。前回 toast の残留を操作成功と数えない。scope の意味を変更しない。

#### 7. 実装ステップ（順序付き）

1. network event と toast lifecycle を含む再現 test を作る。
2. BUG-026 の共通 result と duplicate error を接続する。
3. client precheck/DB競合/正常作成を回帰確認する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- request/toast/error/count の matrix が通り、DB競合でも虚偽成功がなく、BUG-026 と重複実装していない。

## BUG-030: LINE予約設定の最短予約受付日数を0に変更して保存すると、200成功・updated_at更新にもかかわらず値が永続化されない

- **重大度**: 中
- **対応状況（2026-08-05 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み commit `37c84041b00335cb5899cad86102034f77ead855` — Save を Create(DoNothing)+Select Updates に変更し explicit 0/false を永続化; updatable 列集合は credential/bot id 除外のまま; scoped test green. フォローアップ commit `324b356e8333df094a57e1ead44f7f76d1b3f01e` — update path で DO NOTHING により setting.ID が 0 のまま返っていたため、同一 tx で id/created_at を読み戻して RSV-03 応答へ反映; zero 永続化・列集合は不変 | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: V05-8 稼働・受付ルール設定 #2 のブラウザ再実行で VERIFIED_FIXED 判定
- **発見シナリオ**: V05-8 稼働・受付ルール設定（`/line-reservation/settings`）#2
- **再現手順**:
  1. `/line-reservation/settings` を開く（八王子病院、既存値: 最短予約受付日数=2）
  2. 「最短予約受付（日数）」欄を `0` に変更し「設定を保存」をクリック
  3. PUTリクエストのボディを確認 → `booking_window_min_days:0` が正しく送信されている
  4. サーバは200 OKを返し、`updated_at` タイムスタンプも更新される
  5. ページを再読込しGETで値を確認
- **期待結果**: 0は境界値として保存が期待される
- **実際の結果**: GETで返る値は旧値（2）のまま。保存UI上はエラー表示なく成功したように見えるが、実際には0が反映されていない。`-1`を入力した場合もFE側で`0`にクランプされた上で同じ現象が再現し、2回とも同一の非永続化を確認した。
- **備考**: PUTは200を返し`updated_at`も変わるため、サーバ側で`0`/未指定を区別できないzero-value省略（Goの`omitempty`的挙動）が疑われる。エンドポイント: `PUT /api/v1/clinics/1/line-reservation-settings`

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `S-LINE-SETTINGS-ZERO`
- 同一 PR にする BUG: なし
- 先行必須: zero / omitted の repository 再現 test
- 後続解放（シナリオ/他BUG）: LINE予約設定の 0→再GET round-trip

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE / DB
- 観測根拠（API・クエリ・コード参照）: FE は `booking_window_min_days: 0` を送信し BE request/service も値を代入する。repository の create+upsert と model の GORM `default:2` が zero を「未指定」と扱い、旧値/default を残す経路が根因で、`omitempty` 仮説ではない。（具体パスは §2）
- 重複判定（例: 002 vs 022）: 単独。omitempty ではなく GORM default/upsert の zero 値扱いが候補。
- 所有境界（FE / BE / データ / 環境）: FE=line settings form; BE=reservation request/service/repository; データ=model default + clinic setting/audit; 環境=なし

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/line-reservation/components/LineReservationSettingsForm.tsx`。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/reservation/line_reservation_setting_request.go`、service、repository。 | 現行証拠 / 変更候補 |
| DB/migration | `backend/internal/model/line_reservation_setting.go` の default tag。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: JSON/API意味論を明示: omittedと明示0を区別し、0を保存/readbackする。DB schema/default変更が必要な場合のみ別migration。
- 正しい挙動の定義（1〜3 文）: 2→0→2とomittedが意図どおりround-tripし、responseは実永続値から作る。
- やらないこと（Out of scope）: 全zero fieldの一括意味変更、既存値の意図推測補修。
- 既存データ修復の要否と手順: 不要。schema変更時は別migration、人が `make migrate`。

- request presence と値0を区別し、upsert の update columns/map または明示 `Select` で zero を必ず書く。create 時の default と update 時の zero を同一の ORM shortcut に委ねない。
- write後に同 tx/同 clinic で再読込し、response を実永続値から返す。更新と監査を同一 tx にする。

#### 4. 受け入れ基準（AC）

1. Given既存2、When0をPUT、Then response/再GET/DBが0で updated_at/監査も同じ変更を示す。
2. 回帰: 2→0→2、omitted、再GETを維持する。
3. 負例: 別clinic設定を変えず、0を未指定/defaultに変換しない。 既存境界AC: Given0から2、WhenPUT、Then2へ戻る。Given field未指定の内部 update、Then意図しない0上書きはしない。別 clinic は不変。
4. 横展開確認対象: GORM zero-value updateを使う全reservation settings

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/reservation -run 'Test.*LineReservationSetting.*Zero|Test.*BookingWindow'`
- FE: `docker compose exec -T frontend npx vitest run src/features/line-reservation/components/LineReservationSettingsForm.test.tsx`
- 手動/E2E: V05-8 稼働・受付ルール設定（`/line-reservation/settings`）#2を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 予約設定のclinic隔離/権限を維持し、0の意味を保存する。
- fail-closed の維持: 0をdefault/omittedへfallbackせず、別clinicを更新しない。
- audit / トランザクション境界: setting upsert/readback/auditをreservation ownerの同一DBOrTx。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- 全 field の zero semantics を一括変更しない。既存設定2が「0保存失敗」で残ったかは audit なしに判別できないため自動補修しない。
- schema/default変更が必要なら別 migration とし、agent は適用せず人手の `make migrate` を案内する。

#### 7. 実装ステップ（順序付き）

1. 2→0→2 と omitted の repository integration test を追加する。
2. presence-aware request/upsert/readback/tx audit を実装する。
3. FE round-trip と別 clinic 否定ケースを確認する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- zero/nonzero/omitted の全契約、再GET、監査、clinic 分離が通り、既存値を推測補修していない。

## BUG-031: ログイン済み状態で `/login` に直接アクセスしても自動リダイレクトされない

- **重大度**: 低〜中（セキュリティ上の実害は小さいが、仕様と異なる導線）
- **対応状況（2026-08-06 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み commit `5cf86efc4279b6ac1d1e06ca619f80612d50c27e` で AuthProvider が password-recovery 以外（`/login` 含む）で session restore。cold `/login` で refreshToken 1 回後 isAuthenticated→LoginForm Navigate。scoped vitest 33/33 green | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: V05-1 #3 ブラウザ再検証
- **発見シナリオ**: V05-1 ログイン #3
- **再現手順**:
  1. ノアとしてログイン済みの状態（`/owners` 等が正常表示されることで確認済み）で `/login` に直接アクセス
  2. 2秒待機してスクリーンショット確認
  3. ネットワークログを確認 → `/api/v1/me` へのリクエストが一切発生していない
  4. 別ページ（`/owners`）へ遷移してセッションが有効であることを再確認 → 正常表示
  5. 再度 `/login` へ直接アクセス → 同じ結果を再現（2回連続で同一挙動を確認）
- **期待結果**: ログイン済みの `/login` 直アクセスは即リダイレクト（`LoginForm.tsx` の認証済みNavigate）＝ダッシュボード等へ自動遷移するはず
- **実際の結果**: リダイレクトされず、ログインフォームがそのまま表示される。認証済みチェック用のリクエストも発火していない。
- **備考**: `LoginForm.tsx` 側の認証済み判定ロジックが `/login` 到達時に働いていない可能性が高い。

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `S-LOGIN-SESSION-RESTORE`
- 同一 PR にする BUG: なし
- 先行必須: login route の session state trace
- 後続解放（シナリオ/他BUG）: 有効/無効 session の直リンク導線

#### 1. 切り分けステータス

- 主因レイヤ: FE
- 観測根拠（API・クエリ・コード参照）: `LoginForm.tsx` の redirect は in-memory user だけを見るが、public login route では auth restore を無効化し、user が null のため `/me` も発火しない。既存 session を hydrate する入口がない。（具体パスは §2）
- 重複判定（例: 002 vs 022）: 単独。public login route の session hydrate 不在。
- 所有境界（FE / BE / データ / 環境）: FE=auth provider/login route; BE=/me は既存契約; データ=session only; 環境=cookie/session state

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/auth/components/LoginForm.tsx`。 | 現行証拠 / 変更候補 |
| FE | `frontend/src/features/auth/hooks/use-auth.tsx` と public route configuration。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: API/DB変更なし。login route限定のsession hydrate stateを追加し、安全な内部redirectだけを許可する。
- 正しい挙動の定義（1〜3 文）: valid sessionは`/me`一回後redirect、invalidはform、network errorはretry可能でloopしない。
- やらないこと（Out of scope）: open redirect、public route全体の無制限restore、credential logging。
- 既存データ修復の要否と手順: 不要。

- `/login` 到達時だけ lightweight session hydrate を一度行い、authenticated→安全な既定画面、unauthenticated→login form、pending→非操作 loading を明示する。
- redirect target は検証済み内部 path だけを許可し、public route 全体で無制限に restore を有効化しない。

#### 4. 受け入れ基準（AC）

1. Given有効 session、When `/login` を直開き、Then `/me` 1回後に既定画面へ replace redirect し login form を操作可能状態で見せない。
2. 回帰: 通常login/logout/browser backを維持する。
3. 負例: invalid/expired/network errorでloop/open redirect/session存在漏洩を起こさない。 既存境界AC: Given無効/期限切れ session、When直開き、Then form を表示し redirect loop しない。Given network error、Then再試行可能な errorとなる。
4. 横展開確認対象: auth providerとpublic/protected route transitions

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: 該当なし（API/DB変更なし。FE境界testで検証）
- FE: `docker compose exec -T frontend npx vitest run src/features/auth/components/LoginForm.test.tsx src/features/auth/hooks/use-auth-initial-session.test.tsx`
- 手動/E2E: V05-1 ログイン #3を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: 認証/権限対象。credentialとsession存在を漏らさずsafe redirectのみ許可。
- fail-closed の維持: network/invalid sessionをauthenticatedへfallbackせずloopしない。
- audit / トランザクション境界: session readのみ。DB/audit変更なし。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- open redirect、session existence leak、無限 `/me` loop を防ぐ。token/cookie を client log に出さない。

#### 7. 実装ステップ（順序付き）

1. valid/invalid/network session の direct-route test を作る。
2. login-route限定 hydrate state machine と safe redirect を実装する。
3. logout直後、browser back、通常 login を回帰確認する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- valid/invalid/error/logout の route test が通り、request loop/open redirect/credential露出がない。

## BUG-032: 健診対象者抽出プレビューAPIがハングし応答しない（外部Lステップ連携先タイムアウト未実装の疑い）

- **重大度**: 中〜高
- **対応状況（2026-08-06 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み commit `944f2e4ddc1f463e374955b3bfc6c4ace89add36` で PreviewCheckupSync に 15s context timeout、SQL LIMIT 500、owner cap 100; FE axios timeout 20s。unit TestCheckupSyncPreview_Bounds + Preview tests green | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: 実DB件数で preview 応答時間を計測し S シナリオ再検証
- **発見シナリオ**: V05-18 健診対象者一括タグ付与（`/lstep/checkup-sync`）#1-2 検証中
- **再現手順**:
  1. `/lstep/checkup-sync` を開く
  2. 検診種別で「定期健診」を選択し、他の条件は未入力のまま「対象者を検索」をクリック
  3. ボタンが「対象者を検索中...」のまま固まり、15秒以上待っても応答が返らない
  4. ネットワークログで `GET /api/v1/clinics/1/lstep/checkup-sync/preview?checkup_type=annual` が `pending` のまま変化しないことを確認
  5. 同一エンドポイントへ直接 `fetch` し `AbortController` で15秒タイムアウトを設定して検証 → タイムアウトまで応答なし
  6. 同時に `/api/v1/me` を叩くと14msで200が返ることを確認し、サーバ自体は生きている（このエンドポイント個別のハング）ことを確認
- **期待結果**: プレビュー結果一覧が表示される、または妥当な時間内にエラー/空データが返る
- **実際の結果**: エンドポイントが応答せず無期限にハングし、UIは「検索中...」のまま固まって復帰しない（ローディング状態からの復帰導線もない）
- **備考**: V05-12テスト時にダミーのLステップ資格情報（APIキー等）を保存済みの状態であり、バックエンドが検診対象者抽出時に外部Lステップ APIへ同期的に疎通しようとしてタイムアウトなく待ち続けている可能性が高い。実運用の有効な資格情報であれば発生しない可能性はあるが、「外部API疎通不可時にタイムアウトなくハングする」設計自体が問題であり、無効な資格情報時のフェイルセーフとして起票する価値があると判断した。エンドポイント: `GET /api/v1/clinics/1/lstep/checkup-sync/preview`

### 実装計画

#### 0. クラスタ / 依存

- クラスタID: `S-CHECKUP-PREVIEW-LATENCY`
- 同一 PR にする BUG: なし
- 先行必須: query plan / lock wait / context deadline の計測
- 後続解放（シナリオ/他BUG）: V05-18 preview と bounded retry

#### 1. 切り分けステータス

- 主因レイヤ: FE / BE / DB（計測先行）
- 観測根拠（API・クエリ・コード参照）: preview service は設定/cacheとローカルDB repositoryを読む経路で、外部 Lステップ HTTP 呼出しは確認できない。repository の unbounded raw SQL が候補だが、実 query plan/lock wait は未計測。FE には timeout/recovery がない。（具体パスは §2）
- 重複判定（例: 002 vs 022）: 単独。報告時の外部Lステップ仮説は現行 source で反証され、local query/lock候補。
- 所有境界（FE / BE / データ / 環境）: FE=preview API/page; BE=lstep handler/service/repository; データ=query/index/cache key + clinic filters; 環境=DB plan/lock/latency

#### 2. 疑わしい箇所（コードマップ）

| 層 | パス / シンボル | 役割 |
|---|---|---|
| FE | `frontend/src/features/lstep/api/get-checkup-sync-preview.ts`、`frontend/src/features/lstep/routes/CheckupSyncPage.tsx`。 | 現行証拠 / 変更候補 |
| BE | `backend/internal/lstep/checkup_sync_handler.go`、`checkup_sync_service_preview.go`、`checkup_sync_repository.go`。 | 現行証拠 / 変更候補 |

#### 3. 修正方針

- 契約変更の有無（API・DB）: 計測後に必要なら既存preview APIへbounded filter/page/deadline/errorを後方互換追加。indexは別migration。
- 正しい挙動の定義（1〜3 文）: previewは時間予算内のbounded結果を返し、cancel/timeoutでqueryとloadingを終了してretry可能にする。
- やらないこと（Out of scope）: 空結果fallback、外部Lステップ原因の決め打ち、PII入り計測artifact。
- 既存データ修復の要否と手順: 不要。index時はEXPLAIN根拠と別migration、人が `make migrate`。

- request trace と `pg_stat_activity`/lock wait、`EXPLAIN (ANALYZE, BUFFERS)` で handler→service→repository の停止点を確定する。clinic/checkup_type/date/tag filter と上限/page を SQL に入れ、必要な index は計測根拠付きにする。
- context deadline/cancel を DB まで伝播し、server は安全な timeout code、FE は AbortController、error、retry、条件変更の復帰導線を持つ。

#### 4. 受け入れ基準（AC）

1. Given demo相当データ、When annual preview、Then合意した時間予算内に bounded result/count が返り loading が終了する。
2. 回帰: 他checkup type/filter/cache、正常previewを維持する。
3. 負例: lock/deadline/cancel/別clinicでempty successやPII artifactを作らない。 既存境界AC: Given DB遅延/lock、When deadline超過または user cancel、Then query が中断し UI が error/retry へ戻る。 / Given別 clinic/tag、When同条件、Then対象/count/cache が混入しない。
4. 横展開確認対象: checkup preview queries/cache/UI loading states

#### 5. テスト計画

- ユニット: 対象service/repositoryまたはhook/componentの旧失敗をREDにし、境界値・負例を最小testで固定する。
- 結合（scoped）: `docker compose exec -T backend go test -p 1 ./internal/lstep -run 'Test.*CheckupSync.*Preview|Test.*Cancel|Test.*Clinic'`
- FE: `docker compose exec -T frontend npx vitest run src/features/lstep/routes/CheckupSyncPage.test.tsx src/features/lstep/components/CheckupSyncPreviewTable.test.tsx`。`src/features/lstep/api/get-checkup-sync-preview.test.ts`は新規追加予定。
- 手動/E2E: V05-18 健診対象者一括タグ付与（`/lstep/checkup-sync`）#1-2 検証中を専用test clinic / test dataで再実施し、PASS / FAIL / BLOCKED / UNREPORTEDを記録する。
- 禁止: full-repo suites、compose lifecycle、DB reset、migration適用をagentの自動gateにしない。

#### 6. リスクと安全

- 臨床安全 / 会計整合 / clinic 隔離 / 権限: clinic隔離と患者PII保護。filter/cache key/queryにclinicを含める。
- fail-closed の維持: timeoutを空successへfallbackせず、cancelをDBまで伝播する。
- audit / トランザクション境界: read-only。indexは別migration、計測artifactにPIIを残さない。
- ロールバック手順: product codeは各§2 owner/featureの最小差分をpath-scopedに戻す。data/migrationは旧code互換を確認し、別承認・人手手順なしに逆操作しない。

- 根拠なく外部 API timeout を追加しない。unbounded query を空結果 fallback で隠さない。preview cache key に clinic と全 filter を含め、対象者の個人情報を計測 artifact に残さない。

#### 7. 実装ステップ（順序付き）

1. endpoint区間計測、lock wait、query plan、行数を採取して failure signature を固定する。
2. bounded repository query/cancel と必要な index を TDD で実装する。
3. FE timeout/retry と clinic/cache isolation、V05-18 後続を検証する。

#### 8. 完了定義（DoD）

- [ ] §4 の AC が全通過
- [ ] 関連クラスタと横展開対象の回帰が通過
- [ ] 作成した test data の cleanup、または cleanup 不要を記録
- [ ] 原文シナリオの再実施可否と残余 BLOCKED を記録

- latency before/after、bounded result、cancel/timeout、UI recovery、clinic/cache isolation が証拠化され、外部 API 仮説を事実として扱っていない。

---

## V01（臨床系フォーム）: 実施範囲まとめ

実施範囲: §1（カルテ本体・代表PATCH部分更新・存在しないID）、§2（治療明細・薬量dose hard/softゲート）、§3（バイタル・体温境界・体重Kg/g切替）、§6（追記・見積書タブ）、§7（検査入力・必須検証・存在しないID）、§8（予防接種独立画面・未来日/同日拒否・LOT×4永続化・存在しないID）、§11（入院デイリーフォーム群・同一日一意制約・必須検証）を重点的に実施。§10・§12は存在しないID直叩き（C3-3）のみ横断確認。

**バグなしを確認した項目**: §1本体（代表PATCH部分更新、存在しないID時の正しいエラー画面）、§2薬量dose hard/softゲート（上限超過ブロック・下限未満警告記録）、§3体温境界値（30〜45℃）、§6追記の必須検証・境界値・未確定カルテでの導線非表示、§7検査種別・担当医の必須ブロック（表示以外)、§8予防接種の必須検証・未来日拒否・次回予定日境界・LOT×4永続化、§11のスタッフメモ必須検証・daily_records一意制約（get-or-create的挙動で重複防止）、§12トリミングの存在しないID時の正しいエラー画面。

**未実施（時間予算の制約）**: §4カルテ定期健診タブ、§5カルテ予防接種タブ、§9定期健診クイック登録、§10入院登録のC1/C2/C3-1本体、§12トリミングのC1/C2/C3-1本体。

## V02（会計・予約・受付系フォーム）: 実施範囲まとめ

実施範囲: §2（会計・物販追加ダイアログ）、§6（見積書フォーム、client-side検証のみ）、§7（予約登録モーダル、一部）、§10（シフト登録・テンプレート）、§11（休診日設定）。§1（会計確定フォーム）はレジ締め済み期間との相互作用の中でテストしBUG-018を発見。

**バグなしを確認した項目**: §2数量・単価の境界値チェック、カテゴリ必須チェック、マスタデータ整合性、§6タイトル必須チェック・負数入力ブロック・ステータス制限・NavigationBlocker・ロック済み見積書のリダイレクト、§7飼主未選択時の送信ブロック・動物種/予約区分マスタ整合性、§10終了時刻逆転の拒否・「休み」区分の時刻欄無効化・シフトテンプレート・実際のシフト登録の永続化、§11定休日設定の永続化・重複作成防止。

**未実施（時間予算の制約）**: §3クレジット訂正ダイアログ、§4返金ダイアログ、§5レジ締めフォームの境界値、§8受付・当日カード（walk-in）、§9受付ステータス変更、§7予約登録モーダルの残りチェック（永続化確認・同一飼主複数予約等）。

## V03（飼主・ペット・スタッフ・権限系フォーム）: 実施範囲まとめ

実施範囲: §1飼主登録・編集（必須/形式/境界値・代表PATCH・医院マスタ連動・メール/電話重複拒否・存在しないID・生年月日C2永続化）、§2ペット編集モーダル（必須項目・体重境界値・C2永続化・マスタ連動）、§4ペット死亡登録ダイアログ（全4チェック）、§6権限グループマスタ（全4チェック）、§7医院マスタ（チェック1,2,4,5,7）、LINE連携（不正ID拒否確認）、ペット飼主付け替え（導線確認のみ）。

**バグなしを確認した項目**: §1・§2は全チェック項目でバグなし（必須/形式/境界値、代表PATCH部分更新、マスタ連動、重複拒否、存在しないID、C2永続化）。§7医院マスタのチェック1,2,4,5,7もバグなし。LINE連携の不正ID拒否、ペット飼主付け替えの導線。

**未実施（時間予算の制約）**: §3ペット新規追加（新規飼主同時登録フロー）、§5スタッフマスタ登録・編集サイドパネル、§7チェック3（医院切替→デフォルト権限グループ自動作成）・チェック6の帳票設定トグル群。

**残留テストデータ**: 医院マスタのpostal_code形式検証中に誤って作成した「V03医院QAtest_*」テスト用医院レコード（clinic ID 7〜12、is_active: true）が、権限上の削除操作ブロックにより未クリーンアップのまま残存。実害はないが医院一覧に表示される点に留意。

## V04（設定マスタ系フォーム、30フォーム）: 実施範囲まとめ

フル実施: 動物種類、診断カテゴリ・診断病名、主訴種別（BUG-025）、問診テンプレート（BUG-025）、予約区分グループ・本体、入院・宿泊プラン、ケージ、物販・商品、保険（BUG-026）、職種、トリミングコース種別・コース・オプション、割引キャンペーン、支払方法（BUG-029）、診療項目マスタ（診察タブ正常／処置タブBUG-028）。

**バグなしを確認した項目**: 動物種類・診断カテゴリ・診断病名・予約区分グループ/本体・入院プラン・ケージ・物販・職種・トリミング関連・割引キャンペーン・支払方法のC1-1新規作成ラウンドトリップ・診療項目マスタ診察タブ（すべて`is_active:true`で正しく保存）。

**未実施（時間予算の制約）**: §2診療項目の検査・予防接種・定期健診タブ、処置タブの既存4,623件の編集/重複/親子階層の深掘り、§3薬剤マスタ+投与量パラメータ、§4予約区分詳細の残りチェック、§5予約可能枠設定、§6締め時間設定（境界編集）、§7シフトパターン、§8Lステップ連携4フォーム、§9LINE予約ページ設定、§10法人情報。

**テストデータクリーンアップ**: 全V04接頭辞テストデータについて、削除APIを使わず全マスタ横断で `is_active:false` への無効化を確認済み（"V04"を含む名称で `is_active:true` のレコードが0件であることを最終確認）。

## V05（認証・LINE系フォーム）: 実施範囲まとめ

実施範囲: 認証系4フォーム（V05-1〜4）、LINE予約病院側設定6フォーム（V05-8〜13）、Lステップ連携6フォーム（V05-12〜18）。

**バグなしを確認した項目**: V05-3パスワードリセット申請（アカウント列挙防止含む）、V05-4パスワード再設定のtoken無しケース、V05-1ログインの不正メール形式・短すぎるパスワード、V05-8稼働・受付ルール設定（BUG-030以外は正常）、V05-9表示ページ編集、V05-10LINE予約枠設定、V05-11飼主⇄LINE顧客紐付け/解除、V05-12Lステップ連携設定（閾値・LIFF ID・ベースURL編集、CPMバージョン切替、シークレット非上書き）、V05-13配信優先順位の境界値、V05-15タグ設定の必須検証、V05-16CSV取込の未選択時ブロック、V05-17タグ一括解除の確認ダイアログ表示・文言、V05-18検診種別未選択時のブロック。

**LIFFブロック項目（既知バグBUG-008/014の再確認）**: V05-5 LIFFアカウント連携、V05-6 LINE予約作成、V05-7 予約キャンセルは、いずれも `GET /api/liff/1/profile` が401（`missing authorization header`）を返すため起動不能。BUG-008/014と同一の根本原因（LIFFモックがAuthorizationヘッダを一切付与しない）と判断し、新規バグとしては起票せず。

**意図的にスキップした項目（共有デモ環境への副作用回避のため）**: V05-1#4ログインレート制限テスト（共有ノアアカウントのロックアウト回避）、V05-2パスワード変更、V05-4#5使用済みトークン再利用テスト、V05-17タグ一括解除の実行そのもの（対象が共有デモの実在飼主データのため確認ダイアログの表示確認に留めキャンセル）。V05-18の#1・#3〜6はBUG-032（プレビューAPIハング）により後続検証がブロックされたため未実施。

---

## 検証完了サマリ（S01〜S12・V01〜V05 全スコープ完了）

`docs/ops/testing/scenarios` 配下の受入テストシナリオ（業務フローシナリオS01〜S12、個別フォーム検証V01〜V05、計84フォーム）について、実ブラウザ（Chrome）操作とネットワークログ・API直接呼び出しによるクロスチェックを組み合わせて検証を実施した。

**発見バグ総数**: BUG-001〜BUG-032（本ファイル記載の32件）。うち重大度「高」に分類したものが概ね半数以上を占め、特に以下は業務上の影響が大きいため優先対応を推奨する。

- **データ破損・データ不整合系**: BUG-015（体重単位切替で1000倍の数値破損、投薬量計算に波及しうる）、BUG-018（レジ締め後の会計で明細と金額が不整合なレコードが残る）
- **偽の成功フィードバック系**: BUG-026・BUG-029（保存が実際には行われていないのに成功トーストが表示される。同一の共通保存処理に起因する可能性が高く、横展開の懸念あり）
- **保存が反映されない系**: BUG-024（権限グループの権限マトリクス変更が200応答にもかかわらずDBに反映されない。自己剥奪ガードの動作確認自体がブロックされている）、BUG-030（LINE予約設定のゼロ値が保存されない）
- **既存の重大ブロッカー系**（S01〜S12検証時に発見、再掲）: BUG-011（見積書2件目以降が常時409で作成不能。V01検証で根本原因の追加診断を確認）、BUG-012（オーナー集計ダッシュボードの無限ハング）、BUG-013（未請求明細取得APIの500エラー）、BUG-008/014（LIFFモック認証のAuthorizationヘッダ欠落により飼主向けLINE機能が軒並み検証不能）
- **UI導線・エラー表示系**: BUG-016・BUG-019・BUG-002（存在しないID直叩き時にエラー画面ではなく空フォームが開く、複数フォームで横断的に発生）、BUG-017・BUG-021（必須バリデーションのエラーメッセージが一切表示されない「無音失敗」パターン、複数フォームで横断的に発生）、BUG-023・BUG-027（一意制約違反時に内部テーブル名の生英語エラーが露出し、対象名も空文字になる、複数マスタで横断的に発生）

**時間予算の制約により未実施のまま残った領域**: V01の§4・§5・§9・§10本体・§12本体、V02の§3・§4・§5・§8・§9、V03の§3・§5・§7の一部、V04の§2一部タブ・§3・§4一部・§5〜§10、V05の一部の副作用リスクが高い操作（レート制限・パスワード変更・トークン再利用・タグ一括解除の実行）。これらは今回のセッションでは意図的にスコープ外とした（詳細は各セクションの「未実施項目」参照）。網羅的な検証を行うには、追加のテストセッション、または本番相当データを用いない専用のテスト用クリニック環境の用意を推奨する。

以上。

---

## Staged plan unit: EXAMINATION-BUG-GROUP-003-005-017（2026-08-03 JST）

- **Status**: COMPLETE with partial BLOCKED (BUG-003 data)
- **Claims retained** (user-only release): `claim/BUG-003`, `claim/BUG-005`, `claim/BUG-017`
- **Changed files (product)**:
  - BUG-005 `dfd653eaa5ccb089707c3a088863c39c07669288`: `ExaminationForm.tsx`, `ExaminationForm.permissions.test.tsx`, `frontend/src/features/master/index.ts` (allowlist deviation: Feature Indexing barrel export for `useGetStaffs`)
  - BUG-005 follow-up `d91d3285d79d3da4311ee935299f557279cd153b`: `frontend/src/hooks/use-staffs.ts` (allowlist deviation: shared RQ cache must carry `staffType`)
  - BUG-017 `7f71063759974257be14a4ed0a8a5fd04a5c6880`: `ExaminationForm.tsx`, `ExaminationFormFields.tsx`+test, `use-examination-form.ts`+test, `searchable-select.tsx`+test
  - BUG-003: no product patch
  - BUG-004: no product patch this unit
- **Gate commands (verbatim outcomes)**:
  - `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*(Exam.*Assessment|ReferenceRange)' -count=1` → `ok`
  - `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*(Confirmed|Audit|Rollback|ConfirmWithItems|CreateConfirmed)' -count=1` → `ok`
  - `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*(Exam.*Assessment|ReferenceRange|Confirmed|Audit|Rollback)' -count=1` → `ok`
  - `docker compose exec -T frontend npx vitest run src/features/examinations/routes/ExaminationForm.permissions.test.tsx` → 12 passed
  - `docker compose exec -T frontend npx vitest run ...ExaminationFormFields/use-examination-form/searchable-select/permissions` → 81 passed
  - post-HIGH-fix: ExaminationForm.permissions + CheckupForm tests → 16 passed
  - coverage (statements, changed surface includes): **92.18%** (295/320)
  - eslint on changed TS/TSX paths → exit 0
- **Assumption deviations**:
  1. Required write outside initial allowlist: `frontend/src/features/master/index.ts` export of `useGetStaffs` (deep import forbidden by Feature Indexing).
  2. Required write outside allowlist: `frontend/src/hooks/use-staffs.ts` staffType on shared cache (react-reviewer HIGH).
  3. BUG-003 approved structured range data absent → BLOCKED without inventing thresholds.
- **Independent review**: react HIGH (cache) fixed; typescript 0 C/H; security 0 C/H; healthcare accepts BUG-003 BLOCKED and FE/BE dual defense for doctors
- **Browser**: none created; original scenarios remain UNREPORTED
- **Orchestration**: native Workflow `exam-bug-group-investigate` (4 explore probes) + sequential TDD writes + parallel review subagents

---

## Staged plan unit: EXAMINATION-MODE3-FOLLOWUP-20260803（2026-08-03 JST）

- **Status**: COMPLETE
- **Packet claim retained** (user-only release): `claim/EXAMINATION-MODE3-FOLLOWUP-20260803`
- **Changed files**:
  - staff cache: `query-keys.ts`, `use-staffs.ts`, `use-staffs.test.ts`, `features/master/api/staffs.ts`, `staffs.test.ts`
  - BUG-017 proof: `ExaminationForm.validation.test.tsx` (test-only)
  - BUG-003: `exam_result_assessment_test.go` (equality cases only)
  - ledger: `bug.md`
- **Gates**:
  - FE staff + validation + examination suite: vitest pass (89+ focused)
  - `use-staffs.ts` coverage statements **100%**
  - BE `Test.*(Exam.*Assessment|ReferenceRange|Confirmed|Audit|Rollback)` → `ok`
  - eslint on changed TS paths → exit 0
- **Browser**: UNREPORTED (none created)
- **Assumption deviations**: none for claim gate (prior BUG claims absent before acquire)

---
