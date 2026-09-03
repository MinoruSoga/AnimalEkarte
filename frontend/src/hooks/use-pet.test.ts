import { act, renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";

import { queryKeys } from "@/lib/query-keys";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";

import { useGetPets } from "./use-pet";

describe("useGetPets", () => {
  it("ページ番号・件数・検索語・動物種をAPIへ転送し、先頭20件外の患者とページ情報を返す", async () => {
    let capturedUrl: URL | undefined;
    server.use(
      http.get("/api/v1/pets", ({ request }) => {
        capturedUrl = new URL(request.url);
        return HttpResponse.json({
          data: [
            {
              id: 21,
              clinic_id: 1,
              owner_id: 42,
              name: "もも",
              pet_name_kana: "モモ",
              status: "alive",
            },
          ],
          total: 21,
          page: 2,
          limit: 20,
        });
      }),
    );

    const { result } = renderHook(
      () =>
        useGetPets(undefined, {
          page: 2,
          limit: 20,
          search: "もも",
          species: "2",
          includeDeceased: true,
        }),
      { wrapper: createTestWrapper() },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(capturedUrl?.searchParams.get("page")).toBe("2");
    expect(capturedUrl?.searchParams.get("limit")).toBe("20");
    expect(capturedUrl?.searchParams.get("search")).toBe("もも");
    expect(capturedUrl?.searchParams.get("species")).toBe("2");
    expect(capturedUrl?.searchParams.get("include_deceased")).toBe("true");
    expect(result.current.data?.[0]).toEqual(
      expect.objectContaining({
        id: "21",
        name: "もも",
        petNameKana: "モモ",
      }),
    );
    expect(result.current.total).toBe(21);
    expect(result.current.page).toBe(2);
    expect(result.current.limit).toBe(20);
  });

  it("引数なしの既存呼び出しは従来どおり一覧条件を送信しない", async () => {
    let capturedUrl: URL | undefined;
    server.use(
      http.get("/api/v1/pets", ({ request }) => {
        capturedUrl = new URL(request.url);
        return HttpResponse.json({ data: [] });
      }),
    );

    const { result } = renderHook(() => useGetPets(), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(capturedUrl?.search).toBe("");
    expect(result.current.data).toEqual([]);
  });

  it("飼主変更中は前の飼主のペットをplaceholderとして返さない", async () => {
    let releaseSecondOwner: (() => void) | undefined;
    const secondOwnerPending = new Promise<void>((resolve) => {
      releaseSecondOwner = resolve;
    });
    server.use(
      http.get("/api/v1/pets", async ({ request }) => {
        const ownerId = new URL(request.url).searchParams.get("owner_id");
        if (ownerId === "99") await secondOwnerPending;
        return HttpResponse.json({
          data: [
            {
              id: ownerId === "99" ? 99 : 42,
              clinic_id: 1,
              owner_id: Number(ownerId),
              name: ownerId === "99" ? "次の飼主のペット" : "前の飼主のペット",
              pet_name_kana: "",
              status: "alive",
            },
          ],
        });
      }),
    );
    let ownerId = "42";
    const { result, rerender } = renderHook(() => useGetPets(ownerId), {
      wrapper: createTestWrapper(),
    });
    await waitFor(() =>
      expect(result.current.data?.[0]?.ownerId).toBe("42"),
    );

    ownerId = "99";
    rerender();

    expect(result.current.data).toBeUndefined();
    await act(async () => {
      releaseSecondOwner?.();
    });
    await waitFor(() =>
      expect(result.current.data?.[0]?.ownerId).toBe("99"),
    );
  });

  it("死亡ペットを含める場合はAPIへinclude_deceased=trueを明示してstatusを変換する", async () => {
    let capturedUrl: URL | undefined;
    server.use(
      http.get("/api/v1/pets", ({ request }) => {
        capturedUrl = new URL(request.url);
        return HttpResponse.json({
          data: [
            {
              id: 7,
              clinic_id: 1,
              owner_id: 42,
              name: "ポチ",
              status: "deceased",
              deceased_at: "2026-07-10T12:00:00+09:00",
            },
          ],
        });
      }),
    );

    const { result } = renderHook(
      () => useGetPets("42", { includeDeceased: true }),
      { wrapper: createTestWrapper() },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(capturedUrl?.searchParams.get("owner_id")).toBe("42");
    expect(capturedUrl?.searchParams.get("include_deceased")).toBe("true");
    expect(result.current.data?.[0]).toEqual(
      expect.objectContaining({ id: "7", status: "死亡" }),
    );
  });

  it("通常一覧と死亡ペットを含む一覧は異なるquery keyを使う", () => {
    expect(queryKeys.pets.list("42")).not.toEqual(
      queryKeys.pets.list("42", { includeDeceased: true }),
    );
  });

  it("通常一覧ではinclude_deceasedを送信しない", async () => {
    let capturedUrl: URL | undefined;
    server.use(
      http.get("/api/v1/pets", ({ request }) => {
        capturedUrl = new URL(request.url);
        return HttpResponse.json({ data: [] });
      }),
    );

    const { result } = renderHook(() => useGetPets("42"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(capturedUrl?.searchParams.get("include_deceased")).toBeNull();
  });
});
