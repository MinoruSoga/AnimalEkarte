import { expect, type Page } from "@playwright/test";

export type SyntheticRenderedAssertion =
  | { kind: "visibleText"; value: string }
  | { kind: "value"; selector: string; value: string }
  | { kind: "valueContains"; selector: string; value: string }
  | { kind: "containsText"; selector: string; value: string }
  | { kind: "selectedOption"; selector: string; value: string }
  | { kind: "clickTabIfVisible"; value: string }
  | { kind: "checked"; selector: string }
  | { kind: "disabled"; selector: string };

interface AuditResult {
  unnamed: string[];
  undersized: string[];
  overflow: number;
}

export async function inspectPage(page: Page): Promise<AuditResult> {
  return page.evaluate(() => {
    const unnamed: string[] = [];
    const undersized: string[] = [];
    const elements = document.querySelectorAll<HTMLElement>(
      "a[href], button, input:not([type='hidden']), select, textarea, [role='button'], [role='link'], [role='tab'], [role='switch']",
    );

    for (const element of elements) {
      const style = getComputedStyle(element);
      const rect = element.getBoundingClientRect();
      if (
        element.getAttribute("aria-hidden") === "true" ||
        style.display === "none" ||
        style.visibility === "hidden" ||
        rect.width === 0 ||
        rect.height === 0
      ) {
        continue;
      }

      const labelledBy = element.getAttribute("aria-labelledby");
      const labelledText = labelledBy
        ?.split(/\s+/)
        .map((id) => document.getElementById(id)?.textContent ?? "")
        .join(" ");
      const explicitLabels = element.id
        ? [...document.querySelectorAll<HTMLLabelElement>("label[for]")]
            .filter((label) => label.htmlFor === element.id)
            .map((label) => label.textContent ?? "")
            .join(" ")
        : "";
      const wrappingLabelElement = element.closest("label");
      const wrappingLabel = wrappingLabelElement?.textContent ?? "";
      const role = element.getAttribute("role");
      const supportsContentName =
        role !== "combobox" &&
        element.matches("a[href], button, [role='button'], [role='link'], [role='tab']");
      const inputValueName =
        element instanceof HTMLInputElement && ["button", "submit", "reset"].includes(element.type)
          ? element.value
          : "";
      const name = [
        element.getAttribute("aria-label"),
        labelledText,
        explicitLabels,
        wrappingLabel,
        supportsContentName ? element.textContent : "",
        inputValueName,
        element.getAttribute("title"),
      ].find((candidate) => candidate?.trim());
      const descriptor = [
        element.tagName.toLowerCase(),
        element.id ? `#${element.id}` : "",
        `[role=${role ?? "native"}]`,
        element.getAttribute("name") ? `[name=${element.getAttribute("name")}]` : "",
        element.getAttribute("type") ? `[type=${element.getAttribute("type")}]` : "",
        element.hasAttribute("aria-label") ? "[aria-label]" : "",
        element.hasAttribute("title") ? "[title]" : "",
        element.dataset.testid ? `[testid=${element.dataset.testid}]` : "",
        element.classList.length > 0
          ? `[class=${[...element.classList].slice(0, 6).join(".")}]`
          : "",
      ].join("");
      if (!name) unnamed.push(descriptor);

      let targetRect = rect;
      const usesAssociatedLabelTarget =
        role === "checkbox" ||
        role === "radio" ||
        role === "switch" ||
        (element instanceof HTMLInputElement &&
          (element.type === "checkbox" || element.type === "radio" || element.type === "file"));
      if (usesAssociatedLabelTarget) {
        const candidates = [
          rect,
          ...[...document.querySelectorAll<HTMLLabelElement>("label[for]")]
            .filter((label) => element.id !== "" && label.htmlFor === element.id)
            .map((label) => label.getBoundingClientRect()),
          ...(wrappingLabelElement ? [wrappingLabelElement.getBoundingClientRect()] : []),
        ];
        targetRect = candidates.reduce((largest, candidate) =>
          Math.min(candidate.width, candidate.height) > Math.min(largest.width, largest.height)
            ? candidate
            : largest,
        );
      }
      if (targetRect.width < 44 || targetRect.height < 44) {
        undersized.push(`${descriptor}:${Math.round(targetRect.width)}x${Math.round(targetRect.height)}`);
      }
    }

    return {
      unnamed: [...new Set(unnamed)],
      undersized: [...new Set(undersized)],
      overflow: document.documentElement.scrollWidth - document.documentElement.clientWidth,
    };
  });
}

export async function assertSyntheticRenderedState(
  page: Page,
  assertions: readonly SyntheticRenderedAssertion[],
  context: string,
): Promise<void> {
  for (const assertion of assertions) {
    if (assertion.kind === "visibleText") {
      await expect(
        page.getByText(assertion.value, { exact: false }).filter({ visible: true }).first(),
        `${context}: visible synthetic state`,
      ).toBeVisible();
      continue;
    }

    if (assertion.kind === "clickTabIfVisible") {
      const tab = page.getByRole("tab", { name: assertion.value }).filter({ visible: true });
      if (await tab.count()) {
        await tab.first().click();
      } else {
        const tabLabel = page.getByText(assertion.value, { exact: true }).filter({ visible: true });
        if (await tabLabel.count()) await tabLabel.first().click();
      }
      continue;
    }

    const locator = page.locator(assertion.selector).filter({ visible: true }).first();
    await expect(locator, `${context}: ${assertion.kind} ${assertion.selector}`).toBeVisible();
    if (assertion.kind === "value") {
      await expect(locator).toHaveValue(assertion.value);
    } else if (assertion.kind === "valueContains") {
      await expect(locator).toHaveValue(new RegExp(assertion.value));
    } else if (assertion.kind === "containsText") {
      await expect(locator).toContainText(assertion.value);
    } else if (assertion.kind === "selectedOption") {
      await locator.click();
      const option = page.getByRole("option", { name: assertion.value, exact: true });
      await expect(option).toHaveAttribute("data-state", "checked");
      await page.keyboard.press("Escape");
    } else if (assertion.kind === "checked") {
      await expect(locator).toBeChecked();
    } else {
      await expect(locator).toBeDisabled();
    }
  }
}
