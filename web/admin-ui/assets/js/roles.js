// Role Management Module
const Roles = {
    async loadRoles() {
        try {
            const response = await API.roles.list();
            const container = document.getElementById('roles-list-container');

            if (!response.roles || response.roles.length === 0) {
                container.innerHTML = '<p>No roles found</p>';
                return;
            }

            container.innerHTML = response.roles.map(role => `
                <div class="card">
                    <h3>${role.name}</h3>
                    <p>${role.description || 'No description'}</p>
                    <button class="btn btn-secondary" onclick="Roles.viewRole('${role.id}')">View Details</button>
                    <button class="btn btn-danger" onclick="Roles.deleteRole('${role.id}')">Delete</button>
                </div>
            `).join('');
        } catch (error) {
            alert('Failed to load roles: ' + error.message);
        }
    },

    async viewRole(roleId) {
        try {
            const role = await API.roles.get(roleId);
            const perms = await API.roles.getPermissions(roleId);

            showModal(`
                <h2>${role.role.name}</h2>
                <p>${role.role.description}</p>
                <h3>Permissions</h3>
                <ul>${(perms.permissions || []).map(p => `<li>${p.resource}:${p.action}</li>`).join('')}</ul>
                <button class="btn btn-secondary" onclick="hideModal()">Close</button>
            `);
        } catch (error) {
            alert('Failed to load role: ' + error.message);
        }
    },

    async deleteRole(roleId) {
        if (!confirm('Delete this role?')) return;
        try {
            await API.roles.delete(roleId);
            await this.loadRoles();
        } catch (error) {
            alert('Failed to delete role: ' + error.message);
        }
    },

    async createRole() {
        const name = prompt('Role name:');
        const description = prompt('Description:');
        if (!name) return;

        try {
            await API.roles.create({ name, description });
            await this.loadRoles();
        } catch (error) {
            alert('Failed to create role: ' + error.message);
        }
    }
};
