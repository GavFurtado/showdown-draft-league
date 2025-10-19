import React from 'react';
import { DraftedPokemon } from '../api/data_interfaces';
import { getPokemonDefensiveProfile, Type, Effectiveness, AllTypes } from '../utils/typeChart';
import { formatPokemonName } from '../utils/nameFormatter';

interface DefensiveTypeChartProps {
    roster: DraftedPokemon[];
}

const getEffectivenessColor = (effectiveness: Effectiveness): string => {
    switch (effectiveness) {
        case Effectiveness.IMMUNE:
            return 'bg-gray-600 text-white font-bold';
        case Effectiveness.BARELY_EFFECTIVE:
            return 'bg-green-700 text-white font-bold';
        case Effectiveness.NOT_VERY_EFFECTIVE:
            return 'bg-lime-700 text-white font-bold';
        case Effectiveness.SUPER_EFFECTIVE:
            return 'bg-amber-700 text-white font-bold';
        case Effectiveness.EXTREMELY_EFFECTIVE:
            return 'bg-red-700 text-white font-bold';
        case Effectiveness.NEUTRAL:
        default:
            return 'bg-gray-200 text-gray-500 font-bold text-shadow-2xs';
    }
};

const formatEffectiveness = (effectiveness: Effectiveness): string => {
    if (effectiveness === 0.0) return '0x';
    if (effectiveness === 0.25) return '0.25x';
    if (effectiveness === 0.5) return '0.5x';
    if (effectiveness === 1.0) return '1x';
    if (effectiveness === 2.0) return '2x';
    if (effectiveness === 4.0) return '4x';
    return (effectiveness as number).toString() + 'x';
};

const formatTypeName = (typeName: string): string => {
    if (typeName === 'fighting') return 'fight';
    if (typeName === 'normal') return 'norm';
    if (typeName === 'ground') return 'grnd';
    if (typeName === 'psychic') return 'psych';
    if (typeName === 'water') return 'water';
    if (typeName === 'flying') return 'fly';
    if (typeName === 'steel') return 'steel';
    if (typeName === 'fairy') return 'fairy';
    if (typeName === 'grass') return 'grass';
    return typeName.slice(0, 4);
};

export const DefensiveTypeChart: React.FC<DefensiveTypeChartProps> = ({ roster }) => {
    if (!roster || roster.length === 0) {
        return <p className="p-4 text-text-secondary">No Pokémon in this roster to display defensive types.</p>;
    }

    return (
        <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-700 table-fixed">
                <thead className="bg-background-primary">
                    <tr>
                        <th scope="col" className="px-2 py-3 text-text-on-nav text-left text-[10px] uppercase tracking-wider w-40">
                            Pokémon
                        </th>
                        {AllTypes.map(type => (
                            <th key={type} scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-12">
                                {formatTypeName(type)}
                            </th>
                        ))}
                    </tr>
                </thead>
                <tbody className="bg-white">
                    {roster.map(dp => {
                        const [defensiveProfile, _didAbilityMatter, affectedTypes] = getPokemonDefensiveProfile(
                            dp.PokemonSpecies.Types[0] as Type,
                            dp.PokemonSpecies.Types[1] ? dp.PokemonSpecies.Types[1] as Type : null,
                            dp.PokemonSpecies.Abilities
                        );

                        return (
                            <tr key={dp.ID} className='border-b border-gray-900/10'>
                                <td className="px-2 py-2 whitespace-nowrap text-sm font-medium text-gray-900">
                                    <div className="flex items-center">
                                        <img src={dp.PokemonSpecies.Sprites.FrontDefault} alt={dp.PokemonSpecies.Name} className="h-8 w-8 mr-2" />
                                        <span className={dp.PokemonSpecies.Name.length > 14 ? 'text-xs' : 'text-sm'}>
                                            {formatPokemonName(dp.PokemonSpecies.Name)}
                                        </span>
                                    </div>
                                </td>                                {AllTypes.map(attackingType => {
                                    const effectiveness = defensiveProfile[attackingType] ?? Effectiveness.NEUTRAL;
                                    const isAffectedByAbility = affectedTypes.includes(attackingType);
                                    return (
                                        <td key={attackingType} className={`px-2 py-2 whitespace-nowrap text-center text-xs ${getEffectivenessColor(effectiveness)} ${isAffectedByAbility ? 'font-bold' : ''}`}>
                                            {formatEffectiveness(effectiveness)}{isAffectedByAbility ? ' *' : ''}
                                        </td>
                                    );
                                })}
                            </tr>
                        );
                    })}
                </tbody>
            </table>
        </div>
    );
};
