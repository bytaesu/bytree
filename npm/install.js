const https = require("https");
const fs = require("fs");
const path = require("path");

const VERSION = require("./package.json").version;
const REPO = "bytaesu/bytree";

async function install() {
  const { url, isZip } = getDownloadInfo();
  const ext = process.platform === "win32" ? ".exe" : "";
  const dest = path.join(__dirname, "bin", `bytree${ext}`);

  console.log(`Downloading bytree v${VERSION}...`);
  await download(url, dest, isZip);
  if (process.platform !== "win32") {
    fs.chmodSync(dest, 0o755);
  }
  console.log("bytree installed successfully.");
}

function getDownloadInfo() {
  const platform = process.platform;
  const arch = process.arch;

  const osMap = { darwin: "darwin", linux: "linux", win32: "windows" };
  const archMap = { x64: "amd64", arm64: "arm64" };

  const goOS = osMap[platform];
  const goArch = archMap[arch];

  if (!goOS || !goArch) {
    throw new Error(`Unsupported platform: ${platform}/${arch}`);
  }

  const isZip = platform === "win32";
  const ext = isZip ? ".zip" : ".tar.gz";
  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/bytree_${VERSION}_${goOS}_${goArch}${ext}`;
  return { url, isZip };
}

function download(url, dest, isZip, redirects = 0) {
  if (redirects > 5) {
    return Promise.reject(new Error("Too many redirects"));
  }
  return new Promise((resolve, reject) => {
    https
      .get(url, (res) => {
        if (res.statusCode === 302 || res.statusCode === 301) {
          return download(res.headers.location, dest, isZip, redirects + 1)
            .then(resolve, reject);
        }

        if (res.statusCode !== 200) {
          return reject(
            new Error(`Download failed: HTTP ${res.statusCode}\n  ${url}`)
          );
        }

        if (isZip) {
          extractZip(res, dest).then(resolve, reject);
        } else {
          extractTarGz(res, dest).then(resolve, reject);
        }
      })
      .on("error", reject);
  });
}

function extractTarGz(stream, dest) {
  const zlib = require("zlib");
  const gunzip = zlib.createGunzip();
  const chunks = [];

  return new Promise((resolve, reject) => {
    stream.on("error", reject);
    stream
      .pipe(gunzip)
      .on("data", (chunk) => chunks.push(chunk))
      .on("end", () => {
        const buffer = Buffer.concat(chunks);
        const binary = extractFromTar(buffer);
        if (!binary) {
          return reject(new Error("Could not find bytree binary in archive"));
        }
        fs.writeFileSync(dest, binary);
        resolve();
      })
      .on("error", reject);
  });
}

function extractFromTar(buffer) {
  let offset = 0;
  while (offset < buffer.length - 512) {
    const header = buffer.slice(offset, offset + 512);

    if (header.every((b) => b === 0)) break;

    const name = header.toString("utf8", 0, 100).replace(/\0/g, "");
    const sizeStr = header
      .toString("utf8", 124, 136)
      .replace(/\0/g, "")
      .trim();
    const size = parseInt(sizeStr, 8) || 0;

    offset += 512;

    if (name.includes("bytree") && !name.endsWith("/") && size > 0) {
      return buffer.slice(offset, offset + size);
    }

    offset += Math.ceil(size / 512) * 512;
  }
  return null;
}

function extractZip(stream, dest) {
  const chunks = [];

  return new Promise((resolve, reject) => {
    stream.on("error", reject);
    stream
      .on("data", (chunk) => chunks.push(chunk))
      .on("end", () => {
        const buffer = Buffer.concat(chunks);
        const binary = extractFromZip(buffer);
        if (!binary) {
          return reject(new Error("Could not find bytree.exe in archive"));
        }
        fs.writeFileSync(dest, binary);
        resolve();
      });
  });
}

function extractFromZip(buffer) {
  // Simple ZIP extraction - find bytree.exe in local file headers
  let offset = 0;
  while (offset < buffer.length - 30) {
    // Local file header signature: 0x04034b50
    if (
      buffer[offset] !== 0x50 ||
      buffer[offset + 1] !== 0x4b ||
      buffer[offset + 2] !== 0x03 ||
      buffer[offset + 3] !== 0x04
    ) {
      break;
    }

    const nameLen = buffer.readUInt16LE(offset + 26);
    const extraLen = buffer.readUInt16LE(offset + 28);
    const compSize = buffer.readUInt32LE(offset + 18);
    const name = buffer.toString("utf8", offset + 30, offset + 30 + nameLen);

    const dataStart = offset + 30 + nameLen + extraLen;

    if (name.includes("bytree") && !name.endsWith("/") && compSize > 0) {
      const method = buffer.readUInt16LE(offset + 8);
      const data = buffer.slice(dataStart, dataStart + compSize);
      if (method === 0) {
        return data;
      }
      if (method === 8) {
        const zlib = require("zlib");
        return zlib.inflateRawSync(data);
      }
    }

    offset = dataStart + compSize;
  }
  return null;
}

install().catch((err) => {
  console.error("Failed to install bytree:", err.message);
  console.error("You can install manually from:");
  console.error(`  https://github.com/${REPO}/releases`);
  process.exit(1);
});
