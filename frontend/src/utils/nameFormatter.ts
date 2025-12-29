export const formatPokemonName = (name: string): string => {
    return name
        .replace(/-/g, ' ')
        .split(' ')
        .map(word => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ');
};

export const formatAbilityName = (name: string): string => {
    return name
        .replace(/-/g, ' ')
        .split(' ')
        .map(word => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ');
};

export const getPlayerSlug = (name: string): string => {
    return name.toLowerCase()
        .replace(/[^a-z0-9_.]+/g, '-') // Allow lowercase, numbers, underscores, dots. Replace others with hyphen.
        .replace(/(^-|-$)+/g, '');     // Remove leading/trailing hyphens.
};