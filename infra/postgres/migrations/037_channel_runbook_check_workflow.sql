alter table channel_runbook_checks
  add column if not exists check_status text not null default 'DONE',
  add column if not exists assignee text not null default '',
  add column if not exists due_at timestamptz;

create index if not exists channel_runbook_checks_status_idx
  on channel_runbook_checks (check_status, due_at);
