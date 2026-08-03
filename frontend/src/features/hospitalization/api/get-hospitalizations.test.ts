import { beforeEach, describe, expect, it, vi } from "vitest";
import { axios } from "@/lib/axios";
import {
  HOSPITALIZATION_FILTER_STATUS,
  HOSPITALIZATION_LIST_DEFAULT_LIMIT,
  HOSPITALIZATION_LIST_DEFAULT_PAGE,
  toHospitalizationWireStatus,
} from "../constants";
import { getHospitalizations } from "./get-hospitalizations";

vi.mock("@/lib/axios", () => ({
  axios: {
    get: vi.fn(),
  },
}));

const mockedGet = vi.mocked(axios.get);

function mockPage(items: Array<{ id: number; status: string }>, total = items.length) {
  mockedGet.mockResolvedValueOnce({
    data: {
      data: items.map((item) => ({
        id: item.id,
        clinic_id: 1,
        pet_id: 10,
        owner_id: 20,
        start_date: "2026-03-25T00:00:00Z",
        end_date: "",
        status: item.status,
        hospitalization_type: "hospitalization",
        memo: "",
        owner_request: "",
        staff_notes: "",
        created_at: "2026-03-25T00:00:00Z",
        updated_at: "2026-03-25T00:00:00Z",
      })),
      total,
      page: HOSPITALIZATION_LIST_DEFAULT_PAGE,
      limit: HOSPITALIZATION_LIST_DEFAULT_LIMIT,
    },
  });
}

describe("toHospitalizationWireStatus (BUG-009 tab → wire mapping)", () => {
  it("maps active → admitted", () => {
    expect(toHospitalizationWireStatus(HOSPITALIZATION_FILTER_STATUS.ACTIVE)).toBe("admitted");
  });
  it("maps reserved → reserved", () => {
    expect(toHospitalizationWireStatus(HOSPITALIZATION_FILTER_STATUS.RESERVED)).toBe("reserved");
  });
  it("maps discharged → discharged", () => {
    expect(toHospitalizationWireStatus(HOSPITALIZATION_FILTER_STATUS.DISCHARGED)).toBe("discharged");
  });
  it("maps all → undefined (no status param)", () => {
    expect(toHospitalizationWireStatus(HOSPITALIZATION_FILTER_STATUS.ALL)).toBeUndefined();
  });
});

describe("getHospitalizations request params (BUG-009)", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("sends status=admitted for active tab with default page/limit", async () => {
    mockPage([{ id: 1, status: "admitted" }], 5);
    await getHospitalizations({ statusFilter: HOSPITALIZATION_FILTER_STATUS.ACTIVE });
    expect(mockedGet).toHaveBeenCalledWith("/v1/hospitalizations", {
      params: {
        page: HOSPITALIZATION_LIST_DEFAULT_PAGE,
        limit: HOSPITALIZATION_LIST_DEFAULT_LIMIT,
        status: "admitted",
      },
    });
  });

  it("sends status=reserved for reserved tab", async () => {
    mockPage([{ id: 2, status: "reserved" }], 3);
    await getHospitalizations({ statusFilter: HOSPITALIZATION_FILTER_STATUS.RESERVED });
    expect(mockedGet).toHaveBeenCalledWith("/v1/hospitalizations", {
      params: expect.objectContaining({ status: "reserved" }),
    });
  });

  it("sends status=discharged for discharged tab", async () => {
    mockPage([{ id: 3, status: "discharged" }], 1);
    await getHospitalizations({ statusFilter: HOSPITALIZATION_FILTER_STATUS.DISCHARGED });
    expect(mockedGet).toHaveBeenCalledWith("/v1/hospitalizations", {
      params: expect.objectContaining({ status: "discharged" }),
    });
  });

  it("omits status for all tab", async () => {
    mockPage([{ id: 1, status: "admitted" }, { id: 2, status: "reserved" }], 40);
    await getHospitalizations({ statusFilter: HOSPITALIZATION_FILTER_STATUS.ALL });
    const params = mockedGet.mock.calls[0]?.[1]?.params as Record<string, unknown>;
    expect(params).toEqual({
      page: HOSPITALIZATION_LIST_DEFAULT_PAGE,
      limit: HOSPITALIZATION_LIST_DEFAULT_LIMIT,
    });
    expect(params).not.toHaveProperty("status");
  });

  it("preserves server total (page-window contract)", async () => {
    mockPage([{ id: 1, status: "reserved" }], 37);
    const result = await getHospitalizations({
      statusFilter: HOSPITALIZATION_FILTER_STATUS.RESERVED,
    });
    expect(result.total).toBe(37);
    expect(result.data).toHaveLength(1);
    expect(result.page).toBe(HOSPITALIZATION_LIST_DEFAULT_PAGE);
    expect(result.limit).toBe(HOSPITALIZATION_LIST_DEFAULT_LIMIT);
  });

  it("forwards start_date/end_date with status", async () => {
    mockPage([{ id: 1, status: "admitted" }], 1);
    await getHospitalizations({
      statusFilter: HOSPITALIZATION_FILTER_STATUS.ACTIVE,
      startDate: "2026-01-01",
      endDate: "2026-01-31",
      page: 2,
      limit: 20,
    });
    expect(mockedGet).toHaveBeenCalledWith("/v1/hospitalizations", {
      params: {
        page: 2,
        limit: 20,
        status: "admitted",
        start_date: "2026-01-01",
        end_date: "2026-01-31",
      },
    });
  });
});
