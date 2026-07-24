#!/usr/bin/env python3
"""
Convert semgrep --json output into a simple HTML report.

Newer versions of Semgrep removed the built-in --html flag, so this fills
that gap: run semgrep with --json, then feed the result through this script.

Usage:
    semgrep --config=p/golang . --json --output=semgrep-baseline.json
    python semgrep_report.py semgrep-baseline.json semgrep-baseline.html
"""
import json
import sys
from collections import Counter
from html import escape


def load_results(json_path):
    with open(json_path, "r", encoding="utf-8") as f:
        data = json.load(f)
    return data.get("results", []), data.get("errors", [])


def severity_of(finding):
    return finding.get("extra", {}).get("severity", "UNKNOWN").upper()


def render_html(results, errors, project_name):
    counts = Counter(severity_of(r) for r in results)
    total = len(results)

    severity_order = ["ERROR", "WARNING", "INFO", "UNKNOWN"]
    severity_colors = {
        "ERROR": "#c0392b",
        "WARNING": "#d68910",
        "INFO": "#2471a3",
        "UNKNOWN": "#7f8c8d",
    }

    summary_cells = "".join(
        f'<div class="stat" style="border-left:4px solid {severity_colors[s]}">'
        f'<div class="stat-num">{counts.get(s, 0)}</div>'
        f'<div class="stat-label">{s}</div></div>'
        for s in severity_order
        if counts.get(s, 0) > 0
    )

    rows = ""
    for r in sorted(results, key=lambda x: severity_order.index(severity_of(x)) if severity_of(x) in severity_order else 99):
        sev = severity_of(r)
        color = severity_colors.get(sev, "#7f8c8d")
        path = escape(r.get("path", ""))
        line = r.get("start", {}).get("line", "?")
        check_id = escape(r.get("check_id", ""))
        message = escape(r.get("extra", {}).get("message", "").strip())
        rows += f"""
        <tr>
          <td><span class="badge" style="background:{color}">{sev}</span></td>
          <td>{path}:{line}</td>
          <td class="check-id">{check_id}</td>
          <td>{message}</td>
        </tr>"""

    error_note = ""
    if errors:
        error_note = f'<p class="errors">{len(errors)} scan error(s) occurred — see the JSON file for details.</p>'

    return f"""<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Semgrep Report — {escape(project_name)}</title>
<style>
  body {{ font-family: -apple-system, Segoe UI, Arial, sans-serif; margin: 2rem; background: #f7f7f8; color: #1a1a1a; }}
  h1 {{ margin-bottom: 0.2rem; }}
  .subtitle {{ color: #666; margin-top: 0; }}
  .summary {{ display: flex; gap: 1rem; margin: 1.5rem 0; flex-wrap: wrap; }}
  .stat {{ background: white; padding: 0.75rem 1.25rem; border-radius: 6px; box-shadow: 0 1px 3px rgba(0,0,0,0.08); }}
  .stat-num {{ font-size: 1.6rem; font-weight: 700; }}
  .stat-label {{ font-size: 0.8rem; color: #666; letter-spacing: 0.05em; }}
  table {{ width: 100%; border-collapse: collapse; background: white; border-radius: 6px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,0.08); }}
  th, td {{ text-align: left; padding: 0.6rem 0.9rem; border-bottom: 1px solid #eee; vertical-align: top; }}
  th {{ background: #fafafa; font-size: 0.8rem; text-transform: uppercase; color: #666; }}
  .badge {{ color: white; padding: 0.15rem 0.5rem; border-radius: 4px; font-size: 0.75rem; font-weight: 600; }}
  .check-id {{ font-family: monospace; font-size: 0.85rem; color: #444; }}
  .errors {{ color: #c0392b; }}
  .total {{ font-weight: 600; }}
</style>
</head>
<body>
  <h1>Semgrep Report</h1>
  <p class="subtitle">{escape(project_name)}</p>
  <p class="total">{total} finding(s)</p>
  <div class="summary">{summary_cells}</div>
  {error_note}
  <table>
    <tr><th>Severity</th><th>Location</th><th>Rule</th><th>Message</th></tr>
    {rows if rows else '<tr><td colspan="4">No findings 🎉</td></tr>'}
  </table>
</body>
</html>"""


def main():
    if len(sys.argv) < 3:
        print("Usage: python semgrep_report.py <input.json> <output.html> [project name]")
        sys.exit(1)

    json_path, html_path = sys.argv[1], sys.argv[2]
    project_name = sys.argv[3] if len(sys.argv) > 3 else json_path

    results, errors = load_results(json_path)
    html = render_html(results, errors, project_name)

    with open(html_path, "w", encoding="utf-8") as f:
        f.write(html)

    print(f"Wrote {html_path} ({len(results)} findings)")


if __name__ == "__main__":
    main()
