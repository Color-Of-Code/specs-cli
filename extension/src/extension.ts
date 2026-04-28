import * as vscode from "vscode";
import { getOutput } from "./cli";
import { registerCommands } from "./commands";
import { registerCRTree } from "./crTree";
import { registerStatusBar } from "./statusBar";

export function activate(context: vscode.ExtensionContext): void {
    const out = getOutput();
    out.appendLine(`Specs extension activated (v${context.extension.packageJSON.version})`);
    registerCommands(context);
    registerCRTree(context);
    registerStatusBar(context);
}

export function deactivate(): void {
    // nothing
}
