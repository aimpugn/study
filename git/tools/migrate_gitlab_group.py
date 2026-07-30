#!/usr/bin/env python3
"""Provision missing GitLab groups/projects and optionally migrate Git refs.

The safe default is plan-only. Existing target projects are never modified.
Only projects created by this tool and recorded in its state file are eligible
for the optional Git migration phase.
"""

from __future__ import annotations

import argparse
import base64
import json
import os
import re
import shutil
import subprocess
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Mapping, Sequence


SOURCE_HEADS_REFSPEC = "+refs/heads/*:refs/heads/*"
SOURCE_TAGS_REFSPEC = "+refs/tags/*:refs/tags/*"
TARGET_HEADS_REFSPEC = "refs/heads/*:refs/heads/*"
TARGET_TAGS_REFSPEC = "refs/tags/*:refs/tags/*"
RETRYABLE_HTTP_STATUS = {429, 502, 503, 504}
STATE_VERSION = 1


class MigrationError(RuntimeError):
    """An expected operational failure that should be shown without traceback."""


class ApiError(MigrationError):
    """A GitLab API request failed."""

    def __init__(self, method: str, url: str, status: int, message: str) -> None:
        super().__init__(f"{method} {url} failed with HTTP {status}: {message}")
        self.method = method
        self.url = url
        self.status = status
        self.message = message


class RejectRedirects(urllib.request.HTTPRedirectHandler):
    """Fail instead of forwarding an authentication header to a redirect."""

    def redirect_request(
        self,
        request: urllib.request.Request,
        file_pointer: Any,
        code: int,
        message: str,
        headers: Mapping[str, str],
        new_url: str,
    ) -> None:
        del request, file_pointer, code, message, headers, new_url
        return None


def normalize_base_url(value: str) -> str:
    value = value.rstrip("/")
    if value.endswith("/api/v4"):
        value = value[: -len("/api/v4")]
    parsed = urllib.parse.urlparse(value)
    if parsed.scheme not in {"http", "https"} or not parsed.netloc:
        raise MigrationError(f"invalid GitLab base URL: {value}")
    if parsed.username or parsed.password:
        raise MigrationError("do not put credentials in the GitLab base URL")
    if parsed.query or parsed.fragment:
        raise MigrationError(
            f"GitLab base URL cannot contain query or fragment: {value}"
        )
    return value


def encode_path(value: str | int) -> str:
    return urllib.parse.quote(str(value), safe="")


def relative_path(root: str, full_path: str) -> str:
    if full_path == root:
        return ""
    prefix = f"{root}/"
    if not full_path.startswith(prefix):
        raise MigrationError(f"{full_path!r} is not below source group {root!r}")
    return full_path[len(prefix) :]


def map_group_path(source_root: str, target_root: str, source_path: str) -> str:
    relative = relative_path(source_root, source_path)
    return target_root if not relative else f"{target_root}/{relative}"


def display_event(action: str, source: str, target: str = "", detail: str = "") -> None:
    mapping = f"{source} -> {target}" if target else source
    suffix = f" ({detail})" if detail else ""
    print(f"{action:<22} {mapping}{suffix}", flush=True)


class GitLabClient:
    """Small GitLab REST client using only the Python standard library."""

    def __init__(
        self,
        base_url: str,
        token: str,
        *,
        timeout: float = 30.0,
        retries: int = 3,
    ) -> None:
        self.base_url = normalize_base_url(base_url)
        self.api_url = f"{self.base_url}/api/v4"
        self.token = token
        self.timeout = timeout
        self.retries = retries
        self._open = urllib.request.build_opener(RejectRedirects()).open

    def _request(
        self,
        method: str,
        path: str,
        *,
        query: Mapping[str, str | int | bool] | None = None,
        payload: Mapping[str, Any] | None = None,
        allow_404: bool = False,
    ) -> tuple[Any | None, Mapping[str, str]]:
        url = f"{self.api_url}{path}"
        if query:
            encoded_query = urllib.parse.urlencode(
                {
                    key: str(value).lower() if isinstance(value, bool) else value
                    for key, value in query.items()
                }
            )
            url = f"{url}?{encoded_query}"

        body = None
        headers = {
            "Accept": "application/json",
            "PRIVATE-TOKEN": self.token,
            "User-Agent": "study-gitlab-group-migrator/1",
        }
        if payload is not None:
            body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
            headers["Content-Type"] = "application/json"

        for attempt in range(self.retries + 1):
            request = urllib.request.Request(
                url,
                data=body,
                headers=headers,
                method=method,
            )
            try:
                with self._open(request, timeout=self.timeout) as response:
                    raw = response.read()
                    try:
                        data = json.loads(raw.decode("utf-8")) if raw else None
                    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
                        raise MigrationError(
                            f"{method} {url} returned invalid JSON"
                        ) from exc
                    return data, dict(response.headers.items())
            except urllib.error.HTTPError as exc:
                try:
                    raw_error = exc.read().decode("utf-8", errors="replace")
                finally:
                    exc.close()
                if exc.code == 404 and allow_404:
                    return None, dict(exc.headers.items())
                can_retry = method in {"GET", "HEAD", "PUT"}
                if (
                    can_retry
                    and exc.code in RETRYABLE_HTTP_STATUS
                    and attempt < self.retries
                ):
                    retry_after = exc.headers.get("Retry-After", "")
                    delay = float(retry_after) if retry_after.isdigit() else 2**attempt
                    time.sleep(min(delay, 30.0))
                    continue
                message = str(raw_error.strip() or exc.reason)
                message = message.replace(self.token, "<redacted>")[:4000]
                raise ApiError(method, url, exc.code, message) from exc
            except urllib.error.URLError as exc:
                if method in {"GET", "HEAD", "PUT"} and attempt < self.retries:
                    time.sleep(2**attempt)
                    continue
                raise MigrationError(f"{method} {url} failed: {exc.reason}") from exc

        raise AssertionError("request retry loop terminated unexpectedly")

    def _paginated(
        self,
        path: str,
        *,
        query: Mapping[str, str | int | bool] | None = None,
    ) -> list[dict[str, Any]]:
        page = 1
        per_page = 100
        items: list[dict[str, Any]] = []
        base_query = dict(query or {})

        while True:
            current_query = {**base_query, "page": page, "per_page": per_page}
            data, headers = self._request("GET", path, query=current_query)
            if not isinstance(data, list):
                raise MigrationError(f"expected a list response from {path}")
            items.extend(data)

            next_page = next(
                (
                    value
                    for key, value in headers.items()
                    if key.lower() == "x-next-page"
                ),
                "",
            ).strip()
            if next_page:
                page = int(next_page)
                continue
            if len(data) < per_page:
                break
            page += 1

        return items

    def get_group(self, full_path: str) -> dict[str, Any] | None:
        data, _ = self._request(
            "GET",
            f"/groups/{encode_path(full_path)}",
            allow_404=True,
        )
        return data if isinstance(data, dict) else None

    def list_descendant_groups(self, group_id: int) -> list[dict[str, Any]]:
        return self._paginated(
            f"/groups/{group_id}/descendant_groups",
            query={"order_by": "id", "sort": "asc"},
        )

    def list_group_projects(self, group_id: int) -> list[dict[str, Any]]:
        return self._paginated(
            f"/groups/{group_id}/projects",
            query={
                "include_subgroups": True,
                "with_shared": False,
                "simple": False,
                "order_by": "id",
                "sort": "asc",
            },
        )

    def create_group(
        self,
        source_group: Mapping[str, Any],
        parent_id: int,
        *,
        copy_visibility: bool,
    ) -> dict[str, Any]:
        payload: dict[str, Any] = {
            "name": source_group["name"],
            "path": source_group["path"],
            "parent_id": parent_id,
        }
        description = source_group.get("description")
        if description:
            payload["description"] = description
        if copy_visibility and source_group.get("visibility"):
            payload["visibility"] = source_group["visibility"]

        data, _ = self._request("POST", "/groups", payload=payload)
        if not isinstance(data, dict):
            raise MigrationError("GitLab returned an invalid group creation response")
        return data

    def get_project(self, full_path: str) -> dict[str, Any] | None:
        data, _ = self._request(
            "GET",
            f"/projects/{encode_path(full_path)}",
            allow_404=True,
        )
        return data if isinstance(data, dict) else None

    def create_project(
        self,
        source_project: Mapping[str, Any],
        namespace_id: int,
        *,
        copy_visibility: bool,
    ) -> dict[str, Any]:
        payload: dict[str, Any] = {
            "name": source_project["name"],
            "path": source_project["path"],
            "namespace_id": namespace_id,
            "initialize_with_readme": False,
        }
        description = source_project.get("description")
        if description:
            payload["description"] = description
        if copy_visibility and source_project.get("visibility"):
            payload["visibility"] = source_project["visibility"]

        data, _ = self._request("POST", "/projects", payload=payload)
        if not isinstance(data, dict):
            raise MigrationError("GitLab returned an invalid project creation response")
        return data

    def set_default_branch(self, project_id: int, default_branch: str) -> None:
        self._request(
            "PUT",
            f"/projects/{project_id}",
            payload={"default_branch": default_branch},
        )


class StateStore:
    """Persistent ownership and resume state. Tokens are never stored."""

    def __init__(self, path: Path, context: Mapping[str, str]) -> None:
        self.path = path
        self.context = dict(context)
        self.data = self._load()

    def _load(self) -> dict[str, Any]:
        if not self.path.exists():
            return {
                "version": STATE_VERSION,
                "context": self.context,
                "projects": {},
            }

        try:
            data = json.loads(self.path.read_text(encoding="utf-8"))
        except (OSError, json.JSONDecodeError) as exc:
            raise MigrationError(f"cannot read state file {self.path}: {exc}") from exc

        if data.get("version") != STATE_VERSION:
            raise MigrationError(f"unsupported state version in {self.path}")
        if data.get("context") != self.context:
            raise MigrationError(
                f"state context mismatch in {self.path}; use a separate work directory"
            )
        if not isinstance(data.get("projects"), dict):
            raise MigrationError(f"invalid project state in {self.path}")
        return data

    def get_project(self, source_project_id: int) -> dict[str, Any] | None:
        value = self.data["projects"].get(str(source_project_id))
        return value if isinstance(value, dict) else None

    def update_project(self, source_project_id: int, **values: Any) -> None:
        key = str(source_project_id)
        current = dict(self.data["projects"].get(key, {}))
        current.update(values)
        self.data["projects"][key] = current
        self.save()

    def save(self) -> None:
        self.path.parent.mkdir(parents=True, exist_ok=True)
        temporary = self.path.with_name(f"{self.path.name}.tmp")
        temporary.write_text(
            json.dumps(self.data, ensure_ascii=False, indent=2, sort_keys=True) + "\n",
            encoding="utf-8",
        )
        temporary.replace(self.path)


class GitRunner:
    def __init__(self, git_executable: str = "git") -> None:
        resolved = shutil.which(git_executable)
        if not resolved:
            raise MigrationError(f"Git executable not found: {git_executable}")
        self.git_executable = resolved

    @staticmethod
    def _authentication_environment(
        token: str | None,
        auth_url: str | None,
    ) -> dict[str, str]:
        environment = os.environ.copy()
        environment["GIT_TERMINAL_PROMPT"] = "0"
        if token:
            if not auth_url:
                raise MigrationError(
                    "an authentication URL is required when a Git token is used"
                )
            credential = base64.b64encode(f"oauth2:{token}".encode("utf-8")).decode(
                "ascii"
            )
            index = int(environment.get("GIT_CONFIG_COUNT", "0"))
            environment["GIT_CONFIG_COUNT"] = str(index + 2)
            environment[f"GIT_CONFIG_KEY_{index}"] = (
                f"http.{auth_url}.extraHeader"
            )
            environment[f"GIT_CONFIG_VALUE_{index}"] = (
                f"Authorization: Basic {credential}"
            )
            environment[f"GIT_CONFIG_KEY_{index + 1}"] = "http.followRedirects"
            environment[f"GIT_CONFIG_VALUE_{index + 1}"] = "false"
        return environment

    def run(
        self,
        arguments: Sequence[str],
        *,
        cwd: Path | None = None,
        token: str | None = None,
        auth_url: str | None = None,
        capture: bool = True,
        check: bool = True,
    ) -> subprocess.CompletedProcess[str]:
        command = [self.git_executable, *arguments]
        result = subprocess.run(
            command,
            cwd=cwd,
            env=self._authentication_environment(token, auth_url),
            text=True,
            stdout=subprocess.PIPE if capture else None,
            stderr=subprocess.STDOUT if capture else None,
            check=False,
        )
        if check and result.returncode != 0:
            message = result.stdout.strip() if result.stdout else "see Git output above"
            raise MigrationError(
                f"Git command failed with exit code {result.returncode}: {message}"
            )
        return result


class GitRepositoryMigrator:
    def __init__(
        self,
        runner: GitRunner,
        work_root: Path,
        *,
        source_token: str,
        target_token: str,
        allow_insecure_http: bool,
        migrate_lfs: bool,
    ) -> None:
        self.runner = runner
        self.work_root = work_root
        self.source_token = source_token
        self.target_token = target_token
        self.allow_insecure_http = allow_insecure_http
        self.migrate_lfs = migrate_lfs

    def _validate_url(self, value: str) -> None:
        parsed = urllib.parse.urlparse(value)
        if parsed.scheme not in {"http", "https"} or not parsed.netloc:
            raise MigrationError(
                f"token-based migration requires an HTTP(S) repository URL: {value}"
            )
        if parsed.username or parsed.password:
            raise MigrationError(
                f"repository URL must not contain credentials: {value}"
            )
        if parsed.scheme == "http" and not self.allow_insecure_http:
            raise MigrationError(
                f"refusing to send a token over plain HTTP: {value}; "
                "use HTTPS or explicitly pass --allow-insecure-http"
            )

    def _repository_directory(self, target_path: str, source_project_id: int) -> Path:
        safe_name = re.sub(r"[^A-Za-z0-9._-]+", "_", target_path).strip("_")
        return self.work_root / "repositories" / f"{source_project_id}-{safe_name}.git"

    def _ensure_remote(self, repository: Path, name: str, url: str) -> None:
        existing = self.runner.run(
            ["remote", "get-url", name],
            cwd=repository,
            capture=True,
            check=False,
        )
        if existing.returncode == 0:
            self.runner.run(["remote", "set-url", name, url], cwd=repository)
        else:
            self.runner.run(["remote", "add", name, url], cwd=repository)

    @staticmethod
    def _parse_local_refs(output: str) -> dict[str, str]:
        refs: dict[str, str] = {}
        for line in output.splitlines():
            if not line.strip():
                continue
            object_id, ref_name = line.split(" ", 1)
            refs[ref_name] = object_id
        return refs

    @staticmethod
    def _parse_remote_refs(output: str) -> dict[str, str]:
        refs: dict[str, str] = {}
        for line in output.splitlines():
            if not line.strip():
                continue
            object_id, ref_name = line.split(None, 1)
            if ref_name.endswith("^{}"):
                continue
            refs[ref_name] = object_id
        return refs

    @staticmethod
    def _ref_mismatches(
        expected_refs: Mapping[str, str],
        actual_refs: Mapping[str, str],
    ) -> list[str]:
        mismatches = [
            f"{ref_name}: expected {expected_id}, actual {actual_refs.get(ref_name)}"
            for ref_name, expected_id in sorted(expected_refs.items())
            if actual_refs.get(ref_name) != expected_id
        ]
        mismatches.extend(
            f"{ref_name}: unexpected target ref {actual_id}"
            for ref_name, actual_id in sorted(actual_refs.items())
            if ref_name not in expected_refs
        )
        return mismatches

    def migrate(
        self,
        source_project: Mapping[str, Any],
        target_project: Mapping[str, Any],
    ) -> None:
        source_url = str(source_project["http_url_to_repo"])
        target_url = str(target_project["http_url_to_repo"])
        self._validate_url(source_url)
        self._validate_url(target_url)

        target_path = str(target_project["path_with_namespace"])
        repository = self._repository_directory(
            target_path,
            int(source_project["id"]),
        )
        repository.parent.mkdir(parents=True, exist_ok=True)

        if not repository.exists():
            self.runner.run(["init", "--bare", str(repository)])

        self._ensure_remote(repository, "source", source_url)
        self._ensure_remote(repository, "target", target_url)

        self.runner.run(
            [
                "fetch",
                "--prune",
                "--progress",
                "source",
                SOURCE_HEADS_REFSPEC,
                SOURCE_TAGS_REFSPEC,
            ],
            cwd=repository,
            token=self.source_token,
            auth_url=source_url,
            capture=False,
        )
        self.runner.run(["fsck", "--full"], cwd=repository)

        expected_output = self.runner.run(
            [
                "for-each-ref",
                "--format=%(objectname) %(refname)",
                "refs/heads",
                "refs/tags",
            ],
            cwd=repository,
        ).stdout
        expected_refs = self._parse_local_refs(expected_output)

        if not expected_refs:
            display_event(
                "EMPTY_SOURCE",
                str(source_project["path_with_namespace"]),
                target_path,
            )
            return

        if self.migrate_lfs:
            lfs_check = self.runner.run(
                ["lfs", "version"],
                cwd=repository,
                check=False,
            )
            if lfs_check.returncode != 0:
                raise MigrationError("--migrate-lfs requires git-lfs")
            self.runner.run(
                ["lfs", "fetch", "--all", "source"],
                cwd=repository,
                token=self.source_token,
                auth_url=source_url,
                capture=False,
            )
            self.runner.run(
                ["lfs", "push", "--all", "target"],
                cwd=repository,
                token=self.target_token,
                auth_url=target_url,
                capture=False,
            )

        push_refspecs = [TARGET_HEADS_REFSPEC, TARGET_TAGS_REFSPEC]
        self.runner.run(
            ["push", "--dry-run", "--porcelain", "target", *push_refspecs],
            cwd=repository,
            token=self.target_token,
            auth_url=target_url,
            capture=False,
        )
        self.runner.run(
            [
                "push",
                "--atomic",
                "--progress",
                "--porcelain",
                "target",
                *push_refspecs,
            ],
            cwd=repository,
            token=self.target_token,
            auth_url=target_url,
            capture=False,
        )

        actual_output = self.runner.run(
            ["ls-remote", "--heads", "--tags", "target"],
            cwd=repository,
            token=self.target_token,
            auth_url=target_url,
        ).stdout
        actual_refs = self._parse_remote_refs(actual_output)

        mismatches = self._ref_mismatches(expected_refs, actual_refs)
        if mismatches:
            raise MigrationError(
                "target ref verification failed:\n" + "\n".join(mismatches)
            )


@dataclass(frozen=True)
class ControllerOptions:
    apply: bool
    migrate_created: bool
    migrate_lfs: bool
    copy_visibility: bool
    project_pattern: re.Pattern[str] | None
    max_projects: int | None


class GroupMigrationController:
    def __init__(
        self,
        source: GitLabClient,
        target: GitLabClient,
        state: StateStore,
        migrator: GitRepositoryMigrator,
        *,
        source_group_path: str,
        target_group_path: str,
        options: ControllerOptions,
    ) -> None:
        self.source = source
        self.target = target
        self.state = state
        self.migrator = migrator
        self.source_group_path = source_group_path.strip("/")
        self.target_group_path = target_group_path.strip("/")
        self.options = options
        self.failures = 0

    def _provision_groups(
        self,
        source_root: Mapping[str, Any],
        target_root: Mapping[str, Any],
        descendants: Sequence[Mapping[str, Any]],
    ) -> dict[str, Mapping[str, Any] | None]:
        target_groups: dict[str, Mapping[str, Any] | None] = {
            self.target_group_path: target_root
        }
        sorted_groups = sorted(
            descendants,
            key=lambda group: (
                str(group["full_path"]).count("/"),
                str(group["full_path"]),
            ),
        )

        for source_group in sorted_groups:
            source_path = str(source_group["full_path"])
            target_path = map_group_path(
                self.source_group_path,
                self.target_group_path,
                source_path,
            )
            existing = self.target.get_group(target_path)
            if existing:
                display_event("SKIP_GROUP_EXISTS", source_path, target_path)
                target_groups[target_path] = existing
                continue

            if not self.options.apply:
                display_event("PLAN_CREATE_GROUP", source_path, target_path)
                target_groups[target_path] = None
                continue

            parent_path = target_path.rsplit("/", 1)[0]
            parent_group = target_groups.get(parent_path)
            if not parent_group:
                parent_group = self.target.get_group(parent_path)
            if not parent_group:
                raise MigrationError(
                    f"target parent group is missing: {parent_path}"
                )

            created = self.target.create_group(
                source_group,
                int(parent_group["id"]),
                copy_visibility=self.options.copy_visibility,
            )
            display_event("CREATE_GROUP", source_path, target_path)
            target_groups[target_path] = created

        return target_groups

    def _selected_projects(
        self,
        projects: Sequence[dict[str, Any]],
    ) -> list[dict[str, Any]]:
        selected = []
        for project in projects:
            source_path = str(project["path_with_namespace"])
            if self.options.project_pattern and not self.options.project_pattern.search(
                source_path
            ):
                continue
            selected.append(project)
            if (
                self.options.max_projects is not None
                and len(selected) >= self.options.max_projects
            ):
                break
        return selected

    def _required_descendants(
        self,
        descendants: Sequence[dict[str, Any]],
        projects: Sequence[dict[str, Any]],
    ) -> list[dict[str, Any]]:
        if (
            self.options.project_pattern is None
            and self.options.max_projects is None
        ):
            return list(descendants)

        required_paths: set[str] = set()
        for project in projects:
            namespace = str(project["namespace"]["full_path"])
            while namespace != self.source_group_path:
                required_paths.add(namespace)
                parent, separator, _ = namespace.rpartition("/")
                parent_is_below_root = parent.startswith(
                    f"{self.source_group_path}/"
                )
                if (
                    not separator
                    or (
                        parent != self.source_group_path
                        and not parent_is_below_root
                    )
                ):
                    raise MigrationError(
                        f"project namespace is outside source group: {namespace}"
                    )
                namespace = parent

        return [
            group
            for group in descendants
            if str(group["full_path"]) in required_paths
        ]

    def _migrate_owned_project(
        self,
        source_project: Mapping[str, Any],
        target_project: Mapping[str, Any],
    ) -> None:
        source_id = int(source_project["id"])
        source_path = str(source_project["path_with_namespace"])
        target_path = str(target_project["path_with_namespace"])
        try:
            display_event("MIGRATE_REFS", source_path, target_path)
            self.migrator.migrate(source_project, target_project)
            default_branch = source_project.get("default_branch")
            if default_branch:
                self.target.set_default_branch(
                    int(target_project["id"]),
                    str(default_branch),
                )
            self.state.update_project(
                source_id,
                status="verified",
                lfs_migrated=self.options.migrate_lfs,
                error=None,
            )
            display_event("VERIFY_PASS", source_path, target_path)
        except Exception as exc:
            self.state.update_project(
                source_id,
                status="migration_failed",
                error=str(exc),
            )
            display_event("MIGRATION_FAIL", source_path, target_path, str(exc))
            self.failures += 1

    def _provision_projects(
        self,
        source_root: Mapping[str, Any],
        target_groups: Mapping[str, Mapping[str, Any] | None],
        projects: Sequence[dict[str, Any]],
    ) -> None:
        del source_root
        for source_project in projects:
            source_id = int(source_project["id"])
            source_path = str(source_project["path_with_namespace"])
            source_namespace = str(source_project["namespace"]["full_path"])
            target_namespace_path = map_group_path(
                self.source_group_path,
                self.target_group_path,
                source_namespace,
            )
            target_path = f"{target_namespace_path}/{source_project['path']}"
            existing = self.target.get_project(target_path)
            record = self.state.get_project(source_id)

            if existing:
                if (
                    record
                    and record.get("status") == "creation_failed"
                    and record.get("target_path") == target_path
                ):
                    display_event(
                        "BLOCK_CREATE_AMBIGUOUS",
                        source_path,
                        target_path,
                        "project appeared after a failed create request",
                    )
                    self.failures += 1
                    continue

                state_owned = bool(
                    record
                    and record.get("created_by_tool") is True
                    and int(record.get("target_project_id", -1)) == int(existing["id"])
                    and record.get("target_path") == target_path
                )
                needs_lfs_migration = bool(
                    self.options.migrate_lfs
                    and record
                    and record.get("lfs_migrated") is not True
                )
                if (
                    state_owned
                    and self.options.apply
                    and self.options.migrate_created
                    and (
                        record.get("status") != "verified"
                        or needs_lfs_migration
                    )
                ):
                    action = (
                        "RESUME_LFS"
                        if record.get("status") == "verified"
                        else "RESUME_CREATED"
                    )
                    display_event(action, source_path, target_path)
                    self._migrate_owned_project(source_project, existing)
                elif state_owned and record.get("status") == "verified":
                    display_event("SKIP_VERIFIED", source_path, target_path)
                else:
                    display_event("SKIP_PROJECT_EXISTS", source_path, target_path)
                continue

            if not self.options.apply:
                display_event("PLAN_CREATE_PROJECT", source_path, target_path)
                continue

            target_namespace = target_groups.get(target_namespace_path)
            if not target_namespace:
                target_namespace = self.target.get_group(target_namespace_path)
            if not target_namespace:
                display_event(
                    "CREATE_PROJECT_FAIL",
                    source_path,
                    target_path,
                    f"target namespace missing: {target_namespace_path}",
                )
                self.failures += 1
                continue

            try:
                created = self.target.create_project(
                    source_project,
                    int(target_namespace["id"]),
                    copy_visibility=self.options.copy_visibility,
                )
                self.state.update_project(
                    source_id,
                    source_path=source_path,
                    target_path=target_path,
                    target_project_id=int(created["id"]),
                    created_by_tool=True,
                    status="created",
                    error=None,
                )
                display_event("CREATE_PROJECT", source_path, target_path)
                if self.options.migrate_created:
                    self._migrate_owned_project(source_project, created)
            except Exception as exc:
                self.state.update_project(
                    source_id,
                    source_path=source_path,
                    target_path=target_path,
                    created_by_tool=False,
                    status="creation_failed",
                    error=str(exc),
                )
                display_event(
                    "CREATE_PROJECT_FAIL",
                    source_path,
                    target_path,
                    str(exc),
                )
                self.failures += 1

    def run(self) -> int:
        source_root = self.source.get_group(self.source_group_path)
        if not source_root:
            raise MigrationError(
                f"source group does not exist or is not visible: "
                f"{self.source_group_path}"
            )
        target_root = self.target.get_group(self.target_group_path)
        if not target_root:
            raise MigrationError(
                f"target root group must already exist: {self.target_group_path}"
            )

        descendants = self.source.list_descendant_groups(int(source_root["id"]))
        projects = self.source.list_group_projects(int(source_root["id"]))
        selected_projects = self._selected_projects(projects)
        if self.options.project_pattern and not selected_projects:
            raise MigrationError(
                "no source projects match the --project-regex pattern"
            )
        selected_descendants = self._required_descendants(
            descendants,
            selected_projects,
        )
        display_event(
            "DISCOVER",
            self.source_group_path,
            self.target_group_path,
            (
                f"{len(selected_descendants)}/{len(descendants)} subgroups, "
                f"{len(selected_projects)}/{len(projects)} projects selected"
            ),
        )

        target_groups = self._provision_groups(
            source_root,
            target_root,
            selected_descendants,
        )
        self._provision_projects(
            source_root,
            target_groups,
            selected_projects,
        )
        return 1 if self.failures else 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description=(
            "Create missing GitLab subgroups/projects and optionally migrate "
            "branches and tags for projects created by this tool."
        )
    )
    parser.add_argument("--source-url", required=True, help="source GitLab base URL")
    parser.add_argument("--target-url", required=True, help="target GitLab base URL")
    parser.add_argument("--source-group", required=True, help="source group full path")
    parser.add_argument(
        "--target-group",
        required=True,
        help="existing target group path",
    )
    parser.add_argument(
        "--source-token-env",
        default="SOURCE_GITLAB_TOKEN",
        help="environment variable containing source token",
    )
    parser.add_argument(
        "--target-token-env",
        default="TARGET_GITLAB_TOKEN",
        help="environment variable containing target token",
    )
    parser.add_argument(
        "--work-dir",
        type=Path,
        default=Path("gitlab-group-migration-work"),
        help="persistent state and bare repository directory",
    )
    parser.add_argument(
        "--apply",
        action="store_true",
        help="create missing subgroups and projects; default is plan-only",
    )
    parser.add_argument(
        "--migrate-created",
        action="store_true",
        help=(
            "migrate branches/tags only for projects created by this tool; "
            "requires --apply"
        ),
    )
    parser.add_argument(
        "--migrate-lfs",
        action="store_true",
        help="also fetch and push all Git LFS objects; requires --migrate-created",
    )
    parser.add_argument(
        "--copy-visibility",
        action="store_true",
        help="copy source group/project visibility instead of target defaults",
    )
    parser.add_argument(
        "--allow-insecure-http",
        action="store_true",
        help="allow tokens to be used over plain HTTP (not recommended)",
    )
    parser.add_argument(
        "--project-regex",
        help="only process source project full paths matching this regex",
    )
    parser.add_argument(
        "--max-projects",
        type=int,
        help="process at most this many selected projects",
    )
    parser.add_argument(
        "--request-timeout",
        type=float,
        default=30.0,
        help="GitLab API request timeout in seconds",
    )
    return parser


def main(argv: Sequence[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)

    if args.migrate_created and not args.apply:
        parser.error("--migrate-created requires --apply")
    if args.migrate_lfs and not args.migrate_created:
        parser.error("--migrate-lfs requires --migrate-created")
    if args.max_projects is not None and args.max_projects < 1:
        parser.error("--max-projects must be at least 1")

    source_token = os.environ.get(args.source_token_env, "")
    target_token = os.environ.get(args.target_token_env, "")
    if not source_token:
        parser.error(f"missing token environment variable: {args.source_token_env}")
    if not target_token:
        parser.error(f"missing token environment variable: {args.target_token_env}")

    try:
        source_url = normalize_base_url(args.source_url)
        target_url = normalize_base_url(args.target_url)
    except MigrationError as exc:
        parser.error(str(exc))
    source_group = args.source_group.strip("/")
    target_group = args.target_group.strip("/")
    if not source_group:
        parser.error("--source-group cannot be empty")
    if not target_group:
        parser.error("--target-group cannot be empty")
    if source_url == target_url and source_group == target_group:
        parser.error("source and target identify the same GitLab group")
    for label, base_url in (("source", source_url), ("target", target_url)):
        if (
            urllib.parse.urlparse(base_url).scheme == "http"
            and not args.allow_insecure_http
        ):
            parser.error(
                f"{label} URL uses plain HTTP; use HTTPS or explicitly pass "
                "--allow-insecure-http"
            )

    project_pattern = None
    if args.project_regex:
        try:
            project_pattern = re.compile(args.project_regex)
        except re.error as exc:
            parser.error(f"invalid --project-regex: {exc}")

    work_dir = args.work_dir.resolve()
    context = {
        "source_url": source_url,
        "source_group": source_group,
        "target_url": target_url,
        "target_group": target_group,
    }

    try:
        source_client = GitLabClient(
            source_url,
            source_token,
            timeout=args.request_timeout,
        )
        target_client = GitLabClient(
            target_url,
            target_token,
            timeout=args.request_timeout,
        )
        state = StateStore(work_dir / "state.json", context)
        runner = GitRunner()
        migrator = GitRepositoryMigrator(
            runner,
            work_dir,
            source_token=source_token,
            target_token=target_token,
            allow_insecure_http=args.allow_insecure_http,
            migrate_lfs=args.migrate_lfs,
        )
        controller = GroupMigrationController(
            source_client,
            target_client,
            state,
            migrator,
            source_group_path=source_group,
            target_group_path=target_group,
            options=ControllerOptions(
                apply=args.apply,
                migrate_created=args.migrate_created,
                migrate_lfs=args.migrate_lfs,
                copy_visibility=args.copy_visibility,
                project_pattern=project_pattern,
                max_projects=args.max_projects,
            ),
        )
        return controller.run()
    except KeyboardInterrupt:
        print("interrupted", file=sys.stderr)
        return 130
    except MigrationError as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
