/**
 * ================================
 * User Status
 * ================================
 */

export const USER_STATUS_OPTIONS = [
  { value: "established", label: "Established" },
  { value: "active", label: "Active" },
  { value: "suspended", label: "Suspended" },
  { value: "archived", label: "Archived" },
] as const;

export type UserStatus = (typeof USER_STATUS_OPTIONS)[number]["value"];

export function getStatusLabel(status?: UserStatus | string | null): string {
  if (!status) return "";
  const option = USER_STATUS_OPTIONS.find((opt) => opt.value === status);
  return option?.label ?? status;
}

/**
 * ================================
 * Affiliation Period
 * ================================
 */

/**
 * 4月始まり
 * 0期 = 2024年1〜3月
 */
export const AFFILIATION_START_YEAR = 2024;

export const AFFILIATION_QUARTERS = [
  { month: 4, name: "A" },
  { month: 7, name: "B" },
  { month: 10, name: "C" },
  { month: 1, name: "D" }, // 翌暦年1月
] as const;

export type QuarterName = (typeof AFFILIATION_QUARTERS)[number]["name"];

/**
 * 現在の年度（4月始まり）を取得
 */
function getCurrentFiscalYear(date = new Date()): number {
  const month = date.getMonth() + 1;
  return month >= 4 ? date.getFullYear() : date.getFullYear() - 1;
}

/**
 * 選択可能な所属期を取得
 */
export function getSelectableAffiliationPeriods(): Array<{
  value: string;
  label: string;
}> {
  const now = new Date();
  const currentMonth = now.getMonth() + 1;
  const currentFiscalYear = getCurrentFiscalYear(now);

  const periods: string[] = ["0"]; // 0期（2024年1-3月）

  for (
    let fiscalYear = AFFILIATION_START_YEAR;
    fiscalYear <= currentFiscalYear;
    fiscalYear++
  ) {
    const periodNumber = fiscalYear - AFFILIATION_START_YEAR + 1;

    const padded = periodNumber < 10 ? `0${periodNumber}` : `${periodNumber}`;

    for (const q of AFFILIATION_QUARTERS) {
      // 今年度の未来分は除外
      if (fiscalYear === currentFiscalYear) {
        const fiscalMonthNow =
          currentMonth >= 4 ? currentMonth : currentMonth + 12;

        const fiscalMonthQuarter = q.month === 1 ? 13 : q.month;

        if (fiscalMonthNow < fiscalMonthQuarter) {
          continue;
        }
      }

      periods.push(`${padded}${q.name}`);
    }
  }

  return periods.map((period) => ({
    value: period,
    label: period,
  }));
}

/**
 * 所属期ラベルを取得
 */
export function getAffiliationPeriodLabel(period?: string | null): string {
  if (!period) return "";
  return period;
}

/**
 * 所属期の詳細情報を取得
 */
export function getAffiliationPeriodInfo(period: string): {
  isEarly: boolean;
  fiscalYear: number;
  calendarYear: number;
  month?: number;
  label: string;
} {
  if (period === "0") {
    return {
      isEarly: true,
      fiscalYear: AFFILIATION_START_YEAR,
      calendarYear: AFFILIATION_START_YEAR,
      month: undefined,
      label: "早期活動者",
    };
  }

  if (period.length < 2) {
    throw new Error("Invalid period format");
  }

  const letter = period.slice(-1);
  const numPart = period.slice(0, -1);
  const num = parseInt(numPart, 10);

  if (Number.isNaN(num)) {
    throw new Error("Invalid period number");
  }

  const quarter = AFFILIATION_QUARTERS.find((q) => q.name === letter);

  if (!quarter) {
    throw new Error("Invalid quarter letter");
  }

  const fiscalYear = AFFILIATION_START_YEAR + num - 1;

  const calendarYear = quarter.month === 1 ? fiscalYear + 1 : fiscalYear;

  return {
    isEarly: false,
    fiscalYear,
    calendarYear,
    month: quarter.month,
    label: `${quarter.month}月以降`,
  };
}
