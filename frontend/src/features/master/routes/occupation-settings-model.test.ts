import { describe, expect, it } from "vitest";

import type { OccupationFormData } from "../lib/occupation-side-panel-model";
import {
  buildOccupationCreateRequest,
  buildOccupationUpdateRequest,
} from "./occupation-settings-model";

function makeFormData(overrides: Partial<OccupationFormData> = {}): OccupationFormData {
  return {
    name: "看護師",
    description: "",
    isActive: true,
    ...overrides,
  };
}

describe("buildOccupationUpdateRequest", () => {
  it("sends empty description so a non-empty note can be cleared (EMR-448)", () => {
    const request = buildOccupationUpdateRequest(makeFormData({ description: "" }));
    expect(request).toHaveProperty("description", "");
  });

  it("sends edited description as is", () => {
    const request = buildOccupationUpdateRequest(makeFormData({ description: "夜間担当" }));
    expect(request.description).toBe("夜間担当");
  });
});

describe("buildOccupationCreateRequest", () => {
  it("keeps create omitting an empty description", () => {
    const request = buildOccupationCreateRequest(makeFormData({ description: "" }));
    expect(request.description).toBeUndefined();
  });
});
