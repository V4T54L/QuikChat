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

/**
 * @typedef {object} Group
 * @property {string} id
 * @property {string} handle
 * @property {string} name
 * @property {string} photo_url
 * @property {string} owner_id
 * @property {string} created_at
 * @property {string} updated_at
 */

import { getAccessToken, clearTokens } from './store.js';

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
    if (token) { // Only add Authorization header if token exists
        headers['Authorization'] = `Bearer ${token}`;
    }

    if (!(options.body instanceof FormData)) {
        headers['Content-Type'] = 'application/json';
    }

    const config = {
        ...options,
        headers,
    };

    try {
        const response = await fetch(url, config);

        if (response.status === 401) {
            clearTokens();
            window.location.href = '/';
            throw new Error('Unauthorized');
        }

        if (!response.ok) {
            const errorText = await response.text();
            let errorMessage = errorText || `HTTP error! status: ${response.status}`;
            try {
                const errorData = JSON.parse(errorText);
                errorMessage = errorData.message || errorMessage;
            } catch (e) {
                // Not a JSON error, use plain text
            }
            throw new Error(errorMessage);
        }

        if (response.status === 204) {
            return null; // Return null for no content
        }

        return response.json();
    } catch (error) {
        console.error('API request failed:', error);
        throw error;
    }
}

// --- Auth ---
/** @returns {Promise<User>} */
export const signup = (username, password) => request('/auth/signup', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
});

/** @returns {Promise<AuthTokens>} */
export const login = (username, password) => request('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
});

export const logout = () => {
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

// --- Groups ---
/** @returns {Promise<Group>} */
export const createGroup = (formData) => request('/groups', {
    method: 'POST',
    body: formData,
});

/** @returns {Promise<Group[]>} */
export const searchGroups = (query) => request(`/groups/search?q=${encodeURIComponent(query)}`);

/** @returns {Promise<void>} */
export const joinGroup = (handle) => request(`/groups/${handle}/join`, { method: 'POST' });

/** @returns {Promise<Group>} */
export const getGroupDetails = (groupId) => request(`/groups/${groupId}`);

/** @returns {Promise<Group>} */
export const updateGroup = (groupId, formData) => request(`/groups/${groupId}`, {
    method: 'PUT',
    body: formData,
});

/** @returns {Promise<void>} */
export const addGroupMember = (groupId, friendId) => request(`/groups/${groupId}/members`, {
    method: 'POST',
    body: JSON.stringify({ friend_id: friendId }),
});

/** @returns {Promise<void>} */
export const removeGroupMember = (groupId, memberId) => request(`/groups/${groupId}/members/${memberId}`, { method: 'DELETE' });

/** @returns {Promise<void>} */
export const leaveGroup = (groupId) => request(`/groups/${groupId}/leave`, { method: 'POST' });

/** @returns {Promise<void>} */
export const transferGroupOwnership = (groupId, newOwnerId) => request(`/groups/${groupId}/transfer-ownership`, {
    method: 'PUT',
    body: JSON.stringify({ new_owner_id: newOwnerId }),
});

/** @returns {Promise<Group[]>} */
export const listMyGroups = () => request('/groups/me');

