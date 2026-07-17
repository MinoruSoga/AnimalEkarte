# 飼主・ペット一覧 サーバサイド化（理想形）実装計画 — #266

> **正本**: GitHub [#266](https://github.com/MinoruSoga/AnimalEkarte/issues/266)（本書は詳細設計。裁定の変更は #266 コメントが優先）
> **PO 決定（2026-07-17 夜）**: 応急案（owners API + EXISTS）ではなく、**画面の行粒度（ペット行）に一致した読み取り API へ再設計**する。ソートヘッダ 6 列は撤去、species/生死フィルタとペット名検索はサーバサイドで維持。
> **期限**: UAT（#254）前 = 実質 7/18 午前。検証環境 = ローカル（フルデータ投入済み・owners 10,370 / pets 15,654 実測）

## 0. 問題の要約

`/owners` 画面（実体はペット行テーブル）が FE で飼主全件を取得してから展開・絞り込みする設計のため、フルデータ規模で 104 ページ分のリクエストが発火し白画面になる（#266）。根本原因は **UI の行粒度（ペット）と API のリソース粒度（飼主）の不一致** — FE はこの穴をクライアント全件結合で埋めていた。

## 1. UX 決定事項（確定・再議論しない）

| 要素 | 決定 |
|---|---|
| ソートヘッダ 6 列 | **撤去**。デフォルト順固定（下記）。1 万件で全件並べ替え閲覧の業務は無い |
| species フィルタ | **維持・サーバサイド化**（`species` パラメータ） |
| 生死フィルタ | **維持・サーバサイド化**。**デフォルト = 生存のみ**（死亡ペットの誤案内防止は臨床・感情安全の機能。「死亡含む」は明示切替） |
| 検索 | **維持・サーバサイド化**。対象 = ペット名/ペットカナ + 飼主名/飼主カナ/電話。**プレースホルダ文言を実カバレッジと一致させる**（できると表示してできないのが最悪） |
| ページング | サーバサイド（既存 envelope: `data` + `meta.total/page/limit`） |

## 2. API 設計（目標契約）

**新設ではなく既存 `GET /api/v1/pets` の拡張**（一覧基盤は実装済み: `pet_handler.go:14` `ListPets` → `parsePagination` → `newListPetQuery(...).toServiceFilters()` → `petRepository.FindAll(clinicID, ownerID, page, limit, search)`）。

```
GET /api/v1/pets?page=1&limit=50&search=<q>&species=<id|slug>&include_deceased=false
```

- **limit clamp**: サーバ側で最大 100 に強制（`parsePagination` の現状挙動を実測し、上限が無ければ追加。**これが全件取得を構造的に不可能にする本丸**）
- **search**: `pets.name / pets.pet_name_kana` + owners JOIN で `owners.name / owners.name_kana / owners.phone`（ILIKE + 既存の NormalizeKana パターン。medical_record_repository.go:152-170 の実装を踏襲）
- **species**: 型付きパラメータ。`pets.animal_species_id = ?`（FE 選択肢はマスタ由来なので ID で受ける。slug 文字列で WHERE を組ませない）
- **include_deceased**: bool・デフォルト false（`pets.deceased_at IS NULL`。生存判定列は deceased_at — status 列ではない点に注意）
- **順序の安定性**: `ORDER BY owners.name_kana ASC, pets.id ASC`（**一意タイブレーカ必須** — 無いとページ送りで行の重複/欠落が起きる）
- **レスポンス**: 既存 `petListResponse`（`pet_response.go:78-100`）を拡張。`LastVisit` は**実装済み**（:98・model 由来）。**不足は飼主サマリのみ** → `owner` オブジェクト（id / owner_number / name / name_kana / phone / is_dangerous）を埋め込む

### 互換性

- 既存の owners API（`GET /owners`）は**一切変更しない**（飼主検索モーダル・LINE 連携等が使用中）
- `GET /pets` の既存呼び出し元（owner-report の `useGetPets(ownerId)` 等）は ownerID フィルタ経由 — **新パラメータは全て optional のため後方互換**

## 3. BE 実装手順

1. **現状実測（着手時 15 分）**: `newListPetQuery` の既存フィルタ項目／`FindAll` の search 対象カラム／`parsePagination` の limit 上限有無／`petListResponse` に owner 埋め込みが本当に無いか、を grep で確定
2. `pet_repository.go` `FindAll` 拡張:
   - `PetListFilters` 構造体化（ownerID / search / speciesID / includeDeceased）— 引数追加の羅列にしない
   - owners JOIN（search と owner サマリ取得を兼ねる）。**clinic スコープは pets.clinic_id + owners 側も同値で二重に張る**（クロステナント read 監査の既存パターン踏襲）
   - Preload を使う場合は **clinic_id 述語必須**（preload_clinic_scope lint が検出する — P3.1）
3. service / handler: filters の受け渡しと bind 検証（`binding:"omitempty"` + species は uint、include_deceased は bool）
4. `docs/api.yaml` 同期 → `make codegen`（**自動実行禁止 — ユーザーに手動実行を依頼**）
5. **インデックス実測**: ローカルフルデータで `EXPLAIN ANALYZE`。`pets(clinic_id, deceased_at)` / owners JOIN / name_kana ILIKE の実行計画を確認し、必要なら新規 migration でインデックス追加（**採番は実ファイル最大+1**・既適用編集禁止）。目標: 一覧 1 ページ p95 < 500ms

## 4. FE 実装手順

対象: `frontend/src/features/owners/`（routes/OwnersList.tsx + components/OwnersListTable.tsx）。**feature 名・URL `/owners` は変えない**（リネームは diff 爆発・別コミットでも今はやらない）。

1. 全件取得ロジック（page=1..N のクロール）を**削除**し、`GET /pets` のページ単位フェッチに置換（1 画面 = 1 リクエスト）
2. 検索ボックス・species/生死フィルタを**クエリパラメータへ配線**（UI 新設なし。既存 PropertyFilter/Select を流用）。入力はデバウンス（既存パターンがあれば踏襲）
3. ソートヘッダ 6 列を撤去（`DataTable` の sortable 指定を外すだけに留め、コンポーネント破壊はしない）
4. ページネーション UI: 既存 `usePagination` デフォルトに委譲
5. 検索プレースホルダを「飼主名・カナ・電話・ペット名」に修正
6. 危険度バッジ（`OwnersListTable.tsx:304-308`）は owner サマリ + pet.danger_level で従来表示を維持（#229/#234 の系との整合を壊さない）

## 5. テスト

- **BE**: FindAll — search（ペット名/飼主名/カナ正規化）/ species / include_deceased デフォルト false / limit clamp / **順序安定性（同 kana で id タイブレーク）** / **クロステナント隔離**（他 clinic の pet/owner が混入しない）
- **FE**: 一覧描画（owner サマリ列）/ フィルタ変更でクエリパラメータが変わる / **全件クロールが存在しない**（fetch モックの呼び出し回数 = 1）
- 検証コマンド（scoped）:
  ```
  docker compose exec backend go test ./internal/repository/ -run TestPetRepository
  docker compose exec backend go test ./internal/handler/ -run TestPet
  docker compose exec frontend npx vitest run src/features/owners
  ```

## 6. 受け入れ検証（ローカル・フルデータ）

1. `http://localhost:5173/owners`（またはローカル FE ポート）を開き **3 秒以内に初期表示**
2. ネットワークタブで `GET /pets?page=1&limit=...` が **1 本だけ**発火していること（クロール消滅の実証）
3. 検索「マメラ」（ペット名）と飼主名・電話での引き当て、species 絞り、死亡含む切替、ページ送り
4. `EXPLAIN ANALYZE` の結果を #266 にコメント（p95 目標との突合）
5. 完了後 STG 実 URL（api.stg.noah-karte.com 経由・同一フルデータ）で 1〜3 を再確認

## 7. スコープ外・リスク

- **スコープ外**: 他画面の全件取得パターンの棚卸し（#266 受け入れ条件の別項目 — fetch-limits 使用 6 箇所のレビュー。本画面完了後に同セッションで実施）／owners API 側の拡張／feature リネーム
- **リスク 1**: owners JOIN + ILIKE がフルデータで遅い → インデックス追加で対処（手順 3-5）。それでも遅い場合のみ pg_trgm GIN を検討（計測後）
- **リスク 2**: `petListResponse` の変更が既存呼び出し元（owner-report 等）の型に波及 → optional 追加のみで対応し、既存フィールドは触らない
- **リスク 3**: 時間切れ → その場合のフォールバックは「search + ページングのみで出荷し、フィルタ 2 種を UAT 前の別コミットに分割」。**limit clamp と 1 リクエスト化だけは絶対に落とさない**（それが #266 の本体）

## 8. 実践ゲート（PRODUCT_PHILOSOPHY）

- **① 責任者**: PO 決定 2026-07-17（本書ヘッダ）。業務目的 = 受付が来院飼主を実データ規模で 3 秒以内に引き当てる
- **② 削除される工程**: FE の全件クロール（104 リクエスト）・クライアント側ソート/フィルタ実装・ソートヘッダ 6 列
- **③ 簡素化**: 全件閲覧 → 検索主動線 + 賢いデフォルト（生存のみ・kana 順）
- **④ メトリクス**: /owners 初期表示 白画面（∞）→ 3 秒以内。1 画面のリクエスト数 104+ → 1
