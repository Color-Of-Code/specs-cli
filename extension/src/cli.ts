// Resolves and invokes the bundled (or user-configured) specs CLI binary.
import * as path from "path";
import * as fs from "fs";
import * as cp from "child_process";
import * as vscode from "vscode";

let outputChannel: vscode.OutputChannel | undefined;

export function getOutput(): vscode.OutputChannel {
    if (!outputChannel) {
        outputChannel = vscode.window.createOutputChannel("Specs");
    }
    return outputChannel;
}

/**
 * Resolution order:
 *   1. specs.cliPath setting (absolute path or relative to workspace)
 *   2. specs.useGlobalBinary === true: 'specs' on PATH
 *   3. extension's bundled binary at <extensionPath>/bin/specs[.exe]
 *   4. fallback: 'specs' on PATH
 */
export function resolveBinary(context: vscode.ExtensionContext): string {
    const cfg = vscode.workspace.getConfiguration("specs");
    const explicit = cfg.get<string>("cliPath", "").trim();
    if (explicit) {
        return explicit;
    }
    const useGlobal = cfg.get<boolean>("useGlobalBinary", false);
    const exe = process.platform === "win32" ? "specs.exe" : "specs";
    const bundled = path.join(context.extensionPath, "bin", exe);
    if (!useGlobal && fs.existsSync(bundled)) {
        return bundled;
    }
    return "specs"; // resolved via PATH
}

export interface RunResult {
    stdout: string;
    stderr: string;
    exitCode: number;
}

/** Runs the CLI and captures stdout/stderr. Logs to the Specs output channel. */
export async function runAndCapture(
    context: vscode.ExtensionContext,
    args: string[],
    cwd?: string
): Promise<RunResult> {
    const bin = resolveBinary(context);
    const out = getOutput();
    out.appendLine(`$ ${bin} ${args.join(" ")}  (cwd=${cwd ?? "<none>"})`);
    return new Promise((resolve) => {
        const proc = cp.spawn(bin, args, { cwd, env: process.env });
        let stdout = "";
        let stderr = "";
        proc.stdout.on("data", (d) => (stdout += d.toString()));
        proc.stderr.on("data", (d) => (stderr += d.toString()));
        proc.on("error", (err) => {
            resolve({ stdout, stderr: stderr + String(err), exitCode: 127 });
        });
        proc.on("close", (code) => {
            out.appendLine(`  -> exit ${code}`);
            resolve({ stdout, stderr, exitCode: code ?? 0 });
        });
    });
}

/** Runs the CLI in a dedicated VS Code terminal so the user can interact with it. */
export function runInTerminal(
    context: vscode.ExtensionContext,
    args: string[],
    cwd?: string,
    name = "Specs"
): vscode.Terminal {
    const bin = resolveBinary(context);
    const term = vscode.window.createTerminal({ name, cwd });
    // Quote args containing spaces; CLI args here are simple flags/values.
    const quoted = args.map((a) => (/[\s"'$]/.test(a) ? JSON.stringify(a) : a));
    term.sendText(`${bin} ${quoted.join(" ")}`);
    term.show();
    return term;
}

/** Returns the workspace folder containing a .specs.yaml (or specs/.specs.yaml), or undefined. */
export function findSpecsFolder(): vscode.WorkspaceFolder | undefined {
    const folders = vscode.workspace.workspaceFolders ?? [];
    for (const f of folders) {
        const root = f.uri.fsPath;
        if (
            fs.existsSync(path.join(root, ".specs.yaml")) ||
            fs.existsSync(path.join(root, "specs", ".specs.yaml"))
        ) {
            return f;
        }
    }
    return folders[0];
}

/** Returns the directory that contains .specs.yaml (the specs root). */
export function findSpecsRoot(folder: vscode.WorkspaceFolder): string | undefined {
    const candidates = [folder.uri.fsPath, path.join(folder.uri.fsPath, "specs")];
    for (const c of candidates) {
        if (fs.existsSync(path.join(c, ".specs.yaml"))) {
            return c;
        }
    }
    return undefined;
}
