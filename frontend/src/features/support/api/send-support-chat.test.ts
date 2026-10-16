import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook } from "@testing-library/react";
import { createElement, type ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/lib/axios", () => ({
  axios: {
    post: vi.fn(async () => ({
      data: {
        reply: "ok",
        sources: [],
        retrieval: {
          mode: "keyword",
          reason: "disabled",
          manual_revision: "f".repeat(64),
        } satisfies SupportChatRetrieval,
      },
    })),
  },
}));
import { axios } from "@/lib/axios";
import { sendSupportChat, useSendSupportChat } from "./send-support-chat";
import type { SupportChatRetrieval } from "../types";

function wrapper({ children }: { children: ReactNode }) {
  return createElement(
    QueryClientProvider,
    { client: new QueryClient({ defaultOptions: { mutations: { retry: false } } }) },
    children,
  );
}

describe("sendSupportChat", () => {
  it("serializes only message/revision/candidates; signal stays in config", async () => {
    const ac = new AbortController();
    await sendSupportChat(
      {
        message: "m",
        manual_revision: "f".repeat(64),
        keyword_candidates: [{ category: "screens", slug: "s" }],
      },
      ac.signal,
    );
    const call = (axios.post as ReturnType<typeof vi.fn>).mock.calls[0];
    if (!call) throw new Error("expected axios call");
    expect(JSON.stringify(call[1])).not.toContain("signal");
    expect(Object.keys(call[1]).sort()).toEqual([
      "keyword_candidates",
      "manual_revision",
      "message",
    ]);
    expect(call[2].signal).toBe(ac.signal);
  });

  it("real mutation: mutateAsync keeps signal out of JSON body", async () => {
    const { result } = renderHook(() => useSendSupportChat(), { wrapper });
    const ac = new AbortController();
    await result.current.mutateAsync({
      message: "m",
      manual_revision: "f".repeat(64),
      keyword_candidates: [{ category: "screens", slug: "s" }],
      signal: ac.signal,
    });
    const call = (axios.post as ReturnType<typeof vi.fn>).mock.calls.at(-1);
    if (!call) throw new Error("expected axios call");
    expect(Object.keys(call[1]).sort()).toEqual([
      "keyword_candidates",
      "manual_revision",
      "message",
    ]);
    expect(JSON.stringify(call[1])).not.toContain("signal");
    expect(call[2].signal).toBe(ac.signal);
  });
});
