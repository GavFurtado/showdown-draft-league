import re
import os

def process_file(filepath: str) -> None:
    """
    Reads a Go source file and ensures that each exported struct field
    has both a `json` and a `gorm` tag based on its field name.

    Args:
        filepath: Path to the Go source file to process.
    """

    # Read file contents
    with open(filepath, "r", encoding="utf-8") as f:
        lines: list[str] = f.readlines()

    new_lines: list[str] = []

    # Regex pattern to match struct field lines
    field_pattern: re.Pattern[str] = re.compile(
        r"""^
        (\s*)                     # indentation
        ([A-Za-z_]\w*)            # field name
        \s+
        ([\w\.\*\[\]]+)           # field type (allows slices, pointers, etc.)
        (?:\s+`([^`]*)`)?         # optional tag content (without backticks)
        \s*$
        """,
        re.VERBOSE,
    )

    for line in lines:
        # Regex to capture:
        # Group 1: Indentation
        # Group 2: Field Name (PascalCase)
        # Group 3: Field Type
        # Group 4: Full existing tag string (optional, including backticks)
        # Group 5: Content inside the backticks (the actual tags)
        match: re.Match[str] | None = field_pattern.match(line)
        if not match:
            new_lines.append(line)
            continue

        indent: str = match.group(1)
        field_name: str = match.group(2)
        field_type: str = match.group(3)
        tags_content: str = match.group(4) or ""

        # Skip unexported fields or comments
        if not field_name[0].isupper() or field_name.startswith("//"):
            new_lines.append(line)
            continue

        # --- Process JSON tag ---
        new_json_tag: str = f'json:"{field_name}"'

        if 'json:"-"' in tags_content:
            # Keep json:"-" untouched
            pass
        elif "json:" in tags_content:
            # Replace existing json tag
            tags_content = re.sub(r'json:"[^"]*"', new_json_tag, tags_content)
        else:
            # Add missing json tag
            tags_content = f"{tags_content} {new_json_tag}".strip()

        # --- Process GORM tag ---
        new_gorm_column: str = f"column:{field_name}"

        if "gorm:" in tags_content:
            if "column:" in tags_content:
                # Replace existing column value
                tags_content = re.sub(r'column:[^;"]*', new_gorm_column, tags_content)
            else:
                # Append new column attribute to gorm tag
                tags_content = re.sub(
                    r'gorm:"([^"]*)"',
                    lambda m: f'gorm:"{m.group(1)};{new_gorm_column}"',
                    tags_content,
                )
        else:
            # Add new gorm tag if missing
            tags_content = f"{tags_content} gorm:\"{new_gorm_column}\"".strip()

        # Rebuild the struct line
        new_line: str = f"{indent}{field_name} {field_type} `{tags_content}`\n"
        new_lines.append(new_line)

    # Write updated content back to the file
    with open(filepath, "w", encoding="utf-8") as f:
        f.writelines(new_lines)


def main() -> None:
    """Processes all Go model files in the project."""

    files_to_process: list[str] = [
        "./internal/models/user.go",
        "./internal/models/league.go",
        "./internal/models/player.go",
        "./internal/models/drafted-pokemon.go",
        "./internal/models/draft.go",
        "./internal/models/league-pokemon.go",
        "./internal/models/game.go",
        "./internal/models/pokemon-species.go",
        "./internal/models/player-roster.go",
        "./internal/common/types.go",
    ]

    for path in files_to_process:
        if os.path.exists(path):
            print(f"Processing {path}")
            process_file(path)
        else:
            print(f"⚠️ Skipped missing file: {path}")


if __name__ == "__main__":
    main()
