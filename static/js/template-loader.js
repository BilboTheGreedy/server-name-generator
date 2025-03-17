/**
 * Template loader - loads HTML templates into the main content area
 */
document.addEventListener('DOMContentLoaded', function() {
    // Templates to load
    const templates = [
        { id: 'dashboard', path: '/static/templates/dashboard.html' },
        { id: 'generate', path: '/static/templates/generate.html' },
        { id: 'manage', path: '/static/templates/manage.html' },
        { id: 'users', path: '/static/templates/users.html' },
        { id: 'apikeys', path: '/static/templates/apikeys.html' },
        { id: 'statistics', path: '/static/templates/statistics.html' },
        { id: 'apiExplorer', path: '/static/templates/api-explorer.html' }
    ];
    
    // Load all templates
    Promise.all(templates.map(template => 
        fetch(template.path)
            .then(response => {
                if (!response.ok) {
                    throw new Error(`Failed to load template: ${template.path}`);
                }
                return response.text();
            })
            .then(html => {
                document.getElementById(template.id).innerHTML = html;
                return template.id;
            })
            .catch(error => {
                console.error(`Error loading template ${template.path}:`, error);
                document.getElementById(template.id).innerHTML = `
                    <div class="alert alert-danger">
                        <i class="bi bi-exclamation-triangle-fill me-2"></i>
                        Failed to load content: ${error.message}
                    </div>
                `;
                return template.id;
            })
    )).then(results => {
        console.log('All templates loaded:', results.join(', '));
        // Trigger a custom event to signal templates are loaded
        document.dispatchEvent(new CustomEvent('templatesLoaded'));
    });
});