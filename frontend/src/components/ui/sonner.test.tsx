import { render } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ToasterProps } from "sonner";

import { Toaster } from "./sonner";

// EMR-381: 左上固定のトーストがサイドナビ・ページヘッダと重なるため、
// 既定位置をヘッダ下の右上へ変更する。実際の描画位置は Sonner の position prop が
// 決めるため、ラッパーが Sonner へ渡す値を直接検証する（sonner 内部DOM表現は
// ライブラリ側の責務で、jsdom では描画を再現しない）。
const captured = vi.hoisted(() => ({
  position: undefined as ToasterProps["position"],
  offset: undefined as ToasterProps["offset"],
  mobileOffset: undefined as ToasterProps["mobileOffset"],
  duration: undefined as number | undefined,
  closeButton: undefined as boolean | undefined,
}));

vi.mock("sonner", async (importOriginal) => {
  const actual = await importOriginal<typeof import("sonner")>();
  return {
    ...actual,
    Toaster: (props: ToasterProps) => {
      captured.position = props.position;
      captured.offset = props.offset;
      captured.mobileOffset = props.mobileOffset;
      captured.duration = props.duration;
      captured.closeButton = props.closeButton;
      return null;
    },
  };
});

describe("Toaster (sonner) — EMR-381 通知配置", () => {
  beforeEach(() => {
    captured.position = undefined;
    captured.offset = undefined;
    captured.mobileOffset = undefined;
    captured.duration = undefined;
    captured.closeButton = undefined;
  });

  it("既定では top-right に通知を表示する（左上のサイドナビ領域を避ける）", () => {
    render(<Toaster />);

    expect(captured.position).toBe("top-right");
  });

  it("ヘッダ下の offset・duration・closeButton の既定を維持する", () => {
    render(<Toaster />);

    expect(captured.offset).toBe(72);
    expect(captured.duration).toBe(4000);
    expect(captured.closeButton).toBe(true);
  });

  // F003: モバイルは Sonner 既定 mobileOffset=16 でページヘッダと重なる。
  // desktop の offset=72 と同じヘッダクリアランスをモバイルにも適用する。
  it("モバイルでもヘッダクリアランスを確保する（mobileOffset=72）", () => {
    render(<Toaster />);

    expect(captured.mobileOffset).toBe(72);
  });

  it("呼び出し側の props は既定値を上書きできる", () => {
    render(<Toaster position="bottom-left" offset={16} mobileOffset={8} />);

    expect(captured.position).toBe("bottom-left");
    expect(captured.offset).toBe(16);
    expect(captured.mobileOffset).toBe(8);
  });
});
