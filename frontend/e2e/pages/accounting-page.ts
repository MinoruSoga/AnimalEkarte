import { type Locator, type Page } from '@playwright/test';
import { BasePage } from './base-page';

type GotoOptions = Parameters<Page['goto']>[1];

/** Accounting screens (`/accounting`, `/accounting/reports`). Shared by accounting-flow + accounting-smoke. */
export class AccountingPage extends BasePage {
  gotoList(): ReturnType<Page['goto']> {
    return this.open('/accounting');
  }

  gotoUnpaid(): ReturnType<Page['goto']> {
    return this.open('/accounting?tab=unpaid');
  }

  /** accounting-flow passes `{ waitUntil: 'domcontentloaded' }`; accounting-smoke passes nothing (default `load`). */
  gotoReports(options?: GotoOptions): ReturnType<Page['goto']> {
    return this.page.goto('/accounting/reports', options);
  }

  listTab(): Locator {
    return this.page.getByRole('tab', { name: '会計一覧' });
  }

  unpaidTab(): Locator {
    return this.page.getByRole('tab', { name: '未納者一覧' });
  }

  sameDayTab(): Locator {
    return this.page.getByRole('tab', { name: '当日会計' });
  }

  searchToggle(): Locator {
    return this.page.getByRole('button', { name: '検索' });
  }

  searchInput(): Locator {
    return this.page.getByPlaceholder('飼主名、ペット名...');
  }

  /** First "Iris" cell inside the table body. */
  irisCell(): Locator {
    return this.page.locator('tbody').getByText('Iris', { exact: false }).first();
  }

  /** First table row that contains "Iris". */
  irisRow(): Locator {
    return this.page.locator('tbody tr').filter({ hasText: 'Iris' }).first();
  }

  detailHeading(): Locator {
    return this.page.getByRole('heading', { name: '会計精算' });
  }

  reportsHeading(): Locator {
    return this.page.getByRole('heading', { name: '月次集計レポート' });
  }

  firstCombobox(): Locator {
    return this.page.getByRole('combobox').first();
  }

  confirmButton(): Locator {
    return this.page.getByRole('button', { name: /会計を確定する|修正を保存する/ });
  }
}
