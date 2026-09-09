import { type Locator } from "@playwright/test";
import { BasePage } from "./base-page";

/**
 * Generic master-settings CRUD screen (`/settings/*`).
 *
 * Shared controls across masters: `#master-title`, row edit, toolbar delete,
 * and the delete confirmation dialog. Callers navigate with `open(path)` and
 * assert page-specific headings / search placeholders inline.
 */
export class SettingsMasterPage extends BasePage {
  masterTitleInput(): Locator {
    return this.page.locator("#master-title");
  }

  newButton(): Locator {
    return this.page.getByRole("button", { name: "新規登録" });
  }

  saveButton(): Locator {
    return this.page.getByRole("button", { name: "保存" });
  }

  cancelButton(): Locator {
    return this.page.getByRole("button", { name: "キャンセル" });
  }

  /** Table row whose cells contain `text`. */
  rowContaining(text: string): Locator {
    return this.page.locator("tbody tr").filter({ hasText: text });
  }

  /**
   * Row edit/action button for the row containing `text`.
   * Masters use accessible names like `編集: …` or `…を編集` (not bare "操作").
   */
  rowActionButton(text: string): Locator {
    return this.rowContaining(text).getByRole("button", { name: /編集|を編集|操作/ });
  }

  /** Toolbar delete button (aria-label "削除"). */
  deleteButton(): Locator {
    return this.page.getByLabel("削除");
  }

  deleteDialog(): Locator {
    return this.page.getByRole("alertdialog");
  }

  /** Confirm button inside the delete alertdialog (masters vary: 削除 / 削除する). */
  deleteConfirmButton(confirmLabel: string | RegExp = "削除"): Locator {
    return this.deleteDialog().getByRole("button", { name: confirmLabel });
  }

  searchToggle(): Locator {
    return this.page.getByLabel("検索");
  }

  searchInput(placeholder: string): Locator {
    return this.page.getByPlaceholder(placeholder);
  }

  /** Open PropertyFilter search and fill the placeholder-specific input. */
  async searchFor(placeholder: string, term: string): Promise<void> {
    await this.searchToggle().click();
    await this.searchInput(placeholder).fill(term);
  }

  /** Medicine side-panel unit price field. */
  medicinePriceInput(): Locator {
    return this.page.getByLabel("単価(税込)");
  }

  toast(): Locator {
    return this.page.locator("[data-sonner-toast]");
  }
}
