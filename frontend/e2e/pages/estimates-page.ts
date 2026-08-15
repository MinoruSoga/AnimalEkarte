import { type Locator, type Page } from '@playwright/test';
import { BasePage } from './base-page';

/** Estimate screens (`/estimates`). */
export class EstimatesPage extends BasePage {
  gotoList(): ReturnType<Page['goto']> {
    return this.open('/estimates');
  }

  gotoNew(): ReturnType<Page['goto']> {
    return this.open('/estimates/new');
  }

  listHeading(): Locator {
    return this.heading('見積書管理', 1);
  }

  newFormHeading(): Locator {
    return this.heading('新規見積書作成', 1);
  }

  /** Detail heading "見積書 <estimateNo>" (title may be empty in seed). */
  detailHeading(): Locator {
    return this.heading(/見積書\s+\S+/, 1);
  }

  editHeading(): Locator {
    return this.heading('見積書編集', 1);
  }

  newButton(): Locator {
    return this.page.getByRole('button', { name: '新規見積書登録' });
  }

  createButton(): Locator {
    return this.page.getByRole('button', { name: '作成' });
  }

  editButton(): Locator {
    return this.page.getByRole('button', { name: '編集' });
  }

  titleInput(): Locator {
    return this.page.getByPlaceholder('見積書タイトルを入力');
  }

  searchInput(): Locator {
    return this.page.getByPlaceholder('見積番号、タイトル、飼主名...');
  }

  /** Row with status badge text 下書き (editable; not approved/rejected). */
  draftRow(): Locator {
    return this.page.locator('tbody tr').filter({ hasText: '下書き' }).first();
  }

  /**
   * Detail link on the estimate-no cell (DataTableRowLink).
   * Row body click does not navigate — use this link.
   */
  draftDetailLink(): Locator {
    return this.draftRow().getByRole('link').first();
  }

  searchToggle(): Locator {
    return this.page.getByRole('button', { name: '検索' });
  }
}
