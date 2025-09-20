import { getState, setState } from './state.js';

const BASE_URL = '/api/v1';

async function request(endpoint, options = {}) {
    const { accessToken } = getState();
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers,
    };

    if (accessToken) {
        headers['Authorization'] = `Bearer ${accessToken}`;
    }

    const config = {
        ...options,
        headers,
    };

    const response = await fetch(`${BASE_URL}${endpoint}`, config);

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: 'An unknown error occurred' }));
        throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
    }

    if (response.status === 204) { // No Content
        return null;
    }

    return response.json();
}

export const api = {
    login: (username, password) => request('/login', {
        method: 'POST',
        body: JSON.stringify({ username, password }),
    }),
    register: (username, password) => request('/register', {
        method: 'POST',
        body: JSON.stringify({ username, password }),
    }),
    logout: (refreshToken) => request('/logout', {
        method: 'POST',
        body: JSON.stringify({ refreshToken }),
    }),
    getFriends: () => request('/friends'),
    getGroups: () => {
        // This endpoint doesn't exist, so we'll mock it for now.
        // In a real app, you'd fetch groups the user is a member of.
        console.warn("API call to getGroups is mocked.");
        return Promise.resolve([]); 
    },
    getMessages: (chatId) => {
        // This endpoint doesn't exist. Messages are received via WebSocket.
        // A real app would have an endpoint to fetch message history.
        console.warn("API call to getMessages is mocked.");
        return Promise.resolve([]);
    },
};

