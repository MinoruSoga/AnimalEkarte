import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { ErrorBoundary } from '@/shared-liff/ErrorBoundary';
import { ErrorPage, type ErrorPageTheme } from '@/shared-liff/ErrorPage';
import './index.css';
import { App } from './App';

const ERROR_PAGE_THEME: ErrorPageTheme = {
  bg: 'bg-liff-brand-bg',
  heading: 'text-gray-800',
  body: 'text-gray-500',
  button: 'bg-liff-brand',
  buttonHover: 'hover:bg-liff-brand-dark',
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
