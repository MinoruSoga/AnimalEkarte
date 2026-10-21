import { QueryClient } from "@tanstack/react-query";
import { describe, expect, it } from "vitest";
import type { IssuePending } from "./certificate-issue-reducer";
import { certificateOperationRegistry, reprintStoreFor } from "./certificate-operation-registry";

const pending = {
  key: "11111111-1111-4111-8111-111111111111",
  request: {},
} as unknown as IssuePending;

describe("未確定操作の登録簿", () => {
  it("同じ QueryClient では同じ登録簿・同じ接種記録では同じ再印刷 store を返す", () => {
    const client = new QueryClient();
    const registry = certificateOperationRegistry(client);
    registry.issuePending.set(101, pending);
    expect(certificateOperationRegistry(client).issuePending.get(101)).toBe(pending);
    expect(reprintStoreFor(registry, 101)).toBe(reprintStoreFor(registry, 101));
    expect(reprintStoreFor(registry, 101)).not.toBe(reprintStoreFor(registry, 102));
  });

  it("QueryClient が別なら共有しない（テスト間・利用者間で持ち越さない）", () => {
    const a = certificateOperationRegistry(new QueryClient());
    a.issuePending.set(101, pending);
    expect(certificateOperationRegistry(new QueryClient()).issuePending.size).toBe(0);
  });

  it("queryClient.clear()（ログアウト・医院切替）で消える", () => {
    const client = new QueryClient();
    certificateOperationRegistry(client).issuePending.set(101, pending);
    client.clear();
    expect(certificateOperationRegistry(client).issuePending.size).toBe(0);
  });
});
