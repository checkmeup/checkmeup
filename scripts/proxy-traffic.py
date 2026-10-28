# Summarizes kamal-proxy's request log (JSON lines on stdin) into visitors per day:
# `mise run traffic`. Prints only aggregates, no IPs or raw lines. Real people are
# nearer js_visitors (IPs that also loaded the app's JS) than visitors, which still
# includes bots that send a browser User-Agent (ADR frontend).
import json, re, sys
from collections import Counter, defaultdict

BOT = re.compile(r"bot|crawl|spider|slurp|curl|wget|python|go-http|httpclient|java/|okhttp|axios|node-fetch|"
                 r"headless|lighthouse|monitor|uptime|checkmeup|preview|facebookexternalhit|semrush|ahrefs|"
                 r"mj12|dotbot|petal|bytespider|gptbot|claude|perplexity|scan|zgrab|masscan|nmap|^$", re.I)
ASSET = re.compile(r"^/(api|assets|up)(/|$)|\.[a-z0-9]{1,5}$", re.I)

days = defaultdict(lambda: {"views": 0, "ips": set()})
js_ips = defaultdict(set)  # IPs that also loaded the app's JS: real browsers do, most fake-UA bots don't
paths, refs, bot_by_day = Counter(), Counter(), Counter()
fields, total = None, 0
for line in sys.stdin:
    line = line.strip()
    if not line.startswith("{"):
        continue
    try:
        r = json.loads(line)
    except ValueError:
        continue
    if r.get("msg") != "Request":
        continue
    total += 1
    fields = fields or sorted(r)
    day = str(r.get("time", ""))[:10]
    ua = r.get("user_agent") or ""
    path = (r.get("path") or "").split("?")[0]
    if path.startswith("/assets/") and path.endswith(".js") and not BOT.search(ua):
        js_ips[day].add(r.get("remote_addr") or r.get("client_addr"))
    if r.get("method") != "GET" or r.get("status") != 200 or ASSET.search(path):
        continue
    if BOT.search(ua):
        bot_by_day[day] += 1
        continue
    d = days[day]
    d["views"] += 1
    d["ips"].add(r.get("remote_addr") or r.get("client_addr"))
    paths[path] += 1
    # The Referer header is logged under a key named after it (config/deploy.yml proxy.logging).
    ref = next((v for k, v in r.items() if "referer" in k.lower()), None)
    if ref and "checkmeup.net" not in ref:
        refs[re.sub(r"^https?://([^/]+).*", r"\1", ref)] += 1

print(f"requests logged: {total}; fields: {fields}")
print("day         visitors  js_visitors  pageviews  bot_pageviews")
for day in sorted(days):
    print(f"{day}  {len(days[day]['ips']):8}  {len(days[day]['ips'] & js_ips[day]):11}  {days[day]['views']:9}  {bot_by_day[day]:13}")
print("\ntop pages:")
for p, n in paths.most_common(20):
    print(f"{n:6}  {p}")
if refs:
    print("\ntop external referrers:")
    for h, n in refs.most_common(15):
        print(f"{n:6}  {h}")
