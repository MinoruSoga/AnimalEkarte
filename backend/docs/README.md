# Animal Ekarte API ドキュメント

## 概要

このディレクトリにはAnimal Ekarte APIのドキュメントが含まれています。

## ファイル構成

```
docs/
├── README.md                 # このファイル
├── api-examples.md          # API使用例と詳細ガイド
├── postman-collection.json  # Postmanコレクション
├── swagger.json             # OpenAPI 2.0仕様
├── swagger.yaml             # OpenAPI 2.0仕様（YAML形式）
└── docs.go                  # Swagger生成コード
```

## ドキュメントの種類

### 1. Swagger UI (推奨)

インタラクティブなAPIドキュメント。ブラウザで直接APIを試すことができます。

**アクセス方法:**
```
http://localhost:8080/swagger/index.html
```

**特徴:**
- 🌐 ブラウザベースのインタラクティブUI
- 🧪 APIを直接試せる（Try it out機能）
- 📝 自動生成されたリクエスト/レスポンス例
- 🔍 スキーマの詳細表示
- 📱 モバイル対応

### 2. API使用例ガイド

詳細な使用例とベストプラクティスを含むガイド。

**ファイル:** `api-examples.md`

**内容:**
- 認証方法
- 全エンドポイントの使用例
- エラーハンドリング
- ページング情報
- 検索機能
- 制限事項

### 3. Postmanコレクション

Postman用のAPIコレクション。インポートしてすぐに使用できます。

**ファイル:** `postman-collection.json`

**特徴:**
- 🚀 インポートして即使用可能
- 🔑 APIキー認証設定済み
- 📝 サンプルリクエスト付き
- 🏷️ カテゴリ分け（System, Pets, Owners, Medical Records）
- 🔄 環境変数対応

### 4. OpenAPI仕様

標準的なOpenAPI 2.0仕様ファイル。

**ファイル:**
- `swagger.json` - JSON形式
- `swagger.yaml` - YAML形式

**用途:**
- API仕様の標準化
- クライアントSDK生成
- ドキュメントツール連携
- APIテスト自動化

## クイックスタート

### 1. Swagger UIで試す

1. ブラウザで `http://localhost:8080/swagger/index.html` を開く
2. 任意のエンドポイントを選択
3. "Try it out" をクリック
4. パラメータを入力して "Execute" を実行

### 2. Postmanで試す

1. Postmanを開く
2. `postman-collection.json` をインポート
3. 環境変数 `YOUR_API_KEY` を設定
4. リクエストを実行

### 3. curlで試す

```bash
# ヘルスチェック
curl -X GET "http://localhost:8080/health"

# ペット一覧取得
curl -X GET "http://localhost:8080/api/v1/pets" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

## APIエンドポイント概要

### System（システム）
- `GET /health` - ヘルスチェック
- `GET /` - ウェルカムメッセージ

### Pets（ペット管理）
- `GET /pets` - ペット一覧取得
- `GET /pets/{id}` - ペット詳細取得
- `POST /pets` - ペット作成
- `PUT /pets/{id}` - ペット更新
- `DELETE /pets/{id}` - ペット削除

### Owners（飼い主管理）
- `GET /owners` - 飼い主一覧取得
- `GET /owners/{id}` - 飼い主詳細取得
- `POST /owners` - 飼い主作成
- `PUT /owners/{id}` - 飼い主更新
- `DELETE /owners/{id}` - 飼い主削除

### Medical Records（電子カルテ）
- `GET /medical-records` - カルテ一覧取得
- `GET /medical-records/paginated` - カルテ一覧（ページング）
- `GET /medical-records/{id}` - カルテ詳細取得
- `POST /medical-records` - カルテ作成
- `PUT /medical-records/{id}` - カルテ更新
- `DELETE /medical-records/{id}` - カルテ削除

## 認証

APIキー認証を使用します。リクエストヘッダーに以下を含めてください：

```
Authorization: Bearer YOUR_API_KEY
```

## エラーレスポンス

| ステータスコード | 説明 | 例 |
|---|---|---|
| 200 | 成功 | 正常なレスポンス |
| 201 | 作成成功 | リソース作成完了 |
| 400 | リクエストエラー | バリデーションエラー |
| 404 | 見つからない | リソースが存在しない |
| 500 | サーバーエラー | 内部エラー |

## データモデル

### Pet（ペット）
```json
{
  "id": "uuid",
  "owner_id": "uuid",
  "pet_number": "string",
  "name": "string",
  "species": "string",
  "breed": "string",
  "gender": "string",
  "birth_date": "date",
  "weight": "number",
  "status": "string"
}
```

### Owner（飼い主）
```json
{
  "id": "uuid",
  "name": "string",
  "name_kana": "string",
  "phone": "string",
  "email": "string",
  "address": "string",
  "notes": "string"
}
```

### Medical Record（カルテ）
```json
{
  "id": "uuid",
  "record_no": "string",
  "pet_id": "uuid",
  "owner_id": "uuid",
  "visit_date": "datetime",
  "visit_type": "string",
  "chief_complaint": "string",
  "subjective": "string",
  "objective": "string",
  "assessment": "string",
  "plan": "string",
  "diagnosis": "string",
  "treatment": "string",
  "status": "string"
}
```

## 検索とフィルタリング

### ペット検索
- 名前、種別、品種で検索可能
- ページング対応

### 飼い主検索
- 名前、メール、電話番号で検索可能
- ページング対応

### カルテ検索
- ペットID、飼い主IDでフィルタリング
- 診察タイプ（初診/再診）でフィルタリング
- ステータス（作成中/確定済）でフィルタリング
- 診察日範囲でフィルタリング
- ページング対応

## 制限事項

- APIキーは必須
- 1ページあたりの最大件数: 100件
- テキストフィールドには文字数制限あり
- UUID形式のIDを使用

## サポート

不明点がある場合は以下を参照してください：

- 📧 Email: support@animal-ekarte.com
- 🌐 サポート: http://localhost:8080/support
- 📋 利用規約: http://localhost:8080/terms

## ライセンス

MIT License - 詳細は [LICENSE](https://opensource.org/licenses/MIT) を参照してください。
