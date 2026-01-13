# Animal Ekarte Frontend

動物病院 電子カルテシステムのフロントエンド

## 技術スタック

- **言語**: TypeScript 5.7
- **フレームワーク**: React 19
- **ビルドツール**: Vite 6
- **ルーティング**: React Router v7
- **スタイリング**: CSS（カスタム）
- **アーキテクチャ**: bulletproof-react

## ディレクトリ構成（bulletproof-react）

```
frontend/src/
├── app/                    # アプリケーション層
│   ├── index.tsx           # Appコンポーネント
│   ├── provider.tsx        # プロバイダー設定
│   ├── layouts/            # レイアウトコンポーネント
│   │   └── app-layout.tsx
│   └── routes/             # ルーティング設定
│       └── index.tsx
│
├── config/                 # グローバル設定
│   └── index.ts
│
├── features/               # 機能別モジュール
│   └── pets/               # ペット機能
│       ├── api/            # API呼び出し
│       │   ├── index.ts
│       │   ├── get-pets.ts
│       │   ├── get-pet.ts
│       │   ├── create-pet.ts
│       │   ├── update-pet.ts
│       │   └── delete-pet.ts
│       ├── components/     # 機能固有コンポーネント
│       │   ├── index.ts
│       │   ├── pet-form.tsx
│       │   ├── pet-list.tsx
│       │   └── pets-page.tsx
│       ├── hooks/          # カスタムフック
│       │   ├── index.ts
│       │   ├── use-pets.ts
│       │   └── use-pet-mutations.ts
│       ├── types/          # 型定義
│       │   └── index.ts
│       └── index.ts        # バレルエクスポート
│
├── lib/                    # 共通ライブラリ
│   └── api-client.ts       # APIクライアント
│
├── styles/                 # グローバルスタイル
│   └── index.css
│
└── main.tsx                # エントリーポイント
```

## bulletproof-react の原則

### 1. Feature-Based Architecture
機能ごとにモジュールを分離し、関連するコードを同じ場所に配置:
- `api/` - データ取得・更新
- `components/` - UIコンポーネント
- `hooks/` - ビジネスロジック
- `types/` - 型定義

### 2. Barrel Exports（index.ts）
各ディレクトリに `index.ts` を配置し、パブリックAPIを明示:
```typescript
// features/pets/index.ts
export * from './components/pets-page';
export * from './types';
```

### 3. Absolute Imports
`@/` エイリアスを使用した絶対パス:
```typescript
import { apiClient } from '@/lib/api-client';
import { PetsPage } from '@/features/pets';
```

## 開発環境セットアップ

### Docker使用

```bash
# コンテナ起動
make build

# ログ確認
make logs-front

# 再起動
make restart-front
```

### ローカル開発

```bash
npm install
npm run dev
```

## 新しい機能（Feature）の追加手順

### 1. Feature ディレクトリ作成

```bash
mkdir -p src/features/owners/{api,components,hooks,types}
```

### 2. 型定義（types/index.ts）

```typescript
// src/features/owners/types/index.ts
export type Owner = {
  id: string;
  name: string;
  phone: string;
  email: string;
  created_at: string;
  updated_at: string;
};

export type CreateOwnerInput = {
  name: string;
  phone?: string;
  email?: string;
};
```

### 3. API関数（api/）

```typescript
// src/features/owners/api/get-owners.ts
import { apiClient } from '@/lib/api-client';
import type { Owner } from '../types';

export const getOwners = (): Promise<Owner[]> => {
  return apiClient.get<Owner[]>('/owners');
};
```

```typescript
// src/features/owners/api/index.ts
export * from './get-owners';
export * from './create-owner';
// ...
```

### 4. カスタムフック（hooks/）

```typescript
// src/features/owners/hooks/use-owners.ts
import { useState, useEffect, useCallback } from 'react';
import type { Owner } from '../types';
import { getOwners } from '../api';

export const useOwners = () => {
  const [owners, setOwners] = useState<Owner[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchOwners = useCallback(async () => {
    try {
      setIsLoading(true);
      const data = await getOwners();
      setOwners(data || []);
    } catch {
      setError('飼い主一覧の取得に失敗しました');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchOwners();
  }, [fetchOwners]);

  return { owners, isLoading, error, refetch: fetchOwners };
};
```

### 5. コンポーネント（components/）

```typescript
// src/features/owners/components/owners-page.tsx
import { useOwners } from '../hooks';
import { OwnerList } from './owner-list';

export const OwnersPage = () => {
  const { owners, isLoading, error } = useOwners();

  if (isLoading) return <p>読み込み中...</p>;
  if (error) return <p className="error-message">{error}</p>;

  return <OwnerList owners={owners} />;
};
```

### 6. バレルエクスポート（index.ts）

```typescript
// src/features/owners/index.ts
export * from './components/owners-page';
export * from './types';
```

### 7. ルーティング追加

```typescript
// src/app/routes/index.tsx
import { OwnersPage } from '@/features/owners';

const router = createBrowserRouter([
  {
    path: '/',
    element: <AppLayout />,
    children: [
      { index: true, element: <PetsPage /> },
      { path: 'owners', element: <OwnersPage /> },  // 追加
    ],
  },
]);
```

## コード規約

### ファイル命名
- コンポーネント: `kebab-case.tsx`（例: `pet-form.tsx`）
- フック: `use-*.ts`（例: `use-pets.ts`）
- 型: `index.ts`（ディレクトリ内）

### インポート順序
1. React / フレームワーク
2. 外部ライブラリ
3. 内部モジュール（@/）
4. 相対インポート
5. 型インポート

```typescript
import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '@/lib/api-client';
import { PetForm } from './pet-form';
import type { Pet } from '../types';
```

### 禁止事項
- `any` 型の使用
- 未使用のインポート
- 本番コードでの `console.log`
- ハードコードされた値（環境変数または定数を使用）

## ビルド

```bash
# 本番ビルド
npm run build

# プレビュー
npm run preview

# 型チェック
npx tsc --noEmit
```

## トラブルシューティング

### パスエイリアスエラー
`@/` が解決できない場合、`vite.config.ts` と `tsconfig.json` の設定を確認。

### APIエラー
バックエンドが起動しているか確認し、`vite.config.ts` のプロキシ設定を確認。
