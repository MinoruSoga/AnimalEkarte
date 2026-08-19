import { type Locator, type Page } from '@playwright/test';
import { BasePage } from './base-page';

/** Examination screens (`/examinations`). */
export class ExaminationsPage extends BasePage {
  gotoList(): ReturnType<Page['goto']> {
    return this.open('/examinations');
  }

  gotoSelectPet(): ReturnType<Page['goto']> {
    return this.open('/examinations/select-pet');
  }

  /** List heading, exact-string variant: `{ name: '検査管理', level: 1 }`. */
  listHeading(): Locator {
    return this.heading('検査管理', 1);
  }

  /** List heading, regex variant: `{ name: /検査管理/, level: 1 }` (avoids strict-locator clash). */
  listHeadingPattern(): Locator {
    return this.heading(/検査管理/, 1);
  }

  selectPetHeading(): Locator {
    return this.heading('検査登録 - ペット選択');
  }

  detailHeading(): Locator {
    return this.heading('検査詳細・編集', 1);
  }

  irisText(): Locator {
    return this.page.getByText('Iris').first();
  }

  patientSearchInput(): Locator {
    return this.page.locator('#search');
  }

  firstDetailLink(): Locator {
    return this.page.getByRole('link', { name: /検査詳細:/ }).first();
  }

  saveButton(): Locator {
    return this.page.getByRole('button', { name: '保存' });
  }

  searchInput(): Locator {
    return this.page.getByPlaceholder('飼主名、ペット名、検査種別...');
  }

  /** 4th cell (検査種別) of the first row — positional `td.nth(3)`, kept in one place. */
  firstRowTestTypeCell(): Locator {
    return this.firstRow().locator('td').nth(3);
  }
}
