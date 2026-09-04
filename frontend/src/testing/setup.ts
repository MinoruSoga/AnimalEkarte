import "@testing-library/jest-dom";
import { beforeAll, afterEach, afterAll } from "vitest";
import { server } from "./mocks/node";

// jsdom polyfills for Radix UI primitives (Select / Dropdown / Popover etc).
// jsdom does not implement Pointer Events API or scrollIntoView, which Radix relies on
// internally when toggling open state via mouse interactions in tests.
if (typeof window !== "undefined") {
  if (!Element.prototype.hasPointerCapture) {
    Element.prototype.hasPointerCapture = () => false;
  }
  if (!Element.prototype.setPointerCapture) {
    Element.prototype.setPointerCapture = () => {};
  }
  if (!Element.prototype.releasePointerCapture) {
    Element.prototype.releasePointerCapture = () => {};
  }
  if (!Element.prototype.scrollIntoView) {
    Element.prototype.scrollIntoView = () => {};
  }
  // @radix-ui/react-use-size@1.1.1 uses ResizeObserver directly without a
  // typeof guard. jsdom does not provide it, so we stub a no-op implementation.
  if (!window.ResizeObserver) {
    window.ResizeObserver = class ResizeObserver {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  }
  // Sidebar.tsx and Radix UI call window.matchMedia for responsive breakpoints.
  // jsdom does not implement it; stub with a no-op that always returns matches=false.
  if (!window.matchMedia) {
    Object.defineProperty(window, "matchMedia", {
      writable: true,
      value: (query: string) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: () => {},
        removeListener: () => {},
        addEventListener: () => {},
        removeEventListener: () => {},
        dispatchEvent: () => false,
      }),
    });
  }
}

// Start server before all tests
beforeAll(() => server.listen({ onUnhandledRequest: "error" }));

//  Reset handlers after each test `important for test isolation`
afterEach(() => server.resetHandlers());

//  Close server after all tests
afterAll(() => server.close());
