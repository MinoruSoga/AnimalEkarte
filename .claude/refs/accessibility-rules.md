---
description: Web accessibility standards (WCAG 2.2 AA, a11y)
alwaysApply: false
globs: ["frontend/src/**/*.{tsx,jsx}"]
---

# Accessibility Rules

Web accessibility standards (WCAG 2.2 AA — the same baseline as [docs/spec/design-states.md](../../docs/spec/design-states.md); 2.1 criteria remain included). Visual state specs (hover / focus-visible / disabled / loading / invalid), measured contrast ratios and touch-target sizes are owned by design-states.md; this file covers semantic HTML, labels, ARIA and keyboard behavior.

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

// ❌ NEVER use positive tabIndex to control order — it breaks natural DOM
// tab order and is unmaintainable as the page grows.
<div>
  <button tabIndex={1}>First</button>
  <button tabIndex={2}>Next</button>
  <button tabIndex={3}>Last</button>
</div>

// ✅ Let natural DOM order define Tab order. For composite widgets
// (toolbars, tab lists) that need arrow-key navigation within a group,
// use the "roving tabindex" pattern: only the active item gets tabIndex={0},
// all others get tabIndex={-1}, and arrow keys move focus programmatically.
function Toolbar({ items, activeIndex, onNavigate }: ToolbarProps) {
  return (
    <div role="toolbar">
      {items.map((item, i) => (
        <button
          key={item.id}
          tabIndex={i === activeIndex ? 0 : -1}
          ref={i === activeIndex ? activeRef : undefined}
          onKeyDown={(e) => handleArrowKeys(e, i, onNavigate)}
        >
          {item.label}
        </button>
      ))}
    </div>
  );
}
```

### 5. Color Contrast

```
Target: WCAG 2.2 AA or better (measured values: docs/spec/design-states.md §2)
- Normal text: 4.5:1
- Large text (18pt+ or 14pt bold+): 3:1
- Non-text UI (borders, focus rings, icons): 3:1

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

### 9. WCAG 2.2 — criteria added since 2.1 (AA)

Claiming WCAG 2.2 AA means covering the criteria 2.2 added on top of 2.1. None of them is fully automatable (axe-core checks only `target-size`), so each has an owner document and a manual review check.

| SC | Requirement | Owner / rule | Review check |
|---|---|---|---|
| 2.4.11 Focus Not Obscured (Minimum) | A focused element is not entirely hidden by author content (sticky bars, FAB, toasts, dialogs) | `design-patterns.md` §5.1 / §5.2 (fixed elements do not overlap operable elements; trailing padding) | Tab through the page at 1280px and 768px; the focused element stays at least partially visible under fixed bars, the help FAB and open toasts |
| 2.5.7 Dragging Movements | Any drag operation has a single-pointer, non-drag alternative | `design-patterns.md` §2.6 (sortable lists: focused handle + arrow keys, or an up/down action in the row menu) | Every `dnd-kit` / sortable surface can be reordered without dragging; the alternative is reachable with keyboard and with a single tap |
| 2.5.8 Target Size (Minimum) | Targets are at least 24×24 CSS px, or have enough spacing | `design-states.md` §1.1 (44×44 hit area; dense row actions `h-9` min-w-11 per design-system §7.2 still exceed 24px) | `target-size` rule enabled in `.storybook/preview.ts`; manual check for inline links/icons inside text |
| 3.2.6 Consistent Help | A help mechanism that appears on multiple pages is in the same relative position on each | `design-patterns.md` §5.1 (fixed help button position) | The floating help button keeps the same corner and offset on every route |
| 3.3.7 Redundant Entry | Information already entered in a flow is auto-populated or selectable | Screen specs under `docs/spec/screens/`; product-philosophy (二重入力禁止) | Multi-step forms (select-pet → record, estimate → accounting) reuse owner/pet context instead of asking again |
| 3.3.8 Accessible Authentication (Minimum) | Login has no cognitive function test unless an alternative exists | `docs/spec/screens/` login spec | Password managers and paste are not blocked; no CAPTCHA/puzzle without an alternative |

## Checklist

- [ ] Semantic HTML (button, form, label etc.)
- [ ] WCAG 2.2 additions reviewed (focus not obscured, no drag-only interaction, 24px targets, consistent help position, no redundant entry, accessible login) — see §9
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
# Lighthouse (Chrome DevTools → Lighthouse タブ)
docker compose exec frontend pnpm lighthouse  # a11y score check
```

## References

- [WCAG 2.2](https://www.w3.org/WAI/WCAG22/quickref/)
- [ARIA Practices](https://www.w3.org/WAI/ARIA/apg/)
- [WebAIM](https://webaim.org/)
