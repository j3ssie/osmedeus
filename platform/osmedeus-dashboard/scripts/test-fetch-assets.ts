import { fetchHttpAssets } from "../lib/api/assets";

async function main() {
  process.env.NEXT_PUBLIC_USE_MOCK = "true";
  const resAll = await fetchHttpAssets(undefined, { page: 1, pageSize: 10 });
  console.log("All Workspaces -> count:", resAll.data.length, "total:", resAll.pagination.totalItems);
  console.log("Sample:", resAll.data[0]);

  const resWs = await fetchHttpAssets("ws-001", { page: 1, pageSize: 10 });
  console.log("Workspace ws-001 -> count:", resWs.data.length, "total:", resWs.pagination.totalItems);
  console.log("Sample:", resWs.data[0]);
}

main().catch((e) => {
  console.error("Test failed:", e);
  process.exit(1);
});
