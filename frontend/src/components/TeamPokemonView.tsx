import React, { useState, useCallback } from 'react';
import { DraftedPokemon, PokemonStat } from '../api/data_interfaces';
import PokemonCard from './draftCards';
import { formatPokemonName, formatAbilityName } from '../utils/nameFormatter';
import ToggleSwitch from './ToggleSwitch'; // Ensure this import is present

const formatTypeName = (typeName: string): string => {
    return typeName.charAt(0).toUpperCase() + typeName.slice(1);
};

interface TeamPokemonViewProps {
    roster: DraftedPokemon[];
}

const TeamPokemonView: React.FC<TeamPokemonViewProps> = ({ roster }) => {
    const [isCardView, setIsCardView] = useState(false);
    const [currentlyFlippedCardId, setCurrentlyFlippedCardId] = useState<string | null>(null);

    const handleCardFlip = useCallback((pokemonId: string) => {
        setCurrentlyFlippedCardId(prevId => (prevId === pokemonId ? null : pokemonId));
    }, []);

    const handleImageError = useCallback((e: React.SyntheticEvent<HTMLImageElement, Event>) => {
        e.currentTarget.onerror = null;
        e.currentTarget.src = `https://placehold.co/150x150/cccccc/333333?text=No+Image`;
    }, []);


    const calcBST = (s: PokemonStat) =>
        s.Hp +
        s.Attack +
        s.Defense +
        s.SpecialAttack +
        s.SpecialDefense +
        s.Speed;
    const renderTable = () => (
        <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-background-primary">
                    <tr>
                        <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold uppercase tracking-wider w-60">Pokémon ({roster.length})</th>
                        <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold uppercase tracking-wider">Types</th>
                        <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold uppercase tracking-wider">Abilities</th>
                        <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold uppercase tracking-wider">Cost</th>
                        <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold uppercase tracking-wider">BST</th>
                        <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold uppercase tracking-wider">HP</th>
                        <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold uppercase tracking-wider">Att</th>
                        <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold uppercase tracking-wider">Def</th>
                        <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold uppercase tracking-wider">SpA</th>
                        <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold uppercase tracking-wider">SpD</th>
                        <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold uppercase tracking-wider">Spe</th>
                    </tr>
                </thead>
                <tbody className="bg-white">
                    {roster.map(dp => {
                        const s = dp.PokemonSpecies.Stats;
                        const bst = calcBST(s);

                        return (
                            <tr key={dp.ID} className="border-b border-gray-900/10">
                                <td className="px-2 py-3 whitespace-nowrap text-sm font-medium text-gray-900">
                                    <div className="flex items-center">
                                        <img
                                            src={dp.PokemonSpecies.Sprites.FrontDefault}
                                            alt={dp.PokemonSpecies.Name}
                                            className="h-8 w-8 mr-2"
                                        />
                                        <span className={dp.PokemonSpecies.Name.length > 14 ? 'text-xs' : 'text-sm'}>
                                            {formatPokemonName(dp.PokemonSpecies.Name)}
                                        </span>
                                    </div>
                                </td>

                                <td className="px-2 py-3 whitespace-nowrap text-sm text-gray-500">
                                    {dp.PokemonSpecies.Types.map(formatTypeName).join(', ')}
                                </td>

                                <td className="px-2 py-3 whitespace-nowrap text-sm text-gray-500">
                                    {dp.PokemonSpecies.Abilities.map(ability => (
                                        <div
                                            key={ability.Name}
                                            className={ability.IsHidden ? 'italic text-gray-400' : ''}
                                        >
                                            {formatAbilityName(ability.Name)}
                                        </div>
                                    ))}
                                </td>

                                <td className="px-2 py-3 whitespace-nowrap text-sm font-semibold text-gray-500">
                                    {dp.LeaguePokemon.Cost}
                                </td>

                                <td className="px-2 py-3 whitespace-nowrap text-sm font-semibold text-gray-500">{bst}</td>
                                <td className="px-2 py-3 whitespace-nowrap text-sm font-semibold text-gray-500">{s.Hp}</td>
                                <td className="px-2 py-3 whitespace-nowrap text-sm font-semibold text-gray-500">{s.Attack}</td>
                                <td className="px-2 py-3 whitespace-nowrap text-sm font-semibold text-gray-500">{s.Defense}</td>
                                <td className="px-2 py-3 whitespace-nowrap text-sm font-semibold text-gray-500">{s.SpecialAttack}</td>
                                <td className="px-2 py-3 whitespace-nowrap text-sm font-semibold text-gray-500">{s.SpecialDefense}</td>
                                <td className="px-2 py-3 whitespace-nowrap text-sm font-semibold text-gray-500">{s.Speed}</td>

                            </tr>
                        );
                    })}

                    {roster.length > 0 && (
                        <tr className="bg-gray-100">
                            <td className="px-2 py-3 whitespace-nowrap text-sm font-bold">Averages</td>
                            <td className="px-2 py-3 whitespace-nowrap text-sm text-text-secondary"></td>
                            <td className="px-2 py-3 whitespace-nowrap text-sm text-text-secondary"></td>

                            <td className="px-2 py-3 whitespace-nowrap text-sm text-text-secondary">
                                {(roster.reduce((sum, dp) => sum + dp.LeaguePokemon.Cost, 0) / roster.length).toFixed(1)}
                            </td>
                            <td className="px-2 py-3 whitespace-nowrap text-sm text-text-secondary">
                                {(
                                    roster.reduce(
                                        (sum, dp) => sum + calcBST(dp.PokemonSpecies.Stats),
                                        0
                                    ) / roster.length
                                ).toFixed(1)}
                            </td>
                            <td className="px-2 py-3 whitespace-nowrap text-sm text-text-secondary">
                                {(roster.reduce((sum, dp) => sum + dp.PokemonSpecies.Stats.Hp, 0) / roster.length).toFixed(1)}
                            </td>
                            <td className="px-2 py-3 whitespace-nowrap text-sm text-text-secondary">
                                {(roster.reduce((sum, dp) => sum + dp.PokemonSpecies.Stats.Attack, 0) / roster.length).toFixed(1)}
                            </td>
                            <td className="px-2 py-3 whitespace-nowrap text-sm text-text-secondary">
                                {(roster.reduce((sum, dp) => sum + dp.PokemonSpecies.Stats.Defense, 0) / roster.length).toFixed(1)}
                            </td>
                            <td className="px-2 py-3 whitespace-nowrap text-sm text-text-secondary">
                                {(roster.reduce((sum, dp) => sum + dp.PokemonSpecies.Stats.SpecialAttack, 0) / roster.length).toFixed(1)}
                            </td>
                            <td className="px-2 py-3 whitespace-nowrap text-sm text-text-secondary">
                                {(roster.reduce((sum, dp) => sum + dp.PokemonSpecies.Stats.SpecialDefense, 0) / roster.length).toFixed(1)}
                            </td>
                            <td className="px-2 py-3 whitespace-nowrap text-sm text-text-secondary">
                                {(roster.reduce((sum, dp) => sum + dp.PokemonSpecies.Stats.Speed, 0) / roster.length).toFixed(1)}
                            </td>

                        </tr>
                    )}
                </tbody>
            </table>
        </div>
    );

    const renderCardGrid = () => (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-2">
            {roster.map(dp => (
                <PokemonCard
                    key={dp.LeaguePokemonID}
                    leaguePokemonId={dp.LeaguePokemonID}
                    pokemon={dp.PokemonSpecies}
                    cost={dp.LeaguePokemon.Cost}
                    onImageError={handleImageError}
                    addPokemonToWishlist={() => { }} // No wishlist functionality in teamsheet view
                    removePokemonFromWishlist={() => { }} // No wishlist functionality in teamsheet view
                    isPokemonInWishlist={() => false} // Always false in teamsheet view
                    isFlipped={currentlyFlippedCardId === dp.LeaguePokemonID}
                    onFlip={handleCardFlip}
                    isDraftable={false} // Not draftable in teamsheet view
                    onDraft={() => { }} // No draft functionality in teamsheet view
                    isAvailable={true} // Always available in teamsheet view
                    isMyTurn={false} // Not relevant in teamsheet view
                    viewMode="teamsheet"
                    cardSize="default"
                />
            ))}
        </div>
    );

    return (
        <div className="bg-background-surface p-6 rounded-lg shadow-md">
            <div className="flex justify-between items-center mb-4">
                <h2 className="text-2xl font-bold text-text-primary">Pokémon Roster Details</h2>
                <ToggleSwitch
                    isOn={isCardView}
                    onToggle={() => setIsCardView(!isCardView)}
                    label="View as Cards"
                />
            </div>
            {roster.length > 0 ? (
                isCardView ? renderCardGrid() : renderTable()
            ) : (
                <p className="text-text-secondary">No Pokémon to display in this view.</p>
            )}
        </div>
    );
};

export default TeamPokemonView;
