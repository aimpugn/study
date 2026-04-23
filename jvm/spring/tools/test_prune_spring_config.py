#!/usr/bin/env python3
from __future__ import annotations

import io
import sys
import tempfile
import unittest
from contextlib import redirect_stdout
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import prune_spring_config as prune


class PruneSpringConfigTest(unittest.TestCase):
    def run_tool(self, root: Path, *args: str) -> str:
        out = io.StringIO()
        with redirect_stdout(out):
            rc = prune.main([str(root), *args])
        self.assertEqual(rc, 0)
        return out.getvalue()

    def test_dry_run_removes_profile_suffix_without_reading_file(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            bad = root / "kcf-outbound-gateway-dev.properties"
            bad.write_bytes(b"\xff\xfe\x00")

            output = self.run_tool(root)

            self.assertIn("REMOVE kcf-outbound-gateway-dev.properties", output)
            self.assertIn("skipped-read", output)
            self.assertTrue(bad.exists())

    def test_write_removes_application_profile_outside_keep_set(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            dev = root / "application-dev.properties"
            local = root / "application-local.properties"
            dev.write_text("secret.password=company\n", encoding="utf-8")
            local.write_text("server.port=8080\n", encoding="utf-8")

            output = self.run_tool(root, "--write")

            self.assertIn("REMOVE application-dev.properties", output)
            self.assertFalse(dev.exists())
            self.assertTrue(local.exists())

    def test_properties_redacts_examples_and_removes_profile_document(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            app = root / "application.properties"
            app.write_text(
                "\n".join(
                    [
                        "kcf.kog.url=http://10.1.2.3:8980",
                        "svc.kmsg.kmsg-sndr-sys-apik=abcdef123456",
                        "#---",
                        "spring.config.activate.on-profile=dev",
                        "svc.kpnFcrm.key=1234567890ABCDEfGHIjklmn1234ABCD",
                        "",
                    ]
                ),
                encoding="utf-8",
            )

            output = self.run_tool(root, "--write")
            text = app.read_text(encoding="utf-8")

            self.assertIn("profile-docs-removed=1", output)
            self.assertIn("redactions=2", output)
            self.assertIn("http://127.0.0.1:8980", text)
            self.assertIn("<REDACTED_SECRET>", text)
            self.assertNotIn("10.1.2.3", text)
            self.assertNotIn("abcdef123456", text)
            self.assertNotIn("1234567890ABCDEfGHIjklmn1234ABCD", text)

    def test_yaml_redacts_examples_and_removes_profile_document(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            app = root / "application.yml"
            app.write_text(
                "\n".join(
                    [
                        "spring:",
                        "  datasource:",
                        "    username: username",
                        "    password: passwd01",
                        "    jdbc-url: jdbc:log4jdbc:mysql://10.2.3.4:3306/kcfinfo",
                        "svc:",
                        "  url: http://10.3.4.5:6102",
                        "  server-ip-1: 192.123.45.12",
                        "  kfs-fp2-token-value: Abc1E116ABCDEFGHIJKLMNFF62ABCB2E5025ABC8xxxxxxxxxFABCE58B99789456123",
                        "  fep-sftp-ip: 10.23.45.67",
                        "  kmsg-sndr-sys-apik: abcdefghi1jk",
                        "---",
                        "spring:",
                        "  config:",
                        "    activate:",
                        "      on-profile: real",
                        "unsafe:",
                        "  password: real-secret",
                        "",
                    ]
                ),
                encoding="utf-8",
            )

            output = self.run_tool(root, "--write")
            text = app.read_text(encoding="utf-8")

            self.assertIn("profile-docs-removed=1", output)
            self.assertIn("<REDACTED_USERNAME>", text)
            self.assertIn("<REDACTED_PASSWORD>", text)
            self.assertIn("<REDACTED_SECRET>", text)
            self.assertIn("jdbc:log4jdbc:mysql://127.0.0.1:3306/kcfinfo", text)
            self.assertIn("http://127.0.0.1:6102", text)
            self.assertNotIn("10.2.3.4", text)
            self.assertNotIn("10.3.4.5", text)
            self.assertNotIn("192.123.45.12", text)
            self.assertNotIn("real-secret", text)

    def test_kept_profile_document_survives(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            app = root / "application.yaml"
            app.write_text(
                "\n".join(
                    [
                        "---",
                        "spring:",
                        "  config:",
                        "    activate:",
                        "      on-profile: local",
                        "server:",
                        "  address: 127.0.0.1",
                        "",
                    ]
                ),
                encoding="utf-8",
            )

            self.run_tool(root, "--write")

            text = app.read_text(encoding="utf-8")
            self.assertIn("on-profile: local", text)
            self.assertIn("127.0.0.1", text)

    def test_default_scan_skips_build_outputs(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            target = root / "target" / "classes"
            target.mkdir(parents=True)
            generated = target / "application-dev.properties"
            generated.write_text("password=secret\n", encoding="utf-8")

            output = self.run_tool(root, "--write")

            self.assertIn("summary remove=0 update=0 keep=0", output)
            self.assertTrue(generated.exists())

    def test_application_word_is_not_treated_as_api_secret(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            app = root / "application.properties"
            app.write_text("spring.application.name=demo-service\n", encoding="utf-8")

            output = self.run_tool(root, "--write")

            self.assertIn("KEEP   application.properties", output)
            self.assertEqual("spring.application.name=demo-service\n", app.read_text(encoding="utf-8"))

    def test_secret_spring_placeholder_keeps_env_name_and_redacts_default(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            app = root / "application.yml"
            app.write_text(
                "notification:\n"
                "  slack-token: ${SVC_NOTIFICATION_GATEWAY_SLACK_TOKEN:xoxb-49755-secret}\n"
                "  api-token: ${SVC_NOTIFICATION_GATEWAY_API_TOKEN}\n",
                encoding="utf-8",
            )

            output = self.run_tool(root, "--write")
            text = app.read_text(encoding="utf-8")

            self.assertIn("redactions=1", output)
            self.assertIn("slack-token: ${SVC_NOTIFICATION_GATEWAY_SLACK_TOKEN:<REDACTED_SECRET>}", text)
            self.assertIn("api-token: ${SVC_NOTIFICATION_GATEWAY_API_TOKEN}", text)
            self.assertNotIn("xoxb-49755-secret", text)

    def test_properties_secret_spring_placeholder_keeps_env_name_and_redacts_default(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            app = root / "application.properties"
            app.write_text(
                "svc.notification.gateway.slack-token=${SVC_NOTIFICATION_GATEWAY_SLACK_TOKEN:xoxb-49755-secret}\n",
                encoding="utf-8",
            )

            self.run_tool(root, "--write")

            self.assertEqual(
                "svc.notification.gateway.slack-token=${SVC_NOTIFICATION_GATEWAY_SLACK_TOKEN:<REDACTED_SECRET>}\n",
                app.read_text(encoding="utf-8"),
            )

    def test_h2_sa_username_is_preserved(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            app = root / "application.yml"
            app.write_text(
                "spring:\n"
                "  datasource:\n"
                "    username: sa\n"
                "    password: passwd01\n",
                encoding="utf-8",
            )

            output = self.run_tool(root, "--write")
            text = app.read_text(encoding="utf-8")

            self.assertIn("redactions=1", output)
            self.assertIn("username: sa", text)
            self.assertIn("password: <REDACTED_PASSWORD>", text)

    def test_username_placeholder_preserves_safe_default(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            app = root / "application.properties"
            app.write_text(
                "spring.datasource.username=${SPRING_DATASOURCE_USERNAME:sa}\n"
                "audit.datasource.username=${AUDIT_DATASOURCE_USERNAME:kcfadmin}\n",
                encoding="utf-8",
            )

            output = self.run_tool(root, "--write")
            text = app.read_text(encoding="utf-8")

            self.assertIn("redactions=1", output)
            self.assertIn("spring.datasource.username=${SPRING_DATASOURCE_USERNAME:sa}", text)
            self.assertIn("audit.datasource.username=${AUDIT_DATASOURCE_USERNAME:<REDACTED_USERNAME>}", text)
            self.assertNotIn("kcfadmin", text)


if __name__ == "__main__":
    unittest.main()
