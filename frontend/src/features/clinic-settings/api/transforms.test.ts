import { describe, it, expect } from "vitest";
import { transformClinic } from "./transforms";
import type { Clinic } from "@/types/generated/models";

const minimal: Clinic = {
  id: 1,
  clinic_id: 1,
  name: "テスト動物病院",
  postal_code: "123-4567",
  address: "東京都新宿区1-1",
  phone_number: "03-1234-5678",
  fax_number: "03-1234-5679",
  registration_number: "REG-001",
  director_name: "山田院長",
  email: "test@example.com",
  website: "https://example.com",
  logo_url: null,
  is_active: true,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

describe("transformClinic", () => {
  it("id をそのまま返す", () => {
    expect(transformClinic({ ...minimal, id: 5 }).id).toBe(5);
  });

  it("name をそのまま返す", () => {
    expect(transformClinic({ ...minimal, name: "新病院" }).name).toBe("新病院");
  });

  it("postal_code を postalCode にマップする", () => {
    expect(transformClinic({ ...minimal, postal_code: "987-6543" }).postalCode).toBe("987-6543");
  });

  it("phone_number を phoneNumber にマップする", () => {
    expect(transformClinic({ ...minimal, phone_number: "06-9999-0000" }).phoneNumber).toBe(
      "06-9999-0000",
    );
  });

  it("fax_number を faxNumber にマップする", () => {
    expect(transformClinic({ ...minimal, fax_number: "06-9999-0001" }).faxNumber).toBe(
      "06-9999-0001",
    );
  });

  it("registration_number を registrationNumber にマップする", () => {
    expect(transformClinic(minimal).registrationNumber).toBe("REG-001");
  });

  it("director_name を directorName にマップする", () => {
    expect(transformClinic({ ...minimal, director_name: "新院長" }).directorName).toBe("新院長");
  });

  it("logo_url を logoUrl にマップする", () => {
    expect(transformClinic({ ...minimal, logo_url: "https://cdn/logo.png" }).logoUrl).toBe(
      "https://cdn/logo.png",
    );
  });

  it("logo_url が null のとき null を返す", () => {
    expect(transformClinic({ ...minimal, logo_url: null }).logoUrl).toBeNull();
  });

  it("is_active を isActive にマップする", () => {
    expect(transformClinic({ ...minimal, is_active: false }).isActive).toBe(false);
  });
});
