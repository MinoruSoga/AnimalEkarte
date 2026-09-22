import { describe, expect, it } from "vitest";

import type { TrimmingCourse, TrimmingOption } from "../api/trimming";
import {
  trimmingCourseToFormData,
  trimmingOptionToFormData,
  type CourseFormData,
  type OptionFormData,
} from "../lib/trimming-side-panel-model";
import {
  buildTrimmingCourseCreateRequest,
  buildTrimmingCourseUpdateRequest,
  buildTrimmingOptionCreateRequest,
  buildTrimmingOptionUpdateRequest,
} from "./trimming-settings-model";

function makeCourseForm(overrides: Partial<CourseFormData> = {}): CourseFormData {
  return {
    name: "カット",
    price: "1200",
    targetSize: "medium",
    courseTypeId: "3",
    duration: "60",
    description: "",
    isActive: true,
    ...overrides,
  };
}

function makeOptionForm(overrides: Partial<OptionFormData> = {}): OptionFormData {
  return {
    name: "爪切り",
    price: "1200",
    duration: "15",
    combinable: true,
    description: "",
    isActive: true,
    ...overrides,
  };
}

describe("buildTrimmingCourseCreateRequest", () => {
  it("keeps empty price as null and zero price as 0", () => {
    const empty = buildTrimmingCourseCreateRequest(makeCourseForm({ price: "" }));
    expect(empty.price).toBeNull();

    const zero = buildTrimmingCourseCreateRequest(makeCourseForm({ price: "0" }));
    expect(zero.price).toBe(0);

    const priced = buildTrimmingCourseCreateRequest(makeCourseForm({ price: "1200" }));
    expect(priced.price).toBe(1200);
  });
});

describe("buildTrimmingCourseUpdateRequest", () => {
  it("persists edited course price 2300", () => {
    const request = buildTrimmingCourseUpdateRequest(makeCourseForm({ price: "2300" }));
    expect(request.price).toBe(2300);
  });
});

describe("buildTrimmingOptionCreateRequest", () => {
  it("keeps empty option price as null and zero as 0", () => {
    const empty = buildTrimmingOptionCreateRequest(makeOptionForm({ price: "" }));
    expect(empty.price).toBeNull();

    const zero = buildTrimmingOptionCreateRequest(makeOptionForm({ price: "0" }));
    expect(zero.price).toBe(0);
  });
});

describe("buildTrimmingOptionUpdateRequest", () => {
  it("persists edited option price 2300", () => {
    const request = buildTrimmingOptionUpdateRequest(makeOptionForm({ price: "2300" }));
    expect(request.price).toBe(2300);
  });
});

describe("trimming form reread", () => {
  it('maps null price to empty string and 0 to "0"', () => {
    const nullCourse = trimmingCourseToFormData({
      id: "1",
      name: "カット",
      price: null,
      targetSize: "medium",
      courseTypeId: "3",
      duration: 60,
      description: "",
      isActive: true,
    } as TrimmingCourse);
    expect(nullCourse.price).toBe("");

    const zeroCourse = trimmingCourseToFormData({
      id: "1",
      name: "カット",
      price: 0,
      targetSize: "medium",
      courseTypeId: "3",
      duration: 60,
      description: "",
      isActive: true,
    } as TrimmingCourse);
    expect(zeroCourse.price).toBe("0");

    const nullOption = trimmingOptionToFormData({
      id: "2",
      name: "爪切り",
      price: null,
      duration: 15,
      combinable: true,
      description: "",
      isActive: true,
    } as TrimmingOption);
    expect(nullOption.price).toBe("");
  });
});
