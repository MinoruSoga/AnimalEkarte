# FE-154: HospitalizationDetail — デイリーカルテの日付が NaN年NaN月NaN日（undefined）と表示される

## 概要

`/hospitalization/:id` の詳細ページにある「デイリーカルテ」セクションで、現在表示している日付が `NaN年NaN月NaN日（undefined）` と表示される。

## 影響範囲

- `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx`
- 入院詳細ページ全体（すべての入院レコードで発生）

## 現状の挙動（バグ）

```
NaN年NaN月NaN日（undefined）
```

「前日」「翌日」ナビゲーションボタンが disabled になっており、日付のナビゲーションが機能しない。

## 期待する挙動

```
2026年4月7日（火）
```

現在日付または入院期間内の最初の日付が正しく表示される。

## 推定原因

日付の初期化ロジックで Date オブジェクトが正しく生成されていない可能性がある：

```tsx
// DailyRecordsTab.tsx — 推定コード
const [currentDate, setCurrentDate] = useState(new Date(hospitalization.admissionDate));
// admissionDate が undefined または不正フォーマットの場合 NaN になる
```

または日付フォーマット関数への引数が undefined になっている可能性がある。

## 優先度

HIGH — デイリーカルテの日付ナビゲーションが完全に機能せず、「この日の記録を作成」は表示されるが日付が不明なため、作成しても正しい日付で保存されない可能性がある。

## 関連

- `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx`
- 発見日: 2026-04-07（RBAC テスト中）
