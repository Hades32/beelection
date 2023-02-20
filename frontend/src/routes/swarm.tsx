import { h } from "preact";

interface Props {
  id: string;
}

// Note: `id` comes from the URL, courtesy of our router
const Swarm = ({ id }: Props) => {
  return (
    <div>
      <h1>Swarm: {atob(id)}</h1>
      <p>TODO.</p>
    </div>
  );
};

export default Swarm;
