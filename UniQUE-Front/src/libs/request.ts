export const parseUA = (userAgent: string) => {
  // Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36
  const ua = userAgent.toLowerCase();
  let browser = "Unknown";
  let os = "Unknown";

  if (ua.includes("chrome")) {
    browser = "Google Chrome";
  } else if (ua.includes("safari")) {
    browser = "Safari";
  } else if (ua.includes("firefox")) {
    browser = "Mozilla Firefox";
  } else if (ua.includes("edg")) {
    browser = "Microsoft Edge";
  }

  if (ua.includes("mac os x")) {
    os = "macOS";
  } else if (ua.includes("windows nt")) {
    os = "Windows";
  } else if (ua.includes("linux")) {
    os = "Linux";
  } else if (ua.includes("iphone") || ua.includes("ipad")) {
    os = "iOS";
  } else if (ua.includes("android")) {
    os = "Android";
  }

  return { browser, os };
};

export const getRealIPAddress = async (): Promise<string> => {
  const { headers } = await import("next/headers");
  const headersList = await headers();

  // CloudflareのCF-Connecting-IPヘッダーを優先
  const cfConnectingIp = headersList.get("cf-connecting-ip");
  if (cfConnectingIp) {
    return cfConnectingIp;
  }

  // フォールバック: X-Forwarded-For
  const xForwardedFor = headersList.get("x-forwarded-for");
  if (xForwardedFor) {
    // 複数のIPがある場合は最初のものを取得
    return xForwardedFor.split(",")[0].trim();
  }

  // フォールバック: X-Real-IP
  const xRealIp = headersList.get("x-real-ip");
  if (xRealIp) {
    return xRealIp;
  }

  // すべて取得できない場合は不明として返す
  return "unknown";
};
