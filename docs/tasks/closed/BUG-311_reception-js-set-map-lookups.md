# BUG-311: Reception.tsx - O(n²) ネスト検索を Map で O(1) に最適化

## 概要

`frontend/src/features/reception/routes/Reception.tsx` にて、列（column）とアポイントメント（appointment）の
ネスト検索に `Array.find + Array.some` の O(n²) パターンが3箇所使われている。

特にレンダーパス（render path）上の検索は毎レンダーで実行されるため、改善の優先度が高い。

## 違反ルール

`js-set-map-lookups` — O(n) 以上のルックアップは Set/Map による O(1) 化が必要

## 違反箇所

### 1. レンダーパス（最優先・毎レンダー実行）

**File:** `frontend/src/features/reception/routes/Reception.tsx:410`

```typescript
// 現状: filteredColumns を毎レンダーで find+some（O(n²)）
currentStatus={selectedAppointment
  ? filteredColumns.find(c => c.appointments.some(a => a.id === selectedAppointment.id))?.title
  : undefined}
```

**修正案:**
```typescript
// useMemo で selectedAppointment.id をキーにして O(1) ルックアップ
const selectedAppointmentColumnTitle = useMemo(() => {
  if (!selectedAppointment) return undefined;
  for (const col of filteredColumns) {
    for (const apt of col.appointments) {
      if (apt.id === selectedAppointment.id) return col.title;
    }
  }
  return undefined;
}, [filteredColumns, selectedAppointment?.id]);

// 使用箇所
currentStatus={selectedAppointmentColumnTitle}
```

または、より汎用的な Map を構築する場合:
```typescript
const appointmentColumnMap = useMemo(() => {
  const map = new Map<string, string>();
  for (const col of filteredColumns) {
    for (const apt of col.appointments) {
      map.set(apt.id, col.title);
    }
  }
  return map;
}, [filteredColumns]);

currentStatus={selectedAppointment ? appointmentColumnMap.get(selectedAppointment.id) : undefined}
```

---

### 2. handleDragOver イベントハンドラ（Line 106）

```typescript
// columnsRef.current に対して find+some
const sourceColumn = cols.find(col => col.appointments.some(a => a.id === activeId));
```

イベントハンドラ内なので毎フレーム呼ばれることはないが、
`handleDragOver` は頻繁に発火するため改善が望ましい。

**修正案:**
```typescript
// DnD の activeId が変わったときだけ再構築するよう useMemo/useRef で持つか、
// あるいは columnsRef を走査する際に早期 break する実装に変える
// 現状のコードは既に ref 経由なので再レンダー最適化は済んでいる。
// イベントハンドラ内なので Map 構築のオーバーヘッドが find+some より
// 大きくなる可能性があるため、影響度は低い。
```

---

### 3. handleDragEnd イベントハンドラ（Line 147）

Line 106 と同じパターン（`handleDragEnd` の先頭）。
同様の判断で影響度は低い。

---

## 優先度

- **Line 410（レンダーパス）**: HIGH — 毎レンダー実行されるため早期対応を推奨
- **Line 106, 147（イベントハンドラ）**: LOW — ドラッグ操作中のみ実行、現状で実用上の問題は小さい

## 影響範囲

- `frontend/src/features/reception/routes/Reception.tsx`

## 関連ルール

- Vercel React Best Practices: `js-set-map-lookups`
- `.claude/rules/performance-rules.md`
