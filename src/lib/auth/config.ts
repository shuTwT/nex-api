import { randomBytes, scrypt } from "crypto";
import type { NextAuthOptions } from "next-auth";
import GitHubProvider from "next-auth/providers/github";
import { getServerSession } from "next-auth";
import type {
  GetServerSidePropsContext,
  NextApiRequest,
  NextApiResponse,
} from "next";
import Credentials from "next-auth/providers/credentials";
import { PrismaAdapter } from "@auth/prisma-adapter";
import prisma from "@/lib/prisma";
import { type SessionUser, hashPassword, verifyPassword } from "./util";






export function generateToken(length: number = 32): string {
  return randomBytes(length).toString("hex");
}

export const authOptions: NextAuthOptions = {
  session: {
    strategy: "jwt",
  },
  secret: process.env.NEXTAUTH_SECRET!,
  theme:{
    colorScheme:"light"
  },
  pages: {
    signIn: '/auth/signin',
    signOut: '/auth/signout',
     error: '/auth/error', // Error code passed in query string as ?error=
     verifyRequest: '/auth/verify-request', // (used for check email message)
  //   newUser: '/auth/new-user' // New users will be directed here on first sign in (leave the property out if not of interest)
  },
  adapter: PrismaAdapter(prisma),
  providers: [
    Credentials({
      credentials: {
        email: {},
        password: {},
      },
      authorize: async (credentials) => {
        console.log("credentials", credentials);
        let user = null;

        if (!credentials) {
          return null;
        }

        const pwHash = await hashPassword(credentials.password);
        // console.log("pwHash",pwHash)
        //!TODO 验证用户是否存在
        user = await prisma.user.findUnique({
          where: {
            email: credentials.email,
          },
        });

        if (!user) {
          return null;
        }

        console.log("isMatch", credentials.password, user.password);
        const isMatch = await verifyPassword(
          credentials.password,
          user.password,
        );

        if (!isMatch) {
          return null;
        }

        // console.log(user);
        return user;
      },
    }),
    GitHubProvider({
      clientId: process.env.GITHUB_OAUTH_CLIENT_ID!,
      clientSecret: process.env.GITHUB_OAUTH_CLIENT_SECRET!,
    }),
    {
      id: "easy1auth",
      name: "Easy1Auth",
      type: "oauth",
      authorization: process.env.SSO_AUTHORIZATION_URL,
      token: process.env.SSO_TOKEN_URL,
      userinfo: process.env.SSO_USER_INFO_URL,
      clientId: process.env.SSO_CLIENT_ID!,
      clientSecret: process.env.SSO_CLIENT_SECRET!,
      profile(profile) {
        return {
          id: profile.sub,
          name: profile.name,
          email: profile.email,
          role: "admin",
        };
      },
    },
  ],
  callbacks: {
    async signIn({ user, account, profile, email, credentials }) {
      return true;
    },
    async redirect({ url, baseUrl }) {
      // Allows relative callback URLs
      if (url.startsWith("/")) return `${baseUrl}${url}`;
      // Allows callback URLs on the same origin
      else if (new URL(url).origin === baseUrl) return url;
      return baseUrl;
    },
    async session({ session, user, token }) {
      // console.log("session_callback_user", user);
      // console.log("session_callback_token", token);
      if (session?.user) {
        (session.user as SessionUser).role = token.role as string;
        (session.user as SessionUser).id = token.id as string;
      }
      // console.log("session_callback_session", session);
      return session;
    },
    async jwt({ token, user, account, profile, trigger }) {
      // 登录
      // console.log("jwt_callback_user", user);
      // console.log("jwt_callback_account", account);
      // console.log("jwt_callback_profile", profile);
      if (user) {
        token.role = (user as SessionUser).role;
        token.id = (user as SessionUser).id;
        token.name = (user as SessionUser).username;
      }

      // console.log("jwt_callback_token", token);
      return token;
    },
  },
};

export function auth(
  ...args:
    | [GetServerSidePropsContext["req"], GetServerSidePropsContext["res"]]
    | [NextApiRequest, NextApiResponse]
    | []
) {
  return getServerSession(...args, authOptions);
}
