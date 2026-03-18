# テスト戦略 仕様書

本ドキュメントは、Animal Ekarte におけるテストの方針、カバレッジ目標、および実施手順を定義します。

## 1. 全体方針

本システムでは、ビジネスロジックの核心であるバックエンドの **Service層** に対して重点的にテストを配置します。フロントエンドについては、ユーティリティや共通ロジックを中心に、費用対効果の高い箇所からテストを実装します。

---

## 2. バックエンド・テスト (Go)

### 2.1 テスト対象と優先順位
1.  **Service層 (最優先)**: 全てのビジネスバリデーション、データ変換ロジック、計算ロジック。
2.  **Repository層**: 複雑な SQL クエリや GORM の Preload ロジックの検証。
3.  **Handler層**: HTTP ステータスコードと JSON レスポンス形式の検証（主に結合テスト）。

### 2.2 テストパターン: Table-Driven Test
Go の標準的な `struct` 配列を用いたテーブル駆動テストを採用し、正常系・異常系のバリエーションを網羅します。

```go
func TestXxxService_Action(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        mockRet MockReturn
        want    Expectation
        wantErr bool
    }{
        // テストケースをここに列挙
    }
    // ... loop and run
}
```

### 2.3 モック戦略
`Repository` インターフェースに対するカスタムモック実装を作成し、DB への依存を排除した高速な単体テストを実施します。

### 2.4 実行コマンド
```bash
make test        # 全テスト実行
make test-cover  # カバレッジレポート生成
```

---

## 3. フロントエンド・テスト (TypeScript)

### 3.1 採用ライブラリ
- **テストランナー**: `Vitest`
- **テスティングライブラリ**: `React Testing Library`, `user-event`

### 3.2 テスト対象
1.  **Utils / Transforms**: 日付フォーマットや、API レスポンスから UI 用データへの変換ロジック。
2.  **Shared Hooks**: 複数の機能で使われる共通ロジック (`useSortableList` 等)。
3.  **Complex Components (予定)**: `TreatmentTable` 等の複雑なインタラクションを持つ部品。

### 3.3 モック戦略
- **APIモック**: 現時点では疎らですが、将来的に `MSW (Mock Service Worker)` を導入し、ネットワーク層を完全にモック化する方針です。

### 3.4 実行コマンド
```bash
docker compose exec frontend npm run test:run
```

---

## 4. 品質指標 (KPI)

- **バックエンド Service層**: カバレッジ 80% 以上を目標。
- **クリティカルな計算ロジック**: 100% の網羅。
- **CI連携**: GitHub Actions 上でプルリクエストごとに全テストがパスすることを必須とします。
