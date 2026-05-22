import fs from "node:fs";
import path from "node:path";
import nacl from "tweetnacl";
import * as util from "tweetnacl-util";

// 秘密鍵をファイルから読み込む
// 存在しない場合は新規作成
const secretKeyFile =
  process.env.CSRF_SECRET_KEY_PATH ||
  path.join(process.cwd(), "keys", "csrf_secret.key");
const publicKeyFile =
  process.env.CSRF_PUBLIC_KEY_PATH ||
  path.join(process.cwd(), "keys", "csrf_public.key");
let secretKey: Uint8Array<ArrayBufferLike>,
  publicKey: Uint8Array<ArrayBufferLike>;

if (
  fs.existsSync(/*turbopackIgnore: true*/ secretKeyFile) &&
  fs.existsSync(/*turbopackIgnore: true*/ publicKeyFile)
) {
  secretKey = util.decodeBase64(
    fs.readFileSync(/*turbopackIgnore: true*/ secretKeyFile, "utf8"),
  );
  publicKey = secretKey.slice(32); // tweetnaclの秘密鍵は32+32=64バイト
} else {
  if (!fs.existsSync(/*turbopackIgnore: true*/ path.dirname(secretKeyFile))) {
    fs.mkdirSync(/*turbopackIgnore: true*/ path.dirname(secretKeyFile), {
      recursive: true,
    });
  }
  const keyPair = nacl.sign.keyPair();
  secretKey = keyPair.secretKey;
  publicKey = keyPair.publicKey;
  try {
    const fdPubkey = fs.openSync(
      /*turbopackIgnore: true*/
      publicKeyFile,
      fs.constants.O_CREAT | fs.constants.O_EXCL | fs.constants.O_RDWR,
      0o600,
    );
    fs.writeFileSync(fdPubkey, util.encodeBase64(publicKey), {
      mode: 0o600,
    });
    const fdSeckey = fs.openSync(
      /*turbopackIgnore: true*/
      secretKeyFile,
      fs.constants.O_CREAT | fs.constants.O_EXCL | fs.constants.O_RDWR,
      0o600,
    );
    fs.writeFileSync(fdSeckey, util.encodeBase64(secretKey), {
      mode: 0o600,
    });
  } catch {}
}

export const generateCSRFToken = (
  data: string,
  exp: boolean | undefined = true,
): string => {
  const message = util.decodeUTF8(
    `${data}${exp ? `&exp=${Date.now() + 10 * 60 * 1000}` : ""}`,
  ); // 10分有効
  const signedMessage = nacl.sign(message, secretKey);
  return util.encodeBase64(signedMessage);
};

export const VerifyCSRFToken = (
  token: string,
  exp: boolean | undefined = true,
): string | null => {
  try {
    const verified = nacl.sign.open(util.decodeBase64(token), publicKey);
    if (verified) {
      const expire = parseInt(util.encodeUTF8(verified).split("&exp=")[1], 10);
      if (Date.now() > expire && !exp) {
        return null;
      }
      return util.encodeUTF8(verified).split("&exp=")[0];
    } else {
      return null;
    }
  } catch (_e) {
    //TODO: エラーログを記録
    return null;
  }
};
