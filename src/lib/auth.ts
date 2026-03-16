import { randomBytes, scrypt } from "crypto";
import { promisify } from "util";

const randomBytesAsync = promisify(randomBytes);

const SALT_LENGTH = 16;
const KEY_LENGTH = 64;
const SCRYPT_OPTIONS = {
  N: 16384,
  r: 8,
  p: 1,
};

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
  hashedPassword: string
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

export function generateToken(length: number = 32): string {
  return randomBytes(length).toString("hex");
}
