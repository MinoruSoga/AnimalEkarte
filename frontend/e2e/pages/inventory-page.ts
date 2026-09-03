import { type Locator, type Page } from "@playwright/test";
import { BasePage } from "./base-page";

/** Inventory screens (`/inventory`). */
export class InventoryPage extends BasePage {
  gotoList(): ReturnType<Page["goto"]> {
    return this.open("/inventory");
  }

  gotoNew(): ReturnType<Page["goto"]> {
    return this.open("/inventory/new");
  }

  listHeading(): Locator {
    return this.heading("在庫管理");
  }

  newFormHeading(): Locator {
    return this.heading("在庫登録");
  }

  editHeading(): Locator {
    return this.heading("在庫編集");
  }

  nameInput(): Locator {
    return this.page.locator("#name");
  }

  unitInput(): Locator {
    return this.page.locator("#unit");
  }

  quantityInput(): Locator {
    return this.page.locator("#quantity");
  }

  /** Submit button bound to the form via `form="inventory-form"` (attribute selector avoids duplicate matches). */
  submitButton(): Locator {
    return this.page.locator('button[form="inventory-form"]');
  }

  searchInput(): Locator {
    return this.page.getByPlaceholder("品名、保管場所、仕入先...");
  }

  itemText(name: string): Locator {
    return this.page.getByText(name);
  }
}
