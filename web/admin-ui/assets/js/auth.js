// Authentication Module
const Auth = {
    // Check if user is logged in
    isLoggedIn() {
        return !!localStorage.getItem('access_token');
    },

    // Get current user info from token
    getCurrentUser() {
        const token = localStorage.getItem('access_token');
        if (!token) return null;

        try {
            const payload = JSON.parse(atob(token.split('.')[1]));
            return payload;
        } catch (e) {
            return null;
        }
    },

    // Store tokens
    setTokens(accessToken, refreshToken) {
        localStorage.setItem('access_token', accessToken);
        if (refreshToken) {
            localStorage.setItem('refresh_token', refreshToken);
        }
    },

    // Clear tokens
    clearTokens() {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
    },

    // Login
    async login(email, password) {
        try {
            const response = await API.auth.login({ email, password });
            this.setTokens(response.access_token, response.refresh_token);
            return response;
        } catch (error) {
            throw error;
        }
    },

    // Register
    async register(userData) {
        try {
            const response = await API.auth.register(userData);
            this.setTokens(response.access_token, response.refresh_token);
            return response;
        } catch (error) {
            throw error;
        }
    },

    // Logout
    async logout() {
        try {
            await API.auth.logout();
        } catch (e) {
            // Ignore errors on logout
        } finally {
            this.clearTokens();
        }
    },

    // Refresh token if expired
    async refreshIfNeeded() {
        const user = this.getCurrentUser();
        if (!user) return false;

        // Check if token expires in next 5 minutes
        const expiresAt = user.exp * 1000;
        const now = Date.now();
        const fiveMinutes = 5 * 60 * 1000;

        if (expiresAt - now < fiveMinutes) {
            try {
                const refreshToken = localStorage.getItem('refresh_token');
                const response = await API.auth.refresh(refreshToken);
                this.setTokens(response.access_token, response.refresh_token);
                return true;
            } catch (e) {
                this.clearTokens();
                return false;
            }
        }

        return true;
    }
};
