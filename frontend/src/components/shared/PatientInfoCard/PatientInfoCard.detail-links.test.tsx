import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";

import { PatientInfoCard } from "./PatientInfoCard";

// EMR-174: 診療系ヘッダーの飼主名・ペット名から詳細画面へ遷移できるようにする。
// ownerDetailHref / petDetailHref が渡されたときだけ link を出す（未指定は従来表示のまま）。

vi.mock("@/assets/231a870df600a37e011a0e1140e7608b1f4c3340.png", () => ({ default: "/pet.png" }));

vi.mock("@/components/shared/Feedback", () => ({
  ImageWithFallback: ({
    src,
    alt,
    className,
  }: {
    src: string;
    alt: string;
    className?: string;
  }) => <img src={src} alt={alt} className={className} />,
}));

const baseProps = {
  ownerName: "山田 太郎",
  petName: "ポチ",
  petNumber: "0001",
  weight: "5.0kg",
};

function renderCard(props: React.ComponentProps<typeof PatientInfoCard>) {
  return render(
    <MemoryRouter>
      <PatientInfoCard {...props} />
    </MemoryRouter>,
  );
}

describe("PatientInfoCard detail links (EMR-174)", () => {
  it("ownerDetailHref / petDetailHref 指定時は飼主名・ペット名が link になる", () => {
    renderCard({
      ...baseProps,
      ownerDetailHref: "/owners/42",
      petDetailHref: "/owners/42?pet=7",
    });

    expect(screen.getByRole("link", { name: "飼主詳細を開く" })).toHaveAttribute(
      "href",
      "/owners/42",
    );
    expect(screen.getByRole("link", { name: "ペット詳細を開く" })).toHaveAttribute(
      "href",
      "/owners/42?pet=7",
    );
  });

  it("href 未指定時は従来どおり link を出さない", () => {
    renderCard({ ...baseProps });

    expect(screen.queryByRole("link", { name: "飼主詳細を開く" })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "ペット詳細を開く" })).not.toBeInTheDocument();
  });

  it("onOwnerClick が併存する場合は owner 側ボタンを優先し link は出さない", () => {
    const onOwnerClick = vi.fn();
    renderCard({ ...baseProps, onOwnerClick, ownerDetailHref: "/owners/42" });

    expect(screen.getByRole("button", { name: "山田 太郎" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "飼主詳細を開く" })).not.toBeInTheDocument();
  });
});
