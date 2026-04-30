// Init wizard. Maps QuickPick answers to a 'specs init ...' invocation.
// Surface name kept as `specs.bootstrap` for stable keybindings.
import * as vscode from "vscode";
import { runInTerminal, runAndCapture, getOutput } from "./engine";

interface InitAnswers {
  framework: string;
  withModel: boolean;
  withVscode: boolean;
}

const FRAMEWORK_PROMPT =
  "Registered framework name (--framework). Optional `@ref` overrides the registered ref. Leave empty to use the registry's `default` entry.";

export async function runBootstrapWizard(context: vscode.ExtensionContext): Promise<void> {
  const folder = pickFolder();
  if (!folder) {
    return;
  }

  const framework = await vscode.window.showInputBox({
    prompt: FRAMEWORK_PROMPT,
    value: "",
    ignoreFocusOut: true,
  });
  if (framework === undefined) {
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

  const answers: InitAnswers = {
    framework: framework.trim(),
    withModel,
    withVscode,
  };

  const args = buildArgs(answers);

  const out = getOutput();
  out.show(true);
  out.appendLine("Specs init (dry-run preview)");
  const preview = await runAndCapture(context, [...args, "--dry-run"], folder.uri.fsPath);
  out.appendLine(preview.stdout);
  if (preview.stderr) {
    out.appendLine(preview.stderr);
  }
  if (preview.exitCode !== 0) {
    vscode.window.showErrorMessage(
      `init dry-run failed (exit ${preview.exitCode}). See Specs output.`,
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
  runInTerminal(context, args, folder.uri.fsPath, "Specs: init");
}

function buildArgs(a: InitAnswers): string[] {
  const args = ["init"];
  if (a.framework !== "") {
    args.push("--framework", a.framework);
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
      "Specs: open a folder first; init operates on the active workspace.",
    );
    return undefined;
  }
  if (folders.length === 1) {
    return folders[0];
  }
  vscode.window.showWarningMessage(
    "Specs: multi-root workspaces are not yet supported by the init wizard.",
  );
  return undefined;
}
