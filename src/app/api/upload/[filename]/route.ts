import { NextRequest, NextResponse } from "next/server";
import { readFile, stat } from "fs/promises";
import { existsSync } from "fs";
import path from "path";

const UPLOAD_DIR = path.join(process.cwd(), "data", "upload");

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ filename: string }> }
) {
  try {
    const { filename } = await params;
    const filepath = path.join(UPLOAD_DIR, filename);

    if (!existsSync(filepath)) {
      return NextResponse.json(
        { success: false, error: "文件不存在" },
        { status: 404 }
      );
    }

    const fileBuffer = await readFile(filepath);
    const fileStat = await stat(filepath);

    const ext = filename.split(".").pop()?.toLowerCase() || "";
    const contentType = getContentType(ext);

    const headers = new Headers();
    headers.set("Content-Type", contentType);
    headers.set("Content-Length", fileStat.size.toString());
    headers.set("Cache-Control", "public, max-age=31536000, immutable");
    headers.set("Last-Modified", fileStat.mtime.toUTCString());

    return new NextResponse(fileBuffer, {
      status: 200,
      headers,
    });
  } catch (error) {
    console.error("File serve error:", error);
    return NextResponse.json(
      { success: false, error: "文件读取失败" },
      { status: 500 }
    );
  }
}

function getContentType(ext: string): string {
  const contentTypes: Record<string, string> = {
    jpg: "image/jpeg",
    jpeg: "image/jpeg",
    png: "image/png",
    gif: "image/gif",
    webp: "image/webp",
    svg: "image/svg+xml",
    pdf: "application/pdf",
    txt: "text/plain",
    json: "application/json",
  };

  return contentTypes[ext] || "application/octet-stream";
}
