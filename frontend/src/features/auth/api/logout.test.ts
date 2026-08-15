import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/lib/axios", () => ({
  axios: {
    post: vi.fn(),
  },
}));

import { axios } from "@/lib/axios";
import { logout } from "./logout";

describe("logout", () => {
  beforeEach(() => {
    vi.mocked(axios.post).mockReset();
  });

  it("uses the refresh-cookie-compatible path so legacy and current cookies are both sent", async () => {
    vi.mocked(axios.post).mockResolvedValue({ data: {} });

    await logout();

    expect(axios.post).toHaveBeenCalledWith("/v1/auth/refresh/logout");
  });
});
