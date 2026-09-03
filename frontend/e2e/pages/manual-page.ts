import { type Locator, type Page } from "@playwright/test";
import { BasePage } from "./base-page";

/** Manual screens (`/manual`, `/manual/:category/:slug`). */
export class ManualPage extends BasePage {
  gotoIndex(): ReturnType<Page["goto"]> {
    return this.open("/manual");
  }

  /** `path` is the `:category/:slug` segment, e.g. `screens/03-owners`. */
  gotoArticle(path: string): ReturnType<Page["goto"]> {
    return this.open(`/manual/${path}`);
  }

  overviewHeading(): Locator {
    return this.heading("Animal Ekarte 取扱説明書", 1);
  }

  sidebar(): Locator {
    return this.page.getByRole("navigation", { name: "マニュアル目次" });
  }

  sidebarLink(name: string): Locator {
    return this.sidebar().getByRole("link", { name });
  }

  /** All sidebar links (used for count assertions). */
  sidebarLinks(): Locator {
    return this.sidebar().getByRole("link");
  }

  firstSidebarLink(): Locator {
    return this.sidebar().getByRole("link").first();
  }

  categoryTab(name: string): Locator {
    return this.page.getByRole("tab", { name });
  }

  searchInput(): Locator {
    return this.page.getByLabel("マニュアル内検索");
  }
}
