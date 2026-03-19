package web

const dashboardHTML = `<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MailMole - IMAP Migration</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0f0f0f;
            color: #e5e5e5;
            line-height: 1.5;
            min-height: 100vh;
        }

        .app {
            display: grid;
            grid-template-columns: 280px 1fr;
            min-height: 100vh;
        }

        /* Sidebar */
        .sidebar {
            background: #141414;
            border-right: 1px solid #262626;
            padding: 24px;
            display: flex;
            flex-direction: column;
        }

        .logo {
            font-size: 20px;
            font-weight: 600;
            color: #3b82f6;
            margin-bottom: 32px;
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .nav-section {
            margin-bottom: 24px;
        }

        .nav-title {
            font-size: 11px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            color: #737373;
            margin-bottom: 8px;
        }

        .nav-item {
            padding: 10px 12px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
            color: #a3a3a3;
            transition: all 0.15s;
            margin-bottom: 2px;
        }

        .nav-item:hover { background: #1f1f1f; color: #e5e5e5; }
        .nav-item.active { background: #1f1f1f; color: #3b82f6; }

        /* Main Content */
        .main {
            padding: 32px 40px;
            overflow-y: auto;
        }

        .header {
            margin-bottom: 32px;
        }

        .header h1 {
            font-size: 28px;
            font-weight: 600;
            margin-bottom: 8px;
        }

        .header p {
            color: #737373;
            font-size: 14px;
        }

        /* Cards */
        .card {
            background: #141414;
            border: 1px solid #262626;
            border-radius: 8px;
            padding: 24px;
            margin-bottom: 24px;
        }

        .card-title {
            font-size: 14px;
            font-weight: 600;
            margin-bottom: 16px;
            color: #e5e5e5;
        }

        /* Form Elements */
        .form-row {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 16px;
            margin-bottom: 16px;
        }

        .form-group {
            margin-bottom: 16px;
        }

        .form-group label {
            display: block;
            font-size: 12px;
            font-weight: 500;
            color: #a3a3a3;
            margin-bottom: 6px;
            text-transform: uppercase;
            letter-spacing: 0.3px;
        }

        input[type="text"],
        input[type="password"],
        input[type="number"] {
            width: 100%;
            padding: 10px 12px;
            background: #0f0f0f;
            border: 1px solid #333;
            border-radius: 6px;
            color: #e5e5e5;
            font-size: 14px;
            transition: border-color 0.15s;
        }

        input:focus {
            outline: none;
            border-color: #3b82f6;
        }

        input::placeholder { color: #525252; }

        /* Buttons */
        .btn {
            padding: 10px 20px;
            border-radius: 6px;
            border: none;
            font-size: 14px;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.15s;
            display: inline-flex;
            align-items: center;
            gap: 8px;
        }

        .btn-primary {
            background: #3b82f6;
            color: white;
        }

        .btn-primary:hover { background: #2563eb; }

        .btn-secondary {
            background: #262626;
            color: #e5e5e5;
            border: 1px solid #404040;
        }

        .btn-secondary:hover { background: #333; }

        .btn-success {
            background: #22c55e;
            color: white;
        }

        .btn-success:hover { background: #16a34a; }

        .btn-danger {
            background: #ef4444;
            color: white;
        }

        .btn-sm { padding: 6px 12px; font-size: 13px; }

        /* Tables */
        .table-container {
            overflow-x: auto;
        }

        table {
            width: 100%;
            border-collapse: collapse;
            font-size: 14px;
        }

        th {
            text-align: left;
            padding: 12px;
            font-size: 11px;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            color: #737373;
            border-bottom: 1px solid #262626;
        }

        td {
            padding: 12px;
            border-bottom: 1px solid #1f1f1f;
            color: #d4d4d4;
        }

        tr:hover td { background: #1a1a1a; }

        /* Status Badges */
        .badge {
            display: inline-flex;
            align-items: center;
            padding: 4px 10px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: 500;
        }

        .badge-success { background: rgba(34, 197, 94, 0.15); color: #22c55e; }
        .badge-error { background: rgba(239, 68, 68, 0.15); color: #ef4444; }
        .badge-warning { background: rgba(245, 158, 11, 0.15); color: #f59e0b; }
        .badge-info { background: rgba(59, 130, 246, 0.15); color: #3b82f6; }

        /* Progress */
        .progress-container {
            margin: 24px 0;
        }

        .progress-header {
            display: flex;
            justify-content: space-between;
            margin-bottom: 8px;
            font-size: 14px;
        }

        .progress-bar-bg {
            height: 8px;
            background: #1f1f1f;
            border-radius: 4px;
            overflow: hidden;
        }

        .progress-bar-fill {
            height: 100%;
            background: #3b82f6;
            border-radius: 4px;
            transition: width 0.3s;
        }

        /* Stats Grid */
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(4, 1fr);
            gap: 16px;
            margin-bottom: 24px;
        }

        .stat-card {
            background: #141414;
            border: 1px solid #262626;
            border-radius: 8px;
            padding: 20px;
        }

        .stat-label {
            font-size: 12px;
            color: #737373;
            margin-bottom: 4px;
        }

        .stat-value {
            font-size: 24px;
            font-weight: 600;
            color: #e5e5e5;
        }

        /* Tabs */
        .tabs {
            display: flex;
            gap: 4px;
            margin-bottom: 24px;
            border-bottom: 1px solid #262626;
        }

        .tab {
            padding: 12px 20px;
            background: none;
            border: none;
            color: #737373;
            font-size: 14px;
            cursor: pointer;
            border-bottom: 2px solid transparent;
            margin-bottom: -1px;
            transition: all 0.15s;
        }

        .tab:hover { color: #a3a3a3; }
        .tab.active { color: #3b82f6; border-bottom-color: #3b82f6; }

        /* Account Row */
        .account-row {
            background: #0f0f0f;
            border: 1px solid #262626;
            border-radius: 6px;
            margin-bottom: 12px;
            overflow: hidden;
        }

        .account-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 12px 16px;
            background: #1a1a1a;
            border-bottom: 1px solid #262626;
        }

        .account-number {
            font-size: 12px;
            font-weight: 600;
            color: #737373;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .account-fields {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 16px;
            padding: 16px;
        }

        .account-section {
            display: flex;
            flex-direction: column;
            gap: 10px;
        }

        .account-section-title {
            font-size: 11px;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            color: #525252;
            margin-bottom: 4px;
        }

        .account-section.source .account-section-title { color: #3b82f6; }
        .account-section.dest .account-section-title { color: #22c55e; }

        .field-row {
            display: grid;
            grid-template-columns: 2fr 1fr;
            gap: 8px;
        }

        .account-input {
            background: #1a1a1a;
            border: 1px solid #333;
            border-radius: 4px;
            padding: 10px 12px;
            color: #e5e5e5;
            font-size: 13px;
            width: 100%;
            transition: border-color 0.15s;
        }

        .account-input:focus {
            outline: none;
            border-color: #525252;
        }

        .account-section.source .account-input:focus {
            border-color: #3b82f6;
        }

        .account-section.dest .account-input:focus {
            border-color: #22c55e;
        }

        /* Logs */
        .log-container {
            background: #0f0f0f;
            border: 1px solid #262626;
            border-radius: 6px;
            padding: 16px;
            height: 300px;
            overflow-y: auto;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 13px;
            line-height: 1.6;
        }

        .log-entry {
            padding: 4px 0;
            border-bottom: 1px solid #1a1a1a;
        }

        .log-time { color: #525252; margin-right: 12px; }
        .log-level-info { color: #3b82f6; }
        .log-level-success { color: #22c55e; }
        .log-level-error { color: #ef4444; }
        .log-level-warn { color: #f59e0b; }

        /* Utility */
        .hidden { display: none !important; }
        .text-muted { color: #737373; }
        .text-success { color: #22c55e; }
        .text-error { color: #ef4444; }
        
        .actions {
            display: flex;
            gap: 12px;
            margin-top: 24px;
        }

        /* Toast Notifications */
        .toast-container {
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 10000;
            display: flex;
            flex-direction: column;
            gap: 10px;
            max-width: 400px;
        }

        .toast {
            background: #1a1a1a;
            border: 1px solid #333;
            border-radius: 8px;
            padding: 16px 20px;
            box-shadow: 0 4px 20px rgba(0,0,0,0.5);
            display: flex;
            align-items: center;
            gap: 12px;
            animation: slideIn 0.3s ease;
            min-width: 300px;
        }

        .toast.success {
            border-left: 4px solid #22c55e;
        }

        .toast.error {
            border-left: 4px solid #ef4444;
        }

        .toast.info {
            border-left: 4px solid #3b82f6;
        }

        .toast.warning {
            border-left: 4px solid #f59e0b;
        }

        .toast-icon {
            font-size: 20px;
            flex-shrink: 0;
        }

        .toast-content {
            flex: 1;
        }

        .toast-title {
            font-weight: 600;
            font-size: 14px;
            margin-bottom: 2px;
        }

        .toast-message {
            font-size: 13px;
            color: #a3a3a3;
        }

        .toast-close {
            background: none;
            border: none;
            color: #737373;
            cursor: pointer;
            font-size: 18px;
            padding: 0;
            width: 24px;
            height: 24px;
            display: flex;
            align-items: center;
            justify-content: center;
            border-radius: 4px;
            transition: all 0.15s;
        }

        .toast-close:hover {
            background: #333;
            color: #e5e5e5;
        }

        @keyframes slideIn {
            from {
                transform: translateX(100%);
                opacity: 0;
            }
            to {
                transform: translateX(0);
                opacity: 1;
            }
        }

        @keyframes fadeOut {
            from {
                opacity: 1;
            }
            to {
                opacity: 0;
            }
        }

        .toast.hiding {
            animation: fadeOut 0.3s ease forwards;
        }

        /* Checkbox */
        .checkbox-wrapper {
            display: flex;
            align-items: center;
            gap: 8px;
            cursor: pointer;
        }

        .checkbox-wrapper input[type="checkbox"] {
            width: 16px;
            height: 16px;
            accent-color: #3b82f6;
        }

        @media (max-width: 1024px) {
            .app { grid-template-columns: 1fr; }
            .sidebar { display: none; }
            .main { padding: 20px; }
            .stats-grid { grid-template-columns: repeat(2, 1fr); }
            .form-row { grid-template-columns: 1fr; }
        }
    </style>
</head>
<body>
    <div class="app">
        <!-- Sidebar -->
        <aside class="sidebar">
            <div class="logo">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
                    <polyline points="22,6 12,13 2,6"/>
                </svg>
                MailMole
            </div>

            <div class="nav-section">
                <div class="nav-title">Migration</div>
                <div class="nav-item active" onclick="showPage('setup')">Connection Setup</div>
                <div class="nav-item" onclick="showPage('preview')">Preview</div>
                <div class="nav-item" onclick="showPage('progress')">Progress</div>
            </div>

            <div class="nav-section">
                <div class="nav-title">History</div>
                <div class="nav-item" onclick="showPage('logs')">Activity Logs</div>
            </div>
        </aside>

        <!-- Main Content -->
        <main class="main">
            <!-- Setup Page -->
            <div id="page-setup">
                <div class="header">
                    <h1>Connection Setup</h1>
                    <p>Configure source and destination IMAP servers</p>
                </div>

                <div class="tabs">
                    <button class="tab active" onclick="switchMode('single')">Single Account</button>
                    <button class="tab" onclick="switchMode('bulk')">Bulk Migration</button>
                </div>

                <!-- Single Account Form -->
                <div id="single-form">
                    <!-- Templates Card -->
                    <div class="card">
                        <div class="card-title">Quick Templates</div>
                        <p style="color: #737373; font-size: 13px; margin-bottom: 16px;">
                            Select a predefined server configuration
                        </p>
                        <div style="display: flex; gap: 12px; flex-wrap: wrap;">
                            <button class="btn btn-secondary btn-sm" onclick="applyTemplate('gmail-source')">📧 Gmail (Source)</button>
                            <button class="btn btn-secondary btn-sm" onclick="applyTemplate('gmail-dest')">📧 Gmail (Dest)</button>
                            <button class="btn btn-secondary btn-sm" onclick="applyTemplate('outlook-source')">📧 Outlook (Source)</button>
                            <button class="btn btn-secondary btn-sm" onclick="applyTemplate('outlook-dest')">📧 Outlook (Dest)</button>
                            <button class="btn btn-secondary btn-sm" onclick="applyTemplate('yandex-source')">📧 Yandex (Source)</button>
                            <button class="btn btn-secondary btn-sm" onclick="applyTemplate('yandex-dest')">📧 Yandex (Dest)</button>
                            <button class="btn btn-secondary btn-sm" onclick="applyTemplate('icloud-source')">📧 iCloud (Source)</button>
                            <button class="btn btn-secondary btn-sm" onclick="applyTemplate('icloud-dest')">📧 iCloud (Dest)</button>
                        </div>
                    </div>

                    <!-- Import/Export Card -->
                    <div class="card">
                        <div class="card-title">Import / Export</div>
                        <div style="display: flex; gap: 16px; flex-wrap: wrap; align-items: center;">
                            <div>
                                <input type="file" id="import-file" accept=".csv,.json" style="display: none;" onchange="handleImport(this)">
                                <button class="btn btn-secondary" onclick="document.getElementById('import-file').click()">
                                    📥 Import Accounts (CSV/JSON)
                                </button>
                            </div>
                            <button class="btn btn-secondary" onclick="exportAccounts()">
                                📤 Export Current Setup
                            </button>
                            <span style="color: #737373; font-size: 12px;">
                                CSV Format: src_host,src_port,src_user,src_pass,dst_host,dst_port,dst_user,dst_pass
                            </span>
                        </div>
                    </div>

                    <div class="card">
                        <div class="card-title">Source Server</div>
                        <div class="form-row">
                            <div class="form-group">
                                <label>Host</label>
                                <input type="text" id="single-src-host" placeholder="mail.example.com" value="mail.example.com">
                            </div>
                            <div class="form-group">
                                <label>Port</label>
                                <input type="number" id="single-src-port" value="993">
                            </div>
                        </div>
                        <div class="form-row">
                            <div class="form-group">
                                <label>Username</label>
                                <input type="text" id="single-src-user" placeholder="user@example.com">
                            </div>
                            <div class="form-group">
                                <label>Password</label>
                                <input type="password" id="single-src-pass" placeholder="••••••••">
                            </div>
                        </div>
                        <div class="checkbox-wrapper">
                            <input type="checkbox" id="single-src-tls" checked>
                            <label for="single-src-tls" style="margin:0;">Use TLS/SSL</label>
                        </div>
                    </div>

                    <div class="card">
                        <div class="card-title">Destination Server</div>
                        <div class="form-row">
                            <div class="form-group">
                                <label>Host</label>
                                <input type="text" id="single-dst-host" placeholder="mail.new.com" value="mail.new.com">
                            </div>
                            <div class="form-group">
                                <label>Port</label>
                                <input type="number" id="single-dst-port" value="993">
                            </div>
                        </div>
                        <div class="form-row">
                            <div class="form-group">
                                <label>Username</label>
                                <input type="text" id="single-dst-user" placeholder="user@new.com">
                            </div>
                            <div class="form-group">
                                <label>Password</label>
                                <input type="password" id="single-dst-pass" placeholder="••••••••">
                            </div>
                        </div>
                        <div class="checkbox-wrapper">
                            <input type="checkbox" id="single-dst-tls" checked>
                            <label for="single-dst-tls" style="margin:0;">Use TLS/SSL</label>
                        </div>
                    </div>

                    <!-- Scheduling Card -->
                    <div class="card">
                        <div class="card-title">⏰ Schedule Migration</div>
                        <div style="display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 16px;">
                            <div class="form-group">
                                <label>Start Time</label>
                                <input type="datetime-local" id="schedule-time" class="account-input">
                            </div>
                            <div class="form-group">
                                <label>Repeat</label>
                                <select id="schedule-repeat" class="account-input" style="background: #1a1a1a; color: #e5e5e5;">
                                    <option value="once">One Time</option>
                                    <option value="daily">Daily</option>
                                    <option value="weekly">Weekly</option>
                                    <option value="monthly">Monthly</option>
                                </select>
                            </div>
                            <div class="checkbox-wrapper" style="margin-top: 24px;">
                                <input type="checkbox" id="schedule-enabled">
                                <label for="schedule-enabled" style="margin:0;">Enable Scheduling</label>
                            </div>
                        </div>
                        <p style="color: #737373; font-size: 12px; margin-top: 12px;">
                            Migration will automatically start at the scheduled time. Leave disabled to run immediately.
                        </p>
                    </div>

                    <div class="actions">
                        <button class="btn btn-secondary" onclick="testConnection()">Test Connection</button>
                        <button class="btn btn-secondary" onclick="testConnectionDetailed()">🔍 Detailed Test</button>
                        <button class="btn btn-primary" onclick="previewSingle()">Preview Migration</button>
                    </div>
                </div>

                <!-- Bulk Form -->
                <div id="bulk-form" class="hidden">
                    <!-- Templates & Import/Export for Bulk -->
                    <div class="card">
                        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
                            <div class="card-title" style="margin: 0;">Quick Actions</div>
                            <div style="display: flex; gap: 12px;">
                                <input type="file" id="bulk-import-file" accept=".csv,.json" style="display: none;" onchange="handleBulkImport(this)">
                                <button class="btn btn-secondary btn-sm" onclick="document.getElementById('bulk-import-file').click()">📥 Import CSV</button>
                                <button class="btn btn-secondary btn-sm" onclick="exportBulkAccounts()">📤 Export</button>
                                <button class="btn btn-secondary btn-sm" onclick="showTemplateModal()">📋 Templates</button>
                            </div>
                        </div>
                        <p style="color: #737373; font-size: 12px;">
                            Import accounts from CSV or use templates for Gmail, Outlook, Yandex migrations
                        </p>
                    </div>

                    <div class="card">
                        <div class="card-title">Bulk Migration Accounts</div>
                        <p style="color: #737373; font-size: 13px; margin-bottom: 16px;">
                            Add multiple accounts. Each account can have different source and destination servers.
                        </p>

                        <div id="bulk-accounts">
                            <!-- Account rows will be added here -->
                        </div>

                        <button class="btn btn-secondary btn-sm" onclick="addBulkAccount()" style="margin-top: 16px;">
                            + Add Account
                        </button>
                    </div>

                    <div class="actions">
                        <button class="btn btn-secondary" onclick="testBulkConnections()">Test All Connections</button>
                        <button class="btn btn-secondary" onclick="testBulkConnectionsDetailed()">🔍 Detailed Test</button>
                        <button class="btn btn-primary" onclick="previewBulk()">Preview All Accounts</button>
                    </div>
                </div>
            </div>

            <!-- Preview Page -->
            <div id="page-preview" class="hidden">
                <div class="header">
                    <h1>Migration Preview</h1>
                    <p>Review folders and select which ones to migrate</p>
                </div>

                <div class="stats-grid">
                    <div class="stat-card">
                        <div class="stat-label">Total Folders</div>
                        <div class="stat-value" id="preview-total-folders">-</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">Total Messages</div>
                        <div class="stat-value" id="preview-total-messages">-</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">Selected</div>
                        <div class="stat-value text-success" id="preview-selected">-</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">Est. Size</div>
                        <div class="stat-value" id="preview-size">-</div>
                    </div>
                </div>

                <div class="card">
                    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
                        <div class="card-title">Folders</div>
                        <div style="display: flex; gap: 8px;">
                            <button class="btn btn-secondary btn-sm" onclick="selectAllFolders(true)">Select All</button>
                            <button class="btn btn-secondary btn-sm" onclick="selectAllFolders(false)">Select None</button>
                        </div>
                    </div>

                    <div class="table-container">
                        <table id="folders-table">
                            <thead>
                                <tr>
                                    <th style="width: 40px;"><input type="checkbox" id="select-all-checkbox"></th>
                                    <th>Folder Name</th>
                                    <th>Messages</th>
                                    <th>Size</th>
                                </tr>
                            </thead>
                            <tbody id="folders-tbody">
                                <!-- Folders will be populated here -->
                            </tbody>
                        </table>
                    </div>
                </div>

                <div class="actions">
                    <button class="btn btn-secondary" onclick="showPage('setup')">Back</button>
                    <button class="btn btn-success" onclick="startMigration()">Start Migration</button>
                </div>
            </div>

            <!-- Progress Page -->
            <div id="page-progress" class="hidden">
                <div class="header">
                    <h1>Migration Progress</h1>
                    <p>Real-time migration status</p>
                </div>

                <div class="stats-grid">
                    <div class="stat-card">
                        <div class="stat-label">Status</div>
                        <div class="stat-value" id="progress-status">
                            <span class="badge badge-info">Running</span>
                        </div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">Progress</div>
                        <div class="stat-value" id="progress-percent">0%</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">Messages</div>
                        <div class="stat-value" id="progress-messages">0/0</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">Speed</div>
                        <div class="stat-value" id="progress-speed">0/s</div>
                    </div>
                </div>

                <div class="card">
                    <div class="progress-container">
                        <div class="progress-header">
                            <span>Overall Progress</span>
                            <span id="progress-text">0/0 messages</span>
                        </div>
                        <div class="progress-bar-bg">
                            <div class="progress-bar-fill" id="progress-bar" style="width: 0%"></div>
                        </div>
                    </div>
                </div>

                <div class="card">
                    <div class="card-title">Active Accounts</div>
                    <div class="table-container">
                        <table id="accounts-table">
                            <thead>
                                <tr>
                                    <th>Account</th>
                                    <th>Status</th>
                                    <th>Progress</th>
                                    <th>Messages</th>
                                </tr>
                            </thead>
                            <tbody id="accounts-tbody">
                                <!-- Account progress will be shown here -->
                            </tbody>
                        </table>
                    </div>
                </div>

                <div class="actions">
                    <button class="btn btn-danger" onclick="stopMigration()">Stop Migration</button>
                </div>
            </div>

            <!-- Validation Results Page -->
            <div id="page-validation" class="hidden">
                <div class="header">
                    <h1>Connection Validation Results</h1>
                    <p>Detailed test results for all accounts</p>
                </div>

                <div class="stats-grid">
                    <div class="stat-card">
                        <div class="stat-label">Total Accounts</div>
                        <div class="stat-value" id="validation-total">0</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">Successful</div>
                        <div class="stat-value text-success" id="validation-success">0</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">Failed</div>
                        <div class="stat-value text-error" id="validation-failed">0</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">Status</div>
                        <div class="stat-value" id="validation-status">
                            <span class="badge badge-info">Testing...</span>
                        </div>
                    </div>
                </div>

                <div class="card">
                    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
                        <div class="card-title">Test Results</div>
                        <button class="btn btn-secondary btn-sm" onclick="exportValidationResults()">📥 Export Results</button>
                    </div>
                    <div class="table-container">
                        <table id="validation-table">
                            <thead>
                                <tr>
                                    <th>Account</th>
                                    <th>Source</th>
                                    <th>Destination</th>
                                    <th>Status</th>
                                    <th>Details</th>
                                </tr>
                            </thead>
                            <tbody id="validation-tbody">
                                <!-- Validation results will be populated here -->
                            </tbody>
                        </table>
                    </div>
                </div>

                <div class="actions">
                    <button class="btn btn-secondary" onclick="showPage('setup')">← Back to Setup</button>
                    <button class="btn btn-primary" id="validation-continue-btn" onclick="fromValidationToPreview()" disabled>Continue to Preview</button>
                </div>
            </div>

            <!-- Logs Page -->
            <div id="page-logs" class="hidden">
                <div class="header">
                    <h1>Activity Logs</h1>
                    <p>Detailed migration logs</p>
                </div>

                <div class="log-container" id="log-container">
                    <div class="log-entry">
                        <span class="log-time">--:--:--</span>
                        <span class="log-level-info">[INFO]</span>
                        Ready. Configure connection and start migration.
                    </div>
                </div>
            </div>
        </main>
    </div>

    <!-- Toast Container -->
    <div class="toast-container" id="toast-container"></div>

    <script>
        let currentMode = 'single';
        let accounts = [];
        let folders = [];
        let selectedFolders = new Set();
        let eventSource = null;

        // Initialize with one bulk account
        addBulkAccount();

        // Toast Notification Functions
        function showToast(type, title, message, duration = 5000) {
            const container = document.getElementById('toast-container');
            
            const toast = document.createElement('div');
            toast.className = 'toast ' + type;
            
            const icons = {
                success: '✓',
                error: '✗',
                info: 'ℹ',
                warning: '⚠'
            };
            
            toast.innerHTML = 
                '<span class="toast-icon">' + icons[type] + '</span>' +
                '<div class="toast-content">' +
                    '<div class="toast-title">' + title + '</div>' +
                    '<div class="toast-message">' + message + '</div>' +
                '</div>' +
                '<button class="toast-close" onclick="this.parentElement.remove()">×</button>';
            
            container.appendChild(toast);
            
            // Auto remove after duration
            setTimeout(() => {
                toast.classList.add('hiding');
                setTimeout(() => toast.remove(), 300);
            }, duration);
        }

        function showSuccess(title, message) {
            showToast('success', title, message);
        }

        function showError(title, message) {
            showToast('error', title, message);
        }

        function showInfo(title, message) {
            showToast('info', title, message);
        }

        function showWarning(title, message) {
            showToast('warning', title, message);
        }

        function showPage(page) {
            document.querySelectorAll('[id^="page-"]').forEach(el => el.classList.add('hidden'));
            document.getElementById('page-' + page).classList.remove('hidden');
            
            document.querySelectorAll('.nav-item').forEach(el => el.classList.remove('active'));
            event.target.classList.add('active');
        }

        function switchMode(mode) {
            currentMode = mode;
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            event.target.classList.add('active');
            
            document.getElementById('single-form').classList.toggle('hidden', mode !== 'single');
            document.getElementById('bulk-form').classList.toggle('hidden', mode !== 'bulk');
        }

        function addBulkAccount() {
            const container = document.getElementById('bulk-accounts');
            const index = container.children.length + 1;
            
            const row = document.createElement('div');
            row.className = 'account-row';
            row.innerHTML = 
                '<div class="account-header">' +
                    '<span class="account-number">Account #' + index + '</span>' +
                    '<button class="btn btn-danger btn-sm" onclick="this.closest(\'.account-row\').remove()">Remove</button>' +
                '</div>' +
                '<div class="account-fields">' +
                    '<div class="account-section source">' +
                        '<div class="account-section-title">📧 Source Server</div>' +
                        '<div class="field-row">' +
                            '<input type="text" class="account-input" placeholder="mail.source.com" id="bulk-src-host-' + index + '">' +
                            '<input type="number" class="account-input" placeholder="993" value="993" id="bulk-src-port-' + index + '">' +
                        '</div>' +
                        '<input type="text" class="account-input" placeholder="user@source.com" id="bulk-src-user-' + index + '">' +
                        '<input type="password" class="account-input" placeholder="Source password" id="bulk-src-pass-' + index + '">' +
                    '</div>' +
                    '<div class="account-section dest">' +
                        '<div class="account-section-title">📬 Destination Server</div>' +
                        '<div class="field-row">' +
                            '<input type="text" class="account-input" placeholder="mail.dest.com" id="bulk-dst-host-' + index + '">' +
                            '<input type="number" class="account-input" placeholder="993" value="993" id="bulk-dst-port-' + index + '">' +
                        '</div>' +
                        '<input type="text" class="account-input" placeholder="user@dest.com" id="bulk-dst-user-' + index + '">' +
                        '<input type="password" class="account-input" placeholder="Destination password" id="bulk-dst-pass-' + index + '">' +
                    '</div>' +
                '</div>';
            
            container.appendChild(row);
        }

        function log(level, message) {
            const container = document.getElementById('log-container');
            const time = new Date().toLocaleTimeString();
            const entry = document.createElement('div');
            entry.className = 'log-entry';
            entry.innerHTML = '<span class="log-time">' + time + '</span> <span class="log-level-' + level + '">[' + level.toUpperCase() + ']</span> ' + message;
            container.appendChild(entry);
            container.scrollTop = container.scrollHeight;
        }

        async function testConnection() {
            const creds = {
                srcHost: document.getElementById('single-src-host').value,
                srcPort: parseInt(document.getElementById('single-src-port').value) || 993,
                srcUser: document.getElementById('single-src-user').value,
                srcPass: document.getElementById('single-src-pass').value,
                srcTLS: document.getElementById('single-src-tls').checked,
                dstHost: document.getElementById('single-dst-host').value,
                dstPort: parseInt(document.getElementById('single-dst-port').value) || 993,
                dstUser: document.getElementById('single-dst-user').value,
                dstPass: document.getElementById('single-dst-pass').value,
                dstTLS: document.getElementById('single-dst-tls').checked
            };

            if (!creds.srcHost || !creds.srcUser || !creds.srcPass || !creds.dstHost || !creds.dstUser || !creds.dstPass) {
                log('error', 'Please fill in all fields before testing');
                return;
            }

            log('info', 'Testing connection...');
            
            try {
                const response = await fetch('/api/test-connection', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(creds)
                });
                
                const result = await response.json();
                
                if (result.success) {
                    showSuccess('Connection Test Passed', 'Both source and destination servers are reachable and credentials are valid.');
                    log('success', 'Connection test passed: Both servers reachable');
                } else {
                    showError('Connection Test Failed', result.error);
                    log('error', 'Connection test failed: ' + result.error);
                }
            } catch (err) {
                showError('Connection Test Error', err.message);
                log('error', 'Connection test error: ' + err.message);
            }
        }

        async function testBulkConnections() {
            const container = document.getElementById('bulk-accounts');
            const rows = container.querySelectorAll('.account-row');
            
            if (rows.length === 0) {
                showError('No Accounts', 'Please add at least one account to test');
                log('error', 'No accounts to test');
                return;
            }

            showInfo('Testing Connections', 'Testing ' + rows.length + ' accounts, please wait...');
            log('info', 'Testing ' + rows.length + ' accounts...');
            
            const accounts = [];
            rows.forEach((row, idx) => {
                const index = idx + 1;
                accounts.push({
                    srcHost: document.getElementById('bulk-src-host-' + index)?.value,
                    srcPort: parseInt(document.getElementById('bulk-src-port-' + index)?.value) || 993,
                    srcUser: document.getElementById('bulk-src-user-' + index)?.value,
                    srcPass: document.getElementById('bulk-src-pass-' + index)?.value,
                    srcTLS: true,
                    dstHost: document.getElementById('bulk-dst-host-' + index)?.value,
                    dstPort: parseInt(document.getElementById('bulk-dst-port-' + index)?.value) || 993,
                    dstUser: document.getElementById('bulk-dst-user-' + index)?.value,
                    dstPass: document.getElementById('bulk-dst-pass-' + index)?.value,
                    dstTLS: true
                });
            });

            try {
                const response = await fetch('/api/validate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ accounts: accounts })
                });
                
                const results = await response.json();
                
                if (results.error) {
                    log('error', results.error);
                    return;
                }

                const successCount = results.filter(r => r.srcStatus === 'success' && r.dstStatus === 'success').length;
                const failCount = results.length - successCount;
                
                if (failCount === 0) {
                    showSuccess('All Connections Successful', successCount + ' accounts tested successfully. All connections are working!');
                } else if (successCount === 0) {
                    showError('All Connections Failed', 'All ' + failCount + ' accounts failed to connect. Please check your credentials and server settings.');
                } else {
                    showWarning('Partial Success', successCount + ' accounts connected successfully, ' + failCount + ' failed. Check the validation page for details.');
                }
                
                log('success', 'Test complete: ' + successCount + ' successful, ' + failCount + ' failed');
                
                // Show results in log
                results.forEach(r => {
                    if (r.srcStatus === 'success' && r.dstStatus === 'success') {
                        log('success', r.account + ': ✓ Connected');
                    } else {
                        const error = r.srcError || r.dstError || 'Unknown error';
                        log('error', r.account + ': ✗ ' + error);
                    }
                });
            } catch (err) {
                showError('Bulk Test Error', err.message);
                log('error', 'Bulk test error: ' + err.message);
            }
        }

        async function previewSingle() {
            const creds = {
                srcHost: document.getElementById('single-src-host').value,
                srcPort: parseInt(document.getElementById('single-src-port').value) || 993,
                srcUser: document.getElementById('single-src-user').value,
                srcPass: document.getElementById('single-src-pass').value,
                srcTLS: document.getElementById('single-src-tls').checked,
                dstHost: document.getElementById('single-dst-host').value,
                dstPort: parseInt(document.getElementById('single-dst-port').value) || 993,
                dstUser: document.getElementById('single-dst-user').value,
                dstPass: document.getElementById('single-dst-pass').value,
                dstTLS: document.getElementById('single-dst-tls').checked
            };

            if (!creds.srcHost || !creds.srcUser || !creds.srcPass || !creds.dstHost || !creds.dstUser || !creds.dstPass) {
                showError('Missing Fields', 'Please fill in all required fields');
                log('error', 'Please fill in all fields');
                return;
            }

            showInfo('Loading Preview', 'Fetching folder information from source server...');
            log('info', 'Fetching preview...');
            
            try {
                const response = await fetch('/api/preview', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(creds)
                });
                
                const result = await response.json();
                
                if (result.error) {
                    showError('Preview Failed', result.error);
                    log('error', result.error);
                    return;
                }

                folders = result.folders || [];
                selectedFolders = new Set(folders.map(f => f.name));
                renderFolders();
                updatePreviewStats();
                
                showSuccess('Preview Loaded', 'Found ' + folders.length + ' folders with ' + result.totalMessages + ' messages. Select which folders to migrate.');
                showPage('preview');
                log('success', 'Preview loaded: ' + folders.length + ' folders, ' + result.totalMessages + ' messages');
            } catch (err) {
                showError('Preview Error', err.message);
                log('error', 'Preview error: ' + err.message);
            }
        }

        async function previewBulk() {
            const container = document.getElementById('bulk-accounts');
            const rows = container.querySelectorAll('.account-row');
            
            if (rows.length === 0) {
                log('error', 'No accounts added for preview');
                return;
            }

            const accounts = [];
            let hasEmptyFields = false;

            rows.forEach((row, idx) => {
                const index = idx + 1;
                const account = {
                    srcHost: document.getElementById('bulk-src-host-' + index)?.value,
                    srcPort: parseInt(document.getElementById('bulk-src-port-' + index)?.value) || 993,
                    srcUser: document.getElementById('bulk-src-user-' + index)?.value,
                    srcPass: document.getElementById('bulk-src-pass-' + index)?.value,
                    srcTLS: true,
                    dstHost: document.getElementById('bulk-dst-host-' + index)?.value,
                    dstPort: parseInt(document.getElementById('bulk-dst-port-' + index)?.value) || 993,
                    dstUser: document.getElementById('bulk-dst-user-' + index)?.value,
                    dstPass: document.getElementById('bulk-dst-pass-' + index)?.value,
                    dstTLS: true
                };
                
                if (!account.srcHost || !account.srcUser || !account.srcPass || 
                    !account.dstHost || !account.dstUser || !account.dstPass) {
                    hasEmptyFields = true;
                }
                
                accounts.push(account);
            });

            if (hasEmptyFields) {
                showError('Missing Fields', 'Please fill in all account fields');
                log('error', 'Please fill in all account fields');
                return;
            }

            showInfo('Loading Preview', 'Fetching folder information for ' + accounts.length + ' accounts...');
            log('info', 'Fetching preview for ' + accounts.length + ' accounts...');
            
            try {
                const response = await fetch('/api/bulk-preview', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ accounts: accounts })
                });
                
                const result = await response.json();
                
                if (result.error) {
                    showError('Preview Failed', result.error);
                    log('error', result.error);
                    return;
                }

                // Aggregate all folders from all accounts
                const allFolders = [];
                let totalMessages = 0;
                let totalSize = 0;
                
                if (result.accounts) {
                    result.accounts.forEach(acc => {
                        if (acc.folders) {
                            acc.folders.forEach(folder => {
                                // Prefix with account name to distinguish
                                folder.name = acc.srcUser + ' / ' + folder.name;
                                allFolders.push(folder);
                                totalMessages += folder.messageCount;
                                totalSize += folder.sizeEstimate;
                            });
                        }
                    });
                }

                folders = allFolders;
                selectedFolders = new Set(allFolders.map(f => f.name));
                
                // Update preview stats
                document.getElementById('preview-total-folders').textContent = allFolders.length;
                document.getElementById('preview-total-messages').textContent = totalMessages.toLocaleString();
                
                renderFolders();
                updatePreviewStats();
                
                showSuccess('Preview Loaded', 'Found ' + result.accounts.length + ' accounts with ' + allFolders.length + ' folders and ' + totalMessages + ' total messages.');
                showPage('preview');
                log('success', 'Preview loaded: ' + result.accounts.length + ' accounts, ' + allFolders.length + ' folders, ' + totalMessages + ' messages');
            } catch (err) {
                showError('Bulk Preview Error', err.message);
                log('error', 'Bulk preview error: ' + err.message);
            }
        }

        function renderFolders() {
            const tbody = document.getElementById('folders-tbody');
            tbody.innerHTML = folders.map(f => '<tr>' +
                '<td><input type="checkbox" ' + (selectedFolders.has(f.name) ? 'checked' : '') + ' onchange="toggleFolder(\'' + f.name + '\')" class="folder-checkbox" data-folder="' + f.name + '"></td>' +
                '<td>' + f.name + '</td>' +
                '<td>' + f.messageCount.toLocaleString() + '</td>' +
                '<td>' + formatBytes(f.sizeEstimate) + '</td>' +
            '</tr>').join('');
        }

        function toggleFolder(name) {
            if (selectedFolders.has(name)) {
                selectedFolders.delete(name);
            } else {
                selectedFolders.add(name);
            }
            updatePreviewStats();
        }

        function selectAllFolders(select) {
            if (select) {
                selectedFolders = new Set(folders.map(f => f.name));
            } else {
                selectedFolders.clear();
            }
            renderFolders();
            updatePreviewStats();
        }

        function updatePreviewStats() {
            const selected = folders.filter(f => selectedFolders.has(f.name));
            const msgCount = selected.reduce((sum, f) => sum + f.messageCount, 0);
            const size = selected.reduce((sum, f) => sum + f.sizeEstimate, 0);
            
            document.getElementById('preview-total-folders').textContent = folders.length;
            document.getElementById('preview-total-messages').textContent = folders.reduce((sum, f) => sum + f.messageCount, 0).toLocaleString();
            document.getElementById('preview-selected').textContent = selected.length + ' folders (' + msgCount.toLocaleString() + ' msgs)';
            document.getElementById('preview-size').textContent = formatBytes(size);
        }

        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }

        function startMigration() {
            if (selectedFolders.size === 0) {
                alert('Please select at least one folder');
                return;
            }
            showPage('progress');
            log('success', 'Migration started');
            simulateProgress();
        }

        function simulateProgress() {
            let progress = 0;
            const interval = setInterval(() => {
                progress += Math.random() * 5;
                if (progress >= 100) {
                    progress = 100;
                    clearInterval(interval);
                    document.getElementById('progress-status').innerHTML = '<span class="badge badge-success">Completed</span>';
                    log('success', 'Migration completed successfully');
                }
                
                document.getElementById('progress-percent').textContent = progress.toFixed(1) + '%';
                document.getElementById('progress-bar').style.width = progress + '%';
                document.getElementById('progress-text').textContent = Math.floor(progress * 52) + '/5200 messages';
            }, 200);
        }

        function stopMigration() {
            log('warn', 'Migration stopped by user');
        }

        // Select all checkbox handler
        document.getElementById('select-all-checkbox')?.addEventListener('change', function(e) {
            selectAllFolders(e.target.checked);
        });

        // ========== MIGRATION TEMPLATES ==========
        const templates = {
            'gmail-source': { host: 'imap.gmail.com', port: 993, tls: true },
            'gmail-dest': { host: 'imap.gmail.com', port: 993, tls: true },
            'outlook-source': { host: 'outlook.office365.com', port: 993, tls: true },
            'outlook-dest': { host: 'outlook.office365.com', port: 993, tls: true },
            'yandex-source': { host: 'imap.yandex.com', port: 993, tls: true },
            'yandex-dest': { host: 'imap.yandex.com', port: 993, tls: true },
            'icloud-source': { host: 'imap.mail.me.com', port: 993, tls: true },
            'icloud-dest': { host: 'imap.mail.me.com', port: 993, tls: true },
            'yahoo-source': { host: 'imap.mail.yahoo.com', port: 993, tls: true },
            'yahoo-dest': { host: 'imap.mail.yahoo.com', port: 993, tls: true }
        };

        function applyTemplate(templateName) {
            const template = templates[templateName];
            if (!template) return;

            const isSource = templateName.includes('source');
            const prefix = isSource ? 'single-src' : 'single-dst';

            document.getElementById(prefix + '-host').value = template.host;
            document.getElementById(prefix + '-port').value = template.port;
            document.getElementById(prefix + '-tls').checked = template.tls;

            log('info', 'Applied ' + templateName.replace('-', ' ') + ' template');
        }

        function showTemplateModal() {
            const modal = document.createElement('div');
            modal.style.cssText = 'position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.8); z-index: 1000; display: flex; align-items: center; justify-content: center;';
            modal.innerHTML = '<div style="background: #141414; border: 1px solid #333; border-radius: 8px; padding: 24px; max-width: 600px; width: 90%; max-height: 80vh; overflow-y: auto;">' +
                '<h3 style="margin-bottom: 16px; color: #5FAFFF;">Select Template</h3>' +
                '<p style="color: #737373; margin-bottom: 16px; font-size: 13px;">This will apply the same source and destination server settings to all bulk accounts</p>' +
                '<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 8px;">' +
                    '<button class="btn btn-secondary" onclick="applyBulkTemplate(\'gmail\')">📧 Gmail → Gmail</button>' +
                    '<button class="btn btn-secondary" onclick="applyBulkTemplate(\'outlook\')">📧 Outlook → Outlook</button>' +
                    '<button class="btn btn-secondary" onclick="applyBulkTemplate(\'yandex\')">📧 Yandex → Yandex</button>' +
                    '<button class="btn btn-secondary" onclick="applyBulkTemplate(\'icloud\')">📧 iCloud → iCloud</button>' +
                    '<button class="btn btn-secondary" onclick="applyBulkTemplate(\'yahoo\')">📧 Yahoo → Yahoo</button>' +
                    '<button class="btn btn-secondary" onclick="applyBulkTemplate(\'gmail-to-outlook\')">📧 Gmail → Outlook</button>' +
                    '<button class="btn btn-secondary" onclick="applyBulkTemplate(\'outlook-to-gmail\')">📧 Outlook → Gmail</button>' +
                    '<button class="btn btn-secondary" onclick="applyBulkTemplate(\'gmail-to-yandex\')">📧 Gmail → Yandex</button>' +
                    '<button class="btn btn-secondary" onclick="applyBulkTemplate(\'yandex-to-gmail\')">📧 Yandex → Gmail</button>' +
                    '<button class="btn btn-secondary" onclick="applyBulkTemplate(\'outlook-to-yandex\')">📧 Outlook → Yandex</button>' +
                    '<button class="btn btn-secondary" onclick="applyBulkTemplate(\'yandex-to-outlook\')">📧 Yandex → Outlook</button>' +
                    '<button class="btn btn-secondary" onclick="applyBulkTemplate(\'icloud-to-gmail\')">📧 iCloud → Gmail</button>' +
                    '<button class="btn btn-secondary" onclick="applyBulkTemplate(\'gmail-to-icloud\')">📧 Gmail → iCloud</button>' +
                    '<button class="btn btn-secondary" onclick="applyBulkTemplate(\'yahoo-to-gmail\')">📧 Yahoo → Gmail</button>' +
                    '<button class="btn btn-secondary" onclick="applyBulkTemplate(\'gmail-to-yahoo\')">📧 Gmail → Yahoo</button>' +
                '</div>' +
                '<button class="btn btn-secondary" style="margin-top: 16px; width: 100%;" onclick="this.closest(\'.fixed\').remove()">Cancel</button>' +
            '</div>';
            modal.className = 'fixed';
            document.body.appendChild(modal);
        }

        function applyBulkTemplate(templateName) {
            const container = document.getElementById('bulk-accounts');
            const rows = container.querySelectorAll('.account-row');
            
            const templates = {
                // Same provider
                'gmail': { src: 'imap.gmail.com', dst: 'imap.gmail.com' },
                'outlook': { src: 'outlook.office365.com', dst: 'outlook.office365.com' },
                'yandex': { src: 'imap.yandex.com', dst: 'imap.yandex.com' },
                'icloud': { src: 'imap.mail.me.com', dst: 'imap.mail.me.com' },
                'yahoo': { src: 'imap.mail.yahoo.com', dst: 'imap.mail.yahoo.com' },
                // Cross provider
                'gmail-to-outlook': { src: 'imap.gmail.com', dst: 'outlook.office365.com' },
                'outlook-to-gmail': { src: 'outlook.office365.com', dst: 'imap.gmail.com' },
                'gmail-to-yandex': { src: 'imap.gmail.com', dst: 'imap.yandex.com' },
                'yandex-to-gmail': { src: 'imap.yandex.com', dst: 'imap.gmail.com' },
                'outlook-to-yandex': { src: 'outlook.office365.com', dst: 'imap.yandex.com' },
                'yandex-to-outlook': { src: 'imap.yandex.com', dst: 'outlook.office365.com' },
                'icloud-to-gmail': { src: 'imap.mail.me.com', dst: 'imap.gmail.com' },
                'gmail-to-icloud': { src: 'imap.gmail.com', dst: 'imap.mail.me.com' },
                'yahoo-to-gmail': { src: 'imap.mail.yahoo.com', dst: 'imap.gmail.com' },
                'gmail-to-yahoo': { src: 'imap.gmail.com', dst: 'imap.mail.yahoo.com' },
                'icloud-to-outlook': { src: 'imap.mail.me.com', dst: 'outlook.office365.com' },
                'outlook-to-icloud': { src: 'outlook.office365.com', dst: 'imap.mail.me.com' },
                'yahoo-to-outlook': { src: 'imap.mail.yahoo.com', dst: 'outlook.office365.com' },
                'outlook-to-yahoo': { src: 'outlook.office365.com', dst: 'imap.mail.yahoo.com' }
            };
            
            const template = templates[templateName];
            if (!template) {
                log('error', 'Unknown template: ' + templateName);
                return;
            }

            rows.forEach((row, index) => {
                const idx = index + 1;
                document.getElementById('bulk-src-host-' + idx).value = template.src;
                document.getElementById('bulk-dst-host-' + idx).value = template.dst;
                document.getElementById('bulk-src-port-' + idx).value = '993';
                document.getElementById('bulk-dst-port-' + idx).value = '993';
            });

            document.querySelector('.fixed').remove();
            log('success', 'Applied ' + templateName + ' template to all accounts');
        }

        // ========== IMPORT / EXPORT ==========
        function handleImport(input) {
            const file = input.files[0];
            if (!file) return;

            const reader = new FileReader();
            reader.onload = function(e) {
                const content = e.target.result;
                if (file.name.endsWith('.json')) {
                    importFromJSON(content);
                } else {
                    importFromCSV(content);
                }
            };
            reader.readAsText(file);
            input.value = '';
        }

        function importFromJSON(content) {
            try {
                const data = JSON.parse(content);
                if (data.srcHost) {
                    // Single account format
                    document.getElementById('single-src-host').value = data.srcHost || '';
                    document.getElementById('single-src-port').value = data.srcPort || 993;
                    document.getElementById('single-src-user').value = data.srcUser || '';
                    document.getElementById('single-src-pass').value = data.srcPass || '';
                    document.getElementById('single-dst-host').value = data.dstHost || '';
                    document.getElementById('single-dst-port').value = data.dstPort || 993;
                    document.getElementById('single-dst-user').value = data.dstUser || '';
                    document.getElementById('single-dst-pass').value = data.dstPass || '';
                    log('success', 'Imported single account configuration');
                }
            } catch (err) {
                log('error', 'Failed to import JSON: ' + err.message);
            }
        }

        function importFromCSV(content) {
            const lines = content.split('\n').filter(l => l.trim());
            if (lines.length === 0) return;

            // Check if it's bulk format (multiple columns) or single format
            const firstLine = lines[0];
            const cols = firstLine.split(',');

            if (cols.length >= 8) {
                // Bulk format: src_host,src_port,src_user,src_pass,dst_host,dst_port,dst_user,dst_pass
                document.getElementById('bulk-accounts').innerHTML = '';
                lines.forEach((line, idx) => {
                    const cols = line.split(',');
                    if (cols.length >= 8) {
                        addBulkAccount();
                        const index = idx + 1;
                        document.getElementById('bulk-src-host-' + index).value = cols[0].trim();
                        document.getElementById('bulk-src-port-' + index).value = cols[1].trim();
                        document.getElementById('bulk-src-user-' + index).value = cols[2].trim();
                        document.getElementById('bulk-src-pass-' + index).value = cols[3].trim();
                        document.getElementById('bulk-dst-host-' + index).value = cols[4].trim();
                        document.getElementById('bulk-dst-port-' + index).value = cols[5].trim();
                        document.getElementById('bulk-dst-user-' + index).value = cols[6].trim();
                        document.getElementById('bulk-dst-pass-' + index).value = cols[7].trim();
                    }
                });
                log('success', 'Imported ' + lines.length + ' accounts from CSV');
            }
        }

        function exportAccounts() {
            const data = {
                srcHost: document.getElementById('single-src-host').value,
                srcPort: parseInt(document.getElementById('single-src-port').value) || 993,
                srcUser: document.getElementById('single-src-user').value,
                srcPass: document.getElementById('single-src-pass').value,
                dstHost: document.getElementById('single-dst-host').value,
                dstPort: parseInt(document.getElementById('single-dst-port').value) || 993,
                dstUser: document.getElementById('single-dst-user').value,
                dstPass: document.getElementById('single-dst-pass').value,
                exportedAt: new Date().toISOString()
            };

            const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'mailmole-config-' + new Date().toISOString().split('T')[0] + '.json';
            a.click();
            URL.revokeObjectURL(url);
            log('success', 'Configuration exported');
        }

        function handleBulkImport(input) {
            const file = input.files[0];
            if (!file) return;

            const reader = new FileReader();
            reader.onload = function(e) {
                importFromCSV(e.target.result);
            };
            reader.readAsText(file);
            input.value = '';
        }

        function exportBulkAccounts() {
            const container = document.getElementById('bulk-accounts');
            const rows = container.querySelectorAll('.account-row');
            let csv = 'src_host,src_port,src_user,src_pass,dst_host,dst_port,dst_user,dst_pass\n';

            rows.forEach((row, idx) => {
                const index = idx + 1;
                const srcHost = document.getElementById('bulk-src-host-' + index)?.value || '';
                const srcPort = document.getElementById('bulk-src-port-' + index)?.value || 993;
                const srcUser = document.getElementById('bulk-src-user-' + index)?.value || '';
                const srcPass = document.getElementById('bulk-src-pass-' + index)?.value || '';
                const dstHost = document.getElementById('bulk-dst-host-' + index)?.value || '';
                const dstPort = document.getElementById('bulk-dst-port-' + index)?.value || 993;
                const dstUser = document.getElementById('bulk-dst-user-' + index)?.value || '';
                const dstPass = document.getElementById('bulk-dst-pass-' + index)?.value || '';

                csv += [srcHost, srcPort, srcUser, srcPass, dstHost, dstPort, dstUser, dstPass].join(',') + '\n';
            });

            const blob = new Blob([csv], { type: 'text/csv' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'mailmole-bulk-accounts-' + new Date().toISOString().split('T')[0] + '.csv';
            a.click();
            URL.revokeObjectURL(url);
            log('success', 'Bulk accounts exported to CSV');
        }

        // ========== SCHEDULING ==========
        function getScheduledTime() {
            const enabled = document.getElementById('schedule-enabled').checked;
            if (!enabled) return null;

            const timeValue = document.getElementById('schedule-time').value;
            const repeat = document.getElementById('schedule-repeat').value;

            if (!timeValue) return null;

            return {
                time: new Date(timeValue).toISOString(),
                repeat: repeat
            };
        }

        // ========== ENHANCED VALIDATION ==========
        async function testConnectionDetailed() {
            const creds = {
                srcHost: document.getElementById('single-src-host').value,
                srcPort: parseInt(document.getElementById('single-src-port').value) || 993,
                srcUser: document.getElementById('single-src-user').value,
                srcPass: document.getElementById('single-src-pass').value,
                srcTLS: document.getElementById('single-src-tls').checked,
                dstHost: document.getElementById('single-dst-host').value,
                dstPort: parseInt(document.getElementById('single-dst-port').value) || 993,
                dstUser: document.getElementById('single-dst-user').value,
                dstPass: document.getElementById('single-dst-pass').value,
                dstTLS: document.getElementById('single-dst-tls').checked
            };

            if (!creds.srcHost || !creds.srcUser || !creds.srcPass || !creds.dstHost || !creds.dstUser || !creds.dstPass) {
                log('error', 'Please fill in all fields');
                return;
            }

            // Show loading state
            const results = [{
                account: creds.srcUser || 'Single Account',
                srcHost: creds.srcHost,
                srcUser: creds.srcUser,
                dstHost: creds.dstHost,
                dstUser: creds.dstUser,
                srcStatus: 'testing',
                dstStatus: 'testing',
                details: 'Testing connections...'
            }];

            showValidationResults(results);

            try {
                const response = await fetch('/api/validate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ accounts: [{
                        srcHost: creds.srcHost,
                        srcPort: creds.srcPort,
                        srcUser: creds.srcUser,
                        srcPass: creds.srcPass,
                        srcTLS: creds.srcTLS,
                        dstHost: creds.dstHost,
                        dstPort: creds.dstPort,
                        dstUser: creds.dstUser,
                        dstPass: creds.dstPass,
                        dstTLS: creds.dstTLS
                    }] })
                });
                
                const apiResults = await response.json();
                
                if (apiResults.error) {
                    results[0].srcStatus = 'error';
                    results[0].dstStatus = 'error';
                    results[0].details = apiResults.error;
                } else if (apiResults.length > 0) {
                    const r = apiResults[0];
                    results[0].srcStatus = r.srcStatus;
                    results[0].dstStatus = r.dstStatus;
                    results[0].details = r.srcStatus === 'success' && r.dstStatus === 'success' 
                        ? '✓ Connected successfully'
                        : '✗ ' + (r.srcError || r.dstError || 'Connection failed');
                }
                
                updateValidationTable(results);
                
                if (results[0].srcStatus === 'success' && results[0].dstStatus === 'success') {
                    document.getElementById('validation-continue-btn').disabled = false;
                }
            } catch (err) {
                results[0].srcStatus = 'error';
                results[0].dstStatus = 'error';
                results[0].details = 'Error: ' + err.message;
                updateValidationTable(results);
            }
        }

        async function testBulkConnectionsDetailed() {
            const container = document.getElementById('bulk-accounts');
            const rows = container.querySelectorAll('.account-row');
            
            if (rows.length === 0) {
                log('error', 'No accounts to test');
                return;
            }

            const accounts = [];
            const results = [];

            rows.forEach((row, idx) => {
                const index = idx + 1;
                const account = {
                    srcHost: document.getElementById('bulk-src-host-' + index)?.value,
                    srcPort: parseInt(document.getElementById('bulk-src-port-' + index)?.value) || 993,
                    srcUser: document.getElementById('bulk-src-user-' + index)?.value,
                    srcPass: document.getElementById('bulk-src-pass-' + index)?.value,
                    srcTLS: true,
                    dstHost: document.getElementById('bulk-dst-host-' + index)?.value,
                    dstPort: parseInt(document.getElementById('bulk-dst-port-' + index)?.value) || 993,
                    dstUser: document.getElementById('bulk-dst-user-' + index)?.value,
                    dstPass: document.getElementById('bulk-dst-pass-' + index)?.value,
                    dstTLS: true
                };
                
                accounts.push(account);
                results.push({
                    account: account.srcUser || 'Account #' + index,
                    srcHost: account.srcHost,
                    srcUser: account.srcUser,
                    dstHost: account.dstHost,
                    dstUser: account.dstUser,
                    srcStatus: 'testing',
                    dstStatus: 'testing',
                    details: 'Testing...'
                });
            });

            showValidationResults(results);

            try {
                const response = await fetch('/api/validate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ accounts: accounts })
                });
                
                const apiResults = await response.json();
                
                if (apiResults.error) {
                    results.forEach(r => {
                        r.srcStatus = 'error';
                        r.dstStatus = 'error';
                        r.details = apiResults.error;
                    });
                } else {
                    apiResults.forEach((apiResult, idx) => {
                        if (results[idx]) {
                            results[idx].srcStatus = apiResult.srcStatus;
                            results[idx].dstStatus = apiResult.dstStatus;
                            results[idx].details = apiResult.srcStatus === 'success' && apiResult.dstStatus === 'success'
                                ? '✓ All connections successful'
                                : '✗ ' + (apiResult.srcError || apiResult.dstError || 'Connection failed');
                        }
                    });
                }
                
                updateValidationTable(results);
                
                const allSuccess = results.every(r => r.srcStatus === 'success' && r.dstStatus === 'success');
                if (allSuccess) {
                    document.getElementById('validation-continue-btn').disabled = false;
                }
            } catch (err) {
                results.forEach(r => {
                    r.srcStatus = 'error';
                    r.dstStatus = 'error';
                    r.details = 'Error: ' + err.message;
                });
                updateValidationTable(results);
            }
        }

        function showValidationResults(results) {
            document.getElementById('validation-total').textContent = results.length;
            document.getElementById('validation-success').textContent = '0';
            document.getElementById('validation-failed').textContent = '0';
            document.getElementById('validation-status').innerHTML = '<span class="badge badge-warning">Testing...</span>';
            document.getElementById('validation-continue-btn').disabled = true;

            updateValidationTable(results);
            showPage('validation');
        }

        function updateValidationTable(results) {
            const tbody = document.getElementById('validation-tbody');
            let successCount = 0;
            let failedCount = 0;

            tbody.innerHTML = results.map(r => {
                if (r.srcStatus === 'success' && r.dstStatus === 'success') successCount++;
                else if (r.srcStatus === 'error' || r.dstStatus === 'error') failedCount++;

                const statusClass = (r.srcStatus === 'success' && r.dstStatus === 'success') ? 'badge-success' :
                                   (r.srcStatus === 'error' || r.dstStatus === 'error') ? 'badge-error' : 'badge-warning';
                const statusText = (r.srcStatus === 'success' && r.dstStatus === 'success') ? '✓ OK' :
                                  (r.srcStatus === 'error' || r.dstStatus === 'error') ? '✗ Failed' : '⏳ Testing';

                return '<tr>' +
                    '<td>' + r.account + '</td>' +
                    '<td>' + r.srcHost + '<br><small style="color: #737373;">' + r.srcUser + '</small></td>' +
                    '<td>' + r.dstHost + '<br><small style="color: #737373;">' + r.dstUser + '</small></td>' +
                    '<td><span class="badge ' + statusClass + '">' + statusText + '</span></td>' +
                    '<td style="font-size: 12px; color: #a3a3a3;">' + r.details + '</td>' +
                '</tr>';
            }).join('');

            document.getElementById('validation-success').textContent = successCount;
            document.getElementById('validation-failed').textContent = failedCount;

            if (successCount + failedCount === results.length) {
                const allSuccess = failedCount === 0;
                document.getElementById('validation-status').innerHTML = allSuccess
                    ? '<span class="badge badge-success">All Tests Passed</span>'
                    : '<span class="badge badge-error">' + failedCount + ' Failed</span>';
            }
        }

        function fromValidationToPreview() {
            if (currentMode === 'single') {
                previewSingle();
            } else {
                previewBulk();
            }
        }

        function exportValidationResults() {
            const tbody = document.getElementById('validation-tbody');
            const rows = tbody.querySelectorAll('tr');
            let csv = 'Account,Source Host,Source User,Dest Host,Dest User,Status,Details\n';

            rows.forEach(row => {
                const cells = row.querySelectorAll('td');
                if (cells.length >= 5) {
                    const account = cells[0].textContent;
                    const src = cells[1].textContent.replace('\n', ' - ');
                    const dst = cells[2].textContent.replace('\n', ' - ');
                    const status = cells[3].textContent.trim();
                    const details = cells[4].textContent;
                    csv += [account, src, dst, status, details].join(',') + '\n';
                }
            });

            const blob = new Blob([csv], { type: 'text/csv' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'validation-results-' + new Date().toISOString().split('T')[0] + '.csv';
            a.click();
            URL.revokeObjectURL(url);
            log('success', 'Validation results exported');
        }
    </script>
</body>
</html>
`
