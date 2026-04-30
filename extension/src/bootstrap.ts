// Bootstrap wizard. Maps multi-step QuickPick answers to a
// 'specs bootstrap ...' invocation.
import * as vscode from "vscode";
import { runInTerminal, runAndCapture, getOutput } from "./engine";

interface BootstrapAnswers {
  layout: "folder" | "submodule";
  specsUrl?: string;
  specsRef?: string;
  toolsMode: "managed" | "submodule" | "folder" | "vendor";
  toolsUrl: string;
  toolsRef: string;
  withModel: boolean;
  withVscode: boolean;
}

const DEFAULT_TOOLS_URL = "https://github.com/Color-Of-Code/specs-framework.git";

export async function runBootstrapWizard(context: vscode.ExtensionContext): Promise<void> {
  const folder = pickFolder();
  if (!folder) {
    return;
  }

  const layout = await pickLayout();
  if (!layout) {
    return;
  }

  let specsUrl: string | undefined;
  let specsRef: string | undefined;
  if (layout === "submodule") {
    specsUrl = await vscode.window.showInputBox({
      prompt: "Git URL of the host's specs repo (--specs-url)",
      placeHolder: "https://github.com/<org>/<host-specs>.git",
      ignoreFocusOut: true,
      validateInput: (v) => (v.trim().length === 0 ? "URL is required for submodule layout" : null),
    });
    if (!specsUrl) {
      return;
    }
    specsRef = await vscode.window.showInputBox({
      prompt: "Branch/tag for the specs submodule (optional)",
      placeHolder: "main",
      ignoreFocusOut: true,
    });
  }

  const toolsMode = await pickToolsMode();
  if (!toolsMode) {
    return;
  }

  const toolsUrl = await vscode.window.showInputBox({
    prompt: "Tools content git URL (--tools-url)",
    value: DEFAULT_TOOLS_URL,
    ignoreFocusOut: true,
  });
  if (toolsUrl === undefined) {
    return;
  }
  const toolsRef = await vscode.window.showInputBox({
    prompt: "Tools content ref (--tools-ref)",
    value: "main",
    ignoreFocusOut: true,
  });
  if (toolsRef === undefined) {
    return;
  }

  const extras = await vscode.window.showQuickPick(
    [
      { label: "Create model/ and change-requests/ skeletons", picked: true },
      { label: "Write .vscode/tasks.json", picked: false },
    ],
    {
      canPickMany: true,
      placeHolder: "Optional scaffolding",
      ignoreFocusOut: true,
    },
  );
  if (!extras) {
    return;
  }
  const withModel = extras.some((e) => e.label.startsWith("Create model"));
  const withVscode = extras.some((e) => e.label.startsWith("Write .vscode"));

  const answers: BootstrapAnswers = {
    layout,
    specsUrl,
    specsRef,
    toolsMode,
    toolsUrl,
    toolsRef,
    withModel,
    withVscode,
  };

  const args = buildArgs(answers);

  // Show dry-run preview first.
  const out = getOutput();
  out.show(true);
  out.appendLine("Specs bootstrap (dry-run preview)");
  const preview = await runAndCapture(context, [...args, "--dry-run"], folder.uri.fsPath);
  out.appendLine(preview.stdout);
  if (preview.stderr) {
    out.appendLine(preview.stderr);
  }
  if (preview.exitCode !== 0) {
    vscode.window.showErrorMessage(
      `bootstrap dry-run failed (exit ${preview.exitCode}). See Specs output.`,
    );
    return;
  }

  const choice = await vscode.window.showInformationMessage(
    `Run 'specs ${args.join(" ")}' in ${folder.uri.fsPath}?`,
    { modal: true },
    "Run",
    "Cancel",
  );
  if (choice !== "Run") {
    return;
  }
  runInTerminal(context, args, folder.uri.fsPath, "Specs: bootstrap");
}

function buildArgs(a: BootstrapAnswers): string[] {
  const args = ["bootstrap", "--layout", a.layout];
  if (a.layout === "submodule" && a.specsUrl) {
    args.push("--specs-url", a.specsUrl);
    if (a.specsRef) {
      args.push("--specs-ref", a.specsRef);
    }
  }
  args.push("--tools-mode", a.toolsMode);
  if (a.toolsUrl) {
    args.push("--tools-url", a.toolsUrl);
  }
  if (a.toolsRef) {
    args.push("--tools-ref", a.toolsRef);
  }
  if (a.withModel) {
    args.push("--with-model");
  }
  if (a.withVscode) {
    args.push("--with-vscode");
  }
  return args;
}

function pickFolder(): vscode.WorkspaceFolder | undefined {
  const folders = vscode.workspace.workspaceFolders ?? [];
  if (folders.length === 0) {
    vscode.window.showWarningMessage(
      "Specs: open a folder first; bootstrap operates on the active workspace.",
    );
    return undefined;
  }
  if (folders.length === 1) {
    return folders[0];
  }
  // Multi-root: would require a separate prompt; defer for now.
  vscode.window.showWarningMessage(
    "Specs: multi-root workspaces are not yet supported by the bootstrap wizard.",
  );
  return undefined;
}

async function pickLayout(): Promise<BootstrapAnswers["layout"] | undefined> {
  const items: vscode.QuickPickItem[] = [
    {
      label: "Folder",
      description: "specs/ is a plain folder in the host repo (recommended for new repos)",
    },
    {
      label: "Submodule",
      description: "specs/ is a git submodule of an existing specs repo",
    },
  ];
  const pick = await vscode.window.showQuickPick(items, {
    placeHolder: "How should specs/ be materialised?",
    ignoreFocusOut: true,
  });
  if (!pick) {
    return undefined;
  }
  return pick.label === "Submodule" ? "submodule" : "folder";
}

async function pickToolsMode(): Promise<BootstrapAnswers["toolsMode"] | undefined> {
  const items: vscode.QuickPickItem[] = [
    {
      label: "managed",
      description: "fetch into the user cache, share across projects (recommended)",
    },
    { label: "submodule", description: "add .specs-framework as a submodule of the host repo" },
    { label: "folder", description: "clone .specs-framework next to the specs root" },
    { label: "vendor", description: "snapshot .specs-framework without git history" },
  ];
  const pick = await vscode.window.showQuickPick(items, {
    placeHolder: "How should .specs-framework be materialised?",
    ignoreFocusOut: true,
  });
  if (!pick) {
    return undefined;
  }
  return pick.label as BootstrapAnswers["toolsMode"];
}
