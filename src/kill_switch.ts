import { stopMachine } from "./fly/api";

export const killFunc = async () => {
  console.log("nobody is here. Shutting down...");
  try {
    await stopMachine(process.env.FLY_ALLOC_ID!, process.env.FLY_APP_NAME!);
  } catch (ex: any) {
    console.error("failed to stop machine", ex);
  }
};
let killSwitch = setTimeout(killFunc, 30_000);
export const resetKillSwitch = () => {
  clearInterval(killSwitch);
  killSwitch = setTimeout(killFunc, 30_000);
};
