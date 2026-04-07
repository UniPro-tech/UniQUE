export const ExistingMemberCheck = async (
  period: string,
  custom_id: string,
  external_email: string,
): Promise<{ valid: boolean; joinedAt?: Date }> => {
  const internalEmail = `${
    period && period !== "00" && period !== "0"
      ? `${period.toUpperCase()}.`
      : ""
  }${custom_id}@uniproject.jp`;

  const verifyRes = await fetch(
    `${process.env.GAS_MIGRATE_API_URL}?external_email=${encodeURIComponent(
      external_email,
    )}&internal_email=${encodeURIComponent(internalEmail)}`,
    { method: "GET" },
  );
  const verifyData = await verifyRes.json();
  const joinedAt = new Date(verifyData.joined_at);
  if (verifyRes.ok) {
    return { valid: true, joinedAt };
  }
  return { valid: false };
};
