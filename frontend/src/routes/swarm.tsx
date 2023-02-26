import { useSignal } from "@preact/signals";
import { h } from "preact";
import { useEffect, useRef } from "preact/hooks";
import * as Phaser from "phaser";
import { rateLimit } from "../utils/rate-limit";

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
  const phaserParent = useRef<HTMLDivElement>();
  const websocket = useRef<WebSocket>();

  const send = rateLimit((o: object) => {
    if (!websocket.current) {
      return;
    }
    websocket.current.send(JSON.stringify(o));
  }, 250);

  useEffect(() => {
    let wsOpen = false;
    const ws = new WebSocket(
      (window.location.protocol === "http:" ? "ws://" : "wss://") +
        swarmAddr
          // WSL reliability hack
          .replace("127.0.0.1", "localhost") +
        `?clientID=${clientID}`
    );
    websocket.current = ws;
    ws.onclose = (e) => {
      wsOpen = false;
      console.log("ws closed", e);
      stateMessage.value = "closed";
    };
    ws.onerror = (e) => {
      console.log("ws error", e);
      stateMessage.value = "errored";
    };
    ws.onopen = (e) => {
      wsOpen = true;
      console.log("ws opened", e);
      stateMessage.value = "open";
      send({
        State: 0,
      });
    };
    ws.onmessage = (e) => {
      // console.log("ws msg", e);
      stateMessage.value = "msg-data: " + e.data;
    };
  }, []);

  useEffect(() => {
    let graphics: Phaser.GameObjects.Graphics;
    let deathZone: Phaser.Geom.Circle;

    const game = new Phaser.Game({
      type: Phaser.WEBGL,
      width: 800,
      height: 600,
      backgroundColor: "#000",
      parent: phaserParent.current,
      scene: {
        preload: function () {
          // this.load.atlas(
          //   "flares",
          //   "assets/particles/flares.png",
          //   "assets/particles/flares.json"
          // );
        },
        create: function () {
          // let emitZone = new Phaser.Geom.Rectangle(0, 0, 800, 600);

          //  Any particles that enter this shape will be killed instantly
          deathZone = new Phaser.Geom.Circle(0, 0, 48);

          // let particles = this.add.particles("flares");

          // let emitter = particles.createEmitter({
          //   frame: ["red", "green", "blue"],
          //   speed: { min: -20, max: 20 },
          //   lifespan: 10000,
          //   quantity: 2,
          //   scale: { min: 0.1, max: 0.4 },
          //   alpha: { start: 1, end: 0 },
          //   blendMode: "ADD",
          //   emitZone: { source: emitZone },
          //   deathZone: { type: "onEnter", source: deathZone },
          // });

          graphics = this.add.graphics();

          this.input.on("pointermove", function (pointer) {
            deathZone.x = pointer.x;
            deathZone.y = pointer.y;
            send({
              State: deathZone.x * deathZone.y,
            });
          });
        },
        update: function () {
          graphics.clear();

          graphics.lineStyle(1, 0x00ff00, 1);

          graphics.strokeCircleShape(deathZone);
        },
      },
    });

    return () => {
      game.destroy(false);
    };
  }, []);

  return (
    <div>
      <h1>Swarm: {swarmAddr}</h1>
      <p>State: {stateMessage}</p>
      <div ref={phaserParent} />
    </div>
  );
};

export default Swarm;
