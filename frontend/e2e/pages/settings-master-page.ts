import { type Locator } from "@playwright/test";
import { BasePage } from "./base-page";

/**
 * Generic master-settings CRUD screen (`/settings/*`).
 *
 * The page/heading/search-placeholder differ per master; this object only
 * encapsulates the controls that are identical across them — the `#master-title`
 * panel input, the "操作"/"削除" buttons, and the delete confirmation dialog.
 * Callers navigate with `open(path)` and assert page-specific headings inline.
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

  /** Row "操作" (edit) button for the row containing `text`. */
  rowActionButton(text: string): Locator {
    return this.page.locator("tbody tr").filter({ hasText: text }).getByLabel("操作");
  }

  /** Toolbar delete button (aria-label "削除"). */
  deleteButton(): Locator {
    return this.page.getByLabel("削除");
  }

  deleteDialog(): Locator {
    return this.page.getByRole("alertdialog");
  }
}
