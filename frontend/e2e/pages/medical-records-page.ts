import { type Locator, type Page } from "@playwright/test";
import { BasePage } from "./base-page";

/** Medical record screens (`/medical-records`). Shared by clinical-flows + medical-records-create. */
export class MedicalRecordsPage extends BasePage {
  gotoList(): ReturnType<Page["goto"]> {
    return this.open("/medical-records");
  }

  gotoSelectPet(): ReturnType<Page["goto"]> {
    return this.open("/medical-records/select-pet");
  }

  gotoNew(query = ""): ReturnType<Page["goto"]> {
    return this.open(`/medical-records/new${query}`);
  }

  listHeading(): Locator {
    return this.heading("カルテ管理");
  }

  editHeading(): Locator {
    return this.heading("カルテ編集");
  }

  selectPetHeading(): Locator {
    return this.heading("カルテ作成 - ペット選択");
  }

  patientSearchInput(): Locator {
    return this.page.locator("#search");
  }

  patientRow(name: string): Locator {
    return this.page.getByRole("row").filter({ has: this.page.getByText(name, { exact: true }) });
  }

  selectPatientButton(id: string, name: string): Locator {
    return this.page.getByRole("button", {
      name: `選択: ${name} (ID ${id})`,
      exact: true,
    });
  }

  newButton(): Locator {
    return this.page.getByRole("button", { name: "新規カルテ登録" });
  }

  searchInput(): Locator {
    // Placeholder includes treatment/memo fields; match the stable prefix.
    return this.page.getByPlaceholder(/飼主名、ペット名、カルテNo/);
  }

  ownerText(name: string): Locator {
    return this.page.getByText(name, { exact: false }).first();
  }

  petText(name: string): Locator {
    return this.page.getByText(name, { exact: false }).first();
  }

  /** DataTable rows are non-interactive; open detail via the pet-name link. */
  firstDetailLink(): Locator {
    return this.page.getByRole("link", { name: /カルテ詳細:/ }).first();
  }

  saveButton(): Locator {
    return this.page.getByRole("button", { name: "保存" });
  }

  // ---- B-1 follow-up: server pagination / sort / filter (AC-3) ----

  /** ページネーションの「2」ページ番号ボタン。 */
  pageButton(page: number): Locator {
    return this.page.getByRole("button", { name: String(page), exact: true });
  }

  /** SortableHeader の `aria-label="{label}でソート"` に対応する列ヘッダボタン。 */
  sortHeader(label: "診療日" | "飼主名" | "ペット名" | "ステータス"): Locator {
    return this.page.getByRole("button", { name: `${label}でソート` });
  }

  filterAddButton(): Locator {
    return this.page.getByRole("button", { name: "フィルタを追加" });
  }

  filterRemoveButton(propertyLabel: string): Locator {
    return this.page.getByRole("button", { name: `${propertyLabel} フィルタを削除` });
  }
}
