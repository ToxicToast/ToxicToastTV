// Global Modal Functions
function showModal(content) {
    document.getElementById('modal-content').innerHTML = content;
    document.getElementById('modal-overlay').style.display = 'flex';
}

function hideModal() {
    document.getElementById('modal-overlay').style.display = 'none';
}

// Application State
const App = {
    currentView: 'dashboard',

    init() {
        // Check if user is logged in
        if (Auth.isLoggedIn()) {
            this.showMainContent();
            this.loadDashboard();
        } else {
            this.showLoginScreen();
        }

        // Setup event listeners
        this.setupEventListeners();
    },

    setupEventListeners() {
        // Login form
        document.getElementById('login-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const email = document.getElementById('login-email').value;
            const password = document.getElementById('login-password').value;

            try {
                await Auth.login(email, password);
                this.showMainContent();
                this.loadDashboard();
            } catch (error) {
                document.getElementById('login-error').textContent = error.message;
                document.getElementById('login-error').style.display = 'block';
            }
        });

        // Register form
        document.getElementById('register-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const userData = {
                email: document.getElementById('register-email').value,
                username: document.getElementById('register-username').value,
                password: document.getElementById('register-password').value,
                first_name: document.getElementById('register-firstname').value,
                last_name: document.getElementById('register-lastname').value
            };

            try {
                await Auth.register(userData);
                this.showMainContent();
                this.loadDashboard();
            } catch (error) {
                document.getElementById('register-error').textContent = error.message;
                document.getElementById('register-error').style.display = 'block';
            }
        });

        // Show/Hide register
        document.getElementById('show-register-btn').addEventListener('click', () => {
            document.querySelector('.login-card').style.display = 'none';
            document.getElementById('register-card').style.display = 'block';
        });

        document.getElementById('back-to-login-btn').addEventListener('click', () => {
            document.querySelector('.login-card').style.display = 'block';
            document.getElementById('register-card').style.display = 'none';
        });

        // Logout
        document.getElementById('logout-btn').addEventListener('click', async () => {
            await Auth.logout();
            this.showLoginScreen();
        });

        // Navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                const view = e.target.getAttribute('data-view');
                this.showView(view);
            });
        });

        // Refresh buttons
        document.getElementById('refresh-users-btn')?.addEventListener('click', () => Users.loadUsers());
        document.getElementById('refresh-roles-btn')?.addEventListener('click', () => Roles.loadRoles());
        document.getElementById('refresh-permissions-btn')?.addEventListener('click', () => Permissions.loadPermissions());

        // Create buttons
        document.getElementById('create-role-btn')?.addEventListener('click', () => Roles.createRole());
        document.getElementById('create-permission-btn')?.addEventListener('click', () => Permissions.createPermission());

        // Test buttons
        document.getElementById('test-public-btn')?.addEventListener('click', () => this.testEndpoint('public'));
        document.getElementById('test-protected-btn')?.addEventListener('click', () => this.testEndpoint('protected'));
        document.getElementById('test-admin-btn')?.addEventListener('click', () => this.testEndpoint('admin'));

        // Copy token
        document.getElementById('copy-token-btn')?.addEventListener('click', () => {
            const token = localStorage.getItem('access_token');
            navigator.clipboard.writeText(token);
            alert('Token copied to clipboard!');
        });

        // Modal overlay click
        document.getElementById('modal-overlay').addEventListener('click', (e) => {
            if (e.target.id === 'modal-overlay') {
                hideModal();
            }
        });
    },

    showLoginScreen() {
        document.getElementById('login-screen').style.display = 'block';
        document.getElementById('main-content').style.display = 'none';
        document.getElementById('logout-btn').style.display = 'none';
        document.getElementById('username-display').textContent = 'Not logged in';
    },

    showMainContent() {
        document.getElementById('login-screen').style.display = 'none';
        document.getElementById('main-content').style.display = 'block';
        document.getElementById('logout-btn').style.display = 'inline-block';

        const user = Auth.getCurrentUser();
        document.getElementById('username-display').textContent = user?.username || user?.email || 'User';
    },

    showView(viewName) {
        // Update nav
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
            if (link.getAttribute('data-view') === viewName) {
                link.classList.add('active');
            }
        });

        // Hide all views
        document.querySelectorAll('.view').forEach(view => {
            view.classList.remove('active');
        });

        // Show selected view
        document.getElementById(`${viewName}-view`).classList.add('active');
        this.currentView = viewName;

        // Load view data
        switch (viewName) {
            case 'dashboard':
                this.loadDashboard();
                break;
            case 'users':
                Users.loadUsers();
                break;
            case 'roles':
                Roles.loadRoles();
                break;
            case 'permissions':
                Permissions.loadPermissions();
                break;
            case 'test-auth':
                this.loadTestAuth();
                break;
        }
    },

    async loadDashboard() {
        try {
            const [profile, users, roles, permissions] = await Promise.all([
                API.auth.getMyProfile(),
                API.users.list(),
                API.roles.list(),
                API.permissions.list()
            ]);

            // Update stats
            document.getElementById('total-users').textContent = users.users?.length || 0;
            document.getElementById('total-roles').textContent = roles.roles?.length || 0;
            document.getElementById('total-permissions').textContent = permissions.permissions?.length || 0;
            document.getElementById('my-roles').textContent = profile.roles?.length || 0;

            // Update profile
            document.getElementById('profile-info').innerHTML = `
                <p><strong>Email:</strong> ${profile.user.email}</p>
                <p><strong>Username:</strong> ${profile.user.username}</p>
                <p><strong>Roles:</strong> ${profile.roles?.join(', ') || 'None'}</p>
                <p><strong>Permissions:</strong> ${profile.permissions?.length || 0} permissions</p>
            `;
        } catch (error) {
            console.error('Failed to load dashboard:', error);
        }
    },

    loadTestAuth() {
        const token = localStorage.getItem('access_token');
        document.getElementById('access-token-display').value = token || 'No token';

        try {
            const decoded = Auth.getCurrentUser();
            document.getElementById('decoded-claims').textContent = JSON.stringify(decoded, null, 2);
        } catch (e) {
            document.getElementById('decoded-claims').textContent = 'Invalid token';
        }
    },

    async testEndpoint(type) {
        const resultsDiv = document.getElementById('test-results');

        try {
            let result;
            switch (type) {
                case 'public':
                    result = await API.test.public();
                    break;
                case 'protected':
                    result = await API.test.protected();
                    break;
                case 'admin':
                    result = await API.test.admin();
                    break;
            }

            resultsDiv.innerHTML += `
                <div class="test-result">
                    <strong>✅ ${type.toUpperCase()} endpoint successful</strong>
                    <pre>${JSON.stringify(result, null, 2)}</pre>
                </div>
            `;
        } catch (error) {
            resultsDiv.innerHTML += `
                <div class="test-result" style="border-left-color: #e74c3c;">
                    <strong>❌ ${type.toUpperCase()} endpoint failed</strong>
                    <pre>${error.message}</pre>
                </div>
            `;
        }
    }
};

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    App.init();
});
