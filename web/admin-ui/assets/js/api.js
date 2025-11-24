// API Configuration
const API_BASE_URL = 'http://localhost:8080/api';

// API Client
const API = {
    // Helper to get auth header
    getHeaders() {
        const token = localStorage.getItem('access_token');
        const headers = {
            'Content-Type': 'application/json'
        };
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }
        return headers;
    },

    // Generic request handler
    async request(endpoint, options = {}) {
        const url = `${API_BASE_URL}${endpoint}`;
        const config = {
            ...options,
            headers: {
                ...this.getHeaders(),
                ...options.headers
            }
        };

        try {
            const response = await fetch(url, config);
            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || `HTTP ${response.status}`);
            }

            return data;
        } catch (error) {
            console.error('API Error:', error);
            throw error;
        }
    },

    // Auth endpoints
    auth: {
        register: (userData) => API.request('/auth/register', {
            method: 'POST',
            body: JSON.stringify(userData)
        }),

        login: (credentials) => API.request('/auth/login', {
            method: 'POST',
            body: JSON.stringify(credentials)
        }),

        logout: () => API.request('/auth/logout', { method: 'POST' }),

        validate: (token) => API.request('/auth/validate', {
            method: 'POST',
            body: JSON.stringify({ token })
        }),

        refresh: (refreshToken) => API.request('/auth/refresh', {
            method: 'POST',
            body: JSON.stringify({ refresh_token: refreshToken })
        }),

        getMyProfile: () => API.request('/auth/me', { method: 'GET' })
    },

    // User endpoints
    users: {
        list: () => API.request('/auth/users', { method: 'GET' }),
        get: (id) => API.request(`/auth/users/${id}`, { method: 'GET' }),
        update: (id, data) => API.request(`/auth/users/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data)
        }),
        delete: (id) => API.request(`/auth/users/${id}`, { method: 'DELETE' }),
        activate: (id) => API.request(`/auth/users/${id}/activate`, { method: 'POST' }),
        deactivate: (id) => API.request(`/auth/users/${id}/deactivate`, { method: 'POST' }),
        updatePassword: (id, newPassword) => API.request(`/auth/users/${id}/password`, {
            method: 'PUT',
            body: JSON.stringify({ new_password: newPassword })
        })
    },

    // Role endpoints
    roles: {
        list: () => API.request('/auth/roles', { method: 'GET' }),
        get: (id) => API.request(`/auth/roles/${id}`, { method: 'GET' }),
        create: (data) => API.request('/auth/roles', {
            method: 'POST',
            body: JSON.stringify(data)
        }),
        update: (id, data) => API.request(`/auth/roles/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data)
        }),
        delete: (id) => API.request(`/auth/roles/${id}`, { method: 'DELETE' }),
        getPermissions: (roleId) => API.request(`/auth/roles/${roleId}/permissions`, { method: 'GET' }),
        assignPermission: (roleId, permissionId) => API.request(`/auth/roles/${roleId}/permissions`, {
            method: 'POST',
            body: JSON.stringify({ permission_id: permissionId })
        }),
        revokePermission: (roleId, permissionId) => API.request(`/auth/roles/${roleId}/permissions/${permissionId}`, {
            method: 'DELETE'
        })
    },

    // Permission endpoints
    permissions: {
        list: () => API.request('/auth/permissions', { method: 'GET' }),
        get: (id) => API.request(`/auth/permissions/${id}`, { method: 'GET' }),
        create: (data) => API.request('/auth/permissions', {
            method: 'POST',
            body: JSON.stringify(data)
        }),
        update: (id, data) => API.request(`/auth/permissions/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data)
        }),
        delete: (id) => API.request(`/auth/permissions/${id}`, { method: 'DELETE' })
    },

    // RBAC endpoints
    rbac: {
        getUserRoles: (userId) => API.request(`/auth/users/${userId}/roles`, { method: 'GET' }),
        assignRole: (userId, roleId) => API.request(`/auth/users/${userId}/roles`, {
            method: 'POST',
            body: JSON.stringify({ role_id: roleId })
        }),
        revokeRole: (userId, roleId) => API.request(`/auth/users/${userId}/roles/${roleId}`, {
            method: 'DELETE'
        }),
        getUserPermissions: (userId) => API.request(`/auth/users/${userId}/permissions`, { method: 'GET' }),
        checkPermission: (userId, resource, action) => API.request(`/auth/users/${userId}/check-permission`, {
            method: 'POST',
            body: JSON.stringify({ resource, action })
        })
    },

    // Test endpoints
    test: {
        public: () => API.request('/test/public', { method: 'GET' }),
        protected: () => API.request('/test/protected', { method: 'GET' }),
        admin: () => API.request('/test/admin', { method: 'GET' })
    }
};
