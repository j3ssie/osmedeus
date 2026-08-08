import { NextResponse } from "next/server";

export const dynamic = "force-static";

export async function POST(request: Request) {
  try {
    const body = await request.json().catch(() => ({}));
    const username = body?.username ?? "osmedeus";
    const token = `mock-${Buffer.from(username).toString("base64")}`;
    return NextResponse.json({ token });
  } catch {
    return NextResponse.json({ error: true, message: "Invalid request" }, { status: 400 });
  }
}
