import { type Locator, type Page } from "@playwright/test";
import { DEMO_ACCOUNTING_KANA_PET } from "../helpers/demo-seed";
import { BasePage } from "./base-page";

type GotoOptions = Parameters<Page["goto"]>[1];

/** Accounting screens (`/accounting`, `/accounting/reports`). Shared by accounting-flow + accounting-smoke. */
export class AccountingPage extends BasePage {
  gotoList(): ReturnType<Page["goto"]> {
    return this.open("/accounting");
  }

  gotoUnpaid(): ReturnType<Page["goto"]> {
    return this.open("/accounting?tab=unpaid");
  }

  /** accounting-flow passes `{ waitUntil: 'domcontentloaded' }`; accounting-smoke passes nothing (default `load`). */
  gotoReports(options?: GotoOptions): ReturnType<Page["goto"]> {
    return this.page.goto("/accounting/reports", options);
  }

  listTab(): Locator {
    return this.page.getByRole("tab", { name: "会計一覧" });
  }

  unpaidTab(): Locator {
    return this.page.getByRole("tab", { name: "未納者一覧" });
  }

  sameDayTab(): Locator {
    return this.page.getByRole("tab", { name: "当日会計" });
  }

  searchToggle(): Locator {
    return this.page.getByRole("button", { name: "検索" });
  }

  searchInput(): Locator {
    return this.page.getByPlaceholder("飼主名、ペット名...");
  }

  /**
   * Pet cell used for client-side kana-symmetry smoke.
   * Prefer DEMO_ACCOUNTING_KANA_PET (on page 1) over Iris (no billing in 003_demo).
   */
  kanaPetCell(): Locator {
    return this.page
      .locator("tbody")
      .getByText(DEMO_ACCOUNTING_KANA_PET.displayName, { exact: true })
      .first();
  }

  /** @deprecated Prefer kanaPetCell — Iris has no billing rows in current seed. */
  irisCell(): Locator {
    return this.kanaPetCell();
  }

  /** First table row that contains the kana-smoke pet name. */
  irisRow(): Locator {
    return this.page
      .locator("tbody tr")
      .filter({ hasText: DEMO_ACCOUNTING_KANA_PET.displayName })
      .first();
  }

  /**
   * Detail navigation link inside a row (DataTableRowLink on the date cell).
   * Row click alone does not navigate — only this link does.
   */
  firstDetailLink(): Locator {
    return this.firstRow().getByRole("link").first();
  }

  kanaPetDetailLink(): Locator {
    return this.irisRow().getByRole("link").first();
  }

  detailHeading(): Locator {
    return this.page.getByRole("heading", { name: "会計精算" });
  }

  reportsHeading(): Locator {
    return this.page.getByRole("heading", { name: "月次集計レポート" });
  }

  firstCombobox(): Locator {
    return this.page.getByRole("combobox").first();
  }

  confirmButton(): Locator {
    return this.page.getByRole("button", { name: /会計を確定する|修正を保存する/ });
  }
}
