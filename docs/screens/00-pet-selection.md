# ペット選択 仕様書

## 概要
- **画面の目的**: カルテ・入院・トリミング・会計等の各機能で、新規作成時にペットを選択するための共通中間画面
- **URLパターン**: `/:feature/select-pet`
  - `/medical-records/select-pet`
  - `/hospitalization/select-pet`
  - `/trimming/select-pet`
  - `/accounting/select-pet`
  - `/examinations/select-pet`
  - `/vaccinations/select-pet`
- **アクセス権限**: 認証済ユーザー全員

## 画面レイアウト
```
┌────────────────────────────────────────────────────────────┐
│ ← 戻る  [ページタイトル] ペット選択                        │
├────────────────────────────────────────────────────────────┤
│ ■ 検索条件（グリッド: 2列/4列）                            │
│ 飼主No | 飼主名 | 飼主名(カナ) | 電話番号                  │
│ ペット名 | ペット名(カナ) | 種別 | 住所                    │
│                          [クリア] [検索]                   │
├────────────────────────────────────────────────────────────┤
│ 検索結果テーブル                                           │
│ 飼主No | 飼主名 | ペット番号 | ペット名 | 種 | 生死 |     │
│ 生年月日 | 担当医 | 操作ボタン                             │
└────────────────────────────────────────────────────────────┘
```

## 表示項目（検索フォーム）

| フィールド名 | プレースホルダー例 | 備考 |
|------------|-----------------|------|
| 飼主No | 例: 30042 | `owners.id` |
| 飼主名 | 例: 林 文明 | `owners.owner_name` |
| 飼主名(カナ) | 例: ハヤシ フミアキ | `owners.owner_name_kana` |
| 電話番号 | 例: 090-1234-5678 | `owners.phone` |
| ペット名 | 例: Iris | `pets.name` |
| ペット名(カナ) | 例: イリス | `pets.name_kana` |
| 種別 | 例: 犬 | `pets.animal_species_id` → `animal_species` |
| 住所 | 例: 東京都 | `owners.address1` / `owners.home_address1` |

## 表示項目（検索結果テーブル）

| フィールド名 | 型 | 説明 | 備考 |
|------------|-----|------|------|
| 飼主No | string | 飼主ID番号 | `owners.id` |
| 飼主名 | string | 飼い主氏名 | `owners.owner_name` |
| ペット番号 | string | ペットの患者番号 | `pets.pet_number` |
| ペット名 | string | ペットの名前 | `pets.name` |
| 種 | string | 犬/猫等 | `pets.animal_species_id` → `animal_species` |
| 生死 | enum | 生存/死亡ステータス | `pets.status` |
| 生年月日 | date | ペットの誕生日 | `pets.birth_date` |
| 担当医 | string | 担当獣医師名 | `staffs.name` |

## UI コンポーネント
- **`[S] PetSearchForm`**: 検索条件フォーム（8フィールドのグリッドレイアウト。条件入力後に「検索」ボタンで非同期検索。条件クリアボタン付き）
- **`[S] PetSearchResultsTable`**: 検索結果テーブル（行クリックまたは「選択」ボタンでペット選択）
- **`[H] usePetSearch`**: 検索状態管理フック（`searchParams`, `results`, `isSearching`, `hasSearched`）
- **`[S] PageLayout`**: ページコンテナ（戻るボタン付き）

## ユーザーアクション

| アクション | トリガー | 処理内容 | 遷移先 |
|-----------|---------|---------|--------|
| 検索 | 「検索」ボタンまたはEnterキー | 条件に一致するペット一覧を取得 | 同画面（結果表示） |
| クリア | 「クリア」ボタン | 検索条件をリセット | 同画面 |
| ペット選択 | 行クリックまたは「選択」ボタン | 選択ペットIDをクエリパラメータに付与 | 遷移元の新規作成フォーム (`?petId=xxx`) |
| 戻る | 戻るボタン | 遷移元画面に戻る | `/medical-records` 等 |

## 画面遷移

| 遷移元 | 遷移先 | 条件 |
|--------|--------|------|
| カルテ一覧 `/medical-records` | カルテフォーム `/medical-records/new?petId=xxx` | ペット選択後 |
| 入院一覧 `/hospitalization` | 入院フォーム `/hospitalization/new?petId=xxx` | ペット選択後 |
| トリミング一覧 `/trimming` | トリミングフォーム `/trimming/new?petId=xxx` | ペット選択後 |
| 会計一覧 `/accounting` | 会計精算 `/accounting/new?petId=xxx` | ペット選択後 |

## バリデーション
- 検索ボタンは条件が1つ以上入力された場合のみ有効（`hasConditions` チェック）
- 検索結果が0件の場合は「該当するペットが見つかりません」を表示
- 検索実行中は「クリア」「検索」ボタンを disabled

## 状態管理

| 状態 | 型 | 説明 |
|------|-----|------|
| `searchParams` | `PetSearchParams` | 8フィールドの検索条件オブジェクト |
| `results` | `Pet[]` | 検索結果一覧 |
| `isSearching` | `boolean` | 検索実行中フラグ |
| `hasSearched` | `boolean` | 検索済みフラグ（初期状態と結果0件の表示切替に使用） |

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/pets` | ペット一覧取得（検索パラメータ付き） | 実装済 |

## 実装状況
- フロントエンド(ui-sample): 実装済（`[R] MedicalRecordPetSelection`、`[S] PetSearchForm` + `[S] PetSearchResultsTable` の共通コンポーネント使用）
- バックエンドAPI: 実装済（`/pets` エンドポイント）

## 備考
- 旧仕様のリアルタイム検索（テキスト入力毎にフィルタリング）から、**検索ボタン式の非同期検索**に変更されている
- 検索フォームは `PetSearchForm` 共通コンポーネントとして全featureで再利用される
- `location.state` に `from`（戻り先）と `activeTab`（カルテフォームの初期タブ）を渡す
