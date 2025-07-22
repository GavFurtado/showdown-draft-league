# Agent Guidelines for Showdown Draft League Backend (Go)

This document outlines essential commands and code style guidelines for agents working on this Go project.

## 1. Build, Lint, and Test Commands

*   **Run Development Server**: `go run ./cmd/main.go`
*   **Build Production Binary**: `go build -o ./build/server .`
*   **Run All Tests**: `go test ./...`
*   **Run Tests for a Specific Package**: `go test ./internal/services` (replace with actual package path)
*   **Format Code**: `go fmt ./...`
*   **Lint Code (Static Analysis)**: `go vet ./...`

## 2. Code Style Guidelines

*   **Imports**: Group imports. Standard library packages first, then external packages, then internal project packages. Each group should be separated by a blank line.
*   **Formatting**: Adhere to `go fmt` standards.
*   **Naming Conventions**:
    *   Exported (public) identifiers (functions, variables, types) use `CamelCase`.
    *   Unexported (private) identifiers use `camelCase`.
    *   Acronyms (e.g., `HTTP`, `URL`, `ID`) should be all uppercase when exported (e.g., `HTTPClient`, `UserID`) and all lowercase when unexported (e.g., `httpClient`, `userID`).
*   **Error Handling**: Errors are returned as the last return value. Always check errors explicitly using `if err != nil`. Propagate errors where appropriate.
*   **Structure**: Follow the existing layered architecture (controllers, services, repositories, models).
*   **Comments**: Add comments sparingly, focusing on *why* a piece of code exists or *what* complex logic does, rather than *how* it works.
- **Import Paths:** Standard internal package import path prefix is `github.com/GavFurtado/showdown-draft-league/new-backend/internal/`.
- **Naming:** Follow standard Go conventions (camelCase for variables/functions, PascalCase for types/structs/packages).
- **Imports:** Group standard library imports separately from third-party imports.
- **Formatting:** Use `go fmt`.
- **Error Handling:** Use the common package for common errors. These kinds of errors (in `errors.go`) are returned in Service Layer, but maybe used elsewhere if relevant. Note the convention for logging followed throughout the codebase
- **Types:** Use Go's built-in types and define structs for complex data. A `types.go` in the common package is used to define extra structs (such as, but not limited to, Request structs and DTO structs)
- **Comments:** Comments for structs and methods/functions include the name (identifier) of the 

### External Style Guides
No `.cursor/rules/`, `.cursorrules`, or `.github/copilot-instructions.md` found.

## 3. Agent Interaction Policy

*   **Explicit Approval**: Agents must never implement changes to the codebase or system state without explicit user approval. All modifications require a clear 'go ahead' from the user.

### Model Descriptions and Relationships

#### 1. `User` Model (`user.go`)

* **Description:** The `User` model represents an individual user of your platform. This is the top-level entity for anyone interacting with the system, primarily identified by their Discord and Showdown usernames.
* **Purpose:** To store core user information, manage their administrative status, and link to all the leagues they participate in or create.
* **Key Fields:**
    * `ID`: Unique identifier for the user (UUID).
    * `DiscordID`: Unique ID from Discord, used for authentication and possibly bot interactions.
    * `DiscordUsername`: The user's Discord username.
    * `ShowdownUsername`: The user's Pokémon Showdown username. This is crucial for tracking game results.
    * `IsAdmin`: Flag to indicate if the user has administrative privileges.
* **Relationships:**
    * **`LeaguesCreated []League`**: A user can create multiple leagues (one-to-many relationship where `CommissionerUserID` in `League` points to `UserID`).
    * **`Players []Player`**: A user can be a player in multiple leagues (one-to-many relationship where `UserID` in `Player` points to `UserID`).
* **Ties things together:** The `User` model is the foundational entry point for individuals using your platform, linking them to their roles as league commissioners or players across various leagues.

#### 2. `League` Model (`league.go`)

* **Description:** The `League` model defines a single draft tournament instance. It encapsulates the rules, status, and participants of a specific league.
* **Purpose:** To manage the overall structure and state of a Pokemon draft league, including its rules, duration, and who is participating.
* **Key Fields:**
    * `ID`: Unique identifier for the league (UUID).
    * `Name`: Name of the league.
    * `StartDate`, `EndDate`: Defines the league's active period.
    * `RulesetID`: (Nullable) A foreign key to a `Ruleset` model (which you might define later) to specify league-specific rules.
    * `Status`: Current state of the league (e.g., `SETUP`, `DRAFTING`, `REGULARSEASON`). Uses a custom `LeagueStatus` enum for type safety.
    * `MaxPokemonPerPlayer`: The maximum number of Pokémon a player can have on their roster.
    * `IsSnakeRoundDraft`: Determines the draft order style.
    * `CommissionerUserID`: The ID of the user who created and manages the league.
* **Relationships:**
    * **`CommissionerUser User`**: A league has one commissioner (`CommissionerUserID` points to `UserID`).
    * **`Players []Player`**: A league has many players (one-to-many relationship where `LeagueID` in `Player` points to `LeagueID`).
    * **`DefinedPokemon []LeaguePokemon`**: A league has many `LeaguePokemon`, representing its specific draft pool (one-to-many where `LeagueID` in `LeaguePokemon` points to `LeagueID`).
    * **`AllDraftedPokemon []DraftedPokemon`**: A league has many `DraftedPokemon`, representing all Pokémon ever drafted within it (one-to-many where `LeagueID` in `DraftedPokemon` points to `LeagueID`).
* **Ties things together:** The `League` model is the central hub for a specific tournament. It connects commissioners, players, the available Pokémon pool, and all drafted Pokémon within that tournament.

#### 3. `Player` Model (`player.go`)

* **Description:** The `Player` model represents a `User`'s participation *within a specific `League`*. A user can be a player in multiple leagues, and each instance of their participation is a separate `Player` record.
* **Purpose:** To track player-specific information relevant to a single league, such as their in-league name, team name, win/loss record, draft points, and draft position.
* **Key Fields:**
    * `ID`: Unique identifier for this player's entry in a league (UUID).
    * `UserID`: The associated `User`'s ID.
    * `LeagueID`: The `League`'s ID this player belongs to.
    * `InLeagueName`: The player's chosen name for this specific league.
    * `TeamName`: The name of their team.
    * `Wins`, `Losses`: Records for their performance in the league.
    * `DraftPoints`: Points available for drafting or free agency.
    * `DraftPosition`: Their turn order in the draft.
    * `IsCommissioner`: A flag indicating if *this player instance* is a commissioner for *this specific league* (redundant with `League.CommissionerUserID` but can be useful for quick checks if a user is a commissioner in the context of their `Player` record).
* **Relationships:**
    * **`User User`**: A player record belongs to one `User` (`UserID` points to `UserID`).
    * **`League League`**: A player record belongs to one `League` (`LeagueID` points to `LeagueID`).
    * **`Roster []PlayerRoster`**: A player has many `PlayerRoster` entries, representing the Pokémon on their active team (one-to-many where `PlayerID` in `PlayerRoster` points to `PlayerID`).
* **Ties things together:** The `Player` model acts as a bridge between a generic `User` and a specific `League`, holding all the league-specific stats and roles for that user. It also links directly to their active `PlayerRoster`.

#### 4. `PokemonSpecies` Model (`pokemon-species.go`)

* **Description:** The `PokemonSpecies` model stores static, universal data about a specific Pokémon species (e.g., Pikachu, Charizard). This data is typically loaded from an external source (like the PokeAPI) and doesn't change per league or draft.
* **Purpose:** To provide a central repository of Pokémon data, including their types, abilities, base stats, and sprites, which can then be referenced by various league-specific Pokémon instances.
* **Key Fields:**
    * `ID`: Numerical ID, likely from an external source like PokeAPI (e.g., 25 for Pikachu).
    * `Name`: The species name (e.g., "Pikachu").
    * `Types`: Array of Pokémon types (e.g., ["Electric"]).
    * `Abilities`: JSONB field storing a slice of `Ability` structs.
    * `Stats`: JSONB field storing a `BaseStats` struct.
    * `Sprites`: JSONB field storing `Sprites` struct for image URLs.
* **Relationships:** This model acts as a lookup table.
    * It is referenced by `LeaguePokemon` (`PokemonSpeciesID` in `LeaguePokemon` points to `PokemonSpecies.ID`).
    * It is referenced by `DraftedPokemon` (`PokemonSpeciesID` in `DraftedPokemon` points to `PokemonSpecies.ID`).
* **Ties things together:** This model is the canonical source of information for any Pokémon. It decouples generic Pokémon data from league-specific or player-specific instances.

#### 5. `LeaguePokemon` Model (`league-pokemon.go`)

* **Description:** The `LeaguePokemon` model defines which Pokémon species are available for drafting *within a particular league*, along with their specific cost for that league. It essentially represents a league's "draft pool".
* **Purpose:** To customize the set of available Pokémon and their costs for each individual league, allowing for unique league configurations.
* **Key Fields:**
    * `ID`: Unique identifier (UUID).
    * `LeagueID`: The `League` this available Pokémon belongs to.
    * `PokemonSpeciesID`: The `PokemonSpecies` that is available.
    * `Cost`: The draft cost of this Pokémon *in this specific league*.
    * `IsAvailable`: Flag indicating if this Pokémon species can currently be drafted in this league.
* **Relationships:**
    * **`League League`**: Belongs to one `League` (`LeagueID` points to `LeagueID`).
    * **`PokemonSpecies PokemonSpecies`**: Refers to one `PokemonSpecies` (`PokemonSpeciesID` points to `PokemonSpecies.ID`).
* **Ties things together:** `LeaguePokemon` links a `League` to a `PokemonSpecies`, providing league-specific properties like cost and availability for that species, forming the draft pool.

#### 6. `Draft` Model (`draft.go`)

* **Description:** The `Draft` model manages the real-time state and progression of the draft process for a specific league.
* **Purpose:** To orchestrate the draft, track the current turn, round, and accumulated picks, and manage the draft's start and end times.
* **Key Fields:**
    * `ID`: Unique identifier for the draft (UUID).
    * `LeagueID`: The `League` this draft belongs to.
    * `Status`: Current state of the draft (e.g., `PENDING`, `STARTED`, `COMPLETED`). Uses a custom `DraftStatus` enum.
    * `CurrentTurnPlayerID`: (Nullable) The `Player` whose turn it currently is.
    * `CurrentRound`, `CurrentPickInRound`: Tracks the draft's progress.
    * `PlayersWithAccumulatedPicks`: JSONB map to handle complex draft mechanics (e.g., trading picks).
    * `TurnTimeLimit`: How long a player has to make a pick.
* **Relationships:**
    * **`League League`**: A draft belongs to one `League` (`LeagueID` points to `LeagueID`).
    * **`CurrentTurnPlayer Player`**: The current player making a pick (`CurrentTurnPlayerID` points to `PlayerID`).
* **Ties things together:** The `Draft` model is the state machine for the drafting phase of a `League`. It directly references the `League` it's managing and the `Player` whose turn it is.

#### 7. `DraftedPokemon` Model (`drafted-pokemon.go`)

* **Description:** The `DraftedPokemon` model represents a specific instance of a `PokemonSpecies` that has been successfully drafted by a `Player` within a `League`. This is distinct from `LeaguePokemon` which defines availability; `DraftedPokemon` tracks actual ownership.
* **Purpose:** To record which player drafted which Pokémon species, when it was drafted (round/pick), and its current status (e.g., if it has been released).
* **Key Fields:**
    * `ID`: Unique identifier for this drafted instance (UUID).
    * `LeagueID`: The `League` where this Pokémon was drafted.
    * `PlayerID`: The `Player` who drafted this Pokémon.
    * `PokemonSpeciesID`: The `PokemonSpecies` that was drafted.
    * `DraftRoundNumber`, `DraftPickNumber`: Details of when it was drafted.
    * `IsReleased`: Flag if the Pokémon has been released back to the available pool. This is used for the weekly points drop/pickup system, where players can use accumulated points to drop a drafted Pokemon (setting `IsReleased` to true) and pick up an available one (setting `IsReleased` to false on a released Pokemon and assigning it to them).
* **Relationships:**
    * **`League League`**: Belongs to one `League` (`LeagueID` points to `LeagueID`).
    * **`Player Player`**: Was drafted by one `Player` (`PlayerID` points to `PlayerID`).
    * **`PokemonSpecies PokemonSpecies`**: Is an instance of one `PokemonSpecies` (`PokemonSpeciesID` points to `PokemonSpecies.ID`).
    * **`LeaguePokemon LeaguePokemon`**: You have a `LeaguePokemonID` field but no relationship tag. If you intend to link `DraftedPokemon` back to the specific `LeaguePokemon` entry that enabled its draft, you'd need to uncomment/add a `LeaguePokemonID` field and a corresponding relationship. *However, the current setup implies `DraftedPokemon` tracks the `PokemonSpecies` that was drafted, not necessarily the specific `LeaguePokemon` entry. This might be a missing or intended abstraction.*
* **Ties things together:** `DraftedPokemon` forms the historical record of draft picks and links `League`, `Player`, and `PokemonSpecies` to represent an actual Pokémon acquisition. It directly feeds into a `Player`'s active `PlayerRoster`.

#### 8. `PlayerRoster` Model (`player-roster.go`)

* **Description:** The `PlayerRoster` model represents a `DraftedPokemon` that is currently on a specific `Player`'s active team.
* **Purpose:** To manage the current active lineup of Pokémon for each player in a league.
* **Key Fields:**
    * `ID`: Unique identifier (UUID).
    * `PlayerID`: The `Player` this roster entry belongs to.
    * `DraftedPokemonID`: The specific `DraftedPokemon` instance that is on the roster. This has a `unique` constraint, implying a `DraftedPokemon` can only be on one `PlayerRoster` at a time.
* **Relationships:**
    * **`Player Player`**: Belongs to one `Player` (`PlayerID` points to `PlayerID`).
    * **`DraftedPokemon DraftedPokemon`**: Refers to one `DraftedPokemon` (`DraftedPokemonID` points to `DraftedPokemonID`).
* **Ties things together:** `PlayerRoster` is the direct link from a `Player` to their actively chosen `DraftedPokemon`. It essentially defines a player's team for competitive play.

#### 9. `Game` Model (`game.go`)

* **Description:** The `Game` model represents a single best-of-X series played between two `Player`s in a `League`.
* **Purpose:** To track individual match outcomes, scores, and associated replay links within a league's competitive season.
* **Key Fields:**
    * `ID`: Unique identifier (UUID).
    * `LeagueID`: The `League` the game took place in.
    * `Player1ID`, `Player2ID`: The two `Player`s participating.
    * `WinnerID`, `LoserID`: (Nullable) The IDs of the winning and losing players.
    * `Player1Wins`, `Player2Wins`: Scores for the series.
    * `RoundNumber`: The week/round of the league season this game occurred in.
    * `Status`: Current state of the game (e.g., `pending`, `completed`, `disputed`).
    * `ReportedByUserID`: The `User` who reported the game results.
    * `ShowdownReplayLinks`: Array of URLs to Pokémon Showdown replays.
* **Relationships:**
    * **`League League`**: A game belongs to one `League` (`LeagueID` points to `LeagueID`).
    * **`Player1 Player`**, **`Player2 Player`**: The two players involved (`Player1ID`, `Player2ID` point to `PlayerID`).
    * **`Winner Player`**, **`Loser Player`**: The winning and losing players (`WinnerID`, `LoserID` point to `PlayerID`).
    * **`Reporter User`**: The user who reported the game (`ReportedByUserID` points to `UserID`).
* **Ties things together:** The `Game` model links `League`s, `Player`s, and `User`s to record and manage competitive matches, directly contributing to `Player` win/loss records.
