export type LocalizedString = string | Record<string, string>;

/** Extract a display string from a possibly-localized field. */
export function locStr(val: LocalizedString | undefined | null): string {
  if (!val) return "";
  if (typeof val === "string") return val;
  return val.en || Object.values(val)[0] || "";
}

/** Safely extract an array from an API response, handling wrapped responses. */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function asArray<T>(res: any): T[] {
  const d = res?.data?.data ?? res?.data;
  if (Array.isArray(d)) return d;
  if (d && typeof d === "object" && Array.isArray(d.items)) return d.items;
  if (d && typeof d === "object" && Array.isArray(d.entries)) return d.entries;
  return [];
}
