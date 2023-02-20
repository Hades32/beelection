import { useSignal } from "@preact/signals";
import { h } from "preact";
import { route } from "preact-router";
import { newSession } from "../api";

const Home = () => {
  const startSwarmDisabled = useSignal(false);
  const startSwarm = async () => {
    startSwarmDisabled.value = true;
    const s = await newSession();
    route(`/swarm/${btoa(s.Address)}`);
  };

  return (
    <div class="flex flex-col items-center">
      <img
        class="m-8"
        src="../../assets/preact-logo.svg"
        alt="Preact Logo"
        height="160"
        width="160"
      />
      <h1 class="text-2xl m-2">Let's beelect!</h1>
      <p class="m-2">
        Click below to start a new swarm. Then share the URL with your friends
      </p>
      <button
        onClick={startSwarm}
        disabled={startSwarmDisabled}
        class={`m-2 px-4 py-2 bg-amber-700 disabled:bg-stone-700 disabled:text-stone-300 hover:bg-amber-600 focus:bg-amber-800 text-amber-200 font-semibold rounded ${
          startSwarmDisabled.value ? "pulsate" : ""
        }`}
      >
        Start
      </button>
    </div>
  );
};

export default Home;
