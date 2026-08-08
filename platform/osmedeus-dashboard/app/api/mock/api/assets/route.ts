import { NextResponse } from "next/server";
import { mockHttpAssets } from "@/lib/mock/data/http-assets";

export const dynamic = "force-static";

export async function GET() {
  const workspace = "ws-001";
  const offset = 0;
  const limit = 20;

  const all = mockHttpAssets[workspace] || [];
  const sliced = all.slice(offset, offset + limit);

  const data = sliced.map((a) => ({
    id: a.id,
    workspace: workspace,
    host: a.url.replace(/^https?:\/\//, "").split("/")[0],
    url: a.url,
    status_code: a.statusCode,
    title: a.title,
    technologies: a.technologies,
    created_at: a.createdAt.toISOString(),
  }));

  return NextResponse.json({
    data,
    pagination: {
      total: all.length,
      offset,
      limit,
    },
  });
}
