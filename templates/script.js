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
    const formData = new FormData(form);

    const endpoint = {
        'ngram-duplicate-finder': '/ngram_finder',
        'interactive-mode': '/interactive_mode',
        'automatic-mode': '/automatic_mode',
        'heuristic-mode': 'heuristic_mode'
    }[analysisMethod.value];

    try {
        const response = await fetch(endpoint, {
            method: 'POST',
            body: formData,
        });

        const result = await response.json();
        alert('Success: ' + JSON.stringify(result));
    } catch (error) {
        console.error('Error:', error);
        alert('An error occurred.');
    }
});
