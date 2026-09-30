# コンポーネント仕様書テンプレート

> 使い方: このファイルを `docs/spec/components/<component-name>.md` にコピーして記述する。
> 正本関係: 値=design-system.md、状態=design-states.md、層=design-token-layers.md。
> 仕様書は「実装の現在形」を記述する。理想形との差分は「既知のギャップ」節に書く。

```markdown
# <ComponentName>

> 1行説明 / 配置: `frontend/src/components/ui/<name>.tsx`（variants: `<name>-variants.ts`）
> 使用箇所: 約 N ファイル / Storybook: `components/ui/<name>.stories.tsx`

## Anatomy
<!-- DOM構造・スロット・data-slot 属性 -->

## Props / Variants
| prop | 型 | 既定 | 説明 |

## 状態（design-states.md §1 に基づく）
| 状態 | 実装 | トークン/クラス | 備考 |
|------|------|----------------|------|
| hover | ○ | … | |
| focus-visible | ○ | … | |
| disabled | ○ | … | |
| loading | ○/✕ | … | |
| invalid | ○/✕/— | … | |

## アクセシビリティ要件
- ロール/aria 属性、キーボード操作、必須ラベル

## 使い方（Do / Don't）
- Do / Don't を箇条書き

## 既知のギャップ
- 仕様と実装の差異、未実装の状態、計画中の改善
```

## 記述ルール

- 状態表は design-states.md §1 のマトリクスの行順に揃える。実装がない状態は `✕` とし「既知のギャップ」にも記載する。
- 「使用箇所」は `rg -l "@/components/ui/<name>"` の実測値を書く（概数可）。
- Story の有無をヘッダに書く。ui/* プリミティブは Story を持つのが目標（監査提案 C20）。
