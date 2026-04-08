# BUG-123: /settings インデックスページのマスタカードが権限フィルタリングされていない

## 概要
`/settings` のマスタ設定インデックスページ（`MasterSettingsIndex`）で、全マスタカードが
ハードコードで表示されている。ユーザーの権限に応じたフィルタリングが行われておらず、
`master-permission` の view 権限がない一般ユーザーにも「権限グループマスタ」カードが表示される。

## 脆弱性分類
- **CWE-284**: Improper Access Control (UI層)
- **影響**: 権限のない機能のカードが表示される（UX 問題 + カードクリックでページ遷移可能）

## 再現手順
1. `vet@example.com` / `password`（一般権限）でログイン
2. `http://localhost:3003/settings` にアクセス
3. ページ下部「スタッフ・権限」セクションを確認
4. **結果**: 「権限グループマスタ」カードが表示される（一般権限では `master-permission` = アクセス不可）

## 期待する動作
- 各マスタカードに対応するリソースの `canView` を確認し、権限がないカードは非表示にする
- カテゴリ内の全カードが非表示の場合、カテゴリ見出し（セクション）も非表示にする

## 現状コード

### `frontend/src/features/master/routes/MasterSettingsIndex.tsx`

#### セクション定義（権限チェックなし）
```typescript
const MASTER_SECTIONS: SectionDef[] = [
  { title: "基本設定", keys: ["clinic", "animal_species"] },
  { title: "カルテ", keys: ["treatmentItems", "diagnosisGroup", "inquiry_template", "medicine"] },
  { title: "診療関連マスタ", keys: ["serviceType"] },
  { title: "入院・ケージ管理", keys: ["hospitalization", "cage"] },
  { title: "トリミング関連", keys: ["trimmingGroup"] },
  { title: "会計・商品", keys: ["merchandise_item", "insurance"] },
  { title: "スタッフ・権限", keys: ["staff", "occupations", "permission_group"] },
  // ↑ "permission_group" が無条件に表示される
];
```

#### renderCard 関数（権限チェックなし）
```typescript
function renderCard(key: MasterCardKey) {
  // Check if it is a group key
  if (key in GROUP_CARD_CONFIG) {
    const groupKey = key as GroupKey;
    const cfg = GROUP_CARD_CONFIG[groupKey];
    const Icon = cfg.IconComponent;
    return (
      <CardRow
        key={groupKey}
        label={cfg.label}
        description={cfg.description}
        icon={<Icon className={ICON.action} />}
        count={undefined}
        onClick={() => navigate(cfg.path)}
      />
    );
  }

  // Individual category card — 権限チェックなし
  const cat = key as MasterSettingsCategory;
  const cfg = CATEGORY_CONFIG[cat];
  const Icon = cfg.IconComponent;
  return (
    <CardRow
      key={cat}
      label={cfg.label}
      description={cfg.description}
      icon={<Icon className={ICON.action} />}
      count={undefined}
      onClick={() => navigate(cfg.settingsPath)}
    />
  );
}
```

#### セクションレンダリング（フィルタなし）
```typescript
{MASTER_SECTIONS.map((section) => (
  <div key={section.title} className="mb-5">
    <div className={`px-1 pb-1.5 text-base ${C.text40} ...`}>
      {section.title}
    </div>
    <div className={`bg-white rounded-lg border ...`}>
      {section.keys.map((key) => renderCard(key))}
      {/* ↑ 全カードが無条件にレンダリングされる */}
    </div>
  </div>
))}
```

### 比較: サイドバーの正しい実装

`frontend/src/components/shared/Layout/Sidebar.tsx:134-162`:
```typescript
function SidebarItemWithPermission({ item }: { item: NavItem }) {
  const { canView } = usePermission(item.resource ?? "");
  if (item.resource !== undefined && !canView) return null;  // ✅ 権限なしなら非表示
  // ...
}
```

## カードとリソースのマッピング

| カードキー | 表示ラベル | リソース定数 | 一般権限 canView |
|-----------|-----------|-------------|-----------------|
| `clinic` | 医院マスタ | `ResourceHospitalSettings` | true (view only) |
| `animal_species` | 動物種類マスタ | `ResourceMasterAnimalSpecies` | true |
| `treatmentItems` | 診療項目マスタ | `ResourceMasterMedical` | true |
| `diagnosisGroup` | 診断マスタ | `ResourceMasterMedical` | true |
| `inquiry_template` | 問診テンプレート | `ResourceMasterMedical` | true |
| `medicine` | 薬剤マスタ | `ResourceMasterMedical` | true |
| `serviceType` | 予約区分マスタ | `ResourceMasterServiceType` | true |
| `hospitalization` | 入院マスタ | `ResourceMasterHosp` | true |
| `cage` | ケージマスタ | `ResourceMasterHosp` | true |
| `trimmingGroup` | トリミングマスタ | `ResourceMasterTrim` | true |
| `merchandise_item` | 商品マスタ | `ResourceMasterMerchandise` | true |
| `insurance` | 保険マスタ | `ResourceMasterInsurance` | true |
| `staff` | スタッフマスタ | `ResourceMasterStaff` | true |
| `occupations` | 職種マスタ | `ResourceMasterStaff` | true |
| `permission_group` | 権限グループマスタ | `ResourceMasterPermission` | **false** |

## 修正方針

### 1. カードキーに resource プロパティを追加

`CATEGORY_CONFIG` と `GROUP_CARD_CONFIG` に `resource` フィールドを追加:

```typescript
// constants/category-config.ts に追加
interface CategoryConfigEntry {
  // ... 既存フィールド
  resource: Resource;  // 追加
}
```

### 2. 権限チェック付きカードラッパーコンポーネント

React Hook のルール上、`usePermission` は条件付きで呼べないため、
カード単位のラッパーコンポーネントを作成:

```typescript
function PermissionFilteredCard({ cardKey, renderCard }: {
  cardKey: MasterCardKey;
  renderCard: (key: MasterCardKey) => ReactNode;
}) {
  const resource = getResourceForCardKey(cardKey);
  const { canView } = usePermission(resource);
  if (!canView) return null;
  return <>{renderCard(cardKey)}</>;
}
```

### 3. セクションの空チェック

カテゴリ内の全カードが非表示の場合、セクション見出しも非表示にする:

```typescript
{MASTER_SECTIONS.map((section) => {
  const visibleCards = section.keys.filter(key => {
    const resource = getResourceForCardKey(key);
    return hasPermission(resource, "view");
  });
  if (visibleCards.length === 0) return null;  // セクションごと非表示

  return (
    <div key={section.title} className="mb-5">
      {/* ... */}
      {visibleCards.map((key) => renderCard(key))}
    </div>
  );
})}
```

**注意**: `hasPermission` は hook 内でのみ使用可能。上記は概念的な擬似コード。
実装時はサイドバーの `SidebarItemWithPermission` パターンを参考にすること。

## 優先度
**Medium** — 情報漏洩ではないが、UX 上の問題。
BUG-121 でルートガードが適用されれば、カードをクリックしても `AccessDenied` が表示されるため、
セキュリティ上の実害は BUG-121 修正後は軽微。

## 関連チケット
- BUG-121: `/settings` ルートガード未適用（先に対応すべき）
- BUG-122: バックエンド API 権限チェック

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Frontend ベストプラクティス
- **Conditional Render**: 権限チェック結果に基づくカード表示/非表示は `? (...) : null` を使用（`&&` 禁止）
- **Design Tokens**: カードのスタイリングは既存の `C`, `STYLE`, `ICON` 定数を維持（`MasterSettingsIndex` は準拠済み）
- **Feature Indexing**: Resource 定数は `@/types/generated/models` から import

### `.claude/rules/typescript-react.md` — React 19 Patterns
- **Hook のルール**: `usePermission` は条件分岐内で呼べない。カード単位のラッパーコンポーネントで対応すること
- **memo()**: 新規ラッパーコンポーネントがリスト内で大量にレンダリングされる場合は `memo()` を検討
- **useCallback**: ラッパーコンポーネントに渡す `renderCard` 関数は `useCallback` で安定化

### `.claude/rules/security.md` — Input Validation
> "Validate on both client and server"

UI レベルの権限フィルタリングは **UX 改善** が主目的。
セキュリティ上は BUG-121（ルートガード）と BUG-122（API 権限チェック）が防御の本体。
本チケットは多層防御の UI 層として機能する。

### プロジェクト内参照実装
`frontend/src/components/shared/Layout/Sidebar.tsx:134-162` の `SidebarItemWithPermission` が
同一パターンの参照実装:

```typescript
// ✅ 参照実装: サイドバーの権限フィルタリング
function SidebarItemWithPermission({ item }: { item: NavItem }) {
  const { canView } = usePermission(item.resource ?? "");
  if (item.resource !== undefined && !canView) return null;  // 権限なしなら非表示
  return <SidebarItem item={item} />;
}
```

**`MasterSettingsIndex` のカードにも同じパターンを適用する。**

## 関連ファイル
- `frontend/src/features/master/routes/MasterSettingsIndex.tsx` — 修正対象
- `frontend/src/features/master/constants/category-config.ts` — カテゴリ設定（resource 追加）
- `frontend/src/components/shared/Layout/Sidebar.tsx:134-162` — 参考実装（`SidebarItemWithPermission`）
- `frontend/src/features/auth/hooks/use-auth.tsx:104-114` — `hasPermission` / `usePermission`
