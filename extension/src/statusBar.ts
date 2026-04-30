// Phase E4 — Status bar item showing current CR slug or tools_dir SHA.
import * as cp from "child_process";
import * as vscode from "vscode";
import { runAndCapture, findSpecsFolder, findSpecsRoot } from "./cli";

interface DoctorJSON {
  version: string;
  specs_root: string;
  tools_dir: string;
  tools_rev?: string;
  templates_schema?: number;
  compatible: boolean;
  compatible_message?: string;
}

// Coalesce burst events (e.g. multiple file saves in quick succession) into
// a single CLI invocation.
const REFRESH_DEBOUNCE_MS = 250;

export function registerStatusBar(context: vscode.ExtensionContext): void {
  const item = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
  item.command = "workbench.action.quickOpen";
  item.tooltip = "Specs (click to open palette)";
  context.subscriptions.push(item);

  const refresh = async () => {
    const folder = findSpecsFolder();
    if (!folder) {
      item.hide();
      return;
    }
    const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;
    const branch = currentBranch(folder.uri.fsPath);
    const cr = parseCRBranch(branch);

    let label: string;
    let tooltipExtra = "";
    if (cr) {
      label = `$(git-pull-request) ${cr}`;
      tooltipExtra = `On change-request branch ${branch}`;
    } else {
      const doctor = await loadDoctor(context, cwd);
      const rev = doctor?.tools_rev ?? "?";
      label = `$(book) tools@${rev}`;
      if (doctor && !doctor.compatible) {
        label = `$(warning) tools@${rev}`;
        tooltipExtra = doctor.compatible_message ?? "templates_schema mismatch";
      } else if (doctor) {
        tooltipExtra =
          `version ${doctor.version}` +
          (doctor.templates_schema ? ` \u2022 templates_schema=${doctor.templates_schema}` : "");
      }
    }
    item.text = label;
    item.tooltip = `Specs \u2014 ${tooltipExtra}\nClick to open Command Palette`;
    item.command = {
      command: "workbench.action.showCommands",
      title: "Specs commands",
      arguments: [],
    };
    item.show();
  };

  let pending: NodeJS.Timeout | undefined;
  const scheduleRefresh = () => {
    if (pending) {
      clearTimeout(pending);
    }
    pending = setTimeout(() => {
      pending = undefined;
      void refresh();
    }, REFRESH_DEBOUNCE_MS);
  };
  context.subscriptions.push({
    dispose: () => {
      if (pending) {
        clearTimeout(pending);
      }
    },
  });

  // Initial render.
  void refresh();

  // Event-driven refresh: workspace, config, branch and focus changes.
  const folder = findSpecsFolder();
  const subscriptions: vscode.Disposable[] = [
    vscode.workspace.onDidChangeWorkspaceFolders(scheduleRefresh),
    vscode.window.onDidChangeWindowState((s) => {
      if (s.focused) {
        scheduleRefresh();
      }
    }),
    vscode.commands.registerCommand("specs.statusBar.refresh", refresh),
  ];

  if (folder) {
    // Watch `.specs.yaml` anywhere under the workspace folder.
    const cfgWatcher = vscode.workspace.createFileSystemWatcher(
      new vscode.RelativePattern(folder, "**/.specs.yaml"),
    );
    cfgWatcher.onDidCreate(scheduleRefresh);
    cfgWatcher.onDidChange(scheduleRefresh);
    cfgWatcher.onDidDelete(scheduleRefresh);
    subscriptions.push(cfgWatcher);

    // Watch `.git/HEAD` for branch switches (covers terminal git checkouts).
    const headWatcher = vscode.workspace.createFileSystemWatcher(
      new vscode.RelativePattern(folder, ".git/HEAD"),
    );
    headWatcher.onDidChange(scheduleRefresh);
    headWatcher.onDidCreate(scheduleRefresh);
    subscriptions.push(headWatcher);
  }

  context.subscriptions.push(...subscriptions);
}

function currentBranch(repoRoot: string): string {
  try {
    const out = cp.execFileSync("git", ["-C", repoRoot, "symbolic-ref", "--short", "HEAD"], {
      stdio: ["ignore", "pipe", "ignore"],
    });
    return out.toString().trim();
  } catch {
    return "";
  }
}

/** Returns the CR slug if branch matches cr/<id>-<slug> (or CR-...). */
function parseCRBranch(branch: string): string | undefined {
  const m = /^(?:cr|CR)[\/-](\d+)[-_](.+)$/.exec(branch);
  if (!m) {
    return undefined;
  }
  const [, id, slug] = m;
  return `CR-${id.padStart(3, "0")}-${slug}`;
}

async function loadDoctor(
  context: vscode.ExtensionContext,
  cwd: string,
): Promise<DoctorJSON | undefined> {
  const res = await runAndCapture(context, ["doctor", "--json"], cwd);
  if (res.exitCode !== 0) {
    return undefined;
  }
  try {
    return JSON.parse(res.stdout) as DoctorJSON;
  } catch {
    return undefined;
  }
}
