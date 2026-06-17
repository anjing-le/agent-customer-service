create table if not exists channel_runbook_check_events (
  id text primary key,
  action text not null,
  channel text not null,
  runbook_status text not null,
  check_status text not null,
  check_id text not null default '',
  step text not null,
  step_index integer not null,
  action_ref text not null default '',
  report_id text not null default '',
  assignee text not null default '',
  due_at timestamptz,
  actor text not null default '',
  note text not null default '',
  created_at timestamptz not null default now()
);

create index if not exists channel_runbook_check_events_created_at_idx
  on channel_runbook_check_events (created_at desc, id desc);

create index if not exists channel_runbook_check_events_channel_idx
  on channel_runbook_check_events (channel, action, created_at desc);
