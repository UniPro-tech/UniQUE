export const isClient = typeof window !== "undefined";

export function isMobile() {
  if (!isClient) return false;
  return window.innerWidth <= 768;
}
