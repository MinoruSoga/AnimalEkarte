# Animal Ekarte Frontend

動物病院 電子カルテシステムのフロントエンド

## 技術スタック

- **言語**: TypeScript 5.7
- **フレームワーク**: React 18
- **ビルドツール**: Vite 6
- **スタイル**: Tailwind CSS 4
- **UIライブラリ**: shadcn/ui (Radix UI)
- **ルーティング**: React Router v6
- **アイコン**: lucide-react

## ディレクトリ構成

```
frontend/src/
├── components/           # 共通コンポーネント
│   ├── ui/               # shadcn/ui コンポーネント (40+)
│   │   ├── button.tsx
│   │   ├── dialog.tsx
│   │   ├── select.tsx
│   │   └── ...
│   ├── shared/           # 共有UIコンポーネント
│   │   ├── PageLayout.tsx
│   │   ├── PatientInfoCard.tsx
│   │   ├── DataTable.tsx
│   │   └── ...
│   ├── figma/            # Figma生成ヘルパー
│   └── Sidebar.tsx       # サイドバーナビゲーション
│
├── features/             # 機能別モジュール
│   ├── dashboard/        # ダッシュボード（カンバン）
│   │   ├── components/
│   │   ├── hooks/
│   │   └── routes/
│   ├── owners/           # 飼い主管理
│   ├── medical-records/  # カルテ管理
│   ├── reservations/     # 予約管理
│   ├── hospitalization/  # 入院管理
│   ├── examinations/     # 検査管理
│   ├── accounting/       # 会計
│   ├── vaccinations/     # ワクチン管理
│   ├── trimming/         # トリミング
│   ├── master/           # マスタ設定
│   ├── clinic/           # クリニック設定
│   └── inventory/        # 在庫管理（スタブ）
│
├── lib/                  # ユーティリティ
│   ├── constants.ts      # 定数・モックデータ
│   └── utils.ts          # ヘルパー関数
│
├── types/                # 型定義
│   └── index.ts          # 共通型定義
│
├── styles/               # グローバルスタイル
│   └── index.css         # Tailwind CSS
│
├── assets/               # 画像等のアセット
│
├── App.tsx               # ルーティング定義
└── main.tsx              # エントリーポイント
```

## 機能一覧

| 機能 | パス | 説明 |
|------|------|------|
| ダッシュボード | `/` | 予約カンバン、本日の予定 |
| 飼い主管理 | `/owners` | 飼い主・ペット情報管理 |
| カルテ管理 | `/medical-records` | 診察記録、治療計画 |
| 予約管理 | `/reservations` | 予約の登録・変更 |
| 入院管理 | `/hospitalization` | 入院・ホテル管理 |
| 検査 | `/examinations` | 検査記録 |
| 会計 | `/accounting` | 精算処理 |
| ワクチン | `/vaccinations` | 予防接種管理 |
| トリミング | `/trimming` | トリミング予約 |
| マスタ設定 | `/settings/*` | 各種マスタ管理 |

## コンポーネント構成

### shadcn/ui コンポーネント (`components/ui/`)

Radix UIベースの再利用可能なコンポーネント:
- Button, Input, Select, Checkbox
- Dialog, Sheet, Popover
- Table, Tabs, Accordion
- Calendar, DatePicker
- Toast (Sonner)
- その他40以上

### 共有コンポーネント (`components/shared/`)

- `PageLayout` - ページレイアウト
- `FormHeader` - フォームヘッダー
- `PatientInfoCard` - 患者情報カード
- `DataTable` - データテーブル
- `PrimaryButton` - プライマリボタン

## 開発環境

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

## ビルド

```bash
# 本番ビルド
npm run build

# 型チェック
npx tsc --noEmit

# Lint
npm run lint
```

## コード規約

### ファイル命名
- コンポーネント: `PascalCase.tsx`（例: `PatientInfoCard.tsx`）
- フック: `use*.ts`（例: `usePets.ts`）
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
import { Button } from '@/components/ui/button';
import { PatientInfoCard } from '../components';
import type { Pet } from '@/types';
```

### Absolute Imports

`@/` エイリアスを使用:
```typescript
import { Button } from '@/components/ui/button';
import { Pet } from '@/types';
```

## トラブルシューティング

### パスエイリアスエラー
`@/` が解決できない場合、`vite.config.ts` と `tsconfig.json` の設定を確認。

### APIエラー
バックエンドが起動しているか確認し、`vite.config.ts` のプロキシ設定を確認。

### ビルドエラー
```bash
# node_modules再インストール
rm -rf node_modules
npm install
```
