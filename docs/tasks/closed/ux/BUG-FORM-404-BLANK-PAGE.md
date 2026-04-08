# BUG-FORM-404-BLANK-PAGE: フォームページで存在しないIDにアクセスすると空白ページが表示される

## ステータス
🔴 **未修正**

## 優先度
Medium

## 再現手順
1. `/trimming/1` にアクセス（id=1 は DBに存在しない）
2. ページが空白（コンテンツエリアが完全に空）になることを確認

## 症状
APIが 404 を返した際に、フォームコンポーネントがエラー状態を表示せず
ページのコンテンツエリアが空白になる。
ヘッダーのタイトルと「保存」ボタンのみ残る。

## 確認済みページ
- `/trimming/:id`（id=1 → API 404）

## 他に同様の問題が発生しうるページ
- `/medical-records/:id`
- `/hospitalization/:id`
- `/examinations/:id`
- `/accounting/:id`
- その他 `/:id` 形式の全フォームページ

## 根本原因
フォームコンポーネントが API 404 エラーをハンドリングせず、
`data` が `null/undefined` の場合のフォールバック表示がない。

## 修正方針
各フォームの API hook でエラー検出時に
`ErrorFallback` コンポーネントまたは「データが見つかりません」メッセージを表示する。

```tsx
const { data, isLoading, error } = useGetTrimming(id);
if (isLoading) return <LoadingFallback />;
if (error || !data) return <ErrorFallback message="トリミング記録が見つかりません" />;
```

または `router.tsx` に `errorElement` を設定して React Router の Error Boundary で捕捉する。
