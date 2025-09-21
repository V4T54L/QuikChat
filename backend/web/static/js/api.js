/**
 * @typedef {object} User
 * @property {string} id
 * @property {string} username
 * @property {string} profile_pic_url
 * @property {string} created_at
 * @property {string} updated_at
 */

/**
 * @typedef {object} AuthTokens
 * @property {string} access_token
 * @property {string} refresh_token
 */

import { getAccessToken, setTokens, clearTokens } from './store.js';

const API_BASE = '/api/v1';

async function request(endpoint, options = {}) {
    const url = `${API_BASE}${endpoint}`;
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers,
    };

    const token = getAccessToken();
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    // For FormData, let the browser set the Content-Type
    if (options.body instanceof FormData) {
        delete headers['Content-Type'];
    }

    const config = {
        ...options,
        headers,
    };

    const response = await fetch(url, config);

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: response.statusText }));
        throw new Error(errorData.message || 'An API error occurred');
    }

    if (response.status === 204) {
        return null;
    }

    return response.json();
}

/** @returns {Promise<User>} */
export const signup = (username, password) => request('/auth/signup', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
});

/** @returns {Promise<AuthTokens>} */
export const login = async (username, password) => {
    const tokens = await request('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ username, password }),
    });
    setTokens(tokens.access_token, tokens.refresh_token);
    return tokens;
};

export const logout = async () => {
    // Implementation will require refresh token from store
    // For now, just clear local tokens
    clearTokens();
};

/** @returns {Promise<User>} */
export const getMe = () => request('/users/me');

/** @returns {Promise<User>} */
export const updateProfile = (formData) => request('/users/me', {
    method: 'PUT',
    body: formData,
});

/** @returns {Promise<User>} */
export const getUserByUsername = (username) => request(`/users/${username}`);

