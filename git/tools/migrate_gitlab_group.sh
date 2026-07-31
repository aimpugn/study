#!/usr/bin/env bash
#
# Create missing GitLab subgroups/projects and optionally migrate Git refs.
#
# Runtime dependencies:
#   - Bash 4+
#   - curl, git, base64 and standard GNU userland
#   - jq, or Perl with JSON::PP + MIME::Base64
#
# The default mode is plan-only. Existing target projects are never modified.

set -Eeuo pipefail
set +x
IFS=$'\n\t'
LC_ALL=C
export LC_ALL

readonly STATE_VERSION="1"
readonly SOURCE_HEADS_REFSPEC="+refs/heads/*:refs/heads/*"
readonly SOURCE_TAGS_REFSPEC="+refs/tags/*:refs/tags/*"
readonly TARGET_HEADS_REFSPEC="refs/heads/*:refs/heads/*"
readonly TARGET_TAGS_REFSPEC="refs/tags/*:refs/tags/*"

SOURCE_URL=""
TARGET_URL=""
SOURCE_GROUP=""
TARGET_GROUP=""
TARGET_GIT_URL_BASE=""
SOURCE_TOKEN_ENV="SOURCE_GITLAB_TOKEN"
TARGET_TOKEN_ENV="TARGET_GITLAB_TOKEN"
WORK_DIR="gitlab-group-migration-work"
PROJECT_REGEX=""
MAX_PROJECTS=""
REQUEST_TIMEOUT="30"
APPLY=0
MIGRATE_CREATED=0
MIGRATE_LFS=0
COPY_VISIBILITY=0
ALLOW_INSECURE_HTTP=0

SOURCE_TOKEN=""
TARGET_TOKEN=""
JSON_BACKEND=""
TMP_ROOT=""
STATE_FILE=""
API_COUNTER=0
FAILURES=0
CREATED_PROJECTS=0
SKIPPED_PROJECTS=0
VERIFIED_PROJECTS=0

usage() {
    cat <<'EOF'
Usage:
  migrate_gitlab_group.sh \
    --source-url URL \
    --target-url URL \
    --source-group GROUP_PATH \
    --target-group GROUP_PATH \
    [options]

Required:
  --source-url URL        Source GitLab base URL
  --target-url URL        Target GitLab base URL
  --source-group PATH     Source group full path
  --target-group PATH     Existing target root group full path

Options:
  --source-token-env NAME Environment variable holding the source token
                          (default: SOURCE_GITLAB_TOKEN)
  --target-token-env NAME Environment variable holding the target token
                          (default: TARGET_GITLAB_TOKEN)
  --work-dir DIR          Persistent state and bare repositories
                          (default: ./gitlab-group-migration-work)
  --project-regex REGEX   Process only matching source project full paths
  --max-projects N        Process at most N selected projects
  --request-timeout SEC   curl request timeout (default: 30)
  --target-git-url-base URL
                          Force target Git push/verify through this URL origin
                          instead of the API-advertised repository origin
  --apply                 Create missing subgroups and projects
  --migrate-created       Migrate branches/tags for tool-created projects
                          (requires --apply)
  --migrate-lfs           Also migrate all Git LFS objects
                          (requires --migrate-created)
  --copy-visibility       Copy source visibility instead of target defaults
  --allow-insecure-http   Allow tokens over plain HTTP
  -h, --help              Show this help

The safe default is plan-only. Target projects that already exist are skipped.
EOF
}

die() {
    printf 'ERROR: %s\n' "$*" >&2
    exit 1
}

event() {
    local action=$1
    local source=$2
    local target=${3-}
    local detail=${4-}
    local mapping=$source
    local suffix=""

    if [[ -n $target ]]; then
        mapping="${source} -> ${target}"
    fi
    if [[ -n $detail ]]; then
        suffix=" (${detail})"
    fi
    printf '%-24s %s%s\n' "$action" "$mapping" "$suffix"
}

cleanup() {
    if [[ -n ${TMP_ROOT:-} && -d $TMP_ROOT ]]; then
        rm -rf -- "$TMP_ROOT"
    fi
}

trap cleanup EXIT
trap 'printf "ERROR: interrupted\n" >&2; exit 130' INT TERM

parse_args() {
    while (($#)); do
        case $1 in
            --source-url)
                (($# >= 2)) || die "--source-url requires a value"
                SOURCE_URL=$2
                shift 2
                ;;
            --target-url)
                (($# >= 2)) || die "--target-url requires a value"
                TARGET_URL=$2
                shift 2
                ;;
            --source-group)
                (($# >= 2)) || die "--source-group requires a value"
                SOURCE_GROUP=$2
                shift 2
                ;;
            --target-group)
                (($# >= 2)) || die "--target-group requires a value"
                TARGET_GROUP=$2
                shift 2
                ;;
            --source-token-env)
                (($# >= 2)) || die "--source-token-env requires a value"
                SOURCE_TOKEN_ENV=$2
                shift 2
                ;;
            --target-token-env)
                (($# >= 2)) || die "--target-token-env requires a value"
                TARGET_TOKEN_ENV=$2
                shift 2
                ;;
            --work-dir)
                (($# >= 2)) || die "--work-dir requires a value"
                WORK_DIR=$2
                shift 2
                ;;
            --project-regex)
                (($# >= 2)) || die "--project-regex requires a value"
                PROJECT_REGEX=$2
                shift 2
                ;;
            --max-projects)
                (($# >= 2)) || die "--max-projects requires a value"
                MAX_PROJECTS=$2
                shift 2
                ;;
            --request-timeout)
                (($# >= 2)) || die "--request-timeout requires a value"
                REQUEST_TIMEOUT=$2
                shift 2
                ;;
            --target-git-url-base)
                (($# >= 2)) ||
                    die "--target-git-url-base requires a value"
                TARGET_GIT_URL_BASE=$2
                shift 2
                ;;
            --apply)
                APPLY=1
                shift
                ;;
            --migrate-created)
                MIGRATE_CREATED=1
                shift
                ;;
            --migrate-lfs)
                MIGRATE_LFS=1
                shift
                ;;
            --copy-visibility)
                COPY_VISIBILITY=1
                shift
                ;;
            --allow-insecure-http)
                ALLOW_INSECURE_HTTP=1
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                die "unknown argument: $1"
                ;;
        esac
    done
}

require_command() {
    command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

normalize_base_url() {
    local value=$1
    value=${value%/}
    value=${value%/api/v4}

    case $value in
        https://*)
            ;;
        http://*)
            ((ALLOW_INSECURE_HTTP)) ||
                die "plain HTTP requires --allow-insecure-http: $value"
            ;;
        *)
            die "invalid GitLab base URL: $value"
            ;;
    esac

    [[ $value != *"?"* && $value != *"#"* ]] ||
        die "GitLab base URL cannot contain query or fragment: $value"

    local authority=${value#*://}
    authority=${authority%%/*}
    [[ $authority != *"@"* ]] ||
        die "do not put credentials in a GitLab base URL"

    printf '%s\n' "$value"
}

normalize_git_url_base() {
    local value=$1
    value=${value%/}

    case $value in
        https://*)
            ;;
        http://*)
            ((ALLOW_INSECURE_HTTP)) ||
                die "plain HTTP requires --allow-insecure-http: $value"
            ;;
        *)
            die "invalid target Git URL base: $value"
            ;;
    esac

    [[ $value != *"?"* && $value != *"#"* ]] ||
        die "target Git URL base cannot contain query or fragment: $value"

    local authority=${value#*://}
    [[ -n $authority && $authority != *"/"* && $authority != *"@"* ]] ||
        die "--target-git-url-base must contain only scheme, host, and optional port"

    printf '%s\n' "$value"
}

trim_group_path() {
    local value=$1
    value=${value#/}
    value=${value%/}
    [[ -n $value ]] || die "group path cannot be empty"
    printf '%s\n' "$value"
}

validate_token() {
    local label=$1
    local token=$2

    [[ -n $token ]] || die "$label token is empty"
    [[ $token != *$'\n'* && $token != *$'\r'* ]] ||
        die "$label token contains a newline"
    [[ $token != *\"* ]] ||
        die "$label token contains a character unsafe for curl configuration"
    [[ $token != *\\* ]] ||
        die "$label token contains a character unsafe for curl configuration"
}

select_json_backend() {
    if command -v jq >/dev/null 2>&1; then
        JSON_BACKEND="jq"
        return
    fi

    if command -v perl >/dev/null 2>&1 &&
        perl -MJSON::PP -MMIME::Base64 -MEncode -e 'exit 0' \
            >/dev/null 2>&1; then
        JSON_BACKEND="perl"
        return
    fi

    die "JSON parser missing: install jq, or Perl with JSON::PP and MIME::Base64"
}

preflight() {
    ((BASH_VERSINFO[0] >= 4)) || die "Bash 4 or later is required"

    local command_name
    for command_name in \
        curl git base64 awk sort sed tr mktemp diff grep cut dirname mv
    do
        require_command "$command_name"
    done
    select_json_backend

    if ((MIGRATE_LFS)); then
        git lfs version >/dev/null 2>&1 ||
            die "--migrate-lfs requires git-lfs"
    fi

    [[ -n $SOURCE_URL ]] || die "--source-url is required"
    [[ -n $TARGET_URL ]] || die "--target-url is required"
    [[ -n $SOURCE_GROUP ]] || die "--source-group is required"
    [[ -n $TARGET_GROUP ]] || die "--target-group is required"

    ((MIGRATE_CREATED == 0 || APPLY == 1)) ||
        die "--migrate-created requires --apply"
    ((MIGRATE_LFS == 0 || MIGRATE_CREATED == 1)) ||
        die "--migrate-lfs requires --migrate-created"

    if [[ -n $MAX_PROJECTS ]]; then
        [[ $MAX_PROJECTS =~ ^[1-9][0-9]*$ ]] ||
            die "--max-projects must be a positive integer"
    fi
    [[ $REQUEST_TIMEOUT =~ ^[1-9][0-9]*$ ]] ||
        die "--request-timeout must be a positive integer"
    [[ $SOURCE_TOKEN_ENV =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] ||
        die "invalid --source-token-env name"
    [[ $TARGET_TOKEN_ENV =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] ||
        die "invalid --target-token-env name"

    if [[ -n $PROJECT_REGEX ]]; then
        local regex_status=0
        grep -E "$PROJECT_REGEX" </dev/null >/dev/null 2>&1 ||
            regex_status=$?
        ((regex_status != 2)) || die "invalid --project-regex"
    fi

    SOURCE_URL=$(normalize_base_url "$SOURCE_URL")
    TARGET_URL=$(normalize_base_url "$TARGET_URL")
    SOURCE_GROUP=$(trim_group_path "$SOURCE_GROUP")
    TARGET_GROUP=$(trim_group_path "$TARGET_GROUP")
    if [[ -n $TARGET_GIT_URL_BASE ]]; then
        TARGET_GIT_URL_BASE=$(normalize_git_url_base "$TARGET_GIT_URL_BASE")
    fi

    [[ $SOURCE_URL != "$TARGET_URL" || $SOURCE_GROUP != "$TARGET_GROUP" ]] ||
        die "source and target identify the same GitLab group"

    SOURCE_TOKEN=${!SOURCE_TOKEN_ENV-}
    TARGET_TOKEN=${!TARGET_TOKEN_ENV-}
    validate_token "source" "$SOURCE_TOKEN"
    validate_token "target" "$TARGET_TOKEN"

    mkdir -p -- "$WORK_DIR"
    WORK_DIR=$(cd -- "$WORK_DIR" && pwd -P)
    STATE_FILE="$WORK_DIR/state.tsv"
    TMP_ROOT=$(mktemp -d "$WORK_DIR/.tmp.XXXXXX")
}

urlencode() {
    local input=$1
    local output=""
    local character
    local index

    for ((index = 0; index < ${#input}; index++)); do
        character=${input:index:1}
        case $character in
            [a-zA-Z0-9.~_-])
                output+=$character
                ;;
            *)
                printf -v output '%s%%%02X' "$output" "'$character"
                ;;
        esac
    done
    printf '%s\n' "$output"
}

b64_encode() {
    printf '%s' "$1" | base64 | tr -d '\n'
}

b64_decode() {
    [[ -n $1 ]] || return 0
    printf '%s' "$1" | base64 --decode
}

json_array_length() {
    local file=$1
    if [[ $JSON_BACKEND == "jq" ]]; then
        jq -er 'if type == "array" then length else error("not an array") end' \
            "$file"
        return
    fi

    perl -MJSON::PP -0777 -e '
        my $file = shift;
        open my $fh, "<:raw", $file or die "$file: $!";
        local $/;
        my $value = decode_json(<$fh>);
        ref($value) eq "ARRAY" or die "not an array\n";
        print scalar(@$value), "\n";
    ' "$file"
}

json_emit_records() {
    local kind=$1
    local file=$2

    if [[ $JSON_BACKEND == "jq" ]]; then
        case $kind in
            group-array)
                jq -r '
                    .[] |
                    [
                        (.id | tostring),
                        (.name | @base64),
                        .path,
                        .full_path,
                        (.visibility // "")
                    ] | @tsv
                ' "$file"
                ;;
            group-object)
                jq -r '
                    [
                        (.id | tostring),
                        (.name | @base64),
                        .path,
                        .full_path,
                        (.visibility // "")
                    ] | @tsv
                ' "$file"
                ;;
            project-array)
                jq -r '
                    .[] |
                    [
                        (.id | tostring),
                        (.name | @base64),
                        .path,
                        .path_with_namespace,
                        .namespace.full_path,
                        (.visibility // ""),
                        ((.default_branch // "") | @base64),
                        (.http_url_to_repo | @base64)
                    ] | @tsv
                ' "$file"
                ;;
            project-object)
                jq -r '
                    [
                        (.id | tostring),
                        (.name | @base64),
                        .path,
                        .path_with_namespace,
                        .namespace.full_path,
                        (.visibility // ""),
                        ((.default_branch // "") | @base64),
                        (.http_url_to_repo | @base64)
                    ] | @tsv
                ' "$file"
                ;;
            *)
                die "unknown JSON record kind: $kind"
                ;;
        esac
        return
    fi

    perl -MJSON::PP -MMIME::Base64 -MEncode -0777 -e '
        my ($kind, $file) = @ARGV;
        open my $fh, "<:raw", $file or die "$file: $!";
        local $/;
        my $value = decode_json(<$fh>);
        my @items = $kind =~ /-array$/ ? @$value : ($value);
        sub b64 {
            my ($item) = @_;
            $item = "" unless defined $item;
            return encode_base64(encode_utf8("$item"), "");
        }
        for my $item (@items) {
            if ($kind =~ /^group-/) {
                print join("\t",
                    $item->{id},
                    b64($item->{name}),
                    $item->{path},
                    $item->{full_path},
                    ($item->{visibility} // "")
                ), "\n";
            } elsif ($kind =~ /^project-/) {
                print join("\t",
                    $item->{id},
                    b64($item->{name}),
                    $item->{path},
                    $item->{path_with_namespace},
                    $item->{namespace}->{full_path},
                    ($item->{visibility} // ""),
                    b64($item->{default_branch}),
                    b64($item->{http_url_to_repo})
                ), "\n";
            } else {
                die "unknown record kind: $kind\n";
            }
        }
    ' "$kind" "$file"
}

json_make_payload() {
    local kind=$1
    local output=$2
    shift 2

    if [[ $JSON_BACKEND == "jq" ]]; then
        case $kind in
            group)
                local name=$1 path=$2 parent_id=$3 visibility=$4 copy=$5
                jq -n \
                    --arg name "$name" \
                    --arg path "$path" \
                    --argjson parent_id "$parent_id" \
                    --arg visibility "$visibility" \
                    --argjson copy "$copy" '
                        {
                            name: $name,
                            path: $path,
                            parent_id: $parent_id
                        } +
                        (if $copy then {visibility: $visibility} else {} end)
                    ' >"$output"
                ;;
            project)
                local name=$1 path=$2 namespace_id=$3 visibility=$4 copy=$5
                jq -n \
                    --arg name "$name" \
                    --arg path "$path" \
                    --argjson namespace_id "$namespace_id" \
                    --arg visibility "$visibility" \
                    --argjson copy "$copy" '
                        {
                            name: $name,
                            path: $path,
                            namespace_id: $namespace_id,
                            initialize_with_readme: false
                        } +
                        (if $copy then {visibility: $visibility} else {} end)
                    ' >"$output"
                ;;
            default-branch)
                jq -n --arg default_branch "$1" \
                    '{default_branch: $default_branch}' >"$output"
                ;;
            *)
                die "unknown JSON payload kind: $kind"
                ;;
        esac
        return
    fi

    perl -MJSON::PP -MEncode -e '
        my ($kind, $output, @args) = @ARGV;
        my $payload;
        if ($kind eq "group") {
            my ($name, $path, $parent_id, $visibility, $copy) = @args;
            $payload = {
                name => decode_utf8($name),
                path => $path,
                parent_id => 0 + $parent_id
            };
            $payload->{visibility} = $visibility if $copy eq "true";
        } elsif ($kind eq "project") {
            my ($name, $path, $namespace_id, $visibility, $copy) = @args;
            $payload = {
                name => decode_utf8($name),
                path => $path,
                namespace_id => 0 + $namespace_id,
                initialize_with_readme => JSON::PP::false
            };
            $payload->{visibility} = $visibility if $copy eq "true";
        } elsif ($kind eq "default-branch") {
            $payload = {default_branch => $args[0]};
        } else {
            die "unknown payload kind: $kind\n";
        }
        open my $fh, ">:raw", $output or die "$output: $!";
        print {$fh} JSON::PP->new->utf8->canonical->encode($payload);
    ' "$kind" "$output" "$@"
}

api_request() {
    local token=$1
    local method=$2
    local url=$3
    local payload=$4
    local output=$5
    local allow_404=${6:-0}
    local headers="$TMP_ROOT/headers.$API_COUNTER"
    local status
    local protocol="=https"
    local -a curl_args=(
        --silent
        --show-error
        --request "$method"
        --connect-timeout 15
        --max-time "$REQUEST_TIMEOUT"
        --max-redirs 0
        --output "$output"
        --dump-header "$headers"
        --write-out '%{http_code}'
    )

    ((API_COUNTER += 1))
    ((ALLOW_INSECURE_HTTP)) && protocol="=http,https"
    curl_args+=(--proto "$protocol")

    if [[ $method == "GET" || $method == "HEAD" || $method == "PUT" ]]; then
        curl_args+=(
            --retry 3
            --retry-delay 1
            --retry-max-time 30
        )
    fi
    if [[ -n $payload ]]; then
        curl_args+=(
            --header 'Content-Type: application/json'
            --data-binary "@$payload"
        )
    fi

    if ! status=$(
        printf 'header = "PRIVATE-TOKEN: %s"\n' "$token" |
            curl --config - "${curl_args[@]}" "$url"
    ); then
        printf 'ERROR: %s %s failed at the transport layer\n' \
            "$method" "$url" >&2
        return 1
    fi

    if [[ $status =~ ^2[0-9][0-9]$ ]]; then
        return 0
    fi
    if [[ $status == "404" && $allow_404 == "1" ]]; then
        return 44
    fi

    local error_body=""
    if [[ -f $output ]]; then
        error_body=$(tr '\r\n\t' '   ' <"$output")
        error_body=${error_body//"$token"/<redacted>}
        error_body=${error_body:0:4000}
    fi
    printf 'ERROR: %s %s failed with HTTP %s: %s\n' \
        "$method" "$url" "$status" "$error_body" >&2
    return 1
}

paginate_records() {
    local token=$1
    local base_url=$2
    local endpoint=$3
    local record_kind=$4
    local output=$5
    local page=1
    local page_file
    local count
    local separator="?"

    [[ $endpoint == *"?"* ]] && separator="&"
    : >"$output"

    while :; do
        page_file="$TMP_ROOT/page.${API_COUNTER}.json"
        api_request \
            "$token" \
            "GET" \
            "${base_url}/api/v4${endpoint}${separator}page=${page}&per_page=100" \
            "" \
            "$page_file" ||
            return 1
        count=$(json_array_length "$page_file") || return 1
        json_emit_records "$record_kind" "$page_file" >>"$output" ||
            return 1
        ((count < 100)) && break
        ((page += 1))
    done
}

state_init() {
    local context
    context=$(
        printf 'CONTEXT\t%s\t%s\t%s\t%s' \
            "$(b64_encode "$SOURCE_URL")" \
            "$(b64_encode "$SOURCE_GROUP")" \
            "$(b64_encode "$TARGET_URL")" \
            "$(b64_encode "$TARGET_GROUP")"
    )

    if [[ ! -f $STATE_FILE ]]; then
        {
            printf '# gitlab-group-migration-state-v%s\n' "$STATE_VERSION"
            printf '%s\n' "$context"
        } >"$STATE_FILE"
        return
    fi

    local version_line
    local context_line
    version_line=$(sed -n '1p' "$STATE_FILE")
    context_line=$(sed -n '2p' "$STATE_FILE")
    [[ $version_line == "# gitlab-group-migration-state-v$STATE_VERSION" ]] ||
        die "unsupported state version: $STATE_FILE"
    [[ $context_line == "$context" ]] ||
        die "state context mismatch; use a different --work-dir"
}

state_get() {
    local source_id=$1
    awk -F '\t' -v source_id="$source_id" '
        $1 == "P" && $2 == source_id {
            print
            found = 1
            exit
        }
        END {
            if (!found) {
                exit 1
            }
        }
    ' "$STATE_FILE"
}

state_update() {
    local source_id=$1
    local target_id=$2
    local source_path=$3
    local target_path=$4
    local created_by_tool=$5
    local status=$6
    local lfs_migrated=$7
    local temporary="$STATE_FILE.tmp"

    awk -F '\t' -v OFS='\t' \
        -v source_id="$source_id" \
        -v target_id="$target_id" \
        -v source_path="$source_path" \
        -v target_path="$target_path" \
        -v created="$created_by_tool" \
        -v status="$status" \
        -v lfs="$lfs_migrated" '
        $1 == "P" && $2 == source_id {
            print "P", source_id, target_id, source_path, target_path,
                created, status, lfs
            updated = 1
            next
        }
        {
            print
        }
        END {
            if (!updated) {
                print "P", source_id, target_id, source_path, target_path,
                    created, status, lfs
            }
        }
    ' "$STATE_FILE" >"$temporary"
    mv -- "$temporary" "$STATE_FILE"
}

map_group_path() {
    local source_path=$1
    if [[ $source_path == "$SOURCE_GROUP" ]]; then
        printf '%s\n' "$TARGET_GROUP"
        return
    fi
    [[ $source_path == "$SOURCE_GROUP/"* ]] ||
        die "$source_path is outside source group $SOURCE_GROUP"
    printf '%s/%s\n' "$TARGET_GROUP" "${source_path#"$SOURCE_GROUP/"}"
}

target_group_id() {
    local map_file=$1
    local group_path=$2
    awk -F '\t' -v path="$group_path" '
        $1 == path {
            print $2
            found = 1
            exit
        }
        END {
            if (!found) {
                exit 1
            }
        }
    ' "$map_file"
}

select_projects() {
    local all_projects=$1
    local selected=$2
    local required_groups=$3
    local selected_count=0
    local id name_b64 path full_path namespace visibility
    local default_branch_b64 http_url_b64

    : >"$selected"
    : >"$required_groups"

    while IFS=$'\t' read -r \
        id name_b64 path full_path namespace visibility \
        default_branch_b64 http_url_b64
    do
        [[ -n $id ]] || continue
        if [[ -n $PROJECT_REGEX && ! $full_path =~ $PROJECT_REGEX ]]; then
            continue
        fi

        printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
            "$id" "$name_b64" "$path" "$full_path" "$namespace" \
            "$visibility" "$default_branch_b64" "$http_url_b64" \
            >>"$selected"
        ((selected_count += 1))

        if [[ -n $PROJECT_REGEX || -n $MAX_PROJECTS ]]; then
            local current_namespace=$namespace
            while [[ $current_namespace != "$SOURCE_GROUP" ]]; do
                [[ $current_namespace == "$SOURCE_GROUP/"* ]] ||
                    die "project namespace is outside source group: $namespace"
                printf '%s\n' "$current_namespace" >>"$required_groups"
                current_namespace=${current_namespace%/*}
            done
        fi

        if [[ -n $MAX_PROJECTS && $selected_count -ge $MAX_PROJECTS ]]; then
            break
        fi
    done <"$all_projects"

    if [[ -n $PROJECT_REGEX && $selected_count -eq 0 ]]; then
        die "no source projects match --project-regex"
    fi

    sort -u -o "$required_groups" "$required_groups"
    printf '%s\n' "$selected_count"
}

select_groups() {
    local all_groups=$1
    local required_groups=$2
    local selected_groups=$3
    local filtered="$TMP_ROOT/groups.filtered.tsv"
    local id name_b64 path full_path visibility

    : >"$filtered"
    while IFS=$'\t' read -r id name_b64 path full_path visibility; do
        [[ -n $id ]] || continue
        if [[ -n $PROJECT_REGEX || -n $MAX_PROJECTS ]]; then
            grep -Fqx -- "$full_path" "$required_groups" || continue
        fi
        printf '%s\t%s\t%s\t%s\t%s\n' \
            "$id" "$name_b64" "$path" "$full_path" "$visibility" \
            >>"$filtered"
    done <"$all_groups"

    awk -F '\t' 'BEGIN {OFS = "\t"} {
        copy = $4
        depth = gsub(/\//, "/", copy)
        print depth, $0
    }' "$filtered" |
        sort -t $'\t' -k1,1n -k5,5 |
        cut -f2- >"$selected_groups"
}

get_optional_object() {
    local token=$1
    local base_url=$2
    local endpoint=$3
    local output=$4
    local status

    if api_request \
        "$token" \
        "GET" \
        "${base_url}/api/v4${endpoint}" \
        "" \
        "$output" \
        1
    then
        return 0
    else
        status=$?
        ((status == 44)) && return 44
        return 1
    fi
}

provision_groups() {
    local selected_groups=$1
    local target_root_id=$2
    local target_map=$3
    local id name_b64 path source_full visibility
    local target_full target_body target_record target_id
    local parent_path parent_id name payload response

    printf '%s\t%s\n' "$TARGET_GROUP" "$target_root_id" >"$target_map"

    while IFS=$'\t' read -r id name_b64 path source_full visibility; do
        [[ -n $id ]] || continue
        target_full=$(map_group_path "$source_full")
        target_body="$TMP_ROOT/target-group.${API_COUNTER}.json"

        if get_optional_object \
            "$TARGET_TOKEN" \
            "$TARGET_URL" \
            "/groups/$(urlencode "$target_full")" \
            "$target_body"
        then
            target_record=$(json_emit_records group-object "$target_body") ||
                return 1
            IFS=$'\t' read -r target_id _ <<<"$target_record"
            printf '%s\t%s\n' "$target_full" "$target_id" >>"$target_map"
            event "SKIP_GROUP_EXISTS" "$source_full" "$target_full"
            continue
        else
            local optional_status=$?
            ((optional_status == 44)) || return 1
        fi

        if ((APPLY == 0)); then
            event "PLAN_CREATE_GROUP" "$source_full" "$target_full"
            continue
        fi

        parent_path=${target_full%/*}
        parent_id=$(target_group_id "$target_map" "$parent_path") ||
            die "target parent group is missing: $parent_path"
        name=$(b64_decode "$name_b64")
        payload="$TMP_ROOT/group-payload.${API_COUNTER}.json"
        response="$TMP_ROOT/group-response.${API_COUNTER}.json"
        json_make_payload \
            group \
            "$payload" \
            "$name" \
            "$path" \
            "$parent_id" \
            "$visibility" \
            "$([[ $COPY_VISIBILITY == 1 ]] && printf true || printf false)"

        api_request \
            "$TARGET_TOKEN" \
            "POST" \
            "$TARGET_URL/api/v4/groups" \
            "$payload" \
            "$response" ||
            return 1
        target_record=$(json_emit_records group-object "$response") ||
            return 1
        IFS=$'\t' read -r target_id _ <<<"$target_record"
        printf '%s\t%s\n' "$target_full" "$target_id" >>"$target_map"
        event "CREATE_GROUP" "$source_full" "$target_full"
    done <"$selected_groups"
}

validate_repository_url() {
    local url=$1

    if [[ ${MIGRATOR_TEST_MODE:-0} == "1" && $url == file://* ]]; then
        return
    fi
    case $url in
        https://*)
            ;;
        http://*)
            if ((ALLOW_INSECURE_HTTP == 0)); then
                printf 'ERROR: repository URL uses plain HTTP: %s\n' \
                    "$url" >&2
                return 1
            fi
            ;;
        *)
            printf 'ERROR: repository URL must use HTTP(S): %s\n' \
                "$url" >&2
            return 1
            ;;
    esac
    local authority=${url#*://}
    authority=${authority%%/*}
    if [[ $authority == *"@"* ]]; then
        printf 'ERROR: repository URL contains credentials: %s\n' \
            "$url" >&2
        return 1
    fi
}

target_repository_url() {
    local advertised_url=$1
    local authority without_scheme path rewritten_url

    if [[ -z $TARGET_GIT_URL_BASE ]]; then
        validate_repository_url "$advertised_url" || return 1
        printf '%s\n' "$advertised_url"
        return
    fi

    case $advertised_url in
        https://*|http://*)
            ;;
        *)
            printf 'ERROR: API-advertised repository URL must use HTTP(S): %s\n' \
                "$advertised_url" >&2
            return 1
            ;;
    esac
    authority=${advertised_url#*://}
    authority=${authority%%/*}
    if [[ $authority == *"@"* ]]; then
        printf 'ERROR: API-advertised repository URL contains credentials: %s\n' \
            "$advertised_url" >&2
        return 1
    fi

    without_scheme=${advertised_url#*://}
    [[ $without_scheme == */* ]] || {
        printf 'ERROR: target repository URL has no path: %s\n' \
            "$advertised_url" >&2
        return 1
    }
    path=/${without_scheme#*/}
    rewritten_url="${TARGET_GIT_URL_BASE}${path}"
    validate_repository_url "$rewritten_url" || return 1
    printf '%s\n' "$rewritten_url"
}

git_with_token() {
    local token=$1
    local url=$2
    shift 2
    local credential
    credential=$(printf 'oauth2:%s' "$token" | base64 | tr -d '\n')

    GIT_TERMINAL_PROMPT=0 \
    GIT_CONFIG_COUNT=2 \
    GIT_CONFIG_KEY_0="http.${url}.extraHeader" \
    GIT_CONFIG_VALUE_0="Authorization: Basic ${credential}" \
    GIT_CONFIG_KEY_1="http.followRedirects" \
    GIT_CONFIG_VALUE_1="false" \
        git "$@"
}

ensure_remote() {
    local repository=$1
    local name=$2
    local url=$3

    if git -C "$repository" remote get-url "$name" >/dev/null 2>&1; then
        git -C "$repository" remote set-url "$name" "$url"
    else
        git -C "$repository" remote add "$name" "$url"
    fi
}

migrate_repository() {
    local source_id=$1
    local source_path=$2
    local source_repo_url=$3
    local target_path=$4
    local advertised_target_repo_url=$5
    local target_repo_url safe_name repository expected actual_raw actual

    validate_repository_url "$source_repo_url" || return 1
    target_repo_url=$(target_repository_url "$advertised_target_repo_url") ||
        return 1
    if [[ $target_repo_url != "$advertised_target_repo_url" ]]; then
        event \
            "OVERRIDE_TARGET_GIT_URL" \
            "$advertised_target_repo_url" \
            "$target_repo_url"
    fi

    safe_name=$(printf '%s' "$target_path" | tr -c 'A-Za-z0-9._-' '_')
    repository="$WORK_DIR/repositories/${source_id}-${safe_name}.git"
    mkdir -p -- "$(dirname -- "$repository")" || return 1

    if [[ ! -d $repository ]]; then
        git init --bare "$repository" >/dev/null || return 1
    fi
    ensure_remote "$repository" source "$source_repo_url" || return 1
    ensure_remote "$repository" target "$target_repo_url" || return 1

    git_with_token \
        "$SOURCE_TOKEN" \
        "$source_repo_url" \
        -C "$repository" \
        fetch \
        --prune \
        --progress \
        source \
        "$SOURCE_HEADS_REFSPEC" \
        "$SOURCE_TAGS_REFSPEC" ||
        return 1
    git -C "$repository" fsck --full || return 1

    expected="$TMP_ROOT/expected-refs.${source_id}"
    git -C "$repository" for-each-ref \
        --format='%(objectname) %(refname)' \
        refs/heads refs/tags |
        sort >"$expected" ||
        return 1

    if ((MIGRATE_LFS)); then
        git_with_token \
            "$SOURCE_TOKEN" \
            "$source_repo_url" \
            -C "$repository" \
            lfs fetch --all source ||
            return 1
        git_with_token \
            "$TARGET_TOKEN" \
            "$target_repo_url" \
            -C "$repository" \
            lfs push --all target ||
            return 1
    fi

    if [[ -s $expected ]]; then
        git_with_token \
            "$TARGET_TOKEN" \
            "$target_repo_url" \
            -C "$repository" \
            push \
            --dry-run \
            --porcelain \
            target \
            "$TARGET_HEADS_REFSPEC" \
            "$TARGET_TAGS_REFSPEC" ||
            return 1
        git_with_token \
            "$TARGET_TOKEN" \
            "$target_repo_url" \
            -C "$repository" \
            push \
            --atomic \
            --progress \
            --porcelain \
            target \
            "$TARGET_HEADS_REFSPEC" \
            "$TARGET_TAGS_REFSPEC" ||
            return 1
    else
        event "EMPTY_SOURCE" "$source_path" "$target_path"
    fi

    actual_raw="$TMP_ROOT/actual-refs-raw.${source_id}"
    actual="$TMP_ROOT/actual-refs.${source_id}"
    git_with_token \
        "$TARGET_TOKEN" \
        "$target_repo_url" \
        -C "$repository" \
        ls-remote --heads --tags target >"$actual_raw" ||
        return 1
    awk '$2 !~ /\^\{\}$/ {print $1 " " $2}' "$actual_raw" |
        sort >"$actual" ||
        return 1

    if ! diff -u "$expected" "$actual"; then
        printf 'ERROR: target ref verification failed: %s\n' \
            "$target_path" >&2
        return 1
    fi
}

migrate_and_record() {
    local source_id=$1
    local source_path=$2
    local source_repo_url=$3
    local target_id=$4
    local target_path=$5
    local target_repo_url=$6
    local default_branch=$7
    local lfs_state=0

    event "MIGRATE_REFS" "$source_path" "$target_path"
    if ! migrate_repository \
        "$source_id" \
        "$source_path" \
        "$source_repo_url" \
        "$target_path" \
        "$target_repo_url"
    then
        state_update \
            "$source_id" \
            "$target_id" \
            "$source_path" \
            "$target_path" \
            1 \
            "migration_failed" \
            0
        event "MIGRATION_FAIL" "$source_path" "$target_path"
        ((FAILURES += 1))
        return
    fi

    if [[ -n $default_branch ]]; then
        local payload="$TMP_ROOT/default-branch.${API_COUNTER}.json"
        local response="$TMP_ROOT/default-branch-response.${API_COUNTER}.json"
        json_make_payload default-branch "$payload" "$default_branch"
        if ! api_request \
            "$TARGET_TOKEN" \
            "PUT" \
            "$TARGET_URL/api/v4/projects/$target_id" \
            "$payload" \
            "$response"
        then
            state_update \
                "$source_id" \
                "$target_id" \
                "$source_path" \
                "$target_path" \
                1 \
                "migration_failed" \
                0
            event "MIGRATION_FAIL" \
                "$source_path" \
                "$target_path" \
                "default branch update failed"
            ((FAILURES += 1))
            return
        fi
    fi

    ((MIGRATE_LFS)) && lfs_state=1
    state_update \
        "$source_id" \
        "$target_id" \
        "$source_path" \
        "$target_path" \
        1 \
        "verified" \
        "$lfs_state"
    event "VERIFY_PASS" "$source_path" "$target_path"
    ((VERIFIED_PROJECTS += 1))
}

provision_projects() {
    local selected_projects=$1
    local target_map=$2
    local source_id name_b64 path source_path source_namespace visibility
    local default_branch_b64 source_repo_url_b64
    local target_namespace target_path target_body target_record
    local target_id target_repo_url_b64
    local target_repo_url default_branch source_repo_url
    local state_record state_target_id state_target_path state_created
    local state_status state_lfs state_owned needs_lfs
    local namespace_id name payload response

    while IFS=$'\t' read -r \
        source_id name_b64 path source_path source_namespace visibility \
        default_branch_b64 source_repo_url_b64
    do
        [[ -n $source_id ]] || continue
        target_namespace=$(map_group_path "$source_namespace")
        target_path="$target_namespace/$path"
        target_body="$TMP_ROOT/target-project.${API_COUNTER}.json"
        state_record=""
        state_record=$(state_get "$source_id" 2>/dev/null || true)

        if get_optional_object \
            "$TARGET_TOKEN" \
            "$TARGET_URL" \
            "/projects/$(urlencode "$target_path")" \
            "$target_body"
        then
            target_record=$(json_emit_records project-object "$target_body") ||
                return 1
            IFS=$'\t' read -r \
                target_id _ _ _ _ _ _ target_repo_url_b64 \
                <<<"$target_record"
            target_repo_url=$(b64_decode "$target_repo_url_b64")

            if [[ -n $state_record ]]; then
                IFS=$'\t' read -r \
                    _ _ state_target_id _ state_target_path state_created \
                    state_status state_lfs <<<"$state_record"
                if [[ $state_status == "creation_failed" &&
                    $state_target_path == "$target_path" ]]
                then
                    event \
                        "BLOCK_CREATE_AMBIGUOUS" \
                        "$source_path" \
                        "$target_path" \
                        "project appeared after a failed create request"
                    ((FAILURES += 1))
                    continue
                fi
            else
                state_target_id=""
                state_target_path=""
                state_created=0
                state_status=""
                state_lfs=0
            fi

            state_owned=0
            if [[ $state_created == "1" &&
                $state_target_id == "$target_id" &&
                $state_target_path == "$target_path" ]]
            then
                state_owned=1
            fi
            needs_lfs=0
            if ((MIGRATE_LFS)) && [[ $state_lfs != "1" ]]; then
                needs_lfs=1
            fi

            if ((state_owned && APPLY && MIGRATE_CREATED)) &&
                [[ $state_status != "verified" || $needs_lfs == "1" ]]
            then
                if [[ $state_status == "verified" ]]; then
                    event "RESUME_LFS" "$source_path" "$target_path"
                else
                    event "RESUME_CREATED" "$source_path" "$target_path"
                fi
                default_branch=$(b64_decode "$default_branch_b64")
                source_repo_url=$(b64_decode "$source_repo_url_b64")
                migrate_and_record \
                    "$source_id" \
                    "$source_path" \
                    "$source_repo_url" \
                    "$target_id" \
                    "$target_path" \
                    "$target_repo_url" \
                    "$default_branch"
            elif ((state_owned)) && [[ $state_status == "verified" ]]; then
                event "SKIP_VERIFIED" "$source_path" "$target_path"
            else
                event "SKIP_PROJECT_EXISTS" "$source_path" "$target_path"
                ((SKIPPED_PROJECTS += 1))
            fi
            continue
        else
            local optional_status=$?
            ((optional_status == 44)) || return 1
        fi

        if ((APPLY == 0)); then
            event "PLAN_CREATE_PROJECT" "$source_path" "$target_path"
            continue
        fi

        namespace_id=$(target_group_id "$target_map" "$target_namespace") ||
            die "target namespace is missing: $target_namespace"
        name=$(b64_decode "$name_b64")
        payload="$TMP_ROOT/project-payload.${API_COUNTER}.json"
        response="$TMP_ROOT/project-response.${API_COUNTER}.json"
        json_make_payload \
            project \
            "$payload" \
            "$name" \
            "$path" \
            "$namespace_id" \
            "$visibility" \
            "$([[ $COPY_VISIBILITY == 1 ]] && printf true || printf false)"

        if ! api_request \
            "$TARGET_TOKEN" \
            "POST" \
            "$TARGET_URL/api/v4/projects" \
            "$payload" \
            "$response"
        then
            state_update \
                "$source_id" \
                0 \
                "$source_path" \
                "$target_path" \
                0 \
                "creation_failed" \
                0
            event "CREATE_PROJECT_FAIL" "$source_path" "$target_path"
            ((FAILURES += 1))
            continue
        fi

        target_record=$(json_emit_records project-object "$response") ||
            return 1
        IFS=$'\t' read -r \
            target_id _ _ _ _ _ _ target_repo_url_b64 \
            <<<"$target_record"
        target_repo_url=$(b64_decode "$target_repo_url_b64")
        state_update \
            "$source_id" \
            "$target_id" \
            "$source_path" \
            "$target_path" \
            1 \
            "created" \
            0
        event "CREATE_PROJECT" "$source_path" "$target_path"
        ((CREATED_PROJECTS += 1))

        if ((MIGRATE_CREATED)); then
            default_branch=$(b64_decode "$default_branch_b64")
            source_repo_url=$(b64_decode "$source_repo_url_b64")
            migrate_and_record \
                "$source_id" \
                "$source_path" \
                "$source_repo_url" \
                "$target_id" \
                "$target_path" \
                "$target_repo_url" \
                "$default_branch"
        fi
    done <"$selected_projects"
}

main() {
    parse_args "$@"
    preflight
    state_init

    local source_root="$TMP_ROOT/source-root.json"
    local target_root="$TMP_ROOT/target-root.json"
    local source_root_record target_root_record source_root_id target_root_id
    local all_groups="$TMP_ROOT/all-groups.tsv"
    local all_projects="$TMP_ROOT/all-projects.tsv"
    local selected_groups="$TMP_ROOT/selected-groups.tsv"
    local selected_projects="$TMP_ROOT/selected-projects.tsv"
    local required_groups="$TMP_ROOT/required-groups.txt"
    local target_map="$TMP_ROOT/target-groups.tsv"
    local selected_count
    local group_count
    local project_count

    api_request \
        "$SOURCE_TOKEN" \
        "GET" \
        "$SOURCE_URL/api/v4/groups/$(urlencode "$SOURCE_GROUP")" \
        "" \
        "$source_root" ||
        die "source group does not exist or is not visible: $SOURCE_GROUP"
    source_root_record=$(json_emit_records group-object "$source_root")
    IFS=$'\t' read -r source_root_id _ <<<"$source_root_record"

    api_request \
        "$TARGET_TOKEN" \
        "GET" \
        "$TARGET_URL/api/v4/groups/$(urlencode "$TARGET_GROUP")" \
        "" \
        "$target_root" ||
        die "target root group must already exist: $TARGET_GROUP"
    target_root_record=$(json_emit_records group-object "$target_root")
    IFS=$'\t' read -r target_root_id _ <<<"$target_root_record"

    paginate_records \
        "$SOURCE_TOKEN" \
        "$SOURCE_URL" \
        "/groups/$source_root_id/descendant_groups?order_by=id&sort=asc" \
        group-array \
        "$all_groups"
    paginate_records \
        "$SOURCE_TOKEN" \
        "$SOURCE_URL" \
        "/groups/$source_root_id/projects?include_subgroups=true&with_shared=false&simple=false&order_by=id&sort=asc" \
        project-array \
        "$all_projects"

    selected_count=$(
        select_projects "$all_projects" "$selected_projects" "$required_groups"
    )
    select_groups "$all_groups" "$required_groups" "$selected_groups"
    group_count=$(awk 'END {print NR + 0}' "$selected_groups")
    project_count=$(awk 'END {print NR + 0}' "$all_projects")

    event \
        "DISCOVER" \
        "$SOURCE_GROUP" \
        "$TARGET_GROUP" \
        "$group_count groups, $selected_count/$project_count projects selected"

    provision_groups \
        "$selected_groups" \
        "$target_root_id" \
        "$target_map"
    provision_projects "$selected_projects" "$target_map"

    event \
        "SUMMARY" \
        "$TARGET_GROUP" \
        "" \
        "created=$CREATED_PROJECTS, existing=$SKIPPED_PROJECTS, verified=$VERIFIED_PROJECTS, failures=$FAILURES"

    ((FAILURES == 0))
}

if [[ ${BASH_SOURCE[0]} == "$0" ]]; then
    main "$@"
fi
