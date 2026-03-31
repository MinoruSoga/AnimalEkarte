# [master] PropertyRow のホバー背景を削除し、SelectTrigger のみホバーにする（全マスタページ対応）

## 優先度
低

## 種別
UIバグ / UX改善

## 対象ファイル
- `frontend/src/components/shared/SidePeek/PropertyRow.tsx`
- `frontend/src/lib/design-tokens.ts`（`STYLE.selectCompact`）
- `frontend/src/features/master/routes/MedicineSettings.tsx`（`SELECT_TRIGGER_FULL`）

---

## 問題

サイドピーク内の `PropertyRow` では、行全体（ラベル + 値エリア）に `hover:bg-[rgba(55,53,47,0.04)]` が適用されている。
その結果、**ラベル部分にホバーしただけで行全体がハイライトされる**。

ユーザーが期待する挙動は「セレクトボックス（SelectTrigger）にホバーしたときのみハイライト」であり、ラベルにはホバーフィードバックが不要。

---

## 現状の挙動

```
PropertyRow 外側 div
├─ [ラベル div] ← ここにホバーしても行全体が光る（不要）
└─ [値エリア div]
     └─ SelectTrigger (SELECT_TRIGGER_FULL / STYLE.selectCompact)
          └─ hover:bg も持つ → 二重ハイライトになる（master-020 で報告済み）
```

### 関係するクラス

| 箇所 | ホバークラス |
|------|-------------|
| `PropertyRow` 外側 div | `hover:bg-[rgba(55,53,47,0.04)]` ← **削除対象** |
| `SELECT_TRIGGER_FULL`（MedicineSettings） | `hover:bg-[rgba(55,53,47,0.04)]` ← **維持** |
| `STYLE.selectCompact`（他マスタページ共通） | `hover:bg-[rgba(55,53,47,0.04)]` ← **維持** |

---

## master-020 との関係

master-020 では「SelectTrigger のホバーを削除して PropertyRow のホバーを残す」方針を提案していたが、
**本イシューの要件と逆方向**である。

正しい修正方針は以下の通り：

| 対象 | master-020 の提案 | 本イシューの正しい方針 |
|------|-----------------|----------------------|
| `PropertyRow` ホバー | 維持 | **削除** |
| `SelectTrigger` ホバー | 削除 | **維持** |

master-020 は**本イシューの修正で自動的に解消**される（PropertyRow ホバーを削除すれば二重ハイライトも消える）。

---

## 修正方針

### 1. `PropertyRow.tsx` からホバーを削除

```diff
// frontend/src/components/shared/SidePeek/PropertyRow.tsx
 <div
-  className={`flex gap-2 py-2 px-2 -mx-2 rounded-[3px] ${C.hoverBgLight} transition-colors min-h-[40px]`}
+  className={`flex gap-2 py-2 px-2 -mx-2 rounded-[3px] transition-colors min-h-[40px]`}
 >
```

### 2. `SELECT_TRIGGER_FULL`・`STYLE.selectCompact` はそのまま維持

MedicineSettings の `SELECT_TRIGGER_FULL` および `STYLE.selectCompact` はすでに `${C.hoverBgLight}` を持つため、**変更不要**。SelectTrigger 単体のホバーは維持される。

---

## 影響範囲

`PropertyRow` は `MasterSidePanel` を使うすべてのマスタページで使用されているため、
**一箇所の修正（PropertyRow.tsx）で全ページ一括対応**できる。

対応対象の Select 使用箇所（PropertyRow 内）：

| ページ | 対象 Select |
|--------|------------|
| MedicineSettings | 親カテゴリ（SELECT_TRIGGER_FULL） |
| MedicineSettings | 剤形（SELECT_TRIGGER_FULL） |
| MedicineSettings | 単位（SELECT_TRIGGER_FULL） |
| HospitalizationSettings | 対象体格（STYLE.selectCompact） |
| HospitalizationSettings | 料金単位（STYLE.selectCompact） |
| DiagnosisSettings | 各 Select（STYLE.selectCompact） |
| TrimmingSettings | 各 Select（STYLE.selectCompact） |
| その他マスタページ全般 | 同上 |

---

## 修正後の挙動

- ラベルにホバー → **ハイライトなし**
- SelectTrigger にホバー → **SelectTrigger のみ `rgba(55,53,47,0.04)` でハイライト** ✅
- 二重ハイライト問題（master-020）も同時解消 ✅
