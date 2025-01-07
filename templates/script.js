const form = document.getElementById('analysis-method-form');
const analysisMethod = document.getElementById('analysis-method');
const ngramSettings = document.getElementById('ngram-settings');
const interactiveSettings = document.getElementById('interactive-settings');
const automaticSettings = document.getElementById('automatic-settings');
const heuristicSettings = document.getElementById('heuristic-settings');

analysisMethod.addEventListener('change', () => {
    const method = analysisMethod.value;
    [ngramSettings, interactiveSettings, automaticSettings].forEach(setting => setting.style.display = 'none');
    if (method === 'ngram-duplicate-finder') ngramSettings.style.display = 'block';
    if (method === 'interactive-mode') interactiveSettings.style.display = 'block';
    if (method === 'automatic-mode') automaticSettings.style.display = 'block';
    if (method === 'heuristic-mode') heuristicSettings.style.display = 'block';
});

form.addEventListener('submit', async (event) => {
    event.preventDefault();

    let requestData = {};
    const method = analysisMethod.value;

    if (method === 'ngram-duplicate-finder') {
        requestData = {
            min_clone_slider: parseInt(document.getElementById('min-clone-length').value, 10),
            max_edit_slider: parseInt(document.getElementById('max-edit-distance').value, 10),
            max_fuzzy_slider: parseInt(document.getElementById('max-fuzzy-hash-distance').value, 10),
            source_language: document.getElementById('source-language').value,
        };
    } else if (method === 'interactive-mode') {
        requestData = {
            min_clone_slider: parseInt(document.getElementById('interactive-min-length').value, 10),
            max_clone_slider: parseInt(document.getElementById('interactive-max-length').value, 10),
            min_group_slider: parseInt(document.getElementById('group-power').value, 10),
            extension_checkbox: document.getElementById('archetype').checked,
        };
    } else if (method === 'automatic-mode') {
        requestData = {
            length_slider: parseInt(document.getElementById('auto-min-clone-length').value, 10),
            convert_checkbox: document.getElementById('strict-filter').checked,
            archetype_slider: parseInt(document.getElementById('archetype-length').value, 10),
            strict_filtering_checkbox: document.getElementById('strict-filter').checked,
        };
    } else if (method === 'heuristic-mode') {
        requestData = {
            extention_point_checkbox: document.getElementById('extension-value').checked,
        };
    }

    console.log('Request data:', requestData);

    const endpoint = {
        'ngram-duplicate-finder': '/ngram_finder',
        'interactive-mode': '/interactive_mode',
        'automatic-mode': '/automatic_mode',
        'heuristic-mode': '/heuristic_finder'
    }[method];

    try {
        const response = await fetch(endpoint, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(requestData),
        });

        if (!response.ok) {
            const text = await response.text();
            throw new Error(`Server returned ${response.status}: ${text}`);
        }

        const result = await response.json();
        console.info('Server result:', JSON.stringify(result));
        alert('Success: ' + JSON.stringify(result));
    } catch (error) {
        console.error('Error:', error);
        alert('An error occurred: ' + error.message);
    }
});
