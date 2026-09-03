import { describe, expect, it } from "vitest";

import type { InterviewTemplateFormData } from "../lib/interview-template-side-panel-model";
import {
  buildInterviewTemplateCreateRequest,
  buildInterviewTemplateUpdateRequest,
} from "./interview-template-settings-model";

const ACTIVE_FORM: InterviewTemplateFormData = {
  category: "chief_complaint",
  title: "V04テンプレ",
  content: "内容",
  isActive: true,
};

const INACTIVE_FORM: InterviewTemplateFormData = {
  category: "history",
  title: "無効テンプレ",
  content: "",
  isActive: false,
};

describe("buildInterviewTemplateCreateRequest", () => {
  it("既定 active の form data から is_active: true を含む payload を返す", () => {
    expect(buildInterviewTemplateCreateRequest(ACTIVE_FORM)).toEqual({
      category: "chief_complaint",
      title: "V04テンプレ",
      content: "内容",
      is_active: true,
    });
  });

  it("明示 inactive の form data からは is_active: false を返す", () => {
    expect(buildInterviewTemplateCreateRequest(INACTIVE_FORM)).toEqual({
      category: "history",
      title: "無効テンプレ",
      content: "",
      is_active: false,
    });
  });
});

describe("buildInterviewTemplateUpdateRequest", () => {
  it("update でも is_active を送信する（回帰）", () => {
    expect(buildInterviewTemplateUpdateRequest(ACTIVE_FORM)).toEqual({
      category: "chief_complaint",
      title: "V04テンプレ",
      content: "内容",
      is_active: true,
    });
    expect(buildInterviewTemplateUpdateRequest(INACTIVE_FORM)).toEqual({
      category: "history",
      title: "無効テンプレ",
      content: "",
      is_active: false,
    });
  });
});
