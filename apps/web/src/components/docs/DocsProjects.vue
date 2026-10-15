<template>
  <section id="projects" class="scroll-mt-24">
    <h2 class="mb-4 text-[26px] font-semibold tracking-[-0.03em] text-[var(--text-strong)]">
      Projects
    </h2>
    <p class="text-sm leading-relaxed mb-4 text-[var(--text-dim)]">
      A project groups monitors, usually one per client. Each monitor, of any type, can be in one
      project at a time. Create projects on the
      <strong class="text-[var(--text-strong)]">Projects</strong> page and pick each one's monitors
      there, or set the project on a monitor's form, where
      <strong class="text-[var(--text-strong)]">New project…</strong> creates one without leaving
      the page.
    </p>
    <ul class="space-y-2 text-sm text-[var(--text-dim)]">
      <li class="flex items-start gap-2">
        <span class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5 bg-[var(--accent)]"></span>
        <span
          ><strong class="text-[var(--text-strong)]">Grouped list</strong> — the Monitors page shows
          one section per project, with how many of its monitors need attention, and "No project"
          last. A project filter sits next to the search box, and
          <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
            >/monitors?project=…</code
          >
          links straight to one project's monitors.</span
        >
      </li>
      <li class="flex items-start gap-2">
        <span class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5 bg-[var(--accent)]"></span>
        <span
          ><strong class="text-[var(--text-strong)]">Alerts name the project</strong> — every alert
          for a project's monitor starts with the project in brackets, like
          <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
            >[Acme] nightly-backup missed its ping</code
          >, on Telegram, email (subject and body), Slack and SMS. Webhook events carry
          <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
            >projectId</code
          >
          and
          <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
            >projectName</code
          >. Alerts for monitors without a project are unchanged.</span
        >
      </li>
      <li class="flex items-start gap-2">
        <span class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5 bg-[var(--accent)]"></span>
        <span
          ><strong class="text-[var(--text-strong)]">Deleting is safe</strong> — deleting a project
          keeps its monitors running, just without a project.</span
        >
      </li>
      <li class="flex items-start gap-2">
        <span class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5 bg-[var(--accent)]"></span>
        <span
          ><strong class="text-[var(--text-strong)]">Every plan</strong> — projects don't count
          against your plan, up to 1000 per account.</span
        >
      </li>
    </ul>
    <p class="text-sm leading-relaxed mt-4 text-[var(--text-dim)]">
      Notification channels, status pages and maintenance windows are still chosen per monitor;
      project-wide defaults for them aren't available yet.
    </p>
  </section>
</template>
