// results.js - Interactive mode results display
class InteractiveResults {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.groups = new Map();
        this.archetypes = new Map();
    }

    // Initialize the results display
    init() {
        this.container.innerHTML = `
            <div class="results-header">
                <h2>Interactive Analysis Results</h2>
                <div class="results-controls">
                    <button id="expand-all">Expand All</button>
                    <button id="collapse-all">Collapse All</button>
                    <button id="export-results">Export Results</button>
                </div>
            </div>
            <div class="results-content"></div>
        `;

        // Add event listeners
        document.getElementById('expand-all').addEventListener('click', () => this.expandAll());
        document.getElementById('collapse-all').addEventListener('click', () => this.collapseAll());
        document.getElementById('export-results').addEventListener('click', () => this.exportResults());
    }

    // Display analysis results
    displayResults(data) {
        const { groups, archetypes } = data;
        this.groups = new Map(Object.entries(groups));
        this.archetypes = new Map(Object.entries(archetypes));
        
        const content = this.container.querySelector('.results-content');
        content.innerHTML = '';

        this.groups.forEach((fragments, groupId) => {
            const archetype = this.archetypes.get(groupId) || '';
            const groupElement = this.createGroupElement(groupId, fragments, archetype);
            content.appendChild(groupElement);
        });
    }

    // Create a group element
    createGroupElement(groupId, fragments, archetype) {
        const div = document.createElement('div');
        div.className = 'result-group';
        div.innerHTML = `
            <div class="group-header">
                <h3>Group ${groupId}</h3>
                <button class="toggle-group">Expand</button>
            </div>
            <div class="group-content" style="display: none;">
                ${archetype ? `<div class="archetype">
                    <h4>Archetype:</h4>
                    <pre>${archetype}</pre>
                </div>` : ''}
                <div class="fragments">
                    <h4>Fragments (${fragments.length}):</h4>
                    <ul>
                        ${fragments.map((frag, i) => `
                            <li>
                                <span class="fragment-number">${i + 1}.</span>
                                <pre>${frag}</pre>
                            </li>
                        `).join('')}
                    </ul>
                </div>
                <div class="group-actions">
                    <button class="delete-group">Delete Group</button>
                    <button class="merge-group">Merge with...</button>
                </div>
            </div>
        `;

        // Add event listeners
        const toggleBtn = div.querySelector('.toggle-group');
        const content = div.querySelector('.group-content');
        toggleBtn.addEventListener('click', () => {
            const isExpanded = content.style.display !== 'none';
            content.style.display = isExpanded ? 'none' : 'block';
            toggleBtn.textContent = isExpanded ? 'Expand' : 'Collapse';
        });

        return div;
    }

    // Expand all groups
    expandAll() {
        this.container.querySelectorAll('.group-content').forEach(content => {
            content.style.display = 'block';
        });
        this.container.querySelectorAll('.toggle-group').forEach(btn => {
            btn.textContent = 'Collapse';
        });
    }

    // Collapse all groups
    collapseAll() {
        this.container.querySelectorAll('.group-content').forEach(content => {
            content.style.display = 'none';
        });
        this.container.querySelectorAll('.toggle-group').forEach(btn => {
            btn.textContent = 'Expand';
        });
    }

    // Export results to JSON
    exportResults() {
        const data = {
            groups: Object.fromEntries(this.groups),
            archetypes: Object.fromEntries(this.archetypes)
        };
        
        const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'interactive-analysis-results.json';
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }
}

// Export the class
window.InteractiveResults = InteractiveResults; 