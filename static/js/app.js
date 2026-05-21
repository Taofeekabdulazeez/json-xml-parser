document.addEventListener('DOMContentLoaded', () => {
    // DOM Elements - Inputs & Gutters
    const jsonInput = document.getElementById('jsonInput');
    const xmlInput = document.getElementById('xmlInput');
    const jsonLines = document.getElementById('jsonLines');
    const xmlLines = document.getElementById('xmlLines');

    // DOM Elements - Settings
    const rootElement = document.getElementById('rootElement');
    const arrayMode = document.getElementById('arrayMode');
    const attrPrefix = document.getElementById('attrPrefix');
    const textKey = document.getElementById('textKey');
    const autoType = document.getElementById('autoType');
    const prettyPrint = document.getElementById('prettyPrint');

    // DOM Elements - Status Footers
    const jsonFooter = document.getElementById('jsonFooter');
    const xmlFooter = document.getElementById('xmlFooter');

    // DOM Elements - Buttons
    const formatJsonBtn = document.getElementById('formatJson');
    const clearJsonBtn = document.getElementById('clearJson');
    const copyJsonBtn = document.getElementById('copyJson');
    const formatXmlBtn = document.getElementById('formatXml');
    const clearXmlBtn = document.getElementById('clearXml');
    const copyXmlBtn = document.getElementById('copyXml');

    // DOM Elements - Modal Drawer
    const apiDocBtn = document.getElementById('apiDocBtn');
    const closeModal = document.getElementById('closeModal');
    const apiModal = document.getElementById('apiModal');

    // Track which editor is currently being edited to prevent infinite circular conversion loops
    let activeEditor = null;
    let lastModifiedEditor = 'json'; // Defaults to json

    // --- Line Numbering Gutter Sync ---
    function updateLineNumbers(textarea, gutter) {
        const lines = textarea.value.split('\n');
        const lineCount = lines.length;
        let lineNumbersHTML = '';
        for (let i = 1; i <= lineCount; i++) {
            lineNumbersHTML += `<div>${i}</div>`;
        }
        gutter.innerHTML = lineNumbersHTML;
    }

    function syncScroll(textarea, gutter) {
        gutter.scrollTop = textarea.scrollTop;
    }

    const initGutter = (textarea, gutter) => {
        updateLineNumbers(textarea, gutter);
        textarea.addEventListener('input', () => updateLineNumbers(textarea, gutter));
        textarea.addEventListener('scroll', () => syncScroll(textarea, gutter));
    };

    initGutter(jsonInput, jsonLines);
    initGutter(xmlInput, xmlLines);

    // Re-sync when textarea is resized/loaded
    window.addEventListener('resize', () => {
        syncScroll(jsonInput, jsonLines);
        syncScroll(xmlInput, xmlLines);
    });

    // --- Debouncing Helper ---
    function debounce(fn, delay) {
        let timeout;
        return function(...args) {
            clearTimeout(timeout);
            timeout = setTimeout(() => fn.apply(this, args), delay);
        };
    }

    // --- Status Indicator UI Helpers ---
    function setStatusReady(footer) {
        footer.innerHTML = `<span class="status-indicator success"><i class="fa-solid fa-check-circle"></i> Ready</span>`;
    }

    function setStatusConverting(footer) {
        footer.innerHTML = `<span class="status-indicator pending"><i class="fa-solid fa-circle-notch fa-spin"></i> Converting...</span>`;
    }

    function setStatusError(footer, message) {
        footer.innerHTML = `<span class="status-indicator error" title="${escapeHtml(message)}"><i class="fa-solid fa-triangle-exclamation"></i> Error: ${escapeHtml(message)}</span>`;
    }

    function escapeHtml(text) {
        return text
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    }

    // --- Bidirectional AJAX Converters ---
    async function convertJSONToXML() {
        const jsonVal = jsonInput.value.trim();
        if (!jsonVal) {
            xmlInput.value = '';
            updateLineNumbers(xmlInput, xmlLines);
            setStatusReady(jsonFooter);
            setStatusReady(xmlFooter);
            return;
        }

        // Validate local JSON syntax first
        try {
            JSON.parse(jsonVal);
        } catch (e) {
            setStatusError(jsonFooter, "Invalid JSON structure: " + e.message);
            return;
        }

        setStatusConverting(xmlFooter);
        setStatusReady(jsonFooter);

        const params = new URLSearchParams({
            root: rootElement.value || 'root',
            array_mode: arrayMode.value,
            attr_prefix: attrPrefix.value || '@',
            pretty: prettyPrint.checked ? 'true' : 'false'
        });

        try {
            const response = await fetch(`/api/v1/convert/json-to-xml?${params.toString()}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: jsonVal
            });

            if (response.ok) {
                const xmlText = await response.text();
                xmlInput.value = xmlText;
                updateLineNumbers(xmlInput, xmlLines);
                setStatusReady(xmlFooter);
            } else {
                const errData = await response.json();
                setStatusError(xmlFooter, errData.error || "Unknown server error during conversion");
            }
        } catch (err) {
            setStatusError(xmlFooter, "Connection failed: " + err.message);
        }
    }

    async function convertXMLToJSON() {
        const xmlVal = xmlInput.value.trim();
        if (!xmlVal) {
            jsonInput.value = '';
            updateLineNumbers(jsonInput, jsonLines);
            setStatusReady(jsonFooter);
            setStatusReady(xmlFooter);
            return;
        }

        setStatusConverting(jsonFooter);
        setStatusReady(xmlFooter);

        const params = new URLSearchParams({
            attr_prefix: attrPrefix.value || '@',
            text_key: textKey.value || '#text',
            auto_type: autoType.checked ? 'true' : 'false',
            pretty: prettyPrint.checked ? 'true' : 'false'
        });

        try {
            const response = await fetch(`/api/v1/convert/xml-to-json?${params.toString()}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/xml' },
                body: xmlVal
            });

            if (response.ok) {
                const jsonText = await response.text();
                jsonInput.value = jsonText;
                updateLineNumbers(jsonInput, jsonLines);
                setStatusReady(jsonFooter);
            } else {
                const errData = await response.json();
                setStatusError(jsonFooter, errData.error || "Unknown server error during conversion");
            }
        } catch (err) {
            setStatusError(jsonFooter, "Connection failed: " + err.message);
        }
    }

    const debouncedConvertJSON = debounce(convertJSONToXML, 300);
    const debouncedConvertXML = debounce(convertXMLToJSON, 300);

    // --- Input Focus & Change Bindings ---
    jsonInput.addEventListener('focus', () => {
        activeEditor = 'json';
        lastModifiedEditor = 'json';
    });
    
    xmlInput.addEventListener('focus', () => {
        activeEditor = 'xml';
        lastModifiedEditor = 'xml';
    });

    jsonInput.addEventListener('input', () => {
        if (activeEditor === 'json') {
            debouncedConvertJSON();
        }
    });

    xmlInput.addEventListener('input', () => {
        if (activeEditor === 'xml') {
            debouncedConvertXML();
        }
    });

    // Re-trigger conversion if settings are modified
    const triggerSettingsReconversion = () => {
        if (lastModifiedEditor === 'json') {
            convertJSONToXML();
        } else {
            convertXMLToJSON();
        }
    };

    [rootElement, arrayMode, attrPrefix, textKey, autoType, prettyPrint].forEach(elem => {
        elem.addEventListener('change', triggerSettingsReconversion);
        if (elem.tagName === 'INPUT' && elem.type === 'text') {
            elem.addEventListener('input', debounce(triggerSettingsReconversion, 250));
        }
    });

    // --- Action Button Handlers ---

    // Formatting JSON locally
    formatJsonBtn.addEventListener('click', () => {
        const val = jsonInput.value.trim();
        if (!val) return;
        try {
            const parsed = JSON.parse(val);
            jsonInput.value = JSON.stringify(parsed, null, prettyPrint.checked ? 2 : 0);
            updateLineNumbers(jsonInput, jsonLines);
            setStatusReady(jsonFooter);
            convertJSONToXML();
        } catch (e) {
            setStatusError(jsonFooter, "Cannot format. Invalid JSON: " + e.message);
        }
    });

    // Formatting XML via backend converter (with pretty print true)
    formatXmlBtn.addEventListener('click', async () => {
        const val = xmlInput.value.trim();
        if (!val) return;
        
        setStatusConverting(xmlFooter);
        const params = new URLSearchParams({
            attr_prefix: attrPrefix.value || '@',
            text_key: textKey.value || '#text',
            auto_type: autoType.checked ? 'true' : 'false',
            pretty: 'true' // Force pretty format
        });

        try {
            // Round-trip to format it nicely: convert XML -> JSON -> XML
            const toJSONRes = await fetch(`/api/v1/convert/xml-to-json?${params.toString()}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/xml' },
                body: val
            });

            if (toJSONRes.ok) {
                const jsonText = await toJSONRes.text();
                const jsonParsed = JSON.parse(jsonText);
                
                // Get the main root object key
                const rootKey = Object.keys(jsonParsed)[0];
                const rootBody = jsonParsed[rootKey];
                
                const toXMLParams = new URLSearchParams({
                    root: rootKey,
                    array_mode: arrayMode.value,
                    attr_prefix: attrPrefix.value || '@',
                    pretty: 'true'
                });

                const toXMLRes = await fetch(`/api/v1/convert/json-to-xml?${toXMLParams.toString()}`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(rootBody)
                });

                if (toXMLRes.ok) {
                    const finalXml = await toXMLRes.text();
                    xmlInput.value = finalXml;
                    updateLineNumbers(xmlInput, xmlLines);
                    setStatusReady(xmlFooter);
                    // Update JSON panel too
                    jsonInput.value = jsonText;
                    updateLineNumbers(jsonInput, jsonLines);
                    setStatusReady(jsonFooter);
                } else {
                    const err = await toXMLRes.json();
                    setStatusError(xmlFooter, err.error);
                }
            } else {
                const err = await toJSONRes.json();
                setStatusError(xmlFooter, err.error);
            }
        } catch (err) {
            setStatusError(xmlFooter, "Formatting failed: " + err.message);
        }
    });

    // Clear buttons
    clearJsonBtn.addEventListener('click', () => {
        jsonInput.value = '';
        updateLineNumbers(jsonInput, jsonLines);
        setStatusReady(jsonFooter);
        if (activeEditor === 'json' || lastModifiedEditor === 'json') {
            xmlInput.value = '';
            updateLineNumbers(xmlInput, xmlLines);
            setStatusReady(xmlFooter);
        }
    });

    clearXmlBtn.addEventListener('click', () => {
        xmlInput.value = '';
        updateLineNumbers(xmlInput, xmlLines);
        setStatusReady(xmlFooter);
        if (activeEditor === 'xml' || lastModifiedEditor === 'xml') {
            jsonInput.value = '';
            updateLineNumbers(jsonInput, jsonLines);
            setStatusReady(jsonFooter);
        }
    });

    // Copy to clipboard helpers
    function copyText(textarea, button) {
        if (!textarea.value.trim()) return;
        
        navigator.clipboard.writeText(textarea.value).then(() => {
            const originalHTML = button.innerHTML;
            button.innerHTML = `<i class="fa-solid fa-check" style="color: var(--color-success)"></i>`;
            button.style.transform = 'scale(1.15)';
            
            setTimeout(() => {
                button.innerHTML = originalHTML;
                button.style.transform = 'scale(1)';
            }, 1500);
        }).catch(err => {
            console.error('Failed to copy: ', err);
        });
    }

    copyJsonBtn.addEventListener('click', () => copyText(jsonInput, copyJsonBtn));
    copyXmlBtn.addEventListener('click', () => copyText(xmlInput, copyXmlBtn));

    // --- Modal Interactivity ---
    apiDocBtn.addEventListener('click', (e) => {
        e.preventDefault();
        apiModal.classList.add('active');
    });

    closeModal.addEventListener('click', () => {
        apiModal.classList.remove('active');
    });

    // Close on backdrop click
    apiModal.addEventListener('click', (e) => {
        if (e.target === apiModal) {
            apiModal.classList.remove('active');
        }
    });

    // Close on ESC key
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && apiModal.classList.contains('active')) {
            apiModal.classList.remove('active');
        }
    });
});
