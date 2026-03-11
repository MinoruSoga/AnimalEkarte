# React 命名規則 Skills ガイド

インストール済みの命名規則関連スキルと推奨プロンプトのリファレンス。

## インストール済みスキル一覧

| スキル | インストール元 | 用途 |
|--------|--------------|------|
| `/naming-analyzer` | `softaworks/agent-toolkit` | 変数・関数・型名の分析・改善提案 |
| `/component-naming` | `sgcarstrends/sgcarstrends` | Reactコンポーネント命名規則の適用 |

---

## `/naming-analyzer` — 変数・関数・型名の分析

**概要:** 変数・定数・関数・クラス・インターフェース・型・APIエンドポイントなど全ての命名を分析し、改善提案を行う。React/TypeScriptの規約（camelCase/PascalCase/UPPER_SNAKE_CASE）にも対応。

**チェック観点:**
- 曖昧・不明瞭な名前
- 意味を損なう略語
- 不一致な命名規則
- 名前と挙動が乖離している
- 長すぎ・短すぎ
- Boolean に `is`/`has`/`can`/`should` プレフィックスがついているか
****
### 推奨プロンプト

```
/naming-analyzer
frontend/src/features/owners/ 配下の全ファイルの命名を分析し、改善提案をしてください。

【プロジェクト規約（CLAUDE.md準拠）】
- コンポーネント: PascalCase
- 関数・変数: camelCase
- 定数: UPPER_SNAKE_CASE
- 型・Interface: PascalCase
- ファイル: kebab-case
- Boolean変数: is/has/can/should プレフィックス
- カスタムhooks: use プレフィックス

問題箇所を重大度（Critical/Warning/Info）で分類して報告し、修正してください。
```

```
/naming-analyzer
frontend/src/features/ 配下の hooks（useXxx.ts）の命名を分析してください。

【チェック観点】
- use プレフィックスが統一されているか
- hooks の返り値の変数名が意図を明確に示しているか
- イベントハンドラが handle プレフィックスで統一されているか
```

```
/naming-analyzer
frontend/src/features/accounting/api/ 配下のAPI関数・型名を分析してください。

【チェック観点】
- APIレスポンス型が Response/Dto サフィックスで統一されているか
- transform関数の命名が一貫しているか（to~/from~）
- React Queryのキー名が統一されているか
```

---

## `/component-naming` — Reactコンポーネント命名

**概要:** Reactコンポーネントに「ドメイン + 役割」パターンを適用し、PascalCase・複合コンポーネント・ファイル名の一致を強制する。

**命名ルール:**
1. **PascalCase** — 全コンポーネントに適用
2. **ドメイン + 役割パターン** — `OwnerCard`（Owner=ドメイン, Card=役割）
3. **複合コンポーネント** — サブパーツはドット記法（`OwnerCard.Title`）
4. **ファイル名とコンポーネント名を一致** — `OwnerCard.tsx` → `export function OwnerCard`

### 推奨プロンプト

```
/component-naming
frontend/src/features/ 配下の全コンポーネントの命名を確認し、規約違反を修正してください。

【プロジェクト規約】
- ドメイン + 役割パターン（例: OwnerCard, PetEditModal, DashboardKanban）
- PascalCase 必須
- ファイル名は kebab-case（owner-card.tsx）、コンポーネント名は PascalCase（OwnerCard）
- 汎用すぎる名前禁止（Card, Item, List 単体はNG）

違反箇所を列挙し、適切な名前を提案して修正してください。
```

```
/component-naming
frontend/src/components/shared/ 配下の共有コンポーネントの命名を確認してください。

【チェック観点】
- shared コンポーネントは役割が明確な名前になっているか（DataTable, SortableHeader など）
- feature固有の名前が混入していないか
- Props型の命名が ComponentNameProps パターンに統一されているか
```

---

## 組み合わせワークフロー

### 新規 feature 実装後の命名チェック

```
# Step 1: コンポーネント名の確認・修正
/component-naming
frontend/src/features/[feature名]/ 配下の全コンポーネントを命名規約に照らして修正してください。

# Step 2: 変数・関数・型名の分析
/naming-analyzer
frontend/src/features/[feature名]/ 配下の全ファイルの命名を分析し、改善してください。
```

### frontend 全体の命名統一

```
/naming-analyzer /component-naming
frontend/src/features/ 配下の全ファイルを命名規約に基づいてリファクタリングしてください。

【優先順位】
1. コンポーネント名（PascalCase + ドメイン+役割パターン）
2. カスタムhooks（useプレフィックス統一）
3. イベントハンドラ（handleプレフィックス統一）
4. Boolean変数（is/has/can/shouldプレフィックス）
5. 型・Interface名（PascalCase + 意図を示す名前）

修正後に `docker compose exec frontend npm run lint` でエラーがないことを確認してください。
```
