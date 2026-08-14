## BUG-001: 飼主・ペット一覧の検索が「姓 スペース 名」形式でヒットしない

- **重大度**: 中（受付業務で頻出する検索パターンが機能しない）
- **対応状況（2026-08-07 JST）**: IMPLEMENTED_UNVERIFIED | **根拠**: 到達済み product fix `d7bf32f2214d6bb6c252b99b001d2ed2044de7c9`（`PetRepository.FindAll` 空白非依存フルネーム / `owners.id` 文字列一致の飼主No / `pets.pet_number` ILIKE / 空白のみ fail-closed / clinic-scoped JOIN）。Agent2 再検証（03:22 JST）: AC 補強 subtest `TestPetRepository_FindAll_Search/pet_number_partial_ILIKE` を含む scoped PASS（BE: `docker compose exec -T backend go test -p 1 ./internal/pet -run 'TestPetRepository_FindAll_Search|TestPetRepository_FindAll_Kana' -count=1 -v` exit 0, ok 0.359s; FE: `docker compose exec -T frontend npx vitest run` owners loaders/list/table 3 files 32 tests exit 0）。claim `claim/BUG-001` 保持、実装 branch `fix/BUG-001` @ `25ff98d97` | **原文シナリオ再検証**: UNREPORTED | **次のアクション**: S01 手順1を専用 synthetic fixture でブラウザ再検証し、人間/Agent3 が `VERIFIED_FIXED` 可否を判定（agent は VERIFIED_FIXED にしない）
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

[REDACTED_SECRET_LIKE_LINE]
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
