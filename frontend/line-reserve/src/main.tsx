import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { ErrorBoundary } from '@/shared-liff/ErrorBoundary';
import { ErrorPage, type ErrorPageTheme } from '@/shared-liff/ErrorPage';
import './index.css';
import { App } from './App';

const ERROR_PAGE_THEME: ErrorPageTheme = {
  bg: 'bg-noah-teal-light',
  heading: 'text-noah-teal-dark',
  body: 'text-noah-text-sub',
  button: 'bg-noah-teal',
  buttonHover: 'hover:bg-noah-teal-dark',
};

const rootElement = document.getElementById('root');
if (!rootElement) {
  throw new Error('Root element not found');
}

createRoot(rootElement).render(
  <StrictMode>
    <ErrorBoundary
      fallback={
        <ErrorPage
          message="エラーが発生しました。お手数ですが、もう一度開き直してください。"
          theme={ERROR_PAGE_THEME}
        />
      }
    >
      <App />
    </ErrorBoundary>
  </StrictMode>,
);
