import * as vscode from "vscode";
import { runInTerminal, findSpecsFolder, findSpecsRoot, getOutput } from "./cli";

export function activate(context: vscode.ExtensionContext): void {
    const out = getOutput();
    out.appendLine(`Specs extension activated (v${context.extension.packageJSON.version})`);

    context.subscriptions.push(
        vscode.commands.registerCommand("specs.doctor", () => runSimple(context, ["doctor"])),
        vscode.commands.registerCommand("specs.lint", () => runSimple(context, ["lint"]))
    );
}

export function deactivate(): void {
    // nothing
}

function runSimple(context: vscode.ExtensionContext, args: string[]): void {
    const folder = findSpecsFolder();
    if (!folder) {
        vscode.window.showWarningMessage("Specs: no workspace folder is open.");
        return;
    }
    const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;
    runInTerminal(context, args, cwd);
}
