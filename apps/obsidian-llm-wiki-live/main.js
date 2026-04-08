const { Plugin, PluginSettingTab, Setting, Notice, requestUrl } = require("obsidian");
const fs = require("fs");
const os = require("os");
const path = require("path");

const MIRROR_ROOT_FOLDER = "LLM-Wiki";
const DEFAULT_SETTINGS = {
  autoRefreshSeconds: 15,
  lastLoadedAt: "",
  lastMirrorAt: ""
};

class LLMWikiSettingTab extends PluginSettingTab {
  constructor(app, plugin) {
    super(app, plugin);
    this.plugin = plugin;
  }

  display() {
    const { containerEl } = this;
    containerEl.empty();
    containerEl.createEl("h2", { text: "LLM-Wiki Sync" });

    new Setting(containerEl)
      .setName("Update interval seconds")
      .setDesc("How often the plugin syncs the current CLI tenant into Files. Set to 0 to disable background sync.")
      .addText((text) =>
        text
          .setPlaceholder("15")
          .setValue(String(this.plugin.settings.autoRefreshSeconds))
          .onChange(async (value) => {
            const parsed = Number(value);
            this.plugin.settings.autoRefreshSeconds = Number.isFinite(parsed) && parsed >= 0 ? parsed : 15;
            await this.plugin.saveSettings();
            this.plugin.restartSyncTimer();
          })
      );
  }
}

module.exports = class LLMWikiLivePlugin extends Plugin {
  async onload() {
    this.settings = Object.assign({}, DEFAULT_SETTINGS, await this.loadData());
    this.settings.lastLoadedAt = new Date().toISOString();
    await this.saveData(this.settings);

    this.statusBar = this.addStatusBarItem();
    this.setStatus("LLM-Wiki: syncing");

    this.addCommand({
      id: "sync-current-tenant-into-vault",
      name: "Sync current tenant into vault",
      callback: async () => {
        await this.syncNow(true);
      }
    });

    this.addSettingTab(new LLMWikiSettingTab(this.app, this));
    this.restartSyncTimer();
    await this.syncNow(false);
  }

  onunload() {
    this.clearSyncTimer();
  }

  async saveSettings() {
    await this.saveData(this.settings);
  }

  restartSyncTimer() {
    this.clearSyncTimer();
    const seconds = Number(this.settings.autoRefreshSeconds) || 0;
    if (seconds <= 0) {
      this.setStatus("LLM-Wiki: auto sync off");
      return;
    }
    this.syncTimer = window.setInterval(async () => {
      await this.syncNow(false);
    }, seconds * 1000);
  }

  clearSyncTimer() {
    if (this.syncTimer) {
      window.clearInterval(this.syncTimer);
      this.syncTimer = null;
    }
  }

  setStatus(text) {
    if (this.statusBar) {
      this.statusBar.setText(text);
    }
  }

  async syncNow(notify) {
    try {
      this.setStatus("LLM-Wiki: syncing");
      const result = await this.mirrorCurrentTenantToVault();
      this.setStatus(`LLM-Wiki: ${result.tenantID} ${result.changedCount}/${result.documentCount}`);
      if (notify) {
        new Notice(`LLM-Wiki synced ${result.changedCount} files in ${result.rootPath}`);
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      this.setStatus("LLM-Wiki: sync failed");
      new Notice(`LLM-Wiki sync failed: ${message}`);
      throw error;
    }
  }

  loadCurrentCLIProfile() {
    const configPath = path.join(os.homedir(), ".llm-wiki", "config.json");
    if (!fs.existsSync(configPath)) {
      throw new Error("~/.llm-wiki/config.json not found. Run llm-wiki auth login first.");
    }
    const raw = fs.readFileSync(configPath, "utf8");
    const parsed = JSON.parse(raw);
    const activeProfileName = (parsed.current_profile || "default").trim() || "default";
    const profile = parsed.profiles && parsed.profiles[activeProfileName];
    if (!profile) {
      throw new Error(`LLM-Wiki profile ${activeProfileName} not found in ~/.llm-wiki/config.json.`);
    }
    return profile;
  }

  resolveConnection() {
    const profile = this.loadCurrentCLIProfile();
    const baseURL = String(profile.base_url || "").trim().replace(/\/+$/, "");
    const accessToken = String(profile.access_token || "").trim();
    const tenantID = String(profile.tenant_id || "").trim();
    if (!baseURL) {
      throw new Error("LLM-Wiki base URL is missing from ~/.llm-wiki/config.json.");
    }
    if (!accessToken) {
      throw new Error("LLM-Wiki access token is missing from ~/.llm-wiki/config.json.");
    }
    if (!tenantID) {
      throw new Error("LLM-Wiki tenant is missing from ~/.llm-wiki/config.json.");
    }
    return { baseURL, accessToken, tenantID };
  }

  async requestJSON(method, endpoint, body) {
    const connection = this.resolveConnection();
    const response = await requestUrl({
      url: connection.baseURL + endpoint,
      method,
      headers: {
        Authorization: `Bearer ${connection.accessToken}`,
        "Content-Type": "application/json"
      },
      body: body === undefined ? undefined : JSON.stringify(body),
      throw: false
    });
    const payload = response.json ?? null;
    if (response.status >= 400) {
      const message = payload && payload.error && payload.error.message ? payload.error.message : `${response.status}`;
      throw new Error(message);
    }
    return payload;
  }

  async fetchSnapshot() {
    const connection = this.resolveConnection();
    const [whoami, namespaces, documents] = await Promise.all([
      this.requestJSON("GET", "/v1/auth/whoami"),
      this.requestJSON("GET", "/v1/namespaces"),
      this.requestJSON("GET", "/v1/documents")
    ]);
    return {
      connection,
      whoami,
      namespaces: namespaces.items || [],
      documents: documents.items || []
    };
  }

  async mirrorCurrentTenantToVault() {
    const snapshot = await this.fetchSnapshot();
    const tenantFolder = sanitizePathSegment(snapshot.whoami.tenant_id, "tenant");
    const rootPath = `${MIRROR_ROOT_FOLDER}/${tenantFolder}`;

    await this.ensureFolder(MIRROR_ROOT_FOLDER);
    await this.ensureFolder(rootPath);

    const namespaceByID = new Map();
    for (const namespace of snapshot.namespaces) {
      namespaceByID.set(namespace.id, namespace);
      await this.ensureFolder(`${rootPath}/${sanitizePathSegment(namespace.key || namespace.display_name, `namespace-${namespace.id}`)}`);
    }

    let changedCount = 0;
    for (const document of snapshot.documents) {
      const namespace = namespaceByID.get(document.namespace_id);
      const namespaceSegment = sanitizePathSegment(namespace ? (namespace.key || namespace.display_name) : "", `namespace-${document.namespace_id}`);
      const fileName = sanitizePathSegment(document.slug || document.title, `document-${document.id}`) + ".md";
      const filePath = `${rootPath}/${namespaceSegment}/${fileName}`;
      const content = renderMirroredMarkdown(snapshot.whoami.tenant_id, namespace, document);
      const changed = await this.upsertFile(filePath, content);
      if (changed) {
        changedCount += 1;
      }
    }

    const indexPath = `${rootPath}/_llm-wiki-index.md`;
    if (await this.upsertFile(indexPath, renderTenantIndexMarkdown(snapshot))) {
      changedCount += 1;
    }

    this.settings.lastMirrorAt = new Date().toISOString();
    await this.saveData(this.settings);
    return {
      tenantID: snapshot.whoami.tenant_id,
      rootPath,
      documentCount: snapshot.documents.length,
      changedCount
    };
  }

  async ensureFolder(folderPath) {
    const cleaned = folderPath.split("/").map((part) => part.trim()).filter(Boolean).join("/");
    if (!cleaned) {
      return;
    }
    const parts = cleaned.split("/");
    let current = "";
    for (const part of parts) {
      current = current ? `${current}/${part}` : part;
      const existing = this.app.vault.getAbstractFileByPath(current);
      if (!existing) {
        await this.app.vault.createFolder(current);
      }
    }
  }

  async upsertFile(filePath, content) {
    const existing = this.app.vault.getAbstractFileByPath(filePath);
    if (!existing) {
      await this.app.vault.create(filePath, content);
      return true;
    }
    if (existing.children) {
      throw new Error(`Mirror target is a folder, not a file: ${filePath}`);
    }
    const current = await this.app.vault.cachedRead(existing);
    if (current === content) {
      return false;
    }
    await this.app.vault.modify(existing, content);
    return true;
  }
};

function sanitizePathSegment(value, fallback) {
  const base = String(value || "").trim();
  const sanitized = base
    .replace(/[\/\\:*?"<>|]/g, "-")
    .replace(/\s+/g, "-")
    .replace(/-+/g, "-")
    .replace(/^[.-]+|[.-]+$/g, "");
  return sanitized || fallback;
}

function yamlScalar(value) {
  const trimmed = String(value || "").trim();
  if (!trimmed) {
    return `""`;
  }
  return `"${trimmed.replace(/"/g, '\\"')}"`;
}

function renderMirroredMarkdown(tenantID, namespace, document) {
  const namespaceKey = namespace ? (namespace.key || namespace.display_name) : "";
  const namespaceName = namespace ? namespace.display_name || namespace.key : "";
  const title = document.title || document.slug || `Document ${document.id}`;
  const frontmatter = [
    "---",
    "llm_wiki: true",
    `tenant_id: ${yamlScalar(tenantID)}`,
    `namespace_id: ${document.namespace_id}`,
    `namespace_key: ${yamlScalar(namespaceKey)}`,
    `namespace_name: ${yamlScalar(namespaceName)}`,
    `document_id: ${document.id}`,
    `slug: ${yamlScalar(document.slug)}`,
    `title: ${yamlScalar(document.title)}`,
    `status: ${yamlScalar(document.status)}`,
    `current_revision_no: ${document.current_revision_no}`,
    `updated_at: ${yamlScalar(document.updated_at)}`,
    "---",
    ""
  ].join("\n");

  let body = document.content || "";
  if (!body.trim()) {
    body = `# ${title}\n`;
  }
  if (!body.endsWith("\n")) {
    body += "\n";
  }
  return `${frontmatter}\n${body}`;
}

function renderTenantIndexMarkdown(snapshot) {
  const lines = [
    "---",
    "llm_wiki_index: true",
    `tenant_id: ${yamlScalar(snapshot.whoami.tenant_id)}`,
    `base_url: ${yamlScalar(snapshot.connection.baseURL)}`,
    `mirrored_at: ${yamlScalar(new Date().toISOString())}`,
    "---",
    "",
    `# ${snapshot.whoami.tenant_id}`,
    "",
    `- Namespaces mirrored: ${snapshot.namespaces.length}`,
    `- Documents mirrored: ${snapshot.documents.length}`,
    ""
  ];
  for (const namespace of snapshot.namespaces) {
    const docs = snapshot.documents.filter((item) => item.namespace_id === namespace.id);
    lines.push(`## ${namespace.display_name || namespace.key}`);
    lines.push("");
    if (docs.length === 0) {
      lines.push("- No documents");
      lines.push("");
      continue;
    }
    for (const doc of docs) {
      lines.push(`- ${doc.title || doc.slug}`);
    }
    lines.push("");
  }
  return lines.join("\n");
}
