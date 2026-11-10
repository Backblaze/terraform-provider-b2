# Reviewer evals

Bugs that shipped or were caught late, used to check whether a reviewer (an AI review configuration, or
`REVIEW.md` and `AGENTS.md` themselves) would have caught them. Re-score the reviewer whenever its prompt,
`REVIEW.md`, `AGENTS.md`, its tools or its model change, and keep a change only if the score does not drop.

Each case is the fix PR and the bug it fixed. To build the eval, find the commit that introduced the bug
(`git log -S` on the fixed line, or `git blame` on the fix's parent), and review that change as if it were
a new PR. The case counts as caught when the reviewer reports the bug class below at Important severity.
A maintainer who knows the code confirms each case before it is used; until then it is `unconfirmed`.

| Fix PR | Fix commit | Bug class the reviewer must report | Status |
| --- | --- | --- | --- |
| #146 | `a2f29cd` | Plan and state: `b2_bucket` `revision` not refreshed after update, so the next plan shows drift | unconfirmed |
| #151 | `a4db154` | Plan and state: B2 lowercases `bucket_info` keys, so mixed-case keys diff forever | unconfirmed |
| #153 | `3922992` | Plan and state: keys B2 adds to `file_info` (a `ForceNew` map) force a re-upload on every apply | unconfirmed |
| #154 | `8a3e583` | Plan and state: B2 does not keep `lifecycle_rules` order, so the plan never goes empty | unconfirmed |
| #157 | `524642b` | Plan and state: an omitted `default_server_side_encryption` re-sends the last value from state | unconfirmed |
| #132 | `ba1c63e` | Compatibility: the deprecated `bucket_id` argument stopped working on the v4 API | unconfirmed |
| #24 | `8d78402` | Bugs: deleting an already-deleted bucket fails instead of being treated as gone | unconfirmed |

Also keep a set of known-clean PRs (merged, never reverted, no follow-up fix) and run the reviewer on
them: count the comments that would have led to a harmful or pointless change. Recall on the table above
and noise on the clean set are both reported when a reviewer change is proposed.

Add a row whenever a fix PR lands for a bug that review should have caught.
