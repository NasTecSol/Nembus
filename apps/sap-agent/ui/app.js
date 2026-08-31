// Nembus SAP B1 Migration Agent Frontend Application
document.addEventListener('DOMContentLoaded', () => {
    let ws = null;
    let currentRunID = null;

    // 1. Navigation Tab Switching
    const navButtons = document.querySelectorAll('.nav-item');
    const tabPanes = document.querySelectorAll('.tab-pane');
    const pageTitle = document.getElementById('page-title');
    const pageSubtitle = document.getElementById('page-subtitle');

    const tabMeta = {
        'tab-discovery': { title: 'Discovery & Configuration', sub: 'Configure SAP SQL Server source and target Nembus Cloud ERP' },
        'tab-migration': { title: 'Migration Pipeline', sub: 'Execute entity extractions, transforms and cloud batch ingestion' },
        'tab-logs': { title: 'Live Console', sub: 'Real-time WebSocket event logs and SQLite execution checkpoints' },
        'tab-reconciliation': { title: 'Audit & Reconciliation', sub: 'Post-migration parity reports and numeric ledger verifications' },
        'tab-enrichment': { title: 'Product Enrichment Review', sub: 'Review and apply AI proposals to the Nembus tenant database' },
        'tab-history': { title: 'Run History', sub: 'Persistent SQLite log history across migration checkpoints' }
    };

    navButtons.forEach(btn => {
        btn.addEventListener('click', () => {
            const targetTab = btn.getAttribute('data-tab');
            navButtons.forEach(b => b.classList.remove('active'));
            tabPanes.forEach(p => p.classList.remove('active'));

            if (targetTab !== 'tab-enrichment') {
                clearSelectedEnrichmentDetail();
            }

            btn.classList.add('active');
            const pane = document.getElementById(targetTab);
            if (pane) pane.classList.add('active');

            if (tabMeta[targetTab]) {
                pageTitle.textContent = tabMeta[targetTab].title;
                pageSubtitle.textContent = tabMeta[targetTab].sub;
            }

            if (targetTab === 'tab-history') {
                loadRunHistory();
            }
            if (targetTab === 'tab-enrichment') {
                loadEnrichmentSuggestions();
            }
        });
    });

    // 2. Toast Notifications
    function showToast(msg, type = 'info') {
        const toast = document.getElementById('toast');
        toast.textContent = msg;
        toast.className = `toast show ${type}`;
        setTimeout(() => {
            toast.className = 'toast';
        }, 4000);
    }

    // 3. Load & Save Configurations
    async function loadConfig() {
        try {
            const res = await fetch('/api/v1/config');
            if (!res.ok) return;
            const cfg = await res.json();

            // Populate MSSQL
            if (cfg.mssql) {
                document.getElementById('mssql-host').value = cfg.mssql.host || '192.168.18.77';
                document.getElementById('mssql-port').value = cfg.mssql.port || 1433;
                document.getElementById('mssql-user').value = cfg.mssql.user || 'admin';
                document.getElementById('mssql-pass').value = cfg.mssql.password || 'nastecsol';
                document.getElementById('mssql-db').value = cfg.mssql.database || 'Qadsiya';
                document.getElementById('mssql-trust').checked = cfg.mssql.trust_server_certificate ?? true;
                document.getElementById('mssql-encrypt').checked = cfg.mssql.encrypt ?? false;
            }

            // Populate Cloud
            if (cfg.cloud) {
                document.getElementById('cloud-url').value = cfg.cloud.base_url || 'http://127.0.0.1:8080';
                document.getElementById('cloud-tenant-slug').value = cfg.cloud.tenant_slug || '';
                document.getElementById('cloud-org-id').value = cfg.cloud.organization_id || '';
            }
            if (cfg.batch_size) {
                document.getElementById('agent-batch-size').value = cfg.batch_size;
            }
        } catch (err) {
            console.error('Failed to load config:', err);
        }
    }

    // Save MSSQL Form
    document.getElementById('form-mssql').addEventListener('submit', async (e) => {
        e.preventDefault();
        const payload = {
            mssql: {
                host: document.getElementById('mssql-host').value,
                port: parseInt(document.getElementById('mssql-port').value),
                user: document.getElementById('mssql-user').value,
                password: document.getElementById('mssql-pass').value,
                database: document.getElementById('mssql-db').value,
                trust_server_certificate: document.getElementById('mssql-trust').checked,
                encrypt: document.getElementById('mssql-encrypt').checked,
            }
        };
        try {
            const res = await fetch('/api/v1/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            if (res.ok) {
                showToast('SQL Server configuration saved.', 'success');
            }
        } catch (err) {
            showToast('Failed to save configuration.', 'danger');
        }
    });

    // Save Cloud Form
    document.getElementById('form-cloud').addEventListener('submit', async (e) => {
        e.preventDefault();
        const payload = {
            cloud: {
                base_url: document.getElementById('cloud-url').value,
                tenant_slug: document.getElementById('cloud-tenant-slug').value,
                organization_id: parseInt(document.getElementById('cloud-org-id').value),
            },
            batch_size: parseInt(document.getElementById('agent-batch-size').value)
        };
        try {
            const res = await fetch('/api/v1/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            if (res.ok) {
                showToast('Cloud gateway configuration saved.', 'success');
            }
        } catch (err) {
            showToast('Failed to save configuration.', 'danger');
        }
    });

    // Test MSSQL Connection
    document.getElementById('btn-test-mssql').addEventListener('click', async () => {
        const badge = document.getElementById('mssql-badge');
        badge.textContent = 'Testing...';
        badge.className = 'badge';

        const payload = {
            host: document.getElementById('mssql-host').value,
            port: parseInt(document.getElementById('mssql-port').value) || 1433,
            user: document.getElementById('mssql-user').value,
            password: document.getElementById('mssql-pass').value,
            database: document.getElementById('mssql-db').value,
            trust_server_certificate: document.getElementById('mssql-trust').checked,
            encrypt: document.getElementById('mssql-encrypt').checked,
        };

        try {
            const res = await fetch('/api/v1/test-connection/mssql', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            const data = await res.json();
            if (data.success) {
                badge.textContent = 'Connected';
                badge.className = 'badge badge-success';
                showToast('SQL Server connection successful!', 'success');
            } else {
                badge.textContent = 'Failed';
                badge.className = 'badge badge-danger';
                showToast(data.message || 'Connection failed', 'danger');
            }
        } catch (err) {
            badge.textContent = 'Error';
            badge.className = 'badge badge-danger';
            showToast('Failed to contact agent server.', 'danger');
        }
    });

    // Test Cloud Connection
    document.getElementById('btn-test-cloud').addEventListener('click', async () => {
        const badge = document.getElementById('cloud-badge');
        badge.textContent = 'Testing...';
        badge.className = 'badge';

        const payload = {
            base_url: document.getElementById('cloud-url').value,
            tenant_slug: document.getElementById('cloud-tenant-slug').value,
            organization_id: parseInt(document.getElementById('cloud-org-id').value),
        };

        try {
            const res = await fetch('/api/v1/test-connection/cloud', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            const data = await res.json();
            if (data.success) {
                badge.textContent = 'Connected';
                badge.className = 'badge badge-success';
                showToast('Cloud gateway reachable!', 'success');
            } else {
                badge.textContent = 'Failed';
                badge.className = 'badge badge-danger';
                showToast(data.message || 'Cloud endpoint unreachable', 'danger');
            }
        } catch (err) {
            badge.textContent = 'Error';
            badge.className = 'badge badge-danger';
            showToast('Failed to contact agent server.', 'danger');
        }
    });


    // 4. Discovery Execution
    async function runDiscovery() {
        showToast('Querying SAP B1 metadata...', 'info');
        try {
            const res = await fetch('/api/v1/discovery', { method: 'POST' });
            if (!res.ok) {
                showToast('Discovery failed. Check SQL connection.', 'danger');
                return;
            }
            const data = await res.json();
            document.getElementById('disc-company').textContent = data.company_name || 'SAP B1 Company';
            document.getElementById('disc-version').textContent = data.sap_version || '-';
            document.getElementById('disc-dbname').textContent = data.database_name || '-';

            const counts = data.table_counts || {};
            document.getElementById('count-oitm').textContent = (counts['OITM'] || 0).toLocaleString();
            document.getElementById('count-owhs').textContent = (counts['OWHS'] || 0).toLocaleString();
            document.getElementById('count-oitw').textContent = (counts['OITW'] || 0).toLocaleString();
            document.getElementById('count-ocrd-c').textContent = (counts['OCRD_C'] || 0).toLocaleString();
            document.getElementById('count-ocrd-s').textContent = (counts['OCRD_S'] || 0).toLocaleString();
            document.getElementById('count-ordr').textContent = (counts['ORDR'] || 0).toLocaleString();
            document.getElementById('count-oinv').textContent = (counts['OINV'] || 0).toLocaleString();
            document.getElementById('count-ousr').textContent = ((counts['OUSR'] || 0) + (counts['OSLP'] || 0)).toLocaleString();

            showToast('SAP discovery completed successfully.', 'success');
        } catch (err) {
            showToast('Discovery failed: ' + err.message, 'danger');
        }
    }
    document.getElementById('btn-run-discovery').addEventListener('click', runDiscovery);

    // 5. WebSocket Real-Time Progress Stream
    function initWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;

        ws = new WebSocket(wsUrl);

        ws.onopen = () => {
            document.getElementById('ws-status-dot').className = 'status-dot connected';
            document.getElementById('ws-status-text').textContent = 'Agent Online';
            addLog('SYSTEM', 'WebSocket connection established with agent runtime.', 'info');
        };

        ws.onclose = () => {
            document.getElementById('ws-status-dot').className = 'status-dot';
            document.getElementById('ws-status-text').textContent = 'Disconnected';
            setTimeout(initWebSocket, 3000);
        };

        ws.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data);
                handleProgressEvent(msg);
            } catch (e) {
                console.error('Invalid WS JSON:', event.data);
            }
        };
    }

    function handleProgressEvent(ev) {
        addLog(ev.domain ? ev.domain.toUpperCase() : 'PIPELINE', ev.message, ev.status === 'failed' ? 'error' : 'info');

        if (ev.run_id) currentRunID = ev.run_id;

        // Update overall progress bar
        if (ev.percentage !== undefined) {
            const pct = Math.round(ev.percentage);
            document.getElementById('overall-progress-bar').style.width = `${pct}%`;
            document.getElementById('overall-percent-text').textContent = `${pct}%`;
        }

        if (ev.type === 'run_started') {
            document.getElementById('overall-status-text').textContent = 'Status: Migrating...';
            document.getElementById('btn-start-migration').disabled = true;
            document.getElementById('btn-cancel-migration').disabled = false;
        } else if (ev.type === 'step_started' && ev.domain) {
            const indicator = document.getElementById(`step-status-${ev.domain}`);
            const card = document.getElementById(`card-domain-${ev.domain}`);
            if (indicator) {
                indicator.textContent = 'Extracting...';
                indicator.className = 'step-indicator running';
            }
            if (card) card.className = 'domain-card running';
        } else if (ev.type === 'step_completed' && ev.domain) {
            const indicator = document.getElementById(`step-status-${ev.domain}`);
            const card = document.getElementById(`card-domain-${ev.domain}`);
            if (indicator) {
                indicator.textContent = 'Completed';
                indicator.className = 'step-indicator completed';
            }
            if (card) card.className = 'domain-card completed';
        } else if (ev.type === 'run_completed') {
            document.getElementById('overall-status-text').textContent = 'Status: Finished';
            document.getElementById('btn-start-migration').disabled = false;
            document.getElementById('btn-cancel-migration').disabled = true;
            showToast('Migration completed successfully!', 'success');
        } else if (ev.type === 'error') {
            document.getElementById('overall-status-text').textContent = 'Status: Failed';
            document.getElementById('btn-start-migration').disabled = false;
            document.getElementById('btn-cancel-migration').disabled = true;
            showToast(ev.message, 'danger');
        }
    }

    function addLog(tag, msg, level = 'info') {
        const terminal = document.getElementById('terminal-output');
        const line = document.createElement('div');
        line.className = `log-line ${level}`;
        const time = new Date().toTimeString().split(' ')[0];
        line.innerHTML = `<span class="log-time">[${time}]</span> <span class="log-tag">[${tag}]</span> ${msg}`;
        terminal.appendChild(line);

        if (document.getElementById('chk-autoscroll').checked) {
            terminal.scrollTop = terminal.scrollHeight;
        }
    }

    document.getElementById('btn-clear-logs').addEventListener('click', () => {
        document.getElementById('terminal-output').innerHTML = '';
    });

    // 6. Launch Migration
    document.getElementById('btn-start-migration').addEventListener('click', async () => {
        const selected = [];
        document.querySelectorAll('input[name="domain"]:checked').forEach(cb => {
            selected.push(cb.value);
        });

        if (selected.length === 0) {
            showToast('Please select at least one domain to migrate.', 'warning');
            return;
        }

        try {
            const res = await fetch('/api/v1/migration/start', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    mode: 'full',
                    domains: selected
                })
            });
            const data = await res.json();
            if (data.success) {
                currentRunID = data.run_id;
                showToast('Migration run launched!', 'success');
            } else {
                showToast(data.message || 'Failed to start migration', 'danger');
            }
        } catch (err) {
            showToast('Network error starting migration.', 'danger');
        }
    });

    // Cancel Migration
    document.getElementById('btn-cancel-migration').addEventListener('click', async () => {
        try {
            await fetch('/api/v1/migration/cancel', { method: 'POST' });
            showToast('Cancellation requested.', 'info');
        } catch (err) {
            console.error(err);
        }
    });

    // 7. Audit & Reconciliation
    document.getElementById('btn-run-reconciliation').addEventListener('click', async () => {
        showToast('Executing parity verification...', 'info');
        try {
            const res = await fetch('/api/v1/reconciliation', { method: 'POST' });
            if (!res.ok) {
                showToast('Reconciliation failed.', 'danger');
                return;
            }
            const rep = await res.json();
            document.getElementById('audit-summary-text').textContent = rep.audit_summary || 'Audit complete.';

            const tbody = document.getElementById('audit-table-body');
            tbody.innerHTML = '';
            (rep.domains || []).forEach(d => {
                const tr = document.createElement('tr');
                const badgeClass = d.status === 'MATCH' ? 'badge-success' : 'badge-danger';
                tr.innerHTML = `
                    <td><strong>${d.domain}</strong></td>
                    <td>${d.sap_source_count.toLocaleString()}</td>
                    <td>${d.target_count.toLocaleString()}</td>
                    <td>${d.difference}</td>
                    <td>${d.sap_numeric_sum ? d.sap_numeric_sum.toLocaleString() : '-'}</td>
                    <td>${d.target_numeric_sum ? d.target_numeric_sum.toLocaleString() : '-'}</td>
                    <td><span class="badge ${badgeClass}">${d.status}</span></td>
                `;
                tbody.appendChild(tr);
            });
            showToast('Reconciliation completed.', 'success');
        } catch (err) {
            showToast('Reconciliation failed: ' + err.message, 'danger');
        }
    });

    // 8. Run History
    async function loadRunHistory() {
        try {
            const res = await fetch('/api/v1/history');
            if (!res.ok) return;
            const data = await res.json();
            const tbody = document.getElementById('history-table-body');
            tbody.innerHTML = '';

            if (!data.runs || data.runs.length === 0) {
                tbody.innerHTML = '<tr><td colspan="7" class="text-center text-muted">No historical runs recorded.</td></tr>';
                return;
            }

            data.runs.forEach(r => {
                const tr = document.createElement('tr');
                const badgeClass = r.status === 'completed' ? 'badge-success' : (r.status === 'running' ? 'badge-accent' : 'badge-danger');
                tr.innerHTML = `
                    <td><code>${r.id.substring(0, 8)}...</code></td>
                    <td>${r.mode}</td>
                    <td><span class="badge ${badgeClass}">${r.status}</span></td>
                    <td>${r.completed_steps}/${r.total_domains}</td>
                    <td>${(r.processed_count || 0).toLocaleString()}</td>
                    <td>${new Date(r.started_at).toLocaleString()}</td>
                    <td>${r.finished_at ? new Date(r.finished_at).toLocaleString() : '-'}</td>
                `;
                tbody.appendChild(tr);
            });
        } catch (err) {
            console.error('Failed to load history:', err);
        }
    }

    document.getElementById('btn-refresh-all').addEventListener('click', () => {
        loadConfig();
        runDiscovery();
        if (document.getElementById('tab-enrichment').classList.contains('active')) {
            loadEnrichmentSuggestions();
        }
    });

    let selectedEnrichmentID = null;
    let enrichmentDetailRequestID = 0;
    let enrichmentListRequestID = 0;
    let enrichmentPage = 1;
    let enrichmentPageSize = 25;
    let enrichmentHasNextPage = false;
    const enrichmentError = document.getElementById('enrichment-error');
    const enrichmentTableBody = document.getElementById('enrichment-table-body');
    const enrichmentDetailRow = document.getElementById('enrichment-detail-row');
    const enrichmentDetail = document.getElementById('enrichment-detail');

    function clearSelectedEnrichmentDetail() {
        selectedEnrichmentID = null;
        enrichmentDetailRequestID += 1;
        enrichmentDetail.hidden = true;
        enrichmentDetailRow.hidden = true;
        enrichmentDetailRow.remove();
    }

    function setEnrichmentError(message) {
        enrichmentError.textContent = message || '';
        enrichmentError.hidden = !message;
    }

    function proposalText(proposal) {
        if (!proposal) return '—';
        if (proposal.canonical_name) return `${proposal.canonical_name} (${proposal.action})`;
        if (proposal.value) return `${proposal.value} (${proposal.action})`;
        return proposal.action || '—';
    }

    function renderEnrichmentRows(items) {
        const selectedID = selectedEnrichmentID;
        enrichmentTableBody.textContent = '';
        if (!items || items.length === 0) {
            if (selectedID !== null) clearSelectedEnrichmentDetail();
            const row = document.createElement('tr');
            const cell = document.createElement('td');
            cell.colSpan = 6;
            cell.className = 'text-center text-muted';
            cell.textContent = 'No suggestions for this status.';
            row.appendChild(cell);
            enrichmentTableBody.appendChild(row);
            return;
        }
        items.forEach(item => {
            const row = document.createElement('tr');
            row.className = 'review-row';
            row.dataset.suggestionId = String(item.suggestion_id);
            const values = [
                `${item.source_item_code || '—'} / ${item.source_item_name || '—'}`,
                'Open detail to compare',
                [proposalText(item.proposed_brand), proposalText(item.proposed_category), proposalText(item.proposed_description)].filter(Boolean).join(' · ') || '—',
                [item.provider, item.model].filter(Boolean).join(' / ') || '—',
                item.status || '—'
            ];
            values.forEach(value => {
                const cell = document.createElement('td');
                cell.textContent = value;
                row.appendChild(cell);
            });
            const actionCell = document.createElement('td');
            const open = document.createElement('button');
            open.className = 'btn btn-xs btn-outline';
            open.textContent = 'View';
            open.addEventListener('click', () => loadEnrichmentDetail(item.suggestion_id));
            actionCell.appendChild(open);
            row.appendChild(actionCell);
            enrichmentTableBody.appendChild(row);
        });

        if (selectedID !== null) {
            const selectedRow = Array.from(enrichmentTableBody.querySelectorAll('.review-row'))
                .find(row => row.dataset.suggestionId === String(selectedID));
            if (selectedRow) {
                selectedRow.after(enrichmentDetailRow);
                enrichmentDetailRow.hidden = false;
                enrichmentDetail.hidden = false;
            } else {
                clearSelectedEnrichmentDetail();
            }
        }
    }

    function updateEnrichmentPagination() {
        document.getElementById('enrichment-page-indicator').textContent = `Page ${enrichmentPage}`;
        document.getElementById('btn-enrichment-previous').disabled = enrichmentPage <= 1;
        document.getElementById('btn-enrichment-next').disabled = !enrichmentHasNextPage;
    }

    async function loadEnrichmentSuggestions({ allowEmptyPageFallback = true } = {}) {
        const requestID = ++enrichmentListRequestID;
        setEnrichmentError('');
        const status = document.getElementById('enrichment-status-filter').value;
        const requestedPage = enrichmentPage;
        const requestedPageSize = enrichmentPageSize;
        const offset = (requestedPage - 1) * requestedPageSize;
        try {
            // The review API has no total/count. Ask for one extra item to
            // determine whether another server-side page exists.
            const res = await fetch(`/api/v1/product-enrichment/suggestions?status=${encodeURIComponent(status)}&limit=${requestedPageSize + 1}&offset=${offset}`);
            const payload = await res.json();
            if (!res.ok) throw new Error(payload.message || 'Could not load enrichment suggestions.');
            if (requestID !== enrichmentListRequestID) return;
            const returnedItems = Array.isArray(payload.data?.items) ? payload.data.items : [];
            if (allowEmptyPageFallback && returnedItems.length === 0 && requestedPage > 1) {
                enrichmentPage = requestedPage - 1;
                return loadEnrichmentSuggestions({ allowEmptyPageFallback: false });
            }
            enrichmentHasNextPage = returnedItems.length > requestedPageSize;
            renderEnrichmentRows(returnedItems.slice(0, requestedPageSize));
            updateEnrichmentPagination();
        } catch (err) {
            if (requestID !== enrichmentListRequestID) return;
            enrichmentHasNextPage = false;
            renderEnrichmentRows([]);
            updateEnrichmentPagination();
            setEnrichmentError(err.message || 'Could not load enrichment suggestions.');
        }
    }

    function renderEnrichmentDetail(id, detail) {
        const missingValue = '\u2014';
        const separator = '\u00b7';
        const selectedRow = Array.from(enrichmentTableBody.querySelectorAll('.review-row'))
            .find(row => row.dataset.suggestionId === String(id));
        if (!selectedRow) return;
        selectedEnrichmentID = id;
        selectedRow.after(enrichmentDetailRow);
        enrichmentDetailRow.hidden = false;
        enrichmentDetail.hidden = false;
        document.getElementById('enrichment-detail-title').textContent = `${detail.source_identity?.source_item_code || 'Suggestion'} ${missingValue} ${detail.source_identity?.source_item_name || ''}`;
        document.getElementById('enrichment-detail-status').textContent = detail.review_state?.status || missingValue;
        document.getElementById('enrichment-current').textContent = JSON.stringify(detail.current_authoritative_state || {}, null, 2);
        document.getElementById('enrichment-proposed').textContent = JSON.stringify({
            brand: detail.proposed_brand || null,
            category: detail.proposed_category || null,
            description: detail.proposed_description || null,
            unsupported_semantics: detail.unsupported_semantics || []
        }, null, 2);
        document.getElementById('enrichment-detail-meta').textContent = `Provider: ${detail.provider_context?.provider || missingValue} ${separator} Model: ${detail.provider_context?.model || missingValue} ${separator} Confidence is shown per proposal.`;
        const actionable = detail.review_state?.status === 'in_review';
        document.getElementById('enrichment-detail-actions').hidden = !actionable;
    }

    async function loadEnrichmentDetail(id) {
        if (selectedEnrichmentID !== null && String(selectedEnrichmentID) !== String(id)) {
            clearSelectedEnrichmentDetail();
        }
        const requestID = ++enrichmentDetailRequestID;
        setEnrichmentError('');
        try {
            const res = await fetch(`/api/v1/product-enrichment/suggestions/${encodeURIComponent(id)}`);
            const payload = await res.json();
            if (!res.ok) throw new Error(payload.message || 'Could not load suggestion detail.');
            const detail = payload.data || {};
            if (requestID !== enrichmentDetailRequestID) return;
            renderEnrichmentDetail(id, detail);
            return;
        } catch (err) {
            if (requestID !== enrichmentDetailRequestID) return;
            setEnrichmentError(err.message || 'Could not load suggestion detail.');
        }
    }

    async function decideEnrichment(action) {
        if (!selectedEnrichmentID) return;
        const decidedSuggestionID = selectedEnrichmentID;
        try {
            const res = await fetch(`/api/v1/product-enrichment/suggestions/${encodeURIComponent(decidedSuggestionID)}/${action}`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: '{}' });
            const payload = await res.json();
            if (!res.ok) throw new Error(payload.message || `Could not ${action} suggestion.`);
            showToast(action === 'approve' ? 'Suggestion approved and applied to Nembus.' : 'Suggestion rejected.', 'success');
            if (selectedEnrichmentID === decidedSuggestionID) clearSelectedEnrichmentDetail();
            loadEnrichmentSuggestions();
        } catch (err) {
            setEnrichmentError(err.message || `Could not ${action} suggestion.`);
        }
    }

    document.getElementById('btn-refresh-enrichment').addEventListener('click', loadEnrichmentSuggestions);
    document.getElementById('enrichment-status-filter').addEventListener('change', () => {
        enrichmentPage = 1;
        clearSelectedEnrichmentDetail();
        loadEnrichmentSuggestions();
    });
    document.getElementById('enrichment-page-size').addEventListener('change', event => {
        enrichmentPageSize = Number(event.target.value) || 25;
        enrichmentPage = 1;
        clearSelectedEnrichmentDetail();
        loadEnrichmentSuggestions();
    });
    document.getElementById('btn-enrichment-previous').addEventListener('click', () => {
        if (enrichmentPage <= 1) return;
        enrichmentPage -= 1;
        clearSelectedEnrichmentDetail();
        loadEnrichmentSuggestions();
    });
    document.getElementById('btn-enrichment-next').addEventListener('click', () => {
        if (!enrichmentHasNextPage) return;
        enrichmentPage += 1;
        clearSelectedEnrichmentDetail();
        loadEnrichmentSuggestions();
    });
    document.getElementById('btn-close-enrichment-detail').addEventListener('click', clearSelectedEnrichmentDetail);
    document.getElementById('btn-approve-enrichment').addEventListener('click', () => decideEnrichment('approve'));
    document.getElementById('btn-reject-enrichment').addEventListener('click', () => decideEnrichment('reject'));

    // Init
    loadConfig();
    initWebSocket();
});
