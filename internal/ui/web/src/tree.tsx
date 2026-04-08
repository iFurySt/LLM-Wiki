import * as React from "react";
import { createRoot } from "react-dom/client";
import { ThemeProvider, createTheme } from "@mui/material/styles";
import { Box, CssBaseline, FormControl, MenuItem, Select } from "@mui/material";
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

type NSSpace = {
  ns: string;
  displayName: string;
  role?: string;
};

type NSSwitchState = {
  currentNS: string;
  spaces: NSSpace[];
};

declare global {
  interface Window {
    LLMWikiUI?: {
      navigate?: (href: string, options?: { pushHistory?: boolean; scope?: "content" | "reader" }) => Promise<void> | void;
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

function readNSSwitchState(): NSSwitchState | null {
  const payload = document.getElementById("mui-ns-switch-state");
  if (!payload) {
    return null;
  }
  try {
    return JSON.parse(payload.textContent || "") as NSSwitchState;
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
        void window.LLMWikiUI.navigate(hrefMap[itemId], { scope: "reader" });
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
            expansionTrigger="iconContainer"
            expandedItems={expandedItems}
            selectedItems={selectedItemId}
            onExpandedItemsChange={(_event, itemIds) => setExpandedItems(itemIds as string[])}
            onSelectedItemsChange={(_event, itemId) => handleSelectionChange(_event, itemId as string | null)}
            sx={{
              color: "var(--text)",
              overflow: "hidden",
              "& .MuiTreeItem-root": {
                margin: 0,
                "& + .MuiTreeItem-root": {
                  marginTop: "4px",
                },
              },
              "& .MuiTreeItem-content": {
                minHeight: 34,
                borderRadius: "12px",
                paddingLeft: "12px",
                paddingRight: "8px",
                paddingBlock: "2px",
                gap: "6px",
                transition: "160ms ease",
              },
              "& .MuiTreeItem-content:hover": {
                backgroundColor: "var(--accent-soft)",
              },
              "& .MuiTreeItem-content.Mui-expanded": {
                backgroundColor: "transparent",
              },
              "& .MuiTreeItem-content.Mui-selected": {
                backgroundColor: "var(--accent-soft)",
                boxShadow: "inset 0 0 0 1px rgba(166, 75, 42, 0.18)",
              },
              "& .MuiTreeItem-content.Mui-selected:hover": {
                backgroundColor: "var(--accent-soft)",
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
                fontSize: 13,
                fontWeight: 600,
                color: "var(--text)",
              },
              "& .MuiTreeItem-groupTransition": {
                marginLeft: "18px",
                borderLeft: "1px solid var(--line)",
                paddingLeft: "4px",
                paddingTop: "4px",
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

function postNSSwitch(nextNS: string) {
  const form = document.createElement("form");
  form.method = "post";
  form.action = "/ui/ns/switch";

  const input = document.createElement("input");
  input.type = "hidden";
  input.name = "ns";
  input.value = nextNS;

  form.appendChild(input);
  document.body.appendChild(form);
  form.submit();
}

function NSSwitchApp({ state }: { state: NSSwitchState }) {
  const [mode, setMode] = React.useState<"light" | "dark">(currentMode());
  const [value, setValue] = React.useState(state.currentNS);

  React.useEffect(() => {
    const observer = new MutationObserver(() => setMode(currentMode()));
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ["data-theme"] });
    return () => observer.disconnect();
  }, []);

  const theme = React.useMemo(
    () =>
      createTheme({
        palette: { mode },
      }),
    [mode],
  );

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box className="workspace-switch-shell-inner">
        <FormControl fullWidth size="small">
          <Select
            value={value}
            onChange={(event) => {
              const nextValue = event.target.value;
              setValue(nextValue);
              if (nextValue !== state.currentNS) {
                postNSSwitch(nextValue);
              }
            }}
            displayEmpty
            MenuProps={{
              PaperProps: {
                sx: {
                  borderRadius: "18px",
                  border: "1px solid var(--line)",
                  mt: 1,
                  boxShadow: "0 18px 40px var(--shadow-strong)",
                  background: "var(--panel-2)",
                  backdropFilter: "blur(16px)",
                },
              },
            }}
            sx={{
              height: 52,
              borderRadius: "18px",
              background: "linear-gradient(180deg, var(--panel-2), var(--panel))",
              color: "var(--text)",
              boxShadow: "inset 0 1px 0 var(--paper)",
              "& .MuiSelect-select": {
                display: "flex",
                alignItems: "center",
                minHeight: "unset",
                padding: "14px 44px 14px 16px",
                fontSize: 14,
                fontWeight: 600,
              },
              "& .MuiOutlinedInput-notchedOutline": {
                borderColor: "var(--line)",
              },
              "&:hover .MuiOutlinedInput-notchedOutline": {
                borderColor: "var(--line-strong)",
              },
              "&.Mui-focused .MuiOutlinedInput-notchedOutline": {
                borderColor: "var(--accent)",
                borderWidth: "1px",
              },
              "& .MuiSvgIcon-root": {
                color: "var(--muted)",
              },
            }}
          >
            {state.spaces.map((space) => (
              <MenuItem
                key={space.ns}
                value={space.ns}
                sx={{
                  minHeight: 48,
                  borderRadius: "12px",
                  mx: 1,
                  my: 0.5,
                  color: "var(--text)",
                  "&.Mui-selected": {
                    backgroundColor: "var(--accent-soft)",
                  },
                  "&.Mui-selected:hover": {
                    backgroundColor: "var(--accent-soft)",
                  },
                }}
              >
                {space.displayName}
                {space.role ? ` · ${space.role}` : ""}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      </Box>
    </ThemeProvider>
  );
}

let mountedRoot: ReturnType<typeof createRoot> | null = null;
let mountedNode: Element | null = null;
let mountedNSRoot: ReturnType<typeof createRoot> | null = null;
let mountedNSNode: Element | null = null;

function mountFromDOM() {
  const mountNode = document.getElementById("mui-tree-root");
  const state = readTreeState();
  const nsMountNode = document.getElementById("mui-ns-switch-root");
  const nsState = readNSSwitchState();

  if (!mountNode || !state) {
    if (mountedRoot) {
      mountedRoot.unmount();
      mountedRoot = null;
      mountedNode = null;
    }
  } else {
    if (mountedNode !== mountNode) {
      if (mountedRoot) {
        mountedRoot.unmount();
      }
      mountedRoot = createRoot(mountNode);
      mountedNode = mountNode;
    }

    mountedRoot.render(<TreeApp state={state} />);
  }

  if (!nsMountNode || !nsState) {
    if (mountedNSRoot) {
      mountedNSRoot.unmount();
      mountedNSRoot = null;
      mountedNSNode = null;
    }
    return;
  }

  if (mountedNSNode !== nsMountNode) {
    if (mountedNSRoot) {
      mountedNSRoot.unmount();
    }
    mountedNSRoot = createRoot(nsMountNode);
    mountedNSNode = nsMountNode;
  }

  mountedNSRoot.render(<NSSwitchApp state={nsState} />);
}

window.LLMWikiTree = {
  mountFromDOM,
};

mountFromDOM();
