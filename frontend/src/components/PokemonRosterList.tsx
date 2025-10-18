import { log } from 'console';
import { DraftedPokemon, LeaguePokemon } from '../api/data_interfaces';
import { PokemonListItem } from './PokemonListItem';

interface PokemonRosterListProps {
  roster: (DraftedPokemon | LeaguePokemon)[];
  rosterType: 'drafted' | 'pendingPick';
  bgColor?: string;
  onRemove?: (leaguePokemonId: string) => void;
  showRemoveButton?: boolean;
}

export const PokemonRosterList = ({ roster, rosterType, bgColor, onRemove, showRemoveButton }: PokemonRosterListProps) => {
    if (roster.length === 0) {
        return <p className="text-text-secondary">No Pokémon on this roster.</p>;
    }

    return (
        <div className="flex flex-col space-y-2"> {/* Common wrapper styling */}
            {roster.map((item, _) => {
                {/* console.log("PokemonRosterList:: roster item prop: ", item); */ }

                return (
                    <PokemonListItem
                        key={item.ID}
                        pokemon={rosterType === 'drafted' ? (item as DraftedPokemon).PokemonSpecies
                            : (item as LeaguePokemon).PokemonSpecies}
                        cost={rosterType === 'drafted' ? (item as DraftedPokemon).LeaguePokemon.Cost
                            : (item as LeaguePokemon).Cost}
                        leaguePokemonId={rosterType === 'drafted' ? (item as DraftedPokemon).LeaguePokemonID
                            : (item as LeaguePokemon).ID}
                        pickNumber={rosterType === 'drafted' ? (item as DraftedPokemon).DraftPickNumber : undefined}
                        bgColor={bgColor || "bg-gray-50"}
                        onRemove={onRemove}
                        showRemoveButton={showRemoveButton}
                    />
                );
            })}
        </div>
    );
};
