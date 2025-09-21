/**
 * @typedef {import('./api.js').User} User
 */

/**
 * @typedef {object} FriendRequest
 * @property {string} id
 * @property {string} sender_id
 * @property {string} receiver_id
 * @property {User} sender
 * @property {User} receiver
 */

const state = {
    /** @type {User | null} */
    currentUser: null,
    accessToken: null,
    refreshToken: null,
    /** @type {User[]} */
    friends: [],
    /** @type {FriendRequest[]} */
    pendingRequests: [],
};

/** @param {User} user */
export function setCurrentUser(user) {
    state.currentUser = user;
}

export function getCurrentUser() {
    return state.currentUser;
}

export function setTokens(access, refresh) {
    state.accessToken = access;
    state.refreshToken = refresh;
}

export function getAccessToken() {
    return state.accessToken;
}

export function getRefreshToken() {
    return state.refreshToken;
}

export function clearTokens() {
    state.currentUser = null;
    state.accessToken = null;
    state.refreshToken = null;
    state.friends = []; // Clear friends on logout
    state.pendingRequests = []; // Clear pending requests on logout
}

export function setFriends(friends) {
    state.friends = friends;
}

export function getFriends() {
    return state.friends;
}

export function setPendingRequests(requests) {
    state.pendingRequests = requests;
}

export function getPendingRequests() {
    return state.pendingRequests;
}

