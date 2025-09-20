/** @type {import('./types.js').AppState} */
let state = {
    currentUser: null,
    accessToken: null,
    refreshToken: null,
    friends: [],
    groups: [],
    activeChat: null,
};

/** @returns {import('./types.js').AppState} */
export function getState() {
    return { ...state };
}

/** @param {Partial<import('./types.js').AppState>} newState */
export function setState(newState) {
    state = { ...state, ...newState };
}

