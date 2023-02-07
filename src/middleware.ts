// middleware.ts
import type { NextRequest } from "next/server";
import { resetKillSwitch } from "./kill_switch";

// This function can be marked `async` if using `await` inside
export function middleware(request: NextRequest) {
  resetKillSwitch();
}
