import { describe, expect, it } from "vitest";
import {
  treatmentPackageComponentInputSchema,
  treatmentPackageErrorSchema,
  treatmentPackageItemRefSchema,
  treatmentPackageListSchema,
  treatmentPackagePatchSchema,
  treatmentPackageSchema,
} from "./treatment-packages";

const uuid = "a0000000-0000-4000-8000-000000000011";

describe("wire schema（fail-closed・文字列契約）", () => {
  it("型別IDはcanonical文字列のみ。number化・他型FK混入・unknown keyを拒否", () => {
    expect(
      treatmentPackageItemRefSchema.safeParse({ kind: "medicine", medicine_id: "103" }).success,
    ).toBe(true);
    // number ID は拒否（number化しない）
    expect(
      treatmentPackageItemRefSchema.safeParse({ kind: "medicine", medicine_id: 103 }).success,
    ).toBe(false);
    // 他型FK混入は拒否
    expect(
      treatmentPackageItemRefSchema.safeParse({
        kind: "medicine",
        medicine_id: "103",
        vaccine_id: "1",
      }).success,
    ).toBe(false);
    // 未知kind/unknown keyは拒否
    expect(treatmentPackageItemRefSchema.safeParse({ kind: "other" }).success).toBe(false);
  });

  it("quantity は正のcanonical十進のみ（0・末尾0・負を拒否）", () => {
    const q = (v: string) =>
      treatmentPackageComponentInputSchema.safeParse({
        component_key: uuid,
        item: { kind: "consultation", consultation_id: "1" },
        quantity: v,
      }).success;
    expect(q("0.25")).toBe(true);
    expect(q("2")).toBe(true);
    expect(q("0")).toBe(false);
    expect(q("1.230")).toBe(false);
    expect(q("-1")).toBe(false);
  });

  it("PATCH は省略=維持・明示null拒否・version必須", () => {
    expect(treatmentPackagePatchSchema.safeParse({ version: "3" }).success).toBe(true);
    expect(treatmentPackagePatchSchema.safeParse({ version: "3", name: null }).success).toBe(false);
    expect(treatmentPackagePatchSchema.safeParse({ name: "x" }).success).toBe(false);
  });

  it("Package/List は契約外形を受理し余計な key を拒否", () => {
    const pkg = {
      id: "201",
      clinic_id: "7",
      version: "3",
      name: "n",
      description: "",
      is_active: true,
      sort_order: 0,
      components: [
        {
          component_key: uuid,
          item: { kind: "consultation", consultation_id: "1" },
          quantity: "1",
        },
      ],
      created_at: "t",
      updated_at: "t",
    };
    expect(treatmentPackageSchema.safeParse(pkg).success).toBe(true);
    expect(treatmentPackageListSchema.safeParse({ items: [], order_version: "17" }).success).toBe(
      true,
    );
    expect(treatmentPackageSchema.safeParse({ ...pkg, extra: 1 }).success).toBe(false);
  });

  it("error envelope は error/code(+issues) のみ", () => {
    expect(
      treatmentPackageErrorSchema.safeParse({ error: "x", code: "version-conflict" }).success,
    ).toBe(true);
    expect(treatmentPackageErrorSchema.safeParse({ error: "x", code: "bogus" }).success).toBe(
      false,
    );
  });
});
