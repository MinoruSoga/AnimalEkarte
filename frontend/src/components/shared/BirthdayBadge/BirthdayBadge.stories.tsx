import type { Meta, StoryObj } from "@storybook/react-vite";
import { BirthdayBadge } from "./BirthdayBadge";

// 誕生月判定は JST 当月なので、stories も JST の当月/前月を動的生成する
// （固定日付だと月が変わると表示・非表示が崩れる）。
const jstNow = new Date(Date.now() + 9 * 60 * 60 * 1000);
const fmt = (d: Date) =>
  `${d.getUTCFullYear()}-${String(d.getUTCMonth() + 1).padStart(2, "0")}-${String(
    d.getUTCDate(),
  ).padStart(2, "0")}`;
const thisMonth = fmt(new Date(Date.UTC(jstNow.getUTCFullYear(), jstNow.getUTCMonth(), 15)));
const prevMonth = fmt(new Date(Date.UTC(jstNow.getUTCFullYear(), jstNow.getUTCMonth() - 1, 15)));

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/BirthdayBadge",
  component: BirthdayBadge,
  tags: ["autodocs"],
  args: { subjectName: "ポチ", birthDate: thisMonth },
} satisfies Meta<typeof BirthdayBadge>;

type Story = StoryObj<typeof BirthdayBadge>;

export const Default: Story = {};

export const Static: Story = {
  args: { variant: "static", stopPropagation: true },
};

export const NotBirthMonth: Story = {
  args: { birthDate: prevMonth },
};
