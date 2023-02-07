// middleware.ts
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { stopMachine } from "./fly/api";

const killFunc = () => {
  console.log("nobody is here. Shutting down...");
  stopMachine(process.env.FLY_ALLOC_ID!, process.env.FLY_APP_NAME!);
};
let killSwitch = setInterval(killFunc, 30_000);

// This function can be marked `async` if using `await` inside
export function middleware(request: NextRequest) {
  clearInterval(killSwitch);
  killSwitch = setInterval(killFunc, 30_000);
}
