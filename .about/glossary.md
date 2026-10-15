# Checkmeup glossary

Checkmeup is a developer monitoring platform: cron, uptime, SSL, domain expiry, port and DNS monitors with alerts and status pages.

**Article**:
An evergreen blog post explaining one topic, such as a monitor type. It is listed on the blog and indexed by search engines.
_In code_: `kind: 'article'`

**Release note**:
A blog post announcing what one version shipped. It is listed on the Changelog, not the blog, and kept out of search indexes.
_In code_: `kind: 'release'`

**Changelog**:
The single public page that lists every Release note, newest first, each under its own anchor.

**Project**:
A named group of monitors in an org, usually one per client. A monitor belongs to at most one project; deleting a project leaves its monitors ungrouped. Not a workspace: channels, status pages and API keys stay org-wide.
_In code_: `projects`, `project_id`
