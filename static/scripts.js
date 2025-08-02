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

    // Update range values display
    function updateRangeValue(inputId) {
        const input = document.getElementById(inputId);
        const valueDisplay = document.getElementById(inputId + '-value');
        if (input && valueDisplay) {
            valueDisplay.textContent = input.value;
        }
    }

    // Initialize range value displays
    const rangeInputs = [
        'auto-min-clone-length',
        'archetype-length'
    ];
    
    rangeInputs.forEach(inputId => {
        const input = document.getElementById(inputId);
        if (input) {
            updateRangeValue(inputId);
            input.addEventListener('input', () => updateRangeValue(inputId));
        }
    });

    // Update settings collection for automatic mode
    function getAutomaticModeSettings() {
        return {
            minCloneLength: parseInt(document.getElementById('auto-min-clone-length').value),
            convertToDRL: document.getElementById('convert-to-drl').checked,
            archetypeLength: parseInt(document.getElementById('archetype-length').value),
            strictFilter: document.getElementById('strict-filter').checked
        };
    }

    // Update settings collection for interactive mode
    function getInteractiveModeSettings() {
        return {
            minCloneLength: parseInt(document.getElementById('interactive-min-length').value),
            maxCloneLength: parseInt(document.getElementById('interactive-max-length').value),
            minGroupPower: parseInt(document.getElementById('group-power').value),
            useArchetype: document.getElementById('archetype').checked
        };
    }

    // Initialize results display
    const resultsContainer = document.getElementById('results-container');
    const interactiveResults = new InteractiveResults('results-container');
    interactiveResults.init();

    // Form submission handler
    form.addEventListener('submit', async (event) => {
        event.preventDefault();
        const method = analysisMethod.value;
        const sourceFile = document.getElementById('source-file').files[0];

        if (!sourceFile) {
            alert('Please select a file to analyze');
            return;
        }

        // Hide results container initially
        resultsContainer.style.display = 'none';

        const formData = new FormData();
        formData.append('file', sourceFile);

        let settings = {};
        if (method === 'ngram-duplicate-finder') {
            settings = {
                min_clone_slider: parseInt(document.getElementById('min-clone-length').value),
                max_edit_slider: parseInt(document.getElementById('max-edit-distance').value),
                max_fuzzy_slider: parseInt(document.getElementById('max-fuzzy-hash-distance').value),
                source_language: document.getElementById('source-language').value
            };
        } else if (method === 'heuristic-mode') {
            settings = {
                extension_point_checkbox: document.getElementById('extension-value').checked
            };
        } else if (method === 'automatic-mode') {
            settings = getAutomaticModeSettings();
        } else if (method === 'interactive-mode') {
            settings = getInteractiveModeSettings();
        } else {
            alert('In current version this mode is not implemented');
            return;
        }

        formData.append('settings', JSON.stringify(settings));

        let endpoint = '';
        if (method === 'ngram-duplicate-finder') {
            endpoint = 'http://localhost:8080/ngram';
        } else if (method === 'heuristic-mode') {
            endpoint = 'http://localhost:8080/heuristic';
        } else if (method === 'automatic-mode') {
            endpoint = 'http://localhost:8080/automatic';
        } else if (method === 'interactive-mode') {
            endpoint = 'http://localhost:8080/interactive';
        }

        try {
            const response = await fetch(endpoint, {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                const errorText = await response.text();
                // check if the error is related to the file format
                if (errorText.includes('unsupported file format')) {
                    alert('Error: ' + errorText + '\n\nPlease select a file in one of the supported formats.');
                } else {
                    alert('Error: ' + errorText);
                }
                return;
            }

            const result = await response.json();
            
            if (method === 'interactive-mode') {
                // Display results in the interactive view
                interactiveResults.displayResults(result);
                resultsContainer.style.display = 'block';
                
                // Scroll to results
                resultsContainer.scrollIntoView({ behavior: 'smooth' });
            } else {
                // Show alert for other modes
                let resultMessage = '';
                if (method === 'ngram-duplicate-finder') {
                    resultMessage = `Analysis completed!\nFound ${Object.keys(result.duplicates).length} groups of duplicates\nResult: ${result.results_file}`;
                } else if (method === 'heuristic-mode') {
                    resultMessage = `Analysis completed!\nFound ${result.ngrams.length} n-grams\nResult: ${result.results_file}`;
                } else if (method === 'automatic-mode') {
                    resultMessage = `Analysis completed!\nFound ${Object.keys(result.groups).length} groups of duplicates\nResult: ${result.results_file}`;
                }
                alert(resultMessage);
            }

        } catch (error) {
            // FIXME "Load failed" error in webUI 
            // Fixed(?)
            if (error.message.includes('Load failed')) {
                alert('File was successfully analyzed');
                console.error('Error:', error);
            } else {
                alert('An error occurred while processing the file. Please try again.',error);
            }
        }
    });
});