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
    leagueMembersById: (id: string) => `/api/users/${id}/members`,

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
    byId: (id: string) => `/api/members/${id}`,
    profile: (id: string) => `/api/members/${id}/profile`,
    draftPoints: (id: string) => `/api/members/${id}/draft-points`,
    record: (id: string) => `/api/members/${id}/record`,
    draftPosition: (id: string) => `/api/members/${id}/draft-position`,
    roster: (id: string) => `/api/members/${id}/roster`,
  },

  leagues: {
    base: `/api/leagues/`,
    byId: (leagueId: string) => `/api/leagues/${leagueId}`,

    members: {
      base: (leagueId: string) => `/api/leagues/${leagueId}/members`,
      join: (leagueId: string) => `/api/leagues/${leagueId}/members/join`,
    },

    poolEntries: {
      base: (leagueId: string) => `/api/leagues/${leagueId}/pool-entries`,
      byId: (leagueId: string, id: string) => `/api/leagues/${leagueId}/pool-entries/${id}`,
      byAvailable: (leagueId: string) => `/api/leagues/${leagueId}/pool-entries/available`,
      single: (leagueId: string) => `/api/leagues/${leagueId}/pool-entries/single`,
      batch: (leagueId: string) => `/api/leagues/${leagueId}/pool-entries/batch`,
    },

    draft: {
      base: (leagueId: string) => `/api/leagues/${leagueId}/draft`,
      // TF: why is this here? A league can only have one draft right now
      // did I really have this much forethought?
      byId: (leagueId: string, draftId: string) => `/api/leagues/${leagueId}/draft/${draftId}`,
      start: (leagueId: string) => `/api/leagues/${leagueId}/draft/start`,
      pick: (leagueId: string) => `/api/leagues/${leagueId}/draft/pick`,
      skip: (leagueId: string) => `/api/leagues/${leagueId}/draft/skip`,
    },

    draftPicks: {
      base: (leagueId: string) => `/api/leagues/${leagueId}/draft-picks`,
      // this GET does same as base
      history: (leagueId: string) => `/api/leagues/${leagueId}/draft-picks/history`,

      // returned json object key is snake_case 'next_pick_number'
      // TODO: server needs to fix
      nextPickNumber: (leagueId: string) => `/api/leagues/${leagueId}/draft-picks/next-pick-number`,

      // TODO: server needs to rename this to members
      player: {
        byId: (leagueId: string, playerId: string) => `/api/leagues/${leagueId}/draft-picks/player/${playerId}`,
      },
    },

    claims: {
      base: (leagueId: string) => `/api/leagues/${leagueId}/claims`,
      byId: (leagueId: string, id: string) => `/api/leagues/${leagueId}/claims/${id}`,
      released: (leagueId: string) => `/api/leagues/${leagueId}/claims/released`,

      player: {
        byId: (leagueId: string, playerId: string) => `/api/leagues/${leagueId}/claims/player/${playerId}`,
      },
    },

    games: {
      base: (leagueId: string) => `/api/leagues/${leagueId}/games`,
      byId: (leagueId: string, gameId: string) => `/api/leagues/${leagueId}/games/${gameId}`,

      report: (leagueId: string, gameId: string) => `/api/leagues/${leagueId}/games/report/${gameId}`,
      finalize: (leagueId: string, gameId: string) => `/api/leagues/${leagueId}/games/finalize/${gameId}`,

      startSeason: (leagueId: string) => `/api/leagues/${leagueId}/games/start-season`,
      generatePlayoffs: (leagueId: string) => `/api/leagues/${leagueId}/games/generate-playoffs`,

      members: {
        byId: (leagueId: string, memberId: string) => `/api/leagues/${leagueId}/games/members/${memberId}`,
      },
    },

    transfers: {
      startWindow: (leagueId: string) => `/api/leagues/${leagueId}/transfers/start`,
      endWindow: (leagueId: string) => `/api/leagues/${leagueId}/transfers/end`,

      dropClaim: (leagueId: string, claimId: string) => `/api/leagues/${leagueId}/transfers/drop/${claimId}`,
      pickupPoolEntry: (leagueId: string, poolEntryId: string) =>
        `/api/leagues/${leagueId}/transfers/pickup/${poolEntryId}`,
    },
  },
};
