import { type Locator, type Page } from '@playwright/test';
import { BasePage } from './base-page';

/** Medical record screens (`/medical-records`). Shared by clinical-flows + medical-records-create. */
export class MedicalRecordsPage extends BasePage {
  gotoList(): ReturnType<Page['goto']> {
    return this.open('/medical-records');
  }

  gotoSelectPet(): ReturnType<Page['goto']> {
    return this.open('/medical-records/select-pet');
  }

  gotoNew(query = ''): ReturnType<Page['goto']> {
    return this.open(`/medical-records/new${query}`);
  }

  listHeading(): Locator {
    return this.heading('カルテ管理');
  }

  editHeading(): Locator {
    return this.heading('カルテ編集');
  }

  selectPetHeading(): Locator {
    return this.heading('カルテ作成 - ペット選択');
  }

  newButton(): Locator {
    return this.page.getByRole('button', { name: '新規カルテ登録' });
  }

  searchInput(): Locator {
    return this.page.getByPlaceholder('飼主名、ペット名、カルテNo、主訴で検索...');
  }

  hayashiText(): Locator {
    return this.page.getByText('林 文明').first();
  }

  irisText(): Locator {
    return this.page.getByText('Iris').first();
  }

  saveButton(): Locator {
    return this.page.getByRole('button', { name: '保存' });
  }
}
