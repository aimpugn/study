#!/usr/bin/env bash

set -Eeuo pipefail
IFS=$'\n\t'
LC_ALL=C
export LC_ALL

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source=migrate_gitlab_group.sh
source "$SCRIPT_DIR/migrate_gitlab_group.sh"

TEST_ROOT=$(mktemp -d)
TEST_COUNT=0
trap 'rm -rf -- "$TEST_ROOT"' EXIT

fail() {
    printf 'FAIL: %s\n' "$*" >&2
    exit 1
}

pass() {
    ((TEST_COUNT += 1))
    printf 'PASS: %s\n' "$1"
}

assert_eq() {
    local expected=$1
    local actual=$2
    local label=$3
    [[ $actual == "$expected" ]] ||
        fail "$label: expected [$expected], actual [$actual]"
}

select_test_json_backend() {
    select_json_backend
    printf 'JSON backend under test: %s\n' "$JSON_BACKEND"
}

test_json_records_and_payloads() {
    local fixture="$TEST_ROOT/projects.json"
    local payload="$TEST_ROOT/project-payload.json"
    local record

    cat >"$fixture" <<'JSON'
[
  {
    "id": 11,
    "name": "Payment API",
    "path": "payment-api",
    "path_with_namespace": "kcf/payments/payment-api",
    "namespace": {"full_path": "kcf/payments"},
    "visibility": "private",
    "default_branch": "main",
    "http_url_to_repo": "https://source.example/kcf/payments/payment-api.git"
  }
]
JSON

    assert_eq "1" "$(json_array_length "$fixture")" "JSON array length"
    record=$(json_emit_records project-array "$fixture")
    IFS=$'\t' read -r \
        id name_b64 path full_path namespace visibility \
        default_branch_b64 repo_url_b64 <<<"$record"
    assert_eq "11" "$id" "project id"
    assert_eq "Payment API" "$(b64_decode "$name_b64")" "project name"
    assert_eq "payment-api" "$path" "project path"
    assert_eq "kcf/payments/payment-api" "$full_path" "project full path"
    assert_eq "kcf/payments" "$namespace" "project namespace"
    assert_eq "private" "$visibility" "project visibility"
    assert_eq "main" "$(b64_decode "$default_branch_b64")" "default branch"
    assert_eq \
        "https://source.example/kcf/payments/payment-api.git" \
        "$(b64_decode "$repo_url_b64")" \
        "repository URL"

    json_make_payload \
        project \
        "$payload" \
        "Payment API" \
        "payment-api" \
        101 \
        private \
        true

    if command -v jq >/dev/null 2>&1; then
        assert_eq \
            "false" \
            "$(jq -r '.initialize_with_readme' "$payload")" \
            "README initialization"
        assert_eq \
            "101" \
            "$(jq -r '.namespace_id' "$payload")" \
            "namespace id"
    else
        perl -MJSON::PP -0777 -e '
            my $file = shift;
            open my $fh, "<:raw", $file or die "$file: $!";
            local $/;
            my $value = decode_json(<$fh>);
            $value->{initialize_with_readme} and die "README enabled\n";
            $value->{namespace_id} == 101 or die "bad namespace\n";
        ' "$payload"
    fi

    pass "JSON records and project payload"
}

test_state_and_group_mapping() {
    SOURCE_URL="https://source.example"
    TARGET_URL="https://target.example"
    SOURCE_GROUP="kcf"
    TARGET_GROUP="target/kcf"
    WORK_DIR="$TEST_ROOT/state-work"
    mkdir -p -- "$WORK_DIR"
    STATE_FILE="$WORK_DIR/state.tsv"

    state_init
    state_update \
        11 \
        201 \
        "kcf/payments/payment-api" \
        "target/kcf/payments/payment-api" \
        1 \
        "created" \
        0
    local record
    record=$(state_get 11)
    assert_eq \
        $'P\t11\t201\tkcf/payments/payment-api\ttarget/kcf/payments/payment-api\t1\tcreated\t0' \
        "$record" \
        "state record"
    assert_eq \
        "target/kcf/payments" \
        "$(map_group_path "kcf/payments")" \
        "group path mapping"

    pass "state resume record and group mapping"
}

make_fake_curl() {
    local fake_bin=$1
    mkdir -p -- "$fake_bin"

    cat >"$fake_bin/curl" <<'FAKE'
#!/usr/bin/env bash
set -Eeuo pipefail

method="GET"
output=""
headers=""
url=""
printf '%s\n' "$*" >>"$FAKE_CURL_ARGS"
cat >/dev/null

while (($#)); do
    case $1 in
        --request)
            method=$2
            shift 2
            ;;
        --output)
            output=$2
            shift 2
            ;;
        --dump-header)
            headers=$2
            shift 2
            ;;
        --write-out|--connect-timeout|--max-time|--max-redirs|--proto|\
        --retry|--retry-delay|--retry-max-time|--header|--data-binary)
            shift 2
            ;;
        --silent|--show-error)
            shift
            ;;
        --config)
            shift 2
            ;;
        *)
            url=$1
            shift
            ;;
    esac
done

printf '%s %s\n' "$method" "$url" >>"$FAKE_CURL_CALLS"
: >"$headers"
status=200

if [[ $method == "POST" &&
    $url == "https://target.example/api/v4/groups" ]]
then
    cat >"$output" <<'JSON'
{"id":101,"name":"Payments","path":"payments","full_path":"target/kcf/payments","visibility":"private"}
JSON
elif [[ $method == "POST" &&
    $url == "https://target.example/api/v4/projects" ]]
then
    cat >"$output" <<'JSON'
{
  "id":201,
  "name":"Payment API",
  "path":"payment-api",
  "path_with_namespace":"target/kcf/payments/payment-api",
  "namespace":{"full_path":"target/kcf/payments"},
  "visibility":"private",
  "default_branch":null,
  "http_url_to_repo":"https://target.example/target/kcf/payments/payment-api.git"
}
JSON
else
case $url in
    https://source.example/api/v4/groups/kcf)
        cat >"$output" <<'JSON'
{"id":1,"name":"KCF","path":"kcf","full_path":"kcf","visibility":"private"}
JSON
        ;;
    https://target.example/api/v4/groups/target%2Fkcf)
        cat >"$output" <<'JSON'
{"id":100,"name":"KCF","path":"kcf","full_path":"target/kcf","visibility":"private"}
JSON
        ;;
    *"/api/v4/groups/1/descendant_groups?"*)
        cat >"$output" <<'JSON'
[{"id":2,"name":"Payments","path":"payments","full_path":"kcf/payments","visibility":"private"}]
JSON
        ;;
    *"/api/v4/groups/1/projects?"*)
        cat >"$output" <<'JSON'
[
  {
    "id":10,
    "name":"Existing",
    "path":"existing",
    "path_with_namespace":"kcf/existing",
    "namespace":{"full_path":"kcf"},
    "visibility":"private",
    "default_branch":"main",
    "http_url_to_repo":"https://source.example/kcf/existing.git"
  },
  {
    "id":11,
    "name":"Payment API",
    "path":"payment-api",
    "path_with_namespace":"kcf/payments/payment-api",
    "namespace":{"full_path":"kcf/payments"},
    "visibility":"private",
    "default_branch":"main",
    "http_url_to_repo":"https://source.example/kcf/payments/payment-api.git"
  }
]
JSON
        ;;
    https://target.example/api/v4/groups/target%2Fkcf%2Fpayments)
        status=404
        printf '{"message":"404 Group Not Found"}' >"$output"
        ;;
    https://target.example/api/v4/projects/target%2Fkcf%2Fexisting)
        cat >"$output" <<'JSON'
{
  "id":200,
  "name":"Existing",
  "path":"existing",
  "path_with_namespace":"target/kcf/existing",
  "namespace":{"full_path":"target/kcf"},
  "visibility":"private",
  "default_branch":"main",
  "http_url_to_repo":"https://target.example/target/kcf/existing.git"
}
JSON
        ;;
    https://target.example/api/v4/projects/target%2Fkcf%2Fpayments%2Fpayment-api)
        status=404
        printf '{"message":"404 Project Not Found"}' >"$output"
        ;;
    *)
        status=500
        printf '{"message":"unhandled fake route"}' >"$output"
        ;;
esac
fi

printf '%s' "$status"
FAKE
    chmod +x "$fake_bin/curl"
}

test_plan_with_fake_gitlab() {
    local fake_bin="$TEST_ROOT/fake-bin"
    local output="$TEST_ROOT/plan-output.txt"
    local work="$TEST_ROOT/plan-work"
    export FAKE_CURL_ARGS="$TEST_ROOT/fake-curl-args.txt"
    export FAKE_CURL_CALLS="$TEST_ROOT/fake-curl-calls.txt"
    : >"$FAKE_CURL_ARGS"
    : >"$FAKE_CURL_CALLS"
    make_fake_curl "$fake_bin"

    SOURCE_GITLAB_TOKEN="source-secret" \
    TARGET_GITLAB_TOKEN="target-secret" \
    PATH="$fake_bin:$PATH" \
        bash "$SCRIPT_DIR/migrate_gitlab_group.sh" \
            --source-url "https://source.example" \
            --target-url "https://target.example" \
            --source-group "kcf" \
            --target-group "target/kcf" \
            --work-dir "$work" >"$output"

    grep -q '^PLAN_CREATE_GROUP' "$output" ||
        fail "plan did not include subgroup creation"
    grep -q '^SKIP_PROJECT_EXISTS' "$output" ||
        fail "plan did not skip existing project"
    grep -q '^PLAN_CREATE_PROJECT' "$output" ||
        fail "plan did not include missing project"
    ! grep -q 'source-secret\|target-secret' "$FAKE_CURL_ARGS" ||
        fail "token leaked into curl process arguments"
    ! grep -q '^POST ' "$FAKE_CURL_CALLS" ||
        fail "plan mode sent a POST request"

    pass "end-to-end plan with fake GitLab API"
}

test_apply_creates_only_missing_objects() {
    local fake_bin="$TEST_ROOT/fake-apply-bin"
    local output="$TEST_ROOT/apply-output.txt"
    local work="$TEST_ROOT/apply-work"
    local post_count
    export FAKE_CURL_ARGS="$TEST_ROOT/fake-apply-curl-args.txt"
    export FAKE_CURL_CALLS="$TEST_ROOT/fake-apply-curl-calls.txt"
    : >"$FAKE_CURL_ARGS"
    : >"$FAKE_CURL_CALLS"
    make_fake_curl "$fake_bin"

    SOURCE_GITLAB_TOKEN="source-secret" \
    TARGET_GITLAB_TOKEN="target-secret" \
    PATH="$fake_bin:$PATH" \
        bash "$SCRIPT_DIR/migrate_gitlab_group.sh" \
            --source-url "https://source.example" \
            --target-url "https://target.example" \
            --source-group "kcf" \
            --target-group "target/kcf" \
            --work-dir "$work" \
            --apply >"$output"

    grep -q '^CREATE_GROUP' "$output" ||
        fail "apply did not create the missing subgroup"
    grep -q '^CREATE_PROJECT' "$output" ||
        fail "apply did not create the missing project"
    grep -q '^SKIP_PROJECT_EXISTS' "$output" ||
        fail "apply did not preserve the existing project"
    post_count=$(grep -c '^POST ' "$FAKE_CURL_CALLS")
    assert_eq "2" "$post_count" "number of create requests"
    grep -q $'^P\t11\t201\tkcf/payments/payment-api\ttarget/kcf/payments/payment-api\t1\tcreated\t0$' \
        "$work/state.tsv" ||
        fail "created project state was not recorded"
    ! grep -q 'source-secret\|target-secret' "$FAKE_CURL_ARGS" ||
        fail "token leaked into curl process arguments during apply"

    pass "apply creates only missing subgroup and project"
}

test_invalid_project_regex_is_rejected() {
    local output="$TEST_ROOT/invalid-regex-output.txt"

    if SOURCE_GITLAB_TOKEN="source-secret" \
        TARGET_GITLAB_TOKEN="target-secret" \
        bash "$SCRIPT_DIR/migrate_gitlab_group.sh" \
            --source-url "https://source.example" \
            --target-url "https://target.example" \
            --source-group "kcf" \
            --target-group "target/kcf" \
            --work-dir "$TEST_ROOT/invalid-regex-work" \
            --project-regex '[' >"$output" 2>&1
    then
        fail "invalid project regular expression was accepted"
    fi

    grep -q 'invalid --project-regex' "$output" ||
        fail "invalid project regular expression did not produce a clear error"

    pass "invalid project regular expression is rejected before API access"
}

git_run() {
    git "$@" >/dev/null
}

test_local_git_migration_and_extra_ref_detection() {
    local root="$TEST_ROOT/git"
    local source_bare="$root/source.git"
    local target_bare="$root/target.git"
    local working="$root/working"
    local source_refs="$root/source.refs"
    local target_refs="$root/target.refs"
    local main_id

    mkdir -p -- "$root"
    git_run init --bare "$source_bare"
    git_run init --bare "$target_bare"
    git_run init -b main "$working"
    git_run -C "$working" config user.name "Migration Test"
    git_run -C "$working" config user.email "migration@example.invalid"
    printf 'migration test\n' >"$working/README.md"
    git_run -C "$working" add README.md
    git_run -C "$working" commit -m initial
    git_run -C "$working" branch feature/test
    git_run -C "$working" tag v1.0.0
    git_run -C "$working" push "$source_bare" --all
    git_run -C "$working" push "$source_bare" --tags

    SOURCE_TOKEN="source-secret"
    TARGET_TOKEN="target-secret"
    WORK_DIR="$root/migration"
    TMP_ROOT="$root/tmp"
    MIGRATOR_TEST_MODE=1
    MIGRATE_LFS=0
    mkdir -p -- "$WORK_DIR" "$TMP_ROOT"

    migrate_repository \
        1 \
        "kcf/project" \
        "$(realpath "$source_bare" | sed 's#^#file://#')" \
        "target/kcf/project" \
        "$(realpath "$target_bare" | sed 's#^#file://#')"

    git --git-dir="$source_bare" for-each-ref \
        --format='%(objectname) %(refname)' refs/heads refs/tags |
        sort >"$source_refs"
    git --git-dir="$target_bare" for-each-ref \
        --format='%(objectname) %(refname)' refs/heads refs/tags |
        sort >"$target_refs"
    diff -u "$source_refs" "$target_refs" ||
        fail "local branch/tag migration differs"

    main_id=$(git --git-dir="$target_bare" rev-parse refs/heads/main)
    git --git-dir="$target_bare" update-ref refs/heads/target-only "$main_id"
    if migrate_repository \
        1 \
        "kcf/project" \
        "$(realpath "$source_bare" | sed 's#^#file://#')" \
        "target/kcf/project" \
        "$(realpath "$target_bare" | sed 's#^#file://#')" \
        >/dev/null 2>&1
    then
        fail "unexpected target ref was not rejected"
    fi

    pass "local Git branches/tags and unexpected ref rejection"
}

main() {
    select_test_json_backend
    test_json_records_and_payloads
    test_state_and_group_mapping
    test_plan_with_fake_gitlab
    test_apply_creates_only_missing_objects
    test_invalid_project_regex_is_rejected
    test_local_git_migration_and_extra_ref_detection
    printf 'All %s tests passed.\n' "$TEST_COUNT"
}

main "$@"
