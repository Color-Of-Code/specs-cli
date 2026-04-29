import fs from 'fs';
import path from 'path';

const extDir = path.resolve(path.dirname(new URL(import.meta.url).pathname), '..');
const vscodeExtDir = path.join(process.env.HOME || process.env.USERPROFILE, '.vscode/extensions');
const linkTarget = path.join(vscodeExtDir, 'Color-Of-Code.specs');

if (fs.existsSync(linkTarget)) {
  if (fs.lstatSync(linkTarget).isSymbolicLink() && fs.readlinkSync(linkTarget) === extDir) {
    console.log('Symlink already in place:', linkTarget, '->', extDir);
  } else {
    fs.rmSync(linkTarget, { recursive: true, force: true });
    fs.symlinkSync(extDir, linkTarget, 'dir');
    console.log('Updated symlink:', linkTarget, '->', extDir);
  }
} else {
  fs.symlinkSync(extDir, linkTarget, 'dir');
  console.log('Created symlink:', linkTarget, '->', extDir);
}