import { type Locator, type Page } from "@playwright/test";
import { BasePage } from "./base-page";

/** Trimming screens (`/trimming`). */
export class TrimmingPage extends BasePage {
  gotoList(): ReturnType<Page["goto"]> {
    return this.open("/trimming");
  }

  gotoNew(query = ""): ReturnType<Page["goto"]> {
    return this.open(`/trimming/new${query}`);
  }

  listHeading(): Locator {
    return this.heading("トリミング管理");
  }

  selectPetHeading(): Locator {
    return this.heading("トリミング登録 - ペット選択");
  }

  newFormHeading(): Locator {
    return this.heading("トリミング登録");
  }

  newButton(): Locator {
    return this.page.getByRole("button", { name: "新規登録" });
  }

  irisText(): Locator {
    return this.page.getByText("Iris").first();
  }

  saveButton(): Locator {
    return this.page.getByRole("button", { name: "保存" });
  }
}
