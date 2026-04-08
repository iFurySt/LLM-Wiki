import * as React from "react";
import { createRoot } from "react-dom/client";
import { ThemeProvider, createTheme } from "@mui/material/styles";
import { CssBaseline } from "@mui/material";
import { SimpleTreeView, TreeItem } from "@mui/x-tree-view";

type TreeNode = {
  id: string;
  label: string;
  href?: string;
  children?: TreeNode[];
};

type TreeState = {
  tenantId?: string;
  ns?: string;
  statusFilter: string;
  selectedItemId?: string;
  expandedItemIds?: string[];
  folders: TreeNode[];
};

type TreeController = {
  mountFromDOM: () => void;
};

declare global {
  interface Window {
    LLMWikiUI?: {
      navigate?: (href: string, options?: { pushHistory?: boolean }) => Promise<void> | void;
    };
    LLMWikiTree?: TreeController;
  }
}

function readTreeState(): TreeState | null {
  const payload = document.getElementById("mui-tree-state");
  if (!payload) {
    return null;
  }
  try {
    return JSON.parse(payload.textContent || "") as TreeState;
  } catch (_error) {
    return null;
  }
}

function currentMode(): "light" | "dark" {
  return document.documentElement.getAttribute("data-theme") === "light" ? "light" : "dark";
}

function filterTree(nodes: TreeNode[], query: string): TreeNode[] {
  const normalized = query.trim().toLowerCase();
  if (normalized === "") {
    return nodes;
  }

  return nodes
    .map((node) => {
      const matchesSelf = node.label.toLowerCase().includes(normalized);
      const filteredChildren = filterTree(node.children || [], query);
      if (!matchesSelf && filteredChildren.length === 0) {
        return null;
      }
      return {
        ...node,
        children: matchesSelf ? node.children || [] : filteredChildren,
      };
    })
    .filter((node): node is TreeNode => node !== null);
}

function collectExpandedIds(nodes: TreeNode[]): string[] {
  return nodes.filter((node) => (node.children || []).length > 0).map((node) => node.id);
}

function collectHrefMap(nodes: TreeNode[], map: Record<string, string> = {}): Record<string, string> {
  nodes.forEach((node) => {
    if (node.href) {
      map[node.id] = node.href;
    }
    if (node.children) {
      collectHrefMap(node.children, map);
    }
  });
  return map;
}

function TreeApp({ state }: { state: TreeState }) {
  const [mode, setMode] = React.useState<"light" | "dark">(currentMode());
  const [query, setQuery] = React.useState("");
  const [selectedItemId, setSelectedItemId] = React.useState<string | undefined>(state.selectedItemId);
  const [expandedItems, setExpandedItems] = React.useState<string[]>(state.expandedItemIds || []);

  React.useEffect(() => {
    const input = document.querySelector<HTMLInputElement>("[data-tree-search]");
    if (!input) {
      return undefined;
    }
    const handler = () => setQuery(input.value || "");
    input.addEventListener("input", handler);
    return () => input.removeEventListener("input", handler);
  }, []);

  React.useEffect(() => {
    const observer = new MutationObserver(() => setMode(currentMode()));
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ["data-theme"] });
    return () => observer.disconnect();
  }, []);

  const visibleNamespaces = React.useMemo(() => filterTree(state.folders, query), [state.folders, query]);
  const hrefMap = React.useMemo(() => collectHrefMap(state.folders), [state.folders]);

  React.useEffect(() => {
    if (query.trim() !== "") {
      setExpandedItems(collectExpandedIds(visibleNamespaces));
    } else {
      setExpandedItems(state.expandedItemIds || []);
    }
  }, [query, state.expandedItemIds, visibleNamespaces]);

  const theme = React.useMemo(
    () =>
      createTheme({
        palette: { mode },
      }),
    [mode],
  );

  const renderNode = (node: TreeNode): React.ReactNode => (
    <TreeItem
      key={node.id}
      itemId={node.id}
      label={<span className="mui-tree-item-title">{node.label}</span>}
    >
      {(node.children || []).map(renderNode)}
    </TreeItem>
  );

  const handleSelectionChange = (_event: React.SyntheticEvent | null, itemId: string | null) => {
    if (!itemId) {
      return;
    }
    setSelectedItemId(itemId);
    if (hrefMap[itemId]) {
      if (window.LLMWikiUI?.navigate) {
        void window.LLMWikiUI.navigate(hrefMap[itemId]);
        return;
      }
      window.location.assign(hrefMap[itemId]);
    }
  };

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <div className="mui-tree-shell">
        {visibleNamespaces.length === 0 ? (
          <div className="mui-tree-empty">No documents found.</div>
        ) : (
          <SimpleTreeView
            expandedItems={expandedItems}
            selectedItems={selectedItemId}
            onExpandedItemsChange={(_event, itemIds) => setExpandedItems(itemIds as string[])}
            onSelectedItemsChange={(_event, itemId) => handleSelectionChange(_event, itemId as string | null)}
            sx={{
              color: "var(--text)",
              overflow: "hidden",
              "& .MuiTreeItem-root": {
                margin: 0,
              },
              "& .MuiTreeItem-content": {
                minHeight: 30,
                borderRadius: "6px",
                paddingInline: "6px",
                paddingBlock: 0,
                gap: "4px",
              },
              "& .MuiTreeItem-content:hover": {
                backgroundColor: "rgba(255,255,255,0.04)",
              },
              "& .MuiTreeItem-content.Mui-expanded": {
                backgroundColor: "transparent",
              },
              "& .MuiTreeItem-content.Mui-selected": {
                backgroundColor: "rgba(108,178,255,0.16)",
              },
              "& .MuiTreeItem-content.Mui-selected:hover": {
                backgroundColor: "rgba(108,178,255,0.20)",
              },
              "& .MuiTreeItem-iconContainer": {
                width: 20,
                minWidth: 20,
                color: "var(--muted)",
                "& svg": {
                  fontSize: 18,
                },
              },
              "& .MuiTreeItem-label": {
                padding: 0,
                fontSize: 12,
                fontWeight: 500,
                color: "var(--text)",
              },
              "& .MuiTreeItem-groupTransition": {
                marginLeft: "18px",
                borderLeft: "1px solid rgba(255,255,255,0.05)",
              },
            }}
          >
            {visibleNamespaces.map(renderNode)}
          </SimpleTreeView>
        )}
      </div>
    </ThemeProvider>
  );
}

let mountedRoot: ReturnType<typeof createRoot> | null = null;
let mountedNode: Element | null = null;

function mountFromDOM() {
  const mountNode = document.getElementById("mui-tree-root");
  const state = readTreeState();

  if (!mountNode || !state) {
    if (mountedRoot) {
      mountedRoot.unmount();
      mountedRoot = null;
      mountedNode = null;
    }
    return;
  }

  if (mountedNode !== mountNode) {
    if (mountedRoot) {
      mountedRoot.unmount();
    }
    mountedRoot = createRoot(mountNode);
    mountedNode = mountNode;
  }

  mountedRoot.render(<TreeApp state={state} />);
}

window.LLMWikiTree = {
  mountFromDOM,
};

mountFromDOM();
