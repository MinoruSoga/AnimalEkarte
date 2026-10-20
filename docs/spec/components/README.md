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

## カバレッジ（2026-10-20 静的棚卸し）

`frontend/src/components/ui/` の UI プリミティブは **25 本**（`*-variants.ts` / `utils.ts` / `dialog-portaled-overlay.ts` のヘルパを除く）。全 25 本に colocated `*.stories.tsx` があり、design-audit **C21**（`C21_STORIES_ALLOWLIST` は空）が欠落を strict fail で検出する。

| 正本の形 | 本数 | 対象 |
|:--|--:|:--|
| md 仕様書 + stories | 7 | 上表の badge / button / dialog / input / select / table / tabs |
| stories + autodocs のみ（FE-DS2 決裁: 契約が stories から読み取れるもの） | 18 | alert-dialog / calendar / card / checkbox / command / dropdown-menu / label / popover / radio-group / scroll-area / searchable-select / separator / sheet / sonner / switch / textarea / toggle-group / tooltip |

md を追加する基準は design-system.md FE-DS2 のとおり「stories を読んでも分からない振る舞い契約」がある場合のみ。候補として `alert-dialog`（確認必須操作の契約。dialog.md が一部カバー）と `searchable-select`（検索・空結果・読込中の状態遷移）を認識しているが、未起票・未決定。新しいプリミティブを足したら本表の本数と行を同一コミットで更新する。

---

**最新更新**: 2026-10-20（カバレッジ節を追加） | **ステータス**: 本 README がファイル単位の索引の正本（docs/README.md 2層索引の第 2 層）

## このフォルダの役割と使い方

UIプリミティブの構造・状態・契約の入口。

コンポーネント仕様と既存テンプレートを置く。業務フロー・個別画面の操作許可・設計値の第二正本は置かない。

子フォルダ：現在、子フォルダはありません。

主要文書：主要文書は本書の既存索引・上記の読む順を参照します。

読む順：[デザイン値](../design-system.md)・[状態](../design-states.md)・[層](../design-token-layers.md) → 本索引 → 各部品仕様。実装の現在形と既知ギャップという既存の記述方針を維持。

正本・状態：[決定状態と検証状態の規約](../authoring-guide.md#status)に従い、各節・FRの承認根拠と実装／実機状態を別に読む。READMEへの掲載・フォルダ配置は承認や実装済みを表さず、対象外の既存仕様を新たに確定しない。既存の別名入口・索引の責務は維持し、本書は案内を補う。
