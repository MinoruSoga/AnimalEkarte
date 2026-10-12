const mode = process.env.PLANE_MCP_FAKE_FETCH || 'raw-json';
const MARKER = 'UPSTREAM-BODY-MARKER-9f3c';

const requestId = (options) => {
  try {
    const body = JSON.parse(options?.body ?? 'null');
    return body !== null && typeof body === 'object' && Object.hasOwn(body, 'id') ? body.id : null;
  } catch {
    return null;
  }
};

const okText = (body) => ({ ok: true, status: 200, text: async () => body });

globalThis.fetch = async (_url, options) => {
  const rid = JSON.stringify(requestId(options));
  switch (mode) {
    case 'http-401':
      return { ok: false, status: 401, text: async () => `unauthorized ${MARKER}` };
    case 'http-502':
      return { ok: false, status: 502, text: async () => `<html>${MARKER}</html>` };
    case 'http-429':
      return { ok: false, status: 429, text: async () => `rate limited ${MARKER}` };
    case 'fetch-throw':
      throw new Error(`fetch exploded ${MARKER}`);
    case 'body-throw':
      return { ok: true, status: 200, text: async () => { throw new Error(`body read ${MARKER}`); } };
    case 'non-json':
      return okText(`plain text ${MARKER}`);
    case 'malformed-sse':
      return okText(`data: {not json ${MARKER}\n\n`);
    case 'non-rpc-json':
      return okText(`{"ok":true,"marker":"${MARKER}"}`);
    case 'mismatched-id':
      return okText(`{"jsonrpc":"2.0","id":-999,"result":{"ok":true,"marker":"${MARKER}"}}`);
    case 'missing-id':
      return okText(`{"jsonrpc":"2.0","result":{"ok":true,"marker":"${MARKER}"}}`);
    case 'both-result-error':
      return okText(`{"jsonrpc":"2.0","id":${rid},"result":{},"error":{"code":-1,"message":"${MARKER}"}}`);
    case 'neither-result-error':
      return okText(`{"jsonrpc":"2.0","id":${rid},"marker":"${MARKER}"}`);
    case 'sse-mismatched-id':
      return okText('data: {"jsonrpc":"2.0","id":-999,"result":{"ok":true}}\n\n');
    case 'sse-notification-only':
      return okText('data: {"jsonrpc":"2.0","method":"notifications/progress"}\n\n');
    case 'sse-notify-then-response':
      return okText(
        'data: {"jsonrpc":"2.0","method":"notifications/progress"}\n\n'
        + `data: {"jsonrpc":"2.0","id":${rid},"result":{"ok":true}}\n\n`,
      );
    case 'sse':
      return okText(`event: message\ndata: {"jsonrpc":"2.0","id":${rid},"result":{"ok":true}}\n\n`);
    case 'raw-json':
    default:
      return okText(`{"jsonrpc":"2.0","id":${rid},"result":{"ok":true}}`);
  }
};
