import { useState } from 'react';
import { DraftCardProps } from '../api/data_interfaces';

// Helper function to get stat color based on value
const getStatColor = (value: number): string => {
    if (value <= 25) {
        return '#FF0000'; // Red
    } else if (value <= 60) {
        return '#FFA500'; // Orange
    } else if (value <= 89) {
        return '#FFFF00'; // Yellow
    } else if (value <= 120) {
        return '#A0F555'; // Lime Green
    } else if (value <= 199) {
        return '#23CD5E'; // Darker Green
    } else {
        return '#02FFFF'; // Cyan
    }
};

// StatBar Component
interface StatBarProps {
    label: string;
    value: number;
}

const StatBar: React.FC<StatBarProps> = ({ label, value }) => {
    const maxWidth = 255; // Max possible stat value
    const barWidth = (value / maxWidth) * 100; // Calculate width as percentage
    const color = getStatColor(value);

    return (
        <div className="flex items-center gap-2">
            <span className="w-10 text-right text-xs font-medium">{label}:</span>
            <div className="flex-1 bg-gray-600 h-4">
                <div
                    className="h-4"
                    style={{ width: `${barWidth}%`, backgroundColor: color }}
                ></div>
            </div>
            <span className="w-8 text-left text-sm">{value}</span>
        </div>
    );
};

export default function PokemonCard({ pokemon, cost, onImageError }: DraftCardProps) {
    const [isFlipped, setIsFlipped] = useState(false);
    const handleFlip = () => {
        setIsFlipped(!isFlipped);
    };
    const types = pokemon.types.map(t => {
        if (typeof t === 'string' && t.length > 0) {
            return t.charAt(0).toUpperCase() + t.slice(1);
        }
        return t;
    }).join(', ');
    const abilities = pokemon.abilities.map(a => a.name).join(', ')
    return (
        // Container of the whole thing : Sets perspective for effect and provides a clickable area
        <div
            className="group h-70 w-47 rounded-lg shadow-lg relative cursor-pointer [perspective:1000px]"
            onClick={handleFlip}
        >
            {/* Inner container: This is the part that actually flips */}
            <div
                className={`
            relative
            w-full h-full
            transition-transform duration-700 ease-in-out
            [transform-style:preserve-3d]
            ${isFlipped ? '[transform:rotateY(180deg)]' : ''}
            `}
            >
                {/* Front Face of the Card */}
                <div className="absolute inset-0 bg-white rounded-lg p-4 flex flex-col items-center justify-center [backface-visibility:hidden]">
                    {/* Pokémon Image */}
                    <div className="relative w-full h-[100%]">
                        <img
                            src={pokemon.sprites.front_default}
                            alt={pokemon.name}
                            onError={onImageError}
                            className="w-[100%] h-[100%] object-contain mb-4 bg-gray-100 p-2"
                        />
                        {/* Pokémon Cost */}
                        <p className="text-lg font-semibold absolute  bottom-2 right-2 ">
                            {cost}
                        </p>
                    </div>
                    <div className='flex w-[100%] justify-between'>
                        {/* Pokémon Name */}
                        <div>
                            <h3 className="text-lg pb-0 mb-0 font-bold text-gray-800 text-left">
                                {pokemon.name}
                            </h3>
                            <p className='p-0 m-0 text-left text-sm text-gray-600'>{types}</p>
                        </div>
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                            }}
                            className="relative flex items-center align-center justify-center mt-4 h-7.5 w-7.5 rounded-full p-0 hover:border-2 hover: transition-colors">
                            <img
                                src="https://upload.wikimedia.org/wikipedia/commons/5/53/Poké_Ball_icon.svg"
                                className='h-7.5 w-7.5'
                            />
                        </button>
                    </div>

                </div>

                {/* Back Face of the Card */}
                <div className="absolute inset-0 bg-gray-700 text-white rounded-lg p-4 flex flex-col [backface-visibility:hidden] [transform:rotateY(180deg)]">
                    <h3 className="text-l font-bold mb-4 text-center">
                        {pokemon.name}
                    </h3>
                    <div className="flex flex-col gap-1 w-full">
                        <StatBar label="HP" value={pokemon.stats.hp} />
                        <StatBar label="Att" value={pokemon.stats.attack} />
                        <StatBar label="Def" value={pokemon.stats.defense} />
                        <StatBar label="SpA" value={pokemon.stats["special_attack"]} />
                        <StatBar label="SpD" value={pokemon.stats["special_defense"]} />
                        <StatBar label="Spe" value={pokemon.stats.speed} />
                    </div>
                    <div className="text-left text-xs mt-4">
                        <p className="font-xs"><span className='font-bold'>Abilities:</span> {abilities}</p>
                    </div>
                    <p className="text-base text-center mb-auto">

                    </p>
                </div>
            </div>
        </div>
    );
}
