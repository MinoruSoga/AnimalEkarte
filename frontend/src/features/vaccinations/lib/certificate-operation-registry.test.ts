import { QueryClient } from "@tanstack/react-query";
import { describe, expect, it } from "vitest";
import type { IssuePending } from "./certificate-issue-reducer";
import { certificateOperationRegistry, releaseReprintStore, reprintStoreFor, retainReprintStore } from "./certificate-operation-registry";

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

describe("再印刷 store の解放", () => {
  const entry = (hasKey: boolean) => ({
    state: { phase: "unknown" as const, notice: "unavailable" as const },
    pending: hasKey ? { editionId: "1", request: {} as never } : null,
    layout: null,
    checking: false,
  });

  it("未確定の key が無ければ離れるときに外す", () => {
    const registry = certificateOperationRegistry(new QueryClient());
    const store = reprintStoreFor(registry, 101);
    releaseReprintStore(registry, 101, store);
    expect(registry.reprintStores.has(101)).toBe(false);
  });

  it("未確定の key があれば残し、確定した時点で外す", () => {
    const registry = certificateOperationRegistry(new QueryClient());
    const store = reprintStoreFor(registry, 101);
    store.set(1, entry(true));
    releaseReprintStore(registry, 101, store);
    expect(registry.reprintStores.get(101)).toBe(store);
    store.set(1, entry(false));
    expect(registry.reprintStores.has(101)).toBe(false);
  });

  it("解放を予約した後に同じ store が再び使われたら予約を取り消し、使用中の store を消さない", () => {
    const registry = certificateOperationRegistry(new QueryClient());
    const store = reprintStoreFor(registry, 101);
    store.set(1, entry(true));
    releaseReprintStore(registry, 101, store); // 離脱（未確定なので残して予約）
    retainReprintStore(registry, 101, store); // 戻る
    store.set(1, entry(false)); // 確定
    expect(registry.reprintStores.get(101)).toBe(store);
    expect(reprintStoreFor(registry, 101)).toBe(store);
  });
});
