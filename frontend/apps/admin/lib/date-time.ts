type DateLike = string | number | Date;

type TimeFormatOptions = {
  hour12?: boolean;
  includeSeconds?: boolean;
  fallback?: string;
};

function parseDate(value: DateLike): Date | null {
  const parsed = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return null;
  }
  return parsed;
}

export function formatLocalDate(value: DateLike, fallback = "-"): string {
  const parsed = parseDate(value);
  if (!parsed) {
    return fallback;
  }
  return parsed.toLocaleDateString();
}

export function formatLocalDateTime(value: DateLike, fallback = "-"): string {
  const parsed = parseDate(value);
  if (!parsed) {
    return fallback;
  }
  return parsed.toLocaleString();
}

export function formatLocalTime(
  value: DateLike,
  options: TimeFormatOptions = {},
): string {
  const {
    hour12 = false,
    includeSeconds = false,
    fallback = "--:--",
  } = options;
  const parsed = parseDate(value);
  if (!parsed) {
    return fallback;
  }

  return parsed.toLocaleTimeString([], {
    hour12,
    hour: "2-digit",
    minute: "2-digit",
    ...(includeSeconds ? { second: "2-digit" } : {}),
  });
}
