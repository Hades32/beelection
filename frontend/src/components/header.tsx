import { h } from "preact";
import { Link } from "preact-router/match";

const Header = () => (
  <header class="flex flex-row place-content-between text-yellow-500 bg-yellow-900 shadow-md shadow-yellow-700 text-lg">
    <a href="/" class="flex flex-row text-xl mx-2 my-4">
      <img
        class="mx-2"
        src="../../assets/preact-logo-inverse.svg"
        alt="Logo"
        height="32"
        width="32"
      />
      <h1>Beelection</h1>
    </a>
    <nav class="mx-2 my-4 ">
      <NavLink href="/" title="Home" />
      <NavLink href="/profile" title="Me" />
      <NavLink href="/profile/john" title="John" />
    </nav>
  </header>
);

const NavLink = (props: { href: string; title: string }) => (
  <Link
    class="p-2 m-1 rounded hover:bg-yellow-700"
    activeClassName="bg-yellow-800"
    href={props.href}
  >
    {props.title}
  </Link>
);

export default Header;
