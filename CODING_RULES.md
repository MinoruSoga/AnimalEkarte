# Coding Rules - Animal Ekarte

## 概要

本ドキュメントはAnimal Ekarteプロジェクトのコーディング規約を定義する。
全開発者はこのルールに従うこと。

詳細なルールは各サブディレクトリを参照:
- [Frontend Coding Rules](./frontend/CODING_RULES.md)
- [Backend Coding Rules](./backend/CODING_RULES.md)

---

## 1. プロジェクト構成

### 1.1 全体構造

```
AnimalEkarte/
├── backend/                    # Go API サーバー
│   ├── cmd/                    # エントリーポイント
│   ├── internal/               # 内部パッケージ
│   ├── migrations/             # DBマイグレーション
│   └── docs/                   # Swagger
├── frontend/                   # React 19 SPA
│   └── src/
│       ├── main.tsx            # Viteエントリーポイント
│       ├── index.css           # グローバルCSS
│       ├── app/                # アプリケーション層
│       ├── components/         # 共有コンポーネント
│       ├── features/           # 機能別モジュール
│       ├── hooks/              # 共有hooks
│       ├── lib/                # ユーティリティ
│       ├── stores/             # グローバル状態管理
│       ├── types/              # 共有型定義
│       └── testing/            # テスト設定（MSW含む）
├── docker-compose.yml
├── Makefile
└── .env
```

### 1.2 技術スタック

| 層 | 技術 |
|---|---|
| Frontend | React 19, TypeScript 5.7, Vite 6, Tailwind CSS 4, shadcn/ui |
| Backend | Go 1.22+, Gin, GORM |
| Database | PostgreSQL 18 |
| Infrastructure | Docker Compose |

---

## 2. 開発環境

### 2.1 Docker必須

**全てのコマンドはDocker経由で実行する。ローカル実行は禁止。**

```bash
# ❌ 禁止
npm run build
go test ./...

# ✅ 必須
docker compose exec frontend npm run build
docker compose exec backend go test ./...
```

### 2.2 基本コマンド

| タスク | コマンド |
|--------|---------|
| 起動 | `make up` |
| 停止 | `make down` |
| ログ | `make logs` |
| DB接続 | `make db` |
| Frontend ビルド | `docker compose exec frontend npm run build` |
| Frontend Lint | `docker compose exec frontend npm run lint` |
| Frontend テスト | `docker compose exec frontend npm run test:run` |
| Backend テスト | `docker compose exec backend go test ./... -v` |
| Backend Lint | `docker compose exec backend golangci-lint run ./...` |

---

## 3. 共通ルール

### 3.1 命名規則

#### ファイル名

| 言語 | 規則 | 例 |
|------|------|-----|
| TypeScript | kebab-case | `patient-card.tsx`, `use-patient.ts` |
| Go | snake_case | `patient_handler.go`, `pet_repository.go` |
| SQL | snake_case + 番号 | `001_create_owners.sql` |

#### 変数・関数名

| 言語 | 対象 | 規則 | 例 |
|------|------|------|-----|
| TypeScript | 変数・関数 | camelCase | `getPatientById` |
| TypeScript | コンポーネント | PascalCase | `PatientCard` |
| TypeScript | 定数 | UPPER_SNAKE_CASE | `API_BASE_URL` |
| TypeScript | 型・Interface | PascalCase | `Patient`, `ApiResponse` |
| Go | エクスポート | PascalCase | `GetPatient` |
| Go | 非エクスポート | camelCase | `validateInput` |
| Go | パッケージ | lowercase | `handler` |

### 3.2 コメント規則

```typescript
// ❌ 禁止: 何をしているかの説明（コードを読めばわかる）
// ループで患者を取得する
for (const patient of patients) { ... }

// ✅ 推奨: なぜそうしているかの説明
// 患者データは時系列順に表示する要件のため、日付でソート
const sorted = patients.sort((a, b) => a.createdAt - b.createdAt);
```

```go
// ❌ 禁止
// GetPatient gets a patient
func GetPatient(id string) {}

// ✅ 推奨: 特殊なケースや副作用を説明
// GetPatient retrieves a patient by ID.
// Returns ErrNotFound if the patient does not exist.
func GetPatient(ctx context.Context, id string) (*Patient, error) {}
```

### 3.3 エラーハンドリング

**原則: エラーは握りつぶさない。必ず適切に処理する。**

```typescript
// ❌ 禁止
try {
  await fetchData();
} catch (e) {
  // 何もしない
}

// ✅ 必須
try {
  await fetchData();
} catch (error) {
  if (error instanceof ApiError) {
    toast.error(error.message);
  } else {
    toast.error("予期せぬエラーが発生しました");
    console.error(error); // 開発時のみ
  }
}
```

```go
// ❌ 禁止
result, _ := doSomething()

// ✅ 必須
result, err := doSomething()
if err != nil {
    return nil, errors.Wrap(err, "failed to do something")
}
```

---

## 4. Git / バージョン管理

### 4.1 コミットメッセージ

```
<type>(<scope>): <subject>

<body>

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
```

#### Type一覧

| Type | 説明 | 例 |
|------|------|-----|
| `feat` | 新機能 | `feat(owners): 飼主検索機能を追加` |
| `fix` | バグ修正 | `fix(dashboard): 予約表示のタイムゾーン修正` |
| `refactor` | リファクタリング | `refactor(api): エラーハンドリングを統一` |
| `docs` | ドキュメント | `docs: API仕様書を更新` |
| `test` | テスト | `test(owners): OwnerForm のテストを追加` |
| `chore` | その他 | `chore: 依存関係を更新` |
| `perf` | パフォーマンス | `perf(list): 仮想スクロールを導入` |
| `style` | フォーマット | `style: Prettier適用` |

### 4.2 ブランチ戦略

```
main                    # 本番環境
├── develop             # 開発環境
│   ├── feature/xxx     # 機能開発
│   ├── fix/xxx         # バグ修正
│   └── refactor/xxx    # リファクタリング
```

### 4.3 禁止操作

| 操作 | 理由 |
|------|------|
| `git push --force` (main/develop) | 履歴破壊 |
| `git reset --hard` (確認なし) | 作業消失 |
| `--no-verify` | フック回避 |
| 機密情報のコミット | セキュリティ |

---

## 5. セキュリティ

### 5.1 必須対策

| 対策 | 実装方法 |
|------|----------|
| SQLインジェクション | プレースホルダ使用（GORM標準） |
| XSS | React の自動エスケープ + DOMPurify |
| CSRF | SameSite Cookie + CSRFトークン |
| 認証 | JWT + HTTPOnly Cookie |
| 機密情報管理 | 環境変数（`.env`） |

### 5.2 禁止事項

```typescript
// ❌ 禁止: 機密情報のハードコード
const API_KEY = "sk-1234567890";

// ✅ 必須: 環境変数から取得
const API_KEY = import.meta.env.VITE_API_KEY;
```

```go
// ❌ 禁止: 文字列連結によるSQL
db.Raw("SELECT * FROM users WHERE id = " + id)

// ✅ 必須: プレースホルダ
db.Where("id = ?", id).First(&user)
```

### 5.3 ログ出力

```go
// ❌ 禁止: 機密情報のログ出力
slog.Info("user login", slog.String("password", password))

// ✅ 必須: 機密情報をマスク
slog.Info("user login", slog.String("email", email))
```

---

## 6. コードレビュー

### 6.1 レビュー観点

#### 共通
- [ ] 命名は適切か（意図が伝わるか）
- [ ] エラーハンドリングは適切か
- [ ] セキュリティリスクはないか
- [ ] パフォーマンス問題はないか
- [ ] テストは十分か

#### Frontend固有
- [ ] `any` 型を使用していないか
- [ ] `FC` / `forwardRef` を使用していないか
- [ ] feature間importがないか
- [ ] 不要な再レンダリングがないか
- [ ] アクセシビリティは考慮されているか

#### Backend固有
- [ ] Context伝播しているか
- [ ] エラーをラップしているか
- [ ] slogで構造化ログ出力しているか
- [ ] SQLインジェクション対策されているか
- [ ] トランザクション管理は適切か

### 6.2 レビュー手順

1. 静的解析（lint）がパスしていることを確認
2. テストがパスしていることを確認
3. 変更の意図を理解（PR説明を読む）
4. コードを読み、上記観点でチェック
5. 必要に応じて動作確認
6. 指摘事項があればコメント
7. 問題なければApprove

---

## 7. テスト

### 7.1 テスト方針

| テスト種別 | 対象 | カバレッジ目標 |
|----------|------|---------------|
| 単体テスト | ユーティリティ関数、hooks | 80%+ |
| 統合テスト | API、Repository | 70%+ |
| E2Eテスト | 主要ユーザーフロー | 主要パス |

### 7.2 テストファイル配置

```
# Frontend: 同階層に配置
src/lib/utils.ts
src/lib/utils.test.ts

# Backend: 同パッケージ内
internal/service/pet_service.go
internal/service/pet_service_test.go
```

### 7.3 テスト命名

```typescript
describe("calculateAge", () => {
  it("should return 0 for today's date", () => {});
  it("should return correct age for past date", () => {});
  it("should throw error for future date", () => {});
});
```

```go
func TestPetService_GetByID(t *testing.T) {
    t.Run("returns pet when found", func(t *testing.T) {})
    t.Run("returns error when not found", func(t *testing.T) {})
}
```

---

## 8. ドキュメント

### 8.1 必須ドキュメント

| ドキュメント | 場所 | 内容 |
|-------------|------|------|
| API仕様 | `/backend/docs/swagger.yaml` | OpenAPI 3.0 |
| DB設計 | `/docs/ERD.md` | ER図、テーブル定義 |
| 環境構築 | `/README.md` | セットアップ手順 |
| コーディング規約 | `/CODING_RULES.md` | 本ドキュメント |

### 8.2 コード内ドキュメント

```typescript
/**
 * 患者の年齢を計算する
 *
 * @param birthDate - 生年月日
 * @returns 年齢（整数）
 * @throws {Error} 未来の日付が指定された場合
 *
 * @example
 * const age = calculateAge(new Date("2020-01-15"));
 * // => 4 (2024年の場合)
 */
export function calculateAge(birthDate: Date): number {
  // ...
}
```

---

## 9. パフォーマンス

### 9.1 Frontend

| 対策 | 実装 |
|------|------|
| コード分割 | React.lazy + Suspense |
| メモ化 | useMemo, useCallback（必要な場合のみ） |
| 仮想スクロール | 大量リスト表示時 |
| 画像最適化 | WebP + lazy loading |

### 9.2 Backend

| 対策 | 実装 |
|------|------|
| N+1問題回避 | Preload, Joins |
| インデックス | 検索条件カラムに付与 |
| ページネーション | Limit + Offset |
| キャッシュ | Redis（必要に応じて） |

---

## 10. 参照リンク

- [Frontend Coding Rules](./frontend/CODING_RULES.md)
- [Backend Coding Rules](./backend/CODING_RULES.md)
- [ERD（データベース設計）](./docs/ERD.md)
- [API仕様（Swagger）](http://localhost:8080/swagger/index.html)
- [Bulletproof React](https://github.com/alan2207/bulletproof-react)
- [React 19 Release Notes](https://react.dev/blog/2024/12/05/react-19)
- [Effective Go](https://go.dev/doc/effective_go)
