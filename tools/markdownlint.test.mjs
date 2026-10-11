import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { copyFileSync, mkdtempSync, mkdirSync, rmSync, writeFileSync, readFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { test } from "node:test";
import { fileURLToPath } from "node:url";

const repository = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const manifest = JSON.parse(readFileSync(path.join(repository, "package.json"), "utf8"));
const ordinary = "# Fixture\n\nOrdinary text.\n";
const invalid = "# Fixture\n\nTrailing space \n";

function fixture(t) {
  const root = mkdtempSync(path.join(tmpdir(), "fartapp-markdownlint-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  for (const name of [".markdownlint.yaml", ".markdownlintignore"]) {
    copyFileSync(path.join(repository, name), path.join(root, name));
  }
  writeFileSync(path.join(root, "package.json"), JSON.stringify({
    private: true,
    scripts: { "lint:markdown": manifest.scripts["lint:markdown"] },
  }));
  return root;
}

function put(root, name, text) {
  const target = path.join(root, name);
  mkdirSync(path.dirname(target), { recursive: true });
  writeFileSync(target, text);
}

function lint(root) {
  assert.ok(process.env.npm_execpath, "run through npm run test:markdown");
  const result = spawnSync(process.execPath,
    [process.env.npm_execpath, "run", "lint:markdown", "--silent"], {
      cwd: root,
      env: {
        ...process.env,
        PATH: [path.join(repository, "node_modules", ".bin"),
          path.dirname(process.execPath), process.env.PATH].join(path.delimiter),
      },
      encoding: "utf8",
      timeout: 20000,
    });
  assert.ifError(result.error);
  return { status: result.status, output: result.stdout + result.stderr };
}

test("configured rules and math syntax remain accepted", (t) => {
  const root = fixture(t);
  put(root, "README.md", "# Fixture\n\n" + "ordinary ".repeat(30).trimEnd() +
    "\n\n_Emphasized paragraph_\n\nInline math $a+b$ and display math:\n\n$$\na+b=c\n$$\n");
  assert.deepEqual(lint(root), { status: 0, output: "" });
});

test("root exclusions remain excluded", (t) => {
  const root = fixture(t);
  put(root, "README.md", ordinary);
  for (const name of ["node_modules/invalid.md", "vendor/invalid.md", "target/invalid.md",
    "artifacts/invalid.md", ".steward/DECISIONS/invalid.md"]) {
    put(root, name, invalid);
  }
  assert.deepEqual(lint(root), { status: 0, output: "" });
});

for (const name of ["docs/invalid.md", ".github/invalid.md", ".steward/ROADMAP.md",
  "docs/vendor/included.md"]) {
  test(`lint errors in ${name} remain failures`, (t) => {
    const root = fixture(t);
    put(root, "README.md", ordinary);
    put(root, name, invalid);
    const result = lint(root);
    assert.equal(result.status, 1, result.output);
    assert.match(result.output, /MD009/);
    assert.ok(result.output.includes(name), result.output);
  });
}

for (const missing of [false, true]) {
  test(`${missing ? "missing" : "malformed"} explicit config fails`, (t) => {
    const root = fixture(t);
    put(root, "README.md", ordinary);
    if (missing) rmSync(path.join(root, ".markdownlint.yaml"));
    else put(root, ".markdownlint.yaml", "MD013: [\n");
    const result = lint(root);
    assert.equal(result.status, 4, result.output);
    assert.match(result.output, /Cannot read or parse config/);
  });
}
