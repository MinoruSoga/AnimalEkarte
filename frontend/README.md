# Animal Ekarte Frontend

動物病院 電子カルテシステムのフロントエンド

## 技術スタック

| 技術 | バージョン |
|------|-----------|
| React | 19 |
| TypeScript | 5.7 |
| Vite | 6 |
| Tailwind CSS | 4 |
| shadcn/ui | Radix UI |
| React Router | 7 (Data Mode) |
| TanStack Query | v5 |
| Axios | HTTP Client |
| Vitest | テスト |

---

## ⚠️ コマンド実行ルール

**npmコマンドはローカル実行禁止。必ずDocker経由で実行する。**

```bash
# ❌ NG — ローカル実行禁止
npm run dev
npm run build

# ✅ OK — Docker経由
docker compose exec frontend npm run build
docker compose exec frontend npm run lint
docker compose exec frontend npm run test:run
```

| タスク | コマンド |
|--------|---------|
| ビルド | `docker compose exec frontend npm run build` |
| Lint | `docker compose exec frontend npm run lint` |
| テスト | `docker compose exec frontend npm run test:run` |
| テスト (watch) | `docker compose exec frontend npm test` |

---

## ディレクトリ構成（bulletproof-react）

```
frontend/src/
├── main.tsx                # Viteエントリーポイント
│
├── app/                    # アプリケーション層
│   ├── index.tsx           # Appコンポーネント
│   ├── provider.tsx        # プロバイダー統合（QueryClient, Toaster）
│   └── router.tsx          # createBrowserRouter（Data Mode）
│
├── components/             # 共有コンポーネント
│   ├── ui/                 # shadcn/ui（変更禁止）
│   └── shared/             # アプリ固有の共有UI
│
├── features/               # 機能別モジュール（bulletproof-react）
│   ├── owners/             # ★ ベストプラクティス参照実装
│   ├── dashboard/
│   ├── medical-records/
│   ├── reservations/
│   ├── hospitalization/
│   ├── examinations/
│   ├── accounting/
│   ├── vaccinations/
│   ├── trimming/
│   ├── master/
│   └── clinic/
│
├── hooks/                  # 共有hooks
├── lib/                    # ライブラリ設定（axios, react-query）
├── types/                  # 共有型定義（generated/models.ts 含む）
├── utils/                  # ユーティリティ（format, validation）
└── testing/                # テスト設定（MSW）
```

各 feature の内部構造:

```
features/[feature]/
├── api/          # React Query hooks（useQuery, useMutation）
├── components/   # feature 固有 UI
├── hooks/        # ビジネスロジック・UI 状態
├── routes/       # ページコンポーネント
├── types/        # ドメイン型定義
├── loaders.ts    # React Router Data Mode ローダー
└── index.ts      # 公開 API（明示的 named export のみ）
```

---

## 機能一覧

| 機能 | パス | 説明 |
|------|------|------|
| ダッシュボード | `/` | 予約カンバン、本日の予定 |
| 飼主・ペット管理 | `/owners` | 飼主・ペット情報管理 ★参照実装 |
| カルテ管理 | `/medical-records` | 診察記録、治療計画 |
| 予約管理 | `/reservations` | 予約の登録・変更 |
| 入院管理 | `/hospitalization` | 入院・ホテル管理 |
| 検査 | `/examinations` | 検査記録 |
| 会計 | `/accounting` | 精算処理 |
| ワクチン | `/vaccinations` | 予防接種管理 |
| トリミング | `/trimming` | トリミング予約 |
| マスタ設定 | `/settings/*` | 各種マスタ管理 |

---

## ★ ベストプラクティス参照実装: `features/owners/`

`features/owners/` はすべての Vercel React Best Practices パターンを実装した参照実装。
新機能を実装する際はこの feature を手本にすること。

| パターン | 実装ファイル | 内容 |
|---------|------------|------|
| `memo()` によるセクション分割 | `routes/OwnerForm.tsx` | `OwnerInfoSection`, `PetTableRow`, `MembershipTypeButtons` |
| `useCallback` によるハンドラ安定化 | `routes/OwnerForm.tsx` | `handleInputChange`, `handleDeletePetRequest` |
| `lazy()` + `Suspense` の遅延ロード | `routes/OwnerForm.tsx` | `PetEditModal` |
| 静的 JSX のモジュール定数化 | `routes/OwnerForm.tsx` | `PET_TABLE_HEADER` |
| `useDeferredValue` による検索遅延 | `routes/OwnersList.tsx` | `deferredSearchTerm` |
| `useTransition` による pending 管理 | `hooks/useOwnerForm.ts` | `startSaveTransition` |
| `useState(() => ...)` lazy init | `hooks/useOwnerForm.ts` | `mapOwnerToFormData` |
| API 由来 JSX の `useMemo` キャッシュ | `components/PetEditModal.tsx` | `animalSpeciesSelectItems` |
| `Promise.all` による並列フェッチ | `loaders.ts` | `ownersLoader` |
| `? (...) : null` 条件レンダー | `routes/OwnersList.tsx` | `pet.status ?` |
| 直接ファイル import | `routes/OwnersList.tsx` | `../api/delete-owner` |

詳細: `frontend/CODING_RULES.md` Section 12 参照。

---

## コード規約

### ファイル命名

| 対象 | 規則 | 例 |
|------|------|-----|
| コンポーネント | PascalCase.tsx | `OwnerForm.tsx` |
| hooks | use-xxx.ts / useXxx.ts | `useOwnerForm.ts` |
| ユーティリティ | kebab-case.ts | `delete-owner.ts` |
| ディレクトリ | kebab-case | `medical-records/` |

### React 19 禁止事項

| 禁止 | 代替 |
|------|------|
| `FC` / `React.FC` | 関数宣言 |
| `forwardRef` | ref as prop |
| `any` 型 | `unknown` + 型ガード |
| feature 間 import | app 層で合成 |
| `export *` | 明示的 named export |
| `&&` 条件レンダー | `? (...) : null` |
| barrel index 経由 import | 直接ファイル import |

### import 順序

```typescript
// 1. React / Framework
import { useState, memo, useCallback, lazy, Suspense } from "react";
import { useNavigate } from "react-router";

// 2. 外部ライブラリ
import { toast } from "sonner";

// 3. 共有モジュール (@/)
import { Button } from "@/components/ui/button";
import { formatDate } from "@/utils/format/date";  // barrel 経由不可、直接 import

// 4. feature 内部（相対パス）
import { useOwnerForm } from "../hooks/useOwnerForm";

// 5. 型（type keyword 付き）
import type { Owner } from "@/types/owner";
```

---

## トラブルシューティング

### パスエイリアスエラー

`@/` が解決できない場合、`vite.config.ts` と `tsconfig.json` の設定を確認。

### API エラー

バックエンドが起動しているか確認:

```bash
make logs
# または
docker compose ps
```

### ビルドエラー

```bash
# キャッシュクリア＆再ビルド
make clean
```

---

## 参照

| ドキュメント | 説明 |
|-------------|------|
| [コーディング規約](./CODING_RULES.md) | 詳細なルール（Section 12: パフォーマンス最適化） |
| [Claude 設定](./CLAUDE.md) | 実装パターン・禁止事項 |
| [React 19](https://react.dev/blog/2024/12/05/react-19) | 公式リリースノート |
| [bulletproof-react](https://github.com/alan2207/bulletproof-react) | アーキテクチャ参照 |
