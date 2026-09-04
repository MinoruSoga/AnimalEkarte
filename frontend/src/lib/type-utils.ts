import type React from "react";

/**
 * Type-safe utilities for narrowing string values from UI components
 * (shadcn Select, Tabs, ToggleGroup, RadioGroup etc.)
 * that return `string` in their `onValueChange` callbacks.
 */

/**
 * Type guard that checks if a string value is one of the allowed values.
 * Uses `.some()` to avoid `as` type assertions internally.
 */
export function isOneOf<T extends string>(value: string, values: readonly T[]): value is T {
  return values.some((v) => v === value);
}

/**
 * Creates a typed onValueChange handler that validates and narrows
 * the string value before passing it to the setter.
 * Invalid values are silently ignored (no-op).
 *
 * @example
 * const VIEW_VALUES = ["month", "week"] as const;
 * type CalendarView = (typeof VIEW_VALUES)[number];
 *
 * <Select onValueChange={typedSetter(setView, VIEW_VALUES)}>
 */
export function typedSetter<T extends string>(
  setter: ((value: T) => void) | React.Dispatch<React.SetStateAction<T>>,
  validValues: readonly T[],
): (value: string) => void {
  return (value: string) => {
    if (isOneOf(value, validValues)) {
      setter(value);
    }
  };
}
