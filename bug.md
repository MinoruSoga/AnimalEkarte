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

### 派生欠陥 BUG-451 / BUG-452（2026-07-28 いずれも解決済み・台帳からクローズ）

400 の解消により、それまで「0件」に隠れていた別クラスの欠陥が可視化された。**画面が「検索結果 20件」と件数を断定表示するが、実測では描画 20 件に対し API `total` は 15,654 件**だった（`useGetPets` が `page`/`limit` を送らず backend 既定 `limit=20` で打ち切られ、その 20 件に対してのみクライアント側フィルタが掛かる）。実在患者を検索して見つからず「未登録」と誤認し重複カルテを作りうる、[BUG-449](3-session-agent.html#BUG-449) と同じ「システムが黙って不完全な結果を正常な結果として提示する」クラスである。

| | 対象 | 機序 | 解決コミット |
|---|---|---|---|
| **BUG-451** | ペット選択系 7 画面（`usePetSelectionPage` 共有） | `useGetPets` が `page`/`limit` 未送出 → 既定 20 件 → `normalizedIncludes` でページ内フィルタ | `0ff9175ae` → `71500a218`（独立レビュー是正）→ `cbed6f1d1`（クローズ。同 commit で BUG-452 を起票） |
| **BUG-452** | 予約フォームの患者選択モーダル（`PatientSelectionTable`） | 同上 ＋ **`MAX_RESULTS = 20` の二重打ち切り** | `e20a1aeec` → `866c4e8e8` / `844808da8`（レビュー是正）→ `6b7a4bbec`（クローズ） |

**BUG-451 は起票時に影響範囲を「7 画面」と記録していたが、実際は 8 面だった。** 8 面目（`PatientSelectionTable`）は共有フックを使っていないため BUG-451 の修正では直らず、BUG-452 として別途起票・修正した。

いずれも修正方針は同じ: backend の `search` / `species` / `owner_id` 述語とページネーションへ委譲し、クライアント側フィルタを削除する。**backend は無変更**。副産物として `useSearchPets`（`page`/`limit` 未指定の呼び出し口）は利用者ゼロとなり削除した — 残せば 9 面目を誘発する構造だったため。

**住所検索は機能ごと廃止した。** backend の `search` 述語（`backend/internal/pet/repository.go:123-139`）は `owners.address1` / `address2` を対象に持たない。押せない欄を `disabled` で残すと「その条件で検索した」という誤解が残るため、`disabled` 化ではなく削除を選んだ。

### 残存リスク

- **実挙動確認が未消化（BUG-451 / BUG-452 共通）**: ユニット回帰テストでは担保済みだが、ブラウザでの実操作確認は Chrome 拡張未接続・CDP 9222 不通・認証情報が `.secret`（読取禁止）のため未実施。ログイン済みブラウザセッションがあれば ①開いた直後は空 ②総件数とページャ ③ページ跨ぎの選択保持 を確認できる。
- **`include_deceased` が面ごとに不一致**: 予約フォームの患者選択は生存のみ、`usePetSelectionPage` 系 7 面は死亡ペットを含む。予約登録に死亡ペットを出す必要がないなら現状が正しいが、面ごとの差は将来の混乱要因。方針の明示を推奨。
- 開発環境で「コードは新しいが DB は古い」状態は今後も起こり得る。migration を追加した commit を pull した後は `make migrate` を回す運用が必要。
- ~~上記1の通り、repo error をログせずに返す service の経路では未知 pg エラーの SQLSTATE がサーバ側に残らない。~~ → **2026-07-28 解消（`1eb15c6ea`）**。`RespondError` / `RespondErrorWithExtras` が解決 status 500 以上のときだけ `c.Error(err)` を呼ぶようにし、配線済みの `RequestLoggingMiddleware`（`cmd/api/main.go:202`）が SQLSTATE を request 境界で記録するようにした。起票時に「79 呼び出し元の監査が必要」と書いたのは誤りで、実際には `c.Error(` の呼び出しがコードベース全体でゼロであり middleware の `c.Errors` 分岐が本番経路では到達しない dead branch だったため、追加は 1 箇所で足りた。4xx は登録しない（既知 pg コード、特に 23505 の detail が利用者入力を含みうるため）。`go-gin-backend-guidelines.md` §8 の重複ログ禁止には、未知 pg コードに限り境界記録を必須とし domain service 側の重複を許容する例外を追記した（同 commit）。domain service 側のログは 1 件も削除していない。

### 状態（2026-07-28 更新）— 解決済み

- **コード側の対処: 完了**（7 コミット。ユニット/回帰テストと scoped lint は green。最終は `1eb15c6ea` = SQLSTATE の request 境界記録）
- **稼働 DB への migration 適用: 完了**（002/003 とも適用済み。`pets.version` は存在する）
- **旧障害モードの消失を実 HTTP で確認済み**: 認証済みセッションで `GET /api/v1/pets?page=1&limit=20` が **HTTP 200**、`"total": 13984`、`data` 20 件を返す（旧: HTTP 400 `{"error":"入力値が正しくありません"}`）。
- **`/owners` 画面: 一覧を描画**（旧: 「エラー500」画面）。「13,984件中 1-20件」とページ 1〜700 を表示。
- **`/medical-records/select-pet`: 20 行を描画**。エラーでも「該当するペットが見つかりません」でもない。ただし同画面が使う API の `total` は 15,654 であり、この乖離を BUG-451 として起票した（本バグとは別欠陥）。

**状態: 解決済み。** 派生した BUG-451 / BUG-452 も 2026-07-28 に解決・台帳からクローズ済み（上記表を参照）。SQLSTATE がサーバ側に残らない残存リスクも 2026-07-28 に `1eb15c6ea` で解消した。

残る未消化は次の 3 件で、いずれも本バグの再発防止としては閉じている:

| 残件 | 性質 | 次に動くのは誰か |
|---|---|---|
| 実挙動確認（BUG-451 / BUG-452） | ログイン済みブラウザセッションが要る | USER |
| `include_deceased` の面ごと不一致 | 業務判断。要件責任者が必要 | USER |
| pull 後の `make migrate` 運用 | 開発環境の運用ルール | 全員 |

---
