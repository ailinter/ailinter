package knowledge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"
)

// VisualizeOptions controls the visualization output.
type VisualizeOptions struct {
	// OutputPath is the path for the HTML file.
	// If empty, defaults to KnowledgeDir/visualize.html.
	OutputPath string

	// Layout algorithm: "cose", "breadthfirst", "concentric", "circle", "grid".
	// Default: "cose" (built-in force-directed, no plugin needed).
	Layout string

	// Title is the HTML page title.
	Title string

	// OpenBrowser attempts to open the browser after generation (macOS: "open").
	OpenBrowser bool
}

// GenerateVisualization converts the graph to a Cytoscape.js elements array and
// writes a self-contained HTML file. Returns the file path.
func (g *Graph) GenerateVisualization(opts VisualizeOptions) (string, error) {
	if opts.Layout == "" {
		opts.Layout = "cose"
	}
	if opts.Title == "" {
		opts.Title = "AILINTER Knowledge Graph"
	}

	outputPath := opts.OutputPath
	if outputPath == "" {
		outputPath = filepath.Join(g.KnowledgeDir, "visualize.html")
	}

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	// Convert graph to Cytoscape elements
	elements, nodeCounts, edgeCounts, totalNodes, totalEdges := g.toCytoscapeElements()

	// Marshal elements to JSON
	elementsJSON, err := json.Marshal(elements)
	if err != nil {
		return "", fmt.Errorf("marshal elements: %w", err)
	}

	// Build the HTML using the embedded template
	tmpl := template.Must(template.New("viz").Parse(htmlTemplate))

	var buf bytes.Buffer
	data := vizTemplateData{
		Title:        opts.Title,
		LayoutName:   opts.Layout,
		ElementsJSON: template.JS(elementsJSON),
		TotalNodes:   totalNodes,
		TotalEdges:   totalEdges,
		NodeCounts:   nodeCounts,
		EdgeCounts:   edgeCounts,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("write HTML: %w", err)
	}

	if opts.OpenBrowser {
		openBrowser(outputPath)
	}

	return outputPath, nil
}

// vizTemplateData holds the data passed to the HTML template.
type vizTemplateData struct {
	Title        string
	LayoutName   string
	ElementsJSON template.JS
	TotalNodes   int
	TotalEdges   int
	NodeCounts   map[string]int
	EdgeCounts   map[string]int
}

// StartVisualizationServer starts an ephemeral HTTP server serving the
// visualization HTML. Returns the URL and a cancel function.
// The server auto-shuts down after 10 minutes of inactivity.
func (g *Graph) StartVisualizationServer(port int) (string, context.CancelFunc, error) {
	outputPath := filepath.Join(g.KnowledgeDir, "visualize.html")
	if err := ensureVisualizationExists(g, outputPath); err != nil {
		return "", nil, err
	}

	listener, err := listenOnPort(port)
	if err != nil {
		return "", nil, fmt.Errorf("listen: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(filepath.Dir(outputPath))))

	server := &http.Server{}
	var timerMu sync.Mutex
	var timer *time.Timer
	inactivityReset := func() {
		timerMu.Lock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(10*time.Minute, func() {
			server.Close()
		})
		timerMu.Unlock()
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inactivityReset()
		mux.ServeHTTP(w, r)
	})
	server.Handler = handler
	inactivityReset()

	go server.Serve(listener)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-ctx.Done()
		server.Close()
	}()

	addr := listener.Addr().(*net.TCPAddr)
	return fmt.Sprintf("http://localhost:%d", addr.Port), cancel, nil
}

// ensureVisualizationExists generates the visualization HTML if it doesn't already exist.
func ensureVisualizationExists(g *Graph, outputPath string) error {
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		return nil
	}
	_, err := g.GenerateVisualization(VisualizeOptions{OutputPath: outputPath})
	return err
}

// listenOnPort tries to listen on the given port, falling back to a free port.
func listenOnPort(port int) (net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		listener, err = net.Listen("tcp", ":0")
	}
	return listener, err
}

// toCytoscapeElements converts the graph to a JSON-serializable array
// of Cytoscape.js elements (nodes and edges).
// Returns the elements, node type counts, edge type counts, total nodes, total edges.
func (g *Graph) toCytoscapeElements() ([]map[string]interface{}, map[string]int, map[string]int, int, int) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	elements := make([]map[string]interface{}, 0, len(g.Nodes)+len(g.EdgesOut)*2)
	nodeTypeCounts := make(map[string]int)

	elements = append(elements, buildNodeElements(g, &nodeTypeCounts)...)

	edgeElements, edgeTypeCounts, uniqueEdgeCount := buildEdgeElements(g)
	elements = append(elements, edgeElements...)

	return elements, nodeTypeCounts, edgeTypeCounts, len(g.Nodes), uniqueEdgeCount
}

// buildNodeElements creates Cytoscape node elements from the graph.
func buildNodeElements(g *Graph, nodeTypeCounts *map[string]int) []map[string]interface{} {
	var elements []map[string]interface{}
	for _, id := range sortedNodeIDs(g) {
		n := g.Nodes[id]
		data := map[string]interface{}{
			"id":    n.ID,
			"label": n.Label,
			"type":  string(n.Type),
		}
		for k, v := range n.Properties {
			data[k] = v
		}
		elements = append(elements, map[string]interface{}{"data": data})
		(*nodeTypeCounts)[string(n.Type)]++
	}
	return elements
}

// buildEdgeElements creates Cytoscape edge elements from the graph, deduplicating.
func buildEdgeElements(g *Graph) ([]map[string]interface{}, map[string]int, int) {
	seenEdges := make(map[string]bool)
	edgeTypeCounts := make(map[string]int)
	var edgeKeys []string
	edgeMap := make(map[string]Edge)

	for _, outs := range g.EdgesOut {
		for _, e := range outs {
			key := e.From + "→" + e.To + "→" + string(e.Type)
			if seenEdges[key] {
				continue
			}
			seenEdges[key] = true
			edgeKeys = append(edgeKeys, key)
			edgeMap[key] = e
			edgeTypeCounts[string(e.Type)]++
		}
	}

	sort.Strings(edgeKeys)

	var elements []map[string]interface{}
	for _, key := range edgeKeys {
		e := edgeMap[key]
		elements = append(elements, map[string]interface{}{
			"data": map[string]interface{}{
				"id":     key,
				"source": e.From,
				"target": e.To,
				"label":  string(e.Type),
				"type":   string(e.Type),
			},
		})
	}
	return elements, edgeTypeCounts, len(seenEdges)
}

// openBrowser attempts to open the given file or URL in the default browser.
func openBrowser(path string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", path).Start()
	case "linux":
		exec.Command("xdg-open", path).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
	}
}

// htmlTemplate is the self-contained Cytoscape.js visualization template.
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.Title}}</title>
<script src="https://unpkg.com/cytoscape@3.30/dist/cytoscape.min.js"></script>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
html, body { height: 100%; overflow: hidden; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #1a1a2e; color: #e0e0e0; }

#cy {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  left: 280px;
}

#sidebar {
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  width: 280px;
  background: rgba(26, 26, 46, 0.95);
  border-right: 1px solid rgba(255,255,255,0.08);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  z-index: 10;
}

#sidebar-header {
  padding: 16px;
  border-bottom: 1px solid rgba(255,255,255,0.08);
}

#sidebar-header h1 {
  font-size: 16px;
  font-weight: 700;
  color: #8B5CF6;
  margin-bottom: 4px;
}

#sidebar-header .stats {
  font-size: 12px;
  color: #888;
}

#search-box {
  margin: 12px 16px;
  padding: 8px 12px;
  background: rgba(255,255,255,0.06);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 6px;
  color: #e0e0e0;
  font-size: 13px;
  outline: none;
  transition: border-color 0.2s;
}
#search-box:focus {
  border-color: #8B5CF6;
}
#search-box::placeholder { color: #666; }

#filters {
  flex: 1;
  overflow-y: auto;
  padding: 0 16px 16px;
}

.filter-section {
  margin-bottom: 12px;
}

.filter-section h3 {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: #666;
  margin-bottom: 6px;
}

.filter-label {
  display: flex;
  align-items: center;
  padding: 4px 0;
  font-size: 12px;
  cursor: pointer;
  user-select: none;
}

.filter-label input[type="checkbox"] {
  appearance: none;
  -webkit-appearance: none;
  width: 14px;
  height: 14px;
  border: 1.5px solid #555;
  border-radius: 3px;
  margin-right: 8px;
  cursor: pointer;
  position: relative;
  flex-shrink: 0;
}

.filter-label input[type="checkbox"]:checked {
  border-color: #8B5CF6;
  background: #8B5CF6;
}

.filter-label input[type="checkbox"]:checked::after {
  content: "✓";
  position: absolute;
  top: -1px;
  left: 2px;
  font-size: 10px;
  color: #fff;
}

.filter-label .count {
  color: #666;
  margin-left: auto;
  font-size: 11px;
}

.color-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;
  flex-shrink: 0;
}

#inspector {
  border-top: 1px solid rgba(255,255,255,0.08);
  padding: 12px 16px;
  max-height: 200px;
  overflow-y: auto;
  font-size: 12px;
  flex-shrink: 0;
  min-height: 50px;
}

#inspector h3 {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 4px;
}

#inspector p {
  font-size: 11px;
  color: #888;
  margin-bottom: 4px;
  word-break: break-all;
}

#inspector pre {
  font-size: 10px;
  color: #aaa;
  background: rgba(0,0,0,0.2);
  padding: 8px;
  border-radius: 4px;
  max-height: 80px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  margin-top: 4px;
}

#inspector .empty-message {
  color: #666;
  font-style: italic;
  font-size: 12px;
}

.edge-type-badge {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 10px;
  font-weight: 600;
  margin-right: 4px;
}

::-webkit-scrollbar { width: 6px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.1); border-radius: 3px; }
</style>
</head>
<body>

<div id="sidebar">
  <div id="sidebar-header">
    <h1>AILINTER Knowledge Graph</h1>
    <div class="stats">{{.TotalNodes}} Nodes · {{.TotalEdges}} Edges</div>
  </div>
  <input id="search-box" type="text" placeholder="Search nodes..." spellcheck="false">
  <div id="filters">
    <div class="filter-section">
      <h3>Node Types</h3>
      <div id="node-filters"></div>
    </div>
    <div class="filter-section">
      <h3>Edge Types</h3>
      <div id="edge-filters"></div>
    </div>
  </div>
  <div id="inspector">
    <div class="empty-message">Click a node to inspect</div>
  </div>
</div>

<div id="cy"></div>

<script>
const DATA = {{.ElementsJSON}};

const nodeStyleByType = {
  'agent':     { shape: 'hexagon',       color: '#8B5CF6', label: 'Agent' },
  'bug':       { shape: 'triangle',       color: '#EF4444', label: 'Bug' },
  'commit':    { shape: 'diamond',        color: '#F59E0B', label: 'Commit' },
  'file':      { shape: 'rectangle',     color: '#06B6DB', label: 'File' },
  'function':  { shape: 'ellipse',        color: '#10B981', label: 'Function' },
  'package':   { shape: 'round-rectangle', color: '#6366F1', label: 'Package' },
  'spec':      { shape: 'round-diamond',  color: '#EC4899', label: 'Spec' },
  'test':      { shape: 'vee',            color: '#14B8A6', label: 'Test' },
};

const edgeStyleByType = {
  'CALLS':      { color: '#10B981', style: 'bezier',     width: 1,   dash: 'solid' },
  'CONTAINS':   { color: '#6366F1', style: 'bezier',     width: 2,   dash: 'solid' },
  'OWNS':       { color: '#8B5CF6', style: 'bezier',     width: 1.5, dash: 'dashed' },
  'IMPORTS':    { color: '#F59E0B', style: 'bezier',     width: 1,   dash: 'dotted' },
  'CAUSED_BY':  { color: '#EF4444', style: 'bezier',     width: 2,   dash: 'dashed' },
  'TESTS':      { color: '#14B8A6', style: 'bezier',     width: 1,   dash: 'solid' },
  'IMPLEMENTS': { color: '#EC4899', style: 'bezier',     width: 0.5, dash: 'solid' },
  'REFERENCES': { color: '#888',    style: 'bezier',     width: 0.5, dash: 'dotted' },
  'DEPENDS_ON': { color: '#888',    style: 'bezier',     width: 0.5, dash: 'dashed' },
  'AFFECTS':    { color: '#EF4444', style: 'bezier',     width: 0.5, dash: 'dotted' },
  'CHANGED':    { color: '#F59E0B', style: 'bezier',     width: 0.5, dash: 'solid' },
};

function getDefaultNodeStyle(type) {
  const s = nodeStyleByType[type];
  if (s) return s;
  return { shape: 'ellipse', color: '#888', label: type };
}

function getDefaultEdgeStyle(type) {
  const s = edgeStyleByType[type];
  if (s) return s;
  return { color: '#888', style: 'bezier', width: 0.5, dash: 'solid' };
}

// Compute counts from DATA
const nodeTypeCounts = {};
const edgeTypeCounts = {};
DATA.forEach(function(el) {
  if (el.data.source === undefined) {
    const t = el.data.type || 'unknown';
    nodeTypeCounts[t] = (nodeTypeCounts[t] || 0) + 1;
  } else {
    const t = el.data.type || 'unknown';
    edgeTypeCounts[t] = (edgeTypeCounts[t] || 0) + 1;
  }
});

// Build filter UI
const nodeFilterContainer = document.getElementById('node-filters');
const edgeFilterContainer = document.getElementById('edge-filters');

const allNodeTypes = Object.keys(nodeTypeCounts).sort();
const allEdgeTypes = Object.keys(edgeTypeCounts).sort();

const nodeFilterState = {};
const edgeFilterState = {};

allNodeTypes.forEach(function(t) {
  nodeFilterState[t] = true;
  const label = document.createElement('label');
  label.className = 'filter-label';
  const ns = getDefaultNodeStyle(t);
  const colorDot = '<span class="color-dot" style="background:' + ns.color + '"></span>';
  label.innerHTML = '<input type="checkbox" checked data-type="' + t + '" data-kind="node">'
    + colorDot + ns.label + ' <span class="count">' + nodeTypeCounts[t] + '</span>';
  label.querySelector('input').addEventListener('change', function(e) {
    nodeFilterState[t] = e.target.checked;
    applyFilters();
  });
  nodeFilterContainer.appendChild(label);
});

allEdgeTypes.forEach(function(t) {
  edgeFilterState[t] = true;
  const label = document.createElement('label');
  label.className = 'filter-label';
  const es = getDefaultEdgeStyle(t);
  const colorDot = '<span class="color-dot" style="background:' + es.color + '"></span>';
  label.innerHTML = '<input type="checkbox" checked data-type="' + t + '" data-kind="edge">'
    + colorDot + t + ' <span class="count">' + edgeTypeCounts[t] + '</span>';
  label.querySelector('input').addEventListener('change', function(e) {
    edgeFilterState[t] = e.target.checked;
    applyFilters();
  });
  edgeFilterContainer.appendChild(label);
});

// Initialize Cytoscape
var cy = cytoscape({
  container: document.getElementById('cy'),
  elements: DATA,
  style: generateStyles(),
  layout: { name: '{{.LayoutName}}', gravity: 0.25, idealEdgeLength: 100, nodeRepulsion: 8000 },
  minZoom: 0.03,
  maxZoom: 5,
  hideEdgesOnViewport: true,
  wheelSensitivity: 0.3,
});

function generateStyles() {
  const styles = [
    { selector: 'node', style: { 'background-color': '#888', label: 'data(label)', 'color': '#e0e0e0', 'font-size': '10px', 'text-valign': 'bottom', 'text-halign': 'center', 'text-margin-y': 4, 'border-width': 0, 'min-zoomed-font-size': 6 } },
    { selector: 'edge', style: { 'curve-style': 'bezier', 'target-arrow-shape': 'triangle', 'arrow-scale': 0.5, 'line-color': '#888', 'target-arrow-color': '#888', 'width': 0.5 } },
    { selector: 'edge:selected', style: { 'line-color': '#fff', 'target-arrow-color': '#fff', 'width': 2 } },
  ];

  Object.keys(nodeStyleByType).forEach(function(t) {
    const ns = nodeStyleByType[t];
    styles.push({ selector: 'node[type = "' + t + '"]', style: { 'background-color': ns.color, shape: ns.shape } });
  });

  Object.keys(edgeStyleByType).forEach(function(t) {
    const es = edgeStyleByType[t];
    let lineStyle = 'solid';
    if (es.dash === 'dashed') lineStyle = 'dashed';
    else if (es.dash === 'dotted') lineStyle = 'dotted';
    styles.push({
      selector: 'edge[label = "' + t + '"]',
      style: { 'line-color': es.color, 'target-arrow-color': es.color, width: es.width, 'line-style': lineStyle }
    });
  });

  styles.push(
    { selector: 'node.highlighted', style: { opacity: 1 } },
    { selector: 'node.dimmed', style: { opacity: 0.1 } },
    { selector: 'edge.dimmed', style: { opacity: 0.05 } },
    { selector: 'edge.highlighted', style: { opacity: 0.6 } }
  );

  return styles;
}

// Search functionality
var searchTimeout = null;
document.getElementById('search-box').addEventListener('input', function(e) {
  if (searchTimeout) clearTimeout(searchTimeout);
  searchTimeout = setTimeout(function() {
    const query = e.target.value.trim().toLowerCase();
    if (!query) {
      cy.elements().removeClass('highlighted dimmed');
      return;
    }
    cy.nodes().forEach(function(node) {
      const label = (node.data('label') || '').toLowerCase();
      if (label.indexOf(query) !== -1) {
        node.removeClass('dimmed').addClass('highlighted');
      } else {
        node.removeClass('highlighted').addClass('dimmed');
      }
    });
    cy.edges().forEach(function(edge) {
      const src = edge.source();
      const tgt = edge.target();
      if (src.hasClass('highlighted') || tgt.hasClass('highlighted')) {
        edge.removeClass('dimmed').addClass('highlighted');
      } else {
        edge.removeClass('highlighted').addClass('dimmed');
      }
    });
  }, 150);
});

// Filter functionality
function applyFilters() {
  cy.startBatch();

  cy.nodes().forEach(function(node) {
    const t = node.data('type') || 'unknown';
    if (nodeFilterState[t] === undefined || nodeFilterState[t]) {
      node.show();
    } else {
      node.hide();
    }
  });

  cy.edges().forEach(function(edge) {
    const t = edge.data('type') || 'unknown';
    if (edgeFilterState[t] === undefined || edgeFilterState[t]) {
      // Also check if both endpoints are visible
      const src = edge.source();
      const tgt = edge.target();
      if (src.visible() && tgt.visible()) {
        edge.show();
      } else {
        edge.hide();
      }
    } else {
      edge.hide();
    }
  });

  cy.endBatch();
}

// Click-to-inspect
const inspector = document.getElementById('inspector');
cy.on('tap', 'node', function(evt) {
  const node = evt.target;
  const data = node.data();
  const connectedEdges = node.connectedEdges().length;

  let html = '<h3>' + (data.type || 'unknown') + ': ' + (data.label || '') + '</h3>';
  html += '<p>ID: ' + (data.id || '') + '</p>';
  html += '<p>Connected edges: ' + connectedEdges + '</p>';
  // Show edge type badges
  const edgeTypes = {};
  node.connectedEdges().forEach(function(e) {
    const t = e.data('type') || 'unknown';
    edgeTypes[t] = (edgeTypes[t] || 0) + 1;
  });
  if (Object.keys(edgeTypes).length > 0) {
    html += '<p>';
    Object.keys(edgeTypes).sort().forEach(function(t) {
      const es = getDefaultEdgeStyle(t);
      html += '<span class="edge-type-badge" style="background:' + es.color + '22; color:' + es.color + '; border:1px solid ' + es.color + '44">' + t + ' ' + edgeTypes[t] + '</span> ';
    });
    html += '</p>';
  }
  // Omit large Properties like "content" to keep inspector compact
  const displayData = {};
  Object.keys(data).forEach(function(k) {
    if (k === 'id' || k === 'label' || k === 'type') return;
    const v = data[k];
    if (typeof v === 'string' && v.length > 200) {
      displayData[k] = v.substring(0, 200) + '...';
    } else {
      displayData[k] = v;
    }
  });
  if (Object.keys(displayData).length > 0) {
    html += '<pre>' + JSON.stringify(displayData, null, 2) + '</pre>';
  }
  inspector.innerHTML = html;
});
</script>
</body>
</html>`

// ensure htmlTemplate is used
var _ = htmlTemplate
