import { Component, type ReactNode } from "react";

interface Props {
  fallback: ReactNode;
  children: ReactNode;
}

interface State {
  hasError: boolean;
}

/**
 * liff / line-reserve 共通のエラーバウンダリ（FE6-4）。
 * レンダー時例外で画面が白紙化するのを防ぎ、fallback（ErrorPage）へ切り替える。
 * 外部パッケージは使わず React 標準のクラスコンポーネント API のみで実装する。
 */
export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false };

  static getDerivedStateFromError(): State {
    return { hasError: true };
  }

  componentDidCatch(error: unknown): void {
    if (import.meta.env.DEV) {
      console.error("[shared-liff] uncaught render error", error);
    }
  }

  render(): ReactNode {
    return this.state.hasError ? this.props.fallback : this.props.children;
  }
}
