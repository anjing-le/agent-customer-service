create table if not exists channel_runbook_checks (
  id text primary key,
  channel text not null,
  runbook_status text not null,
  step text not null,
  step_index integer not null,
  action_ref text not null default '',
  report_id text not null default '',
  actor text not null default '',
  note text not null default '',
  completed_at timestamptz not null default now()
);

create unique index if not exists channel_runbook_checks_unique_step
  on channel_runbook_checks (channel, runbook_status, step_index, action_ref);

create index if not exists channel_runbook_checks_completed_at_idx
  on channel_runbook_checks (completed_at desc);
