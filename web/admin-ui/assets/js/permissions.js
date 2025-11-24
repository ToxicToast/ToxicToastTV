// Permission Management Module
const Permissions = {
    async loadPermissions() {
        try {
            const response = await API.permissions.list();
            const container = document.getElementById('permissions-list-container');

            if (!response.permissions || response.permissions.length === 0) {
                container.innerHTML = '<p>No permissions found</p>';
                return;
            }

            container.innerHTML = response.permissions.map(perm => `
                <div class="card">
                    <h3>${perm.resource}:${perm.action}</h3>
                    <p>${perm.description || 'No description'}</p>
                    <button class="btn btn-danger" onclick="Permissions.deletePermission('${perm.id}')">Delete</button>
                </div>
            `).join('');
        } catch (error) {
            alert('Failed to load permissions: ' + error.message);
        }
    },

    async deletePermission(permId) {
        if (!confirm('Delete this permission?')) return;
        try {
            await API.permissions.delete(permId);
            await this.loadPermissions();
        } catch (error) {
            alert('Failed to delete permission: ' + error.message);
        }
    },

    async createPermission() {
        const resource = prompt('Resource (e.g., "posts"):');
        const action = prompt('Action (e.g., "read", "write"):');
        const description = prompt('Description:');
        if (!resource || !action) return;

        try {
            await API.permissions.create({ resource, action, description });
            await this.loadPermissions();
        } catch (error) {
            alert('Failed to create permission: ' + error.message);
        }
    }
};
