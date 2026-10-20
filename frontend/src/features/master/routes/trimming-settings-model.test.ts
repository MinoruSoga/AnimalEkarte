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
    shortName: "",
    allowedWeekdays: [],
    imageObjectKey: "",
    imagePreviewUrl: "",
    imageChanged: false,
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

// EMR-457: 短縮名・画像・曜日（利用医院は既存の医院 scope）
describe("trimming course attributes", () => {
  it("omits unregistered optional attributes on create", () => {
    const request = buildTrimmingCourseCreateRequest(makeCourseForm());
    expect(request.short_name).toBeUndefined();
    expect(request.image_object_key).toBeUndefined();
    expect(request.allowed_weekdays).toBeUndefined();
  });

  it("sends trimmed short name, sorted weekdays and uploaded key on create", () => {
    const request = buildTrimmingCourseCreateRequest(
      makeCourseForm({
        shortName: "  シャン ",
        allowedWeekdays: [6, 1],
        imageObjectKey: "trimming-images/1/a.jpg",
        imageChanged: true,
      }),
    );
    expect(request.short_name).toBe("シャン");
    expect(request.allowed_weekdays).toEqual([1, 6]);
    expect(request.image_object_key).toBe("trimming-images/1/a.jpg");
  });

  it("clears short name and weekdays explicitly on update", () => {
    const request = buildTrimmingCourseUpdateRequest(makeCourseForm());
    expect(request.short_name).toBe("");
    expect(request.allowed_weekdays).toEqual([]);
  });

  it("omits an unchanged image on update and sends an empty key for removal", () => {
    const unchanged = buildTrimmingCourseUpdateRequest(
      makeCourseForm({ imageObjectKey: "trimming-images/1/a.jpg" }),
    );
    expect("image_object_key" in unchanged).toBe(false);

    const removed = buildTrimmingCourseUpdateRequest(
      makeCourseForm({ imageObjectKey: "", imageChanged: true }),
    );
    expect(removed.image_object_key).toBe("");
  });

  it("rereads stored attributes into the form", () => {
    const form = trimmingCourseToFormData({
      id: "1",
      name: "カット",
      price: null,
      targetSize: null,
      courseTypeId: null,
      duration: 60,
      description: "",
      isActive: false,
      shortName: "カ",
      allowedWeekdays: [0, 6],
      imageObjectKey: "trimming-images/1/a.jpg",
      imageUrl: "https://signed.example/a.jpg",
    } as TrimmingCourse);
    expect(form.shortName).toBe("カ");
    expect(form.allowedWeekdays).toEqual([0, 6]);
    expect(form.imageObjectKey).toBe("trimming-images/1/a.jpg");
    expect(form.imagePreviewUrl).toBe("https://signed.example/a.jpg");
    expect(form.imageChanged).toBe(false);

    const empty = trimmingCourseToFormData({
      id: "2",
      name: "B",
      shortName: null,
      allowedWeekdays: null,
      imageObjectKey: null,
      imageUrl: null,
    } as TrimmingCourse);
    expect(empty.shortName).toBe("");
    expect(empty.allowedWeekdays).toEqual([]);
    expect(empty.imageObjectKey).toBe("");
  });
});
