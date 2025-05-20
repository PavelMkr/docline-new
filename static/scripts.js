document.addEventListener('DOMContentLoaded', function() {
    // Get DOM elements
    const form = document.getElementById('analysis-method-form');
    const analysisMethod = document.getElementById('analysis-method');
    const ngramSettings = document.getElementById('ngram-settings');
    const interactiveSettings = document.getElementById('interactive-settings');
    const automaticSettings = document.getElementById('automatic-settings');
    const heuristicSettings = document.getElementById('heuristic-settings');
    const submitButton = document.querySelector('button[type="submit"]');
    const formTitle = document.querySelector('h1');

    // Store form values
    let formValues = {
        'ngram-duplicate-finder': {
            'min-clone-length': '25',
            'max-edit-distance': '9',
            'max-fuzzy-hash-distance': '2',
            'source-language': 'english'
        },
        'interactive-mode': {
            'interactive-min-length': '20',
            'interactive-max-length': '50',
            'group-power': '2',
            'archetype': false
        },
        'automatic-mode': {
            'auto-min-clone-length': '20',
            'archetype-length': '5',
            'strict-filter': true
        },
        'heuristic-mode': {
            'extension-value': true
        }
    };

    // Function to save current form values
    function saveCurrentFormValues(method) {
        const values = {};
        let settingsGroup;
        
        // Map method to correct settings group ID
        switch(method) {
            case 'ngram-duplicate-finder':
                settingsGroup = ngramSettings;
                break;
            case 'interactive-mode':
                settingsGroup = interactiveSettings;
                break;
            case 'automatic-mode':
                settingsGroup = automaticSettings;
                break;
            case 'heuristic-mode':
                settingsGroup = heuristicSettings;
                break;
        }

        if (!settingsGroup) return;

        settingsGroup.querySelectorAll('input, select').forEach(element => {
            values[element.id] = element.type === 'checkbox' ? element.checked : element.value;
        });
        formValues[method] = values;
    }

    // Function to restore form values
    function restoreFormValues(method) {
        const values = formValues[method];
        if (!values) return;

        Object.keys(values).forEach(id => {
            const element = document.getElementById(id);
            if (element) {
                if (element.type === 'checkbox') {
                    element.checked = values[id];
                } else {
                    element.value = values[id];
                }
            }
        });
    }

    // Analysis method change handler
    analysisMethod.addEventListener('change', () => {
        const method = analysisMethod.value;
        
        // Save current values before switching
        const currentMethod = Object.keys(formValues).find(key => {
            const settingsId = key === 'ngram-duplicate-finder' ? 'ngram-settings' :
                             key === 'interactive-mode' ? 'interactive-settings' :
                             key === 'automatic-mode' ? 'automatic-settings' :
                             'heuristic-settings';
            return document.getElementById(settingsId).style.display === 'block';
        });
        
        if (currentMethod) {
            saveCurrentFormValues(currentMethod);
        }

        // Hide all settings
        [ngramSettings, interactiveSettings, automaticSettings, heuristicSettings].forEach(setting => {
            if (setting) setting.style.display = 'none';
        });

        // Show selected settings and update UI
        const selectedSettings = {
            'ngram-duplicate-finder': { title: 'Ngram Analysis Interface', button: 'Run Ngram Analysis', element: ngramSettings },
            'interactive-mode': { title: 'Interactive Analysis Interface', button: 'Start Interactive Analysis', element: interactiveSettings },
            'automatic-mode': { title: 'Automatic Analysis Interface', button: 'Run Automatic Analysis', element: automaticSettings },
            'heuristic-mode': { title: 'Heuristic Analysis Interface', button: 'Run Heuristic Analysis', element: heuristicSettings }
        }[method];

        if (selectedSettings) {
            selectedSettings.element.style.display = 'block';
            formTitle.textContent = selectedSettings.title;
            submitButton.textContent = selectedSettings.button;
            restoreFormValues(method);
        }
    });

    // Form submission handler
    form.addEventListener('submit', async (event) => {
        event.preventDefault();
        const method = analysisMethod.value;
        const sourceFile = document.getElementById('source-file').files[0];
        
        if (!sourceFile) {
            alert('Please select a file to analyze');
            return;
        }

        try {
            const formData = new FormData();
            formData.append('file', sourceFile);

            let settings = {};
            if (method === 'ngram-duplicate-finder') {
                settings = {
                    minCloneSlider: parseInt(document.getElementById('min-clone-length').value),
                    maxEditSlider: parseInt(document.getElementById('max-edit-distance').value),
                    maxFuzzySlider: parseInt(document.getElementById('max-fuzzy-hash-distance').value),
                    sourceLanguage: document.getElementById('source-language').value
                };
            } else if (method === 'heuristic-mode') {
                settings = {
                    extensionPointCheckbox: document.getElementById('extension-value').checked
                };
            }

            formData.append('settings', JSON.stringify(settings));

            let endpoint = '';
            if (method === 'ngram-duplicate-finder') {
                endpoint = '/ngram';
            } else if (method === 'heuristic-mode') {
                endpoint = '/heuristic';
            }

            const response = await fetch(endpoint, {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const result = await response.json();
            console.log('Analysis completed:', result);

            // Display results
            let resultMessage = '';
            if (method === 'ngram-duplicate-finder') {
                resultMessage = `Analysis completed!\nFound ${Object.keys(result.duplicates).length} duplicate groups.\nResults saved to: ${result.results_file}`;
            } else if (method === 'heuristic-mode') {
                resultMessage = `Analysis completed!\nFound ${result.ngrams.length} n-grams.\nResults saved to: ${result.results_file}`;
            }
            alert(resultMessage);

        } catch (error) {
            console.error('Error:', error);
            alert('Error: ' + error.message);
        }
    });
});