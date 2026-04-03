#!/usr/bin/env node

const { spawn } = require("child_process");
const path = require("path");
const fs = require("fs");

const ext = process.platform === "win32" ? ".exe" : "";
const bin = path.join(__dirname, `bytree${ext}`);

if (!fs.existsSync(bin)) {
  console.error(
    "bytree binary not found. Try reinstalling: npm install -g bytree"
  );
  process.exit(1);
}

const child = spawn(bin, process.argv.slice(2), { stdio: "inherit" });
child.on("error", (err) => {
  console.error(`Failed to run bytree: ${err.message}`);
  process.exit(1);
});
child.on("close", (code, signal) => {
  if (signal) process.exit(1);
  process.exit(code ?? 1);
});
