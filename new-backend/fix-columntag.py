from __future__ import annotations
import re
from pathlib import Path


def to_snake_case(name: str) -> str:
    """Convert PascalCase or camelCase to snake_case."""
    s1 = re.sub(r"(.)([A-Z][a-z]+)", r"\1_\2", name)
    s2 = re.sub(r"([a-z0-9])([A-Z])", r"\1_\2", s1)
    return s2.lower()


def process_file(filepath: str | Path) -> None:
    path = Path(filepath)
    with path.open("r", encoding="utf-8") as f:
        lines = f.readlines()

    new_lines: list[str] = []

    for line in lines:
        # Only process lines with gorm tag
        if 'gorm:' not in line:
            new_lines.append(line)
            continue

        # Skip lines without `column:` since we only touch those
        if 'column:' not in line:
            new_lines.append(line)
            continue

        # If it’s a relationship field, remove `;column:Something`
        if 'foreignKey:' in line or 'references:' in line:
            line = re.sub(r';\s*column:[A-Za-z_]+', '', line)
            new_lines.append(line)
            continue

        # Otherwise, convert column name to snake_case
        def repl(match: re.Match[str]) -> str:
            original = match.group(1)
            return f'column:{to_snake_case(original)}'

        line = re.sub(r'column:([A-Za-z_]+)', repl, line)
        new_lines.append(line)

    with path.open("w", encoding="utf-8") as f:
        f.writelines(new_lines)


def main() -> None:
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

    for file in files_to_process:
        process_file(file)
        print(f"Processed {file}")


if __name__ == "__main__":
    main()
