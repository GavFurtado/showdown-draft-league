interface LeagueCreateRequest {
    id: string,
    name: string

}

interface UserProfileUpdateRequest {
    name?: string;
    email?: string;
}

interface PlayerProfileUpdateRequest {
    pokemon?: string[]; // Assuming an array of Pokemon names/IDs
}

