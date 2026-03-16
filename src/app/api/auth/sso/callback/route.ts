import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { createSessionToken, setSessionCookie } from "@/lib/session";

export async function GET(request: NextRequest) {
  try {
    const searchParams = request.nextUrl.searchParams;
    const code = searchParams.get("code");
    const state = searchParams.get("state");
    const error = searchParams.get("error");

    // 检查是否有错误
    if (error) {
      console.error("SSO error:", error);
      return NextResponse.redirect(new URL("/?error=sso_auth_failed", request.nextUrl.origin));
    }

    // 验证 state 参数（CSRF 保护）
    const savedState = request.cookies.get("sso_state")?.value;
    if (!state || !savedState || state !== savedState) {
      console.error("Invalid SSO state parameter");
      return NextResponse.redirect(new URL("/?error=invalid_state", request.nextUrl.origin));
    }

    if (!code) {
      console.error("No authorization code received");
      return NextResponse.redirect(new URL("/?error=no_code", request.nextUrl.origin));
    }

    // 使用授权码交换 access token
    const tokenResponse = await fetch(process.env.SSO_TOKEN_URL!, {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
      },
      body: new URLSearchParams({
        grant_type: "authorization_code",
        code: code,
        client_id: process.env.SSO_CLIENT_ID!,
        client_secret: process.env.SSO_CLIENT_SECRET!,
        redirect_uri: `${request.nextUrl.origin}/api/auth/sso/callback`,
      }),
    });

    if (!tokenResponse.ok) {
      const errorData = await tokenResponse.text();
      console.error("Token exchange failed:", errorData);
      return NextResponse.redirect(new URL("/?error=token_exchange_failed", request.nextUrl.origin));
    }

    const tokenData = await tokenResponse.json();
    const accessToken = tokenData.access_token;

    // 使用 access token 获取用户信息
    const userInfoResponse = await fetch(process.env.SSO_USER_INFO_URL!, {
      method: "GET",
      headers: {
        "Authorization": `Bearer ${accessToken}`,
      },
    });

    if (!userInfoResponse.ok) {
      const errorData = await userInfoResponse.text();
      console.error("User info fetch failed:", errorData);
      return NextResponse.redirect(new URL("/?error=user_info_failed", request.nextUrl.origin));
    }

    const userInfo = await userInfoResponse.json();

    // 从用户信息中提取邮箱和用户名
    const email = userInfo.email || userInfo.preferred_username;
    const username = userInfo.preferred_username || userInfo.name || email.split("@")[0];

    console.log(`用户登录:邮箱=${email},用户名=${username}`)

    if (!email) {
      console.error("No email received from SSO provider");
      return NextResponse.redirect(new URL("/?error=no_email", request.nextUrl.origin));
    }

    // 查找或创建用户
    let user = await prisma.user.findUnique({
      where: { email },
      select: {
        id: true,
        email: true,
        username: true,
        role: true,
        credits: true,
      },
    });

    if (!user) {
      // 创建新用户
      user = await prisma.user.create({
        data: {
          email,
          username,
          password: `oauth_${userInfo.sub || "sso"}`, // OAuth 用户不需要密码
          role: "user",
          credits: 1000, // 初始积分
        },
        select: {
          id: true,
          email: true,
          username: true,
          role: true,
          credits: true,
        },
      });
    }

    // 保存或更新 OAuth 账号关联信息
    await prisma.account.upsert({
      where: {
        provider_providerAccountId: {
          provider: "sso",
          providerAccountId: userInfo.sub || email,
        },
      },
      update: {
        access_token: accessToken,
        refresh_token: tokenData.refresh_token,
        expires_at: tokenData.expires_in ? Math.floor(Date.now() / 1000) + tokenData.expires_in : undefined,
        token_type: tokenData.token_type,
        scope: tokenData.scope,
        id_token: tokenData.id_token,
      },
      create: {
        userId: user.id,
        provider: "sso",
        providerAccountId: userInfo.sub || email,
        access_token: accessToken,
        refresh_token: tokenData.refresh_token,
        expires_at: tokenData.expires_in ? Math.floor(Date.now() / 1000) + tokenData.expires_in : undefined,
        token_type: tokenData.token_type,
        scope: tokenData.scope,
        id_token: tokenData.id_token,
      },
    });

    // 创建会话
    const token = createSessionToken({
      id: user.id,
      email: user.email,
      username: user.username,
      role: user.role,
    });

    await setSessionCookie(token);

    // 清除 state cookie
    const response = NextResponse.redirect(new URL("/console", request.nextUrl.origin));
    response.cookies.delete("sso_state");

    return response;
  } catch (error) {
    console.error("SSO callback error:", error);
    return NextResponse.redirect(new URL("/?error=sso_callback_failed", request.nextUrl.origin));
  }
}
