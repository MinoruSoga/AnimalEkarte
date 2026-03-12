# Design Extraction Skills ガイド

URLのスクリーンショットからデザインを解析し、React 19 + TypeScript + Tailwind CSS 4 コンポーネントを生成するスキルのリファレンス。

## インストール済みスキル一覧

| スキル | インストール元 | 用途 |
|--------|--------------|------|
| `design-spec-extraction` | `Cygnusfear/claude-stuff` | スクリーンショット→W3C DTCG準拠のJSON仕様抽出（7パス） |
| `design-md` | `google-labs-code/stitch-skills` | Google Stitch専用：デザインシステム→DESIGN.md生成 |

> **注意:** `design-md` は Google Stitch MCP が必要。一般的なURL解析には使用しない。

---

## `/url-to-react` — URLからReactコンポーネント生成

**概要:** URLを渡すだけで、Chrome DevTools MCPでスクリーンショットを取得し、
`design-spec-extraction`の7パスアーキテクチャでデザイン仕様を抽出、
React 19 + TypeScript + Tailwind CSS 4コンポーネントを生成する。

**実行フロー:**

```
URL入力
  ↓ [Step 1] Chrome DevTools MCP でスクリーンショット取得（1440x900）
  ↓ [Step 2] design-spec-extraction 7パス → .design-specs/{name}/design-spec.json
  ↓ [Step 3] JSON → React 19 + TypeScript + Tailwind CSS 4 コンポーネント生成
  ↓ [Step 4] 型安全・アクセシビリティ・React 19ルール 自己レビュー
```

### 基本プロンプト

```
/url-to-react
https://example.com のデザインを解析して、
React 19 + TypeScript + Tailwind CSS 4 のコンポーネントを生成してください。

【出力先】
frontend/src/components/shared/{ComponentName}/

【技術スタック（厳守）】
- React 19（FC禁止、forwardRef禁止、ref as prop）
- TypeScript（any禁止、Props型を明示定義）
- Tailwind CSS 4（カスタムカラーは @theme で定義）
- shadcn/ui コンポーネントを優先使用

【コンポーネント分割】
Atoms（ボタン・バッジ）→ Molecules（カード・フォームフィールド）→ Organisms（ヘッダー・セクション）の順で分割して出力してください。
```

### モバイルも含める場合

```
/url-to-react
https://example.com のデザインを解析してください。

【スクリーンショット要件】
- デスクトップ: 1440x900
- モバイル: iPhone 14 エミュレート（375x812）

【出力】
- デスクトップ・モバイル両方のレイアウトに対応したレスポンシブコンポーネント
- Tailwind の sm:/md:/lg: プレフィックスを使用
- 出力先: frontend/src/components/shared/{ComponentName}/
```

### 特定セクションのみ抽出

```
/url-to-react
https://example.com のヘッダーナビゲーション部分のみを解析して、
Reactコンポーネントを生成してください。

【抽出対象】
画面上部のナビゲーションバー（ロゴ、メニュー、CTAボタンを含む）

【出力】
frontend/src/components/shared/Navigation/
```

---

## `design-spec-extraction` — スクリーンショットからJSON仕様抽出

**概要:** 画像・スクリーンショットからW3C DTCG 2025.10準拠のJSONデザイン仕様を抽出する。
7パスのシリアル実行で網羅的に抽出し、各パスの結果をファイルに保存する。

**出力ファイル構成:**

```
.design-specs/{project}/
├── pass-1-layout.json      # レイアウト・構造
├── pass-2-colors.json      # カラーパレット
├── pass-3-typography.json  # タイポグラフィ
├── pass-4-components.json  # コンポーネントツリー
├── pass-5-spacing.json     # スペーシング・寸法
├── pass-6-states.json      # 状態・アクセシビリティ
└── design-spec.json        # ★ 最終統合仕様
```

### 推奨プロンプト（スクリーンショットを直接渡す場合）

```
design-spec-extraction スキルを使用して、
添付のスクリーンショットからデザイン仕様を抽出してください。

【プロジェクト名】 {project-name}
【出力ディレクトリ】 .design-specs/{project-name}/

7パスすべてを実行し、各パスの結果を必ずJSONファイルとして保存してください。
最終的に design-spec.json を生成してください。
```

### 抽出後にコード生成する場合

```
.design-specs/{project-name}/design-spec.json を読み込んで、
以下の条件でReactコンポーネントを生成してください。

【技術スタック】
- React 19（FC禁止、forwardRef禁止）
- TypeScript（any禁止）
- Tailwind CSS 4（カラーは @theme に定義）
- shadcn/ui 優先使用

【カラートークンの変換】
design-spec.json の tokens.colors を frontend/src/styles/globals.css の @theme に追加してください:
@theme {
  --color-{name}: {value};
}

【出力先】
frontend/src/components/shared/{ComponentName}/
```

---

## `design-md` — Google Stitch プロジェクト専用

**概要:** Google StitchのプロジェクトからDESIGN.mdを生成する。
**Stitch MCP Serverが必要**。一般的なウェブサイトには使用しない。

### 推奨プロンプト

```
design-md スキルを使用して、
Stitch プロジェクト {project-id} の {screen-name} 画面からDESIGN.mdを生成してください。

【出力先】
docs/DESIGN.md
```

---

## よくある使い方

### 1. 競合他社サイトのUIを参考にコンポーネント化

```
/url-to-react
https://competitor.com/pricing のpricing テーブルセクションのデザインを参考に、
このプロジェクト（Animal Ekarte）のスタイルに合わせたPricingTableコンポーネントを生成してください。

【出力先】 frontend/src/features/hospital-settings/components/PricingTable.tsx
【参考のみ】著作権に注意し、レイアウト・構造パターンのみを参考にすること
```

### 2. 既存ページのデザイントークン抽出

```
Chrome DevTools MCP で http://localhost:3000/hospitalization のスクリーンショットを取得し、
design-spec-extraction で現在のデザイントークンをJSON化してください。

【目的】 デザインシステムのドキュメント化
【出力】 .design-specs/animalekarte-hospitalization/design-spec.json
```

### 3. Figmaエクスポート画像からコンポーネント生成

```
添付のFigmaエクスポート画像（dashboard-mockup.png）を解析して、
ダッシュボードのReactコンポーネントを生成してください。

design-spec-extraction の7パスを実行し、
design-spec.json からコンポーネントを生成してください。

【出力先】 frontend/src/features/dashboard/components/
```

---

## トラブルシューティング

| 問題 | 対処 |
|------|------|
| スクリーンショット取得失敗（認証が必要なページ） | スクリーンショットを手動で撮影して直接渡す |
| `design-md` が動かない | Stitch MCP Serverが設定されていない。一般URLには使わない |
| 7パスが途中で止まる | 前パスのJSONファイルが存在するか確認。再開は途中のパスから可能 |
| Tailwindクラスが認識されない | Tailwind CSS 4 の場合は `@theme` でカスタム変数定義が必要 |
