import '@testing-library/jest-dom';
import { beforeAll, afterEach, afterAll } from 'vitest';
import { server } from './mocks/node';

// Start server before all tests
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));

//  Reset handlers after each test `important for test isolation`
afterEach(() => server.resetHandlers());

//  Close server after all tests
afterAll(() => server.close());
