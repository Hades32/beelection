import { useSignal } from "@preact/signals";
import { h } from "preact";
import { useEffect } from "preact/hooks";

interface Props {
  id: string;
}

const clientID = (() => {
  let id = window.localStorage.getItem("client-id");
  if (!id) {
    id = String(Math.round(1_000_000_000 * Math.random()));
    window.localStorage.setItem("client-id", id);
  }
  return id;
})();

// Note: `id` comes from the URL, courtesy of our router
const Swarm = ({ id }: Props) => {
  const stateMessage = useSignal("loading");
  const numState = useSignal(42);
  const swarmAddr = atob(id);
  useEffect(() => {
    const ws = new WebSocket(
      (window.location.protocol === "http:" ? "ws://" : "wss://") +
        swarmAddr
          // WSL reliability hack
          .replace("127.0.0.1", "localhost") +
        `?clientID=${clientID}`
    );
    ws.onclose = (e) => {
      console.log("ws closed", e);
      stateMessage.value = "closed";
    };
    ws.onerror = (e) => {
      console.log("ws error", e);
      stateMessage.value = "errored";
    };
    ws.onopen = (e) => {
      console.log("ws opened", e);
      stateMessage.value = "open";
      ws.send(
        JSON.stringify({
          State: numState.value,
        })
      );
    };
    ws.onmessage = (e) => {
      // console.log("ws msg", e);
      stateMessage.value = "msg-data: " + e.data;
    };
    numState.subscribe((s) => {
      console.log("new val", s);
      try {
        ws.send(
          JSON.stringify({
            State: s,
          })
        );
      } catch (ex: any) {
        console.error("failed to send", ex);
      }
    });
  }, []);
  return (
    <div>
      <h1>Swarm: {swarmAddr}</h1>
      <p>State: {stateMessage}</p>
      <div>
        <input
          class="m-4 rounded border-gray-700 border-2 w-12"
          type="number"
          value={numState}
          onChange={(e) => (numState.value = Number(e.currentTarget.value))}
        />
      </div>
    </div>
  );
};

export default Swarm;
