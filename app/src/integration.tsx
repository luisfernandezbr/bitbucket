import React, { useEffect, useState, useRef, useCallback } from 'react';
import Icon from '@pinpt/uic.next/Icon';
import Loader from '@pinpt/uic.next/Loader';
import ErrorPage from '@pinpt/uic.next/Error';
import { faCloud, faServer, faExclamationCircle } from '@fortawesome/free-solid-svg-icons';
import {
	useIntegration,
	Account,
	AccountsTable,
	IntegrationType,
	OAuthConnect,
	IAuth,
	Form,
	FormType,
	ConfigAccount,
} from '@pinpt/agent.websdk';

import styles from './styles.module.less';

interface validateResponse {
	accounts: ConfigAccount[];
}

const toAccount = (data: ConfigAccount): Account => {
	return {
		id: data.id,
		public: data.public,
		type: data.type,
		avatarUrl: data.avatarUrl,
		name: data.name || '',
		description: data.description || '',
		totalCount: data.totalCount || 0,
		selected: !!data.selected
	}
};

const AccountList = ({ setError }: { setError: (error: Error | undefined) => void }) => {
	const { config, setConfig, installed, setInstallEnabled, setValidate } = useIntegration();
	const [accounts, setAccounts] = useState<Account[]>([]);
	const [fetching, setFetching] = useState(false);
	const unmounted = useRef(false);

	useEffect(() => {
		if (fetching || accounts.length) {
			return
		}
		setFetching(true);
		const fetch = async () => {
			try {
				config.accounts = config.accounts || {}
				const res: validateResponse = await setValidate(config);
				for (let i = 0; i < res.accounts.length; i++) {
					const obj = toAccount(res.accounts[i]);
					if (installed) {
						const selected = config.accounts[obj.id]?.selected
						obj.selected = !!selected
					}
					accounts.push(obj);
					config.accounts[obj.id] = obj;
				}
				setConfig(config);
				setAccounts(accounts)
				if (!installed && accounts.length > 0) {
					setInstallEnabled(true);
				}
			} catch (err) {
				if (unmounted.current) {
					return;
				}
				setError(err);
			} finally {
				if (unmounted.current) {
					return;
				}
				setFetching(false);
			}
		}
		fetch();
		return () => {
			unmounted.current = true;
		}
	}, [config]);

	if (fetching) {
		return <Loader centered style={{ height: '30rem' }} />;
	}
	return (
		<AccountsTable
			description='For the selected accounts, all repositories, pull requests and other data will automatically be made available in Pinpoint once installed.'
			accounts={accounts}
			entity='repo'
			config={config}
		/>
	);
};

const LocationSelector = ({ setType }: { setType: (val: IntegrationType) => void }) => {
	return (
		<div className={styles.Location}>
			<div className={styles.Button} onClick={() => setType(IntegrationType.CLOUD)}>
				<Icon icon={faCloud} className={styles.Icon} />
				I'm using the <strong>bitbucket.com</strong> cloud service to manage my data
			</div>

			<div className={styles.Button} onClick={() => setType(IntegrationType.SELFMANAGED)}>
				<Icon icon={faServer} className={styles.Icon} />
				I'm using <strong>my own systems</strong> or a <strong>third-party</strong> to manage a BitBucket service
			</div>
		</div>
	);
};

const SelfManagedForm = ({ callback }: { callback: (auth: IAuth) => Promise<void> }) => {
	return <Form type={FormType.BASIC} name='bitbucket' callback={callback} />;
};

const isAuthError = (error: Error | undefined): boolean => {
	return error ? error.message.indexOf('401') > 0 : false;
}

const Integration = () => {
	const { loading, currentURL, config, isFromRedirect, isFromReAuth, setConfig } = useIntegration();
	const [type, setType] = useState<IntegrationType | undefined>(config.integration_type);
	const [error, setError] = useState<Error | undefined>();
	const [, setRerender] = useState(0);

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
						date_ts: Date.now(),
						url: 'https://api.bitbucket.org',
						access_token: profile.Integration.auth.accessToken,
						refresh_token: profile.Integration.auth.refreshToken,
						scopes: profile.Integration.auth.scopes,
					};
					setConfig(config);
					setRerender(Date.now());
				}
			});
		}
	}, [config, loading, isFromRedirect, currentURL]);

	useEffect(() => {
		if (type) {
			config.integration_type = type;
			setConfig(config);
			setRerender(Date.now());
		}
	}, [config, type])

	const basicAuthSet = useCallback(async (auth: IAuth) => {
		if (isAuthError(error)) {
			setError(undefined);
		} else {
			setRerender(Date.now());
		}
	}, [setError, setRerender]);

	if (loading) {
		return <Loader centered />;
	}

	let content;

	if (isFromReAuth || isAuthError(error)) {
		if (config.integration_type === IntegrationType.CLOUD) {
			content = <OAuthConnect name='BitBucket' reauth />;
		} else {
			content = <SelfManagedForm callback={basicAuthSet} />;
		}
	} else {
		if (!config.integration_type) {
			content = <LocationSelector setType={setType} />;
		} else if (config.integration_type === IntegrationType.CLOUD && !config.oauth2_auth) {
			content = <OAuthConnect name='BitBucket' />;
		} else if (config.integration_type === IntegrationType.SELFMANAGED && !config.basic_auth && !config.apikey_auth) {
			content = <SelfManagedForm callback={basicAuthSet} />;
		} else {
			content = <AccountList setError={setError} />;
		}
	}

	return (
		<div className={styles.Wrapper}>
			{error && (
				<div className={styles.Error}>
					<Icon icon={faExclamationCircle} style={{marginRight: '0.5rem'}} /> {isAuthError(error) ? 'Invalid Authorization Credentials' : error.message} 
				</div>
			)}
			{content}
		</div>
	);
};


export default Integration;