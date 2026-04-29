import * as vscode from "vscode";
import { getOutput } from "./cli";
import { registerCommands } from "./commands";
import { registerCRTree } from "./crTree";
import { registerModelTree } from "./modelTree";
import { registerStatusBar } from "./statusBar";
import { registerVisualizePanel } from "./visualizePanel";

export function activate(context: vscode.ExtensionContext): void {
    const out = getOutput();
    out.appendLine(`Specs extension activated (v${context.extension.packageJSON.version})`);
    registerCommands(context);
    registerCRTree(context);
    registerModelTree(context, "specs.requirements", "requirements", "specs.requirements.refresh");
    registerModelTree(context, "specs.features", "features", "specs.features.refresh");
    registerStatusBar(context);
    registerVisualizePanel(context);
}

export function deactivate(): void {
    // nothing
}
