# FE-002: 保険マスタ「種別」フィールドが保存されない（FE/BE仕様不一致）

## 重大度
**Medium** — UIで「種別」を入力しても DB に保存されない（サイレントドロップ）

## 症状

1. 保険マスタ 編集フォームを開く
2. 「種別」フィールドに値を入力
3. 保存ボタンをクリック → 「更新しました」トースト表示
4. DB確認: `insurances` テーブルに `category` カラム自体が存在しない

## 根本原因（3層の問題）

### 1. FE config: category-config.ts

```ts
// frontend/src/features/master/constants/category-config.ts
insurance: {
  labels: { code: "コード", name: "保険会社名", category: "種別" },
  showCategory: true,  // ← UIに「種別」フィールドが表示される
  ...
}
```

### 2. FE ロジック: use-master-items.ts

```ts
// frontend/src/features/master/hooks/use-master-items.ts
const update = (id, updates, callbacks) => {
  const req: UpdateMasterItemRequest = {
    name: updates.name,
    price: updates.price,
    status: updates.status,
    description: updates.description,
    // updates.category が req に含まれていない ← バグ
  };
  updateMutation.mutate({ id, req }, { ... });
};
```

### 3. FE API: update-master-item.ts

```ts
// frontend/src/features/master/api/update-master-item.ts
// payload に category フィールドを追加する処理がない
```

### 4. BE: insurancesテーブル

```sql
-- insurances テーブルに category カラムが存在しない
-- \d insurances: name, is_active, description, coverage_rate, contact_phone, sort_order
```

## 影響

保険マスタで「種別」を入力して保存しても：
- FE: `formData.category` が `req` に含まれずAPIに送信されない
- BE: 仮に送信されても `insurances` テーブルに保存先カラムがない
- 結果: サイレントに無視される（エラーなし）

## 修正方針

### 方針A: 「種別」フィールドを削除する（推奨）
`insurances` テーブルに `category` カラムが設計として不要であれば、FE の category-config から `showCategory: false` に変更する。

### 方針B: 「種別」フィールドを実装する
1. BE: `insurances` テーブルに `insurance_type text` カラムを追加（マイグレーション）
2. BE: `CreateInsurance` / `UpdateInsurance` ハンドラで `insurance_type` を処理
3. FE: `use-master-items.ts` の `update()` / `add()` に `category` フィールドを追加
4. FE: `update-master-item.ts` / `create-master-item.ts` のペイロードに `category` を追加

## 再現確認済み

- 保険マスタ登録フォームで「種別」フィールドに値を入力して保存
- 「更新しました」が表示されるが、DB（insurancesテーブル）に対応カラムなし
- 発見: マスタ設定ページ全ページ 登録テスト中に確認（2026-03-16）
