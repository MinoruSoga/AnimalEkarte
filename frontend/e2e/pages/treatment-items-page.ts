import { type Locator, type Page } from "@playwright/test";
import { BasePage } from "./base-page";

/** Treatment-plan master screen (`/settings/treatment-items`) + the `/settings` index card. */
export class TreatmentItemsPage extends BasePage {
  static readonly PATH = "/settings/treatment-items";

  gotoTab(tab: string): ReturnType<Page["goto"]> {
    return this.open(`${TreatmentItemsPage.PATH}?tab=${tab}`);
  }

  tabUrlPattern(tab: string): RegExp {
    return new RegExp(`/settings/treatment-items\\?tab=${tab}`);
  }

  planMasterHeading(): Locator {
    return this.heading("診療項目マスタ");
  }

  tab(name: string): Locator {
    return this.page.getByRole("tab", { name });
  }

  newButton(): Locator {
    return this.page.getByRole("button", { name: "新規登録" });
  }

  closeButton(): Locator {
    return this.page.getByRole("button", { name: "閉じる" });
  }

  parentCategoryLabel(): Locator {
    return this.page.getByText("親カテゴリ");
  }

  /** Parent-category combobox showing the current "なし（ルート）" selection. */
  parentCategoryCombobox(): Locator {
    return this.page.getByRole("combobox").filter({ hasText: "なし（ルート）" });
  }

  cannotChangeParentText(): Locator {
    return this.page.getByText("子項目があるため変更できません");
  }

  /** Seed root category "注射" (has children) inside the table body. */
  injectionRootItem(): Locator {
    return this.page.locator("tbody").getByText("注射", { exact: true }).first();
  }

  // --- /settings index ---

  gotoSettings(): ReturnType<Page["goto"]> {
    return this.open("/settings");
  }

  /** "主訴カテゴリ" card on the settings index — was `locator('text=主訴カテゴリ').first()`. */
  chiefComplaintCard(): Locator {
    return this.page.getByText("主訴カテゴリ").first();
  }

  body(): Locator {
    return this.page.locator("body");
  }
}
