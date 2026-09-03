import { describe, expect, it } from "vitest";

import type { ChiefComplaintFormData } from "../lib/chief-complaint-side-panel-model";
import {
  buildChiefComplaintCreateRequest,
  buildChiefComplaintUpdateRequest,
} from "./chief-complaint-settings-model";

const ACTIVE_FORM: ChiefComplaintFormData = {
  name: "V04主訴",
  description: "新規主訴",
  isActive: true,
};

const INACTIVE_FORM: ChiefComplaintFormData = {
  name: "無効主訴",
  description: "",
  isActive: false,
};

describe("buildChiefComplaintCreateRequest", () => {
  it("既定 active の form data から is_active: true を含む payload を返す", () => {
    expect(buildChiefComplaintCreateRequest(ACTIVE_FORM)).toEqual({
      name: "V04主訴",
      description: "新規主訴",
      is_active: true,
    });
  });

  it("明示 inactive の form data からは is_active: false を返す", () => {
    expect(buildChiefComplaintCreateRequest(INACTIVE_FORM)).toEqual({
      name: "無効主訴",
      description: undefined,
      is_active: false,
    });
  });
});

describe("buildChiefComplaintUpdateRequest", () => {
  it("update でも is_active を送信する（回帰）", () => {
    expect(buildChiefComplaintUpdateRequest(ACTIVE_FORM)).toEqual({
      name: "V04主訴",
      description: "新規主訴",
      is_active: true,
    });
    expect(buildChiefComplaintUpdateRequest(INACTIVE_FORM)).toEqual({
      name: "無効主訴",
      description: undefined,
      is_active: false,
    });
  });
});
