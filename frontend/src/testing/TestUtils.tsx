import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";

/**
 * FE4-16: 28 テストファイルに私設定義されていた同型 wrapper（変種 (a)(b)(c)）の逐語一般化。
 * CODING_RULES §8.1 が言及する src/testing/TestUtils.tsx の実体化。
 *
 * (a) 素の QueryClientProvider: createTestWrapper()
 * (b) +MemoryRouter: createTestWrapper({ router: true })
 * (c) +initialEntries: createTestWrapper({ initialEntries: ["/aggregation"] })
 *
 * AuthContext 注入・カスタム Routes・hoisted spy を含む複雑な wrapper はスコープ外
 * （各ファイルで個別に private wrapper を維持する）。
 */
export interface CreateTestWrapperOptions {
  initialEntries?: string[];
  router?: boolean;
}

export function createTestWrapper(opts?: CreateTestWrapperOptions) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  const useRouter = opts?.router === true || opts?.initialEntries !== undefined;

  if (!useRouter) {
    return ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  }

  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={opts?.initialEntries}>{children}</MemoryRouter>
    </QueryClientProvider>
  );
}
