import { getAccessToken } from './store.js';
import { showNotification } from './ui.js';

let socket = null;

/**
 * Establishes a WebSocket connection.
 */
function connect() {
    const token = getAccessToken();
    if (!token) {
        console.error('No access token found for WebSocket connection.');
        return;
    }

    if (socket && socket.readyState === WebSocket.OPEN) {
        console.log('WebSocket is already connected.');
        return;
    }

    const protocol = window.location.protocol === 'https' ? 'wss' : 'ws';
    const host = window.location.host;
    const wsUrl = `${protocol}://${host}/ws`;

    socket = new WebSocket(wsUrl);

    socket.onopen = () => {
        console.log('WebSocket connection established.');
        // Send authentication token
        // The backend will get the user ID from the JWT used to establish the connection
        // so we don't need to send it explicitly.
        showNotification('Connected to real-time service.', 'success');
    };

    socket.onmessage = (event) => {
        console.log('WebSocket message received:', event.data);
        try {
            const parsedEvent = JSON.parse(event.data);
            handleIncomingEvent(parsedEvent);
        } catch (error) {
            console.error('Error parsing incoming WebSocket event:', error);
        }
    };

    socket.onclose = (event) => {
        console.log('WebSocket connection closed:', event);
        showNotification('Disconnected from real-time service.', 'error');
        // Optional: implement reconnection logic here
        socket = null;
    };

    socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        showNotification('A real-time connection error occurred.', 'error');
    };
}

/**
 * Sends a message over the WebSocket.
 * @param {string} messageType
 * @param {object} payload
 */
function send(messageType, payload) {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
        console.error('WebSocket is not connected.');
        return;
    }
    const message = JSON.stringify({ type: messageType, payload });
    socket.send(message);
    console.log('WebSocket message sent:', message);
}

/**
 * Handles incoming events from the WebSocket.
 * @param {object} event
 */
function handleIncomingEvent(event) {
    // Example event handling
    switch (event.type) {
        case 'new_message':
            // ui.renderMessage(event.payload);
            showNotification(`New message received!`);
            break;
        case 'friend_request_received':
            // store.addFriendRequest(event.payload);
            // ui.renderFriendRequests();
            showNotification(`New friend request from ${event.payload.sender.username}`);
            break;
        default:
            console.log(`Unhandled event type: ${event.type}`);
            showNotification(`Received a new notification: ${event.type}`);
    }
}

export { connect, send };
