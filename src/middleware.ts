// middleware.ts
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const killFunc = () => {
  console.log("nobody is here. Shutting down...");
  process.exit(0);
};
let killSwitch = setInterval(killFunc, 30_000);

// This function can be marked `async` if using `await` inside
export function middleware(request: NextRequest) {
  clearInterval(killSwitch);
  killSwitch = setInterval(killFunc, 30_000);
}
