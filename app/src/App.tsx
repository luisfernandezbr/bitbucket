import React from 'react';
import { SimulatorInstaller, Integration } from '@pinpt/agent.websdk';
import IntegrationUI from './integration';

function App() {
	// check to see if we are running local and need to run in simulation mode
	if (window === window.parent && window.location.href.indexOf('localhost') > 0) {
		const integration: Integration = {
			name: 'Pinpoint',
			description: 'This is the BitBucket integration for Pinpoint',
			tags: [
					'Source Code',
			],
			installed: false,
			refType: 'bitbucket',
			icon: '',
			publisher: {
				name: 'Pinpoint',
				avatar: '',
				url: 'https://pinpoint.com'
			},
			uiURL: window.location.href
		};
		return <SimulatorInstaller integration={integration} />;
	}
	return (
		<div className="App">
			<IntegrationUI />
		</div>
	);
}

export default App;
