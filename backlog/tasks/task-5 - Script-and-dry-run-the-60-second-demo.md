---
id: TASK-5
title: Script and dry-run the 60-second demo
status: To Do
assignee: []
created_date: '2026-11-16T12:19:00Z'
updated_date: '2026-11-16T12:49:00Z'
labels: []
milestone: m-5
dependencies:
  - TASK-3
  - TASK-4
ordinal: 5000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Once TASK-3 and TASK-4 are done, write the shot list for an honest video of about 60 seconds: 0–30s neighbr is built from its spec, 30–60s one spec change rebuilds one part. Nothing is staged; the video shows what philote really does. Do a timed dry run of every step, including a build from an empty lock, so the user knows where to cut or speed up. Anything slow or ugly the dry run finds is a product task, not a video trick. The user records the video.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A shot list with seconds per shot, short enough for the user to follow while recording
- [ ] #2 Dry run done, with measured times for each step
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Recording: turn off Dark Reader for localhost, or light mode turns brown and Next shows hydration warnings. Use next build + next start, not next dev, so the dev badge doesn't show. A neighbr build with a fresh lock asks the model for 5-6 nodes in parallel; the last one took about 70s.
<!-- SECTION:NOTES:END -->
