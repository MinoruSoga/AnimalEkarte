import { type Locator, type Page } from "@playwright/test";
import { BasePage } from "./base-page";

/** Vaccination screens (`/vaccinations`). */
export class VaccinationsPage extends BasePage {
  gotoList(): ReturnType<Page["goto"]> {
    return this.open("/vaccinations");
  }

  listHeading(): Locator {
    return this.heading("予防接種管理");
  }

  selectPetHeading(): Locator {
    return this.heading("ワクチン接種 - ペット選択");
  }

  detailHeading(): Locator {
    return this.heading("予防接種詳細・編集");
  }

  newButton(): Locator {
    return this.page.getByRole("button", { name: "新規登録" });
  }

  searchInput(): Locator {
    return this.page.getByPlaceholder("飼主名、ペット名、予防接種名...");
  }

  ownerText(name: string): Locator {
    return this.page.getByText(name, { exact: false }).first();
  }

  firstDetailLink(): Locator {
    return this.page.getByRole("link", { name: /予防接種詳細:/ }).first();
  }
}
