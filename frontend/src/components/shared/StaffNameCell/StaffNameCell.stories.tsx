import type { ReactNode } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { StaffNameCell } from "./StaffNameCell";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/StaffNameCell",
  component: StaffNameCell,
  tags: ["autodocs"],
  // EMR-412 R8: 一覧の担当者セル。空欄 / 有効 / 退職等で無効 / 名前解決不可の4状態。
  parameters: {
    layout: "padded",
  },
} satisfies Meta<typeof StaffNameCell>;

type Story = StoryObj<typeof StaffNameCell>;

function TableStory({ children }: { children: ReactNode }) {
  return (
    <table>
      <tbody>
        <tr>{children}</tr>
      </tbody>
    </table>
  );
}

const isValid = (name: string) => name !== "退職 太郎";

/** 空欄。検証せず「-」だけを出す（X15 の記号）。 */
export const Blank: Story = {
  render: () => (
    <TableStory>
      <StaffNameCell staff="" isValidStaff={isValid} />
    </TableStory>
  ),
};

/** アクティブな担当者。氏名は ink（C.text）。 */
export const ValidStaff: Story = {
  render: () => (
    <TableStory>
      <StaffNameCell staff="対応 花子" isValidStaff={isValid} />
    </TableStory>
  ),
};

/** 退職等で無効になった担当者。「（無効）」を文字で併記し色警告に依存しない。 */
export const InactiveStaff: Story = {
  render: () => (
    <TableStory>
      <StaffNameCell staff="退職 太郎" isValidStaff={isValid} />
    </TableStory>
  ),
};

/** 名前解決できない担当者（doctor_id のみ等）。「担当者不明」。 */
export const UnresolvedStaff: Story = {
  render: () => (
    <TableStory>
      <StaffNameCell staff="担当者不明" isValidStaff={isValid} />
    </TableStory>
  ),
};
