import { DraftedPokemon } from "../api/data_interfaces";

interface DefensiveTypeChartProps {
    roster: DraftedPokemon
}

enum Eff { // short for effectiveness
    NEUTRAL = 1.0,
    SUPER_EFFECTIVE = 2.0,
    EXTREMELY_EFFECTIVE = 2.0,
    NOT_VERY_EFFECTIVE = 0.5,
    BARELY_EFFECTIVE = 0.25,
    IMMUNE = 0.0,
}

enum Ty { // short for Types
    NORMAL = "Normal",
    FIRE = "Fire",
    WATER = "Water",
    GRASS = "Grass",
    ELECTRIC = "Electric",
    ICE = "Ice",
    FIGHTING = "Fighting",
    POISON = "Poison",
    GROUND = "Ground",
    FLYING = "Flying",
    PSYCHIC = "Psychic",
    BUG = "Bug",
    ROCK = "Rock",
    GHOST = "Ghost",
    DRAGON = "Dragon",
    DARK = "Dark",
    STEEL = "Steel",
    FAIRY = "Fairy",
}

type TypeEffectivenessMap = {
    [key in Ty]?: Eff;
};

const typeChart: { [key in Ty]?: TypeEffectivenessMap } = {
    [Ty.NORMAL]: {
        [Ty.ROCK]: Eff.NOT_VERY_EFFECTIVE,
        [Ty.STEEL]: Eff.NOT_VERY_EFFECTIVE,
        [Ty.GHOST]: Eff.IMMUNE,
    }
}

export default function DefensiveTypeChart({ roster }: DefensiveTypeChartProps) {


    return
}
