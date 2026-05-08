const storage = {
  get(key, fallback = "") {
    return window.localStorage.getItem(key) || fallback;
  },
  set(key, value) {
    window.localStorage.setItem(key, value);
  },
};

const resources = {
  idols: { label: "Idols", path: "/idols", searchParam: "name", title: "name", meta: ["id", "birthdate"] },
  groups: { label: "Groups", path: "/groups", searchParam: "name", title: "name", meta: ["id", "formation_date"] },
  agencies: { label: "Agencies", path: "/agencies", searchParam: "name", title: "name", meta: ["id", "country"] },
  events: { label: "Events", path: "/events", searchParam: "title", title: "title", meta: ["id", "event_type", "start_date_time"] },
  releases: { label: "Releases", path: "/releases", searchParam: "title", title: "title", meta: ["id", "release_type", "release_date"] },
  tags: { label: "Tags", path: "/tags", searchParam: "name", title: "name", meta: ["id", "category"] },
};

function apiBase() {
  return storage.get("idolApiBase", "/api/v1").replace(/\/$/, "");
}

function readKey() {
  return storage.get("idolReadKey");
}

function writeKey() {
  return storage.get("idolWriteKey");
}

function adminKey() {
  return storage.get("idolAdminKey");
}

function headers(token, json = false) {
  const h = {};
  if (token) h.Authorization = `Bearer ${token}`;
  if (json) h["Content-Type"] = "application/json";
  return h;
}

async function request(path, options = {}) {
  const response = await fetch(`${apiBase()}${path}`, options);
  const text = await response.text();
  let body = text;
  try {
    body = text ? JSON.parse(text) : null;
  } catch (_) {
    body = text;
  }
  if (!response.ok) {
    const message = body && body.message ? body.message : `HTTP ${response.status}`;
    throw new Error(message);
  }
  return body;
}

function pretty(value) {
  return JSON.stringify(value, null, 2);
}

function pickList(payload) {
  if (Array.isArray(payload)) return payload;
  if (Array.isArray(payload?.data)) return payload.data;
  const arrayKey = Object.keys(payload || {}).find((key) => Array.isArray(payload[key]));
  return arrayKey ? payload[arrayKey] : [];
}

function setText(id, value) {
  const el = document.getElementById(id);
  if (el) el.textContent = value;
}

function parseLines(value) {
  return value
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean);
}

function parseJSONField(value) {
  try {
    return JSON.parse(value);
  } catch (error) {
    throw new Error(`JSONの形式が不正です: ${error.message}`);
  }
}

function initSharedSettings() {
  const apiInput = document.getElementById("api-base");
  const readInput = document.getElementById("read-key");
  const writeInput = document.getElementById("write-key");
  const form = document.getElementById("settings-form");
  if (!form) return;

  apiInput.value = apiBase();
  readInput.value = readKey();
  writeInput.value = writeKey();

  form.addEventListener("submit", (event) => {
    event.preventDefault();
    storage.set("idolApiBase", apiInput.value || "/api/v1");
    storage.set("idolReadKey", readInput.value.trim());
    storage.set("idolWriteKey", writeInput.value.trim());
    loadResource(currentResource);
  });
}

let currentResource = "idols";

async function loadResource(resourceName = currentResource, query = "") {
  currentResource = resourceName;
  const config = resources[resourceName];
  const list = document.getElementById("resource-list");
  if (!list) return;

  document.querySelectorAll(".rail-item").forEach((item) => item.classList.toggle("active", item.dataset.resource === resourceName));
  setText("resource-title", config.label);
  setText("list-status", "読み込み中...");
  list.innerHTML = "";

  const params = new URLSearchParams({ page: "1", per_page: "12" });
  if (query) params.set(config.searchParam, query);

  try {
    const payload = await request(`${config.path}?${params.toString()}`, { headers: headers(readKey()) });
    const rows = pickList(payload);
    setText("list-status", `${rows.length}件を表示中`);
    list.innerHTML = rows.map((row) => renderDataCard(row, config)).join("");
  } catch (error) {
    setText("list-status", error.message);
  }
}

function renderDataCard(row, config) {
  const title = row[config.title] || row.title || row.id || "Untitled";
  const meta = config.meta
    .map((key) => row[key] ? `<span>${key}: ${escapeHTML(String(row[key]))}</span>` : "")
    .filter(Boolean)
    .join("<br />");
  const extra = row.tracks ? `<br /><span>tracks: ${row.tracks.length}</span>` : "";
  return `
    <article class="data-card">
      <h3>${escapeHTML(String(title))}</h3>
      <div class="meta">${meta}${extra}</div>
    </article>
  `;
}

function escapeHTML(value) {
  return value.replace(/[&<>"']/g, (char) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[char]);
}

function initAppPage() {
  initSharedSettings();

  document.querySelectorAll(".rail-item").forEach((item) => {
    item.addEventListener("click", () => loadResource(item.dataset.resource));
  });

  document.getElementById("search-form")?.addEventListener("submit", (event) => {
    event.preventDefault();
    loadResource(currentResource, document.getElementById("search-query").value.trim());
  });

  document.querySelectorAll(".tab").forEach((tab) => {
    tab.addEventListener("click", () => {
      document.querySelectorAll(".tab").forEach((t) => t.classList.toggle("active", t === tab));
      document.querySelectorAll(".form-surface").forEach((form) => form.classList.add("hidden"));
      document.getElementById(`${tab.dataset.form}-form`).classList.remove("hidden");
    });
  });

  document.getElementById("submission-form")?.addEventListener("submit", handleSubmission);
  document.getElementById("removal-form")?.addEventListener("submit", handleRemoval);
  document.getElementById("write-form")?.addEventListener("submit", handleWriteCreate);

  if (readKey()) loadResource("idols");
}

async function handleSubmission(event) {
  event.preventDefault();
  const data = new FormData(event.currentTarget);
  try {
    const payload = {
      target_type: data.get("target_type"),
      contributor_email: data.get("contributor_email"),
      source_urls: parseLines(data.get("source_urls")),
      payload: parseJSONField(data.get("payload")),
    };
    const result = await request("/submissions", {
      method: "POST",
      headers: headers(null, true),
      body: JSON.stringify(payload),
    });
    setText("submission-result", pretty(result));
  } catch (error) {
    setText("submission-result", error.message);
  }
}

async function handleRemoval(event) {
  event.preventDefault();
  const data = new FormData(event.currentTarget);
  try {
    const payload = Object.fromEntries(data.entries());
    const result = await request("/removal-requests", {
      method: "POST",
      headers: headers(null, true),
      body: JSON.stringify(payload),
    });
    setText("removal-result", pretty(result));
  } catch (error) {
    setText("removal-result", error.message);
  }
}

async function handleWriteCreate(event) {
  event.preventDefault();
  const data = new FormData(event.currentTarget);
  try {
    const result = await request(`/${data.get("resource")}`, {
      method: "POST",
      headers: headers(writeKey(), true),
      body: JSON.stringify(parseJSONField(data.get("payload"))),
    });
    setText("write-result", pretty(result));
  } catch (error) {
    setText("write-result", error.message);
  }
}

function initAdminSettings() {
  const apiInput = document.getElementById("admin-api-base");
  const keyInput = document.getElementById("admin-key");
  const form = document.getElementById("admin-settings-form");
  if (!form) return;

  apiInput.value = apiBase();
  keyInput.value = adminKey();
  form.addEventListener("submit", (event) => {
    event.preventDefault();
    storage.set("idolApiBase", apiInput.value || "/api/v1");
    storage.set("idolAdminKey", keyInput.value.trim());
    loadAdminDashboard();
  });
}

function initAdminPage() {
  initAdminSettings();
  document.querySelectorAll("[data-admin-refresh]").forEach((button) => {
    button.addEventListener("click", () => loadAdminSection(button.dataset.adminRefresh));
  });
  document.getElementById("apikey-form")?.addEventListener("submit", handleCreateAPIKey);
  document.getElementById("apikey-search-form")?.addEventListener("submit", handleSearchAPIKeys);
  document.getElementById("job-search-form")?.addEventListener("submit", handleJobSearch);
  if (adminKey()) loadAdminDashboard();
}

function loadAdminDashboard() {
  ["usage", "submissions", "removals"].forEach(loadAdminSection);
}

function loadAdminSection(section) {
  if (section === "usage") return loadUsage();
  if (section === "submissions") return loadPendingSubmissions();
  if (section === "removals") return loadPendingRemovals();
  return Promise.resolve();
}

async function loadUsage() {
  try {
    const payload = await request("/admin/analytics/usage?days=7", { headers: headers(adminKey()) });
    const rows = payload.data || [];
    const totals = rows.reduce(
      (acc, row) => {
        acc.requests += row.total_requests || 0;
        acc.errors += row.error_count || 0;
        return acc;
      },
      { requests: 0, errors: 0 },
    );
    document.getElementById("usage-summary").innerHTML = `
      <div class="metric-card"><span>Keys</span><strong>${rows.length}</strong></div>
      <div class="metric-card"><span>Requests</span><strong>${totals.requests}</strong></div>
      <div class="metric-card"><span>Errors</span><strong>${totals.errors}</strong></div>
    `;
    document.getElementById("usage-list").innerHTML = rows.map(renderUsageRow).join("") || `<div class="meta">データなし</div>`;
  } catch (error) {
    document.getElementById("usage-list").innerHTML = `<div class="meta">${escapeHTML(error.message)}</div>`;
  }
}

function renderUsageRow(row) {
  return `
    <div class="table-row">
      <strong>${escapeHTML(row.masked_key || "unknown")}</strong>
      <div class="meta">requests: ${row.total_requests || 0} / success: ${row.success_count || 0} / errors: ${row.error_count || 0}</div>
      <div class="meta">avg latency: ${Math.round(row.avg_latency_ms || 0)}ms / last used: ${escapeHTML(row.last_used_at || "-")}</div>
    </div>
  `;
}

async function handleCreateAPIKey(event) {
  event.preventDefault();
  const data = Object.fromEntries(new FormData(event.currentTarget).entries());
  try {
    const result = await request("/admin/apikeys", {
      method: "POST",
      headers: headers(adminKey(), true),
      body: JSON.stringify(data),
    });
    setText("apikey-result", pretty(result));
  } catch (error) {
    setText("apikey-result", error.message);
  }
}

async function handleSearchAPIKeys(event) {
  event.preventDefault();
  const email = new FormData(event.currentTarget).get("email");
  try {
    const result = await request(`/admin/apikeys?email=${encodeURIComponent(email)}`, { headers: headers(adminKey()) });
    setText("apikey-result", pretty(result));
  } catch (error) {
    setText("apikey-result", error.message);
  }
}

async function loadPendingSubmissions() {
  const target = document.getElementById("submission-review-list");
  try {
    const payload = await request("/submissions/pending", { headers: headers(adminKey()) });
    const rows = payload.submissions || payload.data || [];
    target.innerHTML = rows.map((row) => renderReviewCard(row, "submissions")).join("") || `<div class="meta">審査待ちはありません</div>`;
  } catch (error) {
    target.innerHTML = `<div class="meta">${escapeHTML(error.message)}</div>`;
  }
}

async function loadPendingRemovals() {
  const target = document.getElementById("removal-review-list");
  try {
    const payload = await request("/removal-requests/pending", { headers: headers(adminKey()) });
    const rows = payload.removal_requests || payload.data || [];
    target.innerHTML = rows.map((row) => renderReviewCard(row, "removal-requests")).join("") || `<div class="meta">保留中の削除申請はありません</div>`;
  } catch (error) {
    target.innerHTML = `<div class="meta">${escapeHTML(error.message)}</div>`;
  }
}

function renderReviewCard(row, resource) {
  const id = row.id || row.ID || "";
  const title = row.target_type || row.status || id;
  return `
    <article class="review-card">
      <h3>${escapeHTML(String(title))}</h3>
      <div class="meta">${escapeHTML(pretty(row))}</div>
      <div class="review-actions">
        <button class="success" data-status-update="${resource}" data-id="${escapeHTML(id)}" data-status="approved">承認</button>
        <button class="danger" data-status-update="${resource}" data-id="${escapeHTML(id)}" data-status="rejected">却下</button>
      </div>
    </article>
  `;
}

document.addEventListener("click", async (event) => {
  const button = event.target.closest("[data-status-update]");
  if (!button) return;
  const resource = button.dataset.statusUpdate;
  const id = button.dataset.id;
  const status = button.dataset.status;
  try {
    const path = resource === "submissions" ? `/${resource}/${id}/status` : `/${resource}/${id}`;
    await request(path, {
      method: "PUT",
      headers: headers(adminKey(), true),
      body: JSON.stringify({ status }),
    });
    if (resource === "submissions") loadPendingSubmissions();
    if (resource === "removal-requests") loadPendingRemovals();
  } catch (error) {
    window.alert(error.message);
  }
});

async function handleJobSearch(event) {
  event.preventDefault();
  const jobID = new FormData(event.currentTarget).get("job_id");
  try {
    const result = await request(`/admin/jobs/${encodeURIComponent(jobID)}`, { headers: headers(adminKey()) });
    setText("job-result", pretty(result));
  } catch (error) {
    setText("job-result", error.message);
  }
}

if (document.body.dataset.page === "admin") {
  initAdminPage();
} else {
  initAppPage();
}
