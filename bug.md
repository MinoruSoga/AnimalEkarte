# バグ一覧 (Animal Ekarte)

このファイルは受入テスト等で発見したバグを記録する。各エントリは発見日・重要度・再現手順・根拠を含む。

---

## BUG-2026-07-27-01: `GET /api/v1/pets` が常に 400 を返し、ペット検索/一覧が全滅する【重大・ブロッキング】

- **発見日**: 2026-07-27
- **発見経緯**: S01〜S06 再実行（受入テスト）の事前準備中に発見。2026-07-17時点のテストでは未発生だった新規リグレッション。
- **重要度**: Critical（ブロッキング） — 飼主・ペット検索を前提とするほぼ全業務フロー（カルテ作成のペット選択、飼主一覧画面、S01/S02/S03/S05/S06 の事前条件）に影響する。
- **症状**:
  - `/owners`（飼主・ペット一覧）画面がクラッシュし「エラー500」画面になる。
  - カルテ作成の「ペット選択」画面（`/medical-records/select-pet`）は検索条件を空にして検索しても常に「該当するペットが見つかりません（0件）」と表示される。トリミング・予防接種・検査・会計など他の `select-pet` 画面も同一 API を使うため同様に影響する可能性が高い。
- **再現手順**:
  1. `http://localhost:3003/medical-records/select-pet` を開く。
  2. 検索条件を何も入力せず「検索」を押す（または画面ロード時の自動検索を待つ）。
  3. 「該当するペットが見つかりません」（0件）と表示される。
  4. ブラウザの DevTools / ネットワークログで確認すると、実体は `GET /api/v1/pets?page=1&limit=20` が **HTTP 400** `{"error":"入力値が正しくありません"}` を返しており、フロントエンドはこれを「0件」として誤表示している（0件とエラーを区別していないUIバグも併発）。
  5. 直接 `fetch('/api/v1/pets?page=1&limit=20', {headers:{'X-Requested-With':'XMLHttpRequest'}, credentials:'include'})` を実行しても再現する。クエリパラメータを一切付けない `/api/v1/pets` のみでも同じ 400 が返る。
- **切り分け結果**:
  - 同一セッション・同一権限で `GET /api/v1/owners?page=1&limit=20` は **200 OK** で正常にデータを返す → 認証・セッション・クリニックスコープ解決自体は正常。障害は `pets` 一覧エンドポイント固有。
  - `GET /api/v1/pets/{id}`（単体取得）は正常に動作する（存在しないIDで期待通り404 `not found` を返す）→ 一覧（List）だけが壊れており、単体取得・詳細系は無事。
  - クエリパラメータの組み合わせ（`owner_id`, `search`, `species`, `include_deceased` の有無）を変えても常に同一の 400 本文になる → バインド/バリデーションの個別パラメータエラーではなく、`ListPets` ハンドラの奥（`pet` サービス/リポジトリ層、または共通エラー変換層）で発生している一般化されたフォールバックエラーの可能性が高い。
  - `backend/internal/pet/routes.go`, `pet_handler.go`, `pet_request.go` を読んだ限りコード上の明白な誤りは見当たらず、`repository.go` の `FindAll`（LEFT JOIN owners + `Order("owners.name_kana ASC, pets.id ASC")` を含むクエリ）が疑わしいが、稼働中コンテナのログ/実SQLエラーまでは本セッションのツールから確認できなかった（Docker はユーザー側マシンで稼働しており、サンドボックスから `docker compose logs` 等を直接実行できないため）。
- **回避策（本テストで採用）**: `GET /api/v1/owners` のレスポンスには各飼主に紐づく `pets` 配列がネストされて含まれているため、飼主検索経由でペットIDを特定し、カルテ作成画面へ直接遷移することで一覧画面の検索機能を経由せず後続シナリオの検証を継続した。
- **推奨対応**: バックエンドで `docker compose logs backend` を確認し、`ListPets` → `pet.List` → `repository.FindAll` の実行時に発生している実際のエラー（Postgres エラーコード等）を特定する。あわせてフロントエンド側で 400/500 系エラーを「0件」として握りつぶさず、エラー専用の表示に分離することを推奨。

### 真因（2026-07-27 確定・解決済み）

**真因は「稼働中 DB のスキーマが repo のコードより古い」ことであり、コードの欠陥ではない。**

稼働中 backend のログに残っていた実エラー（verbatim）:

```
2026/07/27 22:15:38 /app/internal/pet/repository.go:154 ERROR: column pets.version does not exist (SQLSTATE 42703)
{"time":"2026-07-27T22:15:38.26046838+09:00","level":"ERROR",
 "source":{"function":"github.com/animal-ekarte/backend/internal/pet.(*petService).List",
 "file":"/app/internal/pet/service.go","line":249},
 "msg":"failed to list pets",
 "error":"database error: ERROR: column pets.version does not exist (SQLSTATE 42703)"}
```

タイムライン:

| 時刻 | 事象 |
|---|---|
| 18:35:59 | backend コンテナ起動。entrypoint が migration 適用 → `Migration summary applied=1 skipped=0 total=1`（当時は `001_init.sql` のみ存在） |
| 21:49:34 | commit `76a00ec70` が `model.Pet.Version` と `002_add_pets_version.sql`（`ALTER TABLE pets ADD COLUMN version`）を追加 |
| 以降 | air がコードをホットリロードするが **migration は再実行されない**。model は `version` 列を要求し、DB には存在しない |

**なぜ pets 一覧だけが壊れ、他は無事だったのか（切り分け結果の説明）**:
GORM v2 は生 `Joins()` を含むクエリでは `SELECT *` ではなく model の DBNames を `pets.<列名>` として**明示列挙**する。

- `repository.go:143` の `Count` — `SELECT count(*)` なので成功
- `repository.go:154` の `Find` — `buildBase()` が `LEFT JOIN owners` を持つため明示列挙 → `pets.version` を要求 → **42703 で失敗**
- `GET /api/v1/owners` — `internal/owner/repository.go` の `FindAll` は `Joins()` を持たず `SELECT *`。ネストされた `pets` も `Preload`（別クエリの `SELECT *`）なので成功
- `GET /api/v1/pets/{id}` — `Joins()` なしのため成功

**なぜ 400「入力値が正しくありません」に化けたのか**:
`backend/internal/httpapi/response.go` は `isPgError(err)` を一律 `400` にマッピングし、`classifyPgError` の既知コード（23503/23505/22003/22P02/23514）以外はすべて汎用文言「入力値が正しくありません」を返していた。42703 は「サーバ側スキーマ欠陥」であって利用者の入力とは無関係だが、UI 上は利用者のせいに見え、原因特定を大きく遅らせた。

### 対処

1. **エラー分類の是正（本質的な再発防止）** — `backend/internal/httpapi/response_pg.go` / `response.go`
   `classifyPgError` を `(message string, known bool)` に変更。既知5コードのみ 400 を維持し、**未知コードは 500** にする。応答本文には pg メッセージ・制約名・テーブル名・SQL を一切出さない（非漏洩テストで固定）。

   SQLSTATE のサーバ側記録は、request context を持つ domain service 側の既存ログが担う（本障害でも `internal/pet/service.go` の `"failed to list pets"` が `(SQLSTATE 42703)` を含む形で実際に出力されていた。上記ログ引用参照）。`ResolveErrorResponse` は 79 箇所から呼ばれる汎用マッピングであり、ここで再度記録すると未知 pg エラー全般が系統的に二重ログになるため、意図的にログを置いていない（`.claude/rules/go-gin-backend-guidelines.md` §8）。
   **残存リスク**: repo error をログせずに返す service があれば、その経路の未知 pg エラーは 500 のみでコードが残らない。ログ責務を境界1箇所へ統一する設計転換は 79 呼び出し元の監査が必要で本対処のスコープ外。
2. **frontend のエラー握り潰し是正** — API 失敗が「0件」と表示されないよう、`usePetSelectionPage` が `error` を返し、共有コンポーネント `PetSelectionResultsTable` がエラー状態を描画するよう変更。`ownersLoader` は上流の HTTP ステータスを保持する（400 を無条件に 500 へ潰さない）。
3. **model↔DDL ドリフト検知ゲート新設** — `backend/internal/lintscan/model_ddl_column_drift_test.go`。

### ⚠ ゲートの限界（過大評価しないこと）

新設した model↔DDL ドリフトゲートは**本障害を検出できない**。`model.Pet.Version` と `002_add_pets_version.sql` は repo 内では最初から整合しており、不整合だったのは「repo の DDL」と「稼働中 DB」の間だからである。ゲートが塞ぐのは「model に列を足して migration を書き忘れた」という別クラス。

**本障害クラスに対する実効的な防御は上記 1（未知 pg コード → 500 + ログ）**である。同じことが再発しても、今度は「入力値エラー」ではなく 500 とログ上の SQLSTATE として現れ、即座に診断できる。

### USER 手動実行が必要（安全境界）

**状況が 2026-07-27 23:03 時点で変化したため、以下を最新として扱うこと。**

migration 自体は既に適用済みになった（別セッションの操作と思われる）。backend コンテナ起動ログ:

```
⏭ Skipping (already applied) file=001_init.sql
⏭ Skipping (already applied) file=002_add_pets_version.sql
⏭ Skipping (already applied) file=003_add_exam_results_exam_type_field_id_index.sql
Migration summary applied=0 skipped=3 total=3
```

したがって `pets.version` は稼働 DB に存在するようになった。

> **注（履歴・現況ではない）**: 23:03:50〜23:13 の間、backend コンテナは `seeds/003_demo` の checksum mismatch により exit 1 で停止していた（本バグとは別件）。23:13:48 に復旧済み。以下の実測はすべて復旧後に採取したものである。

### 残存リスク

- **[BUG-451](3-session-agent.html#BUG-451)（台帳へ起票済み）**: 400 の解消後、ペット選択画面は「0件」ではなく **「検索結果 20件」と件数を断定表示**する。実測では **描画 20 件に対し API `total` は 15,654 件**（`useGetPets` が `page`/`limit` を送らず backend 既定 `limit=20` で打ち切られるため）。ページネーション UI が無く残り 15,634 件へ到達できない。**素朴な受け入れ確認では「直った」ように見えるため、必ず「画面描画件数」と API `total` を両方採取すること。**
- 開発環境で「コードは新しいが DB は古い」状態は今後も起こり得る。migration を追加した commit を pull した後は `make migrate` を回す運用が必要。
- 上記1の通り、repo error をログせずに返す service の経路では未知 pg エラーの SQLSTATE がサーバ側に残らない。

### 状態（2026-07-27 23:2x 時点）— 解決済み

- **コード側の対処: 完了**（6 コミット。ユニット/回帰テストと scoped lint は green）
- **稼働 DB への migration 適用: 完了**（002/003 とも適用済み。`pets.version` は存在する）
- **旧障害モードの消失を実 HTTP で確認済み**: 認証済みセッションで `GET /api/v1/pets?page=1&limit=20` が **HTTP 200**、`"total": 13984`、`data` 20 件を返す（旧: HTTP 400 `{"error":"入力値が正しくありません"}`）。
- **`/owners` 画面: 一覧を描画**（旧: 「エラー500」画面）。「13,984件中 1-20件」とページ 1〜700 を表示。
- **`/medical-records/select-pet`: 20 行を描画**。エラーでも「該当するペットが見つかりません」でもない。ただし同画面が使う API の `total` は 15,654 であり、この乖離が BUG-451 として残る（本バグとは別欠陥・起票済み）。

**状態: 解決済み。** 残るのは別欠陥 BUG-451（台帳で追跡）。

---
