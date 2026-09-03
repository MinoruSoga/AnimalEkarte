import { http, HttpResponse } from "msw";

/**
 * Common API handlers for MSW
 * Add more handlers as needed for specific features
 */
export const handlers = [
  // Health check or session info
  http.get("/api/auth/me", () => {
    return HttpResponse.json({
      id: "1",
      name: "Test User",
      email: "test@example.com",
      role: "admin",
    });
  }),

  // Add more global mocks here
];
