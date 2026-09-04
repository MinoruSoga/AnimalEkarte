import { act, renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";

import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";

import { useCreateBillingConfirmation, useCreateBillingReturn } from "./billing-confirmation";

describe("billing confirmation mutation authority", () => {
  it("does not forward a caller-supplied confirmed_by", async () => {
    let receivedBody: unknown;
    server.use(
      http.post("*/v1/medical-records/:id/billing-confirmation/confirm", async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json({ id: 1 });
      }),
    );

    const { result } = renderHook(() => useCreateBillingConfirmation("42"), {
      wrapper: createTestWrapper(),
    });
    const legacyMutate = result.current.mutateAsync as unknown as (input: {
      confirmed_by: number;
      memo?: string;
    }) => Promise<unknown>;

    await act(async () => {
      await legacyMutate({
        confirmed_by: 999,
        memo: "confirmed",
      });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(receivedBody).toEqual({ memo: "confirmed" });
  });

  it("does not derive returned_by from a stale caller argument", async () => {
    let receivedBody: unknown;
    server.use(
      http.post("*/v1/medical-records/:id/billing-confirmation/return", async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json({ id: 1 });
      }),
    );

    const legacyHook = useCreateBillingReturn as unknown as (
      medicalRecordID: string,
      staleStaffID: number,
    ) => ReturnType<typeof useCreateBillingReturn>;
    const { result } = renderHook(() => legacyHook("42", 999), { wrapper: createTestWrapper() });

    await act(async () => {
      await result.current.mutateAsync({
        return_reason: "needs correction",
      });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(receivedBody).toEqual({
      return_reason: "needs correction",
    });
  });
});
