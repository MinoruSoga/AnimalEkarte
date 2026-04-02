import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./app/index";
import "./index.css";

const root = createRoot(document.getElementById("root")!, {
  onCaughtError: (error: unknown, errorInfo: { componentStack?: string }) => {
    if (import.meta.env.DEV) {
      console.error("[ErrorBoundary]", error, errorInfo.componentStack);
    }
  },
  onUncaughtError: (error: unknown, errorInfo: { componentStack?: string }) => {
    if (import.meta.env.DEV) {
      console.error("[Uncaught]", error, errorInfo.componentStack);
    }
  },
  onRecoverableError: (error: unknown, errorInfo: { componentStack?: string }) => {
    if (import.meta.env.DEV) {
      console.error("[Recoverable]", error, errorInfo.componentStack);
    }
  },
});

root.render(
  <StrictMode>
    <App />
  </StrictMode>
);
