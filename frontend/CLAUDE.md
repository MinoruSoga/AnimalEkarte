# Frontend - React/TypeScript

## ⚠️ コマンド実行ルール

**npmコマンドはローカル実行禁止。必ずDocker経由で実行すること。**

```bash
# ❌ NG
npm run build
npm run lint
npm test

# ✅ OK
docker compose exec frontend npm run build
docker compose exec frontend npm run lint
docker compose exec frontend npm run test:run
```

## よく使うコマンド

| タスク | コマンド |
|--------|---------|
| ビルド | `docker compose exec frontend npm run build` |
| Lint | `docker compose exec frontend npm run lint` |
| テスト | `docker compose exec frontend npm run test:run` |
| テスト(watch) | `docker compose exec frontend npm test` |
| カバレッジ | `docker compose exec frontend npm run test:coverage` |

## 技術スタック

- **言語**: TypeScript 5.7
- **フレームワーク**: React 18
- **ビルド**: Vite 6
- **スタイル**: Tailwind CSS 4
- **UI**: shadcn/ui (Radix UI)
- **ルーティング**: React Router 6
- **テスト**: Vitest + Testing Library

## ディレクトリ構造

```
src/
├── components/
│   ├── ui/        # shadcn/ui コンポーネント
│   ├── shared/    # 共有UIコンポーネント
│   └── Sidebar.tsx
├── features/      # 機能別モジュール
│   ├── dashboard/
│   ├── owners/
│   ├── medical-records/
│   ├── reservations/
│   └── ...
├── lib/           # ユーティリティ
├── types/         # 型定義
└── test/          # テスト設定
```

## コーディング規約

### コンポーネント
```typescript
interface Props {
  patient: Patient;
  onSelect?: (id: string) => void;
}

export const PatientCard: FC<Props> = ({ patient, onSelect }) => {
  // 実装
};
```

### 命名規則
| 対象 | 規則 | 例 |
|------|------|-----|
| コンポーネント | PascalCase | `PatientCard` |
| 関数・変数 | camelCase | `getPatientById` |
| 定数 | UPPER_SNAKE | `API_BASE_URL` |
| ファイル | kebab-case | `patient-card.tsx` |

## 禁止事項

- ❌ any型使用
- ❌ 未使用インポート
- ❌ ハードコード値
- ❌ console.log放置

## ESLint

53件のwarningが残存（主にany型、setState in effect）。
エラーは0件を維持すること。
