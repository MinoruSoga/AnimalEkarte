import { type Locator, type Page } from "@playwright/test";

/**
 * Shared Page Object base.
 *
 * Centralizes the navigation and table-row locators that were previously
 * duplicated inline across every spec (`page.locator('tbody tr').first()` etc).
 * Every locator string here is copied **verbatim** from the original specs — this
 * is a behavior-neutral relocation, not a locator-strategy change.
 */
export class BasePage {
  constructor(protected readonly page: Page) {}

  /** `page.goto(path, { waitUntil: 'domcontentloaded' })` — the suite-wide default. */
  open(path: string): ReturnType<Page["goto"]> {
    return this.page.goto(path, { waitUntil: "domcontentloaded" });
  }

  /** `getByRole('heading', { name?, level? })`. Both options are optional to cover every call site. */
  heading(name?: string | RegExp, level?: number): Locator {
    const options: { name?: string | RegExp; level?: number } = {};
    if (name !== undefined) options.name = name;
    if (level !== undefined) options.level = level;
    return this.page.getByRole("heading", options);
  }

  /** `page.locator('tbody')`. */
  tableBody(): Locator {
    return this.page.locator("tbody");
  }

  /** `page.locator('tbody tr')`. */
  rows(): Locator {
    return this.page.locator("tbody tr");
  }

  /** `page.locator('tbody tr').first()` — the canonical "first data row". */
  firstRow(): Locator {
    return this.page.locator("tbody tr").first();
  }
}
