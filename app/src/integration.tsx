import React, { useEffect, useState, useRef } from 'react';
import { Icon, Loader, Error } from '@pinpt/uic.next';
import {
	useIntegration,
	Account,
	AccountsTable,
	IntegrationType,
	OAuthConnect,
	Graphql,
	IAuth,
	IAppBasicAuth,
	Form,
	FormType,
	Http,
	IOAuth2Auth,
	Config,
} from '@pinpt/agent.websdk';

import styles from './styles.module.less';

interface workspacesResponse {
	is_private: Boolean,
	name: string,
	slug: string,
	type: string,
	uuid: string,
}

function workspacesToAccount(workspaces: workspacesResponse[], config: Config): Account[] {
	let accounts: Account[] = [];
	config.accounts = {}
	for (var i = 0; i < workspaces.length; i++) {
		var workspace = workspaces[i];
		var obj: Account = {
			avatarUrl: "",
			totalCount: 0,
			id: workspace.uuid,
			name: workspace.name,
			description: workspace.slug,
			type: 'ORG',
			public: !workspace.is_private
		};
		accounts.push(obj);
		config.accounts[obj.id] = obj;
	}
	return accounts;
}

const AccountList = ({ workspaces }: { workspaces: workspacesResponse[] | undefined }) => {
	const { config, setConfig } = useIntegration();
	const [accounts, setAccounts] = useState<Account[]>([]);
	const [fetching, setFetching] = useState(false);

	if (accounts.length === 0 && !fetching) {
		const fetch = async () => {
			if (config.basic_auth) {
				let res = await fetchWorkspaces(config.basic_auth as IAppBasicAuth)
				setAccounts(workspacesToAccount(res[1], config))
			} else {
				let res = await fetchWorkspaces(config.oauth2_auth as IOAuth2Auth)
				setAccounts(workspacesToAccount(res[1], config))
			}
			setConfig(config);
			setFetching(false);
		}
		setFetching(true);
		fetch();
	}

	return (
		<AccountsTable
			description="For the selected accounts, all repositories, pull requests and other data will automatically be made available in Pinpoint once installed."
			accounts={accounts}
			entity="repo"
			config={config}
		/>
	);
};

async function fetchWorkspaces(auth: IAppBasicAuth | IOAuth2Auth): Promise<[number, workspacesResponse[]]> {
	return new Promise(async (resolve, reject) => {
		var url = auth.url + "/2.0/workspaces"
		var header: string;
		if ('username' in auth) {
			var basic = (auth as IAppBasicAuth);
			header = "Basic " + btoa(basic.username + ":" + basic.password);
		} else {
			var oauth = (auth as IOAuth2Auth);
			header = "Bearer " + oauth.access_token;
		}
		var res = await Http.get(url, { "Authorization": header });
		if (res[1] === 200) {
			resolve([res[1], res[0].values]);
			return;
		}
		reject("resoponse " + res[1]);
	})
}


const LocationSelector = ({ setType }: { setType: (val: IntegrationType) => void }) => {
	return (
		<div className={styles.Location}>
			<div className={styles.Button} onClick={() => setType(IntegrationType.CLOUD)}>
				<Icon icon={['fas', 'cloud']} className={styles.Icon} />
				I'm using the <strong>bitbucket.com</strong> cloud service to manage my data
			</div>

			<div className={styles.Button} onClick={() => setType(IntegrationType.SELFMANAGED)}>
				<Icon icon={['fas', 'server']} className={styles.Icon} />
				I'm using <strong>my own systems</strong> or a <strong>third-party</strong> to manage a BitBucket service
			</div>
		</div>
	);
};

const SelfManagedForm = ({ setWorkspaces }: { setWorkspaces: (val: workspacesResponse[]) => void }) => {
	async function verify(auth: IAuth): Promise<boolean> {
		return new Promise<boolean>(async resolve => {
			try {
				var res = await fetchWorkspaces(auth as IAppBasicAuth)
				if (res[0] === 200) {
					setWorkspaces(res[1]);
					resolve(true);
					return
				}
				resolve(false);
			} catch (ex) {
				resolve(false);
			}
		});
	}
	return <Form type={FormType.BASIC} name="bitbucket" callback={verify} />;
};

const Integration = () => {
	const { loading, currentURL, config, isFromRedirect, isFromReAuth, setConfig, authorization } = useIntegration();
	const [type, setType] = useState<IntegrationType | undefined>(config.integration_type);
	const [, setRerender] = useState(0);
	const [workspaces, setWorkspaces] = useState<workspacesResponse[]>()


	// ============= OAuth 2.0 =============
	useEffect(() => {
		if (!loading && isFromRedirect && currentURL) {
			const search = currentURL.split('?');
			const tok = search[1].split('&');
			tok.forEach(async token => {
				const t = token.split('=');
				const k = t[0];
				const v = t[1];
				if (k === 'profile') {
					const profile = JSON.parse(atob(decodeURIComponent(v)));
					config.oauth2_auth = {
						url: "https://api.bitbucket.org",
						access_token: profile.Integration.auth.accessToken,
						refresh_token: profile.Integration.auth.refreshToken,
						scopes: profile.Integration.auth.scopes,
					};
					setConfig(config);
					setRerender(Date.now());
				}
			});
		}


		if (type) {
			config.integration_type = type;
			setConfig(config);
			setRerender(Date.now());
		}

	}, [loading, isFromRedirect, currentURL, config, setRerender, setConfig, type]);

	if (loading) {
		return <Loader screen />;
	}

	let content;

	if (isFromReAuth) {
		if (config.integration_type === IntegrationType.CLOUD) {
			content = <OAuthConnect name="GitHub" reauth />
		} else {
			content = <SelfManagedForm setWorkspaces={setWorkspaces} />;
		}
	} else {
		if (!config.integration_type) {
			content = <LocationSelector setType={setType} />;
		} else if (config.integration_type === IntegrationType.CLOUD && !config.oauth2_auth) {
			content = <OAuthConnect name="GitHub" />;
		} else if (config.integration_type === IntegrationType.SELFMANAGED && !config.basic_auth && !config.apikey_auth) {
			content = <SelfManagedForm setWorkspaces={setWorkspaces} />;
		} else {
			content = <AccountList workspaces={workspaces} />
		}
	}

	return (
		<div className={styles.Wrapper}>
			{content}
		</div>
	)
};


export default Integration;