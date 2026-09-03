import { type Locator, type Page } from "@playwright/test";
import { BasePage } from "./base-page";

/** Periodic checkup screens (`/checkups`). */
export class CheckupsPage extends BasePage {
  gotoList(): ReturnType<Page["goto"]> {
    return this.open("/checkups");
  }

  gotoSelectPet(): ReturnType<Page["goto"]> {
    return this.open("/checkups/select-pet");
  }

  gotoNew(query = ""): ReturnType<Page["goto"]> {
    return this.open(`/checkups/new${query}`);
  }

  listHeading(): Locator {
    return this.heading("定期健診", 1);
  }

  selectPetHeading(): Locator {
    return this.heading("定期健診登録 - ペット選択");
  }

  newFormHeading(): Locator {
    return this.heading("定期健診登録");
  }

  newButton(): Locator {
    return this.page.getByRole("button", { name: "新規登録" });
  }

  irisText(): Locator {
    return this.page.getByText("Iris").first();
  }

  patientSearchInput(): Locator {
    return this.page.locator("#search");
  }

  saveButton(): Locator {
    return this.page.getByRole("button", { name: "保存" });
  }

  searchInput(): Locator {
    return this.page.getByPlaceholder("ペット名・飼主名・種別で検索...");
  }
}
