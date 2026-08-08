import fs from "node:fs";
import path from "node:path";

function rm(p) {
  try {
    fs.rmSync(p, { recursive: true, force: true });
  } catch {
    return;
  }
}

function exists(p) {
  try {
    return fs.existsSync(p);
  } catch {
    return false;
  }
}

function deleteMaps(dir) {
  if (!exists(dir)) return;
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  for (const entry of entries) {
    const p = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      deleteMaps(p);
      continue;
    }
    if (entry.isFile() && p.endsWith(".map")) {
      try {
        fs.unlinkSync(p);
      } catch {
        return;
      }
    }
  }
}

rm("build/dev");
rm("build/cache");
rm("build/dev/cache");
rm(".next/cache");
deleteMaps("build");
deleteMaps("out");

