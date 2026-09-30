import type { Meta, StoryObj } from "@storybook/react-vite";
import { NextScheduleField } from "./NextScheduleField";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/NextScheduleField",
  component: NextScheduleField,
  tags: ["autodocs"],
  args: {
    typeId: "next-schedule-type",
    dateId: "next-schedule-date",
    scheduleType: "",
    nextDate: "",
    onScheduleTypeChange: () => {},
    onNextDateChange: () => {},
  },
} satisfies Meta<typeof NextScheduleField>;

type Story = StoryObj<typeof NextScheduleField>;

export const Empty: Story = {};

export const Filled: Story = {
  args: { scheduleType: "ワクチン", nextDate: "2027-03-15" },
};

export const WithError: Story = {
  args: { scheduleType: "ワクチン", nextDate: "", error: "次回予定日を入力してください" },
};
