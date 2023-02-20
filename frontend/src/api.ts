export const newSession = async () => {
  const r = await fetch("/api/session", { method: "POST" });
  return (await r.json()) as {
    Address: string;
  };
};
