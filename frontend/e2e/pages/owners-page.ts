import { type Locator, type Page } from '@playwright/test';
import { BasePage } from './base-page';

/** Owners screens (`/owners`). Shared by owners-flow + owners-search. */
export class OwnersPage extends BasePage {
  gotoList(): ReturnType<Page['goto']> {
    return this.open('/owners');
  }

  gotoDetail(id: number | string): ReturnType<Page['goto']> {
    return this.open(`/owners/${id}`);
  }

  listHeading(): Locator {
    return this.heading('飼主・ペット一覧');
  }

  editHeading(): Locator {
    return this.heading('飼主・ペット　編集');
  }

  registerHeading(): Locator {
    return this.heading('飼主・ペット　登録');
  }

  newButton(): Locator {
    return this.page.getByRole('button', { name: '新規登録' });
  }

  searchInput(): Locator {
    return this.page.getByPlaceholder('飼主名、ペット名、電話番号...');
  }

  /** Owner "林 文明" row/cell — `.first()` covers multiple matching rows. */
  hayashiText(): Locator {
    return this.page.getByText('林 文明').first();
  }

  /** Pet "ピーター" — kana-search assertion target (multiple seed rows). */
  peterText(): Locator {
    return this.page.getByRole('cell', { name: 'ピーター', exact: true }).first();
  }

  ownerNameInput(): Locator {
    return this.page.locator('#ownerName');
  }
}
