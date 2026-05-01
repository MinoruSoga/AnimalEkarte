import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const readSource = (path: string) =>
  readFileSync(resolve(process.cwd(), path), "utf8");

const cardSource = readSource("src/features/owners/components/LineIntegrationCard.tsx");
const confirmLineSource = readSource("src/features/owners/api/confirm-owner-line-id.ts");
const deliveryExclusionSource = readSource("src/features/owners/api/update-owner-delivery-exclusion.ts");
const transferStatusSource = readSource("src/features/owners/api/update-owner-transfer-status.ts");

describe("FEAT-382 LineIntegrationCard", () => {
  it("line-id-confirm hook calls the clinic-prefixed endpoint without a request body", () => {
    expect(confirmLineSource).toContain("/v1/clinics/${clinicId}/owners/${ownerId}/line-id-confirm");
    expect(confirmLineSource).toContain("axios.patch<OwnerApiResponse>");
    expect(confirmLineSource).toContain("transformOwner(data)");
    expect(confirmLineSource).not.toMatch(/line-id-confirm[^`]*`,\s*\{/);
  });

  it("all FEAT-381 mutation hooks update owner cache and invalidate owner-line-tags", () => {
    for (const source of [confirmLineSource, deliveryExclusionSource, transferStatusSource]) {
      expect(source).toContain('setQueryData(["owners", ownerId], owner)');
      expect(source).toContain('invalidateQueries({ queryKey: ["owners", ownerId] })');
      expect(source).toContain('invalidateQueries({ queryKey: ["owner-line-tags", ownerId] })');
    }
  });

  it("delivery-exclusion hook sends excluded and reason to the FEAT-381 endpoint", () => {
    expect(deliveryExclusionSource).toContain("/v1/clinics/${clinicId}/owners/${ownerId}/delivery-exclusion");
    expect(deliveryExclusionSource).toContain("excluded: body.excluded");
    expect(deliveryExclusionSource).toContain("reason: body.reason");
  });

  it("transfer-status hook sends is_transferred to the FEAT-381 endpoint", () => {
    expect(transferStatusSource).toContain("/v1/clinics/${clinicId}/owners/${ownerId}/transfer-status");
    expect(transferStatusSource).toContain("is_transferred: body.is_transferred");
  });

  it("delivery stopped banner covers owner flags, membership, and EXCL tag", () => {
    expect(cardSource).toContain('tags.includes("EXCL_配信停止")');
    expect(cardSource).toContain("owner?.deliveryExcluded");
    expect(cardSource).toContain("owner?.lstepOptOut");
    expect(cardSource).toContain("owner?.isTransferred");
    expect(cardSource).toContain('owner?.membershipType === "他診/準"');
    expect(cardSource).toContain("配信停止中");
  });

  it("delivery exclusion uses the new endpoint and caps the reason input at 100 characters", () => {
    expect(cardSource).not.toContain("useUpdateOwnerLstepOptOut");
    expect(cardSource).not.toContain("updateOptOut");
    expect(cardSource).toContain("maxLength={100}");
    expect(cardSource).toContain("deliveryReasonInputRef.current?.value.trim() || undefined");
    expect(cardSource).toContain("reason: null");
  });

  it("LINE ID confirmation UI is only enabled after LINE linkage", () => {
    expect(cardSource).toContain("is_linked && lineUserId");
    expect(cardSource).toContain("LINE ID 確認済み");
    expect(cardSource).toContain("LINE ID 未確認");
    expect(cardSource).toContain("confirmLineId()");
  });

  it("transfer-on flow is guarded by the specified confirmation copy", () => {
    expect(cardSource).toContain("setConfirmTransferOpen(true)");
    expect(cardSource).toContain("転院フラグを設定すると Lステップへの配信が停止されます。よろしいですか？");
  });
});
