# FE-026: useMasterCRUD<T> カスタム hook 作成

**Status**: Open
**Priority**: High
**Affects**: master feature — 全マスタ設定ページ共通
**Date Created**: 2026-03-17
**Related**: TASK-007, FE-028, FE-029

## Summary

マスタ設定ページ11ページで重複している CRUD state・ハンドラ・検索フィルタ・SidePeek制御を1つのジェネリック hook `useMasterCRUD<T>` に抽出する。推定 ~880行の重複を削減。

## 重複パターン（現在11ページで繰り返し）

```typescript
// state 宣言（全ページ同一）
const [editTarget, setEditTarget] = useState<T | "new" | null>(null);
const [searchTerm, setSearchTerm] = useState("");
const [pendingDelete, setPendingDelete] = useState<T | null>(null);
const [, startSaveTransition] = useTransition();

// 検索フィルタ（全ページ同一）
const deferredSearch = useDeferredValue(searchTerm);
const filteredItems = useMemo(() => { ... }, [data, deferredSearch]);

// ハンドラ（全ページ同一構造）
const handleClose = useCallback(() => setEditTarget(null), []);
const handleSave = useCallback((data) => { startSaveTransition(() => { ... }); }, [...]);
const handleDeleteConfirm = useCallback(() => { ... }, [...]);

// 導出値
const panelItem = editTarget !== null && editTarget !== "new" ? editTarget : null;
```

## 必要な変更

### 新規ファイル: `frontend/src/features/master/hooks/use-master-crud.ts`

```typescript
interface UseMasterCRUDOptions<T, CreateReq, UpdateReq> {
  // データ取得
  data: T[] | undefined;

  // Mutation hooks（各ページが API hook を渡す）
  createMutation: UseMutationResult<T, Error, CreateReq>;
  updateMutation: UseMutationResult<T, Error, { id: string; req: UpdateReq }>;
  deleteMutation: UseMutationResult<void, Error, string>;

  // ラベル（toast メッセージ用）
  entityLabel: string;  // "ケージ", "スタッフ" 等

  // 検索フィルタ（name フィールドデフォルト、カスタマイズ可）
  searchFilter?: (item: T, term: string) => boolean;
}

interface UseMasterCRUDReturn<T> {
  // State
  editTarget: T | "new" | null;
  setEditTarget: (target: T | "new" | null) => void;
  searchTerm: string;
  setSearchTerm: (term: string) => void;
  pendingDelete: T | null;
  setPendingDelete: (item: T | null) => void;
  isSavePending: boolean;

  // 導出値
  filteredItems: T[];
  panelItem: T | null;
  isEditing: boolean;  // editTarget !== null

  // ハンドラ
  handleClose: () => void;
  handleNew: () => void;
  handleEdit: (item: T) => void;
  handleSave: (createReq: CreateReq, updateReq?: UpdateReq) => void;
  handleDeleteRequest: (item: T) => void;
  handleDeleteConfirm: () => void;
  handleDeleteCancel: () => void;
}
```

### Vercel Best Practices 準拠ポイント

- `useTransition` で save pending 管理（`useState(false)` + `setIsPending` 禁止）
- `useDeferredValue` で検索フィルタ遅延
- `useCallback` で全ハンドラ安定化
- `useMemo` でフィルタ結果キャッシュ
- ジェネリック型で型安全性確保

## 使用例（移行後の各ページ）

```typescript
// CageSettings.tsx（移行後）
const crud = useMasterCRUD({
  data: cages,
  createMutation: useCreateCage(),
  updateMutation: useUpdateCage(),
  deleteMutation: useDeleteCage(),
  entityLabel: "ケージ",
});

// crud.editTarget, crud.filteredItems, crud.handleSave 等を使用
// ~80行の state/ハンドラ定義が1行の hook 呼び出しに
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし（ジェネリック T で型安全）
- [ ] `FC` / `forwardRef` なし
- [ ] `useTransition` で pending 管理
- [ ] `useDeferredValue` で検索遅延
- [ ] `useCallback` で全ハンドラ安定化
- [ ] `useMemo` でフィルタ結果キャッシュ

## 完了条件

- [ ] `use-master-crud.ts` 新規作成
- [ ] ジェネリック型パラメータで型安全
- [ ] CageSettings.tsx で動作確認（1ページ先行適用）
- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend pnpm lint` パス）
