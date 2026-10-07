# UI コンポーネント仕様 インデックス (Component Specifications Index)

> **目的**: UI プリミティブのコンポーネント契約（Anatomy・状態表・既知ギャップ）の索引を提供する。
> **読者**: UI コンポーネントを実装・変更する開発者。
> **タイミング**: `frontend/src/components/ui/` 配下の実装変更前。

正本関係: 値は [design-system.md](../design-system.md)、状態は [design-states.md](../design-states.md)、層は [design-token-layers.md](../design-token-layers.md)。各仕様書は「実装の現在形」を記述し、理想形との差分は「既知のギャップ」節に置く（[テンプレート](_TEMPLATE.md) 同様）。

| コンポーネント | 概要 |
|:---|:---|
| [badge.md](badge.md) | ステータス表示バッジ（`badge.tsx`。約 17 ファイルで使用 + `BADGE.*` トークンが status-helpers 経由で数十ステータスに波及） |
| [button.md](button.md) | 汎用ボタン（`button.tsx`。約 120 ファイルで使用する最大消費プリミティブ） |
| [dialog.md](dialog.md) | Radix Dialog ベースのモーダル（`dialog.tsx`） |
| [input.md](input.md) | テキスト入力（`input.tsx`。約 52 ファイルで使用） |
| [select.md](select.md) | Radix Select ベースのドロップダウン選択（`select.tsx`） |
| [table.md](table.md) | 業務 data-table プリミティブ（`table.tsx`。監査 C18: cell override 禁止・C19: row onClick 禁止） |
| [tabs.md](tabs.md) | Radix Tabs ベースのタブ切替（`tabs.tsx`） |
| [_TEMPLATE.md](_TEMPLATE.md) | 新規コンポーネント仕様書テンプレート（`_TEMPLATE.md` をコピーして `<component-name>.md` として起票） |

---

**最新更新**: 2026-10-07 | **ステータス**: 本 README がファイル単位の索引の正本（docs/README.md 2層索引の第 2 層）
