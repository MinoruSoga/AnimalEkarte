# BUG-239: DailyCareLogDialog — getCurrentTime() がコンポーネント内定義 + useState で毎回呼び出し

## 概要
`DailyCareLogDialog.tsx` で以下の2つの違反が重複している：

1. **`rendering-hoist-jsx` 系**: `getCurrentTime` 関数がコンポーネント内で定義されており、レンダーのたびに新しい関数参照が生成される
2. **`rerender-lazy-state-init`**: `useState({..., time: getCurrentTime()})` でオブジェクトリテラルを直接渡しているため、`getCurrentTime()` がレンダーのたびに実行される（`format(new Date(), "HH:mm")` = Date 生成 + 日時フォーマット）

`useState` は初回レンダーの初期値にしか使わないため、2回目以降のレンダーで `getCurrentTime()` を呼ぶのは純粋な無駄である。

## 現状コード

### `features/hospitalization/components/DailyRecord/DailyCareLogDialog.tsx:29-43`
```typescript
export function DailyCareLogDialog({ open, onOpenChange, type, onSave }: DailyCareLogDialogProps) {
    // ❌ コンポーネント内で関数定義 — レンダーごとに新参照
    const getCurrentTime = () => format(new Date(), "HH:mm");

    // ❌ getCurrentTime() がレンダーのたびに呼ばれる（Date 生成 + format）
    const [form, setForm] = useState({
        value: "",
        notes: "",
        time: getCurrentTime()
    });
    const [prevOpen, setPrevOpen] = useState(false);

    // open が変わった時にフォームをリセット（derived state パターン — 許容範囲）
    if (open !== prevOpen) {
        setPrevOpen(open);
        if (open) {
            setForm({ value: "", notes: "", time: getCurrentTime() });
        }
    }
```

## 修正方針

```typescript
// ✅ モジュールスコープに巻き上げ — レンダーごとの再生成なし
function getCurrentTime(): string {
  return format(new Date(), "HH:mm");
}

export function DailyCareLogDialog({ open, onOpenChange, type, onSave }: DailyCareLogDialogProps) {
    // ✅ lazy initializer — 初回マウント時のみ getCurrentTime() 実行
    const [form, setForm] = useState(() => ({
        value: "",
        notes: "",
        time: getCurrentTime()
    }));
    const [prevOpen, setPrevOpen] = useState(false);

    // open が変わった時のリセット — getCurrentTime() は open 時のみ呼ぶ（これは正当）
    if (open !== prevOpen) {
        setPrevOpen(open);
        if (open) {
            setForm({ value: "", notes: "", time: getCurrentTime() });
        }
    }
```

**注意**: `if (open !== prevOpen)` による derived state sync パターンは React の推奨手法（getDerivedStateFromProps の関数コンポーネント版）であり変更不要。

## 準拠すべきプロジェクト規約

### `.claude/rules/typescript-react.md` — rerender-lazy-state-init
> 高コストな `useState` 初期化は `useState(() => ...)` lazy 形式を使用

### `frontend/CODING_RULES.md` Section 12 — rendering-hoist-jsx（関連）
> コンポーネント外の静的関数はモジュールスコープに定義する（毎レンダーでの再生成を避ける）

### プロジェクト内参照実装
`features/estimates/hooks/use-estimate-form.ts:51` — lazy initializer パターン

## 優先度
**Medium** — `format(new Date(), "HH:mm")` は date-fns の書式化処理であり、計算コストは軽微ではない。ダイアログが頻繁に開閉される画面では積み重なる。修正は5分。

## 関連チケット
- BUG-227: 静的 SelectItem JSX のモジュール定数巻き上げ（同種の巻き上げパターン）

## 関連ファイル
- `frontend/src/features/hospitalization/components/DailyRecord/DailyCareLogDialog.tsx:29-43`
