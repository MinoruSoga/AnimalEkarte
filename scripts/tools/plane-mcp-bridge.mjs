#!/usr/bin/env node
/**
 * Stdio-to-HTTP Plane MCP Bridge for Antigravity
 * Bridges JSON-RPC 2.0 messages from stdio to https://mcp.plane.so/http/api-key/mcp
 */
import fs from 'fs';
import os from 'os';
import path from 'path';
import readline from 'readline';

const keyFile = process.env.PLANE_MCP_KEY_FILE || path.join(os.homedir(), '.config/devin/plane-mcp-key');
let apiKey = '';
try {
  apiKey = fs.readFileSync(keyFile, 'utf-8').trim();
} catch (e) {
  process.stderr.write(`Failed to read key file: ${keyFile}\n`);
  process.exit(1);
}

const endpoint = 'https://mcp.plane.so/http/api-key/mcp';
const workspace = 'baritechllc';

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false
});

function writeJsonRpcError(id, code, message) {
  process.stdout.write(JSON.stringify({ jsonrpc: '2.0', id, error: { code, message } }) + '\n');
}

function isObject(v) {
  return v !== null && typeof v === 'object' && !Array.isArray(v);
}

function isValidId(v) {
  return v === null || typeof v === 'string' || (typeof v === 'number' && Number.isFinite(v));
}

function isRpcResponse(v, expectedId) {
  return isObject(v)
    && v.jsonrpc === '2.0'
    && Object.hasOwn(v, 'id')
    && v.id === expectedId
    && Object.hasOwn(v, 'result') !== Object.hasOwn(v, 'error');
}

function isRpcNotification(v) {
  return isObject(v)
    && v.jsonrpc === '2.0'
    && !Object.hasOwn(v, 'id')
    && typeof v.method === 'string';
}

function isValidEnvelope(payload) {
  return isObject(payload)
    && payload.jsonrpc === '2.0'
    && typeof payload.method === 'string'
    && payload.method !== ''
    && (!Object.hasOwn(payload, 'id') || isValidId(payload.id));
}

rl.on('line', async (line) => {
  if (!line.trim()) return;
  let payload;
  try {
    payload = JSON.parse(line);
  } catch {
    writeJsonRpcError(null, -32700, 'Parse error');
    return;
  }
  if (!isValidEnvelope(payload)) {
    writeJsonRpcError(null, -32600, 'Invalid Request');
    return;
  }
  const isRequest = Object.hasOwn(payload, 'id');
  const id = isRequest ? payload.id : null;
  const respond = (s) => {
    if (isRequest) process.stdout.write(s + '\n');
  };
  const fail = (message) => {
    if (isRequest) writeJsonRpcError(id, -32603, message);
  };
  try {
    const res = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json, text/event-stream, */*',
        'Authorization': `Bearer ${apiKey}`,
        'X-API-Key': apiKey,
        'X-Workspace-slug': workspace
      },
      body: JSON.stringify(payload)
    });

    if (!res.ok) {
      fail(`Upstream HTTP error ${res.status}`);
      return;
    }

    const text = await res.text();
    // Parse SSE lines or raw json
    const dataLines = text.split('\n').filter((s) => s.startsWith('data: '));
    if (dataLines.length > 0) {
      const forwards = [];
      let matched = false;
      let malformed = false;
      for (const s of dataLines) {
        const jsonStr = s.slice(6);
        let msg;
        try {
          msg = JSON.parse(jsonStr);
        } catch {
          malformed = true;
          break;
        }
        if (isRpcResponse(msg, id)) {
          matched = true;
          forwards.push(jsonStr);
        } else if (isRpcNotification(msg)) {
          forwards.push(jsonStr);
        }
      }
      if (malformed) {
        fail('Upstream returned malformed SSE payload');
        return;
      }
      if (!matched) {
        fail('Upstream returned an invalid response body');
        return;
      }
      for (const jsonStr of forwards) respond(jsonStr);
      return;
    }
    const trimmed = text.trim();
    let msg;
    try {
      msg = JSON.parse(trimmed);
    } catch {
      msg = null;
    }
    if (isRpcResponse(msg, id)) {
      respond(trimmed);
      return;
    }
    fail('Upstream returned an invalid response body');
  } catch {
    process.stderr.write('Bridge upstream request failed\n');
    fail('Bridge upstream request failed');
  }
});
