# Summarizes kamal-proxy's request log (JSON lines on stdin) into visitors per day:
# `mise run traffic`. Prints only aggregates, no IPs or raw lines. Real people are
# nearer js_visitors (IPs that also loaded the app's JS) than visitors, which still
# includes bots that send a browser User-Agent (ADR frontend), so top pages and
# referrers count only js_visitors' pageviews. An IP signed in that day (a 200 from an
# API route behind sign-in) is a user, not a visitor, and isn't counted at all: while
# the only users are the founder's own accounts, that leaves out their own visits.
#
# Docker keeps kamal-proxy's log in one 10 MB file, about two days, and drops all of it
# when the file rotates. So cron jobs on the server (`mise run traffic-install`) append
# the last 65 minutes of the log to traffic/requests.log every hour, and after midnight
# run this with --yesterday on that file, append that day's summary as one JSON line to
# traffic/days.jsonl and delete the file. The overlapping hours repeat requests, which
# are counted once by request_id. `mise run traffic` feeds the saved lines in before the
# rest; a saved day wins over the log, whose oldest day may be cut short.
import json, re, sys
from collections import Counter, defaultdict
from datetime import datetime, timedelta, timezone

BOT = re.compile(r"bot|crawl|spider|slurp|curl|wget|python|go-http|httpclient|java/|okhttp|axios|node-fetch|"
                 r"headless|lighthouse|monitor|uptime|checkmeup|preview|facebookexternalhit|semrush|ahrefs|"
                 r"mj12|dotbot|petal|bytespider|gptbot|claude|perplexity|scan|zgrab|masscan|nmap|^$", re.I)
ASSET = re.compile(r"^/(api|assets|up)(/|$)|\.[a-z0-9]{1,5}$", re.I)
PUBLIC_API = re.compile(r"^/api/v1/(health|stats|tools|auth)(/|$)")  # the rest needs a session (server.go)

days = defaultdict(lambda: {"ips": set(), "hits": []})  # hits: (ip, path, referrer host)
js_ips = defaultdict(set)  # IPs that also loaded the app's JS: real browsers do, most fake-UA bots don't
signed_in = defaultdict(set)  # IPs with a session that day
bot_by_day = Counter()
saved = {}  # day -> summary line from traffic/days.jsonl
seen = set()  # request IDs already counted
fields, total = None, 0
for line in sys.stdin:
    line = line.strip()
    if not line.startswith("{"):
        continue
    try:
        r = json.loads(line)
    except ValueError:
        continue
    if "day" in r and "msg" not in r:
        saved[r["day"]] = r
        continue
    if r.get("msg") != "Request":
        continue
    rid = r.get("request_id") or line
    if rid in seen:
        continue
    seen.add(rid)
    total += 1
    fields = fields or sorted(r)
    day = str(r.get("time", ""))[:10]
    ua = r.get("user_agent") or ""
    path = (r.get("path") or "").split("?")[0]
    if path.startswith("/assets/") and path.endswith(".js") and not BOT.search(ua):
        js_ips[day].add(r.get("remote_addr") or r.get("client_addr"))
    if path.startswith("/api/v1/") and not PUBLIC_API.match(path) and r.get("status") == 200:
        signed_in[day].add(r.get("remote_addr") or r.get("client_addr"))
    if r.get("method") != "GET" or r.get("status") != 200 or ASSET.search(path):
        continue
    if BOT.search(ua):
        bot_by_day[day] += 1
        continue
    ip = r.get("remote_addr") or r.get("client_addr")
    d = days[day]
    d["ips"].add(ip)
    # The Referer header is logged under a key named after it (config/deploy.yml proxy.logging).
    ref = next((v for k, v in r.items() if "referer" in k.lower()), None)
    host = re.sub(r"^https?://([^/]+).*", r"\1", ref) if ref and "checkmeup.net" not in ref else None
    d["hits"].append((ip, path, host))


def summary(day):
    d = days[day]
    ips = d["ips"] - signed_in[day]
    js = ips & js_ips[day]
    pages = Counter(path for ip, path, _ in d["hits"] if ip in js)
    refs = Counter(host for ip, _, host in d["hits"] if ip in js and host)
    views = sum(1 for ip, _, _ in d["hits"] if ip in ips)
    return {"day": day, "visitors": len(ips), "js_visitors": len(js),
            "pageviews": views, "bot_pageviews": bot_by_day[day],
            "pages": dict(pages.most_common(20)), "referrers": dict(refs.most_common(15))}


if "--yesterday" in sys.argv:
    day = (datetime.now(timezone.utc) - timedelta(days=1)).date().isoformat()
    if day in days:
        print(json.dumps(summary(day)))
    sys.exit()

rows = {**{day: summary(day) for day in days}, **saved}
paths, refs = Counter(), Counter()
for row in rows.values():
    paths.update(row["pages"])
    refs.update(row["referrers"])

print(f"requests logged: {total}; saved days: {len(saved)}; fields: {fields}")
print("day         visitors  js_visitors  pageviews  bot_pageviews  source")
for day in sorted(rows):
    r = rows[day]
    src = "saved" if day in saved else "log"
    print(f"{day}  {r['visitors']:8}  {r['js_visitors']:11}  {r['pageviews']:9}  {r['bot_pageviews']:13}  {src}")
print("\ntop pages (js_visitors only):")
for p, n in paths.most_common(20):
    print(f"{n:6}  {p}")
if refs:
    print("\ntop external referrers (js_visitors only):")
    for h, n in refs.most_common(15):
        print(f"{n:6}  {h}")
