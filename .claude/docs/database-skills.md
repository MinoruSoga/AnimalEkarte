# Database Skills ガイド

インストール済みのDB関連スキルと推奨プロンプトのリファレンス。

## スキル一覧

| スキル | 場所 | 用途 |
|--------|------|------|
| `/database` | `.claude/skills/database/` | PostgreSQL操作・マイグレーション・最適化 |
| `/analyzing-schema` | `.claude/skills/analyzing-schema/` | スキーマ分析・GORMモデル作成・リレーション確認 |
| `/database-design` | `~/.agents/skills/database-design/` | 要件・画面定義書からのスキーマ設計・正規化・インデックス戦略 |
| `/mermaid-diagram-specialist` | `~/.agents/skills/mermaid-diagram-specialist/` | ERD・フローチャート等のMermaid図生成 |

> `/database` `/analyzing-schema` はプロジェクトローカルスキル（`.claude/skills/`）。
> `/database-design` `/mermaid-diagram-specialist` はグローバルスキル（`~/.agents/skills/`）。

---

## `/database` — PostgreSQL操作・マイグレーション

**概要:** PostgreSQLのマイグレーション作成・実行、クエリパフォーマンス最適化、バックアップ・リストア操作を支援。

**命名規則:**
- マイグレーション: `YYYYMMDD_description`
- テーブル: `snake_case`・複数形
- カラム: `snake_case`

### 推奨プロンプト

```
/database
以下の仕様で新規マイグレーションファイルを作成してください。

【追加内容】
- テーブル名: xxx
- カラム: （列挙）
- 外部キー: （列挙）
- インデックス: （列挙）

backend/migrations/ の既存ファイルの命名規則・スタイルに合わせてください。
作成後に `make db` で適用を確認してください。
```

```
/database
backend/migrations/ 配下の全マイグレーションを確認し、以下をチェックしてください。

【チェック観点】
1. インデックスが適切に設定されているか（外部キー・頻繁に検索されるカラム）
2. NOT NULL制約の抜け漏れがないか
3. CASCADE設定が適切か（削除時の整合性）
4. テーブル名・カラム名の命名規則（snake_case・複数形）に違反がないか

問題箇所を列挙し、修正用SQLを提案してください。
```

```
/database
現在のクエリパフォーマンスを分析してください。

【確認手順】
1. PostgreSQL MCPで `EXPLAIN ANALYZE` を実行
2. Seq Scan が発生しているテーブルを特定
3. インデックス追加の提案（`CREATE INDEX CONCURRENTLY` を使用）
4. マイグレーションファイルとして出力
```

---

## `/analyzing-schema` — スキーマ分析・GORMモデル

**概要:** Go/GORM + PostgreSQL 環境に特化。現在の31テーブル構成を理解した上で、スキーマ変更・マイグレーション・GORMモデル定義を支援する。

**ワークフロー:**
1. GORMモデル定義を確認（`backend/internal/model/`）
2. PostgreSQL MCPで実テーブル構造を確認
3. 外部キー制約・関連テーブルを確認
4. マイグレーション内容を決定
5. GORMモデル更新 → マイグレーション作成

### 推奨プロンプト

```
/analyzing-schema
backend/internal/model/ 配下のGORMモデルとPostgreSQLの実テーブルの整合性を確認してください。

【チェック観点】
1. GORMモデルのフィールドと実カラムの型・制約が一致しているか
2. `gorm:"foreignKey:XXX"` の設定と実際の外部キー制約が一致しているか
3. `TableName()` が定義されている場合、実テーブル名と一致しているか
4. GORMのPreloadで使用しているフィールド名がGoのフィールド名と一致しているか

不整合があれば修正方針を提示してください。
```

```
/analyzing-schema
新機能 XXX のために以下のテーブル設計をしてください。

【要件】
- （機能の説明）
- （関連する既存テーブル）

【出力物】
1. GORMモデル定義（backend/internal/model/xxx.go）
2. マイグレーションSQL（backend/migrations/XXX_xxx.sql）
3. 既存テーブルとのリレーション図（テキスト形式）
```

```
/analyzing-schema
backend/migrations/001_init.sql のスキーマ全体を分析してください。

【確認観点】
1. テーブル間のリレーションが正しく設計されているか
2. 正規化が適切か（冗長データ・更新異常が起きないか）
3. インデックス設計の過不足
4. 将来的な拡張で問題になりそうな設計上の懸念点

改善提案があれば優先度付きで列挙してください。
```

---

---

## `/database-design` — 要件・画面定義書からのスキーマ設計

**概要:** 正規化・PK設計・リレーション・FK制約・インデックス戦略を原則に基づいて設計する。画面定義書や要件定義書を入力として、エンティティ抽出からスキーマ設計まで対応。

**設計チェックリスト（スキル内蔵）:**
- 正規化の判断（分離 vs 非正規化）
- PK選択（UUID / ULID / Auto-increment）
- タイムスタンプ戦略（created_at / updated_at / deleted_at）
- リレーション種別（1:1 / 1:N / N:M）
- FK の ON DELETE 設定（CASCADE / SET NULL / RESTRICT）

### 推奨プロンプト

```
/database-design
以下の画面定義書・要件からPostgreSQLのスキーマ設計をしてください。

【プロジェクト前提】
- DB: PostgreSQL 18
- ORM: GORM（Go）
- PK: UUID
- タイムスタンプ: created_at / updated_at / deleted_at（ソフトデリート）
- 命名規則: テーブル名はsnake_case・複数形、カラム名はsnake_case

【画面定義書 / 要件】
（ここに画面定義書の内容を貼り付け）

【出力物】
1. エンティティ一覧と属性
2. テーブル設計（カラム名・型・制約）
3. リレーション定義（FK・ON DELETE設定）
4. インデックス設計
5. 正規化の判断根拠
```

```
/database-design
spec.md の仕様定義書を読み込み、未実装の機能に必要なテーブルを洗い出してください。

【既存テーブル】
backend/migrations/001_init.sql を参照

【出力物】
1. 追加が必要なテーブル一覧
2. 既存テーブルへの追加カラム
3. 設計の判断根拠
```

---

## `/mermaid-diagram-specialist` — ERD・図の生成

**概要:** ERD・フローチャート・シーケンス図・C4アーキテクチャ図などをMermaid形式で生成。`/database-design` で設計したスキーマをERDとして可視化するのに最適。

### 推奨プロンプト

```
/mermaid-diagram-specialist
以下のテーブル設計からMermaid形式のERDを生成し、docs/ERD.md を更新してください。

【テーブル設計】
（/database-design の出力結果を貼り付け）

【出力形式】
- Mermaid erDiagram 記法
- カーディナリティ（||--o{ 等）を正確に記載
- 外部キーのリレーションを全て表現
```

```
/mermaid-diagram-specialist
backend/migrations/001_init.sql の全テーブルからERDを生成し、docs/ERD.md を最新状態に更新してください。
```

---

## 組み合わせワークフロー

### 新規テーブル追加時

```
# Step 1: スキーマ設計
/analyzing-schema
XXX機能のテーブル設計をしてください。
既存の31テーブルとのリレーションも考慮し、GORMモデルとマイグレーションSQLを作成してください。

# Step 2: マイグレーション適用・確認
/database
作成したマイグレーションを backend/migrations/ に配置し、
命名規則・制約・インデックスの問題がないか確認してください。
```

### 要件・画面定義書から ERD まで一気通貫

```
# Step 1: スキーマ設計
/database-design
以下の画面定義書からPostgreSQLスキーマを設計してください。
（画面定義書を貼り付け）

# Step 2: ERD生成
/mermaid-diagram-specialist
上記のスキーマ設計をMermaid ERDに変換し、docs/ERD.md を更新してください。

# Step 3: GORMモデル・マイグレーション作成
/analyzing-schema
設計したスキーマをもとにGORMモデルとマイグレーションSQLを作成してください。
```

### 既存スキーマの健全性チェック

```
/analyzing-schema /database
backend/ 全体のDB設計を以下の観点で監査してください。

1. GORMモデルとPostgreSQLスキーマの不整合
2. 外部キー制約の抜け漏れ
3. インデックス不足によるパフォーマンスリスク
4. 命名規則違反（テーブル名・カラム名）

問題を重大度順（Critical/Warning/Info）に列挙し、修正SQLを提示してください。
```
