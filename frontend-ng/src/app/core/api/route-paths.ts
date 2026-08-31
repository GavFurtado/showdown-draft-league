import { UUID } from '../../shared/types/branded-strings';

export const routePaths = {
  auth: {
    discordLogin: `/auth/discord/login`,
    discordCallback: `/auth/discord/callback`,
    logout: `/auth/logout`,
    // Dev-only impersonation endpoints (backend registers these when ENV=dev).
    dev: {
      users: `/auth/dev/users`,
      login: `/auth/dev/login`,
      memberships: `/auth/dev/memberships`,
    },
  },

  // TODO: Server shouldn't expose this.
  profile: `/api/profile`,
  users: {
    myDiscordDetails: `/api/users/me/discord`,
    myLeagues: `/api/users/me/leagues`,
    profile: `/api/users/profile`,
    me: `/api/users/me`,
    leagueMembersById: (id: UUID) => `/api/users/${id}/members`,

    // TODO: Server needs to expose more endpoints here for admin stuff
    // base:
    // byId:
    // ...and others
  },

  pokemonSpecies: {
    // TODO: server needs to make this kebab-case
    base: `/api/pokemon_species`,
    byName: (name: string) => `/api/pokemon_species/name/${name}`,
    byId: (id: number) => `/api/pokemon_species/${id}`,
  },

  members: {
    byId: (id: UUID) => `/api/members/${id}`,
    profile: (id: UUID) => `/api/members/${id}/profile`,
    draftPoints: (id: UUID) => `/api/members/${id}/draft-points`,
    record: (id: UUID) => `/api/members/${id}/record`,
    draftPosition: (id: UUID) => `/api/members/${id}/draft-position`,
    roster: (id: UUID) => `/api/members/${id}/roster`,
  },

  leagues: {
    base: `/api/leagues/`,
    byId: (leagueId: UUID) => `/api/leagues/${leagueId}`,

    members: {
      base: (leagueId: UUID) => `/api/leagues/${leagueId}/members`,
      join: (leagueId: UUID) => `/api/leagues/${leagueId}/members/join`,
    },

    poolEntries: {
      base: (leagueId: UUID) => `/api/leagues/${leagueId}/pool-entries`,
      byId: (leagueId: UUID, id: UUID) => `/api/leagues/${leagueId}/pool-entries/${id}`,
      byAvailable: (leagueId: UUID) => `/api/leagues/${leagueId}/pool-entries/available`,
      single: (leagueId: UUID) => `/api/leagues/${leagueId}/pool-entries/single`,
      batch: (leagueId: UUID) => `/api/leagues/${leagueId}/pool-entries/batch`,
    },

    draft: {
      base: (leagueId: UUID) => `/api/leagues/${leagueId}/draft`,
      // TF: why is this here? A league can only have one draft right now
      // did I really have this much forethought?
      byId: (leagueId: UUID, draftId: UUID) => `/api/leagues/${leagueId}/draft/${draftId}`,
      start: (leagueId: UUID) => `/api/leagues/${leagueId}/draft/start`,
      pick: (leagueId: UUID) => `/api/leagues/${leagueId}/draft/pick`,
      skip: (leagueId: UUID) => `/api/leagues/${leagueId}/draft/skip`,
    },

    draftPicks: {
      base: (leagueId: UUID) => `/api/leagues/${leagueId}/draft-picks`,
      // this GET does same as base
      history: (leagueId: UUID) => `/api/leagues/${leagueId}/draft-picks/history`,

      // returned json object key is snake_case 'next_pick_number'
      // TODO: server needs to fix
      nextPickNumber: (leagueId: UUID) => `/api/leagues/${leagueId}/draft-picks/next-pick-number`,

      // TODO: server needs to rename this to members
      player: {
        byId: (leagueId: UUID, playerId: UUID) => `/api/leagues/${leagueId}/draft-picks/player/${playerId}`,
      },
    },

    claims: {
      base: (leagueId: UUID) => `/api/leagues/${leagueId}/claims`,
      byId: (leagueId: UUID, id: UUID) => `/api/leagues/${leagueId}/claims/${id}`,
      released: (leagueId: UUID) => `/api/leagues/${leagueId}/claims/released`,

      player: {
        byId: (leagueId: UUID, playerId: UUID) => `/api/leagues/${leagueId}/claims/player/${playerId}`,
      },
    },

    games: {
      base: (leagueId: UUID) => `/api/leagues/${leagueId}/games`,
      byId: (leagueId: UUID, gameId: UUID) => `/api/leagues/${leagueId}/games/${gameId}`,

      report: (leagueId: UUID, gameId: UUID) => `/api/leagues/${leagueId}/games/report/${gameId}`,
      finalize: (leagueId: UUID, gameId: UUID) => `/api/leagues/${leagueId}/games/finalize/${gameId}`,

      startSeason: (leagueId: UUID) => `/api/leagues/${leagueId}/games/start-season`,
      generatePlayoffs: (leagueId: UUID) => `/api/leagues/${leagueId}/games/generate-playoffs`,

      members: {
        byId: (leagueId: UUID, memberId: UUID) => `/api/leagues/${leagueId}/games/members/${memberId}`,
      },
    },

    transfers: {
      startWindow: (leagueId: UUID) => `/api/leagues/${leagueId}/transfers/start`,
      endWindow: (leagueId: UUID) => `/api/leagues/${leagueId}/transfers/end`,

      dropClaim: (leagueId: UUID, claimId: UUID) => `/api/leagues/${leagueId}/transfers/drop/${claimId}`,
      pickupPoolEntry: (leagueId: UUID, poolEntryId: UUID) =>
        `/api/leagues/${leagueId}/transfers/pickup/${poolEntryId}`,
    },
  },
};