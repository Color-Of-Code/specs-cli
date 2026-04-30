// Phase E1 — palette wrappers around CLI subcommands.
import * as vscode from "vscode";
import { runInTerminal, runAndCapture, findSpecsFolder, findSpecsRoot, getOutput } from "./cli";
import { runBootstrapWizard } from "./bootstrap";

type ScaffoldKind = "requirement" | "feature" | "component" | "api" | "service";

export function registerCommands(context: vscode.ExtensionContext): void {
  const reg = (id: string, fn: () => void | Promise<void>) =>
    context.subscriptions.push(vscode.commands.registerCommand(id, fn));

  // Bootstrap wizard.
  reg("specs.bootstrap", () => runBootstrapWizard(context));

  // Lint family.
  reg("specs.lint", () => runTerminal(context, ["lint"]));
  reg("specs.lint.links", () => runTerminal(context, ["lint", "--links"]));
  reg("specs.lint.style", () => runTerminal(context, ["lint", "--style"]));
  reg("specs.lint.baselines", () => runTerminal(context, ["lint", "--baselines"]));

  // Diagnostics.
  reg("specs.doctor", () => runTerminal(context, ["doctor"]));

  // Tools cache.
  reg("specs.toolsUpdate", () => runTerminal(context, ["tools", "update"]));

  // Visualize (writes a file in model/ and opens it).
  reg("specs.visualize.dot", () => visualize(context, "dot"));
  reg("specs.visualize.mermaid", () => visualize(context, "mermaid"));

  // Scaffold a new model file.
  reg("specs.scaffold.requirement", () => scaffold(context, "requirement"));
  reg("specs.scaffold.feature", () => scaffold(context, "feature"));
  reg("specs.scaffold.component", () => scaffold(context, "component"));
  reg("specs.scaffold.api", () => scaffold(context, "api"));
  reg("specs.scaffold.service", () => scaffold(context, "service"));

  // Change-requests.
  reg("specs.cr.new", () => crNew(context));
  reg("specs.cr.status", () => runTerminal(context, ["cr", "status"]));
  reg("specs.cr.drain", () => crDrain(context));

  // Cross-link consistency.
  reg("specs.linkCheck", () => runTerminal(context, ["link", "check"]));

  // Baseline.
  reg("specs.baseline.check", () => runTerminal(context, ["baseline", "check"]));
  reg("specs.baseline.update", () => baselineUpdate(context));
}

// --- helpers ---

function runTerminal(context: vscode.ExtensionContext, args: string[]): void {
  const folder = findSpecsFolder();
  if (!folder) {
    vscode.window.showWarningMessage("Specs: no workspace folder is open.");
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;
  runInTerminal(context, args, cwd);
}

async function scaffold(context: vscode.ExtensionContext, kind: ScaffoldKind): Promise<void> {
  const folder = findSpecsFolder();
  if (!folder) {
    vscode.window.showWarningMessage("Specs: no workspace folder is open.");
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;

  const placeholder =
    kind === "requirement"
      ? "e.g. core/012-some-requirement"
      : kind === "feature" || kind === "component"
      ? "e.g. core/some-slug"
      : "e.g. some-slug";
  const relPath = await vscode.window.showInputBox({
    prompt: `Relative path under model/${kind}s/ (without .md)`,
    placeHolder: placeholder,
    ignoreFocusOut: true,
    validateInput: (v) => (v.trim() ? null : "path is required"),
  });
  if (!relPath) {
    return;
  }
  const title = await vscode.window.showInputBox({
    prompt: `${kind} title (optional — defaults to derived from slug)`,
    placeHolder: "Short, descriptive title",
    ignoreFocusOut: true,
  });
  if (title === undefined) {
    return;
  }
  const args = ["scaffold", kind];
  if (title) {
    args.push("--title", title);
  }
  args.push(relPath);
  runInTerminal(context, args, cwd);
}

async function crNew(context: vscode.ExtensionContext): Promise<void> {
  const folder = findSpecsFolder();
  if (!folder) {
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;

  const id = await vscode.window.showInputBox({
    prompt: "Change-request id (e.g. 042)",
    ignoreFocusOut: true,
    validateInput: (v) => (/^\d+$/.test(v) ? null : "expected a number"),
  });
  if (!id) {
    return;
  }
  const slug = await vscode.window.showInputBox({
    prompt: "Change-request slug (kebab-case)",
    placeHolder: "e.g. add-login-flow",
    ignoreFocusOut: true,
    validateInput: (v) => (/^[a-z0-9]+(-[a-z0-9]+)*$/.test(v) ? null : "expected kebab-case"),
  });
  if (!slug) {
    return;
  }
  const title = await vscode.window.showInputBox({
    prompt: "Change-request title (optional — defaults to derived from slug)",
    ignoreFocusOut: true,
  });
  if (title === undefined) {
    return;
  }
  const args = ["cr", "new", "--id", id, "--slug", slug];
  if (title) {
    args.push("--title", title);
  }
  runInTerminal(context, args, cwd);
}

async function crDrain(context: vscode.ExtensionContext): Promise<void> {
  const folder = findSpecsFolder();
  if (!folder) {
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;

  const id = await vscode.window.showInputBox({
    prompt: "Change-request id (e.g. 042)",
    ignoreFocusOut: true,
    validateInput: (v) => (/^\d+$/.test(v) ? null : "expected a number"),
  });
  if (!id) {
    return;
  }
  runInTerminal(context, ["cr", "drain", "--id", id], cwd);
}

async function baselineUpdate(context: vscode.ExtensionContext): Promise<void> {
  const folder = findSpecsFolder();
  if (!folder) {
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;

  const only = await vscode.window.showInputBox({
    prompt: "Component filter (substring; leave empty for all)",
    ignoreFocusOut: true,
  });
  const args = ["baseline", "update"];
  if (only) {
    args.push("--only", only);
  }
  runInTerminal(context, args, cwd);
}

async function visualize(
  context: vscode.ExtensionContext,
  format: "dot" | "mermaid",
): Promise<void> {
  const folder = findSpecsFolder();
  if (!folder) {
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;

  const ext = format === "dot" ? "dot" : "md";
  const outPath = `model/_visualize.${ext}`;
  const out = getOutput();
  out.appendLine(`Specs: visualize -> ${outPath}`);

  const res = await runAndCapture(
    context,
    ["visualize", "traceability", "--format", format, "--out", outPath],
    cwd,
  );
  if (res.exitCode !== 0) {
    out.appendLine(res.stderr);
    out.appendLine(res.stdout);
    out.show(true);
    vscode.window.showErrorMessage(
      `specs visualize failed (exit ${res.exitCode}). See Specs output for details.`,
    );
    return;
  }
  const uri = vscode.Uri.joinPath(folder.uri, outPath);
  const doc = await vscode.workspace.openTextDocument(uri);
  await vscode.window.showTextDocument(doc, { preview: true });
}
