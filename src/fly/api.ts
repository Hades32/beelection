export class BadStatusError extends Error {
  constructor(public Status: number) {
    super("bad API response code");
  }
}

export async function stopMachine(id: string, app: string) {
  const resp = await fetch(
    `http://${process.env.FLY_API_HOSTNAME}/v1/apps/user-functions/machines/${id}/stop`,
    {
      headers: {
        Authorization: `Bearer ${process.env.FLY_API_TOKEN}`,
        "Content-Type": "application/json",
      },
      method: "POST",
    }
  );
  if (resp.status !== 200) {
    throw new BadStatusError(resp.status);
  }
}
