const PROJECT_RE = /^animalekarte-a4-[a-z0-9][a-z0-9-]{0,39}$/;
const RUN_ID_RE = /^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$/;
const COMMIT_RE = /^[a-f0-9]{40}$/;
const EXPECTED_PORTS = Object.freeze({
  db: 5432,
  backend: 8080,
  frontend: 3000,
});

function assertContract(condition, message) {
  if (!condition) throw new Error(`A4 Compose contract violation: ${message}`);
}

function labelsOf(resource) {
  return resource?.labels ?? {};
}

function validateService({ name, service, runId, revision, project }) {
  assertContract(service && typeof service === "object", `${name} service is required`);
  const labels = labelsOf(service);
  assertContract(
    labels["com.animalekarte.a4.disposable"] === "true",
    `${name} disposable label is required`,
  );
  assertContract(
    labels["com.animalekarte.a4.run-id"] === runId,
    `${name} run-id label must match the other services`,
  );
  assertContract(
    JSON.stringify(Object.keys(service.networks ?? {})) === JSON.stringify(["ekarte-network"]),
    `${name} must use only the dedicated network`,
  );
  assertContract(
    Array.isArray(service.ports) && service.ports.length === 1,
    `${name} must publish exactly one port`,
  );
  const [port] = service.ports;
  assertContract(
    Number(port.target) === EXPECTED_PORTS[name] && port.protocol === "tcp",
    `${name} published port is invalid`,
  );
  assertContract(
    port.host_ip === "127.0.0.1" || port.host_ip === "::1",
    `${name} port must bind to localhost`,
  );
  if (name !== "db") {
    assertContract(
      typeof service.image === "string" && service.image.startsWith(`${project}_${name}:`),
      `${name} image must be project-scoped`,
    );
    const imageRevision = service.build?.labels?.["org.opencontainers.image.revision"];
    assertContract(COMMIT_RE.test(imageRevision ?? ""), `${name} OCI revision must be a full commit`);
    assertContract(
      revision === undefined || revision === imageRevision,
      `${name} OCI revision must match the other application image`,
    );
    return imageRevision;
  }
  return revision;
}

export function validateA4ComposeConfig(config) {
  const project = config?.name;
  assertContract(PROJECT_RE.test(project ?? ""), "project name is invalid");
  const services = config?.services ?? {};
  assertContract(
    JSON.stringify(Object.keys(services).filter((name) => name !== "codegen").sort())
      === JSON.stringify(["backend", "db", "frontend"]),
    "runtime service set must be db, backend, and frontend",
  );

  const runId = labelsOf(services.db)["com.animalekarte.a4.run-id"];
  assertContract(RUN_ID_RE.test(runId ?? ""), "run ID is invalid");
  let revision = validateService({ name: "db", service: services.db, runId, project });
  revision = validateService({
    name: "backend", service: services.backend, runId, revision, project,
  });
  validateService({
    name: "frontend", service: services.frontend, runId, revision, project,
  });

  const databaseMounts = services.db.volumes ?? [];
  assertContract(
    databaseMounts.length === 1
      && databaseMounts[0].type === "volume"
      && databaseMounts[0].source === "postgres_data"
      && databaseMounts[0].target === "/var/lib/postgresql",
    "database must use exactly one named volume",
  );
  assertContract(services.backend.environment?.DB_HOST === "db", "backend DB_HOST must be db");
  assertContract(
    typeof services.backend.environment?.DB_NAME === "string"
      && services.backend.environment.DB_NAME === services.db.environment?.POSTGRES_DB,
    "backend and database names must match",
  );

  const network = config.networks?.["ekarte-network"];
  assertContract(network?.name === `${project}_ekarte-network`, "network must be project-scoped");
  assertContract(network.internal === true, "network must be internal");
  assertContract(network.attachable !== true, "network must not be attachable");
  assertContract(
    labelsOf(network)["com.animalekarte.a4.disposable"] === "true",
    "network disposable label is required",
  );

  const volume = config.volumes?.postgres_data;
  assertContract(
    volume?.name === `${project}_postgres_data`,
    "database volume must be project-scoped",
  );
  assertContract(
    labelsOf(volume)["com.animalekarte.a4.disposable"] === "true",
    "database volume disposable label is required",
  );
  for (const name of ["frontend_node_modules", "go_mod_cache", "go_build_cache"]) {
    const cache = config.volumes?.[name];
    assertContract(cache?.name === `${project}_${name}`, `${name} volume must be project-scoped`);
    assertContract(
      labelsOf(cache)["com.animalekarte.a4.disposable"] === "true",
      `${name} disposable label is required`,
    );
  }

  return {
    composeProject: project,
    runId,
    targetReleaseCommit: revision,
    networkName: network.name,
    databaseVolumeName: volume.name,
  };
}
