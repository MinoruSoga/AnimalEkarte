# FE-245: 画像アップロードのファイルサイズ・型バリデーション欠落

## 概要

複数箇所の画像アップロード機能でファイルサイズの上限チェックがない。
巨大ファイル（数百 MB〜GB）をアップロードできてしまい、ブラウザのメモリ枯渇や
サーバー側への無制限の負荷が発生する可能性がある。

## 問題箇所

### `frontend/src/features/medical-records/components/ImageGalleryFilter.tsx:48-52, 64`

```tsx
// input は accept 属性があるが、ファイルサイズチェックなし
<input
  type="file"
  accept="image/jpeg,image/png,image/gif,application/pdf"
  onChange={handleFileChange}
/>

// handleFileChange でサイズバリデーションなし
const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
  const file = e.target.files?.[0];
  if (!file) return;
  // サイズチェックなし → 1GB の画像もそのまま処理される
  upload(file);
};
```

### `frontend/src/features/trimming/routes/TrimmingForm.tsx:184, 297`

```tsx
// 2箇所のファイルインプット、いずれもサイズチェックなし
<input type="file" accept="image/*" onChange={onStyleImageChange} className="hidden" />
<input type="file" accept="image/*" onChange={onCompletedImageChange} className="hidden" />
```

### `frontend/src/features/trimming/hooks/use-trimming-form.ts:213-231`

```ts
// FileReader.readAsDataURL() でファイル全体をメモリに読み込む前にサイズチェックなし
const handleStyleImageChange = (e: ChangeEvent<HTMLInputElement>) => {
  const file = e.target.files?.[0];
  if (!file) return;
  // サイズチェックなし → 大容量ファイルでブラウザメモリ枯渇の可能性
  const reader = new FileReader();
  reader.readAsDataURL(file);
};
```

## 修正方針

```ts
// 共通バリデーション関数を lib/ に追加
const MAX_IMAGE_SIZE_MB = 10;
const MAX_IMAGE_SIZE_BYTES = MAX_IMAGE_SIZE_MB * 1024 * 1024;

const validateImageFile = (file: File): string | null => {
  if (file.size > MAX_IMAGE_SIZE_BYTES) {
    return `ファイルサイズは ${MAX_IMAGE_SIZE_MB}MB 以下にしてください（現在: ${(file.size / 1024 / 1024).toFixed(1)}MB）`;
  }
  const allowedTypes = ["image/jpeg", "image/png", "image/gif", "image/webp"];
  if (!allowedTypes.includes(file.type)) {
    return "JPEG・PNG・GIF・WebP 形式の画像のみアップロードできます";
  }
  return null;
};

// 各ハンドラで使用
const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
  const file = e.target.files?.[0];
  if (!file) return;
  const error = validateImageFile(file);
  if (error) {
    toast.error(error);
    return;
  }
  upload(file);
};
```

## 影響

- **ブラウザクラッシュ**: `readAsDataURL()` は巨大ファイルを base64 でメモリに展開するため、大容量ファイルでブラウザが応答不能になる可能性がある
- **サーバー負荷**: フロントエンドでのバリデーションなしにサーバーへ送信されると、サーバー側の upload 処理に無制限の負荷がかかる

## 優先度
**High** — ユーザーが誤って大きなファイルをアップロードするとブラウザがハングする可能性がある。

## 関連ファイル
- `frontend/src/features/medical-records/components/ImageGalleryFilter.tsx`
- `frontend/src/features/trimming/routes/TrimmingForm.tsx`
- `frontend/src/features/trimming/hooks/use-trimming-form.ts`
