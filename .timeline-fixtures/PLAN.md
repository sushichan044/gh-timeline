# Timeline fixture runbook

このリポジトリ (`sushichan044/gh-timeline`) で扱う GitHub timeline event
(GraphQL `PullRequestTimelineItems` / `IssueTimelineItems` union) を、
**実データの reference として最小件数の PR / Issue に折り畳んで作る**
ための手順書。

GitHub API のアップデートで union 構成が変わったら、まず
**`.claude/skills/verify-timeline-event-coverage/SKILL.md`**
で差分を取り、必要なら本ファイルの該当 Phase を再実行する。

## Goal

- Live 観測済みイベント種を **23 → 48 / 58** (Core 完了時) まで拡張
- Active 機能 (PR-C / PR-D) で +3 種、最大 **51 / 58**
- 残り 7 種は取得不能 (末尾参照)
- `internal/timeline/coverage_test.go` + GraphQL schema validation が
  並走して回帰検出するので Live 未観測でも信頼できる

## Pre-flight checklist

ターミナルで以下を 1 行ずつ確認。

```sh
gh auth status                                                     # logged in to github.com
git remote -v | grep -F 'sushichan044/gh-timeline'                 # remote OK
gh api repos/sushichan044/gh-timeline --jq .allow_auto_merge        # 必ず true
gh api repos/sushichan044/gh-timeline/collaborators/y-chan          # 200 (招待承認済み) なら review-phase まで進める
```

`allow_auto_merge` が false なら:

```sh
gh api -X PATCH repos/sushichan044/gh-timeline -F allow_auto_merge=true
```

y-chan が collaborator でないなら:

```sh
gh api -X PUT repos/sushichan044/gh-timeline/collaborators/y-chan -F permission=triage
# ↑ 招待後、y-chan が github.com で accept するまで待つ
```

## 全体マップ

| Phase | 内容                                                                                                            | 担当         | 必要前提               |
| ----- | --------------------------------------------------------------------------------------------------------------- | ------------ | ---------------------- |
| 0     | Setup (labels / milestone / dev branch)                                                                         | sushichan044 | Pre-flight 完了        |
| 1     | PR-A 前半 (draft 作成〜force-push)                                                                              | sushichan044 | Phase 0                |
| 2     | PR-B 仕込み (PR-A merge 前の stacked PR)                                                                        | sushichan044 | Phase 1                |
| 3     | y-chan を招待 → review request サイクル発火                                                                     | sushichan044 | Phase 1                |
| 4     | **y-chan 作業** — Approve review + inline thread                                                                | y-chan       | Phase 3                |
| 5     | PR-A 後半 (dismiss / commit comment / demilestone / base 切替 / close-reopen / auto-merge / merge / branch ops) | sushichan044 | Phase 4                |
| 6     | Issue-A / Issue-B                                                                                               | sushichan044 | (Phase 5 と並行可)     |
| 7     | (任意) PR-C deployment / PR-D merge queue                                                                       | sushichan044 | repo settings 追加変更 |
| 8     | 最終カバレッジ確認                                                                                              | sushichan044 | 上記全部               |

## 命名規約

| Resource     | 値                                                      |
| ------------ | ------------------------------------------------------- |
| labels       | `timeline/wip`, `timeline/ready`, `timeline/needs-info` |
| milestone    | `Timeline reference v1`                                 |
| branches     | `dev` (base 切替先), `pr-a/lifecycle`, `pr-b/stacked`   |
| fixture file | `.timeline-fixtures/notes.md`                           |

## Variables (各 Phase で使う環境変数)

```sh
export REPO=sushichan044/gh-timeline
export FRIEND=y-chan
export MILESTONE="Timeline reference v1"
```

PR / Issue 番号は各 Phase で `PR_A`, `PR_B`, `ISSUE_A`, `ISSUE_B` として export する (Phase 1 / 2 / 6 で代入)。

---

## Phase 0: Setup

### Run

```sh
# Labels
gh label create "timeline/wip"        --color "f9d0c4" --description "Reference fixture — WIP" --repo "$REPO"
gh label create "timeline/ready"      --color "0e8a16" --description "Reference fixture — ready" --repo "$REPO"
gh label create "timeline/needs-info" --color "fbca04" --description "Reference fixture — needs info" --repo "$REPO"

# Milestone
gh api -X POST "repos/$REPO/milestones" -f title="$MILESTONE" -f description="Used by .timeline-fixtures/PLAN.md"

# dev branch (PR-A の base 切替先)
git switch main && git pull --ff-only
git switch -c dev
git push -u origin dev
git switch main
```

### Verify

```sh
gh label list --repo "$REPO" --json name --jq '.[] | select(.name|startswith("timeline/")) | .name'
# → 3 件
gh api "repos/$REPO/milestones" --jq '.[].title'
# → "Timeline reference v1" を含む
git ls-remote --heads origin dev | wc -l
# → 1
```

---

## Phase 1: PR-A 前半 (draft → 各種編集 → force-push)

### 1.1 Branch + 初期 commit

```sh
git switch main && git pull --ff-only
git switch -c pr-a/lifecycle

mkdir -p .timeline-fixtures
cat > .timeline-fixtures/notes.md <<'EOF'
# Timeline fixture: kitchen-sink PR

`gh timeline` の reference data 用ダミー。
詳細は `.timeline-fixtures/PLAN.md` を参照。
EOF

git add .timeline-fixtures/notes.md
git commit -m "feat(fixtures): seed kitchen-sink PR target"
git push -u origin pr-a/lifecycle
```

### 1.2 Draft PR を open

```sh
PR_A=$(gh pr create --draft \
  --title "timeline fixture: kitchen sink" \
  --body "Reference PR for gh-timeline. See .timeline-fixtures/PLAN.md." \
  --base main --head pr-a/lifecycle \
  --repo "$REPO" | grep -oE '[0-9]+$')
export PR_A
echo "PR_A=$PR_A"
```

### 1.3 Comment 投稿 → 削除 (`CommentDeletedEvent`)

```sh
CID=$(gh api -X POST "repos/$REPO/issues/$PR_A/comments" -f body="will be deleted" --jq '.id')
gh api -X DELETE "repos/$REPO/issues/comments/$CID"
```

### 1.4 Label / Milestone / Rename / Assign

```sh
gh pr edit "$PR_A" --repo "$REPO" --add-label timeline/wip --add-label timeline/ready  # LabeledEvent ×2
gh pr edit "$PR_A" --repo "$REPO" --remove-label timeline/wip                          # UnlabeledEvent
gh pr edit "$PR_A" --repo "$REPO" --milestone "$MILESTONE"                             # MilestonedEvent
gh pr edit "$PR_A" --repo "$REPO" --title "timeline fixture: kitchen sink (renamed)"   # RenamedTitleEvent
gh pr edit "$PR_A" --repo "$REPO" --add-assignee sushichan044                          # AssignedEvent
gh pr edit "$PR_A" --repo "$REPO" --remove-assignee sushichan044                       # UnassignedEvent
gh pr edit "$PR_A" --repo "$REPO" --add-assignee sushichan044
```

### 1.5 Lock / Unlock

```sh
gh api -X PUT "repos/$REPO/issues/$PR_A/lock" -f lock_reason=off-topic   # LockedEvent
gh api -X DELETE "repos/$REPO/issues/$PR_A/lock"                          # UnlockedEvent
```

### 1.6 Draft ↔ Ready 往復

```sh
gh pr ready "$PR_A" --repo "$REPO"                # ReadyForReviewEvent
gh pr ready "$PR_A" --repo "$REPO" --undo         # ConvertToDraftEvent
gh pr ready "$PR_A" --repo "$REPO"
```

### 1.7 追加 commit + force-push

```sh
echo "second line" >> .timeline-fixtures/notes.md
git commit -am "feat(fixtures): add second commit"
git push                                          # PullRequestCommit
git commit --amend --no-edit
git push --force-with-lease                        # HeadRefForcePushedEvent
```

### Verify (Phase 1)

```sh
gh timeline --repo "$REPO" "$PR_A" --json | jq -r '[.[].type] | unique | .[]' | sort
```

期待: `AssignedEvent`, `CommentDeletedEvent`, `ConvertToDraftEvent`, `HeadRefForcePushedEvent`, `LabeledEvent`, `LockedEvent`, `MilestonedEvent`, `PullRequestCommit`, `ReadyForReviewEvent`, `RenamedTitleEvent`, `UnassignedEvent`, `UnlabeledEvent`, `UnlockedEvent`

---

## Phase 2: PR-B 仕込み (stacked PR)

### Why now

PR-A が **merge される前**に PR-B を作っておく必要がある。PR-A が後で
merge されると、PR-B の base が自動で main に切り替わり
`AutomaticBaseChangeSucceededEvent` が発火する。

### Run

```sh
git switch pr-a/lifecycle && git pull
git switch -c pr-b/stacked
echo "stacked content" > .timeline-fixtures/stacked.md
git add .timeline-fixtures/stacked.md
git commit -m "feat(fixtures): stacked on pr-a"
git push -u origin pr-b/stacked

PR_B=$(gh pr create --base pr-a/lifecycle --head pr-b/stacked \
  --title "timeline fixture: stacked PR" \
  --body "Base auto-changes when PR-A merges. See .timeline-fixtures/PLAN.md." \
  --repo "$REPO" | grep -oE '[0-9]+$')
export PR_B
echo "PR_B=$PR_B"

git switch pr-a/lifecycle
```

### Verify

```sh
gh pr view "$PR_B" --repo "$REPO" --json baseRefName --jq .baseRefName
# → "pr-a/lifecycle"
```

`AutomaticBaseChangeSucceededEvent` は Phase 5 の PR-A merge 後に確認。

---

## Phase 3: y-chan に review を依頼

### Run

```sh
# 既に accept 済みであることが前提 (Pre-flight 参照)
gh pr edit "$PR_A" --repo "$REPO" --add-reviewer "$FRIEND"     # ReviewRequestedEvent
gh pr edit "$PR_A" --repo "$REPO" --remove-reviewer "$FRIEND"  # ReviewRequestRemovedEvent
gh pr edit "$PR_A" --repo "$REPO" --add-reviewer "$FRIEND"
```

### Verify

```sh
gh pr view "$PR_A" --repo "$REPO" --json reviewRequests --jq '.reviewRequests[].login'
# → y-chan
```

---

## Phase 4: ⚠️ y-chan 作業

**y-chan に以下を依頼:**

1. github.com/sushichan044/gh-timeline で PR (kitchen sink) を開く
2. "Files changed" タブで `.timeline-fixtures/notes.md` の任意の行に **inline コメント** を 1 つ以上残す
3. **Review changes → Approve** を選択し submit
4. 完了したら sushichan044 へ通知

### sushichan044 側 verify

```sh
gh pr view "$PR_A" --repo "$REPO" --json reviews \
  --jq '.reviews[] | select(.author.login=="y-chan" and .state=="APPROVED") | .id'
# → review id が 1 件以上返ること
```

返らない場合は y-chan の操作を待つ。

---

## Phase 5: PR-A 後半 (review 後 〜 merge 〜 branch ops)

### 5.1 Comment review を自分でも追加 (`PullRequestReview` state=COMMENTED)

```sh
gh pr review "$PR_A" --repo "$REPO" --comment --body "follow-up: looks good overall"
```

### 5.2 Commit comment thread (`PullRequestCommitCommentThread`)

```sh
SHA=$(gh pr view "$PR_A" --repo "$REPO" --json commits --jq '.commits[-1].oid')
gh api -X POST "repos/$REPO/commits/$SHA/comments" -f body="thread on specific commit"
```

### 5.3 y-chan の Approve review を dismiss (`ReviewDismissedEvent`)

```sh
REVIEW_ID=$(gh api "repos/$REPO/pulls/$PR_A/reviews" \
  --jq '[.[] | select(.user.login=="'"$FRIEND"'" and .state=="APPROVED")] | .[0].id')
gh api -X PUT "repos/$REPO/pulls/$PR_A/reviews/$REVIEW_ID/dismissals" \
  -f message="dismissed for fixture purposes"
```

### 5.4 Milestone を外す (`DemilestonedEvent`)

```sh
gh api -X PATCH "repos/$REPO/issues/$PR_A" -F milestone=null
```

### 5.5 Base ref 切替 (`BaseRefChangedEvent`) と dev への force-push (`BaseRefForcePushedEvent`)

```sh
gh pr edit "$PR_A" --repo "$REPO" --base dev                # BaseRefChangedEvent

# dev を force-push
git switch dev && git pull --ff-only
echo "marker" > .timeline-fixtures/dev-marker.md
git add .timeline-fixtures/dev-marker.md
git commit -m "chore(dev): force-push marker"
git push origin dev
git commit --amend --no-edit
git push --force origin dev                                  # BaseRefForcePushedEvent on PR_A
git switch pr-a/lifecycle

gh pr edit "$PR_A" --repo "$REPO" --base main                # base を戻す
```

### 5.6 Close → Reopen

```sh
gh pr close "$PR_A" --repo "$REPO"                           # ClosedEvent
gh pr reopen "$PR_A" --repo "$REPO"                          # ReopenedEvent
```

### 5.7 Auto-merge cycling

```sh
gh pr merge "$PR_A" --repo "$REPO" --auto --rebase           # AutoRebaseEnabledEvent
gh pr merge "$PR_A" --repo "$REPO" --disable-auto             # AutoMergeDisabledEvent
gh pr merge "$PR_A" --repo "$REPO" --auto --squash            # AutoSquashEnabledEvent
gh pr merge "$PR_A" --repo "$REPO" --disable-auto
gh pr merge "$PR_A" --repo "$REPO" --auto --merge             # AutoMergeEnabledEvent
gh pr merge "$PR_A" --repo "$REPO" --disable-auto
```

### 5.8 実 merge (`MergedEvent`)

```sh
gh pr merge "$PR_A" --repo "$REPO" --rebase
```

merge 後、Phase 2 で作った PR_B の timeline に `AutomaticBaseChangeSucceededEvent` が発火しているはず。

### 5.9 Head branch 削除 → 復元

```sh
HEAD_SHA=$(git rev-parse origin/pr-a/lifecycle)
gh api -X DELETE "repos/$REPO/git/refs/heads/pr-a/lifecycle"  # HeadRefDeletedEvent
gh api -X POST "repos/$REPO/git/refs" \
  -f ref="refs/heads/pr-a/lifecycle" -f sha="$HEAD_SHA"        # HeadRefRestoredEvent
```

### Verify (Phase 5)

```sh
gh timeline --repo "$REPO" "$PR_A" --json | jq -r '[.[].type] | unique | sort | .[]'
# 期待: 計 ~25 種 (PR-A 全部入り)

gh timeline --repo "$REPO" "$PR_B" --json | jq -r '[.[].type] | unique | sort | .[]'
# 期待: AutomaticBaseChangeSucceededEvent, PullRequestCommit を含む
```

---

## Phase 6: Issue-A / Issue-B

### 6.1 Issue-A / Issue-B を作成

```sh
ISSUE_A=$(gh issue create --repo "$REPO" \
  --title "timeline fixture: feedback hub" \
  --body "Reference issue. See .timeline-fixtures/PLAN.md." \
  | grep -oE '[0-9]+$')

ISSUE_B=$(gh issue create --repo "$REPO" \
  --title "timeline fixture: duplicate target" \
  --body "Reference issue B." \
  | grep -oE '[0-9]+$')

export ISSUE_A ISSUE_B
echo "ISSUE_A=$ISSUE_A ISSUE_B=$ISSUE_B"
```

### 6.2 Pin / Unpin (`PinnedEvent` / `UnpinnedEvent`)

```sh
gh api -X PUT "repos/$REPO/issues/$ISSUE_A/pin"
gh api -X DELETE "repos/$REPO/issues/$ISSUE_A/pin"
```

### 6.3 Mark as duplicate / Unmark (`MarkedAsDuplicateEvent` / `UnmarkedAsDuplicateEvent`)

```sh
A_NODE=$(gh api "repos/$REPO/issues/$ISSUE_A" --jq .node_id)
B_NODE=$(gh api "repos/$REPO/issues/$ISSUE_B" --jq .node_id)

gh api graphql -f query='
mutation($id: ID!, $canonical: ID!) {
  markAsDuplicate(input: {duplicateId: $id, canonicalId: $canonical}) {
    duplicate { ... on Issue { number } }
  }
}' -f id="$A_NODE" -f canonical="$B_NODE"

gh api graphql -f query='
mutation($id: ID!) {
  unmarkAsDuplicate(input: {duplicateId: $id}) {
    duplicate { ... on Issue { number } }
  }
}' -f id="$A_NODE"
```

### 6.4 Cross-reference (`CrossReferencedEvent` on Issue-A)

```sh
gh api -X POST "repos/$REPO/issues/$ISSUE_B/comments" \
  -f body="cross-reference test → #$ISSUE_A"
```

### 6.5 Lock / Unlock + Close / Reopen on Issue-A

```sh
gh api -X PUT "repos/$REPO/issues/$ISSUE_A/lock" -f lock_reason=off-topic
gh api -X DELETE "repos/$REPO/issues/$ISSUE_A/lock"
gh issue close "$ISSUE_A" --repo "$REPO"
gh issue reopen "$ISSUE_A" --repo "$REPO"
```

### 6.6 ⚠️ Connected / Disconnected — web UI 必須

> `gh` CLI / REST API では Issue ↔ PR の linkage が直接張れない (Issue
> sidebar の "Development" は GraphQL `linkedBranches` でしか触れない)。

**手順:**

1. github.com で Issue-A を開く
2. 右サイドバー "Development" → "Link a pull request" → PR-A を選択
3. 同じ場所で × ボタンで link を外す

これで `ConnectedEvent` と `DisconnectedEvent` が Issue-A の timeline に発火する。

### Verify (Phase 6)

```sh
gh timeline --repo "$REPO" "$ISSUE_A" --json | jq -r '[.[].type] | unique | sort | .[]'
# 期待: ConnectedEvent, DisconnectedEvent, CrossReferencedEvent, MarkedAsDuplicateEvent, UnmarkedAsDuplicateEvent, PinnedEvent, UnpinnedEvent, LockedEvent, UnlockedEvent, ClosedEvent, ReopenedEvent
```

---

## Phase 7 (任意): Active 機能

実施するなら `repo settings の追加変更` が要る (本ファイル冒頭の Pre-flight を超える)。
後追いで構わない。

### 7.A PR-C: deployment (`DeploymentEnvironmentChangedEvent`)

1. Settings → Environments で `staging`, `production` を作成
2. `.github/workflows/timeline-fixture-deploy.yml` を追加 (merge 時に
   `gh api -X POST .../deployments` を 2 つの env に対して順に実行)
3. PR-C を作成して merge → 2 つ目の deployment で
   `DeploymentEnvironmentChangedEvent` が発火

### 7.B PR-D: merge queue (`AddedToMergeQueueEvent` / `RemovedFromMergeQueueEvent`)

1. Settings → Branches → main rule → **Require merge queue** を有効化
2. PR-D を作成 → merge queue に投入
3. (失敗を狙うなら) CI を落とすコミットを混ぜ、queue から外す

---

## Phase 8: 最終カバレッジ確認

### Run

```sh
{
  for n in "$PR_A" "$PR_B" "$ISSUE_A" "$ISSUE_B"; do
    gh timeline --repo "$REPO" "$n" --json
  done
} | jq -s 'flatten | [.[].type] | unique | sort'
```

### 期待 (Core のみ)

```json
[
  "AssignedEvent",
  "AutoMergeDisabledEvent",
  "AutoMergeEnabledEvent",
  "AutoRebaseEnabledEvent",
  "AutoSquashEnabledEvent",
  "AutomaticBaseChangeSucceededEvent",
  "BaseRefChangedEvent",
  "BaseRefForcePushedEvent",
  "ClosedEvent",
  "CommentDeletedEvent",
  "ConnectedEvent",
  "ConvertToDraftEvent",
  "CrossReferencedEvent",
  "DemilestonedEvent",
  "DisconnectedEvent",
  "HeadRefDeletedEvent",
  "HeadRefForcePushedEvent",
  "HeadRefRestoredEvent",
  "IssueComment",
  "LabeledEvent",
  "LockedEvent",
  "MarkedAsDuplicateEvent",
  "MergedEvent",
  "MilestonedEvent",
  "PinnedEvent",
  "PullRequestCommit",
  "PullRequestCommitCommentThread",
  "PullRequestReview",
  "PullRequestReviewThread",
  "ReadyForReviewEvent",
  "RenamedTitleEvent",
  "ReopenedEvent",
  "ReviewDismissedEvent",
  "ReviewRequestRemovedEvent",
  "ReviewRequestedEvent",
  "UnassignedEvent",
  "UnlabeledEvent",
  "UnlockedEvent",
  "UnmarkedAsDuplicateEvent",
  "UnpinnedEvent"
]
```

40 種前後あれば成功。 加えて既 observed の `MentionedEvent`,
`SubscribedEvent`, `UnsubscribedEvent`, `ReferencedEvent`, `DeployedEvent` も
他 repo の smoke で確認済みの前提で合計 **48 / 58** に到達。

---

## 取得不能 7 種 (現時点で生成不可)

| イベント                                                                                                       | 理由                                             |
| -------------------------------------------------------------------------------------------------------------- | ------------------------------------------------ |
| `TransferredEvent`                                                                                             | Issue が別 repo に移動して本 repo から消える     |
| `ConvertedToDiscussionEvent`                                                                                   | 同上 (Discussion へ変換)                         |
| `UserBlockedEvent`                                                                                             | 実在ユーザーを repo block する必要があり侵襲的   |
| `AddedToProjectEvent` / `RemovedFromProjectEvent` / `MovedColumnsInProjectEvent` / `ConvertedNoteToIssueEvent` | Projects (classic) 専用。2025 EOL で新規発火不可 |

これらは `internal/timeline/coverage_test.go` の reflection 網羅テスト +
GraphQL schema validation で間接的に保護されている。

---

## Recovery (途中で失敗した場合)

| 症状                                                         | 対処                                                                                                                      |
| ------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------- |
| `gh pr merge --auto --rebase` が "Auto-merge is not allowed" | Pre-flight の `allow_auto_merge=true` を確認                                                                              |
| `gh pr edit --add-reviewer y-chan` が "Not found"            | y-chan が collaborator invitation を accept していない。再度 `gh api ... /collaborators/y-chan` で招待                    |
| Phase 2 / Phase 5 で `gh pr create` がコンフリクトする       | `pr-a/lifecycle` を origin と sync (`git pull --rebase`) してから push                                                    |
| Phase 5.9 で head ref 復元失敗                               | `gh api git/refs` 投稿時は ref="refs/heads/pr-a/lifecycle" の prefix を忘れない                                           |
| Phase 6.3 で markAsDuplicate が "Field not found"            | GitHub schema 側で field が変わった可能性。`.claude/skills/verify-timeline-event-coverage/` を起動して schema diff を取る |

---

## 完了後

1. このファイルの末尾に completion log を追加 (どの phase をいつ実行したか)
2. 今後 GitHub API が変わって event 構成が動いたら、
   **`.claude/skills/verify-timeline-event-coverage/SKILL.md`** を起動して
   差分を取り、必要なら本 PLAN.md の該当 Phase をアップデートして再実行
