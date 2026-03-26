# TASK-037: FE StaffSettings.tsx の setState in useEffect 修正

**作成日**: 2026-03-27
**ステータス**: Open
**優先度**: 中
**領域**: Frontend

---

## 概要

ESLint の `react-hooks/set-state-in-effect` 警告が検出されている。
`StaffSettings.tsx:130` で `useEffect` 内から `setState` を同期呼出しており、
カスケードレンダリングが発生するリスクがある。

---

## エラー詳細

```
/app/src/features/master/routes/StaffSettings.tsx:130:7
  warning  Calling setState synchronously within an effect can trigger cascading renders
  react-hooks/set-state-in-effect

  128 |   useEffect(() => {
  129 |     if (userDetail && !groupIdsInitialized.current) {
> 130 |       setGroupIds(userDetail.permission_group_ids);
      |       ^^^^^^^^^^^ Avoid calling setState() directly within an effect
  131 |       groupIdsInitialized.current = true;
  132 |     }
  133 |   }, [userDetail]);
```

---

## 現状の意図

`userDetail`（API レスポンス）が取得できたタイミングで、チェックボックスの初期値 `groupIds` を設定している。
`groupIdsInitialized.current` で「初回のみ」ガードしている。

---

## 修正方針

`useState` の lazy initializer + `useMemo` パターンに変更する。

```tsx
// Before
const [groupIds, setGroupIds] = useState<number[]>([]);
const groupIdsInitialized = useRef(false);

useEffect(() => {
  if (userDetail && !groupIdsInitialized.current) {
    setGroupIds(userDetail.permission_group_ids);
    groupIdsInitialized.current = true;
  }
}, [userDetail]);

// After: userDetail が変わったら key で再マウントしてlazy initで解決
// または useMemo で派生値として管理
const [editedGroupIds, setEditedGroupIds] = useState<number[]>(
  () => userDetail?.permission_group_ids ?? []
);
// userDetail が変わる（別ユーザー選択）場合は key={userId} でコンポーネントをリセット
```

パターン選定は実装者に委ねるが、`useEffect` + `setState` の組み合わせは排除すること。

---

## 受入条件

- [ ] `StaffSettings.tsx` から `react-hooks/set-state-in-effect` 警告が消えている
- [ ] `docker compose exec frontend npm run lint` 警告が 6 件（shadcn/ui 由来のみ）に減少
- [ ] 権限グループ編集のユーザー切替時に、チェックボックス初期値が正しく切り替わること（機能デグレなし）
