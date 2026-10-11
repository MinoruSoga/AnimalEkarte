import { describe, expect, it } from "vitest";
import {
  transformReservationWithIntent,
  type ReservationApiResponse,
} from "./reservation-intent-response";

const response: ReservationApiResponse = {
  id: 42,
  clinic_id: 3,
  start_time: "2026-10-11T00:00:00Z",
  end_time: "2026-10-11T00:15:00Z",
  visit_type: "revisit",
  reservation_type_id: 1,
  is_designated: false,
  status: "pending",
  notes: "",
  created_at: "2026-10-11T00:00:00Z",
  updated_at: "2026-10-11T00:00:00Z",
  source: "manual",
  is_staff_delegated: false,
  customer_fields: null,
};

describe("intent response revision", () => {
  it("keeps the server revision verbatim instead of truncating its fractional seconds", () => {
    const revision = "2026-10-11T09:00:00.123456789+09:00";
    expect(transformReservationWithIntent({ ...response, updated_at: revision }).updatedAt).toBe(
      revision,
    );
  });
  it("does not invent a revision for an empty timestamp", () => {
    expect(
      transformReservationWithIntent({ ...response, updated_at: "" }).updatedAt,
    ).toBeUndefined();
  });
});
