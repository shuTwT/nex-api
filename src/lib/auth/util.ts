import { randomBytes, scrypt } from "crypto";
import { AuthOptions, getServerSession } from "next-auth";
import { promisify } from "node:util";


export interface SessionUser {
  id: string;
  email: string;
  username: string;
  role: string;
}


const SALT_LENGTH = 16;
const KEY_LENGTH = 64;
const SCRYPT_OPTIONS = {
  N: 16384,
  r: 8,
  p: 1,
};

const randomBytesAsync = promisify(randomBytes);

export async function hashPassword(password: string): Promise<string> {
  const salt = (await randomBytesAsync(SALT_LENGTH)).toString("hex");
  const derivedKey = await new Promise<Buffer>((resolve, reject) => {
    scrypt(password, salt, KEY_LENGTH, SCRYPT_OPTIONS, (err, derivedKey) => {
      if (err) reject(err);
      else resolve(derivedKey as Buffer);
    });
  });

  return `${salt}:${derivedKey.toString("hex")}`;
}

export async function verifyPassword(
  password: string,
  hashedPassword: string,
): Promise<boolean> {
  const [salt, storedHash] = hashedPassword.split(":");

  if (!salt || !storedHash) {
    return false;
  }

  const derivedKey = await new Promise<Buffer>((resolve, reject) => {
    scrypt(password, salt, KEY_LENGTH, SCRYPT_OPTIONS, (err, derivedKey) => {
      if (err) reject(err);
      else resolve(derivedKey as Buffer);
    });
  });

  return derivedKey.toString("hex") === storedHash;
}

export async function requireAuth(authOptions: AuthOptions): Promise<SessionUser> {
  const session = await getServerSession(authOptions);
  if (!session?.user) {
    throw new Error("Unauthorized");
  }
  
  return session.user as SessionUser;
}

export async function requireAdmin(authOptions: AuthOptions): Promise<SessionUser> {
  const session = await getServerSession(authOptions);
  if(session==null){
    throw new Error("Unauthorized");
  }
  const user = session.user as SessionUser;

  if (user.role !== "admin") {
    throw new Error("Forbidden: Admin access required");
  }
  
  return user;
}
