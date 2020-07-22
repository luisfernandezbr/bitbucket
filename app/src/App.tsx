import React from 'react';
import { SimulatorInstaller, Integration, IProcessingDetail, IProcessingState, IInstalledLocation } from '@pinpt/agent.websdk';
import IntegrationUI from './integration';

function App() {
	// check to see if we are running local and need to run in simulation mode
	if (window === window.parent && window.location.href.indexOf('localhost') > 0) {
		const integration: Integration = {
			name: 'BitBucket',
			description: 'The official BitBucket integration for Pinpoint',
			tags: ['Source Code Management', 'Issue Management'],
			installed: false,
			refType: 'bitbucket',
			icon: `data:image/svg+xml,%3Csvg aria-hidden='true' focusable='false' data-prefix='fab' data-icon='bitbucket' class='svg-inline--fa fa-bitbucket fa-w-16' role='img' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 512 512'%3E%3Cpath fill='currentColor' d='M22.2 32A16 16 0 0 0 6 47.8a26.35 26.35 0 0 0 .2 2.8l67.9 412.1a21.77 21.77 0 0 0 21.3 18.2h325.7a16 16 0 0 0 16-13.4L505 50.7a16 16 0 0 0-13.2-18.3 24.58 24.58 0 0 0-2.8-.2L22.2 32zm285.9 297.8h-104l-28.1-147h157.3l-25.2 147z'%3E%3C/path%3E%3C/svg%3E`,
			publisher: {
				name: 'Pinpoint',
				avatar: 'https://avatars0.githubusercontent.com/u/24400526?s=200&v=4',
				url: 'https://pinpoint.com'
			},
			uiURL: window.location.href
		};

		const processingDetail: IProcessingDetail = {
			createdDate: Date.now() - (86400000 * 5) - 60000,
			processed: true,
			lastProcessedDate: Date.now() - (86400000 * 2),
			lastExportRequestedDate: Date.now() - ((86400000 * 5) + 60000),
			lastExportCompletedDate: Date.now() - (86400000 * 5),
			state: IProcessingState.IDLE,
			throttled: false,
			throttledUntilDate: Date.now() + 2520000,
			paused: false,
			location: IInstalledLocation.CLOUD
		};

		return <SimulatorInstaller integration={integration} processingDetail={processingDetail} />;
	}
	return <IntegrationUI />;
}

export default App;
