from __future__ import annotations

import importlib.util
import io
import sys
import tempfile
import unittest
import urllib.parse
from pathlib import Path
from typing import Any, Mapping, Sequence
from unittest import mock


MODULE_PATH = Path(__file__).with_name("migrate_gitlab_group.py")
SPEC = importlib.util.spec_from_file_location("migrate_gitlab_group", MODULE_PATH)
assert SPEC and SPEC.loader
MODULE = importlib.util.module_from_spec(SPEC)
sys.modules[SPEC.name] = MODULE
SPEC.loader.exec_module(MODULE)

ControllerOptions = MODULE.ControllerOptions
GitRepositoryMigrator = MODULE.GitRepositoryMigrator
GitRunner = MODULE.GitRunner
GitLabClient = MODULE.GitLabClient
GroupMigrationController = MODULE.GroupMigrationController
StateStore = MODULE.StateStore
map_group_path = MODULE.map_group_path
normalize_base_url = MODULE.normalize_base_url


def source_fixture() -> tuple[
    dict[str, Any],
    list[dict[str, Any]],
    list[dict[str, Any]],
]:
    root = {
        "id": 1,
        "name": "KCF",
        "path": "kcf",
        "full_path": "kcf",
        "visibility": "private",
    }
    groups = [
        {
            "id": 2,
            "name": "Payments",
            "path": "payments",
            "full_path": "kcf/payments",
            "parent_id": 1,
            "visibility": "private",
        }
    ]
    projects = [
        {
            "id": 10,
            "name": "Existing",
            "path": "existing",
            "path_with_namespace": "kcf/existing",
            "namespace": {"full_path": "kcf"},
            "description": "",
            "visibility": "private",
            "default_branch": "main",
            "http_url_to_repo": "https://source.example/kcf/existing.git",
        },
        {
            "id": 11,
            "name": "Payment API",
            "path": "payment-api",
            "path_with_namespace": "kcf/payments/payment-api",
            "namespace": {"full_path": "kcf/payments"},
            "description": "payment service",
            "visibility": "private",
            "default_branch": "main",
            "http_url_to_repo": (
                "https://source.example/kcf/payments/payment-api.git"
            ),
        },
    ]
    return root, groups, projects


class FakeSourceClient:
    def __init__(self) -> None:
        self.root, self.groups, self.projects = source_fixture()

    def get_group(self, full_path: str) -> dict[str, Any] | None:
        return self.root if full_path == self.root["full_path"] else None

    def list_descendant_groups(self, group_id: int) -> list[dict[str, Any]]:
        assert group_id == self.root["id"]
        return list(self.groups)

    def list_group_projects(self, group_id: int) -> list[dict[str, Any]]:
        assert group_id == self.root["id"]
        return list(self.projects)


class FakeTargetClient:
    def __init__(self) -> None:
        root = {
            "id": 100,
            "name": "KCF",
            "path": "kcf",
            "full_path": "axlab/division-dev/kcf",
            "visibility": "private",
        }
        existing = {
            "id": 200,
            "name": "Existing",
            "path": "existing",
            "path_with_namespace": "axlab/division-dev/kcf/existing",
            "namespace": {"full_path": root["full_path"]},
            "http_url_to_repo": (
                "https://target.example/axlab/division-dev/kcf/existing.git"
            ),
        }
        self.groups = {root["full_path"]: root}
        self.projects = {existing["path_with_namespace"]: existing}
        self.created_groups: list[dict[str, Any]] = []
        self.created_projects: list[dict[str, Any]] = []
        self.default_branch_updates: list[tuple[int, str]] = []
        self.next_group_id = 101
        self.next_project_id = 201

    def get_group(self, full_path: str) -> dict[str, Any] | None:
        return self.groups.get(full_path)

    def create_group(
        self,
        source_group: Mapping[str, Any],
        parent_id: int,
        *,
        copy_visibility: bool,
    ) -> dict[str, Any]:
        parent = next(
            group
            for group in self.groups.values()
            if group["id"] == parent_id
        )
        full_path = f"{parent['full_path']}/{source_group['path']}"
        created = {
            "id": self.next_group_id,
            "name": source_group["name"],
            "path": source_group["path"],
            "full_path": full_path,
            "parent_id": parent_id,
            "visibility": (
                source_group["visibility"] if copy_visibility else parent["visibility"]
            ),
        }
        self.next_group_id += 1
        self.groups[full_path] = created
        self.created_groups.append(created)
        return created

    def get_project(self, full_path: str) -> dict[str, Any] | None:
        return self.projects.get(full_path)

    def create_project(
        self,
        source_project: Mapping[str, Any],
        namespace_id: int,
        *,
        copy_visibility: bool,
    ) -> dict[str, Any]:
        namespace = next(
            group for group in self.groups.values() if group["id"] == namespace_id
        )
        full_path = f"{namespace['full_path']}/{source_project['path']}"
        created = {
            "id": self.next_project_id,
            "name": source_project["name"],
            "path": source_project["path"],
            "path_with_namespace": full_path,
            "namespace": {"full_path": namespace["full_path"]},
            "visibility": (
                source_project["visibility"]
                if copy_visibility
                else namespace["visibility"]
            ),
            "http_url_to_repo": f"https://target.example/{full_path}.git",
        }
        self.next_project_id += 1
        self.projects[full_path] = created
        self.created_projects.append(created)
        return created

    def set_default_branch(self, project_id: int, default_branch: str) -> None:
        self.default_branch_updates.append((project_id, default_branch))


class FakeMigrator:
    def __init__(self, *, fail: bool = False) -> None:
        self.fail = fail
        self.calls: list[tuple[str, str]] = []

    def migrate(
        self,
        source_project: Mapping[str, Any],
        target_project: Mapping[str, Any],
    ) -> None:
        self.calls.append(
            (
                str(source_project["path_with_namespace"]),
                str(target_project["path_with_namespace"]),
            )
        )
        if self.fail:
            raise MODULE.MigrationError("simulated migration failure")


class AmbiguousCreateTarget(FakeTargetClient):
    def create_project(
        self,
        source_project: Mapping[str, Any],
        namespace_id: int,
        *,
        copy_visibility: bool,
    ) -> dict[str, Any]:
        super().create_project(
            source_project,
            namespace_id,
            copy_visibility=copy_visibility,
        )
        raise MODULE.MigrationError("response lost after project creation")


def options(
    *,
    apply: bool,
    migrate_created: bool,
    migrate_lfs: bool = False,
    project_pattern: Any = None,
    max_projects: int | None = None,
) -> Any:
    return ControllerOptions(
        apply=apply,
        migrate_created=migrate_created,
        migrate_lfs=migrate_lfs,
        copy_visibility=False,
        project_pattern=project_pattern,
        max_projects=max_projects,
    )


class MappingTests(unittest.TestCase):
    def test_group_path_mapping_preserves_relative_path(self) -> None:
        self.assertEqual(
            map_group_path(
                "kcf",
                "axlab/division-dev/kcf",
                "kcf/payments/card",
            ),
            "axlab/division-dev/kcf/payments/card",
        )

    def test_base_url_rejects_embedded_credentials(self) -> None:
        with self.assertRaises(MODULE.MigrationError):
            normalize_base_url("https://token@gitlab.example")


class FakeHttpResponse:
    def __init__(
        self,
        payload: list[dict[str, Any]],
        headers: Mapping[str, str],
    ) -> None:
        self.payload = MODULE.json.dumps(payload).encode("utf-8")
        self.headers = headers

    def __enter__(self) -> "FakeHttpResponse":
        return self

    def __exit__(self, *args: Any) -> None:
        return None

    def read(self) -> bytes:
        return self.payload


class GitLabClientTests(unittest.TestCase):
    def test_pagination_sends_private_token_without_putting_it_in_url(self) -> None:
        responses = [
            FakeHttpResponse([{"id": 1}], {"X-Next-Page": "2"}),
            FakeHttpResponse([{"id": 2}], {"X-Next-Page": ""}),
        ]
        requests: list[Any] = []

        def fake_urlopen(request: Any, *, timeout: float) -> FakeHttpResponse:
            self.assertEqual(timeout, 7.0)
            requests.append(request)
            return responses.pop(0)

        client = GitLabClient(
            "https://source.example",
            "source-secret",
            timeout=7.0,
        )
        with mock.patch.object(client, "_open", fake_urlopen):
            projects = client._paginated("/groups/1/projects")

        self.assertEqual(projects, [{"id": 1}, {"id": 2}])
        self.assertEqual(len(requests), 2)
        self.assertEqual(
            [
                urllib.parse.parse_qs(
                    urllib.parse.urlparse(item.full_url).query
                )["page"]
                for item in requests
            ],
            [["1"], ["2"]],
        )
        for request in requests:
            self.assertEqual(request.get_header("Private-token"), "source-secret")
            self.assertNotIn("source-secret", request.full_url)

    def test_failed_post_is_not_retried_and_token_is_redacted(self) -> None:
        client = GitLabClient(
            "https://target.example",
            "target-secret",
            retries=3,
        )
        calls = 0

        def fail_post(request: Any, *, timeout: float) -> Any:
            del timeout
            nonlocal calls
            calls += 1
            raise MODULE.urllib.error.HTTPError(
                request.full_url,
                503,
                "Service Unavailable",
                {},
                io.BytesIO(b"failure for target-secret"),
            )

        with mock.patch.object(client, "_open", fail_post):
            with self.assertRaises(MODULE.ApiError) as raised:
                client._request(
                    "POST",
                    "/projects",
                    payload={"name": "project"},
                )

        self.assertEqual(calls, 1)
        self.assertNotIn("target-secret", str(raised.exception))
        self.assertIn("<redacted>", str(raised.exception))


class GitRunnerTests(unittest.TestCase):
    def test_authentication_header_is_scoped_to_repository_url(self) -> None:
        repository_url = "https://gitlab.example/group/project.git"
        environment = GitRunner._authentication_environment(
            "secret",
            repository_url,
        )

        count = int(environment["GIT_CONFIG_COUNT"])
        keys = [environment[f"GIT_CONFIG_KEY_{index}"] for index in range(count)]
        self.assertIn(f"http.{repository_url}.extraHeader", keys)
        self.assertNotIn("http.extraHeader", keys)
        self.assertIn("http.followRedirects", keys)
        self.assertNotIn("secret", " ".join(environment.values()))

        runner = GitRunner()
        redirect_setting = runner.run(
            [
                "config",
                "--get-urlmatch",
                "http.followRedirects",
                repository_url,
            ],
            token="secret",
            auth_url=repository_url,
        )
        self.assertEqual(redirect_setting.stdout.strip(), "false")

        unrelated_header = runner.run(
            [
                "config",
                "--get-urlmatch",
                "http.extraHeader",
                "https://other.example/project.git",
            ],
            token="secret",
            auth_url=repository_url,
            check=False,
        )
        self.assertNotEqual(unrelated_header.returncode, 0)
        self.assertEqual(unrelated_header.stdout, "")


class ProvisioningTests(unittest.TestCase):
    def make_state(self, temporary: str) -> StateStore:
        return StateStore(
            Path(temporary) / "state.json",
            {
                "source_url": "https://source.example",
                "source_group": "kcf",
                "target_url": "https://target.example",
                "target_group": "axlab/division-dev/kcf",
            },
        )

    def make_controller(
        self,
        temporary: str,
        target: FakeTargetClient,
        migrator: FakeMigrator,
        *,
        apply: bool,
        migrate_created: bool,
        migrate_lfs: bool = False,
        project_pattern: Any = None,
        max_projects: int | None = None,
    ) -> GroupMigrationController:
        return GroupMigrationController(
            FakeSourceClient(),
            target,
            self.make_state(temporary),
            migrator,
            source_group_path="kcf",
            target_group_path="axlab/division-dev/kcf",
            options=options(
                apply=apply,
                migrate_created=migrate_created,
                migrate_lfs=migrate_lfs,
                project_pattern=project_pattern,
                max_projects=max_projects,
            ),
        )

    def test_plan_does_not_create_or_migrate(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            target = FakeTargetClient()
            migrator = FakeMigrator()
            controller = self.make_controller(
                temporary,
                target,
                migrator,
                apply=False,
                migrate_created=False,
            )

            self.assertEqual(controller.run(), 0)
            self.assertEqual(target.created_groups, [])
            self.assertEqual(target.created_projects, [])
            self.assertEqual(migrator.calls, [])

    def test_project_filter_does_not_create_unrelated_subgroups(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            target = FakeTargetClient()
            migrator = FakeMigrator()
            controller = self.make_controller(
                temporary,
                target,
                migrator,
                apply=True,
                migrate_created=True,
                project_pattern=MODULE.re.compile(r"^kcf/existing$"),
            )

            self.assertEqual(controller.run(), 0)
            self.assertEqual(target.created_groups, [])
            self.assertEqual(target.created_projects, [])
            self.assertEqual(migrator.calls, [])

    def test_nested_project_filter_creates_its_required_subgroup(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            target = FakeTargetClient()
            migrator = FakeMigrator()
            controller = self.make_controller(
                temporary,
                target,
                migrator,
                apply=True,
                migrate_created=False,
                project_pattern=MODULE.re.compile(
                    r"^kcf/payments/payment-api$"
                ),
            )

            self.assertEqual(controller.run(), 0)
            self.assertEqual(
                [group["full_path"] for group in target.created_groups],
                ["axlab/division-dev/kcf/payments"],
            )
            self.assertEqual(
                [
                    project["path_with_namespace"]
                    for project in target.created_projects
                ],
                ["axlab/division-dev/kcf/payments/payment-api"],
            )
            self.assertEqual(migrator.calls, [])

    def test_apply_skips_existing_and_migrates_only_created_project(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            target = FakeTargetClient()
            migrator = FakeMigrator()
            controller = self.make_controller(
                temporary,
                target,
                migrator,
                apply=True,
                migrate_created=True,
            )

            self.assertEqual(controller.run(), 0)
            self.assertEqual(
                [group["full_path"] for group in target.created_groups],
                ["axlab/division-dev/kcf/payments"],
            )
            self.assertEqual(
                [project["path_with_namespace"] for project in target.created_projects],
                ["axlab/division-dev/kcf/payments/payment-api"],
            )
            self.assertEqual(
                migrator.calls,
                [
                    (
                        "kcf/payments/payment-api",
                        "axlab/division-dev/kcf/payments/payment-api",
                    )
                ],
            )
            self.assertEqual(target.default_branch_updates, [(201, "main")])

            state = self.make_state(temporary)
            self.assertIsNone(state.get_project(10))
            self.assertEqual(state.get_project(11)["status"], "verified")

    def test_provision_first_then_migrate_created_project(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            target = FakeTargetClient()
            provision = self.make_controller(
                temporary,
                target,
                FakeMigrator(),
                apply=True,
                migrate_created=False,
            )

            self.assertEqual(provision.run(), 0)
            self.assertEqual(
                self.make_state(temporary).get_project(11)["status"],
                "created",
            )

            migrator = FakeMigrator()
            migrate = self.make_controller(
                temporary,
                target,
                migrator,
                apply=True,
                migrate_created=True,
            )

            self.assertEqual(migrate.run(), 0)
            self.assertEqual(
                migrator.calls,
                [
                    (
                        "kcf/payments/payment-api",
                        "axlab/division-dev/kcf/payments/payment-api",
                    )
                ],
            )
            self.assertEqual(
                self.make_state(temporary).get_project(11)["status"],
                "verified",
            )

    def test_lfs_mode_revisits_tool_owned_project_verified_without_lfs(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            target = FakeTargetClient()
            first = self.make_controller(
                temporary,
                target,
                FakeMigrator(),
                apply=True,
                migrate_created=True,
                migrate_lfs=False,
            )
            self.assertEqual(first.run(), 0)
            self.assertFalse(
                self.make_state(temporary).get_project(11)["lfs_migrated"]
            )

            lfs_migrator = FakeMigrator()
            second = self.make_controller(
                temporary,
                target,
                lfs_migrator,
                apply=True,
                migrate_created=True,
                migrate_lfs=True,
            )
            self.assertEqual(second.run(), 0)
            self.assertEqual(len(lfs_migrator.calls), 1)
            self.assertTrue(
                self.make_state(temporary).get_project(11)["lfs_migrated"]
            )

    def test_failed_tool_created_project_is_resumed(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            target = FakeTargetClient()
            failing = FakeMigrator(fail=True)
            first = self.make_controller(
                temporary,
                target,
                failing,
                apply=True,
                migrate_created=True,
            )

            self.assertEqual(first.run(), 1)
            self.assertEqual(
                self.make_state(temporary).get_project(11)["status"],
                "migration_failed",
            )

            succeeding = FakeMigrator()
            second = self.make_controller(
                temporary,
                target,
                succeeding,
                apply=True,
                migrate_created=True,
            )

            self.assertEqual(second.run(), 0)
            self.assertEqual(len(target.created_projects), 1)
            self.assertEqual(len(succeeding.calls), 1)
            self.assertEqual(
                self.make_state(temporary).get_project(11)["status"],
                "verified",
            )

    def test_ambiguous_create_failure_remains_blocked_on_rerun(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            target = AmbiguousCreateTarget()
            migrator = FakeMigrator()
            first = self.make_controller(
                temporary,
                target,
                migrator,
                apply=True,
                migrate_created=True,
            )

            self.assertEqual(first.run(), 1)
            self.assertEqual(
                self.make_state(temporary).get_project(11)["status"],
                "creation_failed",
            )

            second = self.make_controller(
                temporary,
                target,
                migrator,
                apply=True,
                migrate_created=True,
            )
            self.assertEqual(second.run(), 1)
            self.assertEqual(migrator.calls, [])


class LocalGitMigrator(GitRepositoryMigrator):
    def _validate_url(self, value: str) -> None:
        del value


class GitMigrationIntegrationTests(unittest.TestCase):
    def run_git(
        self,
        runner: GitRunner,
        arguments: Sequence[str],
        cwd: Path | None = None,
    ) -> str:
        return runner.run(arguments, cwd=cwd).stdout

    def test_ref_comparison_rejects_missing_changed_and_extra_refs(self) -> None:
        mismatches = GitRepositoryMigrator._ref_mismatches(
            {
                "refs/heads/main": "111",
                "refs/tags/v1": "222",
            },
            {
                "refs/heads/main": "999",
                "refs/heads/target-only": "333",
            },
        )

        self.assertEqual(len(mismatches), 3)
        self.assertTrue(
            any("refs/heads/main" in mismatch for mismatch in mismatches)
        )
        self.assertTrue(any("refs/tags/v1" in mismatch for mismatch in mismatches))
        self.assertTrue(
            any("refs/heads/target-only" in mismatch for mismatch in mismatches)
        )

    def test_branches_and_tags_are_migrated_to_empty_bare_repository(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            runner = GitRunner()
            source_bare = root / "source.git"
            target_bare = root / "target.git"
            working = root / "working"

            self.run_git(runner, ["init", "--bare", str(source_bare)])
            self.run_git(runner, ["init", "--bare", str(target_bare)])
            self.run_git(runner, ["init", "-b", "main", str(working)])
            self.run_git(runner, ["config", "user.name", "Migration Test"], working)
            self.run_git(
                runner,
                ["config", "user.email", "migration-test@example.invalid"],
                working,
            )
            (working / "README.md").write_text("migration test\n", encoding="utf-8")
            self.run_git(runner, ["add", "README.md"], working)
            self.run_git(runner, ["commit", "-m", "initial"], working)
            self.run_git(runner, ["branch", "feature/test"], working)
            self.run_git(runner, ["tag", "v1.0.0"], working)
            self.run_git(
                runner,
                ["push", str(source_bare), "--all"],
                working,
            )
            self.run_git(
                runner,
                ["push", str(source_bare), "--tags"],
                working,
            )

            migrator = LocalGitMigrator(
                runner,
                root / "migration",
                source_token="",
                target_token="",
                allow_insecure_http=False,
                migrate_lfs=False,
            )
            migrator.migrate(
                {
                    "id": 1,
                    "path_with_namespace": "kcf/project",
                    "http_url_to_repo": source_bare.as_uri(),
                },
                {
                    "id": 2,
                    "path_with_namespace": "target/kcf/project",
                    "http_url_to_repo": target_bare.as_uri(),
                },
            )

            source_refs = self.run_git(
                runner,
                [
                    "--git-dir",
                    str(source_bare),
                    "for-each-ref",
                    "--format=%(objectname) %(refname)",
                    "refs/heads",
                    "refs/tags",
                ],
            )
            target_refs = self.run_git(
                runner,
                [
                    "--git-dir",
                    str(target_bare),
                    "for-each-ref",
                    "--format=%(objectname) %(refname)",
                    "refs/heads",
                    "refs/tags",
                ],
            )
            self.assertEqual(
                sorted(source_refs.splitlines()),
                sorted(target_refs.splitlines()),
            )


if __name__ == "__main__":
    unittest.main()
