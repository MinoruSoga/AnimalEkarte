# Animal Ekarte Frontend

動物病院 電子カルテシステムのフロントエンド

## 技術スタック

| 技術           | バージョン    |
| -------------- | ------------- |
| React          | 19            |
| TypeScript     | 5.7           |
| Vite           | 6             |
| Tailwind CSS   | 4             |
| shadcn/ui      | Radix UI      |
| React Router   | 7 (Data Mode) |
| TanStack Query | v5            |
| Axios          | HTTP Client   |
| Vitest         | テスト        |
| Zustand        | 状態管理      |

---

## ⚠️ コマンド実行ルール

**pnpmコマンドはローカル実行禁止。必ずDocker経由で実行する。**

```bash
# ❌ NG — ローカル実行禁止
pnpm dev
pnpm build

# ✅ OK — Docker経由
docker compose exec frontend pnpm build
docker compose exec frontend pnpm lint
docker compose exec frontend pnpm test:run
```

| タスク         | コマンド                                     |
| -------------- | -------------------------------------------- |
| ビルド         | `docker compose exec frontend pnpm build`    |
| Lint           | `docker compose exec frontend pnpm lint`     |
| テスト         | `docker compose exec frontend pnpm test:run` |
| テスト (watch) | `docker compose exec frontend pnpm test`     |

---

## ディレクトリ構成（bulletproof-react準拠・プロジェクト固有拡張あり）

```
frontend/src/
├── main.tsx                # Viteエントリーポイント
│
├── app/                    # アプリケーション層
│   ├── index.tsx           # Appコンポーネント（AppProvider → RouterProvider）
│   ├── provider.tsx        # QueryClientProvider + Toaster（AuthProvider はここに置かない）
│   ├── router.tsx          # createBrowserRouter（inline lazy パターン）
│   └── pages/              # ★ cross-feature合成ページ（複数featureが必要な場合のみ）
│       └── OwnerFormPage.tsx   # owners + pets を合成する例
│
├── assets/                 # 静的アセット（画像・アイコン等）
│
├── components/             # 共有コンポーネント
│   ├── ui/                 # shadcn/ui（変更禁止）
│   ├── errors/             # RouteErrorBoundary, RootErrorBoundary
│   └── shared/             # アプリ固有共有UI（Layout含む）
│
├── features/               # 機能別モジュール（16 features）
│   ├── auth/               # ★ 認証（ログイン・セッション管理）
│   ├── reception/          # 当日の受付（カンバンボード）
│   ├── owners/             # ★ ベストプラクティス参照実装
│   ├── pets/               # ペット（CRUD API のみ）
│   ├── reservations/       # 予約管理
│   ├── medical-records/    # 電子カルテ
│   ├── hospitalization/    # 入院管理
│   ├── examinations/       # 診察
│   ├── accounting/         # 会計
│   ├── vaccinations/       # ワクチン
│   ├── trimming/           # トリミング
│   ├── inventory/          # 在庫管理
│   ├── estimates/          # 見積
│   ├── shifts/             # シフト管理
│   ├── master/             # マスタ設定（PATTERNS.md 参照）
│   └── hospital-settings/  # 病院設定（クリニックマスタ）
│
│   └── [feature-name]/     # 各featureの構造
│       ├── api/            # フェッチ関数 + React Query hooks
│       │   ├── get-xxx.ts      # getXxx() 生関数 + useGetXxx() hook
│       │   ├── create-xxx.ts   # createXxx() 生関数（複雑フォームは hook なし）
│       │   ├── types.ts        # APIリクエスト/レスポンス型（models.tsから導出）
│       │   ├── transforms.ts   # BackendXxx → Xxx 型変換
│       │   └── index.ts        # 明示的named export
│       ├── components/     # feature 固有 UI
│       ├── hooks/          # useXxxForm, useXxxFilters 等
│       ├── routes/         # ページコンポーネント（★ app/routes/ でなくここ）
│       ├── types/          # feature固有型（UI専用・手書きOK）
│       ├── loaders.ts      # React Router loader（必要な場合のみ）
│       └── index.ts        # 公開 API（明示的 named export のみ）
│
├── hooks/                  # 共有hooks（全feature横断）
├── lib/                    # ライブラリ設定（axios.ts, react-query.ts, utils.ts 等）
├── config/                 # paths.ts（型安全URLマップ）
├── stores/                 # Zustand（sidebar状態のみ）
├── types/
│   ├── generated/          # ★ 自動生成（直接編集禁止）
│   │   └── models.ts       # make codegen（tygo）で生成
│   └── ...                 # 共有ドメイン型
├── utils/                  # 純粋ユーティリティ（format/, constants/ 等）
└── testing/                # テスト設定（MSW）
```

---

## 機能一覧

| 機能             | パス               | 説明                           |
| ---------------- | ------------------ | ------------------------------ |
| ダッシュボード   | `/`                | 予約カンバン、本日の予定       |
| 飼主・ペット管理 | `/owners`          | 飼主・ペット情報管理 ★参照実装 |
| カルテ管理       | `/medical-records` | 診察記録、治療計画             |
| 予約管理         | `/reservations`    | 予約の登録・変更               |
| 入院管理         | `/hospitalization` | 入院・ホテル管理               |
| 診察             | `/examinations`    | 検査記録                       |
| 会計             | `/accounting`      | 精算処理                       |
| ワクチン         | `/vaccinations`    | 予防接種管理                   |
| トリミング       | `/trimming`        | トリミング予約                 |
| 在庫管理         | `/inventory`       | 薬剤・物品在庫                 |
| 見積             | `/estimates`       | 見積作成・管理                 |
| シフト管理       | `/shifts`          | スタッフシフトカレンダー       |
| マスタ設定       | `/settings/*`      | 各種マスタ管理                 |
| 病院設定         | `/settings/clinic` | クリニック基本情報             |

---

## ★ ベストプラクティス参照実装: `features/owners/`

`features/owners/` はすべての Vercel React Best Practices パターンを実装した参照実装。
新機能を実装する際はこの feature を手本にすること。

| パターン                                  | 実装ファイル                  | 内容                                                       |
| ----------------------------------------- | ----------------------------- | ---------------------------------------------------------- |
| `memo()` によるセクション分割             | `routes/OwnerForm.tsx`        | `OwnerInfoSection`, `PetTableRow`, `MembershipTypeButtons` |
| `useCallback` によるハンドラ安定化        | `routes/OwnerForm.tsx`        | `handleInputChange`, `handleDeletePetRequest`              |
| `lazy()` + `Suspense` の遅延ロード        | `routes/OwnerForm.tsx`        | `PetEditModal`                                             |
| 静的 JSX のモジュール定数化               | `routes/OwnerForm.tsx`        | `PET_TABLE_HEADER`                                         |
| `useDeferredValue` による検索遅延         | `routes/OwnersList.tsx`       | `deferredSearchTerm`                                       |
| `useActionState` による送信・pending 管理 | `hooks/use-owner-form.ts`     | `formAction`, `isPending`                                  |
| `useState(() => ...)` lazy init           | `hooks/use-owner-form.ts`     | `mapOwnerToFormData`                                       |
| API 由来 JSX の `useMemo` キャッシュ      | `components/PetEditModal.tsx` | `animalSpeciesSelectItems`                                 |
| `Promise.all` による並列フェッチ          | `loaders.ts`                  | `ownersLoader`                                             |
| `? (...) : null` 条件レンダー             | `routes/OwnersList.tsx`       | `pet.status ?`                                             |
| 直接ファイル import                       | `routes/OwnersList.tsx`       | `../api/delete-owner`                                      |
| cross-feature props 注入                  | `app/pages/OwnerFormPage.tsx` | `petMutations` を `OwnerForm` に注入                       |

詳細: `frontend/CODING_RULES.md` Section 12 参照。

---

## コード規約

### ファイル命名

| 対象           | 規則                                             | 例                                         |
| -------------- | ------------------------------------------------ | ------------------------------------------ |
| コンポーネント | PascalCase.tsx                                   | `OwnerForm.tsx`                            |
| hooks          | kebab-case.ts（ファイル）／ `useXxx`（シンボル） | `use-owner-form.ts` exports `useOwnerForm` |
| API ファイル   | kebab-case.ts                                    | `get-owners.ts`, `create-owner.ts`         |
| ユーティリティ | kebab-case.ts                                    | `format-date.ts`                           |
| ディレクトリ   | kebab-case                                       | `medical-records/`                         |

### React 19 禁止事項

| 禁止                                                           | 代替                                         |
| -------------------------------------------------------------- | -------------------------------------------- |
| `FC` / `React.FC`                                              | 関数宣言                                     |
| `forwardRef`                                                   | ref as prop                                  |
| `any` 型                                                       | `unknown` + 型ガード                         |
| feature 間 import                                              | app/pages/ で合成（props 注入）              |
| `export *`                                                     | 明示的 named export                          |
| `&&` 条件レンダー                                              | `? (...) : null`                             |
| feature 外からの deep import（`@/features/xxx/routes/...` 等） | feature の index.ts 経由（`@/features/xxx`） |
| `useOwners` 等（動詞省略）                                     | `useGetOwners`（動詞 + エンティティ）        |
| `localStorage` に token 保存                                   | httpOnly Cookie + `withCredentials: true`    |

### import 順序

```typescript
// 1. React / Framework
import { useState, memo, useCallback, lazy, Suspense } from "react";
import { useNavigate } from "react-router";

// 2. 外部ライブラリ
import { toast } from "sonner";

// 3. 共有モジュール (@/)
import { Button } from "@/components/ui/button";
import { formatDate } from "@/utils/format/date"; // barrel 経由不可、直接 import

// 4. feature 内部（相対パス）
import { useOwnerForm } from "../hooks/useOwnerForm";

// 5. 型（type keyword 付き）
import type { Owner } from "@/types/owner";
```

---

## ローカル DEV / STG デモアカウント

ログイン画面のデモアカウント一覧は次だけで表示する。

- ローカル Vite DEV（`import.meta.env.DEV`）
- Vercel preview（STG。`frontend-deploy.yml` が `VERCEL_ENV=preview` を焼き込む）

本番（`VERCEL_ENV=production`）では出さない。ワンクリック入力のパスワードは
全デモ共通の `password`（コード固定。production の API は受け付けない）。

STG の API は `https://api.stg.noah-karte.com/api`（CSP `connect-src` と一致）。
`vite.config.ts` が `VERCEL_ENV=preview` のとき `VITE_API_URL` を define する。
本番ビルドは `https://api.noah-karte.com/api`。ローカルは `VITE_API_URL=/api`。

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

| ドキュメント                                                       | 説明                                             |
| ------------------------------------------------------------------ | ------------------------------------------------ |
| [コーディング規約](./CODING_RULES.md)                              | 詳細なルール（Section 12: パフォーマンス最適化） |
| [Claude 設定](./CLAUDE.md)                                         | 実装パターン・禁止事項                           |
| [React 19](https://react.dev/blog/2024/12/05/react-19)             | 公式リリースノート                               |
| [bulletproof-react](https://github.com/alan2207/bulletproof-react) | アーキテクチャ参照                               |
