---
name: verify-timeline-event-coverage
description: Use when GitHub may have shipped a GraphQL schema update affecting timeline event unions (PullRequestTimelineItems / IssueTimelineItems), when the gh-timeline dispatcher fallback is hitting more than expected, before a gh-timeline release, or when the user requests phrases like "verify timeline coverage", "audit gh-timeline schema", "check for new timeline events", "GitHub schema changed". Walks through schema diff → patch → test update → live verification → optional fixture refresh, marking each spot where a human decision is required.
---

# verify-timeline-event-coverage

このリポジトリ (`gh-timeline`) は GitHub の GraphQL union を 2 つ網羅する:

- `PullRequestTimelineItems` (PR timeline 用)
- `IssueTimelineItems` (Issue timeline 用)

GitHub は schema を不定期に更新する。新規 event 追加・既存 field の変更・deprecate
が起きると、本リポジトリの宣言と乖離して dispatcher fallback に落ちたり、
そもそも query が schema error で落ちる可能性がある。

このスキルは **その差分検出 → 修正 → 検証** までを段階的に案内する。

## 起動の判断基準

以下のいずれかが当てはまれば本スキルを invoke する:

- ユーザーが「timeline event coverage」「schema audit」「新しい event 対応」等を依頼
- `coverage_test.go` の `TestDispatchCoverage_*` が fail
- Live で `gh timeline` を叩いたら未知の `__typename` が頻発
- gh-timeline の major release 前の最終チェック
- GitHub の changelog / blog で timeline event 系の追加・変更が観測された

## 全体フロー

```
[Step 1: Schema diff]
        ↓
[Step 2: Triage]  ← 人間判断ポイント
        ↓
[Step 3: Patch code]
        ↓
[Step 4: Update tests]
        ↓
[Step 5: Live verification]  ← 人間判断ポイント
        ↓
[Step 6: (任意) Fixture refresh]  ← PLAN.md へ委譲
```

⚠️ マーク = 人間の判断や手動操作が必要な箇所。

## Step 1: Schema diff

GitHub の現行 union member を introspection で取得し、本リポジトリの宣言と比較する。

```sh
# 作業ディレクトリへ移動
cd "$(git rev-parse --show-toplevel)"

# 1.1 GitHub schema 側の union member
gh api graphql -f query='
query {
  pr: __type(name: "PullRequestTimelineItems") { possibleTypes { name } }
  issue: __type(name: "IssueTimelineItems") { possibleTypes { name } }
}' --jq '.data.pr.possibleTypes[].name'    | sort -u > /tmp/pr_upstream.txt
gh api graphql -f query='
query { issue: __type(name: "IssueTimelineItems") { possibleTypes { name } } }
' --jq '.data.issue.possibleTypes[].name'  | sort -u > /tmp/issue_upstream.txt

# 1.2 本 repo 側の宣言 (graphql tag を抽出)
grep -hE '`graphql:"\.\.\. on [A-Za-z0-9]+"`' internal/timeline/query.go \
  | grep -oE 'on [A-Za-z0-9]+' | sed 's/^on //' | sort -u > /tmp/repo_declared.txt

# 1.3 PR union の差分 — 本 repo は両 union 分を 1 ファイルに宣言しているので
#     pr_upstream に対してのみ厳密チェック (issue union は subset)
echo "== upstream PR にあって repo に無い (新規対応が必要) =="
comm -23 /tmp/pr_upstream.txt /tmp/repo_declared.txt

echo "== repo に宣言があるが upstream PR に無い (deprecated 検証) =="
comm -13 /tmp/pr_upstream.txt /tmp/repo_declared.txt | grep -v -E '^(Issue|PullRequest)$'

echo "== issue union の追加分 (PR 共通でない場合) =="
comm -23 /tmp/issue_upstream.txt /tmp/pr_upstream.txt
```

## Step 2: Triage ⚠️ 人間判断

Step 1 の結果を 4 つに分類する。

| 差分カテゴリ          | 内容                                   | 対処                                                                                                                            |
| --------------------- | -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| (A) 新規追加          | upstream にあって repo にない typename | Step 3 で fragment + dispatch case + handler を追加                                                                             |
| (B) 削除 (deprecated) | repo にあって upstream にない typename | Step 3 で fragment / case / handler を撤去。Projects classic 系は EOL 後しばらく schema には残ることがある — GitHub Docs で確認 |
| (C) フィールド変更    | typename は同じだが field が変わった   | introspection だけだと検出不可。**GitHub Docs / Changelog で event 型の field 一覧を目視確認**                                  |
| (D) 命名変更          | A と B の両方に同時出現                | 既存 fragment / case を rename で対応                                                                                           |

⚠️ **人間判断が要るところ:**

- B の場合: 本当に削除して良いか (関連 reference fixture が壊れないか) を判断
- C の場合: 既存 fragment が必要な field を取り続けているかを Docs で確認
- D の場合: rename と削除の見分けは Docs / Changelog 必須
- 「summary 文言をどう書くか」(API では決まらない設計判断)

## Step 3: Patch code

新規 event `<FooEvent>` 1 つにつき以下を実施:

### 3.1 Fragment 追加 (`internal/timeline/nodes.go`)

actor + createdAt しか使わないなら `commonEvent` を embed:

```go
type fooEventFragment struct {
    commonEvent
    SomeSpecificField githubv4.String
}
```

actor / createdAt が独自フィールド (例: `submittedAt`) なら個別宣言:

```go
type fooEventFragment struct {
    ID          githubv4.ID
    Author      actorFragment
    SubmittedAt githubv4.DateTime
}
```

### 3.2 Query 追加 (`internal/timeline/query.go`)

`prTimelineNode` に追加:

```go
FooEvent fooEventFragment `graphql:"... on FooEvent"`
```

PR / Issue 両 union のメンバなら `issueTimelineNode` にも同じ tag で追加。

### 3.3 Dispatch case 追加 (`internal/timeline/dispatch.go`)

`dispatchPRNode` (および該当すれば `dispatchIssueNode`) の switch に case 追加:

```go
case "FooEvent":
    return handleFooEvent(t, n.FooEvent)
```

### 3.4 Handler 関数 (`internal/timeline/dispatch.go`)

既存 handler の summary は **動詞句 + 対象フィールド** の形に統一されている
(`added label bug`, `assigned bea`, `merged deadbee into main` 等)。フィールドが
空でも動詞句は残すことで、最低限の意味を保つ。新規 handler もこの流儀に揃える:

```go
func handleFooEvent(typename string, f fooEventFragment) Event {
    verb := "did foo to"  // ⚠️ 文言は人間判断
    summary := verb
    if target := string(f.SomeSpecificField); target != "" {
        summary = fmt.Sprintf("%s %s", verb, target)
    }
    return Event{
        Type:      typename,
        Actor:     string(f.Actor.Login),
        Timestamp: f.CreatedAt.Time,
        Summary:   summary,
        Ref:       Ref{NodeID: graphqlIDString(f.ID)},
    }
}
```

ペアになる add/remove 型 event (Labeled/Unlabeled, Assigned/Unassigned,
Milestoned/Demilestoned, ReviewRequested/ReviewRequestRemoved 等) は **同じ
handler を共有して `typename` で動詞を切り替える** のが既存流儀:

```go
verb := "added label"
if typename == "UnlabeledEvent" {
    verb = "removed label"
}
```

⚠️ **人間判断**:

- 動詞の選定 (`added label` か `labeled` か など)
- 補助情報をどこまで summary に押し込むか (truncate(title) や `(column %q)` のような
  括弧書き、`(%s)` の duration 等。`handleConnected` / `handleProjectChange` /
  `handleUserBlocked` が実例)
- SHA を入れるなら `shortSHA()` を通すこと (`handleMerged` 参照)

### 3.5 削除パスは逆順で実施

`B` カテゴリの event は上記 4 つを削除。fragment 型は他の event と共有していないかを `git grep` で確認してから削除する。

## Step 4: Update tests

### 4.1 Coverage test (自動追従)

`internal/timeline/coverage_test.go` は reflection で struct tag を列挙するため、
Step 3.2 で query.go に追加した時点で対応する sub-test が自動で生まれる。

```sh
mise run test -- -run TestDispatchCoverage
# 期待: TestDispatchCoverage_prTimelineNode/<NewType> が新規に PASS
```

### 4.2 Rich summary test (手動追加)

`internal/timeline/dispatch_test.go` の table-driven テストに代表的なケースを追加。
`name` は **何を assert しているかが伝わる文** にする (例: "FooEvent uses the
foo'd verb with target name") — `assert behavior, not implementation` の原則:

```go
{
    name: "FooEvent uses the foo'd verb with target name",
    node: func() prTimelineNode {
        n := prTimelineNode{Typename: "FooEvent"}
        n.FooEvent.Actor.Login = "alice"
        n.FooEvent.CreatedAt = dt(ts)
        n.FooEvent.SomeSpecificField = "bar"
        return n
    }(),
    wantType:    "FooEvent",
    wantActor:   "alice",
    wantSummary: "foo'd bar",
},
```

ペア event を追加した場合は、**両方の typename** と **特定フィールドが空** の
fallback も別ケースとしてカバーする。

### 4.3 Full gate

```sh
mise run fmt
mise run lint
mise run test
```

3 つ全部 0 issue / pass で次へ。

## Step 5: Live verification ⚠️ 人間判断

実 PR / Issue で発火するかを確認:

```sh
# 該当 event が出る PR が分かっていれば
go run . --repo OWNER/REPO PR_NUMBER --json | jq -r '.[] | select(.type=="FooEvent")'

# 出る PR が思い当たらない場合 → GitHub search で探す
gh search prs "<キーワード>" --json url,number --limit 5
```

⚠️ **人間判断**:

- 該当 event を起こす GitHub 機能を知っているなら、それを使っている公開 repo を狙い撃ち
- 思い当たらない / 試行錯誤の余地が少ない場合は Step 6 (本 repo で自前生成) へ
- 機能自体が public/外部から触れない (UserBlockedEvent 等) なら Live 検証は諦め、
  Step 4 の coverage_test + schema validation で保護を完結させる
- 取得不能と判定した event は `.timeline-fixtures/PLAN.md` 末尾の「取得不能 7 種」表に追記

## Step 6 (任意): Fixture refresh

本 repo の reference fixture (`sushichan044/gh-timeline` の PR-A / PR-B / Issue-A / Issue-B)
を新 event に対応させたい場合、`.timeline-fixtures/PLAN.md` の該当 Phase に
操作を追加してから再実行する。

判断基準:

- 新 event が **CLI / GraphQL から発火可能** → PLAN.md の関連 Phase に行を追加
- 新 event が **特定の GitHub 機能 (deployment / merge queue / project)** → PLAN.md Phase 7
  (任意 PR-C / PR-D) を拡張するか、専用 Phase を追加
- ⚠️ **新 event が friend (collaborator) 操作を要する** → PLAN.md 役割分担表に追記し、
  Phase 4 と同様の hand-off を設計

## Completion checklist

スキル完了時、以下が全て満たされること:

```
[ ] Step 1 の schema diff で repo と upstream が乖離なし
[ ] Step 3 の patch がコミット済み (conventional commit: feat/fix/refactor)
[ ] mise run fmt / lint / test が全て 0 issue / pass
[ ] 新 event について Step 5 で Live 観測 or Step 6 の fixture 拡張 or
    取得不能と判定して PLAN.md に追記
[ ] (該当すれば) .timeline-fixtures/PLAN.md を最新化
[ ] commit message に「新 event X 件、削除 Y 件」のような差分サマリを記載
```

## 補足: gotchas

- **`Issue` / `PullRequest` のような top-level union 名** (timelineQuery 自体の
  `... on PullRequest` 等) は本スキルの差分対象外。grep のフィルタで除外する
- **`databaseId` (Int) が int32 overflow** することがある (`comment_id` 4G 超え)。
  fragment では `int64` で受ける (既存 issueCommentFragment / pullRequestReviewFragment 参照)
- **Schema introspection は時々一時的に遅い**。timeout したら少し待ってリトライ
- **`gh timeline` 自体が schema 変更で動かなくなった場合** (= まず query が
  schema error で落ちる) は本スキルではなく直接 `internal/timeline/query.go` を
  最小修正してから本スキルを起動する

## Cross-reference

- 実 reference fixture 生成手順: `.timeline-fixtures/PLAN.md`
- dispatcher 実装: `internal/timeline/dispatch.go`
- query 宣言: `internal/timeline/query.go`
- fragment 構造: `internal/timeline/nodes.go`
- 反射ベース網羅テスト: `internal/timeline/coverage_test.go`
- handler 単体テスト: `internal/timeline/dispatch_test.go`
