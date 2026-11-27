#!/usr/bin/env bash
# Review metrics for merged PRs (see intent/review-process/intent.md): human approvals per PR, review
# rounds, hours a PR waited for its first human review, reverts, and how review findings were answered.
# Usage: scripts/review-metrics.sh [since YYYY-MM-DD] [owner/repo]   (needs gh, jq; read-only)
set -euo pipefail

since=${1:-$(date -u -v-90d +%F 2>/dev/null || date -u -d '90 days ago' +%F)}
repo=${2:-Backblaze/terraform-provider-b2}
owner=${repo%/*}
name=${repo#*/}

query='query($q: String!, $endCursor: String) {
  search(query: $q, type: ISSUE, first: 50, after: $endCursor) {
    pageInfo { hasNextPage endCursor }
    nodes { ... on PullRequest {
      number title createdAt mergedAt author { login }
      timelineItems(itemTypes: [READY_FOR_REVIEW_EVENT], first: 1) { nodes { ... on ReadyForReviewEvent { createdAt } } }
      reviews(first: 100) { nodes { state submittedAt author { login __typename } } }
      reviewThreads(first: 100) { nodes { comments(first: 50) { nodes { body author { login } } } } }
    } }
  }
}'

data=$(gh api graphql --paginate -f q="repo:$owner/$name is:pr is:merged merged:>=$since" -f query="$query" \
  --jq '.data.search.nodes[]' | jq -s '[.[] | select(.author.login != "dependabot")]')

jq -r --arg since "$since" '
  def human: .author.__typename == "User" and (.author.login | test("copilot|\\[bot\\]"; "i") | not);
  def pct(p): if length == 0 then 0 else sort | .[((length - 1) * p | floor)] end;
  def hours(a; b): (((b | fromdateiso8601) - (a | fromdateiso8601)) / 3600 | floor);
  def answer_label: if test("^\\s*Fixed:"; "i") then "fixed" elif test("^\\s*Wrong:"; "i") then "wrong"
    elif test("^\\s*Intentional:"; "i") then "intentional" elif test("^\\s*Didn.t matter:"; "i") then "didnt_matter"
    else empty end;
  [ .[] | . as $pr
    | ([.reviews.nodes[] | select(human and .author.login != $pr.author.login)]) as $hr
    | { n: .number,
        approvals: ([$hr[] | select(.state == "APPROVED")] | length),
        rounds: ([.reviews.nodes[] | select(.state != "PENDING")] | length),
        self_merged_unreviewed: (($hr | length) == 0),
        wait: (if ($hr | length) > 0 then hours(((.timelineItems.nodes[0].createdAt) // .createdAt); ($hr | map(.submittedAt) | min)) else null end),
        revert: (.title | test("^Revert"; "i")),
        answers: [.reviewThreads.nodes[].comments.nodes[1:][] | .body | answer_label] } ] as $rows
  | "Merged PRs since \($since) (Dependabot excluded): \($rows | length)",
    "Human approvals per PR: median \([$rows[].approvals] | pct(0.5)), p95 \([$rows[].approvals] | pct(0.95))",
    "Reviews per PR (all reviewers): median \([$rows[].rounds] | pct(0.5)), p95 \([$rows[].rounds] | pct(0.95))",
    "Merged with no review from anyone but the author: \([$rows[] | select(.self_merged_unreviewed)] | length)",
    "Hours to first human review: median \([$rows[].wait | select(. != null)] | pct(0.5)), p95 \([$rows[].wait | select(. != null)] | pct(0.95))",
    "Revert PRs: \([$rows[] | select(.revert)] | length)",
    "Finding answers: \([$rows[].answers[]] | group_by(.) | map("\(.[0]) \(length)") | join(", ") | if . == "" then "none labelled yet" else . end)"
' <<<"$data"
