import fs from "node:fs";
import path from "node:path";

const FORBIDDEN_DOCKER_ENV = Object.freeze([
  "DOCKER_CERT_PATH",
  "DOCKER_CONFIG",
  "DOCKER_CONTEXT",
  "DOCKER_HOST",
  "DOCKER_TLS",
  "DOCKER_TLS_VERIFY",
]);

export function sanitizedGitEnvironment(environment = process.env) {
  return Object.fromEntries(
    Object.entries(environment).filter(([key]) => !key.startsWith("GIT_")),
  );
}

export function rejectDockerEnvironmentOverrides(environment = process.env) {
  const configured = FORBIDDEN_DOCKER_ENV.filter((key) => environment[key]);
  if (configured.length > 0) {
    throw new Error(`Docker environment overrides are forbidden: ${configured.join(",")}`);
  }
}

export function validateLocalDockerEndpoint({ contextName, endpoint }) {
  if (!/^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$/.test(contextName ?? "")) {
    throw new Error("Docker context name is invalid");
  }
  if (!endpoint?.startsWith("unix:///")) {
    throw new Error("Docker context must use a local Unix socket");
  }
  const socketPath = endpoint.slice("unix://".length);
  if (!path.isAbsolute(socketPath)) {
    throw new Error("Docker Unix socket path must be absolute");
  }
  const socket = fs.statSync(socketPath);
  if (!socket.isSocket()) {
    throw new Error("Docker endpoint is not a Unix socket");
  }
  return Object.freeze({ contextName, endpoint });
}

export function validateLocalDockerAttestation({ contextName, endpoint, daemonId }) {
  validateLocalDockerEndpoint({ contextName, endpoint });
  if (!/^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$/.test(daemonId ?? "")) {
    throw new Error("Docker daemon identity is invalid");
  }
  return Object.freeze({ contextName, endpoint, daemonId });
}
