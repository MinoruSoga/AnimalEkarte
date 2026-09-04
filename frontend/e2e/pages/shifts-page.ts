import { type Locator, type Page } from "@playwright/test";
import { BasePage } from "./base-page";

/** Shift calendar screen (`/shifts`). */
export class ShiftsPage extends BasePage {
  gotoCalendar(): ReturnType<Page["goto"]> {
    return this.open("/shifts");
  }

  calendarHeading(): Locator {
    return this.heading("シフト管理", 1);
  }

  prevMonthButton(): Locator {
    return this.page.getByRole("button", { name: "前月" });
  }

  nextMonthButton(): Locator {
    return this.page.getByRole("button", { name: "翌月" });
  }

  /** "YYYY年M月" month label. */
  monthLabel(): Locator {
    return this.page.getByText(/\d{4}年\d{1,2}月/);
  }

  firstCombobox(): Locator {
    return this.page.getByRole("combobox").first();
  }

  firstOption(): Locator {
    return this.page.getByRole("option").first();
  }
}
