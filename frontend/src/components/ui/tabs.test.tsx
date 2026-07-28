import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { Tabs, TabsList, TabsTrigger } from "./tabs";

describe("Tabs touch targets", () => {
  it("呼び出し側がh-9を指定してもlistとtriggerを44px以上に保つ", () => {
    render(
      <Tabs defaultValue="one">
        <TabsList className="h-9">
          <TabsTrigger value="one" className="h-9 w-8">
            One
          </TabsTrigger>
        </TabsList>
      </Tabs>,
    );

    expect(screen.getByRole("tablist")).toHaveClass("min-h-11");
    expect(screen.getByRole("tab", { name: "One" })).toHaveClass("min-h-11", "min-w-11");
  });
});
