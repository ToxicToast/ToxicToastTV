// User Management Module
const Users = {
    async loadUsers() {
        try {
            const response = await API.users.list();
            const tbody = document.getElementById('users-table-body');

            if (!response.users || response.users.length === 0) {
                tbody.innerHTML = '<tr><td colspan="6">No users found</td></tr>';
                return;
            }

            tbody.innerHTML = response.users.map(user => `
                <tr>
                    <td>${user.id.substring(0, 8)}...</td>
                    <td>${user.email}</td>
                    <td>${user.username}</td>
                    <td><span class="badge badge-${user.status === 'ACTIVE' ? 'success' : 'danger'}">${user.status}</span></td>
                    <td>${new Date(user.created_at).toLocaleString()}</td>
                    <td>
                        <button class="btn btn-secondary" onclick="Users.viewUser('${user.id}')">View</button>
                        <button class="btn btn-${user.status === 'ACTIVE' ? 'warning' : 'success'}"
                                onclick="Users.toggleStatus('${user.id}', '${user.status}')">
                            ${user.status === 'ACTIVE' ? 'Deactivate' : 'Activate'}
                        </button>
                    </td>
                </tr>
            `).join('');
        } catch (error) {
            alert('Failed to load users: ' + error.message);
        }
    },

    async viewUser(userId) {
        try {
            const response = await API.users.get(userId);
            const roles = await API.rbac.getUserRoles(userId);

            showModal(`
                <h2>User Details</h2>
                <div class="form-group">
                    <strong>Email:</strong> ${response.user.email}<br>
                    <strong>Username:</strong> ${response.user.username}<br>
                    <strong>Status:</strong> ${response.user.status}<br>
                    <strong>Roles:</strong> ${roles.roles ? roles.roles.map(r => r.name).join(', ') : 'None'}
                </div>
                <button class="btn btn-secondary" onclick="hideModal()">Close</button>
            `);
        } catch (error) {
            alert('Failed to load user: ' + error.message);
        }
    },

    async toggleStatus(userId, currentStatus) {
        try {
            if (currentStatus === 'ACTIVE') {
                await API.users.deactivate(userId);
            } else {
                await API.users.activate(userId);
            }
            await this.loadUsers();
        } catch (error) {
            alert('Failed to update status: ' + error.message);
        }
    }
};
