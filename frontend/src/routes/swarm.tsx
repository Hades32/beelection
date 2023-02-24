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
          State: 42,
        })
      );
    };
    ws.onmessage = (e) => {
      console.log("ws msg", e);
      stateMessage.value = "msg-data: " + e.data;
    };
  }, []);
  return (
    <div>
      <h1>Swarm: {swarmAddr}</h1>
      <p>State: {stateMessage}</p>
    </div>
  );
};

export default Swarm;
