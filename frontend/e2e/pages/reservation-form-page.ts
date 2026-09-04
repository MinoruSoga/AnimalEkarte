import { type Locator, type Page } from "@playwright/test";
import { BasePage } from "./base-page";

/**
 * New-reservation modal (`/reservations` → "新規予約登録") with its embedded
 * PatientSelectionTable. All result/input locators are scoped to the dialog to
 * avoid colliding with the calendar DOM behind the modal.
 */
export class ReservationFormPage extends BasePage {
  gotoReservations(): ReturnType<Page["goto"]> {
    return this.open("/reservations");
  }

  todayButton(): Locator {
    return this.page.getByRole("button", { name: "今日" });
  }

  newReservationButton(): Locator {
    return this.page.getByRole("button", { name: "新規予約登録" });
  }

  dialog(): Locator {
    return this.page.getByRole("dialog");
  }

  dialogTitle(): Locator {
    return this.dialog().getByText("新規予約作成");
  }

  patientSearchLabel(): Locator {
    return this.dialog().getByText("患者検索");
  }

  /** Server-side cross-field search (`#search`), scoped to the dialog. */
  patientSearchInput(): Locator {
    return this.dialog().locator("#search");
  }

  patientRow(name: string): Locator {
    return this.dialog()
      .getByRole("row")
      .filter({ has: this.page.getByText(name, { exact: true }) });
  }

  selectPatientButton(id: string, name: string): Locator {
    return this.dialog().getByRole("button", {
      name: `選択: ${name} (ID ${id})`,
      exact: true,
    });
  }

  selectedPatientButton(id: string, name: string): Locator {
    return this.dialog().getByRole("button", {
      name: `選択中: ${name} (ID ${id})`,
      exact: true,
    });
  }
}
