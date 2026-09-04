import { describe, it, expect } from "vitest";

describe("MSW Setup", () => {
  it("should mock API calls correctly", async () => {
    const response = await fetch("/api/auth/me");
    const data = await response.json();

    expect(response.status).toBe(200);
    expect(data).toEqual({
      id: "1",
      name: "Test User",
      email: "test@example.com",
      role: "admin",
    });
  });
});
