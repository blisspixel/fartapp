import { readdirSync, readFileSync, realpathSync, statSync } from "node:fs";
import ignore from "ignore";
import { load } from "js-yaml";
import { lint, readConfig } from "markdownlint/sync";

const configPath = ".markdownlint.yaml";
let config;
try {
  config = readConfig(configPath, [load]);
} catch (error) {
  console.error(`Cannot read or parse config file '${configPath}': ${error.message}`);
  process.exit(4);
}

function markdownFiles(excluded) {
  const files = [];
  const ancestors = new Set();
  const extension = ["darwin", "win32"].includes(process.platform) ? /\.md$/i : /\.md$/;

  function walk(directory) {
    const real = realpathSync(directory || ".");
    if (ancestors.has(real)) return;
    ancestors.add(real);
    for (const entry of readdirSync(directory || ".", { withFileTypes: true })) {
      const name = directory ? directory + "/" + entry.name : entry.name;
      if (entry.name === ".git" || excluded.ignores(name)) continue;
      let type = entry;
      if (entry.isSymbolicLink()) {
        try {
          type = statSync(name);
        } catch (error) {
          if (["ENOENT", "ENOTDIR", "ELOOP"].includes(error.code)) continue;
          throw error;
        }
      }
      if (type.isDirectory()) {
        if (!excluded.ignores(name + "/")) walk(name);
      } else if (type.isFile() && extension.test(name)) {
        files.push(name);
      }
    }
    ancestors.delete(real);
  }

  walk("");
  return files.sort();
}

function lintMarkdown() {
  const excluded = ignore();
  try {
    excluded.add(readFileSync(".markdownlintignore", "utf8"));
  } catch (error) {
    if (error.code !== "ENOENT") throw error;
  }
  const files = markdownFiles(excluded);

  if (files.length === 0) {
    console.warn("No matching files found");
    return;
  }

  for (const [file, findings] of Object.entries(lint({ files, config, configParsers: [load] }))) {
    for (const finding of findings) {
      const column = finding.errorRange ? `:${finding.errorRange[0]}` : "";
      const detail = finding.errorDetail ? ` [${finding.errorDetail}]` : "";
      const context = finding.errorContext ? ` [Context: "${finding.errorContext}"]` : "";
      console.error(`${file}:${finding.lineNumber}${column} ${finding.severity} ${finding.ruleNames.join("/")} ${finding.ruleDescription}${detail}${context}`);
      if (finding.severity === "error") process.exitCode = 1;
    }
  }
}

try {
  lintMarkdown();
} catch (error) {
  console.error(error.message);
  process.exitCode = 4;
}
