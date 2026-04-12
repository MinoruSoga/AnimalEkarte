# BUG-328: 検査編集フォームで検査種別・担当医 Select が「選択してください」を表示

**Status**: OPEN  
**Priority**: Medium  
**Discovery**: 機能テスト Section 5 検査管理 (2026-04-12)

## 概要

`/examinations/:id`（編集モード）を開くと、検査種別・担当医の Select コンポーネントが「選択してください」プレースホルダーを表示し、API が返した値を反映しない。API は `exam_type_id: 1, doctor_id: 1` を正しく返しており、transform も `testTypeId: "1", doctorId: "1"` を生成しているため、Radix UI Select のタイミング問題が疑われる。

## 再現手順

1. `/examinations/3` にアクセス（編集モード）
2. ページロード完了を待つ
3. **結果**: 「検査種別」「担当医」Select が「選択してください」を表示
4. **期待**: API から返った値 (検査種別 id=1, 担当医 id=1) が選択済みとして表示される

## 現状コード

### `frontend/src/features/examinations/routes/ExaminationForm.tsx:100`
```tsx
<Select value={formData.testTypeId ?? ""}>
  <SelectTrigger>
    <SelectValue placeholder="選択してください" />
  </SelectTrigger>
  <SelectContent>
    {examTypeSelectItems}  {/* useMemo — examTypes が空配列の間は0件 */}
  </SelectContent>
</Select>
```

### `frontend/src/features/examinations/hooks/use-examination-form.ts:53-56`
```ts
const formData: Partial<ExaminationRecord> =
  isEdit && existingExam
    ? { ...existingExam, ...localOverrides }
    : { status: "依頼中", ... };
```

`existingExam` ロード前は `formData.testTypeId = undefined`（デフォルト分岐）。  
`existingExam` ロード後は `formData.testTypeId = "1"` — この時点で `examTypes` が空なら Select は一致する SelectItem を見つけられない。

### `frontend/src/features/examinations/api/transforms.ts:35`
```ts
testTypeId: String(data.exam_type_id ?? ""),  // "1" — 変換は正しい
doctorId: String(data.doctor_id ?? ""),        // "1" — 変換は正しい
```

## 疑われる原因

Radix UI Select の動作: `SelectContent` が閉じた状態（dropdown unopened）のとき、`SelectValue` は内部 Context 経由で現在の選択ラベルを取得する。`SelectItem` が後から追加された場合（`examTypes` が非同期ロードで遅延）、Context の更新が `SelectValue` の再レンダーをトリガーしない可能性がある。

競合シナリオ:
1. `existingExam` ロード完了 → `formData.testTypeId = "1"` が確定
2. この時点で `examTypes = []`（まだロード中）→ Select は値を解決できず placeholder 表示
3. `examTypes` ロード完了 → `FormFieldsSection` 再レンダー（memo が `examTypes` prop 変更を検知）
4. Select が `value="1"` + `SelectItem value="1"` で再描画 → **ここで正しく表示されるか不確定**

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/features/examinations/routes/ExaminationForm.tsx:99-113` | 検査種別 Select 表示 | ❌ NG |
| `frontend/src/features/examinations/routes/ExaminationForm.tsx:117-131` | 担当医 Select 表示 | ❌ NG |

## 修正方針

### 方針 1: `examTypes` ロード完了まで Select を非表示にする

```tsx
// FormFieldsSection に examTypesLoaded prop を追加
{examTypesLoaded ? (
  <Select value={formData.testTypeId ?? ""}>...
) : (
  <div className="h-10 bg-muted rounded-md animate-pulse" />
)}
```

### 方針 2: Select に `key` を付けて examTypes ロード後に再マウント

```tsx
<Select key={`exam-type-${examTypes.length > 0 ? "loaded" : "loading"}`}
  value={formData.testTypeId ?? ""}>
```

examTypes が空→非空に変わった時に Select を再マウントして Radix 内部状態をリセットする。副作用: ドロップダウンが開いていた場合に閉じる（許容範囲）。

### 方針 3: useMasterItems の isLoading を利用して両データ揃うまで待つ

```tsx
const { data: examTypesRaw, isLoading: examTypesLoading } = useMasterItems("examination");
// FormFieldsSection に isLoading フラグを渡し、スケルトン表示
```

**推奨**: 方針 3（明示的な loading 状態管理）。

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — 条件レンダリング
> 条件レンダーは必ず `? (...) : null`（`&&` 禁止）

スケルトン表示にも同ルール適用。

### プロジェクト内参照実装
- `frontend/src/features/owners/routes/OwnerForm.tsx` — `isLoading` で `Select` のスケルトン表示パターン確認

## 優先度

**Medium** — 保存時に検査種別未入力バリデーションが通らないため操作不能になる可能性がある（ユーザーが手動で再選択すれば回避可能）。

## 関連チケット
- BUG-327: カルテ「検査」タブの「検査取り込み」ボタンが未実装

## 関連ファイル
- `frontend/src/features/examinations/routes/ExaminationForm.tsx:62-197` — FormFieldsSection (memo)
- `frontend/src/features/examinations/hooks/use-examination-form.ts:53-56` — formData 計算
- `frontend/src/features/examinations/api/transforms.ts:35-37` — transform（問題なし）
- `frontend/src/features/master/hooks/use-master-items.ts:15` — useMasterItems
