---
description: Web accessibility standards (WCAG 2.1 AA, a11y)
alwaysApply: false
globs: ["frontend/src/**/*.{tsx,jsx}"]
---

# Accessibility Rules

Web accessibility standards (WCAG 2.1 AA).

## Core Rules

### 1. Semantic HTML

```typescript
// ❌ Non-semantic
<div onClick={handleClick}>Button</div>
<div role="heading">Title</div>

// ✅ Semantic
<button onClick={handleClick}>Button</button>
<h1>Title</h1>
<form>
  <label htmlFor="email">Email Address</label>
  <input id="email" type="email" />
</form>
```

### 2. Form Labels (Required)

```typescript
// ✅ Relate label to input with htmlFor
<label htmlFor="owner-name">Owner Name</label>
<input id="owner-name" type="text" />

// ❌ No label
<input type="text" placeholder="Owner Name" />
```

### 3. ARIA Attributes (Appropriate Use)

```typescript
// ✅ aria-label for function description
<button aria-label="Toggle Navigation">
  <MenuIcon />
</button>

// ✅ aria-live for dynamic updates
<div aria-live="polite" aria-atomic="true">
  {successMessage}
</div>

// ✅ aria-disabled (when HTML disabled unavailable)
<div role="button" aria-disabled="true">
  Saving...
</div>
```

### 4. Keyboard Operation Support

```typescript
// ✅ Respond to Enter/Space
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
  Action
</div>

// ✅ Manage Tab focus order
<div>
  <button tabIndex={1}>First</button>
  <button tabIndex={2}>Next</button>
  <button tabIndex={3}>Last</button>
</div>
```

### 5. Color Contrast

```
Target: WCAG AA or better
- Normal text: 4.5:1
- Large text (18pt+): 3:1

Tools: axe DevTools, WebAIM Contrast Checker
```

### 6. Image Alt Text

```typescript
// ✅ Meaningful alt text
<img src="owner-photo.jpg" alt="Taro Sato's profile photo" />

// ❌ No description
<img src="owner-photo.jpg" alt="image" />

// ✅ Decorative image
<img src="divider.svg" alt="" aria-hidden="true" />
```

### 7. Focus Management

```typescript
// ✅ Focus visible (for keyboard users)
input:focus {
  outline: 2px solid #4A90E2;
  outline-offset: 2px;
}

// ❌ Focus outline removed
input:focus {
  outline: none;  // Harms keyboard users
}

// ✅ Focus trap in modal
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

### 8. Screen Reader Support

```typescript
// ✅ aria-describedby for error description
<input
  id="email"
  aria-describedby="email-error"
  type="email"
/>
<p id="email-error" role="alert">
  Email format incorrect
</p>

// ✅ Hidden label (invisible visually, read by screen reader)
<label htmlFor="search" className="sr-only">
  Search keyword
</label>
<input id="search" type="search" />

// ✅ aria-live for dynamic content
<div aria-live="assertive" className="toast">
  {errorMessage}
</div>
```

## Checklist

- [ ] Semantic HTML (button, form, label etc.)
- [ ] Form labels linked via htmlFor to input
- [ ] Color contrast 4.5:1 or better
- [ ] Image alt text (meaningful)
- [ ] Keyboard operation (Tab, Enter, Space)
- [ ] Focus visible (outline not removed)
- [ ] ARIA: aria-label, aria-live, aria-describedby
- [ ] Screen reader testing (NVDA, JAWS)
- [ ] Lighthouse a11y score > 90

## Automated Audit Tools

```bash
# axe DevTools (Chrome extension)
# WAVE (WebAIM)
# Lighthouse (DevTools)
pppnpm audit:lighthouse  # a11y score check
```

## References

- [WCAG 2.1](https://www.w3.org/WAI/WCAG21/quickref/)
- [ARIA Practices](https://www.w3.org/WAI/ARIA/apg/)
- [WebAIM](https://webaim.org/)
