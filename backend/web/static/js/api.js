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

/**
 * Generic request handler
 * @param {string} endpoint
 * @param {RequestInit} options
 * @returns {Promise<any>}
 */
async function request(endpoint, options = {}) {
    const url = `${API_BASE}${endpoint}`;
    const headers = {
        ...options.headers,
    };

    const token = getAccessToken();
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    if (!(options.body instanceof FormData)) {
        headers['Content-Type'] = 'application/json';
    }

    const config = {
        ...options,
        headers,
    };

    const response = await fetch(url, config);

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: response.statusText }));
        throw new Error(errorData.message || 'An unknown error occurred');
    }

    if (response.status === 204) {
        return;
    }

    return response.json();
}

// --- Auth ---
/** @returns {Promise<User>} */
export const signup = (username, password) => request('/auth/signup', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
});

/** @returns {Promise<AuthTokens>} */
export const login = async (username, password) => {
    const data = await request('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ username, password }),
    });
    if (data.access_token && data.refresh_token) {
        setTokens(data.access_token, data.refresh_token);
    }
    return data;
};

export const logout = async () => {
    // In a real app, you'd also call a backend endpoint to invalidate the refresh token
    clearTokens();
};

// --- Users ---
/** @returns {Promise<User>} */
export const getMe = () => request('/users/me');

/** @returns {Promise<User>} */
export const updateProfile = (formData) => request('/users/me', {
    method: 'PUT',
    body: formData,
});

/** @returns {Promise<User>} */
export const getUserByUsername = (username) => request(`/users/${username}`);


// --- Friends ---
export const sendFriendRequest = (username) => request('/friends/requests', {
    method: 'POST',
    body: JSON.stringify({ username }),
});

export const getPendingFriendRequests = () => request('/friends/requests/pending');

export const acceptFriendRequest = (requestID) => request(`/friends/requests/${requestID}/accept`, {
    method: 'PUT',
});

export const rejectFriendRequest = (requestID) => request(`/friends/requests/${requestID}/reject`, {
    method: 'PUT',
});

export const unfriendUser = (userID) => request(`/friends/${userID}`, {
    method: 'DELETE',
});

export const listFriends = () => request('/friends');

