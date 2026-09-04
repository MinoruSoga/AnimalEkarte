import { type Locator, type Page } from "@playwright/test";
import { BasePage } from "./base-page";

/** Hospitalization screens (`/hospitalization`). Defaults to board view; list view holds the table. */
export class HospitalizationPage extends BasePage {
  gotoList(): ReturnType<Page["goto"]> {
    return this.open("/hospitalization");
  }

  listHeading(): Locator {
    return this.heading("入院・ホテル管理");
  }

  selectPetHeading(): Locator {
    return this.heading("入院・ホテル登録 - ペット選択");
  }

  newButton(): Locator {
    return this.page.getByRole("button", { name: "新規入院登録" });
  }

  /** Board → list view toggle (aria-label "List View"). */
  listViewToggle(): Locator {
    return this.page.getByLabel("List View");
  }

  statusTab(name: string): Locator {
    return this.page.getByRole("tab", { name });
  }
}
