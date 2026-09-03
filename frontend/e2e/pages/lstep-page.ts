import { type Locator, type Page } from "@playwright/test";
import { BasePage } from "./base-page";

/** L-step integration screens (`/lstep/*`). */
export class LstepPage extends BasePage {
  gotoCheckupSync(): ReturnType<Page["goto"]> {
    return this.open("/lstep/checkup-sync");
  }

  gotoDeliveryMonitor(): ReturnType<Page["goto"]> {
    return this.open("/lstep/delivery-monitor");
  }

  gotoAnalytics(): ReturnType<Page["goto"]> {
    return this.open("/lstep/analytics");
  }

  checkupSyncHeading(): Locator {
    return this.heading("健診リマインダー抽出", 1);
  }

  searchTargetsButton(): Locator {
    return this.page.getByRole("button", { name: "対象者を検索" });
  }

  /** `select[name="checkup_type"]` — native select element. */
  checkupTypeSelect(): Locator {
    return this.page.locator('select[name="checkup_type"]');
  }

  ltvAmountInput(): Locator {
    return this.page.getByPlaceholder("例: 50000");
  }

  deliveryMonitorHeading(): Locator {
    return this.heading("自動配信トリガー監視", 1);
  }

  refreshButton(): Locator {
    return this.page.getByRole("button", { name: "更新" });
  }

  filterFrom(): Locator {
    return this.page.getByTestId("filter-from");
  }

  filterTriggerType(): Locator {
    return this.page.getByTestId("filter-trigger-type");
  }

  filterStatus(): Locator {
    return this.page.getByTestId("filter-status");
  }

  analyticsHeading(): Locator {
    return this.heading("Lステップ分析レポート", 1);
  }

  monthlyStatsText(): Locator {
    return this.page.getByText("月次配信統計");
  }

  visitRateText(): Locator {
    return this.page.getByText("配信後来院率");
  }

  csvImportHeading(): Locator {
    return this.heading("友だち属性 CSV インポート");
  }
}
