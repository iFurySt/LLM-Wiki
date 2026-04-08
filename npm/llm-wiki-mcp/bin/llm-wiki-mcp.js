#!/usr/bin/env node

import { readFileSync } from "node:fs";

const options = parseArgs(process.argv.slice(2));
const packageInfo = JSON.parse(readFileSync(new URL("../package.json", import.meta.url), "utf8"));

if (options.help) {
  printHelp();
  process.exit(0);
}

if (options.version) {
  process.stdout.write(`llm-wiki-mcp ${packageInfo.version}\n`);
  process.exit(0);
}

const baseUrl = options.baseUrl || process.env.LLM_WIKI_BASE_URL || "http://127.0.0.1:8234";
const tenantId = options.tenant || process.env.LLM_WIKI_TENANT_ID || "default";

const [{ McpServer, ResourceTemplate }, { StdioServerTransport }, { z }] = await Promise.all([
  import("@modelcontextprotocol/sdk/server/mcp.js"),
  import("@modelcontextprotocol/sdk/server/stdio.js"),
  import("zod")
]);

const server = new McpServer({
  name: "llm-wiki-mcp",
  version: packageInfo.version
});

server.registerTool(
  "llm_wiki_list_spaces",
  {
    title: "List Spaces",
    description: "List spaces for the current LLM-Wiki tenant.",
    inputSchema: {}
  },
  async () => toolResult(await requestJSON("GET", "/v1/spaces"))
);

server.registerTool(
  "llm_wiki_list_namespaces",
  {
    title: "List Namespaces",
    description: "List namespaces for the current LLM-Wiki tenant.",
    inputSchema: {}
  },
  async () => toolResult(await requestJSON("GET", "/v1/namespaces"))
);

server.registerTool(
  "llm_wiki_create_namespace",
  {
    title: "Create Namespace",
    description: "Create a namespace in the current tenant.",
    inputSchema: {
      key: z.string(),
      display_name: z.string(),
      description: z.string().optional(),
      visibility: z.string().optional()
    }
  },
  async (input) => toolResult(await requestJSON("POST", "/v1/namespaces", input))
);

server.registerTool(
  "llm_wiki_archive_namespace",
  {
    title: "Archive Namespace",
    description: "Archive a namespace in the current tenant.",
    inputSchema: {
      id: z.number().int()
    }
  },
  async ({ id }) => toolResult(await requestJSON("POST", `/v1/namespaces/${id}/archive`, {}))
);

server.registerTool(
  "llm_wiki_list_documents",
  {
    title: "List Documents",
    description: "List documents in the current tenant, optionally filtered by namespace or status.",
    inputSchema: {
      namespace_id: z.number().int().optional(),
      status: z.string().optional()
    }
  },
  async ({ namespace_id, status }) => {
    const query = new URLSearchParams();
    if (namespace_id !== undefined) {
      query.set("namespace_id", String(namespace_id));
    }
    if (status) {
      query.set("status", status);
    }
    const suffix = query.size > 0 ? `?${query.toString()}` : "";
    return toolResult(await requestJSON("GET", `/v1/documents${suffix}`));
  }
);

server.registerTool(
  "llm_wiki_get_document",
  {
    title: "Get Document",
    description: "Get a document and its revisions by ID.",
    inputSchema: {
      id: z.number().int()
    }
  },
  async ({ id }) => toolResult(await requestJSON("GET", `/v1/documents/${id}`))
);

server.registerTool(
  "llm_wiki_get_document_by_slug",
  {
    title: "Get Document By Slug",
    description: "Get a document by namespace ID and slug.",
    inputSchema: {
      namespace_id: z.number().int(),
      slug: z.string()
    }
  },
  async ({ namespace_id, slug }) => toolResult(await requestJSON("GET", `/v1/document-by-slug?namespace_id=${namespace_id}&slug=${encodeURIComponent(slug)}`))
);

server.registerTool(
  "llm_wiki_create_document",
  {
    title: "Create Document",
    description: "Create a new document and its first revision.",
    inputSchema: {
      namespace_id: z.number().int(),
      slug: z.string(),
      title: z.string(),
      content: z.string().optional(),
      author_type: z.string().optional(),
      author_id: z.string().optional(),
      change_summary: z.string().optional()
    }
  },
  async (input) => toolResult(await requestJSON("POST", "/v1/documents", input))
);

server.registerTool(
  "llm_wiki_update_document",
  {
    title: "Update Document",
    description: "Update a document and create a new revision.",
    inputSchema: {
      id: z.number().int(),
      title: z.string(),
      content: z.string().optional(),
      author_type: z.string().optional(),
      author_id: z.string().optional(),
      change_summary: z.string().optional()
    }
  },
  async ({ id, ...body }) => toolResult(await requestJSON("PUT", `/v1/documents/${id}`, body))
);

server.registerTool(
  "llm_wiki_archive_document",
  {
    title: "Archive Document",
    description: "Archive a document while preserving revision history.",
    inputSchema: {
      id: z.number().int(),
      author_type: z.string().optional(),
      author_id: z.string().optional(),
      change_summary: z.string().optional()
    }
  },
  async ({ id, ...body }) => toolResult(await requestJSON("POST", `/v1/documents/${id}/archive`, body))
);

server.registerResource(
  "llm-wiki-spaces",
  "llm-wiki://spaces",
  {
    title: "LLM-Wiki Spaces",
    description: "Spaces for the current tenant.",
    mimeType: "application/json"
  },
  async (uri) => ({
    contents: [{ uri: uri.href, mimeType: "application/json", text: JSON.stringify(await requestJSON("GET", "/v1/spaces"), null, 2) }]
  })
);

server.registerResource(
  "llm-wiki-namespaces",
  "llm-wiki://namespaces",
  {
    title: "LLM-Wiki Namespaces",
    description: "Namespaces for the current tenant.",
    mimeType: "application/json"
  },
  async (uri) => ({
    contents: [{ uri: uri.href, mimeType: "application/json", text: JSON.stringify(await requestJSON("GET", "/v1/namespaces"), null, 2) }]
  })
);

server.registerResource(
  "llm-wiki-document",
  new ResourceTemplate("llm-wiki://documents/{id}", { list: undefined }),
  {
    title: "LLM-Wiki Document",
    description: "Read a document by ID.",
    mimeType: "application/json"
  },
  async (uri, params) => ({
    contents: [{ uri: uri.href, mimeType: "application/json", text: JSON.stringify(await requestJSON("GET", `/v1/documents/${params.id}`), null, 2) }]
  })
);

server.registerResource(
  "llm-wiki-document-by-slug",
  new ResourceTemplate("llm-wiki://documents/by-slug/{namespace_id}/{slug}", { list: undefined }),
  {
    title: "LLM-Wiki Document By Slug",
    description: "Read a document by namespace ID and slug.",
    mimeType: "application/json"
  },
  async (uri, params) => ({
    contents: [{
      uri: uri.href,
      mimeType: "application/json",
      text: JSON.stringify(await requestJSON("GET", `/v1/document-by-slug?namespace_id=${params.namespace_id}&slug=${encodeURIComponent(params.slug)}`), null, 2)
    }]
  })
);

const transport = new StdioServerTransport();
await server.connect(transport);

async function requestJSON(method, path, body) {
  const response = await fetch(new URL(path, ensureTrailingSlash(baseUrl)), {
    method,
    headers: {
      "Accept": "application/json",
      "Content-Type": "application/json",
      "X-LLM-Wiki-Tenant-ID": tenantId
    },
    body: body === undefined ? undefined : JSON.stringify(body)
  });

  const text = await response.text();
  if (!response.ok) {
    throw new Error(`LLM-Wiki API ${response.status}: ${text}`);
  }
  return text.length === 0 ? {} : JSON.parse(text);
}

function toolResult(data) {
  return {
    content: [{ type: "text", text: JSON.stringify(data, null, 2) }],
    structuredContent: data
  };
}

function ensureTrailingSlash(value) {
  return value.endsWith("/") ? value : `${value}/`;
}

function parseArgs(args) {
  const parsed = {
    help: false,
    version: false,
    baseUrl: "",
    tenant: ""
  };

  for (let i = 0; i < args.length; i += 1) {
    const value = args[i];
    switch (value) {
      case "--help":
      case "-h":
        parsed.help = true;
        break;
      case "--version":
      case "-v":
        parsed.version = true;
        break;
      case "--base-url":
        parsed.baseUrl = args[i + 1] || "";
        i += 1;
        break;
      case "--tenant":
        parsed.tenant = args[i + 1] || "";
        i += 1;
        break;
      default:
        break;
    }
  }

  return parsed;
}

function printHelp() {
  process.stdout.write(`LLM-Wiki MCP stdio bridge

Usage:
  llm-wiki-mcp [--base-url URL] [--tenant TENANT]
  llm-wiki-mcp --version
  llm-wiki-mcp --help

Options:
  --base-url  LLM-Wiki server base URL. Defaults to LLM_WIKI_BASE_URL or http://127.0.0.1:8234
  --tenant    LLM-Wiki tenant ID. Defaults to LLM_WIKI_TENANT_ID or default
  --version   Print package version
  --help      Show this help message
`);
}
