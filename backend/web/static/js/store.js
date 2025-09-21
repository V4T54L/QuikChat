/**
 * @typedef {import('./api.js').User} User
 */

let state = {
    /** @type {User | null} */
    currentUser: null,
    accessToken: null,
    refreshToken: null,
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
    state.accessToken = null;
    state.refreshToken = null;
    state.currentUser = null;
}

