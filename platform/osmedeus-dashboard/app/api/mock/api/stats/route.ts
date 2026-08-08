import { NextResponse } from "next/server";
import { mockSystemStats } from "@/lib/mock/data/system-stats";

export const dynamic = "force-static";

export async function GET() {
  return NextResponse.json(mockSystemStats);
}
