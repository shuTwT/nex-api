import { NextRequest, NextResponse } from "next/server";
import { randomBytes } from "crypto";

export async function GET(request: NextRequest) {
  try {
    const baseUrl = request.nextUrl.origin;
    const callbackUrl = `${baseUrl}/api/auth/sso/callback`;
    
    // 生成 state 参数用于 CSRF 保护
    const state = randomBytes(32).toString("hex");
    
    // 将 state 存储到 cookie 中，稍后在回调中验证
    const response = NextResponse.redirect(
      `${process.env.SSO_AUTHORIZATION_URL}?` +
        new URLSearchParams({
          client_id: process.env.SSO_CLIENT_ID!,
          redirect_uri: callbackUrl,
          response_type: "code",
          scope: process.env.SSO_SCOPE || "openid profile email",
          state: state,
        })
    );
    
    // 设置 state cookie，有效期 10 分钟
    response.cookies.set("sso_state", state, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      maxAge: 600,
      path: "/",
    });
    
    return response;
  } catch (error) {
    console.error("SSO authorization error:", error);
    return NextResponse.redirect(new URL("/?error=sso_auth_failed", request.nextUrl.origin));
  }
}
