import { PokemonAbility } from "../api/data_interfaces";

export enum Effectiveness {
    NEUTRAL = 1.0,
    SUPER_EFFECTIVE = 2.0,
    NOT_VERY_EFFECTIVE = 0.5,
    IMMUNE = 0.0,
}

// lowercase because db does lowercaps
export enum Type {
    NORMAL = "normal",
    FIRE = "fire",
    WATER = "water",
    GRASS = "grass",
    ELECTRIC = "electric",
    ICE = "ice",
    FIGHTING = "fighting",
    POISON = "poison",
    GROUND = "ground",
    FLYING = "flying",
    PSYCHIC = "psychic",
    BUG = "bug",
    ROCK = "rock",
    GHOST = "ghost",
    DRAGON = "dragon",
    DARK = "dark",
    STEEL = "steel",
    FAIRY = "fairy",
    // no mon is stellar by default
}

export type TypeEffectivenessMap = {
    [key in Type]?: Effectiveness;
};

export const typeChart: { [key in Type]?: TypeEffectivenessMap } = {
    [Type.NORMAL]: {
        [Type.ROCK]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.STEEL]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.GHOST]: Effectiveness.IMMUNE,
    },
    [Type.FIRE]: {
        [Type.FIRE]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.WATER]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.GRASS]: Effectiveness.SUPER_EFFECTIVE,
        [Type.ICE]: Effectiveness.SUPER_EFFECTIVE,
        [Type.BUG]: Effectiveness.SUPER_EFFECTIVE,
        [Type.ROCK]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.DRAGON]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.STEEL]: Effectiveness.SUPER_EFFECTIVE,
    },
    [Type.WATER]: {
        [Type.FIRE]: Effectiveness.SUPER_EFFECTIVE,
        [Type.WATER]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.GRASS]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.GROUND]: Effectiveness.SUPER_EFFECTIVE,
        [Type.ROCK]: Effectiveness.SUPER_EFFECTIVE,
        [Type.DRAGON]: Effectiveness.NOT_VERY_EFFECTIVE,
    },
    [Type.GRASS]: {
        [Type.FIRE]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.WATER]: Effectiveness.SUPER_EFFECTIVE,
        [Type.GRASS]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.POISON]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.GROUND]: Effectiveness.SUPER_EFFECTIVE,
        [Type.FLYING]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.BUG]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.ROCK]: Effectiveness.SUPER_EFFECTIVE,
        [Type.DRAGON]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.STEEL]: Effectiveness.NOT_VERY_EFFECTIVE,
    },
    [Type.ELECTRIC]: {
        [Type.WATER]: Effectiveness.SUPER_EFFECTIVE,
        [Type.GRASS]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.ELECTRIC]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.GROUND]: Effectiveness.IMMUNE,
        [Type.FLYING]: Effectiveness.SUPER_EFFECTIVE,
        [Type.DRAGON]: Effectiveness.NOT_VERY_EFFECTIVE,
    },
    [Type.ICE]: {
        [Type.FIRE]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.WATER]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.GRASS]: Effectiveness.SUPER_EFFECTIVE,
        [Type.ICE]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.GROUND]: Effectiveness.SUPER_EFFECTIVE,
        [Type.FLYING]: Effectiveness.SUPER_EFFECTIVE,
        [Type.DRAGON]: Effectiveness.SUPER_EFFECTIVE,
        [Type.STEEL]: Effectiveness.NOT_VERY_EFFECTIVE,
    },
    [Type.FIGHTING]: {
        [Type.NORMAL]: Effectiveness.SUPER_EFFECTIVE,
        [Type.ICE]: Effectiveness.SUPER_EFFECTIVE,
        [Type.POISON]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.FLYING]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.PSYCHIC]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.BUG]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.ROCK]: Effectiveness.SUPER_EFFECTIVE,
        [Type.GHOST]: Effectiveness.IMMUNE,
        [Type.DARK]: Effectiveness.SUPER_EFFECTIVE,
        [Type.STEEL]: Effectiveness.SUPER_EFFECTIVE,
        [Type.FAIRY]: Effectiveness.NOT_VERY_EFFECTIVE,
    },
    [Type.POISON]: {
        [Type.GRASS]: Effectiveness.SUPER_EFFECTIVE,
        [Type.POISON]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.GROUND]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.ROCK]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.GHOST]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.STEEL]: Effectiveness.IMMUNE,
        [Type.FAIRY]: Effectiveness.SUPER_EFFECTIVE,
    },
    [Type.GROUND]: {
        [Type.FIRE]: Effectiveness.SUPER_EFFECTIVE,
        [Type.GRASS]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.ELECTRIC]: Effectiveness.SUPER_EFFECTIVE,
        [Type.POISON]: Effectiveness.SUPER_EFFECTIVE,
        [Type.FLYING]: Effectiveness.IMMUNE,
        [Type.BUG]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.ROCK]: Effectiveness.SUPER_EFFECTIVE,
        [Type.STEEL]: Effectiveness.SUPER_EFFECTIVE,
    },
    [Type.FLYING]: {
        [Type.GRASS]: Effectiveness.SUPER_EFFECTIVE,
        [Type.ELECTRIC]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.FIGHTING]: Effectiveness.SUPER_EFFECTIVE,
        [Type.BUG]: Effectiveness.SUPER_EFFECTIVE,
        [Type.ROCK]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.STEEL]: Effectiveness.NOT_VERY_EFFECTIVE,
    },
    [Type.PSYCHIC]: {
        [Type.FIGHTING]: Effectiveness.SUPER_EFFECTIVE,
        [Type.POISON]: Effectiveness.SUPER_EFFECTIVE,
        [Type.PSYCHIC]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.DARK]: Effectiveness.IMMUNE,
        [Type.STEEL]: Effectiveness.NOT_VERY_EFFECTIVE,
    },
    [Type.BUG]: {
        [Type.FIRE]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.GRASS]: Effectiveness.SUPER_EFFECTIVE,
        [Type.FIGHTING]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.POISON]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.FLYING]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.PSYCHIC]: Effectiveness.SUPER_EFFECTIVE,
        [Type.GHOST]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.DARK]: Effectiveness.SUPER_EFFECTIVE,
        [Type.STEEL]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.FAIRY]: Effectiveness.NOT_VERY_EFFECTIVE,
    },
    [Type.ROCK]: {
        [Type.FIRE]: Effectiveness.SUPER_EFFECTIVE,
        [Type.ICE]: Effectiveness.SUPER_EFFECTIVE,
        [Type.FIGHTING]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.GROUND]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.FLYING]: Effectiveness.SUPER_EFFECTIVE,
        [Type.BUG]: Effectiveness.SUPER_EFFECTIVE,
        [Type.STEEL]: Effectiveness.NOT_VERY_EFFECTIVE,
    },
    [Type.GHOST]: {
        [Type.NORMAL]: Effectiveness.IMMUNE,
        [Type.PSYCHIC]: Effectiveness.SUPER_EFFECTIVE,
        [Type.GHOST]: Effectiveness.SUPER_EFFECTIVE,
        [Type.DARK]: Effectiveness.NOT_VERY_EFFECTIVE,
    },
    [Type.DRAGON]: {
        [Type.DRAGON]: Effectiveness.SUPER_EFFECTIVE,
        [Type.STEEL]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.FAIRY]: Effectiveness.IMMUNE,
    },
    [Type.DARK]: {
        [Type.FIGHTING]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.PSYCHIC]: Effectiveness.SUPER_EFFECTIVE,
        [Type.GHOST]: Effectiveness.SUPER_EFFECTIVE,
        [Type.DARK]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.FAIRY]: Effectiveness.NOT_VERY_EFFECTIVE,
    },
    [Type.STEEL]: {
        [Type.FIRE]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.WATER]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.ELECTRIC]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.ICE]: Effectiveness.SUPER_EFFECTIVE,
        [Type.ROCK]: Effectiveness.SUPER_EFFECTIVE,
        [Type.STEEL]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.FAIRY]: Effectiveness.SUPER_EFFECTIVE,
    },
    [Type.FAIRY]: {
        [Type.FIRE]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.FIGHTING]: Effectiveness.SUPER_EFFECTIVE,
        [Type.POISON]: Effectiveness.NOT_VERY_EFFECTIVE,
        [Type.DRAGON]: Effectiveness.SUPER_EFFECTIVE,
        [Type.DARK]: Effectiveness.SUPER_EFFECTIVE,
        [Type.STEEL]: Effectiveness.NOT_VERY_EFFECTIVE,
    },
};

const effectivenessAlteringAbilities: string[] = [
    // immmunities
    "dry-skin",
    "earth-eater",
    "flash-fire",
    "levitate",
    "lightning-rod",
    "sap-sipper",
    "storm-drain",
    "volt-absorb",
    "water-absorb",
    "wonder-guard",

    // resistances
    "purifying-salt",
    "thick-fat",
]


function getSingleTypeEffectiveness(attackingType: Type, defendingType: Type): Effectiveness {
    const effectivenessMap = typeChart[attackingType]

    if (effectivenessMap && effectivenessMap[defendingType] !== undefined) {
        return effectivenessMap[defendingType]
    }
    return Effectiveness.NEUTRAL
}

function getDualTypeEffectiveness(attackingType: Type, type1: Type, type2: Type | null): number {
    let typeEffectiveness: number = getSingleTypeEffectiveness(attackingType, type1);
    if (type2) {
        typeEffectiveness *= getSingleTypeEffectiveness(attackingType, (type2 as Type))
    }
    return typeEffectiveness;
}


export function getPokemonDefensiveProfile(type1: Type, type2: Type | null, abilities: PokemonAbility[]): [TypeEffectivenessMap, boolean, Type[]] {
    let didAbilityMatter = false;
    let defensiveProfile: TypeEffectivenessMap = {};
    let affectedTypes: Type[] = []

    const types: Type[] = [Type.NORMAL, Type.FIRE, Type.WATER, Type.GRASS, Type.ELECTRIC, Type.ICE, Type.FIGHTING, Type.POISON, Type.GROUND, Type.FLYING, Type.PSYCHIC, Type.BUG, Type.ROCK, Type.GHOST, Type.DRAGON, Type.DARK, Type.STEEL, Type.FAIRY]
    for (const attackingType of types) {
        var typeEffectiveness = getDualTypeEffectiveness(attackingType, type1, type2)
        defensiveProfile[attackingType] = typeEffectiveness;
    }

    const abilitiesToFactorIn = abilities.filter(ability =>
        effectivenessAlteringAbilities.includes(ability.Name)
    );
    for (const ability of abilitiesToFactorIn) {
        didAbilityMatter = true;
        switch (ability.Name) {
            case "dry-skin":
                defensiveProfile[Type.WATER]! *= Effectiveness.IMMUNE;
                defensiveProfile[Type.FIRE]! *= Effectiveness.SUPER_EFFECTIVE;
                affectedTypes.push(Type.WATER)
                affectedTypes.push(Type.FIRE)
                break;
            case "earth-eater":
            case "levitate":
                defensiveProfile[Type.GROUND]! *= Effectiveness.IMMUNE;
                affectedTypes.push(Type.GROUND)
                break;
            case "flash-fire":
                defensiveProfile[Type.FIRE]! *= Effectiveness.IMMUNE;
                affectedTypes.push(Type.FIRE)
                break;
            case "lightning-rod":
            case "volt-absorb":
                defensiveProfile[Type.ELECTRIC]! *= Effectiveness.IMMUNE;
                affectedTypes.push(Type.ELECTRIC)
                break;
            case "sap-sipper":
                defensiveProfile[Type.GRASS]! *= Effectiveness.IMMUNE;
                affectedTypes.push(Type.GRASS)
                break;
            case "storm-drain":
            case "water-absorb":
                defensiveProfile[Type.WATER]! *= Effectiveness.IMMUNE;
                affectedTypes.push(Type.WATER)
                break;
            case "purifying-salt":
                defensiveProfile[Type.GHOST]! *= Effectiveness.NOT_VERY_EFFECTIVE;
                affectedTypes.push(Type.GHOST)
                break;
            case "thick-fat":
                defensiveProfile[Type.FIRE]! *= Effectiveness.NOT_VERY_EFFECTIVE;
                defensiveProfile[Type.ICE]! *= Effectiveness.NOT_VERY_EFFECTIVE;
                affectedTypes.push(Type.FIRE)
                affectedTypes.push(Type.ICE)
                break;
            case "wonder-guard":
                Object.keys(defensiveProfile).forEach((key: string) => {
                    const attackingType = key as Type;
                    const currentEffectiveness = defensiveProfile[attackingType] ?? Effectiveness.NEUTRAL;
                    // If the current effectiveness is 1.0 (NEUTRAL) or less, make it IMMUNE
                    if (currentEffectiveness <= Effectiveness.NEUTRAL) {
                        // add to affectedTypes only if the effectiveness actually changes
                        if (defensiveProfile[attackingType] !== Effectiveness.IMMUNE) {
                            affectedTypes.push(attackingType);
                        }
                        defensiveProfile[attackingType] = Effectiveness.IMMUNE;
                    }
                });
                break;
            default:
                console.log("WHAT THE HELLY!!! An not considered ability to factor in made it into the ability checks for defensive profile calculations: ", ability);
                break;
        }
    }

    return [defensiveProfile, didAbilityMatter, affectedTypes];
}


