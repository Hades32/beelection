import { h } from 'preact';
import { Route, Router } from 'preact-router';

import Header from './header';

// Code-splitting is automated for `routes` directory
import Home from '../routes/home';
import Profile from '../routes/profile';
import Swarm from '../routes/swarm';

const App = () => (
	<div id="app">
		<Header />
		<main class="mx-2 mt-2">
			<Router>
				<Route path="/" component={Home} />
				<Route path="/profile/" component={Profile} user="me" />
				<Route path="/swarm/:id" component={Swarm} />
			</Router>
		</main>
	</div>
);

export default App;
