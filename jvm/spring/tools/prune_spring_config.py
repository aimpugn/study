#!/usr/bin/env python3
"""Prune unsafe Spring config profiles and redact local secrets.

The tool is intentionally conservative about what it prints: reports contain
paths, counts, and action kinds, but never the original secret value.
"""

from __future__ import annotations

import argparse
import dataclasses
import re
import sys
from collections.abc import Callable
from pathlib import Path


DEFAULT_KEEP_PROFILES = frozenset({"common", "junit", "local"})
DEFAULT_DROP_PROFILES = frozenset({"dev", "real", "stg"})
REDACTED_IP = "127.0.0.1"
DEFAULT_EXCLUDED_DIRS = {
    ".git",
    ".gradle",
    ".idea",
    ".mvn",
    ".vscode",
    "__pycache__",
    "build",
    "dist",
    "node_modules",
    "out",
    "target",
}
CONFIG_EXTENSIONS = {".properties", ".yaml", ".yml"}
APPLICATION_CONFIG_RE = re.compile(
    r"^application(?:-(?P<profile>[A-Za-z0-9_.-]+))?\.(?P<ext>properties|ya?ml)$"
)
PROFILE_SUFFIX_RE = re.compile(
    r"^(?P<stem>.+)-(?P<profile>[A-Za-z0-9_.-]+)\.(?P<ext>properties|ya?ml)$"
)
YAML_DOC_SEPARATOR_RE = re.compile(r"^\s*---(?:\s*#.*)?\r?\n?$")
PROPERTIES_DOC_SEPARATOR_RE = re.compile(r"^\s*[#!]\s*---\s*\r?\n?$")
YAML_KEY_RE = re.compile(r"^(?P<indent>\s*)(?:-\s*)?(?P<key>['\"]?[A-Za-z0-9_.-]+['\"]?)\s*:\s*(?P<value>.*)$")
PROPERTIES_ENTRY_RE = re.compile(
    r"^(?P<prefix>\s*(?P<key>[^#!:=\s][^:=\s]*?)(?P<sep>\s*[:=]\s*|\s+))"
    r"(?P<value>.*?)(?P<newline>\r?\n)?$"
)
IPV4_RE = re.compile(
    r"\b(?:(?:25[0-5]|2[0-4]\d|1?\d?\d)\.){3}(?:25[0-5]|2[0-4]\d|1?\d?\d)\b"
)
PROFILE_TOKEN_RE = re.compile(r"[A-Za-z0-9_.-]+")
SPRING_PLACEHOLDER_RE = re.compile(r"\$\{(?P<name>[^}:]+)(?::(?P<default>[^}]*))?\}")
SAFE_USERNAME_VALUES = {"sa"}


@dataclasses.dataclass(frozen=True)
class Settings:
    root: Path
    write: bool


@dataclasses.dataclass
class ChangeStats:
    redactions: int = 0
    profile_documents_removed: int = 0

    @property
    def changed(self) -> bool:
        return self.redactions > 0 or self.profile_documents_removed > 0


@dataclasses.dataclass
class FileAction:
    path: Path
    kind: str
    reason: str
    stats: ChangeStats = dataclasses.field(default_factory=ChangeStats)
    skipped_read: bool = False


@dataclasses.dataclass(frozen=True)
class TextFile:
    text: str
    encoding: str


def is_loopback_or_bind_ip(value: str) -> bool:
    return value.startswith("127.") or value == "0.0.0.0" or value == "255.255.255.255"


def redact_ipv4(value: str, replacement: str) -> tuple[str, int]:
    count = 0

    def replace(match: re.Match[str]) -> str:
        nonlocal count
        found = match.group(0)
        if is_loopback_or_bind_ip(found):
            return found
        count += 1
        return replacement

    return IPV4_RE.sub(replace, value), count


def normalize_key(key: str) -> str:
    return re.sub(r"[^a-z0-9]", "", key.lower())


def normalize_path(parts: list[str]) -> tuple[str, ...]:
    normalized: list[str] = []
    for part in parts:
        stripped = part.strip("'\"")
        normalized.extend(token.lower() for token in stripped.split(".") if token)
    return tuple(normalized)


def key_category(path_parts: list[str]) -> str | None:
    joined = normalize_key(".".join(path_parts))
    leaf = normalize_key(path_parts[-1]) if path_parts else ""
    if any(token in joined for token in ("password", "passwd")) or leaf == "pwd":
        return "password"
    if "username" in joined or leaf in {"user", "userid"}:
        return "username"
    if any(token in joined for token in ("secret", "token", "apikey", "apik", "credential")):
        return "secret"
    if leaf == "key" or any(token in joined for token in ("privatekey", "accesskey", "clientsecret")):
        return "secret"
    return None


def network_key(path_parts: list[str]) -> bool:
    joined = normalize_key(".".join(path_parts))
    leaf = normalize_key(path_parts[-1]) if path_parts else ""
    return (
        leaf in {"url", "uri", "host", "ip", "addr", "address", "jdbcurl"}
        or joined.endswith("jdbcurl")
        or "serverip" in joined
        or "sftpip" in joined
    )


def placeholder_for(category: str) -> str:
    if category == "password":
        return "<REDACTED_PASSWORD>"
    if category == "username":
        return "<REDACTED_USERNAME>"
    return "<REDACTED_SECRET>"


def strip_wrapping_quotes(value: str) -> str:
    stripped = value.strip()
    if len(stripped) >= 2 and stripped[0] == stripped[-1] and stripped[0] in {"'", '"'}:
        return stripped[1:-1]
    return stripped


def safe_username_value(value: str) -> bool:
    return strip_wrapping_quotes(value).strip().lower() in SAFE_USERNAME_VALUES


def value_is_only_spring_placeholders(value: str) -> bool:
    stripped = strip_wrapping_quotes(value)
    return bool(SPRING_PLACEHOLDER_RE.search(stripped)) and not SPRING_PLACEHOLDER_RE.sub("", stripped).strip()


def redact_spring_placeholder_defaults(
    value: str,
    replacement: str,
    should_redact_default: Callable[[str], bool] | None = None,
) -> tuple[str, int, bool]:
    count = 0
    found = False

    def replace(match: re.Match[str]) -> str:
        nonlocal count, found
        found = True
        default = match.group("default")
        if default is None or default == "":
            return match.group(0)
        if should_redact_default is not None and not should_redact_default(default):
            return match.group(0)
        count += 1
        return "${" + match.group("name") + ":" + replacement + "}"

    return SPRING_PLACEHOLDER_RE.sub(replace, value), count, found


def extract_profiles(value: str) -> tuple[set[str], bool]:
    lowered = value.strip().lower()
    tokens = {
        token
        for token in PROFILE_TOKEN_RE.findall(lowered)
        if token not in {"and", "or", "not"}
    }
    has_negation = "!" in lowered or re.search(r"\bnot\b", lowered) is not None
    return tokens, has_negation


def profiles_allowed(value: str, keep_profiles: frozenset[str]) -> bool:
    tokens, has_negation = extract_profiles(value)
    if has_negation:
        return False
    return bool(tokens) and tokens <= keep_profiles


def is_profile_selector(path_parts: list[str]) -> bool:
    path = normalize_path(path_parts)
    return path == ("spring", "config", "activate", "on-profile") or path == ("spring", "profiles")


def read_text(path: Path) -> TextFile:
    data = path.read_bytes()
    for encoding in ("utf-8-sig", "utf-8", "cp949", "euc-kr"):
        try:
            return TextFile(data.decode(encoding), encoding)
        except UnicodeDecodeError:
            continue
    raise UnicodeDecodeError("supported-config-encodings", data, 0, 1, "not utf-8/cp949/euc-kr")


def write_text(path: Path, text_file: TextFile, text: str) -> None:
    encoding = "utf-8" if text_file.encoding == "utf-8-sig" else text_file.encoding
    path.write_bytes(text.encode(encoding))


def profile_from_application_name(path: Path) -> str | None:
    match = APPLICATION_CONFIG_RE.match(path.name)
    if not match:
        return None
    profile = match.group("profile")
    return profile.lower() if profile else None


def profile_suffix(path: Path) -> str | None:
    match = PROFILE_SUFFIX_RE.match(path.name)
    if not match:
        return None
    return match.group("profile").lower()


def should_remove_without_read(path: Path, settings: Settings) -> tuple[bool, str]:
    app_profile = profile_from_application_name(path)
    if app_profile is not None and app_profile not in DEFAULT_KEEP_PROFILES:
        return True, f"application profile '{app_profile}' is not kept"

    suffix = profile_suffix(path)
    if suffix in DEFAULT_DROP_PROFILES:
        return True, f"filename profile suffix '{suffix}' is configured for removal"

    return False, ""


def is_sanitizable_application_file(path: Path) -> bool:
    match = APPLICATION_CONFIG_RE.match(path.name)
    return bool(match and match.group("profile") in (None, "common", "junit", "local"))


def iter_candidate_paths(settings: Settings) -> list[Path]:
    candidates: list[Path] = []
    for path in settings.root.rglob("*"):
        if path.is_dir():
            continue
        if any(part in DEFAULT_EXCLUDED_DIRS for part in path.parts):
            continue
        if path.suffix.lower() not in CONFIG_EXTENSIONS:
            continue
        removable, _ = should_remove_without_read(path, settings)
        if removable or is_sanitizable_application_file(path):
            candidates.append(path)
    return sorted(candidates)


def split_yaml_comment(value: str) -> tuple[str, str]:
    in_single = False
    in_double = False
    escaped = False
    for index, char in enumerate(value):
        if escaped:
            escaped = False
            continue
        if char == "\\" and in_double:
            escaped = True
            continue
        if char == "'" and not in_double:
            in_single = not in_single
            continue
        if char == '"' and not in_single:
            in_double = not in_double
            continue
        if char == "#" and not in_single and not in_double:
            before = value[index - 1] if index > 0 else " "
            if before.isspace():
                return value[:index].rstrip(), value[index:]
    return value.rstrip(), ""


def sanitize_value(path_parts: list[str], value: str, settings: Settings) -> tuple[str, int]:
    newline = "\n" if value.endswith("\n") else ""
    body = value[:-1] if newline else value
    category = key_category(path_parts)
    redactions = 0

    if category == "username" and body.strip():
        if safe_username_value(body):
            return body + newline, redactions
        body, placeholder_count, placeholder_found = redact_spring_placeholder_defaults(
            body,
            placeholder_for(category),
            should_redact_default=lambda default: not safe_username_value(default),
        )
        redactions += placeholder_count
        if placeholder_found and (placeholder_count > 0 or value_is_only_spring_placeholders(body)):
            return body + newline, redactions
        body = placeholder_for(category)
        redactions += 1
    elif category and body.strip():
        body, placeholder_count, placeholder_found = redact_spring_placeholder_defaults(body, placeholder_for(category))
        redactions += placeholder_count
        if placeholder_found and (placeholder_count > 0 or value_is_only_spring_placeholders(body)):
            return body + newline, redactions
        body = placeholder_for(category)
        redactions += 1
    else:
        body, ip_count = redact_ipv4(body, REDACTED_IP)
        redactions += ip_count

    if network_key(path_parts):
        body, ip_count = redact_ipv4(body, REDACTED_IP)
        redactions += ip_count

    return body + newline, redactions


def sanitize_yaml_line(line: str, path_parts: list[str], settings: Settings) -> tuple[str, int]:
    match = YAML_KEY_RE.match(line)
    if not match:
        sanitized, count = redact_ipv4(line, REDACTED_IP)
        return sanitized, count

    value = match.group("value")
    if not value.strip():
        sanitized, count = redact_ipv4(line, REDACTED_IP)
        return sanitized, count

    newline = "\n" if line.endswith("\n") else ""
    body = line[:-1] if newline else line
    value_body = match.group("value")[:-1] if match.group("value").endswith("\n") else match.group("value")
    value_part, comment = split_yaml_comment(value_body)
    prefix = body[: match.start("value")]
    sanitized_value, count = sanitize_value(path_parts, value_part.strip(), settings)
    sanitized_comment, comment_count = redact_ipv4(comment, REDACTED_IP)
    separator = " " if sanitized_comment else ""
    return f"{prefix}{sanitized_value.rstrip()}{separator}{sanitized_comment}{newline}", count + comment_count


def sanitize_properties_line(line: str, settings: Settings) -> tuple[str, int]:
    match = PROPERTIES_ENTRY_RE.match(line)
    if not match:
        sanitized, count = redact_ipv4(line, REDACTED_IP)
        return sanitized, count

    key = match.group("key")
    value = match.group("value")
    newline = match.group("newline") or ""
    path_parts = key.split(".")
    if key.lower() in {"spring.profiles.active", "spring.profiles.default"}:
        tokens, has_negation = extract_profiles(value)
        if has_negation or any(token not in DEFAULT_KEEP_PROFILES for token in tokens):
            return f"{match.group('prefix')}local{newline}", 1

    sanitized_value, count = sanitize_value(path_parts, value, settings)
    return f"{match.group('prefix')}{sanitized_value.rstrip()}{newline}", count


@dataclasses.dataclass
class YamlDocument:
    separator: str
    lines: list[str]


def split_yaml_documents(text: str) -> list[YamlDocument]:
    documents: list[YamlDocument] = []
    separator = ""
    lines: list[str] = []
    for line in text.splitlines(keepends=True):
        if YAML_DOC_SEPARATOR_RE.match(line):
            documents.append(YamlDocument(separator, lines))
            separator = line
            lines = []
        else:
            lines.append(line)
    documents.append(YamlDocument(separator, lines))
    if documents and not documents[0].separator and not documents[0].lines:
        documents.pop(0)
    return documents


def yaml_document_profile(doc: YamlDocument) -> str | None:
    stack: list[tuple[int, str]] = []
    for line in doc.lines:
        match = YAML_KEY_RE.match(line)
        if not match:
            continue
        indent = len(match.group("indent").replace("\t", "    "))
        while stack and indent <= stack[-1][0]:
            stack.pop()
        key = match.group("key").strip("'\"")
        path_parts = [item for _, item in stack] + [key]
        value, _ = split_yaml_comment(match.group("value"))
        if is_profile_selector(path_parts) and value.strip():
            return value.strip().strip("'\"")
        if not value.strip():
            stack.append((indent, key))
    return None


def sanitize_yaml(text: str, settings: Settings) -> tuple[str, ChangeStats]:
    documents = split_yaml_documents(text)
    output: list[str] = []
    stats = ChangeStats()

    for doc in documents:
        profile = yaml_document_profile(doc)
        if profile and not profiles_allowed(profile, DEFAULT_KEEP_PROFILES):
            stats.profile_documents_removed += 1
            continue

        if doc.separator:
            output.append(doc.separator)
        stack: list[tuple[int, str]] = []
        for line in doc.lines:
            match = YAML_KEY_RE.match(line)
            if match:
                indent = len(match.group("indent").replace("\t", "    "))
                while stack and indent <= stack[-1][0]:
                    stack.pop()
                key = match.group("key").strip("'\"")
                path_parts = [item for _, item in stack] + [key]
                sanitized, count = sanitize_yaml_line(line, path_parts, settings)
                value, _ = split_yaml_comment(match.group("value"))
                if not value.strip():
                    stack.append((indent, key))
            else:
                sanitized, count = redact_ipv4(line, REDACTED_IP)
            stats.redactions += count
            output.append(sanitized)

    return "".join(output), stats


@dataclasses.dataclass
class PropertiesDocument:
    separator: str
    lines: list[str]


def split_properties_documents(text: str) -> list[PropertiesDocument]:
    documents: list[PropertiesDocument] = []
    separator = ""
    lines: list[str] = []
    for line in text.splitlines(keepends=True):
        if PROPERTIES_DOC_SEPARATOR_RE.match(line):
            documents.append(PropertiesDocument(separator, lines))
            separator = line
            lines = []
        else:
            lines.append(line)
    documents.append(PropertiesDocument(separator, lines))
    if documents and not documents[0].separator and not documents[0].lines:
        documents.pop(0)
    return documents


def properties_document_profile(doc: PropertiesDocument) -> str | None:
    for line in doc.lines:
        match = PROPERTIES_ENTRY_RE.match(line)
        if not match:
            continue
        key = match.group("key").strip().lower()
        if key in {"spring.config.activate.on-profile", "spring.profiles"}:
            return match.group("value").strip()
    return None


def sanitize_properties(text: str, settings: Settings) -> tuple[str, ChangeStats]:
    documents = split_properties_documents(text)
    output: list[str] = []
    stats = ChangeStats()

    for doc in documents:
        profile = properties_document_profile(doc)
        if profile and not profiles_allowed(profile, DEFAULT_KEEP_PROFILES):
            stats.profile_documents_removed += 1
            continue
        if doc.separator:
            output.append(doc.separator)
        for line in doc.lines:
            sanitized, count = sanitize_properties_line(line, settings)
            stats.redactions += count
            output.append(sanitized)

    return "".join(output), stats


def sanitize_config(text: str, path: Path, settings: Settings) -> tuple[str, ChangeStats]:
    if path.suffix.lower() == ".properties":
        return sanitize_properties(text, settings)
    return sanitize_yaml(text, settings)


def process_file(path: Path, settings: Settings) -> FileAction:
    removable, reason = should_remove_without_read(path, settings)
    if removable:
        if settings.write:
            path.unlink()
        return FileAction(path=path, kind="REMOVE", reason=reason, skipped_read=True)

    text_file = read_text(path)
    sanitized, stats = sanitize_config(text_file.text, path, settings)
    if sanitized == text_file.text:
        return FileAction(path=path, kind="KEEP", reason="no unsafe value or profile document matched", stats=stats)

    if settings.write:
        write_text(path, text_file, sanitized)
    return FileAction(path=path, kind="UPDATE", reason="sanitized kept Spring config", stats=stats)


def build_settings(argv: list[str]) -> Settings:
    parser = argparse.ArgumentParser(
        description="Remove unsafe Spring profile config files and redact secrets from kept application configs."
    )
    parser.add_argument("root", nargs="?", default=".", help="root directory to scan")
    parser.add_argument("--write", action="store_true", help="apply removals and redactions; default is dry-run")
    args = parser.parse_args(argv)
    return Settings(root=Path(args.root).resolve(), write=args.write)


def print_report(actions: list[FileAction], settings: Settings) -> None:
    mode = "WRITE" if settings.write else "DRY-RUN"
    print(f"mode={mode} root={settings.root}")
    for action in actions:
        relative = action.path.relative_to(settings.root)
        details = []
        if action.skipped_read:
            details.append("skipped-read")
        if action.stats.profile_documents_removed:
            details.append(f"profile-docs-removed={action.stats.profile_documents_removed}")
        if action.stats.redactions:
            details.append(f"redactions={action.stats.redactions}")
        suffix = f" ({', '.join(details)})" if details else ""
        print(f"{action.kind:6} {relative} - {action.reason}{suffix}")
    counts = {kind: sum(1 for action in actions if action.kind == kind) for kind in ("REMOVE", "UPDATE", "KEEP")}
    print(f"summary remove={counts['REMOVE']} update={counts['UPDATE']} keep={counts['KEEP']}")
    if not settings.write and (counts["REMOVE"] or counts["UPDATE"]):
        print("dry-run only; rerun with --write to apply these actions")


def main(argv: list[str] | None = None) -> int:
    settings = build_settings(sys.argv[1:] if argv is None else argv)
    if not settings.root.exists() or not settings.root.is_dir():
        print(f"error: root directory does not exist: {settings.root}", file=sys.stderr)
        return 2
    actions = [process_file(path, settings) for path in iter_candidate_paths(settings)]
    print_report(actions, settings)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
