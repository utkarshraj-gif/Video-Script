package main

import (
	"encoding/json"
	"fmt"
)

// renderProductionUI renders a clean, white-and-light-blue, production-level UI with a static, non-shifting Chatbot,
// URL input extraction at the top for processing any video, highlighted key points, multi-color Neo4j concept pills,
// direct working Audio Download & Stream player, an interactive in-app Neo4j Force-Directed Graph Visualizer,
// Qdrant 384-dim vector explorer, and live Cypher query console.
func renderProductionUI(res *EmbedChunksResponse) string {
	resBytes, _ := json.Marshal(res)
	safeJSON := string(resBytes)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Video-RAG | AI Assistant & Knowledge Graph</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg: #f8fafc;
      --card-bg: #ffffff;
      --border: #e2e8f0;
      --border-light: #f1f5f9;
      --primary: #0284c7;
      --primary-hover: #0369a1;
      --primary-light: #f0f9ff;
      --primary-border: #bae6fd;
      --text: #0f172a;
      --text-muted: #64748b;
      --text-subtle: #94a3b8;
      --shadow-sm: 0 1px 3px rgba(0,0,0,0.05), 0 1px 2px rgba(0,0,0,0.03);
      --shadow-md: 0 4px 16px -2px rgba(2, 132, 199, 0.08), 0 2px 6px -1px rgba(0,0,0,0.04);
      --shadow-lg: 0 10px 25px -3px rgba(2, 132, 199, 0.12), 0 4px 10px -2px rgba(0,0,0,0.04);
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    
    /* STATIC VIEWPORT LOCK: Stops page from scrolling up and down */
    html, body {
      height: 100vh;
      max-height: 100vh;
      overflow: hidden;
      margin: 0;
      padding: 0;
      display: flex;
      flex-direction: column;
      background: var(--bg);
      color: var(--text);
      font-family: 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      line-height: 1.5;
    }

    /* Top Navigation Bar */
    .navbar {
      background: #ffffff;
      border-bottom: 1px solid var(--border);
      padding: 10px 24px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      flex-shrink: 0;
      z-index: 40;
      box-shadow: var(--shadow-sm);
    }
    .nav-brand {
      display: flex;
      align-items: center;
      gap: 10px;
      text-decoration: none;
    }
    .brand-icon {
      width: 34px;
      height: 34px;
      background: linear-gradient(135deg, #0284c7 0%%, #38bdf8 100%%);
      border-radius: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #fff;
      font-weight: 800;
      font-size: 1rem;
      box-shadow: 0 3px 8px rgba(2, 132, 199, 0.25);
    }
    .brand-title {
      font-size: 1.1rem;
      font-weight: 800;
      color: #0f172a;
      letter-spacing: -0.02em;
    }
    .brand-subtitle {
      font-size: 0.725rem;
      color: var(--text-muted);
      font-weight: 500;
    }
    .nav-status {
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .status-pill {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 4px 10px;
      border-radius: 9999px;
      font-size: 0.725rem;
      font-weight: 600;
    }
    .pill-qdrant {
      background: var(--primary-light);
      color: var(--primary);
      border: 1px solid var(--primary-border);
    }
    .pill-neo4j {
      background: #f0fdf4;
      color: #15803d;
      border: 1px solid #bbf7d0;
    }
    .pill-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%%;
      background: currentColor;
    }
    .nav-links {
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .btn {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 6px 12px;
      border-radius: 8px;
      font-size: 0.8rem;
      font-weight: 600;
      text-decoration: none;
      cursor: pointer;
      border: none;
      transition: all 0.15s ease;
      font-family: inherit;
    }
    .btn-primary {
      background: var(--primary);
      color: #ffffff;
      box-shadow: 0 2px 4px rgba(2, 132, 199, 0.2);
    }
    .btn-primary:hover {
      background: var(--primary-hover);
      transform: translateY(-1px);
    }
    .btn-outline {
      background: #ffffff;
      color: var(--primary);
      border: 1px solid var(--primary-border);
    }
    .btn-outline:hover {
      background: var(--primary-light);
    }
    .btn-sm {
      padding: 4px 8px;
      font-size: 0.725rem;
    }

    /* URL Extraction Bar (Top Input Box) */
    .url-bar-section {
      background: #ffffff;
      border-bottom: 1px solid var(--border);
      padding: 8px 24px;
      flex-shrink: 0;
      z-index: 30;
    }
    .url-bar-wrap {
      max-width: 1240px;
      margin: 0 auto;
      display: flex;
      align-items: center;
      gap: 10px;
    }
    .url-input-container {
      flex: 1;
      display: flex;
      align-items: center;
      background: #f8fafc;
      border: 1.5px solid var(--border);
      border-radius: 10px;
      padding: 4px 12px;
      transition: all 0.2s;
    }
    .url-input-container:focus-within {
      background: #ffffff;
      border-color: var(--primary);
      box-shadow: 0 0 0 3px rgba(2, 132, 199, 0.12);
    }
    .url-icon {
      font-size: 1rem;
      color: var(--primary);
      margin-right: 8px;
    }
    .url-input {
      flex: 1;
      border: none;
      background: transparent;
      outline: none;
      font-family: inherit;
      font-size: 0.85rem;
      color: var(--text);
      font-weight: 500;
    }
    .url-input::placeholder {
      color: var(--text-subtle);
    }
    .btn-extract {
      padding: 8px 18px;
      font-size: 0.85rem;
      font-weight: 700;
      border-radius: 10px;
      display: inline-flex;
      align-items: center;
      gap: 6px;
      white-space: nowrap;
    }

    /* Compact Video Hero & Tabs */
    .video-hero-compact {
      background: #ffffff;
      border-bottom: 1px solid var(--border);
      padding: 10px 24px;
      flex-shrink: 0;
    }
    .video-hero-inner {
      max-width: 1240px;
      margin: 0 auto;
      display: flex;
      justify-content: space-between;
      align-items: center;
      flex-wrap: wrap;
      gap: 12px;
    }
    .video-info-left {
      display: flex;
      align-items: center;
      gap: 10px;
    }
    .video-play-badge {
      width: 32px;
      height: 32px;
      background: #e0f2fe;
      color: var(--primary);
      border-radius: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 0.85rem;
      font-weight: 800;
    }
    .video-title {
      font-size: 0.95rem;
      font-weight: 800;
      color: #0f172a;
      letter-spacing: -0.01em;
    }
    .video-meta {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 0.725rem;
      color: var(--text-muted);
    }
    .video-id-badge {
      font-family: 'JetBrains Mono', monospace;
      background: var(--primary-light);
      color: var(--primary);
      padding: 1px 6px;
      border-radius: 4px;
      font-size: 0.7rem;
      border: 1px solid var(--primary-border);
    }

    /* Tabs Bar */
    .tab-bar {
      display: flex;
      gap: 6px;
      align-items: center;
      flex-wrap: wrap;
    }
    .tab-btn {
      padding: 6px 14px;
      font-size: 0.8rem;
      font-weight: 600;
      color: var(--text-muted);
      border: 1px solid var(--border);
      background: #ffffff;
      border-radius: 8px;
      cursor: pointer;
      display: flex;
      align-items: center;
      gap: 6px;
      transition: all 0.15s;
    }
    .tab-btn:hover {
      color: var(--primary);
      background: var(--primary-light);
      border-color: var(--primary-border);
    }
    .tab-btn.active {
      color: var(--primary);
      border-color: var(--primary-border);
      background: var(--primary-light);
      font-weight: 700;
    }
    .tab-count {
      background: #e2e8f0;
      color: #334155;
      font-size: 0.675rem;
      padding: 1px 6px;
      border-radius: 9999px;
      font-weight: 700;
    }
    .tab-btn.active .tab-count {
      background: var(--primary);
      color: #ffffff;
    }

    /* MAIN CONTAINER: Static layout with internal scrolling */
    .main-wrap {
      flex: 1;
      min-height: 0;
      width: 100%%;
      max-width: 1240px;
      margin: 0 auto;
      padding: 12px 24px;
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }

    /* TAB 1: Chatbot (STATIC CONTAINER, ONLY MESSAGES SCROLL) */
    .tab-content {
      height: 100%%;
      flex: 1;
      min-height: 0;
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }
    .chat-layout {
      flex: 1;
      min-height: 0;
      height: 100%%;
      display: flex;
      flex-direction: column;
    }
    .chat-box {
      flex: 1;
      min-height: 0;
      height: 100%%;
      background: #ffffff;
      border: 1px solid var(--border);
      border-radius: 12px;
      display: flex;
      flex-direction: column;
      box-shadow: var(--shadow-sm);
      overflow: hidden;
    }
    .chat-header {
      padding: 10px 18px;
      border-bottom: 1px solid var(--border);
      background: #ffffff;
      display: flex;
      justify-content: space-between;
      align-items: center;
      flex-shrink: 0;
    }
    .chat-header-title {
      font-size: 0.875rem;
      font-weight: 700;
      color: #0f172a;
      display: flex;
      align-items: center;
      gap: 6px;
    }

    /* The ONLY scrollable area */
    .chat-messages {
      flex: 1;
      min-height: 0;
      overflow-y: auto;
      overscroll-behavior: contain;
      padding: 16px 20px;
      display: flex;
      flex-direction: column;
      gap: 16px;
      background: #fcfdfe;
    }
    .msg {
      display: flex;
      gap: 10px;
      max-width: 86%%;
    }
    .msg-user {
      align-self: flex-end;
      flex-direction: row-reverse;
    }
    .msg-avatar {
      width: 30px;
      height: 30px;
      border-radius: 50%%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 0.775rem;
      font-weight: 700;
      flex-shrink: 0;
    }
    .avatar-user {
      background: var(--primary);
      color: #ffffff;
    }
    .avatar-bot {
      background: #e0f2fe;
      color: var(--primary);
      border: 1px solid var(--primary-border);
    }
    .msg-bubble {
      padding: 12px 16px;
      border-radius: 12px;
      font-size: 0.875rem;
      line-height: 1.55;
      box-shadow: var(--shadow-sm);
    }
    .msg-user .msg-bubble {
      background: var(--primary);
      color: #ffffff;
      border-bottom-right-radius: 3px;
    }
    .msg-bot .msg-bubble {
      background: #ffffff;
      color: #1e293b;
      border: 1px solid var(--border);
      border-bottom-left-radius: 3px;
    }

    /* Highlighting Elements */
    .highlights-box {
      margin-top: 10px;
      background: var(--primary-light);
      border-left: 3px solid var(--primary);
      border-radius: 0 8px 8px 0;
      padding: 8px 12px;
    }
    .highlights-title {
      font-size: 0.725rem;
      font-weight: 700;
      color: var(--primary);
      text-transform: uppercase;
      letter-spacing: 0.04em;
      margin-bottom: 4px;
      display: flex;
      align-items: center;
      gap: 6px;
    }
    .highlights-list {
      list-style-type: disc;
      padding-left: 16px;
      font-size: 0.825rem;
      color: #334155;
    }
    .highlights-list li {
      margin-bottom: 3px;
    }

    /* Multi-color Neo4j Concept Tags */
    .concepts-row {
      display: flex;
      align-items: center;
      gap: 5px;
      flex-wrap: wrap;
      margin-top: 10px;
      padding-top: 8px;
      border-top: 1px dashed var(--border);
    }
    .concepts-label {
      font-size: 0.7rem;
      font-weight: 700;
      color: var(--text-muted);
      text-transform: uppercase;
      letter-spacing: 0.04em;
    }
    .concept-tag {
      font-size: 0.7rem;
      font-weight: 600;
      padding: 2px 7px;
      border-radius: 6px;
      display: inline-flex;
      align-items: center;
      gap: 4px;
      transition: transform 0.15s;
    }
    .concept-tag:hover {
      transform: translateY(-1px);
    }
    .tag-color-0 { background: #ecfdf5; color: #047857; border: 1px solid #a7f3d0; }
    .tag-color-1 { background: #eff6ff; color: #1d4ed8; border: 1px solid #bfdbfe; }
    .tag-color-2 { background: #f5f3ff; color: #6d28d9; border: 1px solid #ddd6fe; }
    .tag-color-3 { background: #fffbeb; color: #b45309; border: 1px solid #fde68a; }
    .tag-color-4 { background: #fff1f2; color: #be123c; border: 1px solid #fecdd3; }
    .tag-color-5 { background: #f0fdfa; color: #0f766e; border: 1px solid #99f6e4; }

    .msg-actions {
      display: flex;
      align-items: center;
      gap: 6px;
      margin-top: 8px;
      flex-wrap: wrap;
    }
    .yt-btn {
      background: #fee2e2;
      color: #b91c1c;
      border: 1px solid #fca5a5;
      font-size: 0.725rem;
      font-weight: 700;
      padding: 3px 8px;
      border-radius: 6px;
      text-decoration: none;
      display: inline-flex;
      align-items: center;
      gap: 4px;
      transition: background 0.15s;
    }
    .yt-btn:hover {
      background: #fecaca;
    }
    .graph-expand-btn {
      background: #f1f5f9;
      color: #334155;
      border: 1px solid var(--border);
      font-size: 0.725rem;
      font-weight: 600;
      padding: 3px 8px;
      border-radius: 6px;
      cursor: pointer;
      display: inline-flex;
      align-items: center;
      gap: 4px;
    }
    .graph-expand-btn:hover {
      background: #e2e8f0;
    }
    .graph-context-drawer {
      margin-top: 8px;
      background: #f8fafc;
      border: 1px solid var(--border);
      border-radius: 8px;
      padding: 8px 10px;
      font-size: 0.775rem;
      color: #475569;
      display: none;
    }

    /* Suggestions Bar: STATIC */
    .suggestions-bar {
      padding: 8px 18px;
      background: #ffffff;
      border-top: 1px solid var(--border);
      display: flex;
      gap: 6px;
      overflow-x: auto;
      white-space: nowrap;
      flex-shrink: 0;
    }
    .suggestion-chip {
      background: var(--primary-light);
      color: var(--primary);
      border: 1px solid var(--primary-border);
      padding: 4px 10px;
      border-radius: 9999px;
      font-size: 0.75rem;
      font-weight: 600;
      cursor: pointer;
      transition: all 0.15s;
    }
    .suggestion-chip:hover {
      background: #e0f2fe;
      transform: translateY(-1px);
    }

    /* Input area: STATIC AT BOTTOM */
    .chat-input-area {
      padding: 10px 18px;
      background: #ffffff;
      border-top: 1px solid var(--border);
      display: flex;
      gap: 8px;
      flex-shrink: 0;
    }
    .chat-input {
      flex: 1;
      padding: 10px 14px;
      border-radius: 8px;
      border: 1px solid var(--border);
      font-size: 0.875rem;
      outline: none;
      font-family: inherit;
      transition: border-color 0.2s, box-shadow 0.2s;
    }
    .chat-input:focus {
      border-color: var(--primary);
      box-shadow: 0 0 0 3px rgba(2, 132, 199, 0.12);
    }

    /* TAB 2 & 3: Chunks & Graph (Scrollable internally) */
    .tab-scrollable {
      flex: 1;
      min-height: 0;
      overflow-y: auto;
      padding-right: 6px;
    }
    .chunks-grid {
      display: flex;
      flex-direction: column;
      gap: 12px;
      padding-bottom: 20px;
    }
    .chunk-card {
      background: #ffffff;
      border: 1px solid var(--border);
      border-radius: 10px;
      padding: 14px 16px;
      box-shadow: var(--shadow-sm);
      transition: border-color 0.2s;
    }
    .chunk-card:hover {
      border-color: var(--primary);
    }
    .chunk-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 8px;
      flex-wrap: wrap;
      gap: 6px;
    }
    .chunk-meta {
      display: flex;
      align-items: center;
      gap: 6px;
    }
    .chunk-idx {
      background: #f1f5f9;
      color: #334155;
      font-weight: 700;
      font-size: 0.725rem;
      padding: 2px 6px;
      border-radius: 4px;
    }
    .chunk-time {
      background: #e0f2fe;
      color: #0369a1;
      font-weight: 600;
      font-size: 0.725rem;
      padding: 2px 6px;
      border-radius: 4px;
    }
    .chunk-words {
      color: var(--text-muted);
      font-size: 0.725rem;
    }
    .chunk-body {
      font-size: 0.85rem;
      color: #334155;
      line-height: 1.55;
      margin-bottom: 8px;
    }
    .vector-panel {
      margin-top: 8px;
      background: #f8fafc;
      border: 1px solid var(--border);
      border-radius: 6px;
      padding: 10px;
      display: none;
    }
    .vector-text {
      font-family: 'JetBrains Mono', monospace;
      font-size: 0.725rem;
      color: #0369a1;
      max-height: 120px;
      overflow-y: auto;
      word-break: break-all;
      background: #ffffff;
      padding: 6px;
      border-radius: 4px;
      border: 1px solid var(--border);
    }

    /* Visual Graph Tab Styles */
    .graph-card {
      background: #ffffff;
      border: 1px solid var(--border);
      border-radius: 12px;
      padding: 16px;
      box-shadow: var(--shadow-sm);
      margin-bottom: 16px;
    }
    .graph-toolbar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 12px;
      flex-wrap: wrap;
      gap: 8px;
    }
    .graph-stats-row {
      display: flex;
      gap: 8px;
      align-items: center;
      font-size: 0.75rem;
    }
    .graph-stat-badge {
      padding: 3px 8px;
      border-radius: 6px;
      font-weight: 600;
    }
    .stat-video { background: #e0f2fe; color: #0369a1; }
    .stat-chunks { background: #f0fdf4; color: #15803d; }
    .stat-concepts { background: #f5f3ff; color: #6d28d9; }
    .stat-rels { background: #fff7ed; color: #c2410c; }
    
    .canvas-container {
      position: relative;
      background: #fafcff;
      border: 1px solid var(--border);
      border-radius: 10px;
      height: 480px;
      overflow: hidden;
    }
    #graphCanvas {
      width: 100%%;
      height: 100%%;
      display: block;
      cursor: grab;
    }
    #graphCanvas:active {
      cursor: grabbing;
    }
    .graph-controls {
      position: absolute;
      top: 10px;
      right: 10px;
      display: flex;
      flex-direction: column;
      gap: 6px;
      z-index: 10;
    }
    .graph-ctrl-btn {
      width: 32px;
      height: 32px;
      border-radius: 6px;
      border: 1px solid var(--border);
      background: #ffffff;
      font-weight: 800;
      color: #334155;
      display: flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;
      box-shadow: var(--shadow-sm);
      transition: all 0.15s;
    }
    .graph-ctrl-btn:hover {
      background: var(--primary-light);
      color: var(--primary);
    }
    .graph-legend {
      position: absolute;
      bottom: 10px;
      left: 10px;
      background: rgba(255, 255, 255, 0.92);
      backdrop-filter: blur(4px);
      border: 1px solid var(--border);
      border-radius: 8px;
      padding: 6px 12px;
      font-size: 0.7rem;
      display: flex;
      gap: 12px;
      z-index: 10;
    }
    .legend-item {
      display: flex;
      align-items: center;
      gap: 5px;
      font-weight: 600;
    }
    .legend-dot {
      width: 10px;
      height: 10px;
      border-radius: 50%%;
    }

    /* Node Inspector Drawer */
    .node-inspector {
      position: absolute;
      top: 10px;
      left: 10px;
      width: 300px;
      max-height: 440px;
      overflow-y: auto;
      background: rgba(255, 255, 255, 0.96);
      backdrop-filter: blur(6px);
      border: 1px solid var(--primary-border);
      border-radius: 10px;
      padding: 14px;
      box-shadow: var(--shadow-md);
      z-index: 15;
      display: none;
      font-size: 0.8rem;
    }
    .inspector-title {
      font-weight: 800;
      color: var(--primary);
      margin-bottom: 6px;
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    /* Live Cypher Query Console */
    .cypher-console {
      background: #0f172a;
      border-radius: 10px;
      padding: 14px;
      color: #e2e8f0;
      margin-top: 14px;
    }
    .cypher-console-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 8px;
    }
    .cypher-input-row {
      display: flex;
      gap: 8px;
    }
    .cypher-input {
      flex: 1;
      background: #1e293b;
      border: 1px solid #334155;
      border-radius: 6px;
      padding: 8px 12px;
      color: #38bdf8;
      font-family: 'JetBrains Mono', monospace;
      font-size: 0.8rem;
      outline: none;
    }
    .cypher-input:focus {
      border-color: #38bdf8;
    }
    .cypher-result-box {
      margin-top: 10px;
      background: #1e293b;
      border-radius: 6px;
      padding: 10px;
      max-height: 160px;
      overflow-y: auto;
      font-family: 'JetBrains Mono', monospace;
      font-size: 0.75rem;
      color: #94a3b8;
      display: none;
    }

    /* Processing Modal Overlay */
    .modal-overlay {
      position: fixed;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: rgba(15, 23, 42, 0.6);
      backdrop-filter: blur(4px);
      display: none;
      align-items: center;
      justify-content: center;
      z-index: 100;
    }
    .modal-card {
      background: #ffffff;
      border-radius: 14px;
      padding: 24px;
      width: 90%%;
      max-width: 480px;
      box-shadow: var(--shadow-lg);
      border: 1px solid var(--border);
      text-align: center;
    }
    .modal-spinner {
      font-size: 2rem;
      animation: spin 1.2s infinite linear;
      margin-bottom: 12px;
    }
    @keyframes spin {
      0%% { transform: rotate(0deg); }
      100%% { transform: rotate(360deg); }
    }
    .modal-title {
      font-size: 1.1rem;
      font-weight: 800;
      color: #0f172a;
      margin-bottom: 8px;
    }
    .modal-desc {
      font-size: 0.85rem;
      color: var(--text-muted);
      margin-bottom: 16px;
      line-height: 1.5;
    }
    .progress-bar-wrap {
      background: #e2e8f0;
      height: 8px;
      border-radius: 4px;
      overflow: hidden;
      margin-bottom: 12px;
    }
    .progress-bar-fill {
      height: 100%%;
      background: linear-gradient(90deg, #0284c7, #38bdf8);
      width: 15%%;
      transition: width 0.4s ease;
      animation: pulseBar 1.5s infinite alternate;
    }
    @keyframes pulseBar {
      0%% { opacity: 0.7; }
      100%% { opacity: 1; }
    }
    .step-list {
      text-align: left;
      font-size: 0.8rem;
      color: #334155;
      display: flex;
      flex-direction: column;
      gap: 6px;
    }
    .step-item {
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .step-active {
      font-weight: 700;
      color: var(--primary);
    }
  </style>
</head>
<body>

  <!-- Top Navigation Bar -->
  <nav class="navbar">
    <div class="nav-brand">
      <div class="brand-icon">⚡</div>
      <div>
        <div class="brand-title">Video-RAG Studio</div>
        <div class="brand-subtitle">FastEmbed 384-Dim &bull; Qdrant Vector &bull; Neo4j Graph</div>
      </div>
    </div>

    <div class="nav-status">
      <div class="status-pill pill-qdrant">
        <span class="pill-dot"></span>
        <span id="pillQdrantCount">Qdrant: %d Chunks (384-dim)</span>
      </div>
      <div class="status-pill pill-neo4j">
        <span class="pill-dot"></span>
        <span>Neo4j: Graph Online</span>
      </div>
      <div class="status-pill pill-qdrant" style="border-color: #38bdf8;">
        <span class="pill-dot" style="background: #0284c7;"></span>
        <span>🦙 Ollama: Active</span>
      </div>
    </div>

    <div class="nav-links">
      <a href="http://localhost:6335/dashboard#/collections/video_chunks" target="_blank" class="btn btn-outline btn-sm">
        🚀 Qdrant UI
      </a>
      <a href="http://localhost:7474/browser/" target="_blank" class="btn btn-primary btn-sm" id="neo4jBrowserBtn" title="Open Neo4j Browser (Passwordless)">
        🌐 Neo4j Browser
      </a>
    </div>
  </nav>

  <!-- URL Extraction Bar (Paste Link & Click Extract / Submit) -->
  <div class="url-bar-section">
    <div class="url-bar-wrap">
      <div class="url-input-container">
        <span class="url-icon">🔗</span>
        <input type="text" id="videoUrlInput" class="url-input" 
               placeholder="Paste any YouTube URL or Video ID (e.g. https://www.youtube.com/watch?v=Rr4CW2quazA)..."
               value="https://www.youtube.com/watch?v=%s"
               onkeypress="if(event.key==='Enter') extractVideo()">
      </div>
      <button id="extractBtn" onclick="extractVideo()" class="btn btn-primary btn-extract">
        <span>⚡</span>
        <span id="extractBtnText">Extract / Submit</span>
      </button>
    </div>
  </div>

  <!-- Compact Video Hero & Tabs -->
  <div class="video-hero-compact">
    <div class="video-hero-inner">
      <div class="video-info-left">
        <div class="video-play-badge">▶</div>
        <div>
          <div class="video-title" id="videoTitleDisplay">%s</div>
          <div class="video-meta">
            <span>Video ID: <strong class="video-id-badge" id="currentVideoId">%s</strong></span>
            <span>&bull;</span>
            <span>Collection: <strong>%s</strong></span>
            <span>&bull;</span>
            <span>Model: <strong>384-dim all-MiniLM-L6-v2</strong></span>
          </div>
        </div>
      </div>

      <div class="tab-bar">
        <button class="tab-btn active" onclick="switchTab('chat')" id="tabBtnChat">
          💬 AI Assistant <span class="tab-count">Chatbot</span>
        </button>
        <button class="tab-btn" onclick="switchTab('chunks')" id="tabBtnChunks">
          📦 384-Dim Chunks <span class="tab-count" id="tabChunksCount">%d</span>
        </button>
        <button class="tab-btn" onclick="switchTab('graph')" id="tabBtnGraph">
          🕸️ Neo4j Graph <span class="tab-count">Visual</span>
        </button>
        <a href="/api/v1/videos/%s/download-audio?download=true" class="btn btn-primary btn-sm" id="downloadAudioLink" download title="Download MP3 Audio File">
          🎵 Download Audio
        </a>
        <button onclick="toggleAudioPlayer()" class="btn btn-outline btn-sm" id="listenAudioBtn" title="Play audio in browser">
          ▶️ Listen
        </button>
        <a href="https://youtu.be/%s" target="_blank" class="btn btn-outline btn-sm" id="ytVideoLink">
          📺 YouTube
        </a>
      </div>

      <!-- Integrated Audio Player Drawer -->
      <div id="audioPlayerWrap" style="display: none; align-items: center; gap: 10px; margin-top: 8px; width: 100%%; background: var(--primary-light); padding: 8px 12px; border-radius: 8px; border: 1px solid var(--primary-border);">
        <span style="font-size: 0.75rem; font-weight: 700; color: var(--primary); white-space: nowrap;">🎧 Audio Stream:</span>
        <audio id="inlineAudioPlayer" controls style="height: 32px; flex: 1; outline: none;" src="/api/v1/videos/%s/stream-audio">
          Your browser does not support audio playback.
        </audio>
        <a href="/api/v1/videos/%s/download-audio?download=true" class="btn btn-primary btn-sm" download title="Save MP3">
          ⬇️ Save MP3
        </a>
      </div>
    </div>
  </div>

  <!-- MAIN STATIC CONTAINER -->
  <main class="main-wrap">

    <!-- TAB 1: Chatbot Interface (STATIC: ONLY MESSAGES SCROLL) -->
    <div id="tabChat" class="tab-content">
      <div class="chat-layout">
        <div class="chat-box">
          <div class="chat-header">
            <div class="chat-header-title">
              <span>🤖 Video Q&A Assistant</span>
              <span style="font-size: 0.725rem; color: var(--primary); font-weight: 500;">(Qdrant Semantic + Neo4j Graph Expansion)</span>
            </div>
            <button onclick="clearChat()" class="btn btn-outline btn-sm">Clear History</button>
          </div>

          <!-- The ONLY scrollable area -->
          <div class="chat-messages" id="chatMessages">
            <div class="msg msg-bot">
              <div class="msg-avatar avatar-bot">AI</div>
              <div class="msg-bubble">
                Hello! I am your <strong>Video-RAG Assistant</strong>. I have indexed this video into <strong>384-dimensional vectors</strong> in Qdrant and linked its temporal sequence and key concepts in Neo4j.
                <br><br>
                Ask me any question about this video! I will highlight the <strong>key points</strong> and identify <strong>related Neo4j concepts</strong>.
              </div>
            </div>
          </div>

          <!-- Quick Suggestions: STATIC -->
          <div class="suggestions-bar">
            <span style="font-size: 0.725rem; font-weight: 700; color: var(--text-muted); align-self: center;">Try:</span>
            <div class="suggestion-chip" onclick="askQuestion('Summarize this video')">Summarize this</div>
          </div>

          <!-- Chat Input Area: STATIC -->
          <div class="chat-input-area">
            <input type="text" id="chatInput" class="chat-input" placeholder="Ask a question about this video or click 'Summarize this'..." onkeypress="if(event.key==='Enter') sendChatMessage()">
            <button onclick="sendChatMessage()" class="btn btn-primary" id="sendBtn">
              Send
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- TAB 2: Chunks & 384-Dim Vectors (Scrollable internally) -->
    <div id="tabChunks" class="tab-content tab-scrollable" style="display: none;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; flex-shrink: 0;">
        <h3 style="font-size: 0.95rem; font-weight: 700;">Stored Chunks in Qdrant (384-dim Vectors)</h3>
        <button onclick="expandAllVectors()" class="btn btn-outline btn-sm" id="expandAllBtn">Expand All Vectors</button>
      </div>
      <div class="chunks-grid" id="chunksContainer"></div>
    </div>

    <!-- TAB 3: Neo4j Knowledge Graph (INTERACTIVE VISUAL GRAPH) -->
    <div id="tabGraph" class="tab-content tab-scrollable" style="display: none;">
      <div class="graph-card">
        <div class="graph-toolbar">
          <div>
            <h3 style="font-size: 0.95rem; font-weight: 700; color: #0f172a;">🕸️ Interactive Knowledge Graph & Relationships</h3>
            <p style="font-size: 0.775rem; color: var(--text-muted);">
              Visualizing Video root, sequential Chunks (<code>[:NEXT_CHUNK]</code>), and extracted Concepts (<code>[:DISCUSSES]</code>) in Neo4j.
            </p>
          </div>
          <div class="graph-stats-row">
            <span class="graph-stat-badge stat-video">1 Video Node</span>
            <span class="graph-stat-badge stat-chunks" id="graphChunkBadge">%d Chunks</span>
            <span class="graph-stat-badge stat-concepts" id="graphConceptBadge">Concepts</span>
            <button onclick="fetchAndRenderGraph()" class="btn btn-outline btn-sm">🔄 Refresh Graph</button>
          </div>
        </div>

        <!-- Canvas Container with Interactive Force Graph -->
        <div class="canvas-container">
          <canvas id="graphCanvas"></canvas>

          <!-- Controls -->
          <div class="graph-controls">
            <button class="graph-ctrl-btn" onclick="zoomGraph(1.2)" title="Zoom In">+</button>
            <button class="graph-ctrl-btn" onclick="zoomGraph(0.8)" title="Zoom Out">-</button>
            <button class="graph-ctrl-btn" onclick="resetGraphView()" title="Reset View">⊙</button>
          </div>

          <!-- Legend -->
          <div class="graph-legend">
            <div class="legend-item"><span class="legend-dot" style="background: #0284c7;"></span> Video</div>
            <div class="legend-item"><span class="legend-dot" style="background: #38bdf8;"></span> Chunks</div>
            <div class="legend-item"><span class="legend-dot" style="background: #a855f7;"></span> Concepts</div>
            <div class="legend-item"><span style="color: #94a3b8; font-weight: 800;">&rarr;</span> NEXT_CHUNK</div>
          </div>

          <!-- Node Inspector Drawer (shown on node click) -->
          <div class="node-inspector" id="nodeInspector">
            <div class="inspector-title">
              <span id="inspectorTitle">Node Details</span>
              <button onclick="closeInspector()" style="border:none;background:transparent;cursor:pointer;font-size:1rem;">&times;</button>
            </div>
            <div id="inspectorContent">Click on any node to view details.</div>
          </div>
        </div>

        <!-- Live Cypher Query Console -->
        <div class="cypher-console">
          <div class="cypher-console-header">
            <span style="font-weight: 700; font-size: 0.8rem; color: #38bdf8;">⚡ Neo4j Live Cypher Console</span>
            <a href="http://localhost:7474/browser/" target="_blank" style="color: #94a3b8; font-size: 0.725rem; text-decoration: none;">
              Open Full Browser ↗
            </a>
          </div>
          <div class="cypher-input-row">
            <input type="text" id="cypherInput" class="cypher-input" value="MATCH (c:Chunk {video_id: '%s'})-[r:DISCUSSES]->(cp:Concept) RETURN c.timestamp_display AS time, cp.name AS concept LIMIT 15">
            <button onclick="runCustomCypher()" class="btn btn-primary btn-sm" id="runCypherBtn">▶ Run Query</button>
          </div>
          <div class="cypher-result-box" id="cypherResultBox"></div>
        </div>

        <!-- Direct Browser Notice -->
        <div style="margin-top: 14px; padding: 12px; background: var(--primary-light); border-radius: 8px; border: 1px solid var(--primary-border); display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px;">
          <div>
            <div style="font-weight: 700; font-size: 0.8rem; color: #0f172a;">Direct Neo4j Browser Access</div>
            <div style="font-size: 0.725rem; color: var(--text-muted);">
              Neo4j is running passwordless on port 7474 &amp; 7687. Query: <code>MATCH (n) RETURN n LIMIT 25</code>
            </div>
          </div>
          <a href="http://localhost:7474/browser/" target="_blank" class="btn btn-primary btn-sm">
            Launch Neo4j Browser 🌐
          </a>
        </div>
      </div>
    </div>

  </main>

  <!-- Processing Modal Overlay -->
  <div class="modal-overlay" id="processingModal">
    <div class="modal-card">
      <div class="modal-spinner">⚙️</div>
      <div class="modal-title" id="modalTitle">Processing YouTube Video</div>
      <div class="modal-desc" id="modalDesc">
        Executing full Video-RAG pipeline: downloading audio, transcribing with Faster-Whisper, computing 384-dim FastEmbed embeddings, and syncing Neo4j Graph.
      </div>
      <div class="progress-bar-wrap">
        <div class="progress-bar-fill" id="modalProgressBar"></div>
      </div>
      <div class="step-list">
        <div class="step-item step-active" id="step1">✓ 1. Validating YouTube URL & Metadata</div>
        <div class="step-item" id="step2">⏳ 2. Downloading Audio Stream (yt-dlp)</div>
        <div class="step-item" id="step3">🎙️ 3. Transcribing Audio (Faster-Whisper)</div>
        <div class="step-item" id="step4">🧠 4. Computing 384-Dim FastEmbed Vectors</div>
        <div class="step-item" id="step5">🗄️ 5. Upserting Chunks into Qdrant</div>
        <div class="step-item" id="step6">🕸️ 6. Mapping Temporal Sequence in Neo4j</div>
      </div>
    </div>
  </div>

  <script>
    const data = %s;
    const videoId = data.video_id || 'Rr4CW2quazA';

    function init() {
      renderChunks();
      // Setup graph on load
      setTimeout(() => { fetchAndRenderGraph(); }, 300);
    }

    function switchTab(tab) {
      document.getElementById('tabChat').style.display = (tab === 'chat') ? 'flex' : 'none';
      document.getElementById('tabChunks').style.display = (tab === 'chunks') ? 'block' : 'none';
      document.getElementById('tabGraph').style.display = (tab === 'graph') ? 'block' : 'none';

      document.getElementById('tabBtnChat').classList.toggle('active', tab === 'chat');
      document.getElementById('tabBtnChunks').classList.toggle('active', tab === 'chunks');
      document.getElementById('tabBtnGraph').classList.toggle('active', tab === 'graph');

      if (tab === 'graph') {
        setTimeout(() => {
          resizeCanvas();
          if (graphNodes.length === 0) {
            fetchAndRenderGraph();
          } else {
            drawGraph();
            startSimulation();
          }
        }, 60);
      }
    }

    function toggleAudioPlayer() {
      const wrap = document.getElementById('audioPlayerWrap');
      const btn = document.getElementById('listenAudioBtn');
      if (wrap.style.display === 'flex') {
        wrap.style.display = 'none';
        btn.innerHTML = '▶️ Listen';
      } else {
        wrap.style.display = 'flex';
        btn.innerHTML = '⏸️ Close Player';
        const audio = document.getElementById('inlineAudioPlayer');
        if (audio) audio.play();
      }
    }

    // Concept pill color palette
    const tagStyles = [
      'tag-color-0', 'tag-color-1', 'tag-color-2',
      'tag-color-3', 'tag-color-4', 'tag-color-5'
    ];

    function getTagClass(str) {
      var hash = 0;
      for (var i = 0; i < str.length; i++) {
        hash = str.charCodeAt(i) + ((hash << 5) - hash);
      }
      var idx = Math.abs(hash) %% tagStyles.length;
      return tagStyles[idx];
    }

    // Render Chunks Tab
    function renderChunks() {
      const container = document.getElementById('chunksContainer');
      const chunks = data.chunks || [];
      if (chunks.length === 0) {
        container.innerHTML = '<div style="color: var(--text-muted); padding: 20px;">No chunks stored.</div>';
        return;
      }

      var html = '';
      chunks.forEach(function(c, i) {
        const preview = (c.embedding_preview || []).slice(0, 8).map(function(n) { return Number(n).toFixed(4); }).join(', ') + '...';
        const fullVectorStr = JSON.stringify(c.embedding_vector || c.embedding_preview || []);

        html += '<div class="chunk-card">' +
          '<div class="chunk-header">' +
            '<div class="chunk-meta">' +
              '<span class="chunk-idx">#' + c.chunk_index + '</span>' +
              '<span class="chunk-time">⏱️ ' + c.timestamp_display + '</span>' +
              '<span class="chunk-words">' + c.word_count + ' words</span>' +
            '</div>' +
            '<button onclick="toggleVector(' + i + ')" class="btn btn-outline btn-sm" id="vec-btn-' + i + '">' +
              '👁️ View 384-dim Vector' +
            '</button>' +
          '</div>' +
          '<div class="chunk-body">' + escapeHtml(c.text) + '</div>' +
          '<div style="font-size: 0.75rem; color: var(--text-muted); font-family: monospace;">' +
            '384-dim preview: [' + preview + ']' +
          '</div>' +
          '<div class="vector-panel" id="vec-panel-' + i + '">' +
            '<div style="display: flex; justify-content: space-between; margin-bottom: 6px;">' +
              '<span style="font-weight: 700; font-size: 0.725rem; color: var(--primary);">Full 384-Dimensional FastEmbed Array:</span>' +
              '<button onclick="copyVec(' + i + ')" class="btn btn-outline btn-sm" id="copy-btn-' + i + '">📋 Copy Vector</button>' +
            '</div>' +
            '<div class="vector-text" id="vec-val-' + i + '">' + fullVectorStr + '</div>' +
          '</div>' +
        '</div>';
      });

      container.innerHTML = html;
    }

    function toggleVector(idx) {
      const p = document.getElementById('vec-panel-' + idx);
      const b = document.getElementById('vec-btn-' + idx);
      if (p.style.display === 'block') {
        p.style.display = 'none';
        b.textContent = '👁️ View 384-dim Vector';
      } else {
        p.style.display = 'block';
        b.textContent = 'Hide Vector';
      }
    }

    let allVecsExpanded = false;
    function expandAllVectors() {
      allVecsExpanded = !allVecsExpanded;
      const total = (data.chunks || []).length;
      for (var i = 0; i < total; i++) {
        const p = document.getElementById('vec-panel-' + i);
        const b = document.getElementById('vec-btn-' + i);
        if (p && b) {
          p.style.display = allVecsExpanded ? 'block' : 'none';
          b.textContent = allVecsExpanded ? 'Hide Vector' : '👁️ View 384-dim Vector';
        }
      }
      document.getElementById('expandAllBtn').textContent = allVecsExpanded ? 'Collapse All' : 'Expand All Vectors';
    }

    function copyVec(idx) {
      const text = document.getElementById('vec-val-' + idx).textContent;
      navigator.clipboard.writeText(text).then(() => {
        const b = document.getElementById('copy-btn-' + idx);
        b.textContent = '✅ Copied!';
        setTimeout(() => { b.textContent = '📋 Copy Vector'; }, 2000);
      });
    }

    // ==========================================
    // INTERACTIVE NEO4J KNOWLEDGE GRAPH ENGINE
    // ==========================================
    let graphNodes = [];
    let graphEdges = [];
    let graphScale = 1;
    let graphPanX = 0;
    let graphPanY = 0;
    let draggingNode = null;
    let isPanning = false;
    let startPanX = 0;
    let startPanY = 0;
    let animFrame = null;

    async function fetchAndRenderGraph() {
      try {
        const res = await fetch('/api/v1/videos/' + videoId + '/graph');
        const json = await res.json();
        if (json.status !== 'success') return;

        resizeCanvas();
        buildGraphData(json);
        const badge = document.getElementById('graphConceptBadge');
        if (badge) badge.textContent = (json.concepts || []).length + ' Concepts';
        drawGraph();
        startSimulation();
      } catch (e) {
        console.error('Graph fetch error:', e);
      }
    }

    function buildGraphData(g) {
      graphNodes = [];
      graphEdges = [];

      const canvas = document.getElementById('graphCanvas');
      const w = (canvas && canvas.width > 50) ? canvas.width : 800;
      const h = (canvas && canvas.height > 50) ? canvas.height : 480;

      // 1. Root Video Node
      const rootNode = {
        id: 'video_' + (g.video_id || videoId),
        type: 'video',
        label: (g.title || 'Video').substring(0, 18) + '...',
        fullLabel: g.title,
        color: '#0284c7',
        textColor: '#ffffff',
        radius: 26,
        x: w / 2,
        y: h / 2,
        vx: 0,
        vy: 0
      };
      graphNodes.push(rootNode);

      // 2. Chunks (arranged sequentially in a circle)
      const chunks = g.chunks || [];
      const chunkMap = {};
      const numC = chunks.length;
      const radiusC = Math.min(w, h) * 0.32;

      chunks.forEach((c, idx) => {
        const angle = (idx / numC) * 2 * Math.PI;
        const node = {
          id: c.id,
          type: 'chunk',
          index: c.index,
          label: '#' + c.index,
          subLabel: c.time,
          text: c.text,
          time: c.time,
          color: '#38bdf8',
          textColor: '#0f172a',
          radius: 18,
          x: w / 2 + Math.cos(angle) * radiusC + (Math.random() - 0.5) * 20,
          y: h / 2 + Math.sin(angle) * radiusC + (Math.random() - 0.5) * 20,
          vx: 0,
          vy: 0
        };
        graphNodes.push(node);
        chunkMap[c.id] = node;

        // Video -> Chunk HAS_CHUNK
        graphEdges.push({ source: rootNode, target: node, type: 'HAS_CHUNK', color: '#bae6fd', width: 1.5 });
      });

      // 3. Temporal NEXT_CHUNK edges
      (g.temporal_sequences || []).forEach(seq => {
        if (seq.from && seq.to && chunkMap[seq.from] && chunkMap[seq.to]) {
          graphEdges.push({
            source: chunkMap[seq.from],
            target: chunkMap[seq.to],
            type: 'NEXT_CHUNK',
            color: '#0284c7',
            width: 2.5,
            arrow: true
          });
        }
      });

      // 4. Concepts (outer ring)
      const concepts = g.concepts || [];
      const conceptMap = {};
      const numCp = concepts.length;
      const radiusCp = Math.min(w, h) * 0.44;

      concepts.forEach((cp, idx) => {
        const angle = (idx / numCp) * 2 * Math.PI;
        const node = {
          id: 'concept_' + cp,
          type: 'concept',
          label: cp,
          color: '#a855f7',
          textColor: '#ffffff',
          radius: 14,
          x: w / 2 + Math.cos(angle) * radiusCp,
          y: h / 2 + Math.sin(angle) * radiusCp,
          vx: 0,
          vy: 0
        };
        graphNodes.push(node);
        conceptMap[cp] = node;
      });

      // 5. Concept Links DISCUSSES
      (g.concept_links || []).forEach(link => {
        if (chunkMap[link.chunk] && conceptMap[link.concept]) {
          graphEdges.push({
            source: chunkMap[link.chunk],
            target: conceptMap[link.concept],
            type: 'DISCUSSES',
            color: '#e9d5ff',
            width: 1,
            dashed: true
          });
        }
      });
    }

    function resizeCanvas() {
      const canvas = document.getElementById('graphCanvas');
      if (!canvas) return;
      const parent = canvas.parentElement;
      const w = (parent && parent.clientWidth > 50) ? parent.clientWidth : 800;
      const h = (parent && parent.clientHeight > 50) ? parent.clientHeight : 480;
      canvas.width = w;
      canvas.height = h;
    }

    function startSimulation() {
      if (animFrame) cancelAnimationFrame(animFrame);
      let stepCount = 0;

      function step() {
        stepCount++;
        // Gentle physics relaxation
        for (let i = 0; i < graphNodes.length; i++) {
          const n1 = graphNodes[i];
          if (n1 === draggingNode) continue;

          for (let j = i + 1; j < graphNodes.length; j++) {
            const n2 = graphNodes[j];
            const dx = n2.x - n1.x;
            const dy = n2.y - n1.y;
            const dist = Math.sqrt(dx * dx + dy * dy) || 1;
            const minDist = (n1.radius + n2.radius) * 2.2;
            if (dist < minDist) {
              const force = (minDist - dist) / dist * 0.12;
              n1.x -= dx * force;
              n1.y -= dy * force;
              if (n2 !== draggingNode) {
                n2.x += dx * force;
                n2.y += dy * force;
              }
            }
          }
        }

        drawGraph();
        if (stepCount < 180 || draggingNode) {
          animFrame = requestAnimationFrame(step);
        }
      }
      step();
    }

    function drawGraph() {
      const canvas = document.getElementById('graphCanvas');
      if (!canvas) return;
      const ctx = canvas.getContext('2d');
      ctx.clearRect(0, 0, canvas.width, canvas.height);

      ctx.save();
      ctx.translate(graphPanX, graphPanY);
      ctx.scale(graphScale, graphScale);

      // Draw Edges
      graphEdges.forEach(edge => {
        ctx.beginPath();
        ctx.moveTo(edge.source.x, edge.source.y);
        ctx.lineTo(edge.target.x, edge.target.y);
        ctx.strokeStyle = edge.color;
        ctx.lineWidth = edge.width;
        if (edge.dashed) {
          ctx.setLineDash([3, 4]);
        } else {
          ctx.setLineDash([]);
        }
        ctx.stroke();
        ctx.setLineDash([]);

        // Draw Arrow for NEXT_CHUNK
        if (edge.arrow) {
          const dx = edge.target.x - edge.source.x;
          const dy = edge.target.y - edge.source.y;
          const angle = Math.atan2(dy, dx);
          const headLen = 9;
          const tx = edge.target.x - Math.cos(angle) * (edge.target.radius + 3);
          const ty = edge.target.y - Math.sin(angle) * (edge.target.radius + 3);

          ctx.beginPath();
          ctx.moveTo(tx, ty);
          ctx.lineTo(tx - headLen * Math.cos(angle - Math.PI / 6), ty - headLen * Math.sin(angle - Math.PI / 6));
          ctx.lineTo(tx - headLen * Math.cos(angle + Math.PI / 6), ty - headLen * Math.sin(angle + Math.PI / 6));
          ctx.fillStyle = edge.color;
          ctx.fill();
        }
      });

      // Draw Nodes
      graphNodes.forEach(node => {
        ctx.beginPath();
        ctx.arc(node.x, node.y, node.radius, 0, 2 * Math.PI);
        ctx.fillStyle = node.color;
        ctx.fill();
        ctx.lineWidth = 2;
        ctx.strokeStyle = '#ffffff';
        ctx.stroke();

        // Node Label
        ctx.fillStyle = node.textColor || '#ffffff';
        ctx.font = 'bold ' + (node.type === 'concept' ? '10px' : '11px') + ' Plus Jakarta Sans, sans-serif';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText(node.label, node.x, node.y);
      });

      ctx.restore();
    }

    // Graph Canvas Mouse Interactions
    const canvasEl = document.getElementById('graphCanvas');
    canvasEl.addEventListener('mousedown', e => {
      const rect = canvasEl.getBoundingClientRect();
      const mx = (e.clientX - rect.left - graphPanX) / graphScale;
      const my = (e.clientY - rect.top - graphPanY) / graphScale;

      for (let i = graphNodes.length - 1; i >= 0; i--) {
        const n = graphNodes[i];
        const dx = mx - n.x;
        const dy = my - n.y;
        if (Math.sqrt(dx * dx + dy * dy) <= n.radius) {
          draggingNode = n;
          showNodeInspector(n);
          startSimulation();
          return;
        }
      }

      isPanning = true;
      startPanX = e.clientX - graphPanX;
      startPanY = e.clientY - graphPanY;
    });

    window.addEventListener('mousemove', e => {
      if (draggingNode) {
        const rect = canvasEl.getBoundingClientRect();
        draggingNode.x = (e.clientX - rect.left - graphPanX) / graphScale;
        draggingNode.y = (e.clientY - rect.top - graphPanY) / graphScale;
        drawGraph();
      } else if (isPanning) {
        graphPanX = e.clientX - startPanX;
        graphPanY = e.clientY - startPanY;
        drawGraph();
      }
    });

    window.addEventListener('mouseup', () => {
      draggingNode = null;
      isPanning = false;
    });

    canvasEl.addEventListener('wheel', e => {
      e.preventDefault();
      const zoomFactor = e.deltaY < 0 ? 1.1 : 0.9;
      zoomGraph(zoomFactor);
    });

    function zoomGraph(factor) {
      graphScale = Math.max(0.4, Math.min(3.0, graphScale * factor));
      drawGraph();
    }

    function resetGraphView() {
      graphScale = 1;
      graphPanX = 0;
      graphPanY = 0;
      drawGraph();
    }

    function showNodeInspector(node) {
      const drawer = document.getElementById('nodeInspector');
      const title = document.getElementById('inspectorTitle');
      const content = document.getElementById('inspectorContent');
      drawer.style.display = 'block';

      if (node.type === 'video') {
        title.innerHTML = '🎬 Video Node';
        content.innerHTML = '<div style="font-weight:700;color:#0f172a;margin-bottom:4px;">' + escapeHtml(node.fullLabel || node.label) + '</div>' +
          '<div style="color:var(--text-muted);font-size:0.75rem;">Root node connected to all chunks.</div>';
      } else if (node.type === 'chunk') {
        title.innerHTML = '📦 Chunk #' + node.index + ' (' + node.time + ')';
        content.innerHTML = '<div style="color:#334155;margin-bottom:8px;line-height:1.5;">' + escapeHtml(node.text) + '</div>' +
          '<button onclick="askQuestion(\'Explain what happens around chunk ' + node.id + '\')" class="btn btn-primary btn-sm" style="width:100%%;justify-content:center;">💬 Ask AI About This</button>';
      } else if (node.type === 'concept') {
        title.innerHTML = '🏷️ Concept: ' + escapeHtml(node.label);
        content.innerHTML = '<div style="color:#334155;margin-bottom:8px;">Extracted domain entity linked in Neo4j via <code>[:DISCUSSES]</code>.</div>' +
          '<button onclick="askQuestion(\'What does this video say about ' + escapeHtml(node.label) + '?\')" class="btn btn-primary btn-sm" style="width:100%%;justify-content:center;">💬 Ask About ' + escapeHtml(node.label) + '</button>';
      }
    }

    function closeInspector() {
      document.getElementById('nodeInspector').style.display = 'none';
    }

    // Live Cypher Runner
    async function runCustomCypher() {
      const qInput = document.getElementById('cypherInput');
      const q = qInput.value.trim();
      if (!q) return;

      const btn = document.getElementById('runCypherBtn');
      const box = document.getElementById('cypherResultBox');
      btn.disabled = true;
      box.style.display = 'block';
      box.innerHTML = 'Running Cypher query in Neo4j...';

      try {
        const res = await fetch('/api/v1/neo4j/query', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ query: q })
        });
        const json = await res.json();

        if (json.status !== 'success') {
          box.innerHTML = '<span style="color:#f87171;">Error: ' + escapeHtml(json.message || 'Query failed') + '</span>';
          return;
        }

        const results = (json.result && json.result.results) || [];
        if (results.length === 0 || !results[0].data || results[0].data.length === 0) {
          box.innerHTML = '<span style="color:#38bdf8;">Query executed successfully. (0 rows returned)</span>';
          return;
        }

        const cols = results[0].columns || [];
        const rows = results[0].data || [];

        let tableHtml = '<div style="color:#38bdf8;font-weight:700;margin-bottom:6px;">Result: ' + rows.length + ' rows</div>' +
          '<table style="width:100%%;border-collapse:collapse;font-size:0.75rem;color:#e2e8f0;">' +
            '<thead><tr style="border-bottom:1px solid #334155;text-align:left;">' +
              cols.map(c => '<th style="padding:4px 6px;">' + escapeHtml(c) + '</th>').join('') +
            '</tr></thead><tbody>';

        rows.forEach(r => {
          tableHtml += '<tr style="border-bottom:1px solid #1e293b;">' +
            (r.row || []).map(val => '<td style="padding:4px 6px;">' + escapeHtml(typeof val === 'object' ? JSON.stringify(val) : String(val)) + '</td>').join('') +
          '</tr>';
        });

        tableHtml += '</tbody></table>';
        box.innerHTML = tableHtml;
      } catch (err) {
        box.innerHTML = '<span style="color:#f87171;">Request error: ' + escapeHtml(err.message) + '</span>';
      } finally {
        btn.disabled = false;
      }
    }

    // Chatbot Execution
    function askQuestion(q) {
      switchTab('chat');
      document.getElementById('chatInput').value = q;
      sendChatMessage();
    }

    async function sendChatMessage() {
      const input = document.getElementById('chatInput');
      const q = input.value.trim();
      if (!q) return;

      const container = document.getElementById('chatMessages');
      const sendBtn = document.getElementById('sendBtn');

      // Append User message
      container.innerHTML += '<div class="msg msg-user">' +
        '<div class="msg-avatar avatar-user">You</div>' +
        '<div class="msg-bubble">' + escapeHtml(q) + '</div>' +
      '</div>';
      input.value = '';
      container.scrollTop = container.scrollHeight;

      // Loading indicator
      const loaderId = 'loader-' + Date.now();
      container.innerHTML += '<div class="msg msg-bot" id="' + loaderId + '">' +
        '<div class="msg-avatar avatar-bot">AI</div>' +
        '<div class="msg-bubble" style="color: var(--text-muted);">' +
          '⚡ Retrieving grounded chunks & synthesizing answer with Ollama (Local LLM)...' +
        '</div>' +
      '</div>';
      container.scrollTop = container.scrollHeight;

      sendBtn.disabled = true;

      try {
        const res = await fetch('/api/v1/videos/' + videoId + '/chat', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ query: q, video_id: videoId })
        });
        const json = await res.json();

        const loader = document.getElementById(loaderId);
        if (loader) loader.remove();

        if (json.status !== 'success' || !json.answer) {
          container.innerHTML += '<div class="msg msg-bot">' +
            '<div class="msg-avatar avatar-bot">AI</div>' +
            '<div class="msg-bubble" style="color: #b91c1c;">Sorry, could not find matching information in this video.</div>' +
          '</div>';
          return;
        }

        // Render Bot response with Highlights, Neo4j Concept pills, and Action buttons
        var highlightsHtml = '';
        if (json.highlights && json.highlights.length > 0) {
          highlightsHtml = '<div class="highlights-box">' +
            '<div class="highlights-title">💡 Key Highlights</div>' +
            '<ul class="highlights-list">' +
              json.highlights.map(function(h) { return '<li>' + escapeHtml(h) + '</li>'; }).join('') +
            '</ul>' +
          '</div>';
        }

        var conceptsHtml = '';
        if (json.concepts && json.concepts.length > 0) {
          conceptsHtml = '<div class="concepts-row">' +
            '<span class="concepts-label">🏷️ Neo4j Related Concepts:</span>' +
            json.concepts.map(function(cp) {
              return '<span class="concept-tag ' + getTagClass(cp) + '">' + escapeHtml(cp) + '</span>';
            }).join('') +
          '</div>';
        }

        var actionsHtml = '<div class="msg-actions">';
        if (json.youtube_url) {
          actionsHtml += '<a href="' + json.youtube_url + '" target="_blank" class="yt-btn">▶ Watch at ' + json.timestamp + '</a>';
        }
        if (json.source_chunk_id && json.source_chunk_id !== 'summary_multi_chunks') {
          actionsHtml += '<button onclick="toggleBotGraphContext(\'' + loaderId + '-ctx\', \'' + json.source_chunk_id + '\')" class="graph-expand-btn">🕸️ Graph Context</button>';
        }
        actionsHtml += '</div>';

        var contextDrawer = '<div class="graph-context-drawer" id="' + loaderId + '-ctx"></div>';

        var ragBadge = '';
        if (json.rag_source) {
          ragBadge = '<div style="font-size:0.7rem; font-weight:700; color:#0284c7; margin-bottom:8px; display:inline-flex; align-items:center; gap:4px; background:#e0f2fe; padding:2px 8px; border-radius:12px; border:1px solid #bae6fd;">🦙 ' + escapeHtml(json.rag_source) + '</div>';
        }

        container.innerHTML += '<div class="msg msg-bot">' +
          '<div class="msg-avatar avatar-bot">AI</div>' +
          '<div class="msg-bubble">' +
            ragBadge +
            '<div style="line-height:1.6;">' + renderMarkdownText(json.answer) + '</div>' +
            highlightsHtml +
            conceptsHtml +
            actionsHtml +
            contextDrawer +
          '</div>' +
        '</div>';

        container.scrollTop = container.scrollHeight;
      } catch (err) {
        const loader = document.getElementById(loaderId);
        if (loader) loader.remove();
        container.innerHTML += '<div class="msg msg-bot"><div class="msg-avatar avatar-bot">AI</div><div class="msg-bubble" style="color:#b91c1c;">Error: ' + err.message + '</div></div>';
      } finally {
        sendBtn.disabled = false;
      }
    }

    async function toggleBotGraphContext(boxId, chunkId) {
      const box = document.getElementById(boxId);
      if (!box) return;
      if (box.style.display === 'block') {
        box.style.display = 'none';
        return;
      }
      box.style.display = 'block';
      box.innerHTML = '<span style="color: var(--text-muted);">Traversing Neo4j for temporal neighbors...</span>';

      try {
        const res = await fetch('/api/v1/videos/' + videoId + '/chunks/' + chunkId + '/context');
        const json = await res.json();
        if (json.status !== 'success') {
          box.innerHTML = 'No graph context found.';
          return;
        }

        var prevTxt = json.prev_chunk ? '<div><strong style="color: var(--primary);">⏮️ Previous Chunk (' + json.prev_chunk.timestamp_display + '):</strong> ' + escapeHtml(json.prev_chunk.text) + '</div>' : '';
        var nextTxt = json.next_chunk ? '<div style="margin-top: 6px;"><strong style="color: var(--primary);">⏭️ Next Chunk (' + json.next_chunk.timestamp_display + '):</strong> ' + escapeHtml(json.next_chunk.text) + '</div>' : '';

        box.innerHTML = '<div style="font-weight: 700; color: #0f172a; margin-bottom: 4px;">🕸️ Neo4j Temporal Context Window:</div>' + prevTxt + nextTxt;
      } catch (e) {
        box.innerHTML = 'Error: ' + e.message;
      }
    }

    function clearChat() {
      document.getElementById('chatMessages').innerHTML = '<div class="msg msg-bot"><div class="msg-avatar avatar-bot">AI</div><div class="msg-bubble">Chat history cleared. Ask another question about the video!</div></div>';
    }

    // TOP URL EXTRACTION LOGIC
    async function extractVideo() {
      const urlInput = document.getElementById('videoUrlInput');
      const url = urlInput.value.trim();
      if (!url) {
        alert('Please paste a valid YouTube video URL or ID.');
        return;
      }

      const modal = document.getElementById('processingModal');
      const extractBtn = document.getElementById('extractBtn');
      const progressBar = document.getElementById('modalProgressBar');
      modal.style.display = 'flex';
      extractBtn.disabled = true;

      // Animate progress steps
      var pct = 15;
      progressBar.style.width = pct + '%%';

      const timer = setInterval(() => {
        if (pct < 85) {
          pct += 8;
          progressBar.style.width = pct + '%%';
          if (pct >= 30) document.getElementById('step2').className = 'step-item step-active';
          if (pct >= 50) document.getElementById('step3').className = 'step-item step-active';
          if (pct >= 70) document.getElementById('step4').className = 'step-item step-active';
        }
      }, 2500);

      try {
        const response = await fetch('/api/v1/videos/process', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ url: url })
        });

        clearInterval(timer);
        progressBar.style.width = '100%%';
        document.getElementById('step5').className = 'step-item step-active';
        document.getElementById('step6').className = 'step-item step-active';

        const result = await response.json();
        if (result.status === 'success' && result.redirect_url) {
          window.location.href = result.redirect_url;
        } else {
          alert('Extraction failed: ' + (result.message || 'Unknown error'));
          modal.style.display = 'none';
          extractBtn.disabled = false;
        }
      } catch (e) {
        clearInterval(timer);
        alert('Extraction error: ' + e.message);
        modal.style.display = 'none';
        extractBtn.disabled = false;
      }
    }

    function escapeHtml(str) {
      if (!str) return '';
      return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
    }

    function renderMarkdownText(txt) {
      if (!txt) return '';
      let s = escapeHtml(txt);
      // Bold
      s = s.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
      // Italic
      s = s.replace(/\*(.*?)\*/g, '<em>$1</em>');
      // Headings
      s = s.replace(/^### (.*$)/gim, '<div style="font-weight:700;font-size:0.95rem;margin:8px 0 4px;color:#0369a1;">$1</div>');
      s = s.replace(/^## (.*$)/gim, '<div style="font-weight:700;font-size:1rem;margin:10px 0 6px;color:#0284c7;">$1</div>');
      // Bullet points
      s = s.replace(/^[•\-\*]\s+(.*)$/gim, '<div style="margin-left:14px;margin-bottom:4px;display:flex;align-items:flex-start;"><span style="color:#0284c7;margin-right:6px;">•</span><span>$1</span></div>');
      // Line breaks
      s = s.replace(/\n\n/g, '<div style="margin-top:8px;"></div>').replace(/\n/g, '<br>');
      return s;
    }

    init();
  </script>
</body>
</html>`,
		res.TotalChunks,
		res.VideoID,
		res.Title,
		res.VideoID,
		res.CollectionName,
		res.TotalChunks,
		res.VideoID,
		res.VideoID,
		res.VideoID,
		res.VideoID,
		res.TotalChunks,
		res.VideoID,
		safeJSON,
	)
}
