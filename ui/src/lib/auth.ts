const TOKEN_KEY = 'intraclub_jwt';

export function getToken(): string | null {
	if (typeof localStorage === 'undefined') return null;
	return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
	localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken(): void {
	localStorage.removeItem(TOKEN_KEY);
}

export function isLoggedIn(): boolean {
	return getToken() !== null;
}

export function authFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
	const token = getToken();
	const headers = new Headers(init?.headers);
	if (token) {
		headers.set('X-INTRACLUB-TOKEN', token);
	}
	return fetch(input, { ...init, headers });
}
