import * as store from './store.js'; // Changed import style
import * as ui from './ui.js';       // Changed import style
import * as main from './main.js';    // Changed import style

let socket;
let reconnectInterval = 1000; // Start with 1 second

/**
 * Handles incoming WebSocket events and dispatches actions.
 * @param {MessageEvent} event - The raw WebSocket message event.
 */
function handleIncomingEvent(event) {
    const data = JSON.parse(event.data);
    console.log('WS Event Received:', data);

    switch (data.type) {
        case 'new_message':
            main.handleNewMessage(data.payload); // Handled new message
            break;
        case 'friend_request_received':
            ui.showNotification(`New friend request from ${data.payload.sender.username}`, 'info'); // Updated payload access
            main.refreshPendingRequests();
            break;
        case 'friend_request_accepted':
            ui.showNotification(`${data.payload.user.username} accepted your friend request!`, 'success'); // Updated payload access
            main.refreshFriendsList();
            main.refreshPendingRequests();
            break;
        case 'friend_request_rejected':
            ui.showNotification(`${data.payload.user.username} rejected your friend request.`, 'info'); // Updated payload access
            main.refreshPendingRequests();
            break;
        case 'unfriended':
            ui.showNotification(`${data.payload.user.username} unfriended you.`, 'info'); // Updated payload access
            main.refreshFriendsList();
            break;
        case 'group_joined':
            ui.showNotification(`You joined the group: ${data.payload.group.name}`, 'success');
            main.refreshGroupsList();
            break;
        case 'group_member_added':
            ui.showNotification(`${data.payload.adder.username} added you to ${data.payload.group.name}`, 'info');
            main.refreshGroupsList();
            break;
        // Removed other group events as they are not in attempted content
        default:
            console.warn('Unhandled event type:', data.type); // Added default warning
    }
}

export function connect() {
    const token = store.getAccessToken(); // Used store.getAccessToken()
    if (!token) {
        console.log("No access token, WebSocket connection aborted.");
        return;
    }

    // Prevent multiple connections
    if (socket && socket.readyState === WebSocket.OPEN) {
        console.log("WebSocket is already connected.");
        return;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    socket = new WebSocket(`${protocol}//${host}/ws`); // Updated WebSocket URL

    socket.onopen = () => {
        console.log('WebSocket connected');
        reconnectInterval = 1000; // Reset reconnect interval on successful connection
        // Send auth event
        send('auth', { token: token }); // Explicitly send token in payload
        ui.showNotification('Connected to real-time service.', 'success'); // Kept notification
    };

    socket.onmessage = handleIncomingEvent;

    socket.onclose = () => {
        console.log('WebSocket disconnected. Attempting to reconnect...');
        ui.showNotification('Real-time connection lost. Attempting to reconnect...', 'error'); // Kept notification
        setTimeout(connect, reconnectInterval);
        reconnectInterval = Math.min(reconnectInterval * 2, 30000); // Exponential backoff up to 30s
    };

    socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        ui.showNotification('A real-time connection error occurred.', 'error'); // Kept notification
        socket.close(); // This will trigger the onclose handler for reconnection
    };
}

/**
 * Sends a message over the WebSocket connection.
 * @param {string} type - The type of the message.
 * @param {object} payload - The payload of the message.
 */
export function send(type, payload) {
    if (socket && socket.readyState === WebSocket.OPEN) {
        const message = JSON.stringify({ type, payload });
        socket.send(message);
    } else {
        console.error('WebSocket is not connected.');
        ui.showNotification('Not connected to server.', 'error'); // Kept notification
    }
}
