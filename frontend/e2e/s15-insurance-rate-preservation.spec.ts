import { test, expect } from "@playwright/test";
import type { BrowserContext, Page } from "@playwright/test";
import { createAuthedContext } from "./helpers/context";

// S15 再現スクリプト: 会計保険負担割合 — 新規 50/70 のみ提示・既存 90/100 レガシー値保持
// 正本: docs/ops/testing/scenarios/S15-insurance-rate-preservation.md
// 実行: scripts/run-e2e.sh e2e/s15-insurance-rate-preservation.spec.ts（E2E_LOGIN_* はホスト env から供給）
// 証跡: stdout の "S15-EVIDENCE {...}" 行を reports/uat-2026-10-01/ へ保存する。
// fixture: owner+pet・waiting 会計（保険対象明細あり・手順2-4 の再計算表示用）・
// completed 会計（ratio=0.9 レガシー）・保険対象の未請求明細（MR + treatment + 医師確認）を run 内で作成。
// 会計・カルテレコードはシナリオ規約により取消しない（作成物は残る）。pet/owner は afterAll で削除を試みる。
// 手順6: 割合以外の変更は「お預かり金額」変更→「修正を保存する」→確定済み修正確認ダイアログ経由で
// 実 UI 保存経路を通す（PATCH merge で insurance_ratio が既存値保持されるかを検証）。
// 手順7: 正規の新規会計導線 /accounting/new → POST /accountings/complete（Idempotency-Key・
// expected_unbilled_revision は FE が自動付与）。waiting 会計の PATCH 確定は BE が拒否する設計のため使わない。

let context: BrowserContext;
let clinicId = "";
const createdOwnerIds: number[] = [];
const createdPetIds: number[] = [];
const createdBillingIds: number[] = [];
const cleanupFailures: string[] = [];

let petId = 0;
let stdAccId = 0; // 保険未設定・保険対象明細 1000円(税別10%) の waiting 会計（表示・再計算専用）
let legacyAccId = 0; // insurance_ratio=0.9 の completed 会計
let medicalRecordId = 0; // 手順7 用の draft カルテ（confirm 済み billing-confirmation 付き）

function clinicHeaders(): Record<string, string> {
  return clinicId ? { "X-Clinic-ID": clinicId } : {};
}

function evidence(caseName: string, expected: string, actual: string): void {
  console.log(`S15-EVIDENCE ${JSON.stringify({ case: caseName, expected, actual })}`);
}

async function jsonOrThrow(
  page: Page,
  method: string,
  path: string,
  body?: unknown,
  extraHeaders?: Record<string, string>,
) {
  const res = await page.request.fetch(path, {
    method,
    headers: {
      "Content-Type": "application/json",
      "X-Requested-With": "XMLHttpRequest",
      ...clinicHeaders(),
      ...extraHeaders,
    },
    data: body === undefined ? undefined : JSON.stringify(body),
  });
  if (!res.ok()) {
    const text = await res.text().catch(() => "");
    throw new Error(
      `fixture API failed: ${method} ${path} -> ${res.status()} ${text.slice(0, 200)}`,
    );
  }
  if (method === "DELETE") return null;
  return res.json();
}

const INSURANCE_ITEM = {
  category: "examination",
  name: "S15検証 保険対象項目",
  unit_price: 1000,
  quantity: 1,
  discount_rate: 0,
  discount_amount: 0,
  tax_type: "excluded",
  tax_rate: 0.1,
  is_insurance_applicable: true,
  source: "manual",
  sort_order: 0,
};

/** JST 基準の本日 00:00 ISO（scheduled_date は date ではなく datetime 必須）。 */
function jstTodayISO(): string {
  const jst = new Date(Date.now() + 9 * 3600 * 1000);
  return `${jst.toISOString().slice(0, 10)}T00:00:00+09:00`;
}

async function openInsuranceSelect(page: Page): Promise<string[]> {
  const trigger = page.getByRole("combobox", { name: "負担割合" });
  await expect(trigger).toBeVisible();
  await trigger.click();
  const options = page.getByRole("option");
  const count = await options.count();
  const labels: string[] = [];
  for (let i = 0; i < count; i++) {
    labels.push((await options.nth(i).innerText()).trim());
  }
  return labels;
}

test.beforeAll(async ({ browser }) => {
  context = await createAuthedContext(browser);
  const page = await context.newPage();
  const me = (await jsonOrThrow(page, "GET", "/api/v1/me")) as { main_clinic_id?: string };
  clinicId = (me.main_clinic_id ?? "").trim();
  if (!clinicId) throw new Error("main clinic not found in /v1/me");

  const species = (await jsonOrThrow(page, "GET", "/api/v1/masters/animal-species")) as Array<{
    id: number;
    name?: string;
  }>;
  const dog = species.find((s) => s.name === "犬") || species[0];
  if (!dog) throw new Error("animal species master is empty");

  const owner = (await jsonOrThrow(page, "POST", "/api/v1/owners", {
    owner_name: "検証一五 一郎",
    owner_name_kana: "ケンショウイチゴ イチロウ",
  })) as { id: number };
  createdOwnerIds.push(owner.id);
  const pet = (await jsonOrThrow(page, "POST", "/api/v1/pets", {
    owner_id: owner.id,
    animal_species_id: dog.id,
    name: "S15ポチ",
  })) as { id: number };
  petId = pet.id;
  createdPetIds.push(pet.id);

  // 標準 fixture: waiting 会計 + 保険対象明細（明細追加で BE が totals を再計算する）
  const std = (await jsonOrThrow(page, "POST", "/api/v1/accountings", {
    owner_id: owner.id,
    pet_id: petId,
    subtotal: 0,
    tax_total: 0,
    total_amount: 0,
    has_insurance: false,
    status: "waiting",
    scheduled_date: jstTodayISO(),
  })) as { id: number };
  stdAccId = std.id;
  createdBillingIds.push(std.id);
  await jsonOrThrow(page, "POST", "/api/v1/billing-items", {
    ...INSURANCE_ITEM,
    billing_id: std.id,
  });

  // レガシー fixture: POST /complete で payments.insurance_ratio=0.9 を持つ completed 会計。
  // PATCH は payment 系フィールドを拒否するため ratio は complete 経由でのみ設定できる。
  const unbilled = (await jsonOrThrow(
    page,
    "GET",
    `/api/v1/billing-items/unbilled-details?pet_id=${petId}`,
  )) as { revision?: string };
  // complete は Idempotency-Key（有効 UUID）必須。
  const legacy = (await jsonOrThrow(
    page,
    "POST",
    "/api/v1/accountings/complete",
    {
      owner_id: owner.id,
      pet_id: petId,
      scheduled_date: jstTodayISO(),
      has_insurance: true,
      insurance_ratio: 0.9,
      insurance_amount: 900,
      items: [INSURANCE_ITEM],
      payment_splits: [{ method: "cash", amount: 200, received_amount: 200, change_amount: 0 }],
      expected_unbilled_revision: unbilled.revision,
    },
    { "Idempotency-Key": crypto.randomUUID() },
  )) as { id: number };
  legacyAccId = legacy.id;
  createdBillingIds.push(legacy.id);

  // 手順7 fixture: 保険対象の未請求明細はカルテ処置（treatment.is_insurance）から集約される。
  // MR は draft で作成（確定済み MR の会計確認は BE が拒否）→ treatment → 医師確認 confirm で
  // bc.status=confirmed になり unbilled-details の候補へ出る。
  const mr = (await jsonOrThrow(page, "POST", "/api/v1/medical-records", {
    owner_id: String(owner.id),
    pet_id: String(petId),
    visit_date: jstTodayISO().slice(0, 10),
    visit_type: "revisit",
    status: "draft",
  })) as { id: number };
  medicalRecordId = mr.id;
  await jsonOrThrow(page, "POST", `/api/v1/medical-records/${mr.id}/treatments`, {
    item_type: "other",
    content: "S15検証 保険対象処置",
    unit_price: 1000,
    quantity: 1,
    is_insurance: true,
    status: "completed",
    sort_order: 0,
  });
  await jsonOrThrow(page, "POST", `/api/v1/medical-records/${mr.id}/billing-confirmation/confirm`, {
    memo: "S15 fixture: unbilled insurance item",
  });
  const unbilledAfter = (await jsonOrThrow(
    page,
    "GET",
    `/api/v1/billing-items/unbilled-details?pet_id=${petId}`,
  )) as { items?: Array<{ name?: string; is_insurance_applicable?: boolean }> };
  const unbilledNames = (unbilledAfter.items ?? []).map((i) => i.name).join(",");
  if (!unbilledNames.includes("S15検証 保険対象処置")) {
    throw new Error(`fixture unbilled treatment not visible: items=${unbilledNames}`);
  }

  evidence(
    "fixtures",
    "owner+pet, std waiting billing, legacy completed billing (ratio=0.9), confirmed MR + unbilled insurance treatment",
    `std=${stdAccId} legacy=${legacyAccId} mr=${medicalRecordId} unbilled=[${unbilledNames}]`,
  );
  await page.close();
});

test.afterAll(async () => {
  if (!context) return;
  const page = await context.newPage();
  // カルテはペット参照をブロックするため先に削除を試みる（会計明細に取り込まれた場合は 409 → 記録して残す）。
  if (medicalRecordId) {
    try {
      await jsonOrThrow(page, "DELETE", `/api/v1/medical-records/${medicalRecordId}`);
    } catch (error) {
      cleanupFailures.push(`medical-record:${medicalRecordId}`);
      console.log(`S15-CLEANUP-FAIL medical-record:${medicalRecordId} ${String(error)}`);
    }
  }
  for (const id of createdPetIds) {
    try {
      await jsonOrThrow(page, "DELETE", `/api/v1/pets/${id}`);
    } catch (error) {
      cleanupFailures.push(`pet:${id}`);
      console.log(`S15-CLEANUP-FAIL pet:${id} ${String(error)}`);
    }
  }
  for (const id of createdOwnerIds) {
    try {
      await jsonOrThrow(page, "DELETE", `/api/v1/owners/${id}`);
    } catch (error) {
      cleanupFailures.push(`owner:${id}`);
      console.log(`S15-CLEANUP-FAIL owner:${id} ${String(error)}`);
    }
  }
  await page.close();
  await context.close();
  // 会計レコード自体はシナリオ規約（作成レコードを無断取消しない）により残置。
  evidence(
    "cleanup",
    "pets+owners deleted; billings kept per scenario",
    `billings_kept=${createdBillingIds.join(",")} ` +
      (cleanupFailures.length === 0 ? "deleted" : `failed=${cleanupFailures.join(",")}`),
  );
});

test.describe("S15: 保険負担割合", () => {
  test("手順1: 新規会計の負担割合選択肢は 50%/70% のみ", async () => {
    const page = await context.newPage();
    await page.goto(`/accounting/new?petId=${petId}`);
    const toggle = page.getByRole("switch", { name: "ペット保険を利用" });
    await expect(toggle).toBeVisible();
    if ((await toggle.getAttribute("aria-checked")) !== "true") {
      await toggle.click();
    }
    const options = await openInsuranceSelect(page);
    evidence("new-options", "only 50%/70% offered (no 90/100)", `options=[${options.join(",")}]`);
    expect(options).toEqual(["50%", "70%"]);
    await page.close();
  });

  test("手順2-4: 50%/70% で保険負担額が再計算・保険なしは全額自己負担", async () => {
    const page = await context.newPage();
    await page.goto(`/accounting/${stdAccId}`);
    const toggle = page.getByRole("switch", { name: "ペット保険を利用" });
    await expect(toggle).toBeVisible();
    if ((await toggle.getAttribute("aria-checked")) !== "true") {
      await toggle.click();
    }
    // insurance target = 保険対象明細 1000円(税抜)。50% → 500円、70% → 700円。
    await expect(page.getByText("保険負担額")).toBeVisible();
    await expect(page.getByText("500 円")).toBeVisible();
    await openInsuranceSelect(page);
    await page.getByRole("option", { name: "70%" }).click();
    await expect(page.getByText("700 円")).toBeVisible();
    await toggle.click(); // 保険なし
    await expect(page.getByText("保険負担額")).toHaveCount(0);
    evidence("recalc", "50%→500円 / 70%→700円 / OFF→負担額なし", "observed");
    await page.close();
  });

  test("手順5-6: レガシー 90% がそのまま表示され、他項目変更後も丸められない", async () => {
    const page = await context.newPage();
    await page.goto(`/accounting/${legacyAccId}`);
    const trigger = page.getByRole("combobox", { name: "負担割合" });
    await expect(trigger).toBeVisible();
    await expect(trigger).toContainText("90%");
    const options = await openInsuranceSelect(page);
    evidence(
      "legacy-display",
      "existing 90% shown via legacy label map, not rounded",
      `options=[${options.join(",")}]`,
    );
    expect(options).toContain("90%");
    await page.keyboard.press("Escape");

    // 手順6: 割合以外の項目だけ UI 上で変更して保存。完了済み会計にはお預かり金額を
    // 200→300（お釣り=100）に変え、確定済み会計の修正理由を入力して「修正を保存する」→
    // 確認ダイアログ「修正する」で PATCH が走る。保険項目は未変更のため FE は保険フィールドを
    // 省略し、BE merge が既存 0.9 を保持する（この経路こそがシナリオの検証対象）。
    const received = page.locator("#payment-split-0-received");
    await expect(received).toBeVisible();
    // 保険適用済みの請求額（200 円）が表示されていることを確認してから他項目を変更する。
    await expect(page.locator("text=今回の請求金額").locator("..")).toContainText("200");
    await received.fill("300");
    await page.locator("#postCloseReason").fill("S15: 割合以外（お預かり金額）のみ変更");
    // UAT evidence: 更新 PATCH の実リクエスト/レスポンスを記録（保険フィールド省略=merge 保持の証明）。
    page.on("response", async (res) => {
      if (res.request().method() === "PATCH" && res.url().includes(`/accountings/${legacyAccId}`)) {
        const body = await res.text().catch(() => "<unreadable>");
        console.log(`S15-PATCH-RESPONSE status=${res.status()} body=${body.slice(0, 400)}`);
        console.log(`S15-PATCH-REQUEST body=${(res.request().postData() ?? "").slice(0, 600)}`);
      }
    });
    await page.getByRole("button", { name: "修正を保存する" }).click();
    await page.getByRole("button", { name: "修正する" }).click();
    await expect(page.getByText("会計を完了しました")).toBeVisible({ timeout: 15000 });
    await page.reload();
    await expect(trigger).toContainText("90%", { timeout: 15000 });
    const acc = (await jsonOrThrow(page, "GET", `/api/v1/accountings/${legacyAccId}`)) as {
      payments?: Array<{ insurance_ratio?: number; received_amount?: number }>;
    };
    evidence(
      "legacy-preserved",
      "ratio stays 0.9 after UI edit (received 200→300) + save + reload",
      `ratio=${acc.payments?.[0]?.insurance_ratio} received=${acc.payments?.[0]?.received_amount}`,
    );
    expect(acc.payments?.[0]?.insurance_ratio).toBe(0.9);
    await page.close();
  });

  test("手順7: 新規 50% で保存・確定 → 再読込で保持", async () => {
    const page = await context.newPage();
    // 正規の新規会計導線: /accounting/new は未請求明細（confirm 済みカルテの保険対象処置）を
    // 取り込み、確定は POST /accountings/complete の原子 command（FE が revision+冪等キー付与）。
    await page.goto(`/accounting/new?petId=${petId}`);
    await expect(page.getByRole("textbox", { name: /項目名: S15検証 保険対象処置/ })).toBeVisible({
      timeout: 15000,
    });
    const toggle = page.getByRole("switch", { name: "ペット保険を利用" });
    await expect(toggle).toBeVisible();
    if ((await toggle.getAttribute("aria-checked")) !== "true") {
      await toggle.click();
    }
    // 保険対象 1000円 → 50% = 保険負担額 500円、請求額 = 1100-500 = 600円。
    await expect(page.getByText("500 円")).toBeVisible();
    await page.getByRole("button", { name: "支払方法を追加" }).click();
    const received = page.locator("#payment-split-0-received");
    await expect(received).toBeVisible();
    await page.getByRole("button", { name: "丁度" }).click();
    // 新規会計画面は inline + sticky グループの2個の「会計を確定する」を描画するため first を取る。
    const submit = page.getByRole("button", { name: "会計を確定する" }).first();
    await expect(submit).toBeEnabled();
    await submit.click();
    // 新規会計の complete 成功は toast「会計を登録・完了しました」+ 詳細画面への遷移。
    await expect(page.getByText("会計を登録・完了しました")).toBeVisible({ timeout: 15000 });
    await page.waitForURL(/\/accounting\/\d+/, { timeout: 15000 });
    const newId = page.url().match(/\/accounting\/(\d+)/)?.[1] ?? "";
    expect(newId).not.toBe("");
    createdBillingIds.push(Number(newId));
    // 再読込後も負担割合 50%・保険負担額 500円・請求額が保存時と一致することを検証。
    await page.reload();
    const trigger = page.getByRole("combobox", { name: "負担割合" });
    await expect(trigger).toContainText("50%", { timeout: 15000 });
    await expect(page.getByText("500 円")).toBeVisible();
    const acc = (await jsonOrThrow(page, "GET", `/api/v1/accountings/${newId}`)) as {
      status?: string;
      total_amount?: number;
      payments?: Array<{
        insurance_ratio?: number;
        insurance_amount?: number;
        billing_amount?: number;
      }>;
    };
    evidence(
      "saved-50-preserved",
      "ratio=0.5 and amount=500 preserved after UI save+reload",
      `id=${newId} status=${acc.status} total=${acc.total_amount} ` +
        `ratio=${acc.payments?.[0]?.insurance_ratio} ins_amount=${acc.payments?.[0]?.insurance_amount} ` +
        `billing=${acc.payments?.[0]?.billing_amount}`,
    );
    expect(acc.status).toBe("completed");
    expect(acc.payments?.[0]?.insurance_ratio).toBe(0.5);
    expect(acc.payments?.[0]?.insurance_amount).toBe(500);
    await page.close();
  });
});
