const ID_RE = /^[a-f0-9]{64}$/;
const IMAGE_RE = /^sha256:[a-f0-9]{64}$/;

function requireValue(condition, message) {
  if (!condition) throw new Error(`A4 runtime attestation failed: ${message}`);
}

export function buildA4RuntimeReport({
  applyReport,
  applyReportSha256,
  composeSummary,
  inspections,
  imageInspections,
  networkInspection,
  volumeInspection,
  generatedAt,
  emptyBandPreflight,
  backupRestorePreflight,
}) {
  requireValue(applyReport?.status === "PASS", "apply report must be PASS");
  requireValue(applyReport?.targetHost === "db", "apply targetHost must be db");
  requireValue(applyReport?.runId === composeSummary.runId, "apply run ID mismatch");
  requireValue(/^[a-f0-9]{64}$/.test(applyReportSha256 ?? ""), "apply report SHA-256 is invalid");
  requireValue(emptyBandPreflight === "PASS", "empty-band preflight must be PASS");
  requireValue(backupRestorePreflight === "PASS", "backup/restore preflight must be PASS");
  requireValue(Array.isArray(inspections) && inspections.length === 3, "exactly three containers are required");
  requireValue(Array.isArray(imageInspections) && imageInspections.length === 3, "exactly three images are required");

  const byService = new Map(inspections.map((inspection) => [
    inspection.Config?.Labels?.["com.docker.compose.service"],
    inspection,
  ]));
  requireValue(byService.size === 3, "container services must be unique");
  const roles = { backend: "backend", frontend: "frontend", database: "db" };
  const containerIds = {};
  const imageIds = {};
  const expectedPorts = { backend: "8080/tcp", frontend: "3000/tcp", db: "5432/tcp" };
  for (const [role, service] of Object.entries(roles)) {
    const inspection = byService.get(service);
    requireValue(inspection && ID_RE.test(inspection.Id ?? ""), `${service} container is missing`);
    requireValue(inspection.State?.Running === true, `${service} is not running`);
    requireValue(inspection.State?.Health?.Status === "healthy", `${service} is not healthy`);
    requireValue(
      inspection.Config?.Labels?.["com.animalekarte.a4.disposable"] === "true",
      `${service} disposable label is missing`,
    );
    requireValue(
      inspection.Config?.Labels?.["com.animalekarte.a4.run-id"] === composeSummary.runId,
      `${service} run ID label mismatch`,
    );
    requireValue(
      inspection.Config?.Labels?.["com.docker.compose.project"] === composeSummary.composeProject,
      `${service} Compose project mismatch`,
    );
    requireValue(inspection.HostConfig?.NetworkMode === composeSummary.networkName, `${service} network mode mismatch`);
    requireValue(
      JSON.stringify(Object.keys(inspection.NetworkSettings?.Networks ?? {}))
        === JSON.stringify([composeSummary.networkName]),
      `${service} must use only the dedicated network`,
    );
    const bindings = inspection.HostConfig?.PortBindings ?? {};
    requireValue(
      JSON.stringify(Object.keys(bindings)) === JSON.stringify([expectedPorts[service]]),
      `${service} published port mismatch`,
    );
    requireValue(
      bindings[expectedPorts[service]].every(({ HostIp }) => HostIp === "127.0.0.1" || HostIp === "::1"),
      `${service} ports must bind to localhost`,
    );
    requireValue(
      Date.parse(inspection.State?.StartedAt ?? "") <= Date.parse(generatedAt),
      `${service} start time is after the attestation`,
    );
    requireValue(IMAGE_RE.test(inspection.Image ?? ""), `${service} image ID is invalid`);
    containerIds[role] = inspection.Id;
    imageIds[role] = inspection.Image;
  }
  const imagesById = new Map(imageInspections.map((inspection) => [inspection.Id, inspection]));
  for (const role of ["backend", "frontend"]) {
    requireValue(
      imagesById.get(imageIds[role])?.Config?.Labels?.["org.opencontainers.image.revision"]
        === composeSummary.targetReleaseCommit,
      `${role} image revision mismatch`,
    );
  }

  requireValue(
    networkInspection?.Name === composeSummary.networkName
      && networkInspection.Internal === true
      && networkInspection.Attachable === false,
    "dedicated internal network mismatch",
  );
  requireValue(
    networkInspection.Labels?.["com.docker.compose.project"] === composeSummary.composeProject
      && networkInspection.Labels?.["com.animalekarte.a4.disposable"] === "true"
      && networkInspection.Labels?.["com.animalekarte.a4.run-id"] === composeSummary.runId,
    "dedicated network labels mismatch",
  );
  requireValue(
    JSON.stringify(Object.keys(networkInspection.Containers ?? {}).sort())
      === JSON.stringify(Object.values(containerIds).sort()),
    "dedicated network membership mismatch",
  );
  requireValue(
    volumeInspection?.Name === composeSummary.databaseVolumeName
      && volumeInspection.Scope === "local",
    "dedicated local database volume mismatch",
  );
  requireValue(
    volumeInspection.Labels?.["com.docker.compose.project"] === composeSummary.composeProject
      && volumeInspection.Labels?.["com.animalekarte.a4.disposable"] === "true"
      && volumeInspection.Labels?.["com.animalekarte.a4.run-id"] === composeSummary.runId,
    "dedicated database volume labels mismatch",
  );
  const database = byService.get("db");
  requireValue(
    database.Mounts?.length === 1
      && database.Mounts[0].Type === "volume"
      && database.Mounts[0].Name === composeSummary.databaseVolumeName
      && database.Mounts[0].Destination === "/var/lib/postgresql"
      && database.Mounts[0].RW === true,
    "database mount mismatch",
  );
  const envValue = (inspection, name) => (
    inspection.Config?.Env?.find((value) => value.startsWith(`${name}=`))?.slice(name.length + 1)
  );
  requireValue(envValue(database, "POSTGRES_DB") === applyReport.targetDatabase, "database target mismatch");
  requireValue(envValue(byService.get("backend"), "DB_HOST") === "db", "backend DB_HOST mismatch");
  requireValue(
    envValue(byService.get("backend"), "DB_NAME") === applyReport.targetDatabase,
    "backend target database mismatch",
  );
  requireValue(Number.isFinite(Date.parse(generatedAt)), "generatedAt is invalid");

  return {
    schemaVersion: 1,
    status: "PASS",
    generatedAt,
    clinicCode: applyReport.clinicCode,
    clinicOrdinal: applyReport.clinicOrdinal,
    runId: applyReport.runId,
    applyReportSha256,
    targetReleaseCommit: composeSummary.targetReleaseCommit,
    attestationMethod: "DOCKER_INSPECT_AND_GIT_HEAD",
    composeProject: composeSummary.composeProject,
    networkName: composeSummary.networkName,
    databaseVolumeName: composeSummary.databaseVolumeName,
    containerIds,
    backendImageDigest: imageIds.backend,
    frontendImageDigest: imageIds.frontend,
    databaseImageDigest: imageIds.database,
    databaseDisposition: "DISPOSABLE",
    networkIsolation: "LOCALHOST_ONLY",
    targetDatabaseEmptyBandPreflight: "PASS",
    backupRestorePreflight: "PASS",
  };
}
