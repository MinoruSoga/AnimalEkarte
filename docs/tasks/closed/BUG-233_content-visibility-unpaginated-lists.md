# BUG-233: 大量アイテムのリストに content-visibility 未適用（MedicineSettings, TreatmentPlanMaster）

## 概要
`MedicineSettings` および `TreatmentPlanMaster` は100〜200件以上のアイテムをページネーションなしで全件レンダーしている。CSS の `content-visibility: auto` を適用することで、ビューポート外の行のレンダーコストをスキップし、初期描画とスクロール性能を改善できる。

## 現状コード

### `features/master/routes/MedicineSettings.tsx`（推定）
```typescript
// ❌ 全件レンダー — スクロール外の200行も即時レンダー
return (
  <div>
    {medicines.map(medicine => (
      <MedicineRow key={medicine.id} medicine={medicine} />
    ))}
  </div>
);
```

### `features/master/routes/TreatmentPlanMaster.tsx`（推定）
```typescript
// ❌ 同様
return (
  <div>
    {treatmentPlans.map(plan => (
      <TreatmentPlanRow key={plan.id} plan={plan} />
    ))}
  </div>
);
```

## 修正方針

各行（または行コンテナ）に `content-visibility: auto` と `contain-intrinsic-size` を設定する。

```typescript
// ✅ CSS で content-visibility を適用（Tailwind カスタムクラスまたはインラインスタイル）
// 方法1: Tailwind CSS v4 カスタムユーティリティ
// globals.css に追加
// .cv-auto { content-visibility: auto; contain-intrinsic-size: 0 56px; }

// 方法2: インラインスタイル（行の高さが固定の場合）
{medicines.map(medicine => (
  <div
    key={medicine.id}
    style={{ contentVisibility: "auto", containIntrinsicSize: "0 56px" }}
  >
    <MedicineRow medicine={medicine} />
  </div>
))}
```

**注意**: `contain-intrinsic-size` には行の推定高さを指定する。不正確な場合はスクロールバーのジャンプが発生する。確認してから適用すること。

## 影響範囲

| コンポーネント | 推定アイテム数 | 優先度 |
|-------------|-------------|-------|
| `MedicineSettings.tsx` | 200件以上（薬品マスタ） | Medium |
| `TreatmentPlanMaster.tsx` | 100件以上（処置マスタ） | Medium |

## 代替案（より根本的解決）

ページネーション（`Pagination` コンポーネントが既存）またはバーチャルスクロール（`@tanstack/virtual`）を導入する。ただし実装コストが高いため、まず `content-visibility` で即効対処する。

## 準拠すべきプロジェクト規約

### `frontend/CODING_RULES.md` Section 12 — rendering-content-visibility
> ページネーションなしで100件以上をレンダーする場合は `content-visibility: auto` を適用する

## 優先度
**Medium** — マスタデータが増えると顕著に遅くなる。薬品・処置マスタは追加し続けるデータのため早めに対処すべき。

## 関連ファイル
- `frontend/src/features/master/routes/MedicineSettings.tsx`
- `frontend/src/features/master/routes/TreatmentPlanMaster.tsx`
- `frontend/src/styles/globals.css` — カスタムユーティリティ追加先
