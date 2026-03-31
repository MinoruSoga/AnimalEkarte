---
description: Web アクセシビリティ規約（WCAG 2.1 AA、a11y）
alwaysApply: false
globs: ["frontend/src/**/*.{tsx,jsx}"]
---

# Accessibility Rules

Web アクセシビリティ標準（WCAG 2.1 AA）。

## 核心ルール

### 1. セマンティック HTML

```typescript
// ❌ 非セマンティック
<div onClick={handleClick}>ボタン</div>
<div role="heading">タイトル</div>

// ✅ セマンティック
<button onClick={handleClick}>ボタン</button>
<h1>タイトル</h1>
<form>
  <label htmlFor="email">メールアドレス</label>
  <input id="email" type="email" />
</form>
```

### 2. フォームラベル（必須）

```typescript
// ✅ label と input を関連付け
<label htmlFor="owner-name">オーナー名</label>
<input id="owner-name" type="text" />

// ❌ ラベルなし
<input type="text" placeholder="オーナー名" />
```

### 3. ARIA 属性（適切に）

```typescript
// ✅ aria-label で機能説明
<button aria-label="ナビゲーション開閉">
  <MenuIcon />
</button>

// ✅ aria-live で動的更新を通知
<div aria-live="polite" aria-atomic="true">
  {successMessage}
</div>

// ✅ aria-disabled（HTML disabled が使えない場合）
<div role="button" aria-disabled="true">
  保存中...
</div>
```

### 4. キーボード操作対応

```typescript
// ✅ Enter・Space で操作可能
const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
  if (e.key === 'Enter' || e.key === ' ') {
    handleAction();
  }
};

<div
  role="button"
  tabIndex={0}
  onKeyDown={handleKeyDown}
  onClick={handleAction}
>
  アクション
</div>

// ✅ Tab フォーカス順序を管理
<div>
  <button tabIndex={1}>最初</button>
  <button tabIndex={2}>次</button>
  <button tabIndex={3}>最後</button>
</div>
```

### 5. 色コントラスト

```
目標: WCAG AA 以上
- 通常テキスト: 4.5:1
- 大型テキスト（18pt+）: 3:1

チェックツール: axe DevTools, WebAIM Contrast Checker
```

### 6. 画像代替テキスト

```typescript
// ✅ 意味のある alt テキスト
<img src="owner-photo.jpg" alt="佐藤太郎さんのプロフィール写真" />

// ❌ 画像の説明なし
<img src="owner-photo.jpg" alt="画像" />

// ✅ 装飾的な画像
<img src="divider.svg" alt="" aria-hidden="true" />
```

### 7. フォーカス管理

```typescript
// ✅ フォーカス visible（キーボード利用者向け）
input:focus {
  outline: 2px solid #4A90E2;
  outline-offset: 2px;
}

// ❌ フォーカス outline削除
input:focus {
  outline: none;  // キーボード利用者が困る
}

// ✅ モーダル内でフォーカストラップ
useEffect(() => {
  const focusableElements = dialogRef.current?.querySelectorAll(
    'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
  );
  const firstElement = focusableElements?.[0] as HTMLElement;
  const lastElement = focusableElements?.[focusableElements.length - 1] as HTMLElement;

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'Tab') {
      if (e.shiftKey && document.activeElement === firstElement) {
        e.preventDefault();
        lastElement?.focus();
      } else if (!e.shiftKey && document.activeElement === lastElement) {
        e.preventDefault();
        firstElement?.focus();
      }
    }
  };

  dialogRef.current?.addEventListener('keydown', handleKeyDown);
  return () => dialogRef.current?.removeEventListener('keydown', handleKeyDown);
}, []);
```

### 8. スクリーンリーダー対応

```typescript
// ✅ aria-describedby でエラー説明
<input
  id="email"
  aria-describedby="email-error"
  type="email"
/>
<p id="email-error" role="alert">
  メールアドレスの形式が正しくありません
</p>

// ✅ 非表示ラベル（視覚的には非表示だがスクリーンリーダーで読み込まれる）
<label htmlFor="search" className="sr-only">
  検索キーワード
</label>
<input id="search" type="search" />

// ✅ aria-live で動的コンテンツ通知
<div aria-live="assertive" className="toast">
  {errorMessage}
</div>
```

## チェックリスト

- [ ] セマンティック HTML（button, form, label 等）
- [ ] form ラベル htmlFor で input と関連付け
- [ ] 色コントラスト 4.5:1 以上
- [ ] 画像 alt テキスト記載（意味のあるテキスト）
- [ ] キーボード操作対応（Tab, Enter, Space）
- [ ] フォーカス visible（outline 削除なし）
- [ ] ARIA: aria-label, aria-live, aria-describedby
- [ ] スクリーンリーダーテスト（NVDA, JAWS）
- [ ] Lighthouse a11y スコア > 90

## 自動監査ツール

```bash
# axe DevTools (Chrome 拡張)
# Wave (WebAIM)
# Lighthouse (DevTools)
npm run audit:lighthouse  # a11y スコア確認
```

## 参照

- [WCAG 2.1](https://www.w3.org/WAI/WCAG21/quickref/)
- [ARIA Practices](https://www.w3.org/WAI/ARIA/apg/)
- [WebAIM](https://webaim.org/)
