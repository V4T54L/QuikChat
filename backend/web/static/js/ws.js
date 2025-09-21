import { getAccessToken } from './store.js';
import { showNotification } from './ui.js';
import * as api from './api.js';
import * * as store from './store.js';
import * as ui from './ui.js';

let socket = null;

export function connect() {
    const token = getAccessToken();
    if (!token) {
        console.error('No access token found for WebSocket connection.');
        return;
    }

    if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
        console.log('WebSocket is already connected or connecting.');
        return;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/ws`;

    socket = new WebSocket(wsUrl);

    socket.onopen = () => {
        console.log('WebSocket connection established.');
        // Send auth token upon connection
        socket.send(JSON.stringify({ type: 'auth', payload: { token } }));
        showNotification('Connected to real-time service.', 'success');
    };

    socket.onmessage = (event) => {
        try {
            const parsedEvent = JSON.parse(event.data);
            handleIncomingEvent(parsedEvent);
        } catch (error) {
            console.error('Error parsing incoming WebSocket message:', error);
        }
    };

    socket.onclose = (event) => {
        console.log('WebSocket connection closed:', event.reason);
        showNotification('Real-time connection lost. Attempting to reconnect...', 'error');
        // Simple reconnect logic
        setTimeout(connect, 5000);
    };

    socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        showNotification('A real-time connection error occurred.', 'error');
    };
}

export function send(messageType, payload) {
    if (socket && socket.readyState === WebSocket.OPEN) {
        const message = JSON.stringify({ type: messageType, payload });
        socket.send(message);
    } else {
        console.error('WebSocket is not connected.');
    }
}

async function handleIncomingEvent(event) {
    console.log('Received event:', event);

    switch (event.Type) {
        case 'new_message':
            // To be implemented
            showNotification(`New message from ${event.Payload.sender_id}`);
            break;
        case 'friend_request_received':
            showNotification(`New friend request from ${event.Payload.sender.username}!`, 'info');
            // Refresh pending requests
            const requests = await api.getPendingFriendRequests();
            store.setPendingRequests(requests);
            ui.renderPendingRequests(requests, store.getCurrentUser().id);
            break;
        case 'friend_request_accepted':
            showNotification(`${event.Payload.receiver.username} accepted your friend request!`, 'success');
            // Refresh both lists
            const [friends, pending] = await Promise.all([api.listFriends(), api.getPendingFriendRequests()]);
            store.setFriends(friends);
            store.setPendingRequests(pending);
            ui.renderFriendList(friends);
            ui.renderPendingRequests(pending, store.getCurrentUser().id);
            break;
        case 'friend_request_rejected':
            showNotification(`${event.Payload.user.username} rejected your friend request.`, 'info');
            // Refresh pending requests
            const pendingReqs = await api.getPendingFriendRequests();
            store.setPendingRequests(pendingReqs);
            ui.renderPendingRequests(pendingReqs, store.getCurrentUser().id);
            break;
        case 'unfriended':
            showNotification(`You were unfriended by ${event.Payload.user.username}.`, 'info');
            // Refresh friends list
            const friendList = await api.listFriends();
            store.setFriends(friendList);
            ui.renderFriendList(friendList);
            break;
        default:
            console.warn('Unhandled event type:', event.Type);
    }
}

