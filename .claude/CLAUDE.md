# Animal Ekarte - 動物病院電子カルテシステム

## 🎯 コーディング姿勢

**シニアエンジニアとして以下を徹底：**
- 型安全性最優先
- SOLID原則・クリーンアーキテクチャ
- エラーハンドリング徹底
- セキュリティ意識
- パフォーマンス考慮
- 自己レビュー実施

---

## 📋 プロジェクト概要

**名前:** Animal Ekarte
**説明:** 動物病院向け電子カルテ管理システム

---

## 🛠️ 技術スタック

### Backend
- 言語: Go
- フレームワーク: Gin
- ホットリロード: Air

### Frontend
- 言語: TypeScript 5.7
- フレームワーク: React 18
- ビルドツール: Vite 6
- スタイル: Tailwind CSS 4
- UIライブラリ: shadcn/ui (Radix UI)
- ルーティング: React Router 6
- アイコン: lucide-react

### Infrastructure
- データベース: PostgreSQL 18
- コンテナ: Docker Compose
- マイグレーション: SQL files

---

## 📁 ディレクトリ構造

```
AnimalEkarte/
├── backend/
│   ├── cmd/              # エントリーポイント
│   ├── internal/         # 内部パッケージ
│   │   ├── config/       # 設定
│   │   ├── errors/       # エラー定義
│   │   ├── handler/      # HTTPハンドラ
│   │   ├── logger/       # ロガー
│   │   ├── model/        # ドメインモデル
│   │   ├── repository/   # データアクセス
│   │   └── service/      # ビジネスロジック
│   ├── migrations/       # DBマイグレーション
│   ├── docs/             # APIドキュメント (Swagger)
│   ├── .golangci.yml     # Linter設定
│   ├── go.mod
│   └── Dockerfile.dev
├── frontend/
│   ├── src/
│   │   ├── components/   # 共通コンポーネント
│   │   │   ├── ui/       # shadcn/ui コンポーネント
│   │   │   ├── shared/   # 共有UIコンポーネント
│   │   │   ├── figma/    # Figma生成コンポーネント
│   │   │   └── Sidebar.tsx
│   │   ├── features/     # 機能別モジュール
│   │   │   ├── dashboard/      # ダッシュボード
│   │   │   ├── owners/         # 飼い主管理
│   │   │   ├── medical-records/# カルテ管理
│   │   │   ├── reservations/   # 予約管理
│   │   │   ├── hospitalization/# 入院管理
│   │   │   ├── examinations/   # 検査管理
│   │   │   ├── accounting/     # 会計
│   │   │   ├── vaccinations/   # ワクチン
│   │   │   ├── trimming/       # トリミング
│   │   │   ├── master/         # マスタ設定
│   │   │   └── clinic/         # クリニック設定
│   │   ├── lib/          # ユーティリティ
│   │   ├── types/        # 型定義
│   │   ├── styles/       # グローバルスタイル
│   │   ├── assets/       # 画像等のアセット
│   │   ├── App.tsx       # ルーティング定義
│   │   └── main.tsx      # エントリーポイント
│   ├── package.json
│   └── Dockerfile
├── docker-compose.yml
├── Makefile
└── .env
```

---

## 🚀 開発コマンド

| コマンド | 説明 |
|---------|------|
| `make up` | コンテナ起動 |
| `make build` | コンテナ起動（ビルド付き） |
| `make down` | コンテナ停止 |
| `make logs` | 全ログ表示 |
| `make logs-api` | APIログ表示 |
| `make logs-front` | フロントエンドログ表示 |
| `make ps` | コンテナ状態確認 |
| `make db` | DB接続（psql） |
| `make clean` | キャッシュクリア＆再ビルド |
| `make reset` | 完全リセット（データ削除） |
| `make restart-api` | API再起動 |
| `make restart-front` | フロントエンド再起動 |

---

## 📝 命名規則

### Go (Backend)

| 対象 | 規則 | 例 |
|------|------|-----|
| パッケージ | lowercase | `handler`, `repository` |
| エクスポート関数/型 | PascalCase | `GetPatient`, `PatientService` |
| プライベート関数/変数 | camelCase | `validateInput`, `dbConn` |
| 定数 | PascalCase or UPPER_SNAKE | `MaxRetryCount`, `DB_TIMEOUT` |
| インターフェース | PascalCase + er | `Reader`, `PatientRepository` |
| ファイル | snake_case | `patient_handler.go` |

### TypeScript (Frontend)

| 対象 | 規則 | 例 |
|------|------|-----|
| コンポーネント | PascalCase | `PatientCard`, `MedicalRecord` |
| 関数・変数 | camelCase | `getPatientById`, `isActive` |
| 定数 | UPPER_SNAKE_CASE | `API_BASE_URL` |
| ファイル | kebab-case | `patient-card.tsx` |
| 型・インターフェース | PascalCase | `Patient`, `ApiResponse` |

---

## 🔐 環境変数

```bash
# .env
DB_USER=ekarte_user
DB_PASSWORD=<secure_password>
DB_NAME=ekarte_db
```

**注意:** `.env` ファイルは `.gitignore` に含まれています。

---

## 📐 重要なパターン

### Go - エラーハンドリング（センチネルエラー + ラッピング）
```go
// internal/errors/errors.go でセンチネルエラーを定義
var (
    ErrNotFound     = errors.New("resource not found")
    ErrInvalidInput = errors.New("invalid input")
)

// エラーラッピングでコンテキストを追加
func Wrap(err error, message string) error {
    return fmt.Errorf("%s: %w", message, err)
}

// エラー判定は errors.Is() を使用
if errors.Is(err, ErrNotFound) {
    // 404 レスポンス
}
```

### Go - slog 構造化ログ
```go
import "log/slog"

// コンテキスト付きログ出力
slog.InfoContext(ctx, "pet created", slog.String("pet_id", pet.ID.String()))
slog.ErrorContext(ctx, "failed to create pet", slog.String("error", err.Error()))
```

### Go - Context伝播
```go
// 全てのレイヤーで context.Context を第一引数に
func (s *Service) GetPetByID(ctx context.Context, id string) (*Pet, error)
func (r *Repository) GetPetByID(ctx context.Context, id uuid.UUID) (*Pet, error)

// GORMでもContextを使用
r.db.WithContext(ctx).First(&pet, "id = ?", id)
```

### Go - リポジトリパターン
```go
type PatientRepository interface {
    FindByID(ctx context.Context, id string) (*Patient, error)
    Save(ctx context.Context, patient *Patient) error
    Delete(ctx context.Context, id string) error
}
```

### React - コンポーネント構造
```typescript
interface PatientCardProps {
  patient: Patient;
  onSelect?: (id: string) => void;
}

export const PatientCard: FC<PatientCardProps> = ({ patient, onSelect }) => {
  // 実装
};
```

### React - API呼び出し
```typescript
const fetchPatient = async (id: string): Promise<Patient> => {
  const response = await fetch(`/api/patients/${id}`);
  if (!response.ok) {
    throw new Error('Failed to fetch patient');
  }
  return response.json();
};
```

---

## ⚡ 禁止事項

### Go
- ❌ panic の乱用
- ❌ エラーの握りつぶし（`_ = err`）
- ❌ グローバル変数の多用
- ❌ 未使用のインポート

### TypeScript
- ❌ any型使用
- ❌ 未使用インポート
- ❌ ハードコード
- ❌ コンソールログ放置

---

## ✅ 必須事項

- ✅ 適切なエラーハンドリング（センチネルエラー + ラッピング）
- ✅ Context伝播（全関数の第一引数）
- ✅ slog構造化ログ使用
- ✅ 入力値バリデーション
- ✅ 既存パターンに従う
- ✅ 変更前に影響範囲確認
- ✅ SQLインジェクション対策（プレースホルダ使用）
- ✅ golangci-lint でコード品質チェック

---

## 🤖 エージェント構成

| エージェント | モデル | 用途 |
|-------------|--------|------|
| `architect` | Opus | アーキテクチャ設計、セキュリティ監査 |
| `implementer` | Sonnet | 機能実装、テスト作成 |
| `reviewer` | Sonnet | コードレビュー、品質チェック |
| `debugger` | Sonnet | バグ調査、エラー分析 |
| `researcher` | Haiku | コード検索、ファイル探索 |
| `formatter` | Haiku | コミット生成、コード整形 |

**使用方法:**
- 自動委譲: 適切なタスクで自動的に呼び出される
- 明示的: `Use the [agent] agent to...`

---

## 🔧 Docker操作

### ⚠️ 重要: コマンド実行ルール

**npmやgoコマンドはローカルで実行しないこと。必ずDocker経由で実行する。**

```bash
# ❌ NG - ローカル実行
npm run build
go test ./...

# ✅ OK - Docker経由
docker compose exec frontend npm run build
docker compose exec backend go test ./...
```

### コンテナ別コマンド

| タスク | コマンド |
|--------|---------|
| Frontend ビルド | `docker compose exec frontend npm run build` |
| Frontend Lint | `docker compose exec frontend npm run lint` |
| Frontend テスト | `docker compose exec frontend npm run test:run` |
| Backend テスト | `docker compose exec backend go test ./... -v` |
| Backend Lint | `docker compose exec backend golangci-lint run ./...` |
| Backend モジュール更新 | `docker compose exec backend go mod tidy` |

### コンテナ構成
- `ekarte-db`: PostgreSQL 18
- `ekarte-backend`: Go API (port 8080)
- `ekarte-frontend`: React (port 3000)

### ポート
| サービス | ポート |
|---------|--------|
| Frontend | 3000 |
| Backend API | 8080 |
| PostgreSQL | 5432 |

---

## 📊 データベース

### 接続情報
- Host: `localhost` (外部) / `db` (Docker内部)
- Port: `5432`
- Database: `ekarte_db`
- User: `ekarte_user`

### マイグレーション
SQLファイルは `backend/migrations/` に配置。
Docker起動時に自動実行される。

---

## 📚 参照

- [Backend README](backend/README.md)
- [ドキュメント目次](.claude/docs/README.md)
- [設定ガイド](.claude/README.md)
