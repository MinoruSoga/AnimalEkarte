import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { ErrorBoundary } from "@/shared-liff/ErrorBoundary";
import { ErrorPage } from "@/shared-liff/ErrorPage";
import "./index.css";
import { App } from "./App";

const rootElement = document.getElementById("root");
if (!rootElement) {
  throw new Error("Root element not found");
}

createRoot(rootElement).render(
  <StrictMode>
    <ErrorBoundary
      fallback={
        <ErrorPage message="エラーが発生しました。お手数ですが、もう一度開き直してください。" />
      }
    >
      <App />
    </ErrorBoundary>
  </StrictMode>,
);
