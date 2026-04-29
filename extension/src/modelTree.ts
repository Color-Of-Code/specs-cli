// Filesystem-backed tree view for a model/ subdirectory (requirements, features, etc.).
import * as path from "path";
import * as fs from "fs";
import * as vscode from "vscode";
import { findSpecsFolder, findSpecsRoot } from "./cli";

type Node = DirNode | FileNode;

class DirNode {
    readonly kind = "dir" as const;
    constructor(
        public readonly fullPath: string,
        public readonly label: string
    ) {}
}

class FileNode {
    readonly kind = "file" as const;
    constructor(
        public readonly fullPath: string,
        public readonly label: string
    ) {}
}

export class ModelTreeProvider implements vscode.TreeDataProvider<Node> {
    private _onDidChange = new vscode.EventEmitter<Node | void>();
    readonly onDidChangeTreeData = this._onDidChange.event;

    /**
     * @param subdir The subdirectory name under model/ to display (e.g. "requirements", "features").
     */
    constructor(private readonly subdir: string) {}

    refresh(): void {
        this._onDidChange.fire();
    }

    getTreeItem(node: Node): vscode.TreeItem {
        if (node.kind === "dir") {
            const item = new vscode.TreeItem(
                node.label,
                vscode.TreeItemCollapsibleState.Collapsed
            );
            item.iconPath = vscode.ThemeIcon.Folder;
            item.resourceUri = vscode.Uri.file(node.fullPath);
            item.contextValue = "specs.model.dir";
            return item;
        }
        const item = new vscode.TreeItem(
            node.label,
            vscode.TreeItemCollapsibleState.None
        );
        item.resourceUri = vscode.Uri.file(node.fullPath);
        item.iconPath = vscode.ThemeIcon.File;
        item.command = {
            command: "vscode.open",
            title: "Open",
            arguments: [item.resourceUri],
        };
        item.contextValue = "specs.model.file";
        return item;
    }

    getChildren(node?: Node): Node[] {
        if (!node) {
            const root = this.resolveModelSubdir();
            if (!root || !fs.existsSync(root)) {
                return [];
            }
            return this.readDir(root);
        }
        if (node.kind === "dir") {
            return this.readDir(node.fullPath);
        }
        return [];
    }

    private resolveModelSubdir(): string | undefined {
        const folder = findSpecsFolder();
        if (!folder) {
            return undefined;
        }
        const specsRoot = findSpecsRoot(folder) ?? folder.uri.fsPath;
        return path.join(specsRoot, "model", this.subdir);
    }

    private readDir(dir: string): Node[] {
        let entries: fs.Dirent[];
        try {
            entries = fs.readdirSync(dir, { withFileTypes: true });
        } catch {
            return [];
        }
        const nodes: Node[] = [];
        for (const e of entries.sort((a, b) => a.name.localeCompare(b.name))) {
            const full = path.join(dir, e.name);
            if (e.isDirectory()) {
                nodes.push(new DirNode(full, e.name));
            } else if (e.isFile() && e.name.endsWith(".md")) {
                nodes.push(new FileNode(full, e.name));
            }
        }
        return nodes;
    }
}

export function registerModelTree(
    context: vscode.ExtensionContext,
    viewId: string,
    subdir: string,
    refreshCommand: string
): ModelTreeProvider {
    const provider = new ModelTreeProvider(subdir);
    context.subscriptions.push(
        vscode.window.registerTreeDataProvider(viewId, provider),
        vscode.commands.registerCommand(refreshCommand, () => provider.refresh())
    );
    // Auto-refresh when model/<subdir>/ changes.
    const folder = findSpecsFolder();
    if (folder) {
        const root = findSpecsRoot(folder) ?? folder.uri.fsPath;
        const watcher = vscode.workspace.createFileSystemWatcher(
            new vscode.RelativePattern(root, `model/${subdir}/**`)
        );
        watcher.onDidCreate(() => provider.refresh());
        watcher.onDidDelete(() => provider.refresh());
        watcher.onDidChange(() => provider.refresh());
        context.subscriptions.push(watcher);
    }
    return provider;
}
